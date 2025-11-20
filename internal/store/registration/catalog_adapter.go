package registration

import (
	"context"
	"time"

	"github.com/dcm-project/service-provider-api/internal/store"
	"github.com/dcm-project/service-provider-api/internal/store/model"
	"go.uber.org/zap"
)

// RegistrationCatalogAdapter implements registration.CatalogStore
type RegistrationCatalogAdapter struct {
	store store.Store
}

// NewRegistrationCatalogAdapter creates a new catalog adapter
func NewRegistrationCatalogAdapter(s store.Store) *RegistrationCatalogAdapter {
	return &RegistrationCatalogAdapter{store: s}
}

// UpdateCatalogMapping updates which services can fulfill which catalog items
func (a *RegistrationCatalogAdapter) UpdateCatalogMapping(ctx context.Context, serviceID, resourceKind string, catalogItem string) error {
	logger := zap.S().Named("catalog_adapter")

	// Get the service's endpoint from registry using the adapter
	registryAdapter := NewRegistrationRegistryAdapter(a.store)
	provider, err := registryAdapter.GetProvider(ctx, serviceID, resourceKind)
	if err != nil {
		logger.Warnw("Service not found in registry, using placeholder endpoint",
			"service_id", serviceID,
			"resource_kind", resourceKind,
			"error", err,
		)
	}

	endpoint := ""
	if provider != nil {
		endpoint = provider.Endpoint
	}

	now := time.Now()

	// First, mark existing mappings as inactive for this service + resource kind
	if err := a.store.Catalog().DeactivateMappings(ctx, serviceID, resourceKind); err != nil {
		logger.Errorw("Failed to deactivate old catalog mappings", "error", err)
		return err
	}

	// Create or update catalog mapping for the catalog item
	mapping := model.CatalogProviderMapping{
		CatalogName:  catalogItem,
		ServiceID:    serviceID,
		ResourceKind: resourceKind,
		Endpoint:     endpoint,
		Active:       true,
		RegisteredAt: now,
		UpdatedAt:    now,
	}

	// Upsert: try to update existing, or insert new
	if err := a.store.Catalog().UpsertCatalogMapping(ctx, &mapping); err != nil {
		logger.Errorw("Failed to upsert catalog item",
			"catalog_name", catalogItem,
			"service_id", serviceID,
			"error", err,
		)
		return err
	}

	logger.Infow("Updated catalog mapping",
		"service_id", serviceID,
		"resource_kind", resourceKind,
		"catalog_item", catalogItem,
	)

	return nil
}

// RemoveCatalogMapping removes service mappings from catalog
func (a *RegistrationCatalogAdapter) RemoveCatalogMapping(ctx context.Context, serviceID, resourceKind string) error {
	logger := zap.S().Named("catalog_adapter")

	// Mark all mappings as inactive
	if err := a.store.Catalog().DeactivateMappings(ctx, serviceID, resourceKind); err != nil {
		logger.Errorw("Failed to remove catalog mapping", "error", err)
		return err
	}

	logger.Infow("Removed catalog mapping",
		"service_id", serviceID,
		"resource_kind", resourceKind,
	)

	return nil
}

// GetProvidersForCatalogItem returns all active providers that can fulfill a catalog item
func (a *RegistrationCatalogAdapter) GetProvidersForCatalogItem(ctx context.Context, catalogName string) ([]model.CatalogProviderMapping, error) {
	return a.store.Catalog().GetCatalogMappings(ctx, catalogName, true)
}

// GetAllCatalogMappings returns all active catalog provider mappings
func (a *RegistrationCatalogAdapter) GetAllCatalogMappings(ctx context.Context) ([]model.CatalogProviderMapping, error) {
	return a.store.Catalog().ListAllCatalogMappings(ctx, true)
}

// GetAllCatalogItems returns all predefined catalog items
func (a *RegistrationCatalogAdapter) GetAllCatalogItems(ctx context.Context) ([]model.CatalogItem, error) {
	return a.store.Catalog().ListCatalogItems(ctx, true)
}

// CreateCatalogItem creates a new catalog item (admin operation)
func (a *RegistrationCatalogAdapter) CreateCatalogItem(ctx context.Context, item *model.CatalogItem) error {
	return a.store.Catalog().CreateCatalogItem(ctx, item)
}
