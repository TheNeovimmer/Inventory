package handlers

import (
	"net/http"
	"strconv"

	"inventory-ims/internal/database"
	"inventory-ims/internal/models"

	"github.com/gin-gonic/gin"
)

type ProductHandler struct{}

func NewProductHandler() *ProductHandler {
	return &ProductHandler{}
}

type CreateProductRequest struct {
	SKU          string  `json:"sku" binding:"required"`
	Name         string  `json:"name" binding:"required"`
	Description  string  `json:"description"`
	CategoryID   *uint   `json:"category_id"`
	Barcode      string  `json:"barcode"`
	UnitPrice    float64 `json:"unit_price" binding:"required"`
	CostPrice    float64 `json:"cost_price"`
	ReorderPoint int     `json:"reorder_point"`
	HasBOM       bool    `json:"has_bom"`
	ImageURL     string  `json:"image_url"`
}

type UpdateProductRequest struct {
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	CategoryID   *uint   `json:"category_id"`
	Barcode      string  `json:"barcode"`
	UnitPrice    float64 `json:"unit_price"`
	CostPrice    float64 `json:"cost_price"`
	ReorderPoint int     `json:"reorder_point"`
	HasBOM       bool    `json:"has_bom"`
	ImageURL     string  `json:"image_url"`
	IsActive     *bool   `json:"is_active"`
}

func (h *ProductHandler) List(c *gin.Context) {
	var products []models.Product
	query := database.DB.Preload("Category").Preload("Inventory")

	search := c.Query("search")
	if search != "" {
		query = query.Where("name LIKE ? OR sku LIKE ? OR barcode LIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	categoryID := c.Query("category_id")
	if categoryID != "" {
		id, _ := strconv.ParseUint(categoryID, 10, 32)
		query = query.Where("category_id = ?", id)
	}

	active := c.Query("active")
	if active != "" {
		isActive := active == "true"
		query = query.Where("is_active = ?", isActive)
	}

	if err := query.Order("created_at DESC").Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
		return
	}

	var response []models.ProductResponse
	for _, p := range products {
		qty := 0
		if p.Inventory != nil {
			qty = p.Inventory.Quantity
		}
		catName := ""
		if p.Category != nil {
			catName = p.Category.Name
		}
		response = append(response, models.ProductResponse{
			ID:           p.ID,
			SKU:          p.SKU,
			Name:         p.Name,
			Description:  p.Description,
			CategoryID:   p.CategoryID,
			CategoryName: catName,
			Barcode:      p.Barcode,
			UnitPrice:    p.UnitPrice,
			CostPrice:    p.CostPrice,
			ReorderPoint: p.ReorderPoint,
			IsActive:     p.IsActive,
			HasBOM:       p.HasBOM,
			ImageURL:     p.ImageURL,
			Quantity:     qty,
		})
	}

	c.JSON(http.StatusOK, response)
}

func (h *ProductHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	var product models.Product
	if err := database.DB.Preload("Category").Preload("Inventory").First(&product, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	qty := 0
	if product.Inventory != nil {
		qty = product.Inventory.Quantity
	}
	catName := ""
	if product.Category != nil {
		catName = product.Category.Name
	}

	c.JSON(http.StatusOK, models.ProductResponse{
		ID:           product.ID,
		SKU:          product.SKU,
		Name:         product.Name,
		Description:  product.Description,
		CategoryID:   product.CategoryID,
		CategoryName: catName,
		Barcode:      product.Barcode,
		UnitPrice:    product.UnitPrice,
		CostPrice:    product.CostPrice,
		ReorderPoint: product.ReorderPoint,
		IsActive:     product.IsActive,
		HasBOM:       product.HasBOM,
		ImageURL:     product.ImageURL,
		Quantity:     qty,
	})
}

func (h *ProductHandler) Create(c *gin.Context) {
	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existing models.Product
	if err := database.DB.Where("sku = ?", req.SKU).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "SKU already exists"})
		return
	}

	reorderPoint := req.ReorderPoint
	if reorderPoint == 0 {
		reorderPoint = 10
	}

	product := models.Product{
		SKU:          req.SKU,
		Name:         req.Name,
		Description:  req.Description,
		CategoryID:   req.CategoryID,
		Barcode:      req.Barcode,
		UnitPrice:    req.UnitPrice,
		CostPrice:    req.CostPrice,
		ReorderPoint: reorderPoint,
		HasBOM:       req.HasBOM,
		ImageURL:     req.ImageURL,
		IsActive:     true,
	}

	if err := database.DB.Create(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
		return
	}

	inventory := models.Inventory{
		ProductID:        product.ID,
		Quantity:         0,
		ReservedQuantity: 0,
	}
	database.DB.Create(&inventory)

	c.JSON(http.StatusCreated, product)
}

func (h *ProductHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	var product models.Product
	if err := database.DB.First(&product, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	var req UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.CategoryID != nil {
		updates["category_id"] = *req.CategoryID
	}
	if req.Barcode != "" {
		updates["barcode"] = req.Barcode
	}
	if req.UnitPrice > 0 {
		updates["unit_price"] = req.UnitPrice
	}
	if req.CostPrice > 0 {
		updates["cost_price"] = req.CostPrice
	}
	if req.ReorderPoint > 0 {
		updates["reorder_point"] = req.ReorderPoint
	}
	updates["has_bom"] = req.HasBOM
	if req.ImageURL != "" {
		updates["image_url"] = req.ImageURL
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if err := database.DB.Model(&product).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
		return
	}

	c.JSON(http.StatusOK, product)
}

func (h *ProductHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	if err := database.DB.Delete(&models.Product{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted"})
}
