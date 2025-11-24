package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CatalogItem represents a predefined catalog offering that admins create
// Examples: "vm-small", "vm-large", "file-storage", "postgresql-ha"
type CatalogItem struct {
	gorm.Model
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name         string    `gorm:"name;not null;uniqueIndex"`
	DisplayName  string    `gorm:"display_name;not null"`
	Description  string    `gorm:"description"`
	ResourceKind string    `gorm:"resource_kind;not null;index"`
	Active       bool      `gorm:"active;not null;default:true"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// TableName specifies the table name for GORM
func (CatalogItem) TableName() string {
	return "catalog_items"
}

// CatalogProviderMapping represents which services can fulfill which catalog items
// This is the many-to-many relationship between catalog items and services
type CatalogProviderMapping struct {
	gorm.Model
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CatalogName  string    `gorm:"catalog_name;not null;index:idx_catalog_service"`
	ServiceID    string    `gorm:"service_id;not null;index:idx_catalog_service"`
	ResourceKind string    `gorm:"resource_kind;not null"`
	Endpoint     string    `gorm:"endpoint;not null"`
	Active       bool      `gorm:"active;not null;default:true"`
	RegisteredAt time.Time `gorm:"registered_at;not null"`
	UpdatedAt    time.Time `gorm:"updated_at;not null"`
}

// TableName specifies the table name for GORM
func (CatalogProviderMapping) TableName() string {
	return "catalog_provider_mappings"
}
