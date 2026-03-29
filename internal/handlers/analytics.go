package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	"inventory-ims/internal/database"
	"inventory-ims/internal/models"

	"github.com/gin-gonic/gin"
)

type AnalyticsHandler struct{}

func NewAnalyticsHandler() *AnalyticsHandler {
	return &AnalyticsHandler{}
}

type KPIResponse struct {
	TotalProducts     int64   `json:"total_products"`
	TotalCategories   int64   `json:"total_categories"`
	TotalSuppliers    int64   `json:"total_suppliers"`
	TotalWarehouses   int64   `json:"total_warehouses"`
	TotalValue        float64 `json:"total_value"`
	TotalCost         float64 `json:"total_cost"`
	PotentialProfit   float64 `json:"potential_profit"`
	LowStockCount     int     `json:"low_stock_count"`
	OutOfStockCount   int     `json:"out_of_stock_count"`
	PendingPOCount    int64   `json:"pending_po_count"`
	ActiveProduction  int64   `json:"active_production"`
	TodayTransactions int64   `json:"today_transactions"`
	WeekTransactions  int64   `json:"week_transactions"`
	MonthTransactions int64   `json:"month_transactions"`
	AvgDailyTrans     float64 `json:"avg_daily_transactions"`
}

type TrendData struct {
	Date             string  `json:"date"`
	StockIn          int     `json:"stock_in"`
	StockOut         int     `json:"stock_out"`
	NetChange        int     `json:"net_change"`
	TransactionCount int     `json:"transaction_count"`
	TotalValue       float64 `json:"total_value"`
}

type ABCItem struct {
	ProductID   uint    `json:"product_id"`
	ProductName string  `json:"product_name"`
	SKU         string  `json:"sku"`
	Quantity    int     `json:"quantity"`
	UnitCost    float64 `json:"unit_cost"`
	TotalValue  float64 `json:"total_value"`
	Percentage  float64 `json:"percentage"`
	Cumulative  float64 `json:"cumulative"`
	Class       string  `json:"class"`
}

type TopMover struct {
	ProductID   uint    `json:"product_id"`
	ProductName string  `json:"product_name"`
	SKU         string  `json:"sku"`
	TotalIn     int     `json:"total_in"`
	TotalOut    int     `json:"total_out"`
	NetChange   int     `json:"net_change"`
	Turnover    float64 `json:"turnover"`
}

type CategoryStat struct {
	CategoryName string  `json:"category_name"`
	ProductCount int     `json:"product_count"`
	TotalValue   float64 `json:"total_value"`
	TotalQty     int     `json:"total_quantity"`
}

type DashboardData struct {
	KPI          KPIResponse    `json:"kpi"`
	Trends       []TrendData    `json:"trends"`
	CategoryData []CategoryStat `json:"category_data"`
	TopMovers    []TopMover     `json:"top_movers"`
	ABCAnalysis  []ABCItem      `json:"abc_analysis"`
	LowStock     []LowStockItem `json:"low_stock"`
	RecentTX     []RecentTX     `json:"recent_transactions"`
}

func (h *AnalyticsHandler) GetDashboardData(c *gin.Context) {
	var result DashboardData

	result.KPI = h.getKPI()
	result.Trends = h.getTrends(30)
	result.CategoryData = h.getCategoryStats()
	result.TopMovers = h.getTopMovers(30)
	result.ABCAnalysis = h.getABCAnalysis()
	result.LowStock = h.getLowStockItems()
	result.RecentTX = h.getRecentTransactions(10)

	c.JSON(http.StatusOK, result)
}

func (h *AnalyticsHandler) getKPI() KPIResponse {
	var kpi KPIResponse

	database.DB.Model(&models.Product{}).Where("is_active = ?", true).Count(&kpi.TotalProducts)
	database.DB.Model(&models.Category{}).Count(&kpi.TotalCategories)
	database.DB.Model(&models.Supplier{}).Count(&kpi.TotalSuppliers)
	database.DB.Model(&models.Warehouse{}).Count(&kpi.TotalWarehouses)

	var inventory []models.Inventory
	database.DB.Preload("Product").Find(&inventory)

	for _, inv := range inventory {
		if inv.Product != nil {
			kpi.TotalValue += float64(inv.Quantity) * inv.Product.UnitPrice
			kpi.TotalCost += float64(inv.Quantity) * inv.Product.CostPrice

			if inv.Product.ReorderPoint > 0 && inv.Quantity <= inv.Product.ReorderPoint {
				kpi.LowStockCount++
			}
			if inv.Quantity == 0 {
				kpi.OutOfStockCount++
			}
		}
	}

	kpi.PotentialProfit = kpi.TotalValue - kpi.TotalCost

	database.DB.Model(&models.PurchaseOrder{}).Where("status IN ?", []string{"pending", "ordered"}).Count(&kpi.PendingPOCount)

	database.DB.Model(&models.ProductionOrder{}).Where("status = ?", models.ProductionStatusInProgress).Count(&kpi.ActiveProduction)

	today := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.UTC)
	database.DB.Model(&models.Transaction{}).Where("created_at >= ?", today).Count(&kpi.TodayTransactions)

	weekAgo := today.AddDate(0, 0, -7)
	database.DB.Model(&models.Transaction{}).Where("created_at >= ?", weekAgo).Count(&kpi.WeekTransactions)

	monthAgo := today.AddDate(0, -1, 0)
	database.DB.Model(&models.Transaction{}).Where("created_at >= ?", monthAgo).Count(&kpi.MonthTransactions)

	kpi.AvgDailyTrans = float64(kpi.MonthTransactions) / 30.0

	return kpi
}

func (h *AnalyticsHandler) getTrends(days int) []TrendData {
	startDate := time.Now().AddDate(0, 0, -days)

	var transactions []models.Transaction
	database.DB.Where("created_at >= ?", startDate).Order("created_at ASC").Find(&transactions)

	trendMap := make(map[string]*TrendData)

	for i := days; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		trendMap[date] = &TrendData{Date: date}
	}

	for _, t := range transactions {
		date := t.CreatedAt.Format("2006-01-02")
		if trendMap[date] == nil {
			trendMap[date] = &TrendData{Date: date}
		}

		trendMap[date].TransactionCount++

		if t.Type == models.TransactionTypePurchase || t.Type == models.TransactionTypeProductionIn {
			trendMap[date].StockIn += t.Quantity
		} else if t.Type == models.TransactionTypeSale || t.Type == models.TransactionTypeProductionOut {
			trendMap[date].StockOut += -t.Quantity
		}

		trendMap[date].NetChange = trendMap[date].StockIn - trendMap[date].StockOut
	}

	var result []TrendData
	for _, v := range trendMap {
		result = append(result, *v)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Date < result[j].Date
	})

	return result
}

func (h *AnalyticsHandler) getCategoryStats() []CategoryStat {
	var catStats []CategoryStat

	rows, _ := database.DB.Raw(`
		SELECT 
			COALESCE(c.name, 'Uncategorized') as category_name,
			COUNT(DISTINCT p.id) as product_count,
			COALESCE(SUM(i.quantity * p.unit_price), 0) as total_value,
			COALESCE(SUM(i.quantity), 0) as total_quantity
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		LEFT JOIN inventory i ON p.id = i.product_id
		WHERE p.is_active = 1
		GROUP BY c.id, c.name
		ORDER BY total_value DESC
	`).Rows()

	defer rows.Close()
	for rows.Next() {
		var stat CategoryStat
		var catName *string
		var value, qty float64
		rows.Scan(&catName, &stat.ProductCount, &value, &qty)
		if catName != nil {
			stat.CategoryName = *catName
		}
		stat.TotalValue = value
		stat.TotalQty = int(qty)
		catStats = append(catStats, stat)
	}

	return catStats
}

func (h *AnalyticsHandler) getTopMovers(days int) []TopMover {
	startDate := time.Now().AddDate(0, 0, -days)

	type result struct {
		ProductID   uint
		ProductName string
		SKU         string
		TotalIn     int
		TotalOut    int
	}

	var results []result

	rows, _ := database.DB.Raw(`
		SELECT 
			p.id as product_id,
			p.name as product_name,
			p.sku as sku,
			COALESCE(SUM(CASE WHEN t.quantity > 0 THEN t.quantity ELSE 0 END), 0) as total_in,
			COALESCE(SUM(CASE WHEN t.quantity < 0 THEN ABS(t.quantity) ELSE 0 END), 0) as total_out
		FROM transactions t
		JOIN products p ON t.product_id = p.id
		WHERE t.created_at >= ?
		GROUP BY p.id, p.name, p.sku
		ORDER BY (total_in + total_out) DESC
		LIMIT 10
	`, startDate).Rows()

	defer rows.Close()
	for rows.Next() {
		var r result
		rows.Scan(&r.ProductID, &r.ProductName, &r.SKU, &r.TotalIn, &r.TotalOut)
		results = append(results, r)
	}

	var topMovers []TopMover
	for _, r := range results {
		net := r.TotalIn - r.TotalOut
		turnover := 0.0
		if net != 0 {
			turnover = float64(r.TotalIn+r.TotalOut) / float64(net)
		}
		topMovers = append(topMovers, TopMover{
			ProductID:   r.ProductID,
			ProductName: r.ProductName,
			SKU:         r.SKU,
			TotalIn:     r.TotalIn,
			TotalOut:    r.TotalOut,
			NetChange:   net,
			Turnover:    turnover,
		})
	}

	return topMovers
}

func (h *AnalyticsHandler) getABCAnalysis() []ABCItem {
	type prodValue struct {
		ProductID   uint
		ProductName string
		SKU         string
		Quantity    int
		UnitCost    float64
		TotalValue  float64
	}

	var products []models.Product
	database.DB.Preload("Inventory").Where("is_active = ?", true).Find(&products)

	var items []prodValue
	var totalValue float64

	for _, p := range products {
		qty := 0
		if p.Inventory != nil {
			qty = p.Inventory.Quantity
		}
		value := float64(qty) * p.CostPrice
		if value > 0 {
			items = append(items, prodValue{
				ProductID:   p.ID,
				ProductName: p.Name,
				SKU:         p.SKU,
				Quantity:    qty,
				UnitCost:    p.CostPrice,
				TotalValue:  value,
			})
			totalValue += value
		}
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].TotalValue > items[j].TotalValue
	})

	var abc []ABCItem
	var cumulative float64

	for i, item := range items {
		percentage := (item.TotalValue / totalValue) * 100
		cumulative += percentage

		class := "C"
		if cumulative <= 80 {
			class = "A"
		} else if cumulative <= 95 {
			class = "B"
		}

		abc = append(abc, ABCItem{
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			SKU:         item.SKU,
			Quantity:    item.Quantity,
			UnitCost:    item.UnitCost,
			TotalValue:  item.TotalValue,
			Percentage:  percentage,
			Cumulative:  cumulative,
			Class:       class,
		})

		if i >= 19 {
			break
		}
	}

	return abc
}

type LowStockItem struct {
	ProductID    uint   `json:"product_id"`
	ProductName  string `json:"product_name"`
	SKU          string `json:"sku"`
	Quantity     int    `json:"quantity"`
	ReorderPoint int    `json:"reorder_point"`
	Deficit      int    `json:"deficit"`
	Status       string `json:"status"`
}

func (h *AnalyticsHandler) getLowStockItems() []LowStockItem {
	var products []models.Product
	database.DB.Preload("Inventory").Where("is_active = ? AND reorder_point > 0", true).Find(&products)

	var items []LowStockItem
	for _, p := range products {
		if p.Inventory != nil && p.Inventory.Quantity <= p.ReorderPoint {
			status := "low_stock"
			if p.Inventory.Quantity == 0 {
				status = "out_of_stock"
			}
			items = append(items, LowStockItem{
				ProductID:    p.ID,
				ProductName:  p.Name,
				SKU:          p.SKU,
				Quantity:     p.Inventory.Quantity,
				ReorderPoint: p.ReorderPoint,
				Deficit:      p.ReorderPoint - p.Inventory.Quantity,
				Status:       status,
			})
		}
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Deficit > items[j].Deficit
	})

	if len(items) > 10 {
		items = items[:10]
	}

	return items
}

type RecentTX struct {
	ID          uint      `json:"id"`
	ProductName string    `json:"product_name"`
	Type        string    `json:"type"`
	Quantity    int       `json:"quantity"`
	UserName    string    `json:"user_name"`
	CreatedAt   time.Time `json:"created_at"`
}

func (h *AnalyticsHandler) getRecentTransactions(limit int) []RecentTX {
	var txs []models.Transaction
	database.DB.Preload("Product").Preload("User").Order("created_at DESC").Limit(limit).Find(&txs)

	var recent []RecentTX
	for _, t := range txs {
		prodName := ""
		if t.Product != nil {
			prodName = t.Product.Name
		}
		userName := ""
		if t.User != nil {
			userName = t.User.Username
		}
		recent = append(recent, RecentTX{
			ID:          t.ID,
			ProductName: prodName,
			Type:        t.Type,
			Quantity:    t.Quantity,
			UserName:    userName,
			CreatedAt:   t.CreatedAt,
		})
	}

	return recent
}

func (h *AnalyticsHandler) GetABC(c *gin.Context) {
	abc := h.getABCAnalysis()
	c.JSON(http.StatusOK, abc)
}

func (h *AnalyticsHandler) GetTrends(c *gin.Context) {
	days := 30
	if d := c.Query("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil {
			days = parsed
		}
	}

	trends := h.getTrends(days)
	c.JSON(http.StatusOK, trends)
}

func (h *AnalyticsHandler) GetTopMovers(c *gin.Context) {
	days := 30
	if d := c.Query("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil {
			days = parsed
		}
	}

	movers := h.getTopMovers(days)
	c.JSON(http.StatusOK, movers)
}

func (h *AnalyticsHandler) GetCategoryStats(c *gin.Context) {
	stats := h.getCategoryStats()
	c.JSON(http.StatusOK, stats)
}

func (h *AnalyticsHandler) GetKPI(c *gin.Context) {
	kpi := h.getKPI()
	c.JSON(http.StatusOK, kpi)
}

func (h *AnalyticsHandler) PredictLowStock(c *gin.Context) {
	type prediction struct {
		ProductID            uint    `json:"product_id"`
		ProductName          string  `json:"product_name"`
		CurrentStock         int     `json:"current_stock"`
		DailyUsage           float64 `json:"daily_usage"`
		DaysUntilStockout    float64 `json:"days_until_stockout"`
		ReorderPoint         int     `json:"reorder_point"`
		RecommendedOrderDate string  `json:"recommended_order_date"`
	}

	var products []models.Product
	database.DB.Preload("Inventory").Where("is_active = ? AND reorder_point > 0", true).Find(&products)

	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)

	var predictions []prediction

	for _, p := range products {
		if p.Inventory == nil || p.Inventory.Quantity == 0 {
			continue
		}

		var totalOut int64
		database.DB.Model(&models.Transaction{}).
			Where("product_id = ? AND type IN ? AND created_at >= ?",
				p.ID, []string{models.TransactionTypeSale, models.TransactionTypeProductionOut}, thirtyDaysAgo).
			Count(&totalOut)

		dailyUsage := float64(totalOut) / 30.0
		daysUntilStockout := 0.0
		if dailyUsage > 0 {
			daysUntilStockout = float64(p.Inventory.Quantity) / dailyUsage
		}

		daysUntilReorder := float64(p.Inventory.Quantity-p.ReorderPoint) / dailyUsage
		if dailyUsage == 0 {
			daysUntilReorder = 999
		}

		recommendedDate := time.Now().AddDate(0, 0, int(daysUntilReorder-7))
		if daysUntilReorder <= 7 {
			recommendedDate = time.Now()
		}

		if daysUntilStockout < 30 {
			predictions = append(predictions, prediction{
				ProductID:            p.ID,
				ProductName:          p.Name,
				CurrentStock:         p.Inventory.Quantity,
				DailyUsage:           dailyUsage,
				DaysUntilStockout:    daysUntilStockout,
				ReorderPoint:         p.ReorderPoint,
				RecommendedOrderDate: recommendedDate.Format("2006-01-02"),
			})
		}
	}

	sort.Slice(predictions, func(i, j int) bool {
		return predictions[i].DaysUntilStockout < predictions[j].DaysUntilStockout
	})

	c.JSON(http.StatusOK, predictions)
}

func (h *AnalyticsHandler) GetInventoryTurnover(c *gin.Context) {
	days := 90
	if d := c.Query("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil {
			days = parsed
		}
	}

	startDate := time.Now().AddDate(0, 0, -days)

	type turnover struct {
		ProductID    uint    `json:"product_id"`
		ProductName  string  `json:"product_name"`
		SKU          string  `json:"sku"`
		AvgInventory float64 `json:"avg_inventory"`
		TotalOut     int     `json:"total_out"`
		TurnoverRate float64 `json:"turnover_rate"`
	}

	var products []models.Product
	database.DB.Preload("Inventory").Where("is_active = ?", true).Find(&products)

	var results []turnover

	for _, p := range products {
		if p.Inventory == nil {
			continue
		}

		avgInventory := float64(p.Inventory.Quantity)

		var totalOut int64
		database.DB.Model(&models.Transaction{}).
			Where("product_id = ? AND quantity < 0 AND created_at >= ?", p.ID, startDate).
			Count(&totalOut)

		turnoverRate := 0.0
		if avgInventory > 0 {
			turnoverRate = float64(totalOut) / avgInventory * (365.0 / float64(days))
		}

		results = append(results, turnover{
			ProductID:    p.ID,
			ProductName:  p.Name,
			SKU:          p.SKU,
			AvgInventory: avgInventory,
			TotalOut:     int(totalOut),
			TurnoverRate: turnoverRate,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].TurnoverRate > results[j].TurnoverRate
	})

	c.JSON(http.StatusOK, results)
}

func (h *AnalyticsHandler) ExportDashboardJSON(c *gin.Context) {
	var result DashboardData

	result.KPI = h.getKPI()
	result.Trends = h.getTrends(30)
	result.CategoryData = h.getCategoryStats()
	result.TopMovers = h.getTopMovers(30)
	result.ABCAnalysis = h.getABCAnalysis()
	result.LowStock = h.getLowStockItems()
	result.RecentTX = h.getRecentTransactions(10)

	data := result

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate JSON"})
		return
	}

	c.Header("Content-Type", "application/json")
	c.Header("Content-Disposition", "attachment; filename=dashboard_data.json")
	c.String(http.StatusOK, string(jsonData))
}

func init() {
	_ = fmt.Sprintf
}
