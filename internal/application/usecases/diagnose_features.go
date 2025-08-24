package usecases

import (
	"context"
	"historiadorgo/internal/domain/repositories"
)

type DiagnoseFeaturesUseCase struct {
	featureRepo repositories.FeatureManager
}

func NewDiagnoseFeaturesUseCase(featureRepo repositories.FeatureManager) *DiagnoseFeaturesUseCase {
	return &DiagnoseFeaturesUseCase{
		featureRepo: featureRepo,
	}
}

func (uc *DiagnoseFeaturesUseCase) Execute(ctx context.Context, projectKey string) ([]string, error) {
	return uc.featureRepo.ValidateFeatureRequiredFields(ctx, projectKey)
}
