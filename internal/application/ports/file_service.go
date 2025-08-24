package ports

import (
	"context"
	"historiadorgo/internal/domain/entities"
)

type FileService interface {
	ProcessFile(ctx context.Context, filePath, projectKey string, dryRun bool) (*entities.BatchResult, error)
	ValidateFile(ctx context.Context, filePath string) error
	GetPendingFiles(ctx context.Context, inputDir string) ([]string, error)
}
