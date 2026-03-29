package models

import (
	"time"
)

type Setting struct {
	Key       string    `gorm:"primaryKey" json:"key"`
	Value     string    `gorm:"type:text" json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SettingGroup struct {
	Key         string
	Name        string
	Description string
	Settings    []SettingItem
}

type SettingItem struct {
	Key         string
	Name        string
	Type        string // text, number, boolean, select, textarea
	Value       string
	Options     []string
	Required    bool
	Description string
}

var DefaultSettings = []SettingGroup{
	{
		Key:         "company",
		Name:        "Company Information",
		Description: "Basic company details",
		Settings: []SettingItem{
			{Key: "company_name", Name: "Company Name", Type: "text", Value: "My Company"},
			{Key: "company_email", Name: "Email", Type: "text", Value: ""},
			{Key: "company_phone", Name: "Phone", Type: "text", Value: ""},
			{Key: "company_address", Name: "Address", Type: "textarea", Value: ""},
			{Key: "company_tax", Name: "Tax Number", Type: "text", Value: ""},
			{Key: "company_logo", Name: "Logo URL", Type: "text", Value: ""},
		},
	},
	{
		Key:         "currency",
		Name:        "Currency Settings",
		Description: "Currency and number formatting",
		Settings: []SettingItem{
			{Key: "currency_symbol", Name: "Currency Symbol", Type: "text", Value: "$"},
			{Key: "currency_position", Name: "Symbol Position", Type: "select", Value: "before", Options: []string{"before", "after"}},
			{Key: "decimal_places", Name: "Decimal Places", Type: "number", Value: "2"},
			{Key: "thousand_separator", Name: "Thousand Separator", Type: "select", Value: ",", Options: []string{",", ".", " "}},
		},
	},
	{
		Key:         "tax",
		Name:        "Tax Settings",
		Description: "Tax configuration",
		Settings: []SettingItem{
			{Key: "default_tax_rate", Name: "Default Tax Rate (%)", Type: "number", Value: "0"},
			{Key: "tax_number", Name: "Tax Number/ID", Type: "text", Value: ""},
		},
	},
	{
		Key:         "invoice",
		Name:        "Invoice Settings",
		Description: "Invoice numbering and templates",
		Settings: []SettingItem{
			{Key: "invoice_prefix", Name: "Invoice Prefix", Type: "text", Value: "INV"},
			{Key: "invoice_starting_number", Name: "Starting Number", Type: "number", Value: "1"},
			{Key: "quotation_prefix", Name: "Quotation Prefix", Type: "text", Value: "QUO"},
			{Key: "invoice_footer", Name: "Invoice Footer", Type: "textarea", Value: ""},
		},
	},
	{
		Key:         "notification",
		Name:        "Notification Settings",
		Description: "Email and SMS notifications",
		Settings: []SettingItem{
			{Key: "notify_sale", Name: "Notify on Sale", Type: "boolean", Value: "false"},
			{Key: "notify_purchase", Name: "Notify on Purchase", Type: "boolean", Value: "false"},
			{Key: "notify_quotation", Name: "Notify on Quotation", Type: "boolean", Value: "false"},
			{Key: "notify_low_stock", Name: "Notify Low Stock", Type: "boolean", Value: "false"},
		},
	},
	{
		Key:         "system",
		Name:        "System Settings",
		Description: "General system configuration",
		Settings: []SettingItem{
			{Key: "timezone", Name: "Timezone", Type: "text", Value: "UTC"},
			{Key: "language", Name: "Language", Type: "select", Value: "en", Options: []string{"en", "fr", "ar", "es"}},
			{Key: "auto_backup", Name: "Auto Backup", Type: "boolean", Value: "false"},
		},
	},
}
