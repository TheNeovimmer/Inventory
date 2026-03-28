package models

import (
	"time"

	"gorm.io/gorm"
)

type Inventory struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	ProductID        uint           `gorm:"uniqueIndex;not null" json:"product_id"`
	Quantity         int            `gorm:"default:0" json:"quantity"`
	ReservedQuantity int            `gorm:"default:0" json:"reserved_quantity"`
	LastUpdated      time.Time      `json:"last_updated"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
	Product          *Product       `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

func (i *Inventory) AvailableQuantity() int {
	return i.Quantity - i.ReservedQuantity
}
