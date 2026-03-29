package models

import (
	"time"

	"gorm.io/gorm"
)

type StockTransfer struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	TransferNumber  string         `gorm:"size:50;uniqueIndex;not null" json:"transfer_number"`
	FromWarehouseID uint           `gorm:"index;not null" json:"from_warehouse_id"`
	ToWarehouseID   uint           `gorm:"index;not null" json:"to_warehouse_id"`
	Status          string         `gorm:"size:20;default:'pending'" json:"status"`
	Notes           string         `gorm:"type:text" json:"notes"`
	CreatedBy       uint           `gorm:"index" json:"created_by"`
	ApprovedBy      *uint          `json:"approved_by"`
	ApprovedAt      *time.Time     `json:"approved_at"`
	CompletedBy     *uint          `json:"completed_by"`
	CompletedAt     *time.Time     `json:"completed_at"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
	FromWarehouse   *Warehouse     `gorm:"foreignKey:FromWarehouseID" json:"from_warehouse,omitempty"`
	ToWarehouse     *Warehouse     `gorm:"foreignKey:ToWarehouseID" json:"to_warehouse,omitempty"`
	Creator         *User          `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	Items           []TransferItem `gorm:"foreignKey:TransferID" json:"items"`
}

type TransferItem struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	TransferID  uint           `gorm:"index;not null" json:"transfer_id"`
	ProductID   uint           `gorm:"index;not null" json:"product_id"`
	Quantity    int            `gorm:"not null" json:"quantity"`
	ReceivedQty int            `gorm:"default:0" json:"received_qty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	Product     *Product       `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

const (
	TransferStatusPending   = "pending"
	TransferStatusApproved  = "approved"
	TransferStatusInTransit = "in_transit"
	TransferStatusCompleted = "completed"
	TransferStatusCancelled = "cancelled"
)

type TransferResponse struct {
	ID             uint                   `json:"id"`
	TransferNumber string                 `json:"transfer_number"`
	FromWarehouse  WarehouseResponse      `json:"from_warehouse"`
	ToWarehouse    WarehouseResponse      `json:"to_warehouse"`
	Status         string                 `json:"status"`
	Notes          string                 `json:"notes"`
	CreatedBy      uint                   `json:"created_by"`
	CreatorName    string                 `json:"creator_name"`
	ApprovedBy     *uint                  `json:"approved_by"`
	ApprovedAt     *time.Time             `json:"approved_at"`
	CompletedBy    *uint                  `json:"completed_by"`
	CompletedAt    *time.Time             `json:"completed_at"`
	Items          []TransferItemResponse `json:"items"`
	CreatedAt      time.Time              `json:"created_at"`
}

type TransferItemResponse struct {
	ID          uint   `json:"id"`
	ProductID   uint   `json:"product_id"`
	ProductName string `json:"product_name"`
	ProductSKU  string `json:"product_sku"`
	Quantity    int    `json:"quantity"`
	ReceivedQty int    `json:"received_qty"`
}
