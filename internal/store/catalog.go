package store

import (
	"context"
	"time"

	"github.com/dcm-project/service-provider-api/internal/store/model"
	"gorm.io/gorm"
)

type Catalog interface {
	// CatalogItem operations
	GetCatalogItem(ctx context.Context, name string) (*model.CatalogItem, error)
	ListCatalogItems(ctx context.Context, active bool) ([]model.CatalogItem, error)
	ListAllCatalogItems(ctx context.Context) ([]model.CatalogItem, error)
	CreateCatalogItem(ctx context.Context, item *model.CatalogItem) error

	// CatalogProviderMapping operations
	GetCatalogMappings(ctx context.Context, catalogName string, active bool) ([]model.CatalogProviderMapping, error)
	ListAllCatalogMappings(ctx context.Context, active bool) ([]model.CatalogProviderMapping, error)
	GetDistinctResourceKinds(ctx context.Context) ([]string, error)
	UpsertCatalogMapping(ctx context.Context, mapping *model.CatalogProviderMapping) error
	DeactivateMappings(ctx context.Context, serviceID, resourceKind string) error
}

type CatalogStore struct {
	db *gorm.DB
}

var _ Catalog = (*CatalogStore)(nil)

func NewCatalog(db *gorm.DB) Catalog {
	return &CatalogStore{db: db}
}

func (s *CatalogStore) GetCatalogItem(ctx context.Context, name string) (*model.CatalogItem, error) {
	var item model.CatalogItem
	result := s.db.Where("name = ?", name).First(&item)
	if result.Error != nil {
		return nil, result.Error
	}
	return &item, nil
}

func (s *CatalogStore) ListCatalogItems(ctx context.Context, active bool) ([]model.CatalogItem, error) {
	var items []model.CatalogItem
	result := s.db.
		Where("active = ?", active).
		Order("resource_kind ASC, name ASC").
		Find(&items)
	if result.Error != nil {
		return nil, result.Error
	}
	return items, nil
}

func (s *CatalogStore) ListAllCatalogItems(ctx context.Context) ([]model.CatalogItem, error) {
	var items []model.CatalogItem
	result := s.db.Find(&items)
	if result.Error != nil {
		return nil, result.Error
	}
	return items, nil
}

func (s *CatalogStore) CreateCatalogItem(ctx context.Context, item *model.CatalogItem) error {
	result := s.db.Create(item)
	return result.Error
}

func (s *CatalogStore) GetCatalogMappings(ctx context.Context, catalogName string, active bool) ([]model.CatalogProviderMapping, error) {
	var mappings []model.CatalogProviderMapping
	result := s.db.
		Where("catalog_name = ? AND active = ?", catalogName, active).
		Find(&mappings)
	if result.Error != nil {
		return nil, result.Error
	}
	return mappings, nil
}

func (s *CatalogStore) ListAllCatalogMappings(ctx context.Context, active bool) ([]model.CatalogProviderMapping, error) {
	var mappings []model.CatalogProviderMapping
	result := s.db.
		Where("active = ?", active).
		Order("catalog_name ASC, service_id ASC").
		Find(&mappings)
	if result.Error != nil {
		return nil, result.Error
	}
	return mappings, nil
}

func (s *CatalogStore) GetDistinctResourceKinds(ctx context.Context) ([]string, error) {
	var resourceKinds []string
	result := s.db.Table("catalog_provider_mappings").
		Distinct("resource_kind").
		Pluck("resource_kind", &resourceKinds)
	if result.Error != nil {
		return nil, result.Error
	}
	return resourceKinds, nil
}

func (s *CatalogStore) UpsertCatalogMapping(ctx context.Context, mapping *model.CatalogProviderMapping) error {
	now := time.Now()
	result := s.db.
		Where("catalog_name = ? AND service_id = ? AND resource_kind = ?",
			mapping.CatalogName, mapping.ServiceID, mapping.ResourceKind).
		Assign(map[string]interface{}{
			"endpoint":   mapping.Endpoint,
			"active":     mapping.Active,
			"updated_at": now,
		}).
		FirstOrCreate(mapping)
	return result.Error
}

func (s *CatalogStore) DeactivateMappings(ctx context.Context, serviceID, resourceKind string) error {
	result := s.db.Model(&model.CatalogProviderMapping{}).
		Where("service_id = ? AND resource_kind = ?", serviceID, resourceKind).
		Update("active", false)
	return result.Error
}
