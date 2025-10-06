package v1alpha1

import (
	"context"

	"github.com/dcm-project/service-provider-api/internal/api/server"
	"github.com/dcm-project/service-provider-api/internal/service"
	"github.com/dcm-project/service-provider-api/internal/store"
	"go.uber.org/zap"
)

type ServiceHandler struct {
	ps    *service.ProviderService
	store store.Store
}

func NewServiceHandler(store store.Store, providerService *service.ProviderService) *ServiceHandler {
	return &ServiceHandler{
		store: store,
		ps:    providerService,
	}
}

// (GET /health)
func (s *ServiceHandler) ListHealth(ctx context.Context, request server.ListHealthRequestObject) (server.ListHealthResponseObject, error) {
	return server.ListHealth200Response{}, nil
}

// ListProviders (GET /providers)
func (s *ServiceHandler) ListProviders(ctx context.Context, request server.ListProvidersRequestObject) (server.ListProvidersResponseObject, error) {
	logger := zap.S().Named("service-provider")
	logger.Info("Listing service providers... ")
	// TODO
	return nil, nil
}

// GetProvider (GET /provider/{id})
func (s *ServiceHandler) GetProvider(ctx context.Context, request server.GetProviderRequestObject) (server.GetProviderResponseObject, error) {
	logger := zap.S().Named("service-provider")
	logger.Info("Retrieving provider: ", "ID: ", request.Id)
	// TODO
	return nil, nil
}

// CreateApplication (POST /provider/application/{id})
func (s *ServiceHandler) CreateApplication(ctx context.Context, request server.CreateApplicationRequestObject) (server.CreateApplicationResponseObject, error) {
	logger := zap.S().Named("service-provider")
	logger.Info("Creating Application. ", "Application: ", request)

	// TODO
	logger.Info("Application created. ", "Application: ", "app")
	return nil, nil
}

// (DELETE /provider/{id})
func (s *ServiceHandler) DeleteApplication(ctx context.Context, request server.DeleteApplicationRequestObject) (server.DeleteApplicationResponseObject, error) {
	logger := zap.S().Named("service-provider")
	logger.Info("Deleting Application. ", "Application: ", request)
	// TODO
	logger.Info("Application deleted. ", "Application: ", request.Id)
	return nil, nil
}
