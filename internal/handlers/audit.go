package handlers

import (
	"net/http"
	"strconv"
	"time"

	"inventory-ims/internal/database"
	"inventory-ims/internal/models"

	"github.com/gin-gonic/gin"
)

type AuditHandler struct{}

func NewAuditHandler() *AuditHandler {
	return &AuditHandler{}
}

type CreateAuditCycleRequest struct {
	Title       string     `json:"title" binding:"required"`
	WarehouseID uint       `json:"warehouse_id" binding:"required"`
	StartDate   *time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
	Notes       string     `json:"notes"`
}

type CountItemRequest struct {
	ProductID  uint   `json:"product_id" binding:"required"`
	CountedQty int    `json:"counted_qty" binding:"required"`
	Notes      string `json:"notes"`
}

func (h *AuditHandler) ListCycles(c *gin.Context) {
	var cycles []models.AuditCycle
	query := database.DB.Preload("Warehouse").Preload("Creator")

	status := c.Query("status")
	if status != "" {
		query = query.Where("status = ?", status)
	}

	warehouseID := c.Query("warehouse_id")
	if warehouseID != "" {
		id, _ := strconv.ParseUint(warehouseID, 10, 32)
		query = query.Where("warehouse_id = ?", id)
	}

	if err := query.Order("created_at DESC").Find(&cycles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch audit cycles"})
		return
	}

	var response []models.AuditCycleResponse
	for _, cycle := range cycles {
		creatorName := ""
		if cycle.Creator != nil {
			creatorName = cycle.Creator.Username
		}

		whName := ""
		if cycle.Warehouse != nil {
			whName = cycle.Warehouse.Name
		}

		var itemCount, verifiedCount int64
		database.DB.Model(&models.AuditItem{}).Where("audit_cycle_id = ?", cycle.ID).Count(&itemCount)
		database.DB.Model(&models.AuditItem{}).Where("audit_cycle_id = ? AND status IN ?", cycle.ID, []string{models.AuditItemStatusVerified, models.AuditItemStatusAdjusted}).Count(&verifiedCount)

		response = append(response, models.AuditCycleResponse{
			ID:            cycle.ID,
			Title:         cycle.Title,
			WarehouseID:   cycle.WarehouseID,
			WarehouseName: whName,
			Status:        cycle.Status,
			StartDate:     cycle.StartDate,
			EndDate:       cycle.EndDate,
			Notes:         cycle.Notes,
			CreatorName:   creatorName,
			ItemCount:     int(itemCount),
			VerifiedCount: int(verifiedCount),
			CreatedAt:     cycle.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, response)
}

func (h *AuditHandler) GetCycle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid audit cycle ID"})
		return
	}

	var cycle models.AuditCycle
	if err := database.DB.Preload("Warehouse").Preload("Creator").Preload("Items.Product").First(&cycle, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Audit cycle not found"})
		return
	}

	creatorName := ""
	if cycle.Creator != nil {
		creatorName = cycle.Creator.Username
	}

	whName := ""
	if cycle.Warehouse != nil {
		whName = cycle.Warehouse.Name
	}

	items := make([]models.AuditItemResponse, len(cycle.Items))
	for i, item := range cycle.Items {
		prodName := ""
		prodSKU := ""
		if item.Product != nil {
			prodName = item.Product.Name
			prodSKU = item.Product.SKU
		}
		items[i] = models.AuditItemResponse{
			ID:          item.ID,
			ProductID:   item.ProductID,
			ProductName: prodName,
			ProductSKU:  prodSKU,
			SystemQty:   item.SystemQty,
			CountedQty:  item.CountedQty,
			Variance:    item.Variance,
			Status:      item.Status,
			Notes:       item.Notes,
			CountedBy:   item.CountedBy,
			CountedAt:   item.CountedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"id":             cycle.ID,
		"title":          cycle.Title,
		"warehouse_id":   cycle.WarehouseID,
		"warehouse_name": whName,
		"status":         cycle.Status,
		"start_date":     cycle.StartDate,
		"end_date":       cycle.EndDate,
		"notes":          cycle.Notes,
		"creator_name":   creatorName,
		"items":          items,
		"created_at":     cycle.CreatedAt,
	})
}

func (h *AuditHandler) CreateCycle(c *gin.Context) {
	var req CreateAuditCycleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")

	cycle := models.AuditCycle{
		Title:       req.Title,
		WarehouseID: req.WarehouseID,
		Status:      models.AuditStatusPlanned,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		Notes:       req.Notes,
		CreatedBy:   userID.(uint),
	}

	if err := database.DB.Create(&cycle).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create audit cycle"})
		return
	}

	var products []models.Product
	database.DB.Preload("Inventory", "warehouse_id = ?", req.WarehouseID).Where("is_active = ?", true).Find(&products)

	for _, product := range products {
		systemQty := 0
		if product.Inventory != nil {
			systemQty = product.Inventory.Quantity
		}

		auditItem := models.AuditItem{
			AuditCycleID: cycle.ID,
			ProductID:    product.ID,
			SystemQty:    systemQty,
			Status:       models.AuditItemStatusPending,
		}
		database.DB.Create(&auditItem)
	}

	logAudit(c, "audit_cycle", cycle.ID, models.AuditActionCreate, "", "Created audit cycle: "+req.Title, userID.(uint))

	c.JSON(http.StatusCreated, gin.H{"message": "Audit cycle created", "id": cycle.ID})
}

func (h *AuditHandler) StartCycle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid audit cycle ID"})
		return
	}

	userID, _ := c.Get("user_id")

	var cycle models.AuditCycle
	if err := database.DB.First(&cycle, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Audit cycle not found"})
		return
	}

	if cycle.Status != models.AuditStatusPlanned {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Audit cycle cannot be started in current status"})
		return
	}

	now := time.Now()
	if err := database.DB.Model(&cycle).Updates(map[string]interface{}{
		"status":     models.AuditStatusInProgress,
		"start_date": now,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start audit cycle"})
		return
	}

	logAudit(c, "audit_cycle", cycle.ID, models.AuditActionUpdate, models.AuditStatusPlanned, models.AuditStatusInProgress, userID.(uint))

	c.JSON(http.StatusOK, gin.H{"message": "Audit cycle started"})
}

func (h *AuditHandler) CountItem(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid audit cycle ID"})
		return
	}

	var req CountItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")

	var item models.AuditItem
	if err := database.DB.Preload("AuditCycle").Where("audit_cycle_id = ? AND product_id = ?", id, req.ProductID).First(&item).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Audit item not found"})
		return
	}

	if item.AuditCycle.Status != models.AuditStatusInProgress {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Audit cycle is not in progress"})
		return
	}

	now := time.Now()
	variance := req.CountedQty - item.SystemQty

	if err := database.DB.Model(&item).Updates(map[string]interface{}{
		"counted_qty": req.CountedQty,
		"variance":    variance,
		"status":      models.AuditItemStatusCounted,
		"notes":       req.Notes,
		"counted_by":  userID.(uint),
		"counted_at":  now,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit count"})
		return
	}

	logAudit(c, "audit_item", item.ID, models.AuditActionUpdate, "", "Counted: "+strconv.Itoa(req.CountedQty), userID.(uint))

	c.JSON(http.StatusOK, gin.H{"message": "Count submitted", "variance": variance})
}

func (h *AuditHandler) AdjustInventory(c *gin.Context) {
	_, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid audit cycle ID"})
		return
	}

	itemID, err := strconv.ParseUint(c.Param("item_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
		return
	}

	userID, _ := c.Get("user_id")

	var item models.AuditItem
	if err := database.DB.Preload("AuditCycle").First(&item, itemID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Audit item not found"})
		return
	}

	if item.AuditCycle.Status != models.AuditStatusInProgress {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Audit cycle is not in progress"})
		return
	}

	var inventory models.Inventory
	if err := database.DB.Where("product_id = ? AND warehouse_id = ?", item.ProductID, item.AuditCycle.WarehouseID).First(&inventory).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Inventory record not found"})
		return
	}

	oldQty := inventory.Quantity
	inventory.Quantity = item.CountedQty
	inventory.LastUpdated = time.Now()
	database.DB.Save(&inventory)

	database.DB.Model(&item).Updates(map[string]interface{}{
		"status": models.AuditItemStatusAdjusted,
	})

	database.DB.Create(&models.Transaction{
		ProductID: item.ProductID,
		Type:      models.TransactionTypeAdjustment,
		Quantity:  item.CountedQty - oldQty,
		Notes:     "Audit adjustment",
		UserID:    userID.(uint),
	})

	logAudit(c, "inventory", inventory.ID, models.AuditActionUpdate, strconv.Itoa(oldQty), strconv.Itoa(item.CountedQty), userID.(uint))

	c.JSON(http.StatusOK, gin.H{"message": "Inventory adjusted"})
}

func (h *AuditHandler) CompleteCycle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid audit cycle ID"})
		return
	}

	userID, _ := c.Get("user_id")

	var cycle models.AuditCycle
	if err := database.DB.First(&cycle, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Audit cycle not found"})
		return
	}

	if cycle.Status != models.AuditStatusInProgress {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Audit cycle is not in progress"})
		return
	}

	now := time.Now()
	if err := database.DB.Model(&cycle).Updates(map[string]interface{}{
		"status":   models.AuditStatusCompleted,
		"end_date": now,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete audit cycle"})
		return
	}

	logAudit(c, "audit_cycle", cycle.ID, models.AuditActionUpdate, models.AuditStatusInProgress, models.AuditStatusCompleted, userID.(uint))

	c.JSON(http.StatusOK, gin.H{"message": "Audit cycle completed"})
}

func (h *AuditHandler) ListLogs(c *gin.Context) {
	var logs []models.AuditLog
	query := database.DB.Preload("User")

	entityType := c.Query("entity_type")
	if entityType != "" {
		query = query.Where("entity_type = ?", entityType)
	}

	entityID := c.Query("entity_id")
	if entityID != "" {
		id, _ := strconv.ParseUint(entityID, 10, 32)
		query = query.Where("entity_id = ?", id)
	}

	action := c.Query("action")
	if action != "" {
		query = query.Where("action = ?", action)
	}

	userID := c.Query("user_id")
	if userID != "" {
		id, _ := strconv.ParseUint(userID, 10, 32)
		query = query.Where("user_id = ?", id)
	}

	limit := c.DefaultQuery("limit", "100")
	l, _ := strconv.Atoi(limit)
	query = query.Order("created_at DESC").Limit(l)

	if err := query.Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch audit logs"})
		return
	}

	var response []models.AuditLogResponse
	for _, log := range logs {
		userName := ""
		if log.User != nil {
			userName = log.User.Username
		}
		response = append(response, models.AuditLogResponse{
			ID:          log.ID,
			EntityType:  log.EntityType,
			EntityID:    log.EntityID,
			Action:      log.Action,
			OldValue:    log.OldValue,
			Description: log.Description,
			UserName:    userName,
			IPAddress:   log.IPAddress,
			CreatedAt:   log.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, response)
}
