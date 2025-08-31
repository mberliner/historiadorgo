package mocks

import (
	"context"
	"time"
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

// MockProcessFilesUseCase is a mock implementation of ProcessFilesUseCase
type MockProcessFilesUseCase struct {
	ExecuteFunc         func(ctx context.Context, filePath, projectKey string, dryRun bool) (*entities.BatchResult, error)
	ProcessAllFilesFunc func(ctx context.Context, inputDir, projectKey string, dryRun bool) ([]*entities.BatchResult, error)
}

func (m *MockProcessFilesUseCase) Execute(ctx context.Context, filePath, projectKey string, dryRun bool) (*entities.BatchResult, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, filePath, projectKey, dryRun)
	}
	return nil, nil
}

func (m *MockProcessFilesUseCase) ProcessAllFiles(ctx context.Context, inputDir, projectKey string, dryRun bool) ([]*entities.BatchResult, error) {
	if m.ProcessAllFilesFunc != nil {
		return m.ProcessAllFilesFunc(ctx, inputDir, projectKey, dryRun)
	}
	return nil, nil
}

// MockValidateFileUseCase is a mock implementation of ValidateFileUseCase
type MockValidateFileUseCase struct {
	ExecuteFunc func(ctx context.Context, filePath, projectKey string) (*entities.BatchResult, error)
}

func (m *MockValidateFileUseCase) Execute(ctx context.Context, filePath, projectKey string) (*entities.BatchResult, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, filePath, projectKey)
	}
	return nil, nil
}

// MockTestConnectionUseCase is a mock implementation of TestConnectionUseCase
type MockTestConnectionUseCase struct {
	ExecuteFunc func(ctx context.Context) error
}

func (m *MockTestConnectionUseCase) Execute(ctx context.Context) error {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx)
	}
	return nil
}

// DiagnosisResult represents the result of feature diagnosis
type DiagnosisResult struct {
	ProjectKey     string
	RequiredFields []string
}

// MockDiagnoseFeaturesUseCase is a mock implementation of DiagnoseFeaturesUseCase
type MockDiagnoseFeaturesUseCase struct {
	ExecuteFunc func(ctx context.Context, projectKey string) (*DiagnosisResult, error)
}

func (m *MockDiagnoseFeaturesUseCase) Execute(ctx context.Context, projectKey string) (*DiagnosisResult, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, projectKey)
	}
	return nil, nil
}

// MockLogger is a mock implementation of Logger
type MockLogger struct {
	LogCommandStartFunc        func(cmd string, params map[string]interface{})
	LogCommandEndFunc          func(cmd string, success bool, duration time.Duration)
	InfoFunc                   func(msg string)
	WriteFormattedOutputFunc   func(output string)
}

func (m *MockLogger) LogCommandStart(cmd string, params map[string]interface{}) {
	if m.LogCommandStartFunc != nil {
		m.LogCommandStartFunc(cmd, params)
	}
}

func (m *MockLogger) LogCommandEnd(cmd string, success bool, duration time.Duration) {
	if m.LogCommandEndFunc != nil {
		m.LogCommandEndFunc(cmd, success, duration)
	}
}

func (m *MockLogger) Info(msg string) {
	if m.InfoFunc != nil {
		m.InfoFunc(msg)
	}
}

func (m *MockLogger) WriteFormattedOutput(output string) {
	if m.WriteFormattedOutputFunc != nil {
		m.WriteFormattedOutputFunc(output)
	}
}

// MockOutputFormatter is a mock implementation of OutputFormatter
type MockOutputFormatter struct {
	FormatBatchResultFunc           func(result *entities.BatchResult) string
	FormatMultipleBatchResultsFunc  func(results []*entities.BatchResult) string
	FormatValidationFunc            func(result *entities.BatchResult) string
	FormatConnectionTestFunc        func(success bool, error string) string
	FormatDiagnosisFunc             func(result *DiagnosisResult) string
}

func (m *MockOutputFormatter) FormatBatchResult(result *entities.BatchResult) string {
	if m.FormatBatchResultFunc != nil {
		return m.FormatBatchResultFunc(result)
	}
	return ""
}

func (m *MockOutputFormatter) FormatMultipleBatchResults(results []*entities.BatchResult) string {
	if m.FormatMultipleBatchResultsFunc != nil {
		return m.FormatMultipleBatchResultsFunc(results)
	}
	return ""
}

func (m *MockOutputFormatter) FormatValidation(result *entities.BatchResult) string {
	if m.FormatValidationFunc != nil {
		return m.FormatValidationFunc(result)
	}
	return ""
}

func (m *MockOutputFormatter) FormatConnectionTest(success bool, error string) string {
	if m.FormatConnectionTestFunc != nil {
		return m.FormatConnectionTestFunc(success, error)
	}
	return ""
}

func (m *MockOutputFormatter) FormatDiagnosis(result *DiagnosisResult) string {
	if m.FormatDiagnosisFunc != nil {
		return m.FormatDiagnosisFunc(result)
	}
	return ""
}

// MockConfig is a mock implementation of Config
type MockConfig struct {
	ProjectKey      string
	InputDirectory  string
}

func (m *MockConfig) GetProjectKey() string {
	return m.ProjectKey
}

func (m *MockConfig) GetInputDirectory() string {
	return m.InputDirectory
}
