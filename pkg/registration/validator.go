package registration

import (
	"fmt"
	"net/url"

	"github.com/dcm-project/service-provider-api/internal/api/server"
	"github.com/google/uuid"
)

// Validator validates registration requests
type Validator struct{}

// NewValidator creates a new Validator
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateRegistration validates a registration request
func (v *Validator) ValidateRegistration(serviceID, resourceKind, endpoint string, metadata server.ProviderMetadata, operations []string) error {
	if err := v.validateServiceID(serviceID); err != nil {
		return newValidationError("invalid service_id", err)
	}

	if err := v.validateResourceKind(resourceKind); err != nil {
		return newValidationError("invalid resource_kind", err)
	}

	if err := v.validateEndpoint(endpoint); err != nil {
		return newValidationError("invalid endpoint", err)
	}

	if err := v.validateOperations(operations); err != nil {
		return newValidationError("invalid operations", err)
	}

	if err := v.validateMetadata(metadata); err != nil {
		return newValidationError("invalid metadata", err)
	}

	return nil
}

func (v *Validator) validateServiceID(serviceID string) error {
	if serviceID == "" {
		return fmt.Errorf("service_id is required")
	}

	if _, err := uuid.Parse(serviceID); err != nil {
		return fmt.Errorf("service_id must be a valid UUID: %w", err)
	}

	return nil
}

func (v *Validator) validateResourceKind(resourceKind string) error {
	if resourceKind == "" {
		return fmt.Errorf("resource_kind is required")
	}
	return nil
}

func (v *Validator) validateEndpoint(endpoint string) error {
	if endpoint == "" {
		return fmt.Errorf("endpoint is required")
	}

	parsedURL, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("endpoint must be a valid URL: %w", err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("endpoint must use http or https scheme")
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("endpoint must have a valid host")
	}

	return nil
}

func (v *Validator) validateOperations(operations []string) error {
	if len(operations) == 0 {
		return fmt.Errorf("at least one operation is required")
	}

	for _, op := range operations {
		if op == "" {
			return fmt.Errorf("operation cannot be empty")
		}
	}

	return nil
}

func (v *Validator) validateMetadata(metadata server.ProviderMetadata) error {
	if metadata.Zone == "" {
		return fmt.Errorf("metadata.zone is required")
	}

	if metadata.Region == "" {
		return fmt.Errorf("metadata.region is required")
	}

	return nil
}
