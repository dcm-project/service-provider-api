package registration

import (
	"context"
	"time"

	"github.com/dcm-project/service-provider-api/internal/store"
	"github.com/dcm-project/service-provider-api/internal/store/model"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// SeedDefaultCatalogItems creates default catalog items if they don't exist
// This should be called during application initialization
func SeedDefaultCatalogItems(s store.Store) error {
	logger := zap.S().Named("catalog_seed")
	ctx := context.Background()

	// Convert catalog definitions to model.CatalogItem
	defaultItems := make([]model.CatalogItem, len(DefaultCatalogDefinitions))
	for i, def := range DefaultCatalogDefinitions {
		defaultItems[i] = model.CatalogItem{
			ID:           uuid.New(),
			Name:         def.Name,
			DisplayName:  def.DisplayName,
			Description:  def.Description,
			ResourceKind: def.ResourceKind,
			Active:       true,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
	}

	catalogAdapter := NewRegistrationCatalogAdapter(s)

	for _, item := range defaultItems {
		// Check if item already exists
		_, err := s.Catalog().GetCatalogItem(ctx, item.Name)

		if err != nil {
			// Item doesn't exist, create it
			if err := catalogAdapter.CreateCatalogItem(ctx, &item); err != nil {
				logger.Errorw("Failed to create catalog item",
					"name", item.Name,
					"error", err,
				)
				return err
			}
			logger.Infow("Created catalog item",
				"name", item.Name,
				"display_name", item.DisplayName,
			)
		} else {
			logger.Debugw("Catalog item already exists, skipping",
				"name", item.Name,
			)
		}
	}

	logger.Info("Catalog seeding completed")
	return nil
}
