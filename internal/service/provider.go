package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dcm-project/service-provider-api/internal/api/server"
	"github.com/dcm-project/service-provider-api/internal/store"
	"github.com/dcm-project/service-provider-api/internal/store/model"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ProviderService struct {
	store       store.Store
	restyClient resty.Client
}

func NewProviderService(store store.Store, client *resty.Client) *ProviderService {
	return &ProviderService{store: store, restyClient: *client}
}

func (v *ProviderService) CreateProvider(ctx context.Context, request *server.RegisterProviderJSONRequestBody) error {
	logger := zap.S().Named("provider_service:createProvider")
	logger.Info("Creating service provider")

	newProvider := model.Provider{
		ID:           request.Uuid,
		Name:         request.Name,
		Version:      request.Version,
		Endpoint:     request.Endpoint,
		ProviderType: string(request.Type),
	}
	result, err := v.restyClient.R().Get(newProvider.Endpoint + "/health")
	if err != nil {
		logger.Errorw("Failed to connect to provider health endpoint", "error", err, "endpoint", newProvider.Endpoint)
		return fmt.Errorf("failed to connect to provider health endpoint %s: %w", newProvider.Endpoint, err)
	}
	if result.StatusCode() != http.StatusOK {
		logger.Errorw("Provider health endpoint returned non-OK status", "status_code", result.StatusCode(), "endpoint", newProvider.Endpoint)
		return fmt.Errorf("provider health endpoint returned status %d, expected %d", result.StatusCode(), http.StatusOK)
	}

	provider, err := v.store.Provider().Create(ctx, newProvider)
	if err != nil {
		return err
	}
	logger.Info("Successfully created provider: ", provider.ID)
	return nil
}

func (v *ProviderService) GetProvider(ctx context.Context, providerID string) (server.Provider, error) {
	logger := zap.S().Named("provider_service:getProvider")
	logger.Info("Retrieving provider details")

	existingProvider, err := v.store.Provider().Get(ctx, uuid.MustParse(providerID))
	if err != nil {
		return server.Provider{}, fmt.Errorf("provider %s not found", providerID)
	}
	provider := server.Provider{
		Name:     existingProvider.Name,
		Uuid:     existingProvider.ID,
		Type:     existingProvider.ProviderType,
		Version:  existingProvider.Version,
		Endpoint: existingProvider.Endpoint,
	}
	logger.Info("Successfully retrieved provider details")
	return provider, nil
}

func (v *ProviderService) ListProvider(ctx context.Context) (*[]server.Provider, error) {
	logger := zap.S().Named("service_provider:listProviders")
	logger.Info("Retrieving Service Providers")

	providers, err := v.store.Provider().List(ctx)
	if err != nil {
		logger.Error("Failed to list providers", err)
		return &[]server.Provider{}, err
	}

	var providerList []server.Provider
	for _, v := range providers {
		providerList = append(providerList, server.Provider{
			Uuid:     v.ID,
			Name:     v.Name,
			Type:     v.ProviderType,
			Version:  v.Version,
			Endpoint: v.Endpoint,
		})
	}
	logger.Info("Successfully retrieved Service Providers")
	return &providerList, nil
}

func (v *ProviderService) DeleteProvider(ctx context.Context, providerID string) error {
	logger := zap.S().Named("service_provider:DeleteProvider")
	logger.Info("Deleting provider by ID")

	providerUUID := uuid.MustParse(providerID)
	err := v.store.Provider().Delete(ctx, providerUUID)
	if err != nil {
		return err
	}
	logger.Info("Successfully deleted service provider")
	return nil
}

func (v *ProviderService) ListServiceTypes(ctx context.Context) (*[]server.ServiceType, error) {
	logger := zap.S().Named("service_provider:listServiceTypes")
	logger.Info("Retrieving service types")

	providers, err := v.store.Provider().List(ctx)
	if err != nil {
		return nil, err
	}
	var serviceTypeList []server.ServiceType
	for _, provider := range providers {
		result, err := v.restyClient.R().Get(provider.Endpoint + "/api/v1/services")
		if err != nil {
			return nil, err
		}
		if result.StatusCode() != http.StatusOK {
			return nil, fmt.Errorf("failed to list service types: %s", result.Status())
		}
		serviceTypes := result.Body()

		// Parse the JSON response body into a slice of map[string]interface{}
		var providerServiceTypeList []server.ServiceType
		if err := json.Unmarshal(serviceTypes, &providerServiceTypeList); err != nil {
			return nil, fmt.Errorf("failed to parse service types json: %v", err)
		}
		serviceTypeList = append(serviceTypeList, providerServiceTypeList...)
	}
	return &serviceTypeList, nil
}

func (v *ProviderService) GetServiceTypeSchema(ctx context.Context, providerID string, serviceTypeID string) (*map[string]interface{}, error) {
	logger := zap.S().Named("service_provider:getServiceTypeSchema")
	logger.Info("Retrieving service type schema")

	providerUUID := uuid.MustParse(providerID)
	serviceTypeUUID := uuid.MustParse(serviceTypeID)
	provider, err := v.store.Provider().Get(ctx, providerUUID)
	if err != nil {
		return nil, err
	}
	result, err := v.restyClient.R().Get(provider.Endpoint + "/api/v1/services/" + serviceTypeUUID.String() + "/schema")
	if err != nil {
		return nil, err
	}
	if result.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get service type schema: %s", result.Status())
	}
	bodyBytes := result.Body()

	var schemaMap map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &schemaMap); err != nil {
		return nil, fmt.Errorf("failed to parse service type schema json: %v", err)
	}
	return &schemaMap, nil
}
