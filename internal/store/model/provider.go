package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Provider struct {
	gorm.Model
	ID           uuid.UUID `gorm:"primaryKey;"`
	Name         string    `gorm:"name;not null"`
	ProviderType string    `gorm:"providertype;not null"`
	Description  string    `gorm:"description;not null"`
	Endpoint     string    `gorm:"endpoint;not null"`
}

type ProviderList []Provider
