package store

import (
	"gorm.io/gorm"
)

type Store interface {
	Close() error
	Application() ProviderApplication
	Provider() Provider
	Catalog() Catalog
}

type DataStore struct {
	db          *gorm.DB
	application ProviderApplication
	provider    Provider
	catalog     Catalog
}

func NewStore(db *gorm.DB) Store {
	return &DataStore{
		db:          db,
		application: NewProviderApplication(db),
		provider:    NewProvider(db),
		catalog:     NewCatalog(db),
	}
}

func (s *DataStore) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (s *DataStore) Application() ProviderApplication {
	return s.application
}

func (s *DataStore) Provider() Provider {
	return s.provider
}

func (s *DataStore) Catalog() Catalog {
	return s.catalog
}
