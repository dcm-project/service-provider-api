package store

import (
	"context"

	"github.com/dcm-project/service-provider-api/internal/store/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Application interface {
	List(ctx context.Context) (model.ProviderList, error)
	Create(ctx context.Context, app model.Application) (*model.Application, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Get(ctx context.Context, id uuid.UUID) (*model.Application, error)
}

type ApplicationStore struct {
	db *gorm.DB
}

var _ Application = (*ApplicationStore)(nil)

func NewApplication(db *gorm.DB) Application {
	return &ApplicationStore{db: db}
}

func (s *ApplicationStore) List(ctx context.Context) (model.ApplicationList, error) {
	var apps model.ApplicationList
	tx := s.db.Model(&apps)
	result := tx.Find(&apps)
	if result.Error != nil {
		return nil, result.Error
	}
	return apps, nil
}

func (s *ApplicationStore) Delete(ctx context.Context, id uuid.UUID) error {
	result := s.db.Delete(&model.Application{}, id)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (s *ApplicationStore) Create(ctx context.Context, app model.Application) (*model.Application, error) {
	result := s.db.Clauses(clause.Returning{}).Create(&app)
	if result.Error != nil {
		return nil, result.Error
	}

	return &app, nil
}

func (s *ApplicationStore) Get(ctx context.Context, id uuid.UUID) (*model.Application, error) {
	var app model.Application
	result := s.db.First(&app, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &app, nil
}
