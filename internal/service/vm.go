package service

import (
	"context"

	"github.com/dcm-project/service-provider-api/internal/api/server"
	"github.com/dcm-project/service-provider-api/internal/registry"
	"github.com/dcm-project/service-provider-api/internal/registry/vm"
	"go.uber.org/zap"
)

type VMService struct {
	registry *registry.ProviderRegistry
}

func NewVMService(providerRegistry *registry.ProviderRegistry) *VMService {
	return &VMService{registry: providerRegistry}
}

func (v *VMService) CreateVM(ctx context.Context, providerID string, userRequest server.CreateVMJSONRequestBody) (vm.DeclaredVM, error) {
	logger := zap.S().Named("vm_service:create_vm")
	logger.Info("Starting VM creation for: ", *userRequest.Name)

	request := vm.Request{
		OsImage:   *userRequest.OsImage,
		Ram:       *userRequest.Ram,
		Cpu:       *userRequest.Cpu,
		RequestId: *userRequest.Id,
		Namespace: *userRequest.Namespace,
		VMName:    *userRequest.Name,
	}

	p, err := v.registry.GetProvider(providerID)
	if err != nil {
		return vm.DeclaredVM{}, err
	}
	declaredVM, err := p.CreateVM(ctx, request)
	if err != nil {
		return vm.DeclaredVM{}, err
	}
	logger.Info("Successfully created VM", declaredVM)
	return declaredVM, nil
}

func (v *VMService) GetVMProviders(ctx context.Context) *[]server.VMProvider {
	logger := zap.S().Named("vm-service-provider:get_providers")
	logger.Info("Retrieving VM Service Providers")
	providers := v.registry.ListProvider()
	var vmProviderList []server.VMProvider
	for _, v := range providers {
		vmProviderList = append(vmProviderList, server.VMProvider{
			Description: v.Description(),
			Id:          v.ProviderID(),
			Name:        v.Name(),
			Type:        "vm",
		})
	}
	return &vmProviderList
}

func (v *VMService) GetVMProvider(providerID string) (server.GetVMProviderResponseObject, error) {
	logger := zap.S().Named("service-provider:get_provider")
	logger.Info("Retrieving providers by ID: ")
	providerInfo, err := v.registry.GetProvider(providerID)
	if err != nil {
		logger.Error("Failed to retrieve provider info: ", "Provider-ID", providerInfo)
		return nil, err
	}
	logger.Info("Successfully retrieved provider info")
	return server.GetVMProvider200JSONResponse{Id: providerID, Name: providerInfo.Name(), Type: "vm", Description: providerInfo.Description()}, nil
}

func (v *VMService) DeleteVMApplication(ctx context.Context, providerID, appID string) (vm.DeclaredVM, error) {
	logger := zap.S().Named("service-provider:delete_app")
	logger.Info("Deleting VM application", "ID ", appID)
	provider, err := v.registry.GetProvider(providerID)
	if err != nil {
		logger.Error("Failed to retrieve provider information.")
		return vm.DeclaredVM{}, err
	}
	declaredVM, err := provider.DeleteVM(ctx, appID)
	if err != nil {
		logger.Error("Failed to delete VM application ", "ID ", appID)
		return vm.DeclaredVM{}, err
	}

	return declaredVM, nil
}
