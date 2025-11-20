package registration

import (
	"context"
	"fmt"

	"github.com/dcm-project/service-provider-api/internal/store"
	"github.com/dcm-project/service-provider-api/internal/store/model"
	"github.com/dcm-project/service-provider-api/pkg/registration"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// RegistrationRegistryAdapter adapts the Store to implement registration.RegistryStore
type RegistrationRegistryAdapter struct {
	store store.Store
}

// NewRegistrationRegistryAdapter creates a new adapter
func NewRegistrationRegistryAdapter(s store.Store) *RegistrationRegistryAdapter {
	return &RegistrationRegistryAdapter{store: s}
}

// UpsertProvider creates or updates a service registration
func (a *RegistrationRegistryAdapter) UpsertProvider(ctx context.Context, provider registration.RegisteredProvider) error {
	serviceUUID, err := uuid.Parse(provider.ServiceID)
	if err != nil {
		return fmt.Errorf("invalid service ID: %w", err)
	}

	// Check if service exists
	existing, err := a.store.Provider().Get(ctx, serviceUUID)

	dbProvider := model.Provider{
		ID:           serviceUUID,
		Name:         fmt.Sprintf("%s-%s", provider.ServiceID, provider.ResourceKind),
		ProviderType: provider.ResourceKind,
		Description:  fmt.Sprintf("Service for %s resources in %s/%s", provider.ResourceKind, provider.Metadata.Region, provider.Metadata.Zone),
		Endpoint:     provider.Endpoint,
		ApiHost:      provider.Endpoint,
		Operations:   pq.StringArray(provider.Operations),
	}

	if err != nil || existing == nil {
		// Create new service
		_, err = a.store.Provider().Create(ctx, dbProvider)
		return err
	}

	// Update existing service
	_, err = a.store.Provider().Update(ctx, dbProvider)
	return err
}

// GetProvider retrieves a service by ID and resource kind
func (a *RegistrationRegistryAdapter) GetProvider(ctx context.Context, serviceID, resourceKind string) (*registration.RegisteredProvider, error) {
	serviceUUID, err := uuid.Parse(serviceID)
	if err != nil {
		return nil, fmt.Errorf("invalid service ID: %w", err)
	}

	dbProvider, err := a.store.Provider().Get(ctx, serviceUUID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("service not found")
		}
		return nil, err
	}

	// Check if resource kind matches
	if dbProvider.ProviderType != resourceKind {
		return nil, fmt.Errorf("service not found for resource kind %s", resourceKind)
	}

	return &registration.RegisteredProvider{
		ServiceID:    serviceID,
		ResourceKind: dbProvider.ProviderType,
		Endpoint:     dbProvider.Endpoint,
		Operations:   []string(dbProvider.Operations),
		Status:       "active",
		RegisteredAt: dbProvider.CreatedAt,
		UpdatedAt:    dbProvider.UpdatedAt,
	}, nil
}

// DeleteProvider removes a service registration
func (a *RegistrationRegistryAdapter) DeleteProvider(ctx context.Context, serviceID, resourceKind string) error {
	serviceUUID, err := uuid.Parse(serviceID)
	if err != nil {
		return fmt.Errorf("invalid service ID: %w", err)
	}

	// Verify service exists and matches resource kind
	_, err = a.GetProvider(ctx, serviceID, resourceKind)
	if err != nil {
		return err
	}

	return a.store.Provider().Delete(ctx, serviceUUID)
}

// ListProviders lists all registered services for a resource kind
func (a *RegistrationRegistryAdapter) ListProviders(ctx context.Context, resourceKind string) ([]registration.RegisteredProvider, error) {
	dbProviders, err := a.store.Provider().ListByType(ctx, resourceKind)
	if err != nil {
		return nil, err
	}

	providers := make([]registration.RegisteredProvider, 0, len(dbProviders))
	for _, dbProvider := range dbProviders {
		providers = append(providers, registration.RegisteredProvider{
			ServiceID:    dbProvider.ID.String(),
			ResourceKind: dbProvider.ProviderType,
			Endpoint:     dbProvider.Endpoint,
			Operations:   []string(dbProvider.Operations),
			Status:       "active",
			RegisteredAt: dbProvider.CreatedAt,
			UpdatedAt:    dbProvider.UpdatedAt,
		})
	}

	return providers, nil
}
