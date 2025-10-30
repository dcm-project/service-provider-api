package model

import (
	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Provider struct {
	gorm.Model
	ID           uuid.UUID      `gorm:"primaryKey;"`
	Name         string         `gorm:"name;not null"`
	ProviderType string         `gorm:"provider_type;not null"`
	Description  string         `gorm:"description;not null"`
	Endpoint     string         `gorm:"endpoint;not null"`
	Version      string         `gorm:"version;not null"`
	Capabilities pq.StringArray `gorm:"capabilities;type:text[]"`
}

type ProviderList []Provider
