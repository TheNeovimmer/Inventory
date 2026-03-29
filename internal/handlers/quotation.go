package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"inventory-ims/internal/database"
	"inventory-ims/internal/models"

	"github.com/gin-gonic/gin"
)

type QuotationHandler struct{}

func NewQuotationHandler() *QuotationHandler {
	return &QuotationHandler{}
}

type CreateQuotationRequest struct {
	CustomerID     *uint                  `json:"customer_id"`
	WarehouseID    uint                   `json:"warehouse_id"`
	Items          []QuotationItemRequest `json:"items" binding:"required,min=1"`
	TaxRate        float64                `json:"tax_rate"`
	DiscountAmount float64                `json:"discount_amount"`
	ValidUntil     *time.Time             `json:"valid_until"`
	Notes          string                 `json:"notes"`
	Terms          string                 `json:"terms"`
}

type QuotationItemRequest struct {
	ProductID uint    `json:"product_id" binding:"required"`
	Quantity  int     `json:"quantity" binding:"required,min=1"`
	UnitPrice float64 `json:"unit_price" binding:"required"`
	Discount  float64 `json:"discount"`
	TaxRate   float64 `json:"tax_rate"`
}

func (h *QuotationHandler) List(c *gin.Context) {
	var quotations []models.Quotation
	query := database.DB.Preload("Customer").Preload("Creator").Preload("Items")

	status := c.Query("status")
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("created_at DESC").Find(&quotations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch quotations"})
		return
	}

	c.JSON(http.StatusOK, quotations)
}

func (h *QuotationHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid quotation ID"})
		return
	}

	var quotation models.Quotation
	if err := database.DB.Preload("Customer").Preload("Creator").Preload("Items").First(&quotation, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Quotation not found"})
		return
	}

	c.JSON(http.StatusOK, quotation)
}

func (h *QuotationHandler) Create(c *gin.Context) {
	var req CreateQuotationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")

	quotationNumber := generateQuotationNumber()
	warehouseID := req.WarehouseID
	if warehouseID == 0 {
		warehouseID = 1
	}

	subtotal := 0.0
	taxAmount := 0.0

	quotation := models.Quotation{
		QuotationNumber: quotationNumber,
		CustomerID:      req.CustomerID,
		WarehouseID:     warehouseID,
		TaxRate:         req.TaxRate,
		DiscountAmount:  req.DiscountAmount,
		Status:          models.QuotationStatusDraft,
		ValidUntil:      req.ValidUntil,
		Notes:           req.Notes,
		Terms:           req.Terms,
		CreatedBy:       userID.(uint),
	}

	for _, item := range req.Items {
		var product models.Product
		if err := database.DB.First(&product, item.ProductID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Product not found"})
			return
		}

		itemTotal := float64(item.Quantity) * item.UnitPrice
		itemDiscount := item.Discount
		itemTax := itemTotal * (item.TaxRate / 100)

		quotationItem := models.QuotationItem{
			ProductID:   item.ProductID,
			ProductName: product.Name,
			ProductSKU:  product.SKU,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			Discount:    itemDiscount,
			TaxRate:     item.TaxRate,
			TaxAmount:   itemTax,
			Total:       itemTotal - itemDiscount + itemTax,
		}
		quotation.Items = append(quotation.Items, quotationItem)

		subtotal += itemTotal - itemDiscount
		taxAmount += itemTax
	}

	quotation.Subtotal = subtotal
	quotation.TaxAmount = taxAmount
	quotation.Total = subtotal + taxAmount - req.DiscountAmount

	if err := database.DB.Create(&quotation).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create quotation"})
		return
	}

	logAudit(c, "quotation", quotation.ID, models.AuditActionCreate, "", "Created quotation: "+quotationNumber, userID.(uint))

	c.JSON(http.StatusCreated, quotation)
}

func (h *QuotationHandler) UpdateStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid quotation ID"})
		return
	}

	status := c.Query("status")
	if status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status required"})
		return
	}

	var quotation models.Quotation
	if err := database.DB.First(&quotation, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Quotation not found"})
		return
	}

	if err := database.DB.Model(&quotation).Update("status", status).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update quotation"})
		return
	}

	c.JSON(http.StatusOK, quotation)
}

func (h *QuotationHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid quotation ID"})
		return
	}

	userID, _ := c.Get("user_id")

	var quotation models.Quotation
	if err := database.DB.First(&quotation, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Quotation not found"})
		return
	}

	if err := database.DB.Delete(&quotation).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete quotation"})
		return
	}

	logAudit(c, "quotation", quotation.ID, models.AuditActionDelete, "", "Deleted quotation: "+quotation.QuotationNumber, userID.(uint))

	c.JSON(http.StatusOK, gin.H{"message": "Quotation deleted"})
}

func generateQuotationNumber() string {
	return fmt.Sprintf("QUO-%s", time.Now().Format("20060102150405"))
}
