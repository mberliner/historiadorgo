package ports

import (
	"context"
	"historiadorgo/internal/domain/entities"
)

type JiraService interface {
	TestConnection(ctx context.Context) error
	ValidateProject(ctx context.Context, projectKey string) error
	ProcessUserStory(ctx context.Context, story *entities.UserStory, projectKey string, rowNumber int) (*entities.ProcessResult, error)
	DiagnoseFeatures(ctx context.Context, projectKey string) ([]string, error)
}
