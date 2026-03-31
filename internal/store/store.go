// Package store provides database access interfaces and implementations.
package store

import (
	providerstore "github.com/dcm-project/service-provider-manager/internal/store/provider"
	rmstore "github.com/dcm-project/service-provider-manager/internal/store/resource_manager"
	"gorm.io/gorm"
)

type Store interface {
	Close() error
	Provider() providerstore.Provider
	ServiceTypeInstance() rmstore.ServiceTypeInstance
}

type DataStore struct {
	db       *gorm.DB
	provider providerstore.Provider
	instance rmstore.ServiceTypeInstance
}

func NewStore(db *gorm.DB) Store {
	return &DataStore{
		db:       db,
		provider: providerstore.NewProvider(db),
		instance: rmstore.NewServiceTypeInstance(db),
	}
}

func (s *DataStore) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (s *DataStore) Provider() providerstore.Provider {
	return s.provider
}

func (s *DataStore) ServiceTypeInstance() rmstore.ServiceTypeInstance {
	return s.instance
}
