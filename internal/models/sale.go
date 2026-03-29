package models

import (
	"time"

	"gorm.io/gorm"
)

type Sale struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	InvoiceNumber   string         `gorm:"size:50;uniqueIndex;not null" json:"invoice_number"`
	ReferenceNumber string         `gorm:"size:50" json:"reference_number"`
	CustomerID      *uint          `gorm:"index" json:"customer_id"`
	WarehouseID     uint           `gorm:"default:1" json:"warehouse_id"`
	Subtotal        float64        `gorm:"type:decimal(10,2);not null" json:"subtotal"`
	TaxAmount       float64        `gorm:"type:decimal(10,2);default:0" json:"tax_amount"`
	DiscountAmount  float64        `gorm:"type:decimal(10,2);default:0" json:"discount_amount"`
	ShippingAmount  float64        `gorm:"type:decimal(10,2);default:0" json:"shipping_amount"`
	Total           float64        `gorm:"type:decimal(10,2);not null" json:"total"`
	Status          string         `gorm:"size:20;default:'completed'" json:"status"`
	PaymentStatus   string         `gorm:"size:20;default:'paid'" json:"payment_status"`
	PaymentMethod   string         `gorm:"size:50" json:"payment_method"`
	PaidAmount      float64        `gorm:"type:decimal(10,2);default:0" json:"paid_amount"`
	DueAmount       float64        `gorm:"type:decimal(10,2);default:0" json:"due_amount"`
	TaxRate         float64        `gorm:"default:0" json:"tax_rate"`
	Notes           string         `gorm:"type:text" json:"notes"`
	CreatedBy       uint           `gorm:"index" json:"created_by"`
	SaleDate        time.Time      `json:"sale_date"`
	DueDate         *time.Time     `json:"due_date"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
	Customer        *Customer      `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	Creator         *User          `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	Warehouse       *Warehouse     `gorm:"foreignKey:WarehouseID" json:"warehouse,omitempty"`
	Items           []SaleItem     `gorm:"foreignKey:SaleID" json:"items"`
	Payments        []SalePayment  `gorm:"foreignKey:SaleID" json:"payments"`
}

type SaleItem struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	SaleID      uint      `gorm:"index;not null" json:"sale_id"`
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
	Product     *Product  `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

type SalePayment struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	SaleID        uint      `gorm:"index;not null" json:"sale_id"`
	PaymentMethod string    `gorm:"size:50" json:"payment_method"`
	Amount        float64   `gorm:"type:decimal(10,2);not null" json:"amount"`
	Reference     string    `gorm:"size:100" json:"reference"`
	Notes         string    `gorm:"type:text" json:"notes"`
	CreatedAt     time.Time `json:"created_at"`
}

const (
	SaleStatusPending   = "pending"
	SaleStatusCompleted = "completed"
	SaleStatusCancelled = "cancelled"
	SaleStatusReturned  = "returned"
	SaleStatusDraft     = "draft"
)

const (
	PaymentStatusPaid    = "paid"
	PaymentStatusPartial = "partial"
	PaymentStatusDue     = "due"
)

type SaleResponse struct {
	ID             uint                  `json:"id"`
	InvoiceNumber  string                `json:"invoice_number"`
	CustomerID     *uint                 `json:"customer_id"`
	CustomerName   string                `json:"customer_name"`
	Subtotal       float64               `json:"subtotal"`
	TaxAmount      float64               `json:"tax_amount"`
	DiscountAmount float64               `json:"discount_amount"`
	ShippingAmount float64               `json:"shipping_amount"`
	Total          float64               `json:"total"`
	Status         string                `json:"status"`
	PaymentStatus  string                `json:"payment_status"`
	PaymentMethod  string                `json:"payment_method"`
	PaidAmount     float64               `json:"paid_amount"`
	DueAmount      float64               `json:"due_amount"`
	Notes          string                `json:"notes"`
	CreatedBy      uint                  `json:"created_by"`
	CreatorName    string                `json:"creator_name"`
	SaleDate       string                `json:"sale_date"`
	Items          []SaleItemResponse    `json:"items"`
	Payments       []SalePaymentResponse `json:"payments"`
	CreatedAt      string                `json:"created_at"`
}

type SaleItemResponse struct {
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

type SalePaymentResponse struct {
	ID            uint    `json:"id"`
	PaymentMethod string  `json:"payment_method"`
	Amount        float64 `json:"amount"`
	Reference     string  `json:"reference"`
	Notes         string  `json:"notes"`
	CreatedAt     string  `json:"created_at"`
}
