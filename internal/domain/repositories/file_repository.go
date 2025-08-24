package repositories

import (
	"context"
	"historiadorgo/internal/domain/entities"
)

type FileRepository interface {
	ReadFile(ctx context.Context, filePath string) ([]*entities.UserStory, error)
	ValidateFile(ctx context.Context, filePath string) error
	MoveToProcessed(ctx context.Context, filePath string) error
	GetPendingFiles(ctx context.Context, inputDir string) ([]string, error)
}
