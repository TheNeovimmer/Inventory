package handlers

import (
	"net/http"
	"strconv"
	"time"

	"inventory-ims/internal/database"
	"inventory-ims/internal/models"

	"github.com/gin-gonic/gin"
)

type ProductionHandler struct{}

func NewProductionHandler() *ProductionHandler {
	return &ProductionHandler{}
}

type BOMHandler struct{}

func NewBOMHandler() *BOMHandler {
	return &BOMHandler{}
}

type CreateBOMRequest struct {
	ProductID uint                   `json:"product_id" binding:"required"`
	Version   int                    `json:"version"`
	Notes     string                 `json:"notes"`
	Items     []CreateBOMItemRequest `json:"items"`
}

type CreateBOMItemRequest struct {
	ComponentProductID uint    `json:"component_product_id" binding:"required"`
	QuantityRequired   float64 `json:"quantity_required" binding:"required"`
}

func (h *BOMHandler) List(c *gin.Context) {
	var boms []models.BOM
	if err := database.DB.Preload("Product").Preload("Items.ComponentProduct").Where("is_active = ?", true).Find(&boms).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch BOMs"})
		return
	}

	var response []models.BOMResponse
	for _, bom := range boms {
		productName := ""
		if bom.Product != nil {
			productName = bom.Product.Name
		}
		var items []models.BOMItemResponse
		for _, item := range bom.Items {
			compName := ""
			compSKU := ""
			if item.ComponentProduct != nil {
				compName = item.ComponentProduct.Name
				compSKU = item.ComponentProduct.SKU
			}
			items = append(items, models.BOMItemResponse{
				ID:                 item.ID,
				ComponentProductID: item.ComponentProductID,
				ComponentName:      compName,
				ComponentSKU:       compSKU,
				QuantityRequired:   item.QuantityRequired,
			})
		}
		response = append(response, models.BOMResponse{
			ID:          bom.ID,
			ProductID:   bom.ProductID,
			ProductName: productName,
			Version:     bom.Version,
			IsActive:    bom.IsActive,
			Notes:       bom.Notes,
			Items:       items,
		})
	}

	c.JSON(http.StatusOK, response)
}

func (h *BOMHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid BOM ID"})
		return
	}

	var bom models.BOM
	if err := database.DB.Preload("Product").Preload("Items.ComponentProduct").First(&bom, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "BOM not found"})
		return
	}

	c.JSON(http.StatusOK, bom)
}

func (h *BOMHandler) Create(c *gin.Context) {
	var req CreateBOMRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existing models.BOM
	database.DB.Where("product_id = ? AND is_active = ?", req.ProductID, true).First(&existing)
	if existing.ID > 0 {
		database.DB.Model(&existing).Update("is_active", false)
	}

	version := req.Version
	if version == 0 {
		version = 1
	}

	bom := models.BOM{
		ProductID: req.ProductID,
		Version:   version,
		IsActive:  true,
		Notes:     req.Notes,
	}

	if err := database.DB.Create(&bom).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create BOM"})
		return
	}

	for _, item := range req.Items {
		bomItem := models.BOMItem{
			BOMID:              bom.ID,
			ComponentProductID: item.ComponentProductID,
			QuantityRequired:   item.QuantityRequired,
		}
		database.DB.Create(&bomItem)
	}

	c.JSON(http.StatusCreated, bom)
}

func (h *BOMHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid BOM ID"})
		return
	}

	database.DB.Model(&models.BOM{}).Where("id = ?", id).Update("is_active", false)
	c.JSON(http.StatusOK, gin.H{"message": "BOM deactivated"})
}

type CreateProductionRequest struct {
	BOMID    uint   `json:"bom_id" binding:"required"`
	Quantity int    `json:"quantity" binding:"required"`
	Notes    string `json:"notes"`
}

func (h *ProductionHandler) List(c *gin.Context) {
	var orders []models.ProductionOrder
	if err := database.DB.Preload("BOM.Product").Order("created_at DESC").Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch production orders"})
		return
	}

	var response []models.ProductionOrderResponse
	for _, o := range orders {
		productName := ""
		if o.BOM != nil && o.BOM.Product != nil {
			productName = o.BOM.Product.Name
		}
		response = append(response, models.ProductionOrderResponse{
			ID:          o.ID,
			BOMID:       o.BOMID,
			ProductName: productName,
			Quantity:    o.Quantity,
			Status:      o.Status,
			StartedAt:   o.StartedAt,
			CompletedAt: o.CompletedAt,
			Notes:       o.Notes,
			CreatedAt:   o.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, response)
}

func (h *ProductionHandler) Create(c *gin.Context) {
	var req CreateProductionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order := models.ProductionOrder{
		BOMID:    req.BOMID,
		Quantity: req.Quantity,
		Status:   models.ProductionStatusDraft,
		Notes:    req.Notes,
	}

	if err := database.DB.Create(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create production order"})
		return
	}

	c.JSON(http.StatusCreated, order)
}

func (h *ProductionHandler) Start(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	userID, _ := c.Get("user_id")

	var order models.ProductionOrder
	if err := database.DB.Preload("BOM.Items.ComponentProduct").Preload("BOM.Product").First(&order, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	if order.Status != models.ProductionStatusDraft {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order already started"})
		return
	}

	now := time.Now()
	for _, item := range order.BOM.Items {
		if item.ComponentProduct == nil {
			continue
		}
		requiredQty := item.QuantityRequired * float64(order.Quantity)

		var inv models.Inventory
		if err := database.DB.Where("product_id = ?", item.ComponentProductID).First(&inv).Error; err != nil {
			inv = models.Inventory{ProductID: item.ComponentProductID, Quantity: 0}
			database.DB.Create(&inv)
		}

		if inv.Quantity < int(requiredQty) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient stock for component: " + item.ComponentProduct.Name})
			return
		}

		inv.Quantity -= int(requiredQty)
		inv.LastUpdated = now
		database.DB.Save(&inv)

		transaction := models.Transaction{
			ProductID:     item.ComponentProductID,
			Type:          models.TransactionTypeProductionOut,
			Quantity:      -int(requiredQty),
			ReferenceType: "production_order",
			ReferenceID:   &order.ID,
			Notes:         "Production started",
			UserID:        userID.(uint),
		}
		database.DB.Create(&transaction)
	}

	order.Status = models.ProductionStatusInProgress
	order.StartedAt = &now
	database.DB.Save(&order)

	c.JSON(http.StatusOK, order)
}

func (h *ProductionHandler) Complete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	userID, _ := c.Get("user_id")

	var order models.ProductionOrder
	if err := database.DB.Preload("BOM.Product").First(&order, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	if order.Status != models.ProductionStatusInProgress {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order not in progress"})
		return
	}

	now := time.Now()

	var finishedProductInv models.Inventory
	if err := database.DB.Where("product_id = ?", order.BOM.ProductID).First(&finishedProductInv).Error; err != nil {
		finishedProductInv = models.Inventory{ProductID: order.BOM.ProductID, Quantity: 0}
		database.DB.Create(&finishedProductInv)
	}

	finishedProductInv.Quantity += order.Quantity
	finishedProductInv.LastUpdated = now
	database.DB.Save(&finishedProductInv)

	transaction := models.Transaction{
		ProductID:     order.BOM.ProductID,
		Type:          models.TransactionTypeProductionIn,
		Quantity:      order.Quantity,
		ReferenceType: "production_order",
		ReferenceID:   &order.ID,
		Notes:         "Production completed",
		UserID:        userID.(uint),
	}
	database.DB.Create(&transaction)

	order.Status = models.ProductionStatusCompleted
	order.CompletedAt = &now
	database.DB.Save(&order)

	c.JSON(http.StatusOK, order)
}

func (h *ProductionHandler) Cancel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	var order models.ProductionOrder
	if err := database.DB.First(&order, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	if order.Status == models.ProductionStatusCompleted {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot cancel completed order"})
		return
	}

	if order.Status == models.ProductionStatusInProgress {
		userID, _ := c.Get("user_id")
		now := time.Now()

		var bom models.BOM
		database.DB.Preload("Items.ComponentProduct").First(&bom, order.BOMID)

		for _, item := range bom.Items {
			if item.ComponentProduct == nil {
				continue
			}
			requiredQty := item.QuantityRequired * float64(order.Quantity)

			var inv models.Inventory
			database.DB.Where("product_id = ?", item.ComponentProductID).First(&inv)
			inv.Quantity += int(requiredQty)
			inv.LastUpdated = now
			database.DB.Save(&inv)

			transaction := models.Transaction{
				ProductID:     item.ComponentProductID,
				Type:          models.TransactionTypeProductionOut,
				Quantity:      int(requiredQty),
				ReferenceType: "production_order",
				ReferenceID:   &order.ID,
				Notes:         "Production cancelled - stock returned",
				UserID:        userID.(uint),
			}
			database.DB.Create(&transaction)
		}
	}

	order.Status = models.ProductionStatusCancelled
	database.DB.Save(&order)

	c.JSON(http.StatusOK, order)
}
