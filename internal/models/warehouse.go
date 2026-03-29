package models

import (
	"time"

	"gorm.io/gorm"
)

type Warehouse struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"size:100;not null" json:"name"`
	Code        string         `gorm:"size:20;uniqueIndex;not null" json:"code"`
	Location    string         `gorm:"size:255" json:"location"`
	IsDefault   bool           `gorm:"default:false" json:"is_default"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	Description string         `gorm:"type:text" json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type WarehouseResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	Location    string `json:"location"`
	IsDefault   bool   `json:"is_default"`
	IsActive    bool   `json:"is_active"`
	Description string `json:"description"`
}
