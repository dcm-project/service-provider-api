package v1alpha1

import (
	"context"
	"time"

	"github.com/dcm-project/service-provider-api/internal/api/server"
	"github.com/dcm-project/service-provider-api/internal/service"
	"github.com/dcm-project/service-provider-api/internal/store"
	storeregistration "github.com/dcm-project/service-provider-api/internal/store/registration"
	"github.com/dcm-project/service-provider-api/pkg/registration"
	"go.uber.org/zap"
)

type ServiceHandler struct {
	providerService     *service.ProviderService
	registrationHandler *registration.Handler
	store               store.Store
}

func NewServiceHandler(providerService *service.ProviderService) *ServiceHandler {
	return &ServiceHandler{
		providerService: providerService,
	}
}

func (s *ServiceHandler) SetRegistrationHandler(handler *registration.Handler) {
	s.registrationHandler = handler
}

func (s *ServiceHandler) SetStore(store store.Store) {
	s.store = store
}

// ListHealth (GET /health)
func (s *ServiceHandler) ListHealth(ctx context.Context, request server.ListHealthRequestObject) (server.ListHealthResponseObject, error) {
	return server.ListHealth200Response{}, nil
}

// ListProviders (GET /providers)
func (s *ServiceHandler) ListProviders(ctx context.Context, request server.ListProvidersRequestObject) (server.ListProvidersResponseObject, error) {
	logger := zap.S().Named("handler:listProviders")
	logger.Info("Retrieving service providers... ")

	var providerType = ""
	if request.Params.Type != nil && *request.Params.Type != "" {
		providerType = *request.Params.Type
	}

	providers, err := s.providerService.ListProvider(ctx, providerType)
	if err != nil {
		return nil, err
	}
	return server.ListProviders200JSONResponse{
		Providers: providers,
	}, nil
}

// CreateProvider (POST /providers)
func (s *ServiceHandler) CreateProvider(ctx context.Context, request server.CreateProviderRequestObject) (server.CreateProviderResponseObject, error) {
	logger := zap.S().Named("handler:createProvider")
	logger.Info("Creating new service provider")

	newProvider := request.Body
	err := s.providerService.CreateProvider(ctx, newProvider)
	if err != nil {
		return server.CreateProvider400JSONResponse{Error: err.Error()}, nil
	}

	return server.CreateProvider201JSONResponse{
		Description: newProvider.Description,
		Endpoint:    newProvider.Endpoint,
		Id:          newProvider.Id,
		Name:        newProvider.Name,
		Type:        newProvider.Type,
		Operations:  newProvider.Operations,
		ApiHost:     newProvider.ApiHost,
	}, nil
}

// GetProvider (GET /providers/{providerId})
func (s *ServiceHandler) GetProvider(ctx context.Context, request server.GetProviderRequestObject) (server.GetProviderResponseObject, error) {
	logger := zap.S().Named("handler:getProvider")
	logger.Info("Retrieving provider details: ", "ID: ", request.ProviderId)
	providerInfo, err := s.providerService.GetProvider(ctx, request.ProviderId.String())
	if err != nil {
		return server.GetProvider404JSONResponse{Error: err.Error()}, nil
	}
	return server.GetProvider200JSONResponse{
		Description: providerInfo.Description,
		Endpoint:    providerInfo.Endpoint,
		Id:          providerInfo.Id,
		Name:        providerInfo.Name,
		Type:        providerInfo.Type,
		ApiHost:     providerInfo.ApiHost,
		Operations:  providerInfo.Operations,
	}, nil
}

// ApplyProvider (PUT /providers/{providerId}
func (s *ServiceHandler) ApplyProvider(ctx context.Context, request server.ApplyProviderRequestObject) (server.ApplyProviderResponseObject, error) {
	logger := zap.S().Named("handler:applyProvider")
	logger.Info("Updating provider details: ", "ID: ", request.ProviderId)

	_, err := s.providerService.UpdateProvider(ctx, *request.Body)

	if err != nil {
		return server.ApplyProvider404JSONResponse{Error: err.Error()}, nil
	}
	return nil, nil
}

// DeleteProvider (DELETE /providers/{providerId})
func (s *ServiceHandler) DeleteProvider(ctx context.Context, request server.DeleteProviderRequestObject) (server.DeleteProviderResponseObject, error) {
	logger := zap.S().Named("handler:deleteProvider")
	providerID := request.ProviderId.String()
	logger.Info("Deleting provider: ", "ID: ", providerID)

	err := s.providerService.DeleteProvider(ctx, providerID)
	if err != nil {
		return server.DeleteProvider404JSONResponse{}, err
	}
	logger.Info("Successfully deleted service provider")
	return server.DeleteProvider204JSONResponse{Id: providerID}, nil
}

// RegisterProvider (POST /resource/{resourceKind}/provider)
func (s *ServiceHandler) RegisterProvider(ctx context.Context, request server.RegisterProviderRequestObject) (server.RegisterProviderResponseObject, error) {
	logger := zap.S().Named("handler:registerProvider")

	if s.registrationHandler == nil {
		return server.RegisterProvider500JSONResponse{Error: "registration handler not initialized"}, nil
	}

	// Call registration handler directly with OpenAPI types
	resp, err := s.registrationHandler.Register(
		ctx,
		request.Body.ServiceId,
		request.ResourceKind,
		request.Body.Endpoint,
		request.Body.Metadata,
		request.Body.Operations,
	)
	if err != nil {
		regErr, ok := err.(*registration.RegistrationError)
		if ok && regErr.Code == registration.ErrCodeValidation {
			return server.RegisterProvider400JSONResponse{Error: regErr.Error()}, nil
		}
		logger.Errorw("Registration failed", "error", err)
		return server.RegisterProvider500JSONResponse{Error: err.Error()}, nil
	}

	return server.RegisterProvider200JSONResponse(*resp), nil
}

// UnregisterProvider (DELETE /resource/{resourceKind}/provider/{providerId})
func (s *ServiceHandler) UnregisterProvider(ctx context.Context, request server.UnregisterProviderRequestObject) (server.UnregisterProviderResponseObject, error) {
	logger := zap.S().Named("handler:unregisterProvider")

	if s.registrationHandler == nil {
		return server.UnregisterProvider500JSONResponse{Error: "registration handler not initialized"}, nil
	}

	err := s.registrationHandler.Unregister(ctx, request.ProviderId, request.ResourceKind)
	if err != nil {
		regErr, ok := err.(*registration.RegistrationError)
		if ok && regErr.Code == registration.ErrCodeNotFound {
			return server.UnregisterProvider404JSONResponse{Error: regErr.Error()}, nil
		}
		logger.Errorw("Unregistration failed", "error", err)
		return server.UnregisterProvider500JSONResponse{Error: err.Error()}, nil
	}

	return server.UnregisterProvider204Response{}, nil
}

// GetRegisteredProvider (GET /resource/{resourceKind}/provider/{providerId})
func (s *ServiceHandler) GetRegisteredProvider(ctx context.Context, request server.GetRegisteredProviderRequestObject) (server.GetRegisteredProviderResponseObject, error) {
	logger := zap.S().Named("handler:getRegisteredProvider")

	if s.registrationHandler == nil {
		return server.GetRegisteredProvider500JSONResponse{Error: "registration handler not initialized"}, nil
	}

	provider, err := s.registrationHandler.GetRegistration(ctx, request.ProviderId, request.ResourceKind)
	if err != nil {
		regErr, ok := err.(*registration.RegistrationError)
		if ok && regErr.Code == registration.ErrCodeNotFound {
			return server.GetRegisteredProvider404JSONResponse{Error: regErr.Error()}, nil
		}
		logger.Errorw("Failed to get provider", "error", err)
		return server.GetRegisteredProvider500JSONResponse{Error: err.Error()}, nil
	}

	// Convert to OpenAPI response
	return server.GetRegisteredProvider200JSONResponse{
		ServiceId:    &provider.ServiceID,
		ResourceKind: &provider.ResourceKind,
		Endpoint:     &provider.Endpoint,
		Metadata:     &provider.Metadata,
		Operations:   &provider.Operations,
		CatalogItem:  &provider.CatalogItem,
		Status:       &provider.Status,
		RegisteredAt: &provider.RegisteredAt,
		UpdatedAt:    &provider.UpdatedAt,
	}, nil
}

// ListRegisteredProviders (GET /resource/{resourceKind}/provider)
func (s *ServiceHandler) ListRegisteredProviders(ctx context.Context, request server.ListRegisteredProvidersRequestObject) (server.ListRegisteredProvidersResponseObject, error) {
	logger := zap.S().Named("handler:listRegisteredProviders")

	if s.registrationHandler == nil {
		return server.ListRegisteredProviders500JSONResponse{Error: "registration handler not initialized"}, nil
	}

	providers, err := s.registrationHandler.ListRegistrations(ctx, request.ResourceKind)
	if err != nil {
		logger.Errorw("Failed to list providers", "error", err)
		return server.ListRegisteredProviders500JSONResponse{Error: err.Error()}, nil
	}

	// Convert to OpenAPI response types
	result := make([]server.RegisteredProvider, len(providers))
	for i, p := range providers {
		result[i] = server.RegisteredProvider{
			ServiceId:    &p.ServiceID,
			ResourceKind: &p.ResourceKind,
			Endpoint:     &p.Endpoint,
			Metadata:     &p.Metadata,
			Operations:   &p.Operations,
			CatalogItem:  &p.CatalogItem,
			Status:       &p.Status,
			RegisteredAt: &p.RegisteredAt,
			UpdatedAt:    &p.UpdatedAt,
		}
	}

	return server.ListRegisteredProviders200JSONResponse(result), nil
}

// GetRegistry (GET /admin/registry)
func (s *ServiceHandler) GetRegistry(ctx context.Context, request server.GetRegistryRequestObject) (server.GetRegistryResponseObject, error) {
	logger := zap.S().Named("handler:getRegistry")

	if s.store == nil {
		return server.GetRegistry500JSONResponse{Error: "store not initialized"}, nil
	}

	registryAdapter := storeregistration.NewRegistrationRegistryAdapter(s.store)

	// Get unique resource kinds
	resourceKinds, err := s.store.Catalog().GetDistinctResourceKinds(ctx)
	if err != nil {
		logger.Errorw("Failed to get resource kinds", "error", err)
		return server.GetRegistry500JSONResponse{Error: "failed to retrieve registry"}, nil
	}

	// Collect all providers
	providersMap := make(map[string]*registryProvider)

	for _, resourceKind := range resourceKinds {
		providers, err := registryAdapter.ListProviders(ctx, resourceKind)
		if err != nil {
			logger.Warnw("Failed to list providers", "resource_kind", resourceKind, "error", err)
			continue
		}

		for _, provider := range providers {
			entry, exists := providersMap[provider.ServiceID]
			if !exists {
				entry = &registryProvider{
					ServiceID:     provider.ServiceID,
					Metadata:      &provider.Metadata,
					Status:        &provider.Status,
					Registrations: []registryProviderRegistration{},
					RegisteredAt:  &provider.RegisteredAt,
				}
				providersMap[provider.ServiceID] = entry
			}

			entry.Registrations = append(entry.Registrations, registryProviderRegistration{
				ResourceKind: &provider.ResourceKind,
				Endpoint:     &provider.Endpoint,
				Operations:   &provider.Operations,
				CatalogItem:  &provider.CatalogItem,
				RegisteredAt: &provider.RegisteredAt,
			})
		}
	}

	// Build response using the inline struct types from RegistryView
	registryItems := make([]struct {
		Metadata      *server.ProviderMetadata `json:"metadata,omitempty"`
		RegisteredAt  *time.Time               `json:"registered_at,omitempty"`
		Registrations *[]struct {
			CatalogItem  *string    `json:"catalog_item,omitempty"`
			Endpoint     *string    `json:"endpoint,omitempty"`
			Operations   *[]string  `json:"operations,omitempty"`
			RegisteredAt *time.Time `json:"registered_at,omitempty"`
			ResourceKind *string    `json:"resource_kind,omitempty"`
		} `json:"registrations,omitempty"`
		ServiceId *string `json:"service_id,omitempty"`
		Status    *string `json:"status,omitempty"`
	}, 0, len(providersMap))

	for _, entry := range providersMap {
		regs := make([]struct {
			CatalogItem  *string    `json:"catalog_item,omitempty"`
			Endpoint     *string    `json:"endpoint,omitempty"`
			Operations   *[]string  `json:"operations,omitempty"`
			RegisteredAt *time.Time `json:"registered_at,omitempty"`
			ResourceKind *string    `json:"resource_kind,omitempty"`
		}, len(entry.Registrations))

		for i, reg := range entry.Registrations {
			regs[i].ResourceKind = reg.ResourceKind
			regs[i].Endpoint = reg.Endpoint
			regs[i].Operations = reg.Operations
			regs[i].CatalogItem = reg.CatalogItem
			regs[i].RegisteredAt = reg.RegisteredAt
		}

		registryItems = append(registryItems, struct {
			Metadata      *server.ProviderMetadata `json:"metadata,omitempty"`
			RegisteredAt  *time.Time               `json:"registered_at,omitempty"`
			Registrations *[]struct {
				CatalogItem  *string    `json:"catalog_item,omitempty"`
				Endpoint     *string    `json:"endpoint,omitempty"`
				Operations   *[]string  `json:"operations,omitempty"`
				RegisteredAt *time.Time `json:"registered_at,omitempty"`
				ResourceKind *string    `json:"resource_kind,omitempty"`
			} `json:"registrations,omitempty"`
			ServiceId *string `json:"service_id,omitempty"`
			Status    *string `json:"status,omitempty"`
		}{
			ServiceId:     &entry.ServiceID,
			Metadata:      entry.Metadata,
			Status:        entry.Status,
			Registrations: &regs,
			RegisteredAt:  entry.RegisteredAt,
		})
	}

	total := len(registryItems)
	return server.GetRegistry200JSONResponse{
		Total:     &total,
		Providers: &registryItems,
	}, nil
}

// GetCatalog (GET /admin/catalog)
func (s *ServiceHandler) GetCatalog(ctx context.Context, request server.GetCatalogRequestObject) (server.GetCatalogResponseObject, error) {
	logger := zap.S().Named("handler:getCatalog")

	if s.store == nil {
		return server.GetCatalog500JSONResponse{Error: "store not initialized"}, nil
	}

	catalogItems, err := s.store.Catalog().ListAllCatalogItems(ctx)
	if err != nil {
		logger.Errorw("Failed to get catalog items", "error", err)
		return server.GetCatalog500JSONResponse{Error: "failed to retrieve catalog"}, nil
	}

	registryAdapter := storeregistration.NewRegistrationRegistryAdapter(s.store)
	catalogResponse := make([]struct {
		AvailableProviders *[]string `json:"available_providers,omitempty"`
		DisplayName        *string   `json:"display_name,omitempty"`
		Name               *string   `json:"name,omitempty"`
		ResourceKind       *string   `json:"resource_kind,omitempty"`
	}, 0, len(catalogItems))

	for _, item := range catalogItems {
		providers, err := registryAdapter.ListProviders(ctx, item.ResourceKind)
		if err != nil {
			logger.Warnw("Failed to list providers", "catalog_item", item.Name, "error", err)
			continue
		}

		serviceIDs := make([]string, 0, len(providers))
		for _, provider := range providers {
			serviceIDs = append(serviceIDs, provider.ServiceID)
		}

		catalogResponse = append(catalogResponse, struct {
			AvailableProviders *[]string `json:"available_providers,omitempty"`
			DisplayName        *string   `json:"display_name,omitempty"`
			Name               *string   `json:"name,omitempty"`
			ResourceKind       *string   `json:"resource_kind,omitempty"`
		}{
			Name:               &item.Name,
			DisplayName:        &item.DisplayName,
			ResourceKind:       &item.ResourceKind,
			AvailableProviders: &serviceIDs,
		})
	}

	total := len(catalogResponse)
	return server.GetCatalog200JSONResponse{
		Total:        &total,
		CatalogItems: &catalogResponse,
	}, nil
}

// Helper types for registry view
type registryProvider struct {
	ServiceID     string
	Metadata      *server.ProviderMetadata
	Status        *string
	Registrations []registryProviderRegistration
	RegisteredAt  *time.Time
}

type registryProviderRegistration struct {
	ResourceKind *string
	Endpoint     *string
	Operations   *[]string
	CatalogItem  *string
	RegisteredAt *time.Time
}
