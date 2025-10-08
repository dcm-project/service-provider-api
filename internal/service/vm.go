package service

import (
	"context"

	"github.com/dcm-project/service-provider-api/internal/api/server"
	"github.com/dcm-project/service-provider-api/internal/registry"
	"github.com/dcm-project/service-provider-api/internal/registry/vm"
	"github.com/google/uuid"
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

func (v *VMService) GetProviders(ctx context.Context) error {
	logger := zap.S().Named("service-provider:get_providers")
	logger.Info("Starting...")
	// TODO
	return nil
}

func (v *VMService) GetProvider(ctx context.Context) error {
	logger := zap.S().Named("service-provider:get_provider")
	logger.Info("Get...")
	// TODO
	return nil
}

func (v *VMService) DeleteApplication(ctx context.Context, id uuid.UUID) error {
	logger := zap.S().Named("service-provider:delete_app")
	logger.Info("Starting...")
	// TODO
	return nil
}
