package store

import (
	"gorm.io/gorm"
)

type Store interface {
	Close() error
	Application() Application
}

type DataStore struct {
	db          *gorm.DB
	application Application
}

func NewStore(db *gorm.DB) Store {
	return &DataStore{
		db:          db,
		application: NewApplication(db),
	}
}

func (s *DataStore) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (s *DataStore) Application() Application {
	return s.application
}
