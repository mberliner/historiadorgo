package repositories

import (
	"context"
	"historiadorgo/internal/domain/entities"
)

type FeatureManager interface {
	CreateOrGetFeature(ctx context.Context, description, projectKey string) (*entities.FeatureResult, error)
	SearchExistingFeature(ctx context.Context, description, projectKey string) (string, error)
	ValidateFeatureRequiredFields(ctx context.Context, projectKey string) ([]string, error)
}
