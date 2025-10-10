package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dcm-project/service-provider-api/internal/api/server"
	"github.com/dcm-project/service-provider-api/internal/registry/vm"
	"github.com/dcm-project/service-provider-api/internal/store"
	"github.com/dcm-project/service-provider-api/internal/store/model"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type VMService struct {
	store       store.Store
	restyClient resty.Client
}

func NewVMService(store store.Store, client *resty.Client) *VMService {
	return &VMService{store: store, restyClient: *client}
}

func (v *VMService) RegisterProvider(ctx context.Context, request *server.CreateVMProviderJSONRequestBody) error {
	logger := zap.S().Named("vm_service:register_provider")
	logger.Info("Registering VM service provider")

	newProvider := model.Provider{
		Name:         request.Name,
		Endpoint:     request.Endpoint,
		ID:           uuid.MustParse(request.Id),
		Description:  request.Description,
		ProviderType: request.Type,
	}
	result, err := v.restyClient.R().Get(newProvider.Endpoint + "/health")
	if err != nil || result.StatusCode() != http.StatusOK {
		logger.Error("Failed to get health status or health endpoint return OK status")
		return err
	}
	provider, err := v.store.Provider().Create(ctx, newProvider)
	if err != nil {
		return err
	}
	logger.Info("Successfully registered provider: ", provider.ID)
	return nil

}

func (v *VMService) GetProvider(ctx context.Context, providerID string) (vm.Provider, error) {
	logger := zap.S().Named("vm_service:get_provider")
	logger.Info("Retrieving provider details")

	existingProvider, err := v.store.Provider().Get(ctx, uuid.MustParse(providerID))
	if err != nil {
		return vm.Provider{}, fmt.Errorf("provider %s not found", providerID)
	}
	provider := vm.Provider{
		Name:        existingProvider.Name,
		Endpoint:    existingProvider.Endpoint,
		ProviderID:  existingProvider.ID.String(),
		Description: existingProvider.Description,
		Type:        existingProvider.ProviderType,
	}
	logger.Info("Successfully retrieved provider details")
	return provider, nil
}

func (v *VMService) ListProvider(ctx context.Context) (*[]server.VMProvider, error) {
	logger := zap.S().Named("vm-service-provider:get_providers")
	logger.Info("Retrieving VM Service Providers")
	providers, err := v.store.Provider().List(ctx)
	if err != nil {
		logger.Error("Failed to list providers", err)
		return &[]server.VMProvider{}, err
	}

	var vmProviderList []server.VMProvider
	for _, v := range providers {
		vmProviderList = append(vmProviderList, server.VMProvider{
			Description: v.Description,
			Id:          v.ID.String(),
			Name:        v.Name,
			Type:        v.ProviderType,
		})
	}
	logger.Info("Successfully retrieved VM Service Providers")
	return &vmProviderList, nil
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

	p, err := v.GetProvider(ctx, providerID)
	if err != nil {
		return vm.DeclaredVM{}, err
	}

	response, err := v.restyClient.R().
		SetHeader("Content-ProviderType", "application/json").
		SetBody(&request).
		Post(fmt.Sprintf("%s/vm/create", p.Endpoint))
	if err != nil {
		return vm.DeclaredVM{}, err
	}
	if response.StatusCode() != http.StatusOK {
		return vm.DeclaredVM{}, fmt.Errorf("failed to create VM: %s", response.Status())
	}
	declaredVM := vm.DeclaredVM{ID: request.RequestId, RequestInfo: request}
	// TODO save to database
	logger.Info("Successfully created VM", userRequest.Id)
	return declaredVM, nil
}

func (v *VMService) DeleteVMApplication(ctx context.Context, providerID, appID string) (vm.DeclaredVM, error) {
	logger := zap.S().Named("service-provider:delete_app")
	logger.Info("Deleting VM application", "ID ", appID)
	provider, err := v.GetProvider(ctx, providerID)
	if err != nil {
		logger.Error("Failed to retrieve provider information.")
		return vm.DeclaredVM{}, err
	}
	response, err := v.restyClient.R().
		SetHeader("Content-ProviderType", "application/json").
		Delete(fmt.Sprintf("%s/vm/delete/%s", provider.Endpoint, appID))
	if err != nil {
		return vm.DeclaredVM{}, err
	}
	if response.StatusCode() != http.StatusOK {
		logger.Error("Failed to delete VM application ", "ID ", appID)
		return vm.DeclaredVM{}, fmt.Errorf("failed to create VM: %s", response.Status())
	}

	return vm.DeclaredVM{ID: appID}, nil
}
