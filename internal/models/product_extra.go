package models

import (
	"time"

	"gorm.io/gorm"
)

type Brand struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"size:100;not null" json:"name"`
	Slug      string         `gorm:"size:100;uniqueIndex" json:"slug"`
	Logo      string         `gorm:"size:255" json:"logo"`
	Website   string         `gorm:"size:255" json:"website"`
	IsActive  bool           `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Unit struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	Name           string         `gorm:"size:50;not null" json:"name"`
	ShortName      string         `gorm:"size:20;not null" json:"short_name"`
	BaseUnit       *uint          `gorm:"index" json:"base_unit"`
	Operator       string         `gorm:"size:10" json:"operator"` // multiply, divide
	OperationValue float64        `gorm:"default:1" json:"operation_value"`
	IsActive       bool           `gorm:"default:true" json:"is_active"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

type ProductVariant struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	ProductID uint           `gorm:"index;not null" json:"product_id"`
	Name      string         `gorm:"size:100;not null" json:"name"`
	SKU       string         `gorm:"size:50" json:"sku"`
	Barcode   string         `gorm:"size:100" json:"barcode"`
	Price     float64        `gorm:"type:decimal(10,2)" json:"price"`
	CostPrice float64        `gorm:"type:decimal(10,2)" json:"cost_price"`
	ImageURL  string         `gorm:"size:255" json:"image_url"`
	IsActive  bool           `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type ProductType string

const (
	ProductTypeStandard ProductType = "standard"
	ProductTypeVariable ProductType = "variable"
	ProductTypeService  ProductType = "service"
)
