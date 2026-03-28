package handlers

import (
	"encoding/csv"
	"net/http"
	"strconv"
	"time"

	"inventory-ims/internal/database"
	"inventory-ims/internal/models"

	"github.com/gin-gonic/gin"
)

type ReportHandler struct{}

func NewReportHandler() *ReportHandler {
	return &ReportHandler{}
}

type StockLevelReport struct {
	ProductID    uint   `json:"product_id"`
	ProductName  string `json:"product_name"`
	SKU          string `json:"sku"`
	CategoryName string `json:"category_name"`
	Quantity     int    `json:"quantity"`
	ReorderPoint int    `json:"reorder_point"`
	Status       string `json:"status"`
}

type ValuationReport struct {
	ProductID   uint    `json:"product_id"`
	ProductName string  `json:"product_name"`
	SKU         string  `json:"sku"`
	Quantity    int     `json:"quantity"`
	UnitCost    float64 `json:"unit_cost"`
	TotalValue  float64 `json:"total_value"`
}

type DashboardStats struct {
	TotalProducts     int64   `json:"total_products"`
	TotalValue        float64 `json:"total_value"`
	LowStockCount     int     `json:"low_stock_count"`
	PendingPOCount    int64   `json:"pending_po_count"`
	ActiveProduction  int64   `json:"active_production"`
	TodayTransactions int64   `json:"today_transactions"`
}

func (h *ReportHandler) GetDashboard(c *gin.Context) {
	var stats DashboardStats

	database.DB.Model(&models.Product{}).Where("is_active = ?", true).Count(&stats.TotalProducts)

	var inventory []models.Inventory
	database.DB.Preload("Product").Find(&inventory)
	for _, inv := range inventory {
		if inv.Product != nil {
			stats.TotalValue += float64(inv.Quantity) * inv.Product.CostPrice
		}
		if inv.Product != nil && inv.Quantity <= inv.Product.ReorderPoint {
			stats.LowStockCount++
		}
	}

	database.DB.Model(&models.PurchaseOrder{}).Where("status IN ?", []string{"pending", "ordered"}).Count(&stats.PendingPOCount)

	database.DB.Model(&models.ProductionOrder{}).Where("status = ?", models.ProductionStatusInProgress).Count(&stats.ActiveProduction)

	today := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.UTC)
	database.DB.Model(&models.Transaction{}).Where("created_at >= ?", today).Count(&stats.TodayTransactions)

	c.JSON(http.StatusOK, stats)
}

func (h *ReportHandler) GetStockLevels(c *gin.Context) {
	var inventory []models.Inventory
	database.DB.Preload("Product.Category").Find(&inventory)

	var report []StockLevelReport
	for _, inv := range inventory {
		if inv.Product == nil {
			continue
		}
		status := "OK"
		if inv.Quantity <= 0 {
			status = "Out of Stock"
		} else if inv.Product.ReorderPoint > 0 && inv.Quantity <= inv.Product.ReorderPoint {
			status = "Low Stock"
		}

		reorderPoint := inv.Product.ReorderPoint
		if reorderPoint == 0 {
			reorderPoint = 10
		}

		catName := ""
		if inv.Product.Category != nil {
			catName = inv.Product.Category.Name
		}

		report = append(report, StockLevelReport{
			ProductID:    inv.ProductID,
			ProductName:  inv.Product.Name,
			SKU:          inv.Product.SKU,
			CategoryName: catName,
			Quantity:     inv.Quantity,
			ReorderPoint: reorderPoint,
			Status:       status,
		})
	}

	c.JSON(http.StatusOK, report)
}

func (h *ReportHandler) GetValuation(c *gin.Context) {
	var inventory []models.Inventory
	database.DB.Preload("Product").Find(&inventory)

	var report []ValuationReport
	var totalValue float64

	for _, inv := range inventory {
		if inv.Product == nil || inv.Product.CostPrice == 0 {
			continue
		}
		value := float64(inv.Quantity) * inv.Product.CostPrice
		totalValue += value

		report = append(report, ValuationReport{
			ProductID:   inv.ProductID,
			ProductName: inv.Product.Name,
			SKU:         inv.Product.SKU,
			Quantity:    inv.Quantity,
			UnitCost:    inv.Product.CostPrice,
			TotalValue:  value,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"items":       report,
		"total_value": totalValue,
	})
}

func (h *ReportHandler) GetLowStock(c *gin.Context) {
	var products []models.Product
	database.DB.Preload("Inventory").Where("is_active = ? AND reorder_point > 0", true).Find(&products)

	var report []map[string]interface{}
	for _, p := range products {
		if p.Inventory != nil && p.Inventory.Quantity <= p.ReorderPoint {
			report = append(report, map[string]interface{}{
				"product_id":    p.ID,
				"product_name":  p.Name,
				"sku":           p.SKU,
				"current_stock": p.Inventory.Quantity,
				"reorder_point": p.ReorderPoint,
				"needed":        p.ReorderPoint - p.Inventory.Quantity,
			})
		}
	}

	c.JSON(http.StatusOK, report)
}

func (h *ReportHandler) GetTurnover(c *gin.Context) {
	startDate := c.DefaultQuery("start_date", time.Now().AddDate(0, -1, 0).Format("2006-01-02"))
	endDate := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

	start, _ := time.Parse("2006-01-02", startDate)
	end, _ := time.Parse("2006-01-02", endDate)
	end = end.Add(24 * time.Hour)

	var transactions []models.Transaction
	database.DB.Where("created_at BETWEEN ? AND ?", start, end).Order("created_at DESC").Find(&transactions)

	type turnover struct {
		Date      string `json:"date"`
		TotalIn   int    `json:"total_in"`
		TotalOut  int    `json:"total_out"`
		NetChange int    `json:"net_change"`
	}

	dailyMap := make(map[string]*turnover)

	for _, t := range transactions {
		date := t.CreatedAt.Format("2006-01-02")
		if dailyMap[date] == nil {
			dailyMap[date] = &turnover{Date: date}
		}
		if t.Quantity > 0 {
			dailyMap[date].TotalIn += t.Quantity
		} else {
			dailyMap[date].TotalOut += -t.Quantity
		}
		dailyMap[date].NetChange += t.Quantity
	}

	var result []turnover
	for _, v := range dailyMap {
		result = append(result, *v)
	}

	c.JSON(http.StatusOK, result)
}

func (h *ReportHandler) ExportStockLevels(c *gin.Context) {
	var inventory []models.Inventory
	database.DB.Preload("Product.Category").Find(&inventory)

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=stock_levels.csv")

	w := csv.NewWriter(c.Writer)
	defer w.Flush()

	w.Write([]string{"SKU", "Product Name", "Category", "Quantity", "Reorder Point", "Status"})

	for _, inv := range inventory {
		if inv.Product == nil {
			continue
		}
		status := "OK"
		if inv.Quantity <= 0 {
			status = "Out of Stock"
		} else if inv.Product.ReorderPoint > 0 && inv.Quantity <= inv.Product.ReorderPoint {
			status = "Low Stock"
		}

		catName := ""
		if inv.Product.Category != nil {
			catName = inv.Product.Category.Name
		}

		reorderPoint := inv.Product.ReorderPoint
		if reorderPoint == 0 {
			reorderPoint = 10
		}

		w.Write([]string{
			inv.Product.SKU,
			inv.Product.Name,
			catName,
			strconv.Itoa(inv.Quantity),
			strconv.Itoa(reorderPoint),
			status,
		})
	}
}

func (h *ReportHandler) GetTransactions(c *gin.Context) {
	var transactions []models.Transaction
	query := database.DB.Preload("Product").Preload("User", "id, username").Order("created_at DESC")

	limit := c.DefaultQuery("limit", "100")
	l, _ := strconv.Atoi(limit)
	query = query.Limit(l)

	productID := c.Query("product_id")
	if productID != "" {
		id, _ := strconv.ParseUint(productID, 10, 32)
		query = query.Where("product_id = ?", id)
	}

	txType := c.Query("type")
	if txType != "" {
		query = query.Where("type = ?", txType)
	}

	if err := query.Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions"})
		return
	}

	var response []models.TransactionResponse
	for _, t := range transactions {
		prodName := ""
		prodSKU := ""
		if t.Product != nil {
			prodName = t.Product.Name
			prodSKU = t.Product.SKU
		}
		userName := ""
		if t.User != nil {
			userName = t.User.Username
		}
		response = append(response, models.TransactionResponse{
			ID:            t.ID,
			ProductID:     t.ProductID,
			ProductName:   prodName,
			ProductSKU:    prodSKU,
			Type:          t.Type,
			Quantity:      t.Quantity,
			ReferenceType: t.ReferenceType,
			Notes:         t.Notes,
			UserName:      userName,
			CreatedAt:     t.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, response)
}

func (h *ReportHandler) GetRecentTransactions(c *gin.Context) {
	var transactions []models.Transaction
	if err := database.DB.Preload("Product").Preload("User", "id, username").Order("created_at DESC").Limit(10).Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions"})
		return
	}

	type recentTx struct {
		ID          uint      `json:"id"`
		ProductName string    `json:"product_name"`
		Type        string    `json:"type"`
		Quantity    int       `json:"quantity"`
		UserName    string    `json:"user_name"`
		CreatedAt   time.Time `json:"created_at"`
	}

	var result []recentTx
	for _, t := range transactions {
		prodName := ""
		if t.Product != nil {
			prodName = t.Product.Name
		}
		userName := ""
		if t.User != nil {
			userName = t.User.Username
		}
		result = append(result, recentTx{
			ID:          t.ID,
			ProductName: prodName,
			Type:        t.Type,
			Quantity:    t.Quantity,
			UserName:    userName,
			CreatedAt:   t.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, result)
}

func (h *ReportHandler) GetCategoryBreakdown(c *gin.Context) {
	type categoryStat struct {
		CategoryName string  `json:"category_name"`
		ProductCount int     `json:"product_count"`
		TotalValue   float64 `json:"total_value"`
	}

	var results []categoryStat

	rows, _ := database.DB.Raw(`
		SELECT 
			COALESCE(c.name, 'Uncategorized') as category_name,
			COUNT(DISTINCT p.id) as product_count,
			COALESCE(SUM(i.quantity * p.cost_price), 0) as total_value
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		LEFT JOIN inventory i ON p.id = i.product_id
		WHERE p.is_active = 1
		GROUP BY c.id, c.name
		ORDER BY total_value DESC
	`).Rows()

	defer rows.Close()
	for rows.Next() {
		var stat categoryStat
		var catName *string
		var value float64
		rows.Scan(&catName, &stat.ProductCount, &value)
		if catName != nil {
			stat.CategoryName = *catName
		}
		stat.TotalValue = value
		results = append(results, stat)
	}

	c.JSON(http.StatusOK, results)
}
