package models

import (
	"time"

	"gorm.io/gorm"
)

type Product struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	SKU          string         `gorm:"uniqueIndex;size:50;not null" json:"sku"`
	Name         string         `gorm:"size:200;not null" json:"name"`
	Description  string         `gorm:"type:text" json:"description"`
	CategoryID   *uint          `gorm:"index" json:"category_id"`
	Barcode      string         `gorm:"size:100;index" json:"barcode"`
	UnitPrice    float64        `gorm:"type:decimal(10,2);not null" json:"unit_price"`
	CostPrice    float64        `gorm:"type:decimal(10,2)" json:"cost_price"`
	ReorderPoint int            `gorm:"default:10" json:"reorder_point"`
	IsActive     bool           `gorm:"default:true" json:"is_active"`
	HasBOM       bool           `gorm:"default:false" json:"has_bom"`
	ImageURL     string         `gorm:"size:255" json:"image_url"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	Category     *Category      `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Inventory    *Inventory     `gorm:"foreignKey:ProductID" json:"inventory,omitempty"`
}

type ProductResponse struct {
	ID           uint    `json:"id"`
	SKU          string  `json:"sku"`
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	CategoryID   *uint   `json:"category_id"`
	CategoryName string  `json:"category_name"`
	Barcode      string  `json:"barcode"`
	UnitPrice    float64 `json:"unit_price"`
	CostPrice    float64 `json:"cost_price"`
	ReorderPoint int     `json:"reorder_point"`
	IsActive     bool    `json:"is_active"`
	HasBOM       bool    `json:"has_bom"`
	ImageURL     string  `json:"image_url"`
	Quantity     int     `json:"quantity"`
}
