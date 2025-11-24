package service

import (
	"time"

	"github.com/dcm-project/service-provider-api/internal/store"
	"github.com/dcm-project/service-provider-api/internal/store/registration"
	pkgregistration "github.com/dcm-project/service-provider-api/pkg/registration"
)

// RegistrationServiceConfig configuration for registration service
type RegistrationServiceConfig struct {
	// Store for database operations
	Store store.Store
	
	// EndpointCheckEnabled whether to check endpoint reachability during registration
	EndpointCheckEnabled bool
	
	// EndpointCheckTimeout timeout for endpoint health checks
	EndpointCheckTimeout time.Duration
}

// InitializeRegistrationService creates and configures the registration handler
func InitializeRegistrationService(cfg RegistrationServiceConfig) (*pkgregistration.Handler, error) {
	// Create store adapters
	registryStore := registration.NewRegistrationRegistryAdapter(cfg.Store)
	catalogStore := registration.NewRegistrationCatalogAdapter(cfg.Store)
	
	// Create validator
	validator := pkgregistration.NewValidator()
	
	// Create endpoint checker if enabled
	var endpointChecker pkgregistration.EndpointChecker
	if cfg.EndpointCheckEnabled {
		timeout := cfg.EndpointCheckTimeout
		if timeout == 0 {
			timeout = 10 * time.Second
		}
		endpointChecker = NewHTTPEndpointChecker(timeout)
	}
	
	// Create registration handler
	registrationHandler, err := pkgregistration.NewHandler(pkgregistration.Config{
		RegistryStore:   registryStore,
		CatalogStore:    catalogStore,
		Validator:       validator,
		EndpointChecker: endpointChecker,
	})
	if err != nil {
		return nil, err
	}
	
	return registrationHandler, nil
}

// DefaultRegistrationServiceConfig returns default configuration
func DefaultRegistrationServiceConfig(store store.Store) RegistrationServiceConfig {
	return RegistrationServiceConfig{
		Store:                store,
		EndpointCheckEnabled: true,
		EndpointCheckTimeout: 10 * time.Second,
	}
}

