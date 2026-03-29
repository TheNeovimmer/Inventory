package models

import (
	"time"

	"gorm.io/gorm"
)

type AuditLog struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	EntityType  string    `gorm:"size:50;index;not null" json:"entity_type"`
	EntityID    uint      `gorm:"index;not null" json:"entity_id"`
	Action      string    `gorm:"size:20;not null" json:"action"`
	OldValue    string    `gorm:"type:text" json:"old_value"`
	NewValue    string    `gorm:"type:text" json:"new_value"`
	Description string    `gorm:"size:500" json:"description"`
	UserID      uint      `gorm:"index" json:"user_id"`
	IPAddress   string    `gorm:"size:50" json:"ip_address"`
	UserAgent   string    `gorm:"size:255" json:"user_agent"`
	CreatedAt   time.Time `json:"created_at"`
	User        *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

type AuditCycle struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Title       string         `gorm:"size:200;not null" json:"title"`
	WarehouseID uint           `gorm:"index" json:"warehouse_id"`
	Status      string         `gorm:"size:20;default:'planned'" json:"status"`
	StartDate   *time.Time     `json:"start_date"`
	EndDate     *time.Time     `json:"end_date"`
	Notes       string         `gorm:"type:text" json:"notes"`
	CreatedBy   uint           `gorm:"index" json:"created_by"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	Warehouse   *Warehouse     `gorm:"foreignKey:WarehouseID" json:"warehouse,omitempty"`
	Creator     *User          `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	Items       []AuditItem    `gorm:"foreignKey:AuditCycleID" json:"items"`
}

type AuditItem struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	AuditCycleID uint           `gorm:"index;not null" json:"audit_cycle_id"`
	ProductID    uint           `gorm:"index;not null" json:"product_id"`
	SystemQty    int            `gorm:"default:0" json:"system_qty"`
	CountedQty   int            `gorm:"default:0" json:"counted_qty"`
	Variance     int            `gorm:"default:0" json:"variance"`
	Status       string         `gorm:"size:20;default:'pending'" json:"status"`
	Notes        string         `gorm:"type:text" json:"notes"`
	CountedBy    *uint          `json:"counted_by"`
	CountedAt    *time.Time     `json:"counted_at"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	AuditCycle   *AuditCycle    `gorm:"foreignKey:AuditCycleID" json:"audit_cycle,omitempty"`
	Product      *Product       `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

const (
	AuditStatusPlanned      = "planned"
	AuditStatusInProgress   = "in_progress"
	AuditStatusCompleted    = "completed"
	AuditStatusCancelled    = "cancelled"
	AuditItemStatusPending  = "pending"
	AuditItemStatusCounted  = "counted"
	AuditItemStatusVerified = "verified"
	AuditItemStatusAdjusted = "adjusted"
)

const (
	AuditActionCreate = "create"
	AuditActionUpdate = "update"
	AuditActionDelete = "delete"
	AuditActionLogin  = "login"
	AuditActionLogout = "logout"
	AuditActionExport = "export"
	AuditActionImport = "import"
)

type AuditLogResponse struct {
	ID          uint      `json:"id"`
	EntityType  string    `json:"entity_type"`
	EntityID    uint      `json:"entity_id"`
	Action      string    `json:"action"`
	OldValue    string    `json:"old_value"`
	Description string    `json:"description"`
	UserName    string    `json:"user_name"`
	IPAddress   string    `json:"ip_address"`
	CreatedAt   time.Time `json:"created_at"`
}

type AuditCycleResponse struct {
	ID            uint       `json:"id"`
	Title         string     `json:"title"`
	WarehouseID   uint       `json:"warehouse_id"`
	WarehouseName string     `json:"warehouse_name"`
	Status        string     `json:"status"`
	StartDate     *time.Time `json:"start_date"`
	EndDate       *time.Time `json:"end_date"`
	Notes         string     `json:"notes"`
	CreatorName   string     `json:"creator_name"`
	ItemCount     int        `json:"item_count"`
	VerifiedCount int        `json:"verified_count"`
	CreatedAt     time.Time  `json:"created_at"`
}

type AuditItemResponse struct {
	ID          uint       `json:"id"`
	ProductID   uint       `json:"product_id"`
	ProductName string     `json:"product_name"`
	ProductSKU  string     `json:"product_sku"`
	SystemQty   int        `json:"system_qty"`
	CountedQty  int        `json:"counted_qty"`
	Variance    int        `json:"variance"`
	Status      string     `json:"status"`
	Notes       string     `json:"notes"`
	CountedBy   *uint      `json:"counted_by"`
	CountedAt   *time.Time `json:"counted_at"`
}
