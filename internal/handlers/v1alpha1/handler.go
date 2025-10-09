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

// ListVMProviders ListProviders (GET /vm/providers)
func (s *ServiceHandler) ListVMProviders(ctx context.Context, request server.ListVMProvidersRequestObject) (server.ListVMProvidersResponseObject, error) {
	logger := zap.S().Named("service-provider")
	logger.Info("Listing vm service providers... ")
	providers := s.vmService.GetVMProviders(ctx)
	return server.ListVMProviders200JSONResponse{
		Providers: providers,
	}, nil
}

// GetVMProvider GetProvider (GET /vm/provider/{provider-id})
func (s *ServiceHandler) GetVMProvider(ctx context.Context, request server.GetVMProviderRequestObject) (server.GetVMProviderResponseObject, error) {
	logger := zap.S().Named("service-provider")
	logger.Info("Retrieving provider: ", "ID: ", request.ProviderId)
	providerInfo, err := s.vmService.GetVMProvider(request.ProviderId.String())
	if err != nil {
		return server.GetVMProvider400JSONResponse{}, nil
	}
	return providerInfo, nil
}

// CreateVM (POST /vm/provider/{provider-id}/application
func (s *ServiceHandler) CreateVM(ctx context.Context, request server.CreateVMRequestObject) (server.CreateVMResponseObject, error) {
	logger := zap.S().Named("service-provider")
	logger.Info("Creating VM. ", "VM: ", request)
	vm, err := s.vmService.CreateVM(ctx, request.ProviderId.String(), *request.Body)
	if err != nil {
		return nil, err
	}

	logger.Info("Successfully created VM application. ", "VM: ", *request.Body.Name)
	return server.CreateVM201JSONResponse{Id: &vm.ID, Name: &vm.RequestInfo.VMName, Namespace: &vm.RequestInfo.Namespace}, nil
}

// DeleteVM (DELETE /vm/provider/{provider-id})/application
func (s *ServiceHandler) DeleteVM(ctx context.Context, request server.DeleteVMRequestObject) (server.DeleteVMResponseObject, error) {
	logger := zap.S().Named("service-provider")
	logger.Info("Deleting Application. ", "VM: ", request)
	appID := request.Params.AppID.String()
	declaredVM, err := s.vmService.DeleteVMApplication(ctx, request.ProviderId.String(), appID)
	if err != nil {
		logger.Error("Failed to Delete VM application")
		return nil, err
	}
	logger.Info("Successfully deleted VM application. ", "VM: ", request.ProviderId)
	return server.DeleteVM204JSONResponse{
		Id:        &appID,
		Name:      &declaredVM.RequestInfo.VMName,
		Namespace: &declaredVM.RequestInfo.Namespace}, nil
}
