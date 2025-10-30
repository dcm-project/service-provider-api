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

// Health check (GET /health)
func (s *ServiceHandler) ListHealth(ctx context.Context, request server.ListHealthRequestObject) (server.ListHealthResponseObject, error) {
	logger := zap.S().Named("handler:health")
	logger.Info("Checking health... ")

	return server.ListHealth200Response{}, nil
}

// RegisterProvider (POST /provider)
func (s *ServiceHandler) RegisterProvider(ctx context.Context, request server.RegisterProviderRequestObject) (server.RegisterProviderResponseObject, error) {
	logger := zap.S().Named("handler:registerProvider")
	logger.Info("Registering new service provider")

	newProvider := request.Body
	err := s.providerService.CreateProvider(ctx, newProvider)
	if err != nil {
		return server.RegisterProvider400JSONResponse{}, nil
	}
	return server.RegisterProvider201JSONResponse{
		Endpoint: newProvider.Endpoint,
		Name:     newProvider.Name,
	}, nil
}

// ListProviders (GET /providers)
func (s *ServiceHandler) ListProviders(ctx context.Context, request server.ListProvidersRequestObject) (server.ListProvidersResponseObject, error) {
	logger := zap.S().Named("handler:listProviders")
	logger.Info("Retrieving service providers... ")

	providers, err := s.providerService.ListProvider(ctx)
	if err != nil {
		return nil, err
	}
	return server.ListProviders200JSONResponse{
		Providers: providers,
	}, nil
}

// CreateProvider (POST /providers)
func (s *ServiceHandler) CreateProvider(ctx context.Context, request server.RegisterProviderRequestObject) (server.RegisterProviderResponseObject, error) {
	logger := zap.S().Named("handler:createProvider")
	logger.Info("Creating new service provider")

	newProvider := request.Body
	err := s.providerService.CreateProvider(ctx, newProvider)
	if err != nil {
		return server.RegisterProvider400JSONResponse{}, nil
	}

	return server.RegisterProvider201JSONResponse{
		Uuid:     newProvider.Uuid,
		Endpoint: newProvider.Endpoint,
		Name:     newProvider.Name,
		Type:     newProvider.Type,
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
		Endpoint: providerInfo.Endpoint,
		Uuid:     providerInfo.Uuid,
		Name:     providerInfo.Name,
		Type:     providerInfo.Type,
	}, nil
}

// DeleteProvider (DELETE /provider/{providerId})
func (s *ServiceHandler) DeleteProvider(ctx context.Context, request server.DeleteProviderRequestObject) (server.DeleteProviderResponseObject, error) {
	logger := zap.S().Named("handler:deleteProvider")
	providerID := request.ProviderId.String()
	logger.Info("Deleting provider: ", "ID: ", providerID)

	err := s.providerService.DeleteProvider(ctx, providerID)
	if err != nil {
		return server.DeleteProvider404JSONResponse{}, err
	}
	logger.Info("Successfully deleted service provider")
	return server.DeleteProvider400JSONResponse{}, nil
}

// apply provider (PUT /providers/{providerId})
func (s *ServiceHandler) ApplyProvider(ctx context.Context, request server.ApplyProviderRequestObject) (server.ApplyProviderResponseObject, error) {
	logger := zap.S().Named("handler:applyProvider")
	logger.Info("Applying provider details: ", "ID: ", request.ProviderId)

	return server.ApplyProvider201JSONResponse{}, nil
}

// ListServiceTypes (GET /service-types)
func (s *ServiceHandler) ListServiceTypes(ctx context.Context, request server.ListServiceTypesRequestObject) (server.ListServiceTypesResponseObject, error) {
	logger := zap.S().Named("handler:listServiceTypes")
	logger.Info("Retrieving service types... ")

	serviceTypes, err := s.providerService.ListServiceTypes(ctx)
	if err != nil {
		return server.ListServiceTypes400JSONResponse{}, nil
	}

	return server.ListServiceTypes200JSONResponse(*serviceTypes), nil
}

// GetServiceType (GET /service-types/{serviceType})
func (s *ServiceHandler) GetServiceType(ctx context.Context, request server.GetServiceTypeRequestObject) (server.GetServiceTypeResponseObject, error) {
	logger := zap.S().Named("handler:getServiceType")
	logger.Info("Retrieving service type details: ", "ID: ", request.ServiceTypeId)

	// FIXME: store services as part of registration
	serviceTypes, err := s.providerService.ListServiceTypes(ctx)
	if err != nil {
		return server.GetServiceType400JSONResponse{Error: err.Error()}, nil
	}

	for _, serviceType := range *serviceTypes {
		if serviceType.Id.String() == request.ServiceTypeId {
			return server.GetServiceType200JSONResponse{
				Id:   serviceType.Id,
				Name: serviceType.Name,
			}, nil
		}
	}

	return server.GetServiceType404JSONResponse{Error: "Service type not found"}, nil
}

// GetServiceTypeSchema (GET /service-types/{serviceType}/schema)
func (s *ServiceHandler) GetServiceTypeSchema(ctx context.Context, request server.GetServiceTypeSchemaRequestObject) (server.GetServiceTypeSchemaResponseObject, error) {
	logger := zap.S().Named("handler:getServiceTypeSchema")
	logger.Info("Retrieving service type schema: ", "ID: ", request.ServiceTypeId)

	// FIXME: store services as part of registration
	serviceTypes, err := s.providerService.ListServiceTypes(ctx)
	if err != nil {
		return server.GetServiceTypeSchema400JSONResponse{Error: err.Error()}, nil
	}

	for _, serviceType := range *serviceTypes {
		if serviceType.Id.String() == request.ServiceTypeId {
			return server.GetServiceTypeSchema200ApplicationSchemaPlusJSONResponse{
				// TODO: Add schema from provider
				"Schema": serviceType.Description,
			}, nil
		}
	}

	return server.GetServiceTypeSchema404JSONResponse{}, nil
}

// ListServiceInstances (GET /services)
func (s *ServiceHandler) ListServiceInstances(ctx context.Context, request server.ListServiceInstancesRequestObject) (server.ListServiceInstancesResponseObject, error) {
	logger := zap.S().Named("handler:listServiceInstances")
	logger.Info("Retrieving service instances... ")

	return server.ListServiceInstances200JSONResponse{}, nil
}

// CreateServiceInstance (POST /services)
func (s *ServiceHandler) CreateServiceInstance(ctx context.Context, request server.CreateServiceInstanceRequestObject) (server.CreateServiceInstanceResponseObject, error) {
	logger := zap.S().Named("handler:createServiceInstance")
	logger.Info("Creating service instance... ")

	return server.CreateServiceInstance201JSONResponse{}, nil
}

// DeleteServiceInstance (DELETE /services/{serviceId})
func (s *ServiceHandler) DeleteServiceInstance(ctx context.Context, request server.DeleteServiceInstanceRequestObject) (server.DeleteServiceInstanceResponseObject, error) {
	logger := zap.S().Named("handler:deleteServiceInstance")
	logger.Info("Deleting service instance... ")

	return server.DeleteServiceInstance202Response{}, nil
}

// GetServiceInstance (GET /service/{serviceId})
func (s *ServiceHandler) GetServiceInstance(ctx context.Context, request server.GetServiceInstanceRequestObject) (server.GetServiceInstanceResponseObject, error) {
	logger := zap.S().Named("handler:getServiceInstance")
	logger.Info("Retrieving service instance details: ", "ID: ", request.ServiceId)

	return server.GetServiceInstance200JSONResponse{}, nil
}
