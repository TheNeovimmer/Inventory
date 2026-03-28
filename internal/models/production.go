package models

import (
	"time"

	"gorm.io/gorm"
)

type ProductionOrder struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	BOMID       uint           `gorm:"index;not null" json:"bom_id"`
	Quantity    int            `gorm:"not null" json:"quantity"`
	Status      string         `gorm:"size:20;default:'draft'" json:"status"`
	StartedAt   *time.Time     `json:"started_at"`
	CompletedAt *time.Time     `json:"completed_at"`
	Notes       string         `gorm:"type:text" json:"notes"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	BOM         *BOM           `gorm:"foreignKey:BOMID" json:"bom,omitempty"`
}

const (
	ProductionStatusDraft      = "draft"
	ProductionStatusInProgress = "in_progress"
	ProductionStatusCompleted  = "completed"
	ProductionStatusCancelled  = "cancelled"
)

type ProductionOrderResponse struct {
	ID          uint       `json:"id"`
	BOMID       uint       `json:"bom_id"`
	ProductName string     `json:"product_name"`
	Quantity    int        `json:"quantity"`
	Status      string     `json:"status"`
	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
	Notes       string     `json:"notes"`
	CreatedAt   time.Time  `json:"created_at"`
}
