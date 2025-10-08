package v1alpha1

import (
	"context"

	"github.com/dcm-project/service-provider-api/internal/api/server"
	"github.com/dcm-project/service-provider-api/internal/service"
	"go.uber.org/zap"
)

type ServiceHandler struct {
	vmService *service.VMService
}

func NewServiceHandler(providerService *service.VMService) *ServiceHandler {
	return &ServiceHandler{
		vmService: providerService,
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

// CreateVM (POST /vm/provider/{provider-id}
func (s *ServiceHandler) CreateVM(ctx context.Context, request server.CreateVMRequestObject) (server.CreateVMResponseObject, error) {
	logger := zap.S().Named("service-provider")
	logger.Info("Creating VM. ", "VM: ", request)
	vm, err := s.vmService.CreateVM(ctx, request.ProviderId.String(), *request.Body)
	if err != nil {
		return nil, err
	}

	logger.Info("Successfully created VM. ", "VM: ", *request.Body.Name)
	return server.CreateVM201JSONResponse{Id: &vm.ID, Name: &vm.RequestInfo.VMName, Namespace: &vm.RequestInfo.Namespace}, nil
}

// (DELETE /provider/{id})
func (s *ServiceHandler) DeleteApplication(ctx context.Context, request server.DeleteApplicationRequestObject) (server.DeleteApplicationResponseObject, error) {
	logger := zap.S().Named("service-provider")
	logger.Info("Deleting Application. ", "Application: ", request)
	// TODO
	logger.Info("Application deleted. ", "Application: ", request.ProviderId)
	return nil, nil
}
