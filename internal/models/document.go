package models

import (
	"time"

	"gorm.io/gorm"
)

type Document struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	EntityType   string         `gorm:"size:50;index;not null" json:"entity_type"` // product, purchase_order, supplier
	EntityID     uint           `gorm:"index;not null" json:"entity_id"`
	FileName     string         `gorm:"size:255;not null" json:"file_name"`
	OriginalName string         `gorm:"size:255;not null" json:"original_name"`
	FileType     string         `gorm:"size:50;not null" json:"file_type"`
	FileSize     int64          `json:"file_size"`
	FilePath     string         `gorm:"size:500;not null" json:"file_path"`
	Description  string         `gorm:"size:500" json:"description"`
	UploadedBy   uint           `gorm:"index" json:"uploaded_by"`
	CreatedAt    time.Time      `json:"created_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	Uploader     *User          `gorm:"foreignKey:UploadedBy" json:"uploader,omitempty"`
}

type DocumentResponse struct {
	ID           uint      `json:"id"`
	EntityType   string    `json:"entity_type"`
	EntityID     uint      `json:"entity_id"`
	FileName     string    `json:"file_name"`
	OriginalName string    `json:"original_name"`
	FileType     string    `json:"file_type"`
	FileSize     int64     `json:"file_size"`
	Description  string    `json:"description"`
	UploaderName string    `json:"uploader_name"`
	CreatedAt    time.Time `json:"created_at"`
}
