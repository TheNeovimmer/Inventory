package models

import (
	"time"

	"gorm.io/gorm"
)

type BOM struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	ProductID uint           `gorm:"uniqueIndex:idx_bom_product;not null" json:"product_id"`
	Version   int            `gorm:"default:1" json:"version"`
	IsActive  bool           `gorm:"default:true" json:"is_active"`
	Notes     string         `gorm:"type:text" json:"notes"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Product   *Product       `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Items     []BOMItem      `gorm:"foreignKey:BOMID" json:"items"`
}

type BOMItem struct {
	ID                 uint           `gorm:"primaryKey" json:"id"`
	BOMID              uint           `gorm:"uniqueIndex:idx_bom_item;not null" json:"bom_id"`
	ComponentProductID uint           `gorm:"uniqueIndex:idx_bom_item;not null" json:"component_product_id"`
	QuantityRequired   float64        `gorm:"type:decimal(10,3);not null" json:"quantity_required"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
	ComponentProduct   *Product       `gorm:"foreignKey:ComponentProductID" json:"component_product,omitempty"`
}

type BOMResponse struct {
	ID          uint              `json:"id"`
	ProductID   uint              `json:"product_id"`
	ProductName string            `json:"product_name"`
	Version     int               `json:"version"`
	IsActive    bool              `json:"is_active"`
	Notes       string            `json:"notes"`
	Items       []BOMItemResponse `json:"items"`
}

type BOMItemResponse struct {
	ID                 uint    `json:"id"`
	ComponentProductID uint    `json:"component_product_id"`
	ComponentName      string  `json:"component_name"`
	ComponentSKU       string  `json:"component_sku"`
	QuantityRequired   float64 `json:"quantity_required"`
}
