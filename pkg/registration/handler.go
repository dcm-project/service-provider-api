package registration

import (
	"context"
	"fmt"
	"time"

	"github.com/dcm-project/service-provider-api/internal/api/server"
)

// RegistrationError represents errors that occur during provider registration
type RegistrationError struct {
	Code    string
	Message string
	Err     error
}

func (e *RegistrationError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *RegistrationError) Unwrap() error {
	return e.Err
}

// Common registration error codes
const (
	ErrCodeValidation          = "VALIDATION_ERROR"
	ErrCodeNotFound            = "PROVIDER_NOT_FOUND"
	ErrCodeRegistryUpdate      = "REGISTRY_UPDATE_FAILED"
	ErrCodeCatalogUpdate       = "CATALOG_UPDATE_FAILED"
	ErrCodeEndpointUnreachable = "ENDPOINT_UNREACHABLE"
)

// RegisteredProvider represents a service registered in the Resource Registry (domain model)
// This extends the OpenAPI type with additional fields needed internally
type RegisteredProvider struct {
	ServiceID    string
	ResourceKind string
	Endpoint     string
	Metadata     server.ProviderMetadata
	Operations   []string
	CatalogItem  string
	Status       string
	RegisteredAt time.Time
	UpdatedAt    time.Time
}

// RegistryStore interface for Resource Registry operations
// This allows the package to be agnostic of the actual storage implementation
type RegistryStore interface {
	// UpsertProvider creates or updates a provider registration in the Resource Registry
	UpsertProvider(ctx context.Context, provider RegisteredProvider) error

	// GetProvider retrieves a service by ID and resource kind
	GetProvider(ctx context.Context, serviceID, resourceKind string) (*RegisteredProvider, error)

	// DeleteProvider removes a service registration
	DeleteProvider(ctx context.Context, serviceID, resourceKind string) error

	// ListProviders lists all registered providers for a resource kind
	ListProviders(ctx context.Context, resourceKind string) ([]RegisteredProvider, error)
}

// CatalogStore interface for Service Catalog operations
// This allows the package to be agnostic of the actual catalog implementation
type CatalogStore interface {
	// UpdateCatalogMapping updates which catalog item a service can fulfill
	UpdateCatalogMapping(ctx context.Context, serviceID, resourceKind string, catalogItem string) error

	// RemoveCatalogMapping removes service mapping from catalog
	RemoveCatalogMapping(ctx context.Context, serviceID, resourceKind string) error
}

// EndpointChecker validates that a provider endpoint is reachable
type EndpointChecker interface {
	// CheckEndpoint verifies the provider endpoint is reachable and healthy
	CheckEndpoint(ctx context.Context, endpoint string) error
}

// Handler handles Resource Provider registration requests
type Handler struct {
	registryStore   RegistryStore
	catalogStore    CatalogStore
	validator       *Validator
	endpointChecker EndpointChecker
}

func newValidationError(message string, err error) *RegistrationError {
	return &RegistrationError{Code: ErrCodeValidation, Message: message, Err: err}
}

func newRegistryUpdateError(message string, err error) *RegistrationError {
	return &RegistrationError{Code: ErrCodeRegistryUpdate, Message: message, Err: err}
}

func newCatalogUpdateError(message string, err error) *RegistrationError {
	return &RegistrationError{Code: ErrCodeCatalogUpdate, Message: message, Err: err}
}

func newEndpointUnreachableError(endpoint string, err error) *RegistrationError {
	return &RegistrationError{
		Code:    ErrCodeEndpointUnreachable,
		Message: fmt.Sprintf("Provider endpoint %s is unreachable", endpoint),
		Err:     err,
	}
}

// Config for creating a new Handler
type Config struct {
	RegistryStore   RegistryStore
	CatalogStore    CatalogStore
	Validator       *Validator
	EndpointChecker EndpointChecker
}

// NewHandler creates a new registration handler
func NewHandler(cfg Config) (*Handler, error) {
	if cfg.RegistryStore == nil {
		return nil, fmt.Errorf("RegistryStore is required")
	}
	if cfg.CatalogStore == nil {
		return nil, fmt.Errorf("CatalogStore is required")
	}

	validator := cfg.Validator
	if validator == nil {
		validator = NewValidator()
	}

	return &Handler{
		registryStore:   cfg.RegistryStore,
		catalogStore:    cfg.CatalogStore,
		validator:       validator,
		endpointChecker: cfg.EndpointChecker,
	}, nil
}

// Register handles a provider registration request
// This implements the idempotent registration flow described in the ADR
func (h *Handler) Register(ctx context.Context, serviceID, resourceKind, endpoint string, metadata server.ProviderMetadata, operations []string) (*server.RegistrationResponse, error) {
	// 1. Validate the request
	if err := h.validator.ValidateRegistration(serviceID, resourceKind, endpoint, metadata, operations); err != nil {
		return nil, err
	}

	// 2. Optional: Check endpoint reachability
	if h.endpointChecker != nil {
		if err := h.endpointChecker.CheckEndpoint(ctx, endpoint); err != nil {
			return nil, newEndpointUnreachableError(endpoint, err)
		}
	}

	// 3. Check if service already exists (for idempotent registration)
	existingProvider, err := h.registryStore.GetProvider(ctx, serviceID, resourceKind)
	isUpdate := err == nil && existingProvider != nil

	// 4. Create or update service in Resource Registry
	now := time.Now()

	// Catalog item is derived from resource kind (no redundancy in API)
	catalogItem := resourceKind

	registeredProvider := RegisteredProvider{
		ServiceID:    serviceID,
		ResourceKind: resourceKind,
		Endpoint:     endpoint,
		Metadata:     metadata,
		Operations:   operations,
		CatalogItem:  catalogItem,
		Status:       "active",
		RegisteredAt: now,
		UpdatedAt:    now,
	}

	// Preserve original registration time if updating
	if isUpdate {
		registeredProvider.RegisteredAt = existingProvider.RegisteredAt
	}

	if err := h.registryStore.UpsertProvider(ctx, registeredProvider); err != nil {
		return nil, newRegistryUpdateError("failed to update Resource Registry", err)
	}

	// 5. Update Service Catalog mapping
	if err := h.catalogStore.UpdateCatalogMapping(ctx, serviceID, resourceKind, catalogItem); err != nil {
		// Note: this is a partial failure - registry is updated but catalog isn't
		// In production, consider using a transaction or saga pattern
		return nil, newCatalogUpdateError("failed to update Service Catalog", err)
	}

	// 6. Build response
	message := "Service registered successfully"
	if isUpdate {
		message = "Service registration updated successfully"
	}

	status := "active"
	return &server.RegistrationResponse{
		ServiceId:    &serviceID,
		ResourceKind: &resourceKind,
		Status:       &status,
		RegisteredAt: &registeredProvider.RegisteredAt,
		Message:      &message,
	}, nil
}

// Unregister removes a service registration
func (h *Handler) Unregister(ctx context.Context, serviceID, resourceKind string) error {
	// 1. Validate inputs
	if serviceID == "" {
		return newValidationError("serviceID is required", nil)
	}
	if resourceKind == "" {
		return newValidationError("resourceKind is required", nil)
	}

	// 2. Check if service exists
	_, err := h.registryStore.GetProvider(ctx, serviceID, resourceKind)
	if err != nil {
		return &RegistrationError{
			Code:    ErrCodeNotFound,
			Message: fmt.Sprintf("Service %s with resource kind %s not found", serviceID, resourceKind),
			Err:     err,
		}
	}

	// 3. Remove from catalog
	if err := h.catalogStore.RemoveCatalogMapping(ctx, serviceID, resourceKind); err != nil {
		return newCatalogUpdateError("failed to remove catalog mappings", err)
	}

	// 4. Remove from registry
	if err := h.registryStore.DeleteProvider(ctx, serviceID, resourceKind); err != nil {
		return newRegistryUpdateError("failed to remove from registry", err)
	}

	return nil
}

// GetRegistration retrieves a service registration
func (h *Handler) GetRegistration(ctx context.Context, serviceID, resourceKind string) (*RegisteredProvider, error) {
	if serviceID == "" {
		return nil, newValidationError("serviceID is required", nil)
	}
	if resourceKind == "" {
		return nil, newValidationError("resourceKind is required", nil)
	}

	provider, err := h.registryStore.GetProvider(ctx, serviceID, resourceKind)
	if err != nil {
		return nil, &RegistrationError{
			Code:    ErrCodeNotFound,
			Message: fmt.Sprintf("Service %s with resource kind %s not found", serviceID, resourceKind),
			Err:     err,
		}
	}

	return provider, nil
}

// ListRegistrations lists all registrations for a resource kind
func (h *Handler) ListRegistrations(ctx context.Context, resourceKind string) ([]RegisteredProvider, error) {
	if resourceKind == "" {
		return nil, newValidationError("resourceKind is required", nil)
	}

	providers, err := h.registryStore.ListProviders(ctx, resourceKind)
	if err != nil {
		return nil, newRegistryUpdateError("failed to list providers", err)
	}

	return providers, nil
}
