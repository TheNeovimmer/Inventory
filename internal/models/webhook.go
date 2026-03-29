package models

import (
	"time"

	"gorm.io/gorm"
)

type Webhook struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	Name       string         `gorm:"size:100;not null" json:"name"`
	URL        string         `gorm:"size:500;not null" json:"url"`
	Events     string         `gorm:"size:500" json:"events"` // comma-separated events
	Secret     string         `gorm:"size:255" json:"secret"`
	IsActive   bool           `gorm:"default:true" json:"is_active"`
	RetryCount int            `gorm:"default:3" json:"retry_count"`
	Timeout    int            `gorm:"default:30" json:"timeout"` // seconds
	CreatedBy  uint           `gorm:"index" json:"created_by"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
	Creator    *User          `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

type WebhookDelivery struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	WebhookID  uint      `gorm:"index;not null" json:"webhook_id"`
	Event      string    `gorm:"size:50;not null" json:"event"`
	Payload    string    `gorm:"type:text" json:"payload"`
	Response   string    `gorm:"type:text" json:"response"`
	StatusCode int       `json:"status_code"`
	Success    bool      `gorm:"default:false" json:"success"`
	ErrorMsg   string    `gorm:"type:text" json:"error_msg"`
	Attempt    int       `gorm:"default:1" json:"attempt"`
	CreatedAt  time.Time `json:"created_at"`
}

type WebhookResponse struct {
	ID         uint      `json:"id"`
	Name       string    `json:"name"`
	URL        string    `json:"url"`
	Events     []string  `json:"events"`
	IsActive   bool      `json:"is_active"`
	RetryCount int       `json:"retry_count"`
	Timeout    int       `json:"timeout"`
	CreatedBy  uint      `json:"created_by"`
	CreatedAt  time.Time `json:"created_at"`
}

// Webhook Events Constants
const (
	EventInventoryLow        = "inventory.low"
	EventInventoryAdjusted   = "inventory.adjusted"
	EventProductCreated      = "product.created"
	EventProductUpdated      = "product.updated"
	EventPOCreated           = "po.created"
	EventPOStatusChanged     = "po.status_changed"
	EventTransferCreated     = "transfer.created"
	EventTransferCompleted   = "transfer.completed"
	EventProductionStarted   = "production.started"
	EventProductionCompleted = "production.completed"
)
