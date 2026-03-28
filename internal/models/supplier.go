package models

import (
	"time"

	"gorm.io/gorm"
)

type Supplier struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"size:200;not null" json:"name"`
	ContactName string         `gorm:"size:100" json:"contact_name"`
	Email       string         `gorm:"size:100" json:"email"`
	Phone       string         `gorm:"size:50" json:"phone"`
	Address     string         `gorm:"type:text" json:"address"`
	Notes       string         `gorm:"type:text" json:"notes"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type SupplierResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	ContactName string `json:"contact_name"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Address     string `json:"address"`
	Notes       string `json:"notes"`
	IsActive    bool   `json:"is_active"`
}
