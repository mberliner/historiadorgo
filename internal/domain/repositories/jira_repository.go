package repositories

import (
	"context"
	"historiadorgo/internal/domain/entities"
)

type JiraRepository interface {
	TestConnection(ctx context.Context) error
	ValidateProject(ctx context.Context, projectKey string) error
	ValidateSubtaskIssueType(ctx context.Context, projectKey string) error
	ValidateFeatureIssueType(ctx context.Context) error
	ValidateParentIssue(ctx context.Context, issueKey string) error
	CreateUserStory(ctx context.Context, story *entities.UserStory, rowNumber int) (*entities.ProcessResult, error)
	GetIssueTypes(ctx context.Context) ([]map[string]interface{}, error)
}
