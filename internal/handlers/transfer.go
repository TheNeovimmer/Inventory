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

type TransferHandler struct{}

func NewTransferHandler() *TransferHandler {
	return &TransferHandler{}
}

type CreateTransferRequest struct {
	FromWarehouseID uint                  `json:"from_warehouse_id" binding:"required"`
	ToWarehouseID   uint                  `json:"to_warehouse_id" binding:"required"`
	Notes           string                `json:"notes"`
	Items           []TransferItemRequest `json:"items" binding:"required,min=1"`
}

type TransferItemRequest struct {
	ProductID uint `json:"product_id" binding:"required"`
	Quantity  int  `json:"quantity" binding:"required,min=1"`
}

func (h *TransferHandler) List(c *gin.Context) {
	var transfers []models.StockTransfer
	query := database.DB.Preload("FromWarehouse").Preload("ToWarehouse").Preload("Creator").Preload("Items.Product")

	status := c.Query("status")
	if status != "" {
		query = query.Where("status = ?", status)
	}

	fromWH := c.Query("from_warehouse_id")
	if fromWH != "" {
		id, _ := strconv.ParseUint(fromWH, 10, 32)
		query = query.Where("from_warehouse_id = ?", id)
	}

	toWH := c.Query("to_warehouse_id")
	if toWH != "" {
		id, _ := strconv.ParseUint(toWH, 10, 32)
		query = query.Where("to_warehouse_id = ?", id)
	}

	if err := query.Order("created_at DESC").Find(&transfers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transfers"})
		return
	}

	var response []models.TransferResponse
	for _, t := range transfers {
		creatorName := ""
		if t.Creator != nil {
			creatorName = t.Creator.Username
		}

		items := make([]models.TransferItemResponse, len(t.Items))
		for i, item := range t.Items {
			prodName := ""
			prodSKU := ""
			if item.Product != nil {
				prodName = item.Product.Name
				prodSKU = item.Product.SKU
			}
			items[i] = models.TransferItemResponse{
				ID:          item.ID,
				ProductID:   item.ProductID,
				ProductName: prodName,
				ProductSKU:  prodSKU,
				Quantity:    item.Quantity,
				ReceivedQty: item.ReceivedQty,
			}
		}

		response = append(response, models.TransferResponse{
			ID:             t.ID,
			TransferNumber: t.TransferNumber,
			FromWarehouse: models.WarehouseResponse{
				ID:        t.FromWarehouse.ID,
				Name:      t.FromWarehouse.Name,
				Code:      t.FromWarehouse.Code,
				Location:  t.FromWarehouse.Location,
				IsDefault: t.FromWarehouse.IsDefault,
				IsActive:  t.FromWarehouse.IsActive,
			},
			ToWarehouse: models.WarehouseResponse{
				ID:        t.ToWarehouse.ID,
				Name:      t.ToWarehouse.Name,
				Code:      t.ToWarehouse.Code,
				Location:  t.ToWarehouse.Location,
				IsDefault: t.ToWarehouse.IsDefault,
				IsActive:  t.ToWarehouse.IsActive,
			},
			Status:      t.Status,
			Notes:       t.Notes,
			CreatedBy:   t.CreatedBy,
			CreatorName: creatorName,
			ApprovedBy:  t.ApprovedBy,
			ApprovedAt:  t.ApprovedAt,
			CompletedBy: t.CompletedBy,
			CompletedAt: t.CompletedAt,
			Items:       items,
			CreatedAt:   t.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, response)
}

func (h *TransferHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transfer ID"})
		return
	}

	var transfer models.StockTransfer
	if err := database.DB.Preload("FromWarehouse").Preload("ToWarehouse").Preload("Creator").Preload("Items.Product").First(&transfer, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transfer not found"})
		return
	}

	creatorName := ""
	if transfer.Creator != nil {
		creatorName = transfer.Creator.Username
	}

	items := make([]models.TransferItemResponse, len(transfer.Items))
	for i, item := range transfer.Items {
		prodName := ""
		prodSKU := ""
		if item.Product != nil {
			prodName = item.Product.Name
			prodSKU = item.Product.SKU
		}
		items[i] = models.TransferItemResponse{
			ID:          item.ID,
			ProductID:   item.ProductID,
			ProductName: prodName,
			ProductSKU:  prodSKU,
			Quantity:    item.Quantity,
			ReceivedQty: item.ReceivedQty,
		}
	}

	c.JSON(http.StatusOK, models.TransferResponse{
		ID:             transfer.ID,
		TransferNumber: transfer.TransferNumber,
		FromWarehouse: models.WarehouseResponse{
			ID:        transfer.FromWarehouse.ID,
			Name:      transfer.FromWarehouse.Name,
			Code:      transfer.FromWarehouse.Code,
			Location:  transfer.FromWarehouse.Location,
			IsDefault: transfer.FromWarehouse.IsDefault,
			IsActive:  transfer.FromWarehouse.IsActive,
		},
		ToWarehouse: models.WarehouseResponse{
			ID:        transfer.ToWarehouse.ID,
			Name:      transfer.ToWarehouse.Name,
			Code:      transfer.ToWarehouse.Code,
			Location:  transfer.ToWarehouse.Location,
			IsDefault: transfer.ToWarehouse.IsDefault,
			IsActive:  transfer.ToWarehouse.IsActive,
		},
		Status:      transfer.Status,
		Notes:       transfer.Notes,
		CreatedBy:   transfer.CreatedBy,
		CreatorName: creatorName,
		ApprovedBy:  transfer.ApprovedBy,
		ApprovedAt:  transfer.ApprovedAt,
		CompletedBy: transfer.CompletedBy,
		CompletedAt: transfer.CompletedAt,
		Items:       items,
		CreatedAt:   transfer.CreatedAt,
	})
}

func (h *TransferHandler) Create(c *gin.Context) {
	var req CreateTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.FromWarehouseID == req.ToWarehouseID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Source and destination warehouses must be different"})
		return
	}

	userID, _ := c.Get("user_id")

	transferNumber := generateTransferNumber()

	transfer := models.StockTransfer{
		TransferNumber:  transferNumber,
		FromWarehouseID: req.FromWarehouseID,
		ToWarehouseID:   req.ToWarehouseID,
		Status:          models.TransferStatusPending,
		Notes:           req.Notes,
		CreatedBy:       userID.(uint),
	}

	for _, item := range req.Items {
		transfer.Items = append(transfer.Items, models.TransferItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})
	}

	if err := database.DB.Create(&transfer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create transfer"})
		return
	}

	database.DB.Preload("FromWarehouse").Preload("ToWarehouse").Preload("Items.Product").First(&transfer, transfer.ID)

	logAudit(c, "stock_transfer", transfer.ID, models.AuditActionCreate, "", fmt.Sprintf("Created transfer %s", transferNumber), userID.(uint))

	c.JSON(http.StatusCreated, transfer)
}

func (h *TransferHandler) Approve(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transfer ID"})
		return
	}

	userID, _ := c.Get("user_id")

	var transfer models.StockTransfer
	if err := database.DB.First(&transfer, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transfer not found"})
		return
	}

	if transfer.Status != models.TransferStatusPending {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Transfer cannot be approved in current status"})
		return
	}

	now := time.Now()
	approvedBy := userID.(uint)

	if err := database.DB.Model(&transfer).Updates(map[string]interface{}{
		"status":      models.TransferStatusApproved,
		"approved_by": approvedBy,
		"approved_at": now,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve transfer"})
		return
	}

	logAudit(c, "stock_transfer", transfer.ID, models.AuditActionUpdate, transfer.Status, models.TransferStatusApproved, userID.(uint))

	c.JSON(http.StatusOK, gin.H{"message": "Transfer approved"})
}

func (h *TransferHandler) StartTransit(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transfer ID"})
		return
	}

	userID, _ := c.Get("user_id")

	var transfer models.StockTransfer
	if err := database.DB.First(&transfer, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transfer not found"})
		return
	}

	if transfer.Status != models.TransferStatusApproved {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Transfer must be approved first"})
		return
	}

	if err := database.DB.Model(&transfer).Update("status", models.TransferStatusInTransit).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update transfer"})
		return
	}

	logAudit(c, "stock_transfer", transfer.ID, models.AuditActionUpdate, transfer.Status, models.TransferStatusInTransit, userID.(uint))

	c.JSON(http.StatusOK, gin.H{"message": "Transfer in transit"})
}

func (h *TransferHandler) Complete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transfer ID"})
		return
	}

	userID, _ := c.Get("user_id")

	var transfer models.StockTransfer
	if err := database.DB.Preload("Items").First(&transfer, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transfer not found"})
		return
	}

	if transfer.Status != models.TransferStatusApproved && transfer.Status != models.TransferStatusInTransit {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Transfer must be in approved or in transit status"})
		return
	}

	tx := database.DB.Begin()

	for _, item := range transfer.Items {
		var fromInv models.Inventory
		if err := tx.Where("product_id = ? AND warehouse_id = ?", item.ProductID, transfer.FromWarehouseID).First(&fromInv).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Inventory not found in source warehouse"})
			return
		}

		if fromInv.Quantity < item.Quantity {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient stock in source warehouse"})
			return
		}

		fromInv.Quantity -= item.Quantity
		tx.Save(&fromInv)

		var toInv models.Inventory
		if err := tx.Where("product_id = ? AND warehouse_id = ?", item.ProductID, transfer.ToWarehouseID).First(&toInv).Error; err != nil {
			toInv = models.Inventory{
				ProductID:   item.ProductID,
				WarehouseID: transfer.ToWarehouseID,
				Quantity:    0,
			}
			tx.Create(&toInv)
		}

		toInv.Quantity += item.Quantity
		tx.Save(&toInv)

		tx.Create(&models.Transaction{
			ProductID: item.ProductID,
			Type:      models.TransactionTypeTransfer,
			Quantity:  -item.Quantity,
			Notes:     fmt.Sprintf("Transfer %s out", transfer.TransferNumber),
			UserID:    userID.(uint),
		})

		tx.Create(&models.Transaction{
			ProductID: item.ProductID,
			Type:      models.TransactionTypeTransfer,
			Quantity:  item.Quantity,
			Notes:     fmt.Sprintf("Transfer %s in", transfer.TransferNumber),
			UserID:    userID.(uint),
		})
	}

	now := time.Now()
	completedBy := userID.(uint)

	if err := tx.Model(&transfer).Updates(map[string]interface{}{
		"status":       models.TransferStatusCompleted,
		"completed_by": completedBy,
		"completed_at": now,
	}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete transfer"})
		return
	}

	tx.Commit()

	logAudit(c, "stock_transfer", transfer.ID, models.AuditActionUpdate, transfer.Status, models.TransferStatusCompleted, userID.(uint))

	c.JSON(http.StatusOK, gin.H{"message": "Transfer completed"})
}

func (h *TransferHandler) Cancel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transfer ID"})
		return
	}

	userID, _ := c.Get("user_id")

	var transfer models.StockTransfer
	if err := database.DB.First(&transfer, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transfer not found"})
		return
	}

	if transfer.Status == models.TransferStatusCompleted || transfer.Status == models.TransferStatusCancelled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Transfer cannot be cancelled in current status"})
		return
	}

	if err := database.DB.Model(&transfer).Update("status", models.TransferStatusCancelled).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel transfer"})
		return
	}

	logAudit(c, "stock_transfer", transfer.ID, models.AuditActionUpdate, transfer.Status, models.TransferStatusCancelled, userID.(uint))

	c.JSON(http.StatusOK, gin.H{"message": "Transfer cancelled"})
}

func generateTransferNumber() string {
	return fmt.Sprintf("TRF-%s", time.Now().Format("20060102150405"))
}

func logAudit(c *gin.Context, entityType string, entityID uint, action, oldValue, description string, userID uint) {
	ip := c.ClientIP()
	agent := c.GetHeader("User-Agent")

	log := models.AuditLog{
		EntityType:  entityType,
		EntityID:    entityID,
		Action:      action,
		OldValue:    oldValue,
		Description: description,
		UserID:      userID,
		IPAddress:   ip,
		UserAgent:   agent,
	}
	database.DB.Create(&log)
}
