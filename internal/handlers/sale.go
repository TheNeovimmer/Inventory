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

type SaleHandler struct{}

func NewSaleHandler() *SaleHandler {
	return &SaleHandler{}
}

type CreateSaleRequest struct {
	CustomerID     *uint             `json:"customer_id"`
	WarehouseID    uint              `json:"warehouse_id"`
	Items          []SaleItemRequest `json:"items" binding:"required,min=1"`
	TaxRate        float64           `json:"tax_rate"`
	DiscountAmount float64           `json:"discount_amount"`
	ShippingAmount float64           `json:"shipping_amount"`
	PaymentMethod  string            `json:"payment_method"`
	PaidAmount     float64           `json:"paid_amount"`
	Notes          string            `json:"notes"`
	DueDate        *time.Time        `json:"due_date"`
}

type SaleItemRequest struct {
	ProductID uint    `json:"product_id" binding:"required"`
	Quantity  int     `json:"quantity" binding:"required,min=1"`
	UnitPrice float64 `json:"unit_price" binding:"required"`
	Discount  float64 `json:"discount"`
	TaxRate   float64 `json:"tax_rate"`
}

type AddPaymentRequest struct {
	PaymentMethod string  `json:"payment_method" binding:"required"`
	Amount        float64 `json:"amount" binding:"required,min=1"`
	Reference     string  `json:"reference"`
	Notes         string  `json:"notes"`
}

func (h *SaleHandler) List(c *gin.Context) {
	var sales []models.Sale
	query := database.DB.Preload("Customer").Preload("Creator").Preload("Items")

	status := c.Query("status")
	if status != "" {
		query = query.Where("status = ?", status)
	}

	customerID := c.Query("customer_id")
	if customerID != "" {
		id, _ := strconv.ParseUint(customerID, 10, 32)
		query = query.Where("customer_id = ?", id)
	}

	startDate := c.Query("start_date")
	if startDate != "" {
		query = query.Where("created_at >= ?", startDate)
	}

	endDate := c.Query("end_date")
	if endDate != "" {
		query = query.Where("created_at <= ?", endDate)
	}

	if err := query.Order("created_at DESC").Find(&sales).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch sales"})
		return
	}

	var response []models.SaleResponse
	for _, sale := range sales {
		customerName := ""
		if sale.Customer != nil {
			customerName = sale.Customer.Name
		}

		creatorName := ""
		if sale.Creator != nil {
			creatorName = sale.Creator.Username
		}

		items := make([]models.SaleItemResponse, len(sale.Items))
		for i, item := range sale.Items {
			items[i] = models.SaleItemResponse{
				ID:          item.ID,
				ProductID:   item.ProductID,
				ProductName: item.ProductName,
				ProductSKU:  item.ProductSKU,
				Quantity:    item.Quantity,
				UnitPrice:   item.UnitPrice,
				Discount:    item.Discount,
				TaxRate:     item.TaxRate,
				TaxAmount:   item.TaxAmount,
				Total:       item.Total,
			}
		}

		payments := make([]models.SalePaymentResponse, len(sale.Payments))
		for i, p := range sale.Payments {
			payments[i] = models.SalePaymentResponse{
				ID:            p.ID,
				PaymentMethod: p.PaymentMethod,
				Amount:        p.Amount,
				Reference:     p.Reference,
				Notes:         p.Notes,
				CreatedAt:     p.CreatedAt.Format("2006-01-02 15:04"),
			}
		}

		response = append(response, models.SaleResponse{
			ID:             sale.ID,
			InvoiceNumber:  sale.InvoiceNumber,
			CustomerID:     sale.CustomerID,
			CustomerName:   customerName,
			Subtotal:       sale.Subtotal,
			TaxAmount:      sale.TaxAmount,
			DiscountAmount: sale.DiscountAmount,
			ShippingAmount: sale.ShippingAmount,
			Total:          sale.Total,
			Status:         sale.Status,
			PaymentStatus:  sale.PaymentStatus,
			PaymentMethod:  sale.PaymentMethod,
			PaidAmount:     sale.PaidAmount,
			DueAmount:      sale.DueAmount,
			Notes:          sale.Notes,
			CreatedBy:      sale.CreatedBy,
			CreatorName:    creatorName,
			SaleDate:       sale.CreatedAt.Format("2006-01-02"),
			Items:          items,
			Payments:       payments,
			CreatedAt:      sale.CreatedAt.Format("2006-01-02 15:04"),
		})
	}

	c.JSON(http.StatusOK, response)
}

func (h *SaleHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sale ID"})
		return
	}

	var sale models.Sale
	if err := database.DB.Preload("Customer").Preload("Creator").Preload("Items").Preload("Payments").First(&sale, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sale not found"})
		return
	}

	c.JSON(http.StatusOK, sale)
}

func (h *SaleHandler) Create(c *gin.Context) {
	var req CreateSaleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")

	invoiceNumber := generateInvoiceNumber()
	warehouseID := req.WarehouseID
	if warehouseID == 0 {
		warehouseID = 1
	}

	subtotal := 0.0
	taxAmount := 0.0

	tx := database.DB.Begin()

	sale := models.Sale{
		InvoiceNumber:  invoiceNumber,
		CustomerID:     req.CustomerID,
		WarehouseID:    warehouseID,
		TaxRate:        req.TaxRate,
		DiscountAmount: req.DiscountAmount,
		ShippingAmount: req.ShippingAmount,
		Status:         models.SaleStatusCompleted,
		PaymentMethod:  req.PaymentMethod,
		Notes:          req.Notes,
		DueDate:        req.DueDate,
		CreatedBy:      userID.(uint),
		SaleDate:       time.Now(),
	}

	for _, item := range req.Items {
		var product models.Product
		if err := tx.First(&product, item.ProductID).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Product not found"})
			return
		}

		itemTotal := float64(item.Quantity) * item.UnitPrice
		itemDiscount := item.Discount
		itemTax := itemTotal * (item.TaxRate / 100)

		saleItem := models.SaleItem{
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
		sale.Items = append(sale.Items, saleItem)

		subtotal += itemTotal - itemDiscount
		taxAmount += itemTax

		var inv models.Inventory
		if err := tx.Where("product_id = ? AND warehouse_id = ?", item.ProductID, warehouseID).First(&inv).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Inventory not found for product"})
			return
		}

		if inv.Quantity < item.Quantity {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient stock for product: " + product.Name})
			return
		}

		inv.Quantity -= item.Quantity
		inv.LastUpdated = time.Now()
		tx.Save(&inv)

		tx.Create(&models.Transaction{
			ProductID: item.ProductID,
			Type:      models.TransactionTypeSale,
			Quantity:  -item.Quantity,
			Notes:     "Sale: " + invoiceNumber,
			UserID:    userID.(uint),
		})
	}

	sale.Subtotal = subtotal
	sale.TaxAmount = taxAmount
	sale.Total = subtotal + taxAmount + req.ShippingAmount - req.DiscountAmount

	if req.PaidAmount > 0 {
		sale.PaidAmount = req.PaidAmount
		sale.DueAmount = sale.Total - req.PaidAmount
		if sale.DueAmount <= 0 {
			sale.PaymentStatus = models.PaymentStatusPaid
		} else {
			sale.PaymentStatus = models.PaymentStatusPartial
		}
	} else {
		sale.PaidAmount = sale.Total
		sale.DueAmount = 0
		sale.PaymentStatus = models.PaymentStatusPaid
	}

	if err := tx.Create(&sale).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create sale"})
		return
	}

	if req.PaidAmount > 0 {
		payment := models.SalePayment{
			SaleID:        sale.ID,
			PaymentMethod: req.PaymentMethod,
			Amount:        req.PaidAmount,
		}
		tx.Create(&payment)
	}

	tx.Commit()

	logAudit(c, "sale", sale.ID, models.AuditActionCreate, "", "Created sale: "+invoiceNumber, userID.(uint))

	c.JSON(http.StatusCreated, sale)
}

func (h *SaleHandler) AddPayment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sale ID"})
		return
	}

	var req AddPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var sale models.Sale
	if err := database.DB.First(&sale, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sale not found"})
		return
	}

	payment := models.SalePayment{
		SaleID:        sale.ID,
		PaymentMethod: req.PaymentMethod,
		Amount:        req.Amount,
		Reference:     req.Reference,
		Notes:         req.Notes,
	}

	database.DB.Create(&payment)

	sale.PaidAmount += req.Amount
	sale.DueAmount = sale.Total - sale.PaidAmount

	if sale.DueAmount <= 0 {
		sale.PaymentStatus = models.PaymentStatusPaid
	} else {
		sale.PaymentStatus = models.PaymentStatusPartial
	}

	database.DB.Save(&sale)

	c.JSON(http.StatusOK, gin.H{"message": "Payment added", "sale": sale})
}

func (h *SaleHandler) Cancel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sale ID"})
		return
	}

	userID, _ := c.Get("user_id")

	var sale models.Sale
	if err := database.DB.Preload("Items").First(&sale, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sale not found"})
		return
	}

	if sale.Status == models.SaleStatusCancelled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Sale already cancelled"})
		return
	}

	tx := database.DB.Begin()

	for _, item := range sale.Items {
		var inv models.Inventory
		if err := tx.Where("product_id = ? AND warehouse_id = ?", item.ProductID, sale.WarehouseID).First(&inv).Error; err == nil {
			inv.Quantity += item.Quantity
			inv.LastUpdated = time.Now()
			tx.Save(&inv)
		}

		tx.Create(&models.Transaction{
			ProductID: item.ProductID,
			Type:      models.TransactionTypeReturn,
			Quantity:  item.Quantity,
			Notes:     "Sale cancelled: " + sale.InvoiceNumber,
			UserID:    userID.(uint),
		})
	}

	if err := tx.Model(&sale).Update("status", models.SaleStatusCancelled).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel sale"})
		return
	}

	tx.Commit()

	logAudit(c, "sale", sale.ID, models.AuditActionUpdate, sale.Status, models.SaleStatusCancelled, userID.(uint))

	c.JSON(http.StatusOK, gin.H{"message": "Sale cancelled"})
}

func generateInvoiceNumber() string {
	return fmt.Sprintf("INV-%s", time.Now().Format("20060102150405"))
}
