package models

import (
	"time"

	"gorm.io/gorm"
)

type Quotation struct {
	ID              uint            `gorm:"primaryKey" json:"id"`
	QuotationNumber string          `gorm:"size:50;uniqueIndex;not null" json:"quotation_number"`
	CustomerID      *uint           `gorm:"index" json:"customer_id"`
	WarehouseID     uint            `gorm:"default:1" json:"warehouse_id"`
	Subtotal        float64         `gorm:"type:decimal(10,2);not null" json:"subtotal"`
	TaxAmount       float64         `gorm:"type:decimal(10,2);default:0" json:"tax_amount"`
	TaxRate         float64         `gorm:"default:0" json:"tax_rate"`
	DiscountAmount  float64         `gorm:"type:decimal(10,2);default:0" json:"discount_amount"`
	Total           float64         `gorm:"type:decimal(10,2);not null" json:"total"`
	Status          string          `gorm:"size:20;default:'draft'" json:"status"`
	ValidUntil      *time.Time      `json:"valid_until"`
	Notes           string          `gorm:"type:text" json:"notes"`
	Terms           string          `gorm:"type:text" json:"terms"`
	CreatedBy       uint            `gorm:"index" json:"created_by"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	DeletedAt       gorm.DeletedAt  `gorm:"index" json:"-"`
	Customer        *Customer       `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	Creator         *User           `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	Items           []QuotationItem `gorm:"foreignKey:QuotationID" json:"items"`
}

type QuotationItem struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	QuotationID uint      `gorm:"index;not null" json:"quotation_id"`
	ProductID   uint      `gorm:"index;not null" json:"product_id"`
	ProductName string    `gorm:"size:200" json:"product_name"`
	ProductSKU  string    `gorm:"size:50" json:"product_sku"`
	Quantity    int       `gorm:"not null" json:"quantity"`
	UnitPrice   float64   `gorm:"type:decimal(10,2);not null" json:"unit_price"`
	Discount    float64   `gorm:"type:decimal(10,2);default:0" json:"discount"`
	TaxRate     float64   `gorm:"default:0" json:"tax_rate"`
	TaxAmount   float64   `gorm:"type:decimal(10,2);default:0" json:"tax_amount"`
	Total       float64   `gorm:"type:decimal(10,2);not null" json:"total"`
	CreatedAt   time.Time `json:"created_at"`
}

const (
	QuotationStatusDraft     = "draft"
	QuotationStatusSent      = "sent"
	QuotationStatusAccepted  = "accepted"
	QuotationStatusRejected  = "rejected"
	QuotationStatusExpired   = "expired"
	QuotationStatusConverted = "converted"
)

type QuotationResponse struct {
	ID              uint                    `json:"id"`
	QuotationNumber string                  `json:"quotation_number"`
	CustomerID      *uint                   `json:"customer_id"`
	CustomerName    string                  `json:"customer_name"`
	Subtotal        float64                 `json:"subtotal"`
	TaxAmount       float64                 `json:"tax_amount"`
	DiscountAmount  float64                 `json:"discount_amount"`
	Total           float64                 `json:"total"`
	Status          string                  `json:"status"`
	ValidUntil      *time.Time              `json:"valid_until"`
	Notes           string                  `json:"notes"`
	Terms           string                  `json:"terms"`
	CreatedBy       uint                    `json:"created_by"`
	CreatorName     string                  `json:"creator_name"`
	Items           []QuotationItemResponse `json:"items"`
	CreatedAt       string                  `json:"created_at"`
}

type QuotationItemResponse struct {
	ID          uint    `json:"id"`
	ProductID   uint    `json:"product_id"`
	ProductName string  `json:"product_name"`
	ProductSKU  string  `json:"product_sku"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	Discount    float64 `json:"discount"`
	TaxRate     float64 `json:"tax_rate"`
	TaxAmount   float64 `json:"tax_amount"`
	Total       float64 `json:"total"`
}
