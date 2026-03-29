package handlers

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"inventory-ims/internal/database"
	"inventory-ims/internal/models"

	"github.com/gin-gonic/gin"
)

type ImportExportHandler struct{}

func NewImportExportHandler() *ImportExportHandler {
	return &ImportExportHandler{}
}

type ImportResult struct {
	Success   bool     `json:"success"`
	TotalRows int      `json:"total_rows"`
	Imported  int      `json:"imported"`
	Errors    []string `json:"errors"`
}

func (h *ImportExportHandler) ImportProducts(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to open file"})
		return
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read CSV"})
		return
	}

	if len(records) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CSV file is empty or has no data rows"})
		return
	}

	userID, _ := c.Get("user_id")

	var skuIdx, nameIdx, descIdx, catIdx, priceIdx, costIdx, reorderIdx int
	for i, h := range records[0] {
		h = strings.ToLower(strings.TrimSpace(h))
		switch h {
		case "sku":
			skuIdx = i
		case "name", "product_name":
			nameIdx = i
		case "description", "desc":
			descIdx = i
		case "category", "category_id", "category_name":
			catIdx = i
		case "price", "unit_price":
			priceIdx = i
		case "cost", "cost_price":
			costIdx = i
		case "reorder", "reorder_point", "min_stock":
			reorderIdx = i
		}
	}

	result := ImportResult{
		Success:   true,
		TotalRows: len(records) - 1,
		Errors:    []string{},
	}

	tx := database.DB.Begin()

	for i := 1; i < len(records); i++ {
		row := records[i]
		if len(row) < 2 {
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: Not enough columns", i+1))
			continue
		}

		sku := strings.TrimSpace(row[skuIdx])
		name := strings.TrimSpace(row[nameIdx])

		if sku == "" || name == "" {
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: SKU and Name are required", i+1))
			continue
		}

		var existing models.Product
		if err := tx.Where("sku = ?", sku).First(&existing).Error; err == nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: SKU '%s' already exists", i+1, sku))
			continue
		}

		unitPrice := 0.0
		if priceIdx > 0 && priceIdx < len(row) {
			if p, err := strconv.ParseFloat(row[priceIdx], 64); err == nil {
				unitPrice = p
			}
		}

		costPrice := 0.0
		if costIdx > 0 && costIdx < len(row) {
			if p, err := strconv.ParseFloat(row[costIdx], 64); err == nil {
				costPrice = p
			}
		}

		reorderPoint := 10
		if reorderIdx > 0 && reorderIdx < len(row) {
			if p, err := strconv.Atoi(row[reorderIdx]); err == nil {
				reorderPoint = p
			}
		}

		var categoryID *uint
		if catIdx > 0 && catIdx < len(row) && strings.TrimSpace(row[catIdx]) != "" {
			var cat models.Category
			if err := tx.Where("name = ?", strings.TrimSpace(row[catIdx])).First(&cat).Error; err == nil {
				categoryID = &cat.ID
			}
		}

		product := models.Product{
			SKU:          sku,
			Name:         name,
			Description:  getOrEmpty(row, descIdx),
			CategoryID:   categoryID,
			UnitPrice:    unitPrice,
			CostPrice:    costPrice,
			ReorderPoint: reorderPoint,
			IsActive:     true,
		}

		if err := tx.Create(&product).Error; err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: Failed to create product - %s", i+1, err.Error()))
			continue
		}

		tx.Create(&models.Inventory{
			ProductID:   product.ID,
			WarehouseID: 1,
			Quantity:    0,
		})

		result.Imported++
	}

	if result.Imported > 0 {
		tx.Commit()
	} else {
		tx.Rollback()
	}

	if len(result.Errors) > 0 {
		result.Success = false
	}

	logAudit(c, "import", 0, models.AuditActionImport, "", fmt.Sprintf("Imported %d products", result.Imported), userID.(uint))

	c.JSON(http.StatusOK, result)
}

func (h *ImportExportHandler) ImportInventory(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to open file"})
		return
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read CSV"})
		return
	}

	if len(records) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CSV file is empty or has no data rows"})
		return
	}

	userID, _ := c.Get("user_id")

	var skuIdx, qtyIdx, whIdx int
	for i, h := range records[0] {
		h = strings.ToLower(strings.TrimSpace(h))
		switch h {
		case "sku":
			skuIdx = i
		case "quantity", "qty", "stock":
			qtyIdx = i
		case "warehouse", "warehouse_id", "warehouse_code":
			whIdx = i
		}
	}

	result := ImportResult{
		Success:   true,
		TotalRows: len(records) - 1,
		Errors:    []string{},
	}

	for i := 1; i < len(records); i++ {
		row := records[i]
		if len(row) < 2 {
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: Not enough columns", i+1))
			continue
		}

		sku := strings.TrimSpace(row[skuIdx])
		if sku == "" {
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: SKU is required", i+1))
			continue
		}

		qty := 0
		if qtyIdx > 0 && qtyIdx < len(row) {
			if q, err := strconv.Atoi(row[qtyIdx]); err == nil {
				qty = q
			}
		}

		warehouseID := uint(1)
		if whIdx > 0 && whIdx < len(row) && strings.TrimSpace(row[whIdx]) != "" {
			var wh models.Warehouse
			if err := database.DB.Where("code = ? OR name = ?", strings.TrimSpace(row[whIdx]), strings.TrimSpace(row[whIdx])).First(&wh).Error; err == nil {
				warehouseID = wh.ID
			}
		}

		var product models.Product
		if err := database.DB.Where("sku = ?", sku).First(&product).Error; err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: Product with SKU '%s' not found", i+1, sku))
			continue
		}

		var inv models.Inventory
		if err := database.DB.Where("product_id = ? AND warehouse_id = ?", product.ID, warehouseID).First(&inv).Error; err != nil {
			inv = models.Inventory{
				ProductID:   product.ID,
				WarehouseID: warehouseID,
				Quantity:    0,
			}
			database.DB.Create(&inv)
		}

		oldQty := inv.Quantity
		inv.Quantity = qty
		inv.LastUpdated = time.Now()
		database.DB.Save(&inv)

		database.DB.Create(&models.Transaction{
			ProductID: product.ID,
			Type:      models.TransactionTypeAdjustment,
			Quantity:  qty - oldQty,
			Notes:     "CSV import",
			UserID:    userID.(uint),
		})

		result.Imported++
	}

	if len(result.Errors) > 0 {
		result.Success = false
	}

	logAudit(c, "import", 0, models.AuditActionImport, "", fmt.Sprintf("Imported inventory for %d products", result.Imported), userID.(uint))

	c.JSON(http.StatusOK, result)
}

func (h *ImportExportHandler) ExportProducts(c *gin.Context) {
	var products []models.Product
	database.DB.Preload("Category").Where("is_active = ?", true).Order("created_at DESC").Find(&products)

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=products.csv")

	w := csv.NewWriter(c.Writer)
	defer w.Flush()

	w.Write([]string{"SKU", "Name", "Description", "Category", "Unit Price", "Cost Price", "Reorder Point", "Has BOM", "Created At"})

	for _, p := range products {
		catName := ""
		if p.Category != nil {
			catName = p.Category.Name
		}
		w.Write([]string{
			p.SKU,
			p.Name,
			p.Description,
			catName,
			fmt.Sprintf("%.2f", p.UnitPrice),
			fmt.Sprintf("%.2f", p.CostPrice),
			strconv.Itoa(p.ReorderPoint),
			strconv.FormatBool(p.HasBOM),
			p.CreatedAt.Format("2006-01-02"),
		})
	}
}

func (h *ImportExportHandler) ExportInventory(c *gin.Context) {
	var inventory []models.Inventory
	database.DB.Preload("Product").Preload("Warehouse").Find(&inventory)

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=inventory.csv")

	w := csv.NewWriter(c.Writer)
	defer w.Flush()

	w.Write([]string{"SKU", "Product Name", "Warehouse", "Quantity", "Reserved", "Available", "Last Updated"})

	for _, inv := range inventory {
		if inv.Product == nil {
			continue
		}
		whName := "Default"
		if inv.Warehouse != nil {
			whName = inv.Warehouse.Name
		}
		w.Write([]string{
			inv.Product.SKU,
			inv.Product.Name,
			whName,
			strconv.Itoa(inv.Quantity),
			strconv.Itoa(inv.ReservedQuantity),
			strconv.Itoa(inv.AvailableQuantity()),
			inv.LastUpdated.Format("2006-01-02 15:04"),
		})
	}
}

func (h *ImportExportHandler) ExportTransactions(c *gin.Context) {
	var transactions []models.Transaction
	database.DB.Preload("Product").Preload("User").Order("created_at DESC").Limit(1000).Find(&transactions)

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=transactions.csv")

	w := csv.NewWriter(c.Writer)
	defer w.Flush()

	w.Write([]string{"ID", "Date", "Product SKU", "Product Name", "Type", "Quantity", "User", "Notes"})

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
		w.Write([]string{
			strconv.Itoa(int(t.ID)),
			t.CreatedAt.Format("2006-01-02 15:04"),
			prodSKU,
			prodName,
			t.Type,
			strconv.Itoa(t.Quantity),
			userName,
			t.Notes,
		})
	}
}

func (h *ImportExportHandler) ExportTemplate(c *gin.Context) {
	templateType := c.DefaultQuery("type", "products")

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s_template.csv", templateType))

	w := csv.NewWriter(c.Writer)
	defer w.Flush()

	switch templateType {
	case "products":
		w.Write([]string{"SKU", "Name", "Description", "Category", "Unit Price", "Cost Price", "Reorder Point"})
		w.Write([]string{"SKU-001", "Product Name", "Product Description", "Electronics", "99.99", "50.00", "10"})
	case "inventory":
		w.Write([]string{"SKU", "Quantity", "Warehouse"})
		w.Write([]string{"SKU-001", "100", "MAIN"})
	}

	userID, _ := c.Get("user_id")
	logAudit(c, "export", 0, models.AuditActionExport, "", fmt.Sprintf("Downloaded %s template", templateType), userID.(uint))
}

func getOrEmpty(row []string, idx int) string {
	if idx > 0 && idx < len(row) {
		return strings.TrimSpace(row[idx])
	}
	return ""
}

func init() {
	_ = json.Marshal
}
