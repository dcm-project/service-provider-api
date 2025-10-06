package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProviderApplication struct {
	gorm.Model
	ID         uuid.UUID `gorm:"primaryKey;"`
	ProviderID uuid.UUID `gorm:"not null;"` // foreign key
	// TODO
	//Config        object            `gorm:"config;not null"`
}

type ProviderApplicationList []ProviderApplication
