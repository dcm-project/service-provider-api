package service

import (
	"context"

	"github.com/dcm-project/service-provider-api/internal/deploy"
	"github.com/dcm-project/service-provider-api/internal/store"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ProviderService struct {
	store  store.Store
	deploy *deploy.DeployService
}

func NewProviderService(store store.Store, deploy *deploy.DeployService) *ProviderService {
	return &ProviderService{store: store, deploy: deploy}
}

func (s *ProviderService) GetProviders(ctx context.Context) error {
	logger := zap.S().Named("service-provider:get_providers")
	logger.Info("Starting...")
	// TODO
	return nil
}

func (s *ProviderService) GetProvider(ctx context.Context) error {
	logger := zap.S().Named("service-provider:get_provider")
	logger.Info("Get...")
	// TODO
	return nil
}

func (s *ProviderService) CreateApplication(ctx context.Context) error {
	logger := zap.S().Named("service-provider:create_app")
	logger.Info("Starting...")
	// TODO
	return nil
}

func (s *ProviderService) DeleteApplication(ctx context.Context, id uuid.UUID) error {
	logger := zap.S().Named("service-provider:delete_app")
	logger.Info("Starting...")
	// TODO
	return nil
}
