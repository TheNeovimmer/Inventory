package models

import (
	"time"

	"gorm.io/gorm"
)

type Account struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"size:100;not null" json:"name"`
	Code        string         `gorm:"size:20;uniqueIndex;not null" json:"code"`
	AccountType string         `gorm:"size:20;not null" json:"account_type"` // asset, liability, income, expense, equity
	ParentID    *uint          `gorm:"index" json:"parent_id"`
	IsDefault   bool           `gorm:"default:false" json:"is_default"`
	Balance     float64        `gorm:"default:0" json:"balance"`
	Description string         `gorm:"type:text" json:"description"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	Parent      *Account       `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
}

type AccountTransaction struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Date          time.Time `json:"date"`
	Description   string    `gorm:"size:255;not null" json:"description"`
	Type          string    `gorm:"size:20;not null" json:"type"` // deposit, expense, transfer, adjustment
	Amount        float64   `gorm:"type:decimal(10,2);not null" json:"amount"`
	AccountID     uint      `gorm:"index;not null" json:"account_id"`
	ReferenceType string    `gorm:"size:50" json:"reference_type"`
	ReferenceID   *uint     `gorm:"index" json:"reference_id"`
	Reference     string    `gorm:"size:100" json:"reference"`
	Notes         string    `gorm:"type:text" json:"notes"`
	CreatedBy     uint      `gorm:"index" json:"created_by"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Account       *Account  `gorm:"foreignKey:AccountID" json:"account,omitempty"`
	Creator       *User     `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

type PaymentMethod struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:50;not null" json:"name"`
	Code      string    `gorm:"size:20;uniqueIndex;not null" json:"code"`
	IsDefault bool      `gorm:"default:false" json:"is_default"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

const (
	AccountTypeAsset     = "asset"
	AccountTypeLiability = "liability"
	AccountTypeIncome    = "income"
	AccountTypeExpense   = "expense"
	AccountTypeEquity    = "equity"
)

const (
	AccountTransactionTypeDeposit    = "deposit"
	AccountTransactionTypeExpense    = "expense"
	AccountTransactionTypeTransfer   = "transfer"
	AccountTransactionTypeAdjustment = "adjustment"
)

type AccountResponse struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Code        string  `json:"code"`
	AccountType string  `json:"account_type"`
	ParentID    *uint   `json:"parent_id"`
	Balance     float64 `json:"balance"`
	IsDefault   bool    `json:"is_default"`
	IsActive    bool    `json:"is_active"`
}

type AccountTransactionResponse struct {
	ID          uint    `json:"id"`
	Date        string  `json:"date"`
	Description string  `json:"description"`
	Type        string  `json:"type"`
	Amount      float64 `json:"amount"`
	AccountID   uint    `json:"account_id"`
	AccountName string  `json:"account_name"`
	Reference   string  `json:"reference"`
	Notes       string  `json:"notes"`
	CreatedBy   uint    `json:"created_by"`
	CreatorName string  `json:"creator_name"`
	CreatedAt   string  `json:"created_at"`
}
