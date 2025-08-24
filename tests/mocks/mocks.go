package mocks

import (
	"context"
	"historiadorgo/internal/domain/entities"
)

// MockFileRepository is a mock implementation of repositories.FileRepository
type MockFileRepository struct {
	ReadFileFunc        func(ctx context.Context, filePath string) ([]*entities.UserStory, error)
	ValidateFileFunc    func(ctx context.Context, filePath string) error
	MoveToProcessedFunc func(ctx context.Context, filePath string) error
	GetPendingFilesFunc func(ctx context.Context, inputDir string) ([]string, error)
}

func (m *MockFileRepository) ReadFile(ctx context.Context, filePath string) ([]*entities.UserStory, error) {
	if m.ReadFileFunc != nil {
		return m.ReadFileFunc(ctx, filePath)
	}
	return nil, nil
}

func (m *MockFileRepository) ValidateFile(ctx context.Context, filePath string) error {
	if m.ValidateFileFunc != nil {
		return m.ValidateFileFunc(ctx, filePath)
	}
	return nil
}

func (m *MockFileRepository) MoveToProcessed(ctx context.Context, filePath string) error {
	if m.MoveToProcessedFunc != nil {
		return m.MoveToProcessedFunc(ctx, filePath)
	}
	return nil
}

func (m *MockFileRepository) GetPendingFiles(ctx context.Context, inputDir string) ([]string, error) {
	if m.GetPendingFilesFunc != nil {
		return m.GetPendingFilesFunc(ctx, inputDir)
	}
	return nil, nil
}

// MockJiraRepository is a mock implementation of repositories.JiraRepository
type MockJiraRepository struct {
	TestConnectionFunc           func(ctx context.Context) error
	ValidateProjectFunc          func(ctx context.Context, projectKey string) error
	ValidateSubtaskIssueTypeFunc func(ctx context.Context, projectKey string) error
	ValidateFeatureIssueTypeFunc func(ctx context.Context) error
	ValidateParentIssueFunc      func(ctx context.Context, issueKey string) error
	CreateUserStoryFunc          func(ctx context.Context, story *entities.UserStory, rowNumber int) (*entities.ProcessResult, error)
	GetIssueTypesFunc            func(ctx context.Context) ([]map[string]interface{}, error)
}

func (m *MockJiraRepository) TestConnection(ctx context.Context) error {
	if m.TestConnectionFunc != nil {
		return m.TestConnectionFunc(ctx)
	}
	return nil
}

func (m *MockJiraRepository) ValidateProject(ctx context.Context, projectKey string) error {
	if m.ValidateProjectFunc != nil {
		return m.ValidateProjectFunc(ctx, projectKey)
	}
	return nil
}

func (m *MockJiraRepository) ValidateSubtaskIssueType(ctx context.Context, projectKey string) error {
	if m.ValidateSubtaskIssueTypeFunc != nil {
		return m.ValidateSubtaskIssueTypeFunc(ctx, projectKey)
	}
	return nil
}

func (m *MockJiraRepository) ValidateFeatureIssueType(ctx context.Context) error {
	if m.ValidateFeatureIssueTypeFunc != nil {
		return m.ValidateFeatureIssueTypeFunc(ctx)
	}
	return nil
}

func (m *MockJiraRepository) ValidateParentIssue(ctx context.Context, issueKey string) error {
	if m.ValidateParentIssueFunc != nil {
		return m.ValidateParentIssueFunc(ctx, issueKey)
	}
	return nil
}

func (m *MockJiraRepository) CreateUserStory(ctx context.Context, story *entities.UserStory, rowNumber int) (*entities.ProcessResult, error) {
	if m.CreateUserStoryFunc != nil {
		return m.CreateUserStoryFunc(ctx, story, rowNumber)
	}
	return nil, nil
}

func (m *MockJiraRepository) GetIssueTypes(ctx context.Context) ([]map[string]interface{}, error) {
	if m.GetIssueTypesFunc != nil {
		return m.GetIssueTypesFunc(ctx)
	}
	return nil, nil
}

// MockFeatureManager is a mock implementation of repositories.FeatureManager
type MockFeatureManager struct {
	CreateOrGetFeatureFunc            func(ctx context.Context, description, projectKey string) (*entities.FeatureResult, error)
	SearchExistingFeatureFunc         func(ctx context.Context, description, projectKey string) (string, error)
	ValidateFeatureRequiredFieldsFunc func(ctx context.Context, projectKey string) ([]string, error)
}

func (m *MockFeatureManager) CreateOrGetFeature(ctx context.Context, description, projectKey string) (*entities.FeatureResult, error) {
	if m.CreateOrGetFeatureFunc != nil {
		return m.CreateOrGetFeatureFunc(ctx, description, projectKey)
	}
	return nil, nil
}

func (m *MockFeatureManager) SearchExistingFeature(ctx context.Context, description, projectKey string) (string, error) {
	if m.SearchExistingFeatureFunc != nil {
		return m.SearchExistingFeatureFunc(ctx, description, projectKey)
	}
	return "", nil
}

func (m *MockFeatureManager) ValidateFeatureRequiredFields(ctx context.Context, projectKey string) ([]string, error) {
	if m.ValidateFeatureRequiredFieldsFunc != nil {
		return m.ValidateFeatureRequiredFieldsFunc(ctx, projectKey)
	}
	return nil, nil
}
