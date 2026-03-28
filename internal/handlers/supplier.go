package handlers

import (
	"net/http"
	"strconv"
	"time"

	"inventory-ims/internal/database"
	"inventory-ims/internal/models"

	"github.com/gin-gonic/gin"
)

type SupplierHandler struct{}

func NewSupplierHandler() *SupplierHandler {
	return &SupplierHandler{}
}

type CreateSupplierRequest struct {
	Name        string `json:"name" binding:"required"`
	ContactName string `json:"contact_name"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Address     string `json:"address"`
	Notes       string `json:"notes"`
}

func (h *SupplierHandler) List(c *gin.Context) {
	var suppliers []models.Supplier
	if err := database.DB.Where("is_active = ?", true).Order("name").Find(&suppliers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch suppliers"})
		return
	}
	c.JSON(http.StatusOK, suppliers)
}

func (h *SupplierHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid supplier ID"})
		return
	}

	var supplier models.Supplier
	if err := database.DB.First(&supplier, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Supplier not found"})
		return
	}

	c.JSON(http.StatusOK, supplier)
}

func (h *SupplierHandler) Create(c *gin.Context) {
	var req CreateSupplierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	supplier := models.Supplier{
		Name:        req.Name,
		ContactName: req.ContactName,
		Email:       req.Email,
		Phone:       req.Phone,
		Address:     req.Address,
		Notes:       req.Notes,
		IsActive:    true,
	}

	if err := database.DB.Create(&supplier).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create supplier"})
		return
	}

	c.JSON(http.StatusCreated, supplier)
}

func (h *SupplierHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid supplier ID"})
		return
	}

	var supplier models.Supplier
	if err := database.DB.First(&supplier, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Supplier not found"})
		return
	}

	var req CreateSupplierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := database.DB.Model(&supplier).Updates(map[string]interface{}{
		"name":         req.Name,
		"contact_name": req.ContactName,
		"email":        req.Email,
		"phone":        req.Phone,
		"address":      req.Address,
		"notes":        req.Notes,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update supplier"})
		return
	}

	c.JSON(http.StatusOK, supplier)
}

func (h *SupplierHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid supplier ID"})
		return
	}

	database.DB.Model(&models.Supplier{}).Where("id = ?", id).Update("is_active", false)
	c.JSON(http.StatusOK, gin.H{"message": "Supplier deactivated"})
}

type PurchaseOrderHandler struct{}

func NewPurchaseOrderHandler() *PurchaseOrderHandler {
	return &PurchaseOrderHandler{}
}

type CreatePORequest struct {
	SupplierID   uint                  `json:"supplier_id" binding:"required"`
	ExpectedDate *time.Time            `json:"expected_date"`
	Notes        string                `json:"notes"`
	Items        []CreatePOItemRequest `json:"items"`
}

type CreatePOItemRequest struct {
	ProductID uint    `json:"product_id" binding:"required"`
	Quantity  int     `json:"quantity" binding:"required"`
	UnitCost  float64 `json:"unit_cost" binding:"required"`
}

func (h *PurchaseOrderHandler) List(c *gin.Context) {
	var pos []models.PurchaseOrder
	query := database.DB.Preload("Supplier").Preload("Creator").Order("created_at DESC")

	status := c.Query("status")
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Find(&pos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch purchase orders"})
		return
	}

	var response []models.PurchaseOrderResponse
	for _, po := range pos {
		supplierName := ""
		if po.Supplier != nil {
			supplierName = po.Supplier.Name
		}
		creatorName := ""
		if po.Creator != nil {
			creatorName = po.Creator.Username
		}
		response = append(response, models.PurchaseOrderResponse{
			ID:           po.ID,
			SupplierID:   po.SupplierID,
			SupplierName: supplierName,
			Status:       po.Status,
			OrderDate:    po.OrderDate,
			ExpectedDate: po.ExpectedDate,
			ReceivedDate: po.ReceivedDate,
			Total:        po.Total,
			Notes:        po.Notes,
			CreatedBy:    po.CreatedBy,
			CreatorName:  creatorName,
			CreatedAt:    po.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, response)
}

func (h *PurchaseOrderHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid PO ID"})
		return
	}

	var po models.PurchaseOrder
	if err := database.DB.Preload("Supplier").Preload("Items.Product").First(&po, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Purchase order not found"})
		return
	}

	c.JSON(http.StatusOK, po)
}

func (h *PurchaseOrderHandler) Create(c *gin.Context) {
	var req CreatePORequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")

	var total float64
	for _, item := range req.Items {
		total += item.UnitCost * float64(item.Quantity)
	}

	po := models.PurchaseOrder{
		SupplierID:   req.SupplierID,
		Status:       models.POStatusPending,
		OrderDate:    time.Now(),
		ExpectedDate: req.ExpectedDate,
		Total:        total,
		Notes:        req.Notes,
		CreatedBy:    userID.(uint),
	}

	if err := database.DB.Create(&po).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create purchase order"})
		return
	}

	for _, item := range req.Items {
		poItem := models.PurchaseOrderItem{
			POID:      po.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			UnitCost:  item.UnitCost,
		}
		database.DB.Create(&poItem)
	}

	c.JSON(http.StatusCreated, po)
}

func (h *PurchaseOrderHandler) UpdateStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid PO ID"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var po models.PurchaseOrder
	if err := database.DB.First(&po, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Purchase order not found"})
		return
	}

	po.Status = req.Status
	if req.Status == models.POStatusReceived {
		now := time.Now()
		po.ReceivedDate = &now
	}

	database.DB.Save(&po)
	c.JSON(http.StatusOK, po)
}

func (h *PurchaseOrderHandler) Receive(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid PO ID"})
		return
	}

	userID, _ := c.Get("user_id")

	var po models.PurchaseOrder
	if err := database.DB.Preload("Items").First(&po, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Purchase order not found"})
		return
	}

	if po.Status == models.POStatusReceived {
		c.JSON(http.StatusBadRequest, gin.H{"error": "PO already received"})
		return
	}

	now := time.Now()

	for _, item := range po.Items {
		var inv models.Inventory
		if err := database.DB.Where("product_id = ?", item.ProductID).First(&inv).Error; err != nil {
			inv = models.Inventory{ProductID: item.ProductID, Quantity: 0}
			database.DB.Create(&inv)
		}

		receiveQty := item.Quantity - item.ReceivedQty
		inv.Quantity += receiveQty
		inv.LastUpdated = now
		database.DB.Save(&inv)

		item.ReceivedQty = item.Quantity
		database.DB.Save(&item)

		transaction := models.Transaction{
			ProductID:     item.ProductID,
			Type:          models.TransactionTypePurchase,
			Quantity:      receiveQty,
			ReferenceType: "purchase_order",
			ReferenceID:   &po.ID,
			Notes:         "PO Received",
			UserID:        userID.(uint),
		}
		database.DB.Create(&transaction)
	}

	po.Status = models.POStatusReceived
	po.ReceivedDate = &now
	database.DB.Save(&po)

	c.JSON(http.StatusOK, po)
}
