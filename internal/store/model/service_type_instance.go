package model

import (
	"time"
)

type ServiceTypeInstance struct {
	ID            string         `gorm:"primaryKey;type:varchar(63)"`
	ProviderName  string         `gorm:"column:provider_name;not null"`
	Status        string         `gorm:"column:status;not null"`
	StatusMessage string         `gorm:"column:status_message"`
	InstanceName  string         `gorm:"column:instance_name;not null"`
	Spec          map[string]any `gorm:"column:spec;type:jsonb;serializer:json;not null"`
	CreateTime    time.Time      `gorm:"column:create_time;autoCreateTime"`
	UpdateTime    time.Time      `gorm:"column:update_time;autoUpdateTime"`
}

type ServiceTypeInstanceList []ServiceTypeInstance
