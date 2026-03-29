package models

import (
	"time"

	"gorm.io/gorm"
)

type Inventory struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	ProductID        uint           `gorm:"uniqueIndex:idx_product_warehouse;not null" json:"product_id"`
	WarehouseID      uint           `gorm:"uniqueIndex:idx_product_warehouse;default:1" json:"warehouse_id"`
	Quantity         int            `gorm:"default:0" json:"quantity"`
	ReservedQuantity int            `gorm:"default:0" json:"reserved_quantity"`
	LastUpdated      time.Time      `json:"last_updated"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
	Product          *Product       `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Warehouse        *Warehouse     `gorm:"foreignKey:WarehouseID" json:"warehouse,omitempty"`
}

func (i *Inventory) AvailableQuantity() int {
	return i.Quantity - i.ReservedQuantity
}

type InventoryBatch struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	InventoryID uint           `gorm:"index;not null" json:"inventory_id"`
	BatchNumber string         `gorm:"size:50;index" json:"batch_number"`
	Quantity    int            `gorm:"default:0" json:"quantity"`
	ExpiryDate  *time.Time     `json:"expiry_date"`
	Location    string         `gorm:"size:100" json:"location"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
