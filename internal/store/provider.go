package store

import (
	"context"

	"github.com/dcm-project/service-provider-api/internal/store/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Provider interface {
	List(ctx context.Context) (model.ProviderList, error)
	ListByType(ctx context.Context, providerType string) (model.ProviderList, error)
	Create(ctx context.Context, app model.Provider) (*model.Provider, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Update(ctx context.Context, app model.Provider) (*model.Provider, error)
	Get(ctx context.Context, id uuid.UUID) (*model.Provider, error)
}

type ProviderStore struct {
	db *gorm.DB
}

var _ Provider = (*ProviderStore)(nil)

func NewProvider(db *gorm.DB) Provider {
	return &ProviderStore{db: db}
}

func (s *ProviderStore) List(ctx context.Context) (model.ProviderList, error) {
	var provider model.ProviderList
	tx := s.db.Model(&provider)
	result := tx.Find(&provider)
	if result.Error != nil {
		return nil, result.Error
	}
	return provider, nil
}

func (s *ProviderStore) Delete(ctx context.Context, id uuid.UUID) error {
	result := s.db.Delete(&model.Provider{}, id)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (s *ProviderStore) Create(ctx context.Context, app model.Provider) (*model.Provider, error) {
	result := s.db.Clauses(clause.Returning{}).Create(&app)
	if result.Error != nil {
		return nil, result.Error
	}

	return &app, nil
}

func (s *ProviderStore) Get(ctx context.Context, id uuid.UUID) (*model.Provider, error) {
	var provider model.Provider
	result := s.db.First(&provider, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &provider, nil
}

func (s *ProviderStore) Update(ctx context.Context, provider model.Provider) (*model.Provider, error) {
	result := s.db.Clauses(clause.Returning{}).Save(&provider)
	if result.Error != nil {
		return nil, result.Error
	}
	return &provider, nil
}

func (s *ProviderStore) ListByType(ctx context.Context, providerType string) (model.ProviderList, error) {
	var provider model.ProviderList
	result := s.db.Where("provider_type = ?", providerType).Find(&provider)
	if result.Error != nil {
		return nil, result.Error
	}
	return provider, nil
}
