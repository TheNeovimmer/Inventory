package models

import (
	"time"
)

type Transaction struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	ProductID     uint      `gorm:"index;not null" json:"product_id"`
	Type          string    `gorm:"size:30;not null" json:"type"`
	Quantity      int       `gorm:"not null" json:"quantity"`
	ReferenceType string    `gorm:"size:50" json:"reference_type"`
	ReferenceID   *uint     `gorm:"index" json:"reference_id"`
	Notes         string    `gorm:"type:text" json:"notes"`
	UserID        uint      `gorm:"index" json:"user_id"`
	CreatedAt     time.Time `json:"created_at"`
	Product       *Product  `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	User          *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

const (
	TransactionTypePurchase      = "purchase"
	TransactionTypeSale          = "sale"
	TransactionTypeAdjustment    = "adjustment"
	TransactionTypeTransfer      = "transfer"
	TransactionTypeProductionIn  = "production_in"
	TransactionTypeProductionOut = "production_out"
	TransactionTypeReturn        = "return"
)

type TransactionResponse struct {
	ID            uint      `json:"id"`
	ProductID     uint      `json:"product_id"`
	ProductName   string    `json:"product_name"`
	ProductSKU    string    `json:"product_sku"`
	Type          string    `json:"type"`
	Quantity      int       `json:"quantity"`
	ReferenceType string    `json:"reference_type"`
	Notes         string    `json:"notes"`
	UserName      string    `json:"user_name"`
	CreatedAt     time.Time `json:"created_at"`
}
