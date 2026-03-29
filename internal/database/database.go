package database

import (
	"inventory-ims/internal/config"
	"inventory-ims/internal/models"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect(cfg *config.Config) error {
	var err error

	logLevel := logger.Silent
	if cfg.Environment == "development" {
		logLevel = logger.Info
	}

	DB, err = gorm.Open(sqlite.Open(cfg.DatabasePath), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return err
	}

	if err = AutoMigrate(); err != nil {
		return err
	}

	log.Println("Database connected successfully")
	return nil
}

func AutoMigrate() error {
	return DB.AutoMigrate(
		&models.User{},
		&models.Category{},
		&models.Brand{},
		&models.Unit{},
		&models.Product{},
		&models.ProductVariant{},
		&models.Warehouse{},
		&models.Inventory{},
		&models.InventoryBatch{},
		&models.BOM{},
		&models.BOMItem{},
		&models.ProductionOrder{},
		&models.Transaction{},
		&models.Supplier{},
		&models.PurchaseOrder{},
		&models.PurchaseOrderItem{},
		&models.Customer{},
		&models.Sale{},
		&models.SaleItem{},
		&models.SalePayment{},
		&models.Quotation{},
		&models.QuotationItem{},
		&models.SaleReturn{},
		&models.SaleReturnItem{},
		&models.PurchaseReturn{},
		&models.PurchaseReturnItem{},
		&models.StockTransfer{},
		&models.TransferItem{},
		&models.AuditCycle{},
		&models.AuditItem{},
		&models.AuditLog{},
		&models.Account{},
		&models.AccountTransaction{},
		&models.PaymentMethod{},
		&models.Setting{},
		&models.Webhook{},
		&models.WebhookDelivery{},
		&models.Document{},
	)
}
