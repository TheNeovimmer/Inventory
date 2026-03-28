package models

import (
	"time"

	"gorm.io/gorm"
)

type PurchaseOrder struct {
	ID           uint                `gorm:"primaryKey" json:"id"`
	SupplierID   uint                `gorm:"index;not null" json:"supplier_id"`
	Status       string              `gorm:"size:20;default:'pending'" json:"status"`
	OrderDate    time.Time           `json:"order_date"`
	ExpectedDate *time.Time          `json:"expected_date"`
	ReceivedDate *time.Time          `json:"received_date"`
	Total        float64             `gorm:"type:decimal(10,2);default:0" json:"total"`
	Notes        string              `gorm:"type:text" json:"notes"`
	CreatedBy    uint                `gorm:"index" json:"created_by"`
	CreatedAt    time.Time           `json:"created_at"`
	UpdatedAt    time.Time           `json:"updated_at"`
	DeletedAt    gorm.DeletedAt      `gorm:"index" json:"-"`
	Supplier     *Supplier           `gorm:"foreignKey:SupplierID" json:"supplier,omitempty"`
	Creator      *User               `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	Items        []PurchaseOrderItem `gorm:"foreignKey:POID" json:"items"`
}

type PurchaseOrderItem struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	POID        uint           `gorm:"index;not null" json:"po_id"`
	ProductID   uint           `gorm:"index;not null" json:"product_id"`
	Quantity    int            `gorm:"not null" json:"quantity"`
	UnitCost    float64        `gorm:"type:decimal(10,2);not null" json:"unit_cost"`
	ReceivedQty int            `gorm:"default:0" json:"received_qty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	Product     *Product       `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

const (
	POStatusPending   = "pending"
	POStatusOrdered   = "ordered"
	POStatusPartial   = "partial"
	POStatusReceived  = "received"
	POStatusCancelled = "cancelled"
)

type PurchaseOrderResponse struct {
	ID           uint                        `json:"id"`
	SupplierID   uint                        `json:"supplier_id"`
	SupplierName string                      `json:"supplier_name"`
	Status       string                      `json:"status"`
	OrderDate    time.Time                   `json:"order_date"`
	ExpectedDate *time.Time                  `json:"expected_date"`
	ReceivedDate *time.Time                  `json:"received_date"`
	Total        float64                     `json:"total"`
	Notes        string                      `json:"notes"`
	CreatedBy    uint                        `json:"created_by"`
	CreatorName  string                      `json:"creator_name"`
	Items        []PurchaseOrderItemResponse `json:"items"`
	CreatedAt    time.Time                   `json:"created_at"`
}

type PurchaseOrderItemResponse struct {
	ID          uint    `json:"id"`
	ProductID   uint    `json:"product_id"`
	ProductName string  `json:"product_name"`
	ProductSKU  string  `json:"product_sku"`
	Quantity    int     `json:"quantity"`
	UnitCost    float64 `json:"unit_cost"`
	ReceivedQty int     `json:"received_qty"`
	LineTotal   float64 `json:"line_total"`
}
