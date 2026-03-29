package handlers

import (
	"net/http"
	"strconv"

	"inventory-ims/internal/database"
	"inventory-ims/internal/models"

	"github.com/gin-gonic/gin"
)

type WarehouseHandler struct{}

func NewWarehouseHandler() *WarehouseHandler {
	return &WarehouseHandler{}
}

type CreateWarehouseRequest struct {
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Location    string `json:"location"`
	Description string `json:"description"`
	IsDefault   bool   `json:"is_default"`
}

type UpdateWarehouseRequest struct {
	Name        string `json:"name"`
	Code        string `json:"code"`
	Location    string `json:"location"`
	Description string `json:"description"`
	IsDefault   *bool  `json:"is_default"`
	IsActive    *bool  `json:"is_active"`
}

func (h *WarehouseHandler) List(c *gin.Context) {
	var warehouses []models.Warehouse
	query := database.DB.Where("is_active = ?", true)

	search := c.Query("search")
	if search != "" {
		query = query.Where("name LIKE ? OR code LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	if err := query.Order("is_default DESC, name ASC").Find(&warehouses).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch warehouses"})
		return
	}

	response := []models.WarehouseResponse{}
	for _, w := range warehouses {
		response = append(response, models.WarehouseResponse{
			ID:          w.ID,
			Name:        w.Name,
			Code:        w.Code,
			Location:    w.Location,
			IsDefault:   w.IsDefault,
			IsActive:    w.IsActive,
			Description: w.Description,
		})
	}

	c.JSON(http.StatusOK, response)
}

func (h *WarehouseHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid warehouse ID"})
		return
	}

	var warehouse models.Warehouse
	if err := database.DB.First(&warehouse, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Warehouse not found"})
		return
	}

	c.JSON(http.StatusOK, models.WarehouseResponse{
		ID:          warehouse.ID,
		Name:        warehouse.Name,
		Code:        warehouse.Code,
		Location:    warehouse.Location,
		IsDefault:   warehouse.IsDefault,
		IsActive:    warehouse.IsActive,
		Description: warehouse.Description,
	})
}

func (h *WarehouseHandler) Create(c *gin.Context) {
	var req CreateWarehouseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existing models.Warehouse
	if err := database.DB.Where("code = ?", req.Code).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Warehouse code already exists"})
		return
	}

	if req.IsDefault {
		database.DB.Model(&models.Warehouse{}).Where("is_default = ?", true).Update("is_default", false)
	}

	warehouse := models.Warehouse{
		Name:        req.Name,
		Code:        req.Code,
		Location:    req.Location,
		Description: req.Description,
		IsDefault:   req.IsDefault,
		IsActive:    true,
	}

	if err := database.DB.Create(&warehouse).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create warehouse"})
		return
	}

	c.JSON(http.StatusCreated, warehouse)
}

func (h *WarehouseHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid warehouse ID"})
		return
	}

	var warehouse models.Warehouse
	if err := database.DB.First(&warehouse, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Warehouse not found"})
		return
	}

	var req UpdateWarehouseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Code != "" {
		var existing models.Warehouse
		if err := database.DB.Where("code = ? AND id != ?", req.Code, id).First(&existing).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Warehouse code already exists"})
			return
		}
		updates["code"] = req.Code
	}
	if req.Location != "" {
		updates["location"] = req.Location
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.IsDefault != nil {
		if *req.IsDefault {
			database.DB.Model(&models.Warehouse{}).Where("is_default = ?", true).Update("is_default", false)
		}
		updates["is_default"] = *req.IsDefault
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if err := database.DB.Model(&warehouse).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update warehouse"})
		return
	}

	c.JSON(http.StatusOK, warehouse)
}

func (h *WarehouseHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid warehouse ID"})
		return
	}

	var warehouse models.Warehouse
	if err := database.DB.First(&warehouse, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Warehouse not found"})
		return
	}

	if warehouse.IsDefault {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete default warehouse"})
		return
	}

	if err := database.DB.Delete(&warehouse).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete warehouse"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Warehouse deleted"})
}

func (h *WarehouseHandler) GetDefault(c *gin.Context) {
	var warehouse models.Warehouse
	if err := database.DB.Where("is_default = ?", true).First(&warehouse).Error; err != nil {
		var wh models.Warehouse
		if err := database.DB.First(&wh).Error; err == nil {
			c.JSON(http.StatusOK, models.WarehouseResponse{
				ID:        wh.ID,
				Name:      wh.Name,
				Code:      wh.Code,
				Location:  wh.Location,
				IsDefault: wh.IsDefault,
				IsActive:  wh.IsActive,
			})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "No warehouse found"})
		return
	}

	c.JSON(http.StatusOK, models.WarehouseResponse{
		ID:          warehouse.ID,
		Name:        warehouse.Name,
		Code:        warehouse.Code,
		Location:    warehouse.Location,
		IsDefault:   warehouse.IsDefault,
		IsActive:    warehouse.IsActive,
		Description: warehouse.Description,
	})
}
