package service

import (
	"context"
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

func (v *ProviderService) CreateProvider(ctx context.Context, request *server.CreateProviderJSONRequestBody) error {
	logger := zap.S().Named("provider_service:createProvider")
	logger.Info("Creating service provider")

	newProvider := model.Provider{
		Name:         request.Name,
		Endpoint:     request.Endpoint,
		ID:           uuid.MustParse(request.Id),
		Description:  request.Description,
		ProviderType: string(request.Type),
		ApiHost:      request.ApiHost,
		Operations:   request.Operations,
	}
	result, err := v.restyClient.R().Get(newProvider.ApiHost + "/health")
	if err != nil || result.StatusCode() != http.StatusOK {
		logger.Error("Failed to get health status or health endpoint return OK status")
		return fmt.Errorf("health endpoint does not return OK status")
	}

	// TODO Get resource information about the provider

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
		Name:        existingProvider.Name,
		Endpoint:    existingProvider.Endpoint,
		Id:          existingProvider.ID.String(),
		Description: existingProvider.Description,
		Type:        server.ProviderType(existingProvider.ProviderType),
	}
	logger.Info("Successfully retrieved provider details")
	return provider, nil
}

func (v *ProviderService) ListProvider(ctx context.Context, providerType *string) (*[]server.Provider, error) {
	logger := zap.S().Named("service_provider:listProviders")
	logger.Info("Retrieving Service Providers")

	// Filter by type if provided
	var providers model.ProviderList
	var err error
	if providerType != nil {
		providers, err = v.store.Provider().ListByType(ctx, *providerType)
	} else {
		providers, err = v.store.Provider().List(ctx)
	}

	if err != nil {
		logger.Error("Failed to list providers", err)
		return &[]server.Provider{}, err
	}

	var providerList []server.Provider
	for _, v := range providers {
		providerList = append(providerList, server.Provider{
			Description: v.Description,
			Id:          v.ID.String(),
			Name:        v.Name,
			Type:        server.ProviderType(v.ProviderType),
			Endpoint:    v.Endpoint,
			ApiHost:     v.ApiHost,
			Operations:  v.Operations,
		})
	}
	logger.Info("Successfully retrieved Service Providers")
	return &providerList, nil
}

func (v *ProviderService) UpdateProvider(ctx context.Context, updateProvider server.ApplyProviderJSONRequestBody) (server.Provider, error) {
	logger := zap.S().Named("service_provider:UpdateProvider")
	logger.Info("Retrieving Service Providers by ID")

	providerUUID := uuid.MustParse(updateProvider.Id)
	_, err := v.store.Provider().Get(ctx, providerUUID)
	if err != nil {
		logger.Error("ProviderID does not exist in database", err)
		return server.Provider{}, err
	}

	updatedModel := model.Provider{
		ID:           providerUUID,
		Name:         updateProvider.Name,
		Endpoint:     updateProvider.Endpoint,
		Description:  updateProvider.Description,
		ProviderType: string(updateProvider.Type),
		ApiHost:      updateProvider.ApiHost,
		Operations:   updateProvider.Operations,
	}
	_, err = v.store.Provider().Update(ctx, updatedModel)
	if err != nil {
		return server.Provider{}, err
	}
	logger.Info("Successfully updated service provider")
	return updateProvider, nil
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
