// Package resource_manager implements business logic for service type instance management.
package resource_manager

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dcm-project/service-provider-manager/api/v1alpha1/resource_manager"
	"github.com/dcm-project/service-provider-manager/internal/logging"
	"github.com/dcm-project/service-provider-manager/internal/service"
	"github.com/dcm-project/service-provider-manager/internal/store"
	"github.com/dcm-project/service-provider-manager/internal/store/model"
	rmstore "github.com/dcm-project/service-provider-manager/internal/store/resource_manager"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
)

type InstanceService struct {
	store      store.Store
	httpClient *resty.Client
}

func NewInstanceService(store store.Store) *InstanceService {
	client := resty.New().
		SetTimeout(30 * time.Second).
		SetRetryCount(3).
		SetRetryWaitTime(2 * time.Second).
		SetRetryMaxWaitTime(30 * time.Second)

	return &InstanceService{
		store:      store,
		httpClient: client,
	}
}

// CreateInstance creates a new service type instance
func (s *InstanceService) CreateInstance(ctx context.Context, request *resource_manager.ServiceTypeInstance, queryID *string) (*resource_manager.ServiceTypeInstance, error) {
	log := logging.FromContext(ctx)
	providerName := request.ProviderName

	log.Debug("Creating instance", "provider_name", providerName)

	provider, err := s.store.Provider().GetByName(ctx, providerName)
	if err != nil {
		if errors.Is(err, store.ErrProviderNotFound) {
			return nil, service.NewNotFoundError(fmt.Sprintf("provider '%s' not found", providerName))
		}
		log.Error("Failed to retrieve provider", "provider_name", providerName, "error", err)
		return nil, service.NewInternalError(fmt.Sprintf("failed to retrieve provider: %v", err))
	}

	// Check Provider if provider is not in ready state
	if provider.HealthStatus != model.HealthStatusReady {
		log.Warn("Provider not in ready state", "provider_name", providerName, "health_status", provider.HealthStatus)
		return nil, service.NewProviderError(fmt.Sprintf("provider '%s' is not in ready state (current status: %s)", providerName, provider.HealthStatus))
	}

	// Resolve instance ID
	instanceID, err := s.resolveInstanceID(ctx, queryID)
	if err != nil {
		return nil, err
	}

	// Send request to provider endpoint with the resolved ID
	providerResponse, err := s.createInstanceWithProvider(ctx, provider.Endpoint, request, instanceID)
	if err != nil {
		log.Error("Provider provisioning failed", "instance_id", *instanceID, "provider_name", providerName, "error", err)
		return nil, service.NewProviderError(fmt.Sprintf("Error from Provider (%s): %v", providerName, err))
	}

	// Create instance in database
	instance := model.ServiceTypeInstance{
		ID:           *instanceID,
		ProviderName: providerName,
		Status:       providerResponse.Status,
		Spec:         request.Spec,
	}

	created, err := s.store.ServiceTypeInstance().Create(ctx, instance)
	if err != nil {
		log.Error("Failed to create instance in store", "instance_id", *instanceID, "error", err)
		return nil, service.NewInternalError(fmt.Sprintf("failed to create database record for instance %s: %v", providerResponse.ID, err))
	}

	log.Info("Instance created successfully",
		"instance_id", created.ID,
		"provider_name", providerName,
		"status", providerResponse.Status,
	)
	return ModelToAPI(created), nil
}

// GetInstance retrieves an instance by ID
func (s *InstanceService) GetInstance(ctx context.Context, instanceID string) (*resource_manager.ServiceTypeInstance, error) {
	log := logging.FromContext(ctx)
	log.Debug("Getting instance", "instance_id", instanceID)

	instance, err := s.store.ServiceTypeInstance().Get(ctx, instanceID)
	if err != nil {
		if errors.Is(err, rmstore.ErrInstanceNotFound) {
			return nil, service.NewNotFoundError(fmt.Sprintf("instance %s not found", instanceID))
		}
		log.Error("Failed to get instance from store", "instance_id", instanceID, "error", err)
		return nil, service.NewInternalError(fmt.Sprintf("failed to retrieve instance: %v", err))
	}

	return ModelToAPI(instance), nil
}

// ListInstances returns instances with optional filtering and pagination
func (s *InstanceService) ListInstances(ctx context.Context, providerName *string, maxPageSize *int, pageToken *string) (*resource_manager.ServiceTypeInstanceList, error) {
	log := logging.FromContext(ctx)
	log.Debug("Listing instances",
		"provider_filter", providerName,
		"page_size", maxPageSize,
	)

	opts := &rmstore.ServiceTypeInstanceListOptions{
		ProviderName: providerName,
	}

	// Apply max page size (default 50, max 100)
	if maxPageSize != nil {
		if *maxPageSize > 0 && *maxPageSize <= 100 {
			opts.PageSize = *maxPageSize
		} else {
			return nil, service.NewValidationError("page size must be between 1 and 100")
		}
	}

	// Apply page token
	if pageToken != nil && *pageToken != "" {
		opts.PageToken = pageToken
	}

	result, err := s.store.ServiceTypeInstance().List(ctx, opts)
	if err != nil {
		log.Error("Failed to list instances from store", "error", err)
		return nil, service.NewInternalError(fmt.Sprintf("failed to list instances: %v", err))
	}

	// Convert to API types
	apiInstances := make([]resource_manager.ServiceTypeInstance, len(result.Instances))
	for i, inst := range result.Instances {
		apiInstances[i] = *ModelToAPI(&inst)
	}

	log.Debug("Instances listed",
		"count", len(apiInstances),
		"has_next_page", result.NextPageToken != nil,
	)
	apiResult := &resource_manager.ServiceTypeInstanceList{
		Instances:     &apiInstances,
		NextPageToken: result.NextPageToken,
	}

	return apiResult, nil
}

// DeleteInstance removes an instance by ID
func (s *InstanceService) DeleteInstance(ctx context.Context, instanceID string) error {
	log := logging.FromContext(ctx)
	log.Debug("Deleting instance", "instance_id", instanceID)

	instance, err := s.store.ServiceTypeInstance().Get(ctx, instanceID)
	if err != nil {
		if errors.Is(err, rmstore.ErrInstanceNotFound) {
			return service.NewNotFoundError(fmt.Sprintf("instance %s not found", instanceID))
		}
		log.Error("Failed to get instance for deletion", "instance_id", instanceID, "error", err)
		return service.NewInternalError(fmt.Sprintf("failed to retrieve instance: %v", err))
	}

	// Get provider to send delete request
	provider, err := s.store.Provider().GetByName(ctx, instance.ProviderName)
	if err != nil && !errors.Is(err, store.ErrProviderNotFound) {
		log.Error("Failed to retrieve provider for deletion", "instance_id", instanceID, "provider_name", instance.ProviderName, "error", err)
		return service.NewInternalError(fmt.Sprintf("failed to retrieve provider: %v", err))
	}

	// Send delete request to provider if provider still exists
	if provider != nil {
		log.Debug("Deleting instance from provider", "instance_id", instanceID, "provider_name", provider.Name)
		err = s.deleteInstanceWithProvider(ctx, provider.Endpoint, instanceID)
		if err != nil {
			log.Error("Provider deletion failed", "instance_id", instanceID, "provider_name", provider.Name, "error", err)
			return service.NewProviderError(fmt.Sprintf("failed to delete instance (%s): %v", instanceID, err))
		}
	}

	// Delete from database
	err = s.store.ServiceTypeInstance().Delete(ctx, instanceID)
	if err != nil {
		log.Error("Failed to delete instance from store", "instance_id", instanceID, "error", err)
		return service.NewInternalError(fmt.Sprintf("failed to delete database record for instance %s: %v", instanceID, err))
	}

	log.Info("Instance deleted successfully",
		"instance_id", instanceID,
		"provider_name", instance.ProviderName,
	)
	return nil
}

// resolveInstanceID returns the requested ID after checking for conflicts, or generates a new one
func (s *InstanceService) resolveInstanceID(ctx context.Context, queryID *string) (*string, error) {
	log := logging.FromContext(ctx)

	if queryID == nil || *queryID == "" {
		generatedId := uuid.New().String()
		log.Debug("Generated instance ID", "instance_id", generatedId)
		return &generatedId, nil
	}

	requestedID := *queryID

	exists, err := s.store.ServiceTypeInstance().ExistsByID(ctx, requestedID)
	if err != nil {
		log.Error("Failed to check instance ID existence", "instance_id", requestedID, "error", err)
		return nil, service.NewInternalError(fmt.Sprintf("failed to check instance existence: %v", err))
	}
	if exists {
		log.Warn("Duplicate instance ID", "instance_id", requestedID)
		return nil, service.NewConflictError(fmt.Sprintf("instance with ID '%s' already exists", requestedID))
	}

	return &requestedID, nil
}

// createInstanceWithProvider sends the create request to the provider's endpoint
func (s *InstanceService) createInstanceWithProvider(ctx context.Context, endpoint string, request *resource_manager.ServiceTypeInstance, id *string) (*ProviderResponse, error) {
	log := logging.FromContext(ctx)

	var providerResp ProviderResponse

	resp, err := s.httpClient.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetQueryParam("id", *id).
		SetBody(map[string]interface{}{"spec": request.Spec}).
		SetResult(&providerResp).
		Post(endpoint)
	if err != nil {
		log.Error("Failed to connect to provider", "endpoint", endpoint, "error", err)
		return nil, service.NewProviderError(fmt.Sprintf("failed to connect to provider: %v", err))
	}

	if resp.IsError() {
		log.Error("Provider returned error", "endpoint", endpoint, "status", resp.Status())
		return nil, service.NewProviderError(fmt.Sprintf("provider returned error: %s", resp.Status()))
	}

	return &providerResp, nil
}

// deleteInstanceWithProvider sends the delete request to the provider's endpoint
func (s *InstanceService) deleteInstanceWithProvider(ctx context.Context, endpoint string, instanceID string) error {
	log := logging.FromContext(ctx)

	resp, err := s.httpClient.R().
		SetContext(ctx).
		Delete(fmt.Sprintf("%s/%s", endpoint, instanceID))
	if err != nil {
		log.Error("Failed to connect to provider for deletion", "endpoint", endpoint, "instance_id", instanceID, "error", err)
		return fmt.Errorf("failed to connect to provider: %w", err)
	}

	if resp.IsError() && resp.StatusCode() != 404 {
		log.Error("Provider returned error on deletion", "endpoint", endpoint, "instance_id", instanceID, "status", resp.Status())
		return fmt.Errorf("provider returned error: %s", resp.Status())
	}

	return nil
}
