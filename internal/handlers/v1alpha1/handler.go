package v1alpha1

import (
	"context"

	"github.com/dcm-project/service-provider-api/internal/api/server"
	"github.com/dcm-project/service-provider-api/internal/service"
	"go.uber.org/zap"
)

type ServiceHandler struct {
	providerService *service.ProviderService
}

func NewServiceHandler(providerService *service.ProviderService) *ServiceHandler {
	return &ServiceHandler{
		providerService: providerService,
	}
}

// ListHealth (GET /health)
func (s *ServiceHandler) ListHealth(ctx context.Context, request server.ListHealthRequestObject) (server.ListHealthResponseObject, error) {
	return server.ListHealth200Response{}, nil
}

// ListProviders (GET /providers)
func (s *ServiceHandler) ListProviders(ctx context.Context, request server.ListProvidersRequestObject) (server.ListProvidersResponseObject, error) {
	logger := zap.S().Named("handler:listProviders")
	logger.Info("Retrieving service providers... ")

	//TODO implement filter by type

	providers, err := s.providerService.ListProvider(ctx)
	if err != nil {
		return nil, err
	}
	return server.ListProviders200JSONResponse{
		Providers: providers,
	}, nil
}

// CreateProvider (POST /provider/{providerId})
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

// GetProvider (GET /provider/{providerId})
func (s *ServiceHandler) GetProvider(ctx context.Context, request server.GetProviderRequestObject) (server.GetProviderResponseObject, error) {
	logger := zap.S().Named("handler:getProvider")
	logger.Info("Retrieving provider details: ", "ID: ", request.ProviderId)
	providerInfo, err := s.providerService.GetProvider(ctx, request.ProviderId.String())
	if err != nil {
		return server.GetProvider400JSONResponse{Error: err.Error()}, nil
	}
	return server.GetProvider200JSONResponse{
		Description: providerInfo.Description,
		Endpoint:    providerInfo.Endpoint,
		Id:          providerInfo.Id,
		Name:        providerInfo.Name,
		Type:        providerInfo.Type,
	}, nil
}

// ApplyProvider (PUT /provider/{providerId}
func (s *ServiceHandler) ApplyProvider(ctx context.Context, request server.ApplyProviderRequestObject) (server.ApplyProviderResponseObject, error) {
	logger := zap.S().Named("handler:applyProvider")
	logger.Info("Updating provider details: ", "ID: ", request.ProviderId)

	// TODO

	return nil, nil
}

// DeleteProvider (DELETE /provider/{providerId})
func (s *ServiceHandler) DeleteProvider(ctx context.Context, request server.DeleteProviderRequestObject) (server.DeleteProviderResponseObject, error) {
	logger := zap.S().Named("handler:deleteProvider")
	logger.Info("Deleting provider: ", "ID: ", request.ProviderId)

	// TODO

	return nil, nil
}
