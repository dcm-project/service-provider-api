package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Provider struct {
	gorm.Model
	ID                 uuid.UUID `gorm:"primaryKey;"`
	Name               string    `gorm:"name;not null"`
	ServiceType        string    `gorm:"not null"`
	ServiceDescription string    `gorm:"not null"`
	// TODO
	//Config        object            `gorm:"config;not null"`
}

type ProviderList []Provider
