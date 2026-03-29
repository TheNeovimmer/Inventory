package models

import (
	"time"

	"gorm.io/gorm"
)

type Customer struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"size:200;not null" json:"name"`
	Email       string         `gorm:"size:100" json:"email"`
	Phone       string         `gorm:"size:50" json:"phone"`
	Address     string         `gorm:"type:text" json:"address"`
	City        string         `gorm:"size:100" json:"city"`
	State       string         `gorm:"size:100" json:"state"`
	Country     string         `gorm:"size:100" json:"country"`
	PostalCode  string         `gorm:"size:20" json:"postal_code"`
	TaxNumber   string         `gorm:"size:50" json:"tax_number"`
	Balance     float64        `gorm:"default:0" json:"balance"`
	CreditLimit float64        `gorm:"default:0" json:"credit_limit"`
	Points      int            `gorm:"default:0" json:"points"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	Notes       string         `gorm:"type:text" json:"notes"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type CustomerResponse struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Email       string  `json:"email"`
	Phone       string  `json:"phone"`
	Address     string  `json:"address"`
	City        string  `json:"city"`
	State       string  `json:"state"`
	Country     string  `json:"country"`
	TaxNumber   string  `json:"tax_number"`
	Balance     float64 `json:"balance"`
	CreditLimit float64 `json:"credit_limit"`
	Points      int     `json:"points"`
	IsActive    bool    `json:"is_active"`
	Notes       string  `json:"notes"`
	CreatedAt   string  `json:"created_at"`
}
