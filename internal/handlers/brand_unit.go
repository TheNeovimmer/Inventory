package handlers

import (
	"net/http"
	"strconv"

	"inventory-ims/internal/database"
	"inventory-ims/internal/models"

	"github.com/gin-gonic/gin"
)

type BrandHandler struct{}

func NewBrandHandler() *BrandHandler {
	return &BrandHandler{}
}

func (h *BrandHandler) List(c *gin.Context) {
	var brands []models.Brand
	query := database.DB.Where("is_active = ?", true)

	search := c.Query("search")
	if search != "" {
		query = query.Where("name LIKE ?", "%"+search+"%")
	}

	if err := query.Order("name").Find(&brands).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch brands"})
		return
	}

	c.JSON(http.StatusOK, brands)
}

func (h *BrandHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid brand ID"})
		return
	}

	var brand models.Brand
	if err := database.DB.First(&brand, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Brand not found"})
		return
	}

	c.JSON(http.StatusOK, brand)
}

func (h *BrandHandler) Create(c *gin.Context) {
	var req struct {
		Name    string `json:"name" binding:"required"`
		Logo    string `json:"logo"`
		Website string `json:"website"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	brand := models.Brand{
		Name:     req.Name,
		Slug:     req.Name,
		Logo:     req.Logo,
		Website:  req.Website,
		IsActive: true,
	}

	if err := database.DB.Create(&brand).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create brand"})
		return
	}

	c.JSON(http.StatusCreated, brand)
}

func (h *BrandHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid brand ID"})
		return
	}

	var brand models.Brand
	if err := database.DB.First(&brand, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Brand not found"})
		return
	}

	var req struct {
		Name    string `json:"name"`
		Logo    string `json:"logo"`
		Website string `json:"website"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
		updates["slug"] = req.Name
	}
	if req.Logo != "" {
		updates["logo"] = req.Logo
	}
	if req.Website != "" {
		updates["website"] = req.Website
	}

	if err := database.DB.Model(&brand).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update brand"})
		return
	}

	c.JSON(http.StatusOK, brand)
}

func (h *BrandHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid brand ID"})
		return
	}

	var brand models.Brand
	if err := database.DB.First(&brand, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Brand not found"})
		return
	}

	if err := database.DB.Model(&brand).Update("is_active", false).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete brand"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Brand deleted"})
}

type UnitHandler struct{}

func NewUnitHandler() *UnitHandler {
	return &UnitHandler{}
}

func (h *UnitHandler) List(c *gin.Context) {
	var units []models.Unit
	query := database.DB.Where("is_active = ?", true)

	if err := query.Order("name").Find(&units).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch units"})
		return
	}

	c.JSON(http.StatusOK, units)
}

func (h *UnitHandler) Create(c *gin.Context) {
	var req struct {
		Name           string  `json:"name" binding:"required"`
		ShortName      string  `json:"short_name" binding:"required"`
		BaseUnit       *uint   `json:"base_unit"`
		Operator       string  `json:"operator"`
		OperationValue float64 `json:"operation_value"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	unit := models.Unit{
		Name:           req.Name,
		ShortName:      req.ShortName,
		BaseUnit:       req.BaseUnit,
		Operator:       req.Operator,
		OperationValue: req.OperationValue,
		IsActive:       true,
	}

	if unit.OperationValue == 0 {
		unit.OperationValue = 1
	}

	if err := database.DB.Create(&unit).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create unit"})
		return
	}

	c.JSON(http.StatusCreated, unit)
}

func (h *UnitHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid unit ID"})
		return
	}

	var unit models.Unit
	if err := database.DB.First(&unit, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Unit not found"})
		return
	}

	if err := database.DB.Model(&unit).Update("is_active", false).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete unit"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Unit deleted"})
}
