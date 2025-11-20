package client

import (
	"context"
	"log"
	"time"
)

// AutoRegistrar handles automatic registration on startup and re-registration
type AutoRegistrar struct {
	client             *Client
	registrations      []Registration
	reregisterInterval time.Duration
	stopCh             chan struct{}
}

// Registration describes a single resource type registration
type Registration struct {
	ResourceKind string
	Request      *RegistrationRequest
}

// AutoRegistrarConfig configuration
type AutoRegistrarConfig struct {
	Client             *Client
	Registrations      []Registration
	ReregisterInterval time.Duration // 0 = no automatic re-registration
}

// NewAutoRegistrar creates an auto-registrar
func NewAutoRegistrar(cfg AutoRegistrarConfig) *AutoRegistrar {
	if cfg.ReregisterInterval == 0 {
		cfg.ReregisterInterval = 5 * time.Minute
	}

	return &AutoRegistrar{
		client:             cfg.Client,
		registrations:      cfg.Registrations,
		reregisterInterval: cfg.ReregisterInterval,
		stopCh:             make(chan struct{}),
	}
}

// Start registers on startup and optionally re-registers periodically
func (a *AutoRegistrar) Start(ctx context.Context) error {
	// Initial registration
	if err := a.registerAll(ctx); err != nil {
		return err
	}

	// Start periodic re-registration if enabled
	if a.reregisterInterval > 0 {
		go a.reregisterLoop(ctx)
	}

	return nil
}

// Stop stops the auto-registrar
func (a *AutoRegistrar) Stop() {
	close(a.stopCh)
}

func (a *AutoRegistrar) registerAll(ctx context.Context) error {
	for _, reg := range a.registrations {
		resp, err := a.client.Register(ctx, reg.ResourceKind, reg.Request)
		if err != nil {
			log.Printf("Failed to register %s: %v", reg.ResourceKind, err)
			return err
		}
		log.Printf("Registered %s for %s: %s", reg.Request.ServiceID, reg.ResourceKind, resp.Message)
	}
	return nil
}

func (a *AutoRegistrar) reregisterLoop(ctx context.Context) {
	ticker := time.NewTicker(a.reregisterInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := a.registerAll(ctx); err != nil {
				log.Printf("Re-registration failed: %v", err)
			}
		case <-a.stopCh:
			return
		case <-ctx.Done():
			return
		}
	}
}

// UnregisterAll unregisters all resource types
func (a *AutoRegistrar) UnregisterAll(ctx context.Context) error {
	for _, reg := range a.registrations {
		if err := a.client.Unregister(ctx, reg.ResourceKind, reg.Request.ServiceID); err != nil {
			log.Printf("Failed to unregister %s: %v", reg.ResourceKind, err)
			// Continue to unregister others
		}
	}
	return nil
}
