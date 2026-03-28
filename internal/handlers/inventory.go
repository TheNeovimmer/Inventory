package handlers

import (
	"net/http"
	"strconv"
	"time"

	"inventory-ims/internal/database"
	"inventory-ims/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type InventoryHandler struct{}

func NewInventoryHandler() *InventoryHandler {
	return &InventoryHandler{}
}

type AdjustInventoryRequest struct {
	ProductID     uint   `json:"product_id" binding:"required"`
	Quantity      int    `json:"quantity" binding:"required"`
	Type          string `json:"type" binding:"required"` // in, out, adjustment
	Notes         string `json:"notes"`
	ReferenceID   *uint  `json:"reference_id"`
	ReferenceType string `json:"reference_type"`
}

type InventoryResponse struct {
	ProductID    uint    `json:"product_id"`
	ProductName  string  `json:"product_name"`
	ProductSKU   string  `json:"product_sku"`
	Quantity     int     `json:"quantity"`
	ReservedQty  int     `json:"reserved_quantity"`
	AvailableQty int     `json:"available_quantity"`
	ReorderPoint int     `json:"reorder_point"`
	UnitPrice    float64 `json:"unit_price"`
	CategoryName string  `json:"category_name"`
}

func (h *InventoryHandler) List(c *gin.Context) {
	var inventory []models.Inventory
	query := database.DB.Preload("Product.Category")

	productID := c.Query("product_id")
	if productID != "" {
		id, _ := strconv.ParseUint(productID, 10, 32)
		query = query.Where("product_id = ?", id)
	}

	lowStock := c.Query("low_stock")
	if lowStock == "true" {
		query = query.Where("quantity <= (SELECT COALESCE(reorder_point, 10) FROM products WHERE products.id = inventory.product_id)")
	}

	if err := query.Find(&inventory).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch inventory"})
		return
	}

	var response []InventoryResponse
	for _, inv := range inventory {
		if inv.Product == nil {
			continue
		}
		reorderPoint := inv.Product.ReorderPoint
		if reorderPoint == 0 {
			reorderPoint = 10
		}
		catName := ""
		if inv.Product.Category != nil {
			catName = inv.Product.Category.Name
		}
		response = append(response, InventoryResponse{
			ProductID:    inv.ProductID,
			ProductName:  inv.Product.Name,
			ProductSKU:   inv.Product.SKU,
			Quantity:     inv.Quantity,
			ReservedQty:  inv.ReservedQuantity,
			AvailableQty: inv.AvailableQuantity(),
			ReorderPoint: reorderPoint,
			UnitPrice:    inv.Product.UnitPrice,
			CategoryName: catName,
		})
	}

	c.JSON(http.StatusOK, response)
}

func (h *InventoryHandler) Adjust(c *gin.Context) {
	var req AdjustInventoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")

	var product models.Product
	if err := database.DB.First(&product, req.ProductID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	var inventory models.Inventory
	if err := database.DB.Where("product_id = ?", req.ProductID).First(&inventory).Error; err != nil {
		inventory = models.Inventory{
			ProductID: req.ProductID,
			Quantity:  0,
		}
		database.DB.Create(&inventory)
	}

	transactionType := ""
	switch req.Type {
	case "in":
		inventory.Quantity += req.Quantity
		transactionType = models.TransactionTypePurchase
	case "out":
		if inventory.Quantity < req.Quantity {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient stock"})
			return
		}
		inventory.Quantity -= req.Quantity
		transactionType = models.TransactionTypeSale
	case "adjustment":
		transactionType = models.TransactionTypeAdjustment
	default:
		inventory.Quantity = req.Quantity
		transactionType = models.TransactionTypeAdjustment
	}

	inventory.LastUpdated = time.Now()
	database.DB.Save(&inventory)

	transaction := models.Transaction{
		ProductID:     req.ProductID,
		Type:          transactionType,
		Quantity:      req.Quantity,
		ReferenceType: req.ReferenceType,
		ReferenceID:   req.ReferenceID,
		Notes:         req.Notes,
		UserID:        userID.(uint),
	}
	database.DB.Create(&transaction)

	c.JSON(http.StatusOK, inventory)
}

func (h *InventoryHandler) GetAlerts(c *gin.Context) {
	var products []models.Product
	database.DB.Preload("Inventory").Where("is_active = ? AND reorder_point > 0").Find(&products)

	var alerts []map[string]interface{}
	for _, p := range products {
		if p.Inventory != nil && p.Inventory.Quantity <= p.ReorderPoint {
			alerts = append(alerts, map[string]interface{}{
				"product_id":    p.ID,
				"product_name":  p.Name,
				"product_sku":   p.SKU,
				"quantity":      p.Inventory.Quantity,
				"reorder_point": p.ReorderPoint,
			})
		}
	}

	c.JSON(http.StatusOK, alerts)
}

func (h *InventoryHandler) GetHistory(c *gin.Context) {
	productID := c.Query("product_id")
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "product_id required"})
		return
	}

	id, _ := strconv.ParseUint(productID, 10, 32)

	var transactions []models.Transaction
	if err := database.DB.Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, username")
	}).Where("product_id = ?", id).Order("created_at DESC").Limit(100).Find(&transactions).Error; err != nil {
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

func init() {
	_ = database.DB
}
