package models

import (
	"time"

	"gorm.io/gorm"
)

type SaleReturn struct {
	ID           uint             `gorm:"primaryKey" json:"id"`
	ReturnNumber string           `gorm:"size:50;uniqueIndex;not null" json:"return_number"`
	SaleID       uint             `gorm:"index;not null" json:"sale_id"`
	CustomerID   *uint            `gorm:"index" json:"customer_id"`
	Subtotal     float64          `gorm:"type:decimal(10,2);not null" json:"subtotal"`
	TaxAmount    float64          `gorm:"type:decimal(10,2);default:0" json:"tax_amount"`
	Total        float64          `gorm:"type:decimal(10,2);not null" json:"total"`
	Reason       string           `gorm:"type:text" json:"reason"`
	Status       string           `gorm:"size:20;default:'pending'" json:"status"`
	RefundAmount float64          `gorm:"type:decimal(10,2);default:0" json:"refund_amount"`
	RefundMethod string           `gorm:"size:50" json:"refund_method"`
	Notes        string           `gorm:"type:text" json:"notes"`
	CreatedBy    uint             `gorm:"index" json:"created_by"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
	DeletedAt    gorm.DeletedAt   `gorm:"index" json:"-"`
	Sale         *Sale            `gorm:"foreignKey:SaleID" json:"sale,omitempty"`
	Customer     *Customer        `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	Creator      *User            `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	Items        []SaleReturnItem `gorm:"foreignKey:ReturnID" json:"items"`
}

type SaleReturnItem struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ReturnID    uint      `gorm:"index;not null" json:"return_id"`
	SaleItemID  uint      `gorm:"index" json:"sale_item_id"`
	ProductID   uint      `gorm:"index;not null" json:"product_id"`
	ProductName string    `gorm:"size:200" json:"product_name"`
	Quantity    int       `gorm:"not null" json:"quantity"`
	UnitPrice   float64   `gorm:"type:decimal(10,2);not null" json:"unit_price"`
	Total       float64   `gorm:"type:decimal(10,2);not null" json:"total"`
	Reason      string    `gorm:"type:text" json:"reason"`
	CreatedAt   time.Time `json:"created_at"`
}

type PurchaseReturn struct {
	ID              uint                 `gorm:"primaryKey" json:"id"`
	ReturnNumber    string               `gorm:"size:50;uniqueIndex;not null" json:"return_number"`
	PurchaseOrderID uint                 `gorm:"index;not null" json:"purchase_order_id"`
	SupplierID      uint                 `gorm:"index" json:"supplier_id"`
	Subtotal        float64              `gorm:"type:decimal(10,2);not null" json:"subtotal"`
	TaxAmount       float64              `gorm:"type:decimal(10,2);default:0" json:"tax_amount"`
	Total           float64              `gorm:"type:decimal(10,2);not null" json:"total"`
	Reason          string               `gorm:"type:text" json:"reason"`
	Status          string               `gorm:"size:20;default:'pending'" json:"status"`
	RefundAmount    float64              `gorm:"type:decimal(10,2);default:0" json:"refund_amount"`
	Notes           string               `gorm:"type:text" json:"notes"`
	CreatedBy       uint                 `gorm:"index" json:"created_by"`
	CreatedAt       time.Time            `json:"created_at"`
	UpdatedAt       time.Time            `json:"updated_at"`
	DeletedAt       gorm.DeletedAt       `gorm:"index" json:"-"`
	PurchaseOrder   *PurchaseOrder       `gorm:"foreignKey:PurchaseOrderID" json:"purchase_order,omitempty"`
	Supplier        *Supplier            `gorm:"foreignKey:SupplierID" json:"supplier,omitempty"`
	Creator         *User                `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	Items           []PurchaseReturnItem `gorm:"foreignKey:ReturnID" json:"items"`
}

type PurchaseReturnItem struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ReturnID    uint      `gorm:"index;not null" json:"return_id"`
	ProductID   uint      `gorm:"index;not null" json:"product_id"`
	ProductName string    `gorm:"size:200" json:"product_name"`
	Quantity    int       `gorm:"not null" json:"quantity"`
	UnitPrice   float64   `gorm:"type:decimal(10,2);not null" json:"unit_price"`
	Total       float64   `gorm:"type:decimal(10,2);not null" json:"total"`
	Reason      string    `gorm:"type:text" json:"reason"`
	CreatedAt   time.Time `json:"created_at"`
}

const (
	ReturnStatusPending   = "pending"
	ReturnStatusApproved  = "approved"
	ReturnStatusRejected  = "rejected"
	ReturnStatusCompleted = "completed"
)

const (
	RefundMethodCash    = "cash"
	RefundMethodCard    = "card"
	RefundMethodAccount = "account"
)
