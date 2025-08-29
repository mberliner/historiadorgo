package usecases

import (
	"context"
	"errors"
	"strings"
	"testing"

	"historiadorgo/internal/domain/entities"
	"historiadorgo/tests/fixtures"
	"historiadorgo/tests/mocks"
)

func TestProcessFilesUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name              string
		filePath          string
		projectKey        string
		dryRun            bool
		userStories       []*entities.UserStory
		fileError         error
		jiraError         error
		createResponse    *entities.ProcessResult
		wantError         bool
		wantRowsProcessed int
	}{
		{
			name:       "successful processing",
			filePath:   "test.csv",
			projectKey: "PROJ",
			dryRun:     false,
			userStories: []*entities.UserStory{
				fixtures.ValidUserStory1(),
				fixtures.ValidUserStory2(),
			},
			fileError:         nil,
			jiraError:         nil,
			createResponse:    fixtures.SuccessProcessResult(),
			wantError:         false,
			wantRowsProcessed: 2,
		},
		{
			name:       "dry run mode",
			filePath:   "test.csv",
			projectKey: "PROJ",
			dryRun:     true,
			userStories: []*entities.UserStory{
				fixtures.ValidUserStory1(),
			},
			fileError:         nil,
			jiraError:         nil,
			createResponse:    nil, // Not used in dry run
			wantError:         false,
			wantRowsProcessed: 1,
		},
		{
			name:              "file read error",
			filePath:          "nonexistent.csv",
			projectKey:        "PROJ",
			dryRun:            false,
			userStories:       nil,
			fileError:         errors.New("file not found"),
			jiraError:         nil,
			createResponse:    nil,
			wantError:         true,
			wantRowsProcessed: 0,
		},
		{
			name:              "empty file",
			filePath:          "empty.csv",
			projectKey:        "PROJ",
			dryRun:            false,
			userStories:       []*entities.UserStory{},
			fileError:         nil,
			jiraError:         nil,
			createResponse:    nil,
			wantError:         false,
			wantRowsProcessed: 0,
		},
		{
			name:       "jira creation error",
			filePath:   "test.csv",
			projectKey: "PROJ",
			dryRun:     false,
			userStories: []*entities.UserStory{
				fixtures.ValidUserStory1(),
			},
			fileError:         nil,
			jiraError:         errors.New("jira API error"),
			createResponse:    fixtures.ErrorProcessResult(),
			wantError:         false, // Use case doesn't fail, but result contains errors
			wantRowsProcessed: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockFileRepo := &mocks.MockFileRepository{
				ReadFileFunc: func(ctx context.Context, filePath string) ([]*entities.UserStory, error) {
					if tt.fileError != nil {
						return nil, tt.fileError
					}
					return tt.userStories, nil
				},
			}

			mockJiraRepo := &mocks.MockJiraRepository{
				CreateUserStoryFunc: func(ctx context.Context, story *entities.UserStory, rowNumber int) (*entities.ProcessResult, error) {
					if tt.jiraError != nil {
						result := entities.NewProcessResult(story.Row)
						result.Success = false
						result.ErrorMessage = tt.jiraError.Error()
						return result, nil
					}
					return tt.createResponse, nil
				},
			}

			mockFeatureRepo := &mocks.MockFeatureManager{}

			// Create use case
			useCase := NewProcessFilesUseCase(mockFileRepo, mockJiraRepo, mockFeatureRepo)

			// Execute
			result, err := useCase.Execute(ctx, tt.filePath, tt.projectKey, tt.dryRun)

			// Verify error
			if tt.wantError && err == nil {
				t.Errorf("Execute() error = nil, wantError = true")
				return
			}
			if !tt.wantError && err != nil {
				t.Errorf("Execute() error = %v, wantError = false", err)
				return
			}

			// Verify result
			if !tt.wantError {
				if result == nil {
					t.Errorf("Execute() result = nil, want BatchResult")
					return
				}

				if result.TotalRows != len(tt.userStories) {
					t.Errorf("Execute() result.TotalRows = %v, want %v", result.TotalRows, len(tt.userStories))
				}

				if result.ProcessedRows != tt.wantRowsProcessed {
					t.Errorf("Execute() result.ProcessedRows = %v, want %v", result.ProcessedRows, tt.wantRowsProcessed)
				}

				if result.DryRun != tt.dryRun {
					t.Errorf("Execute() result.DryRun = %v, want %v", result.DryRun, tt.dryRun)
				}

				if result.FileName != tt.filePath {
					t.Errorf("Execute() result.FileName = %v, want %v", result.FileName, tt.filePath)
				}
			}
		})
	}
}

func TestProcessFilesUseCase_Execute_BatchProcessing(t *testing.T) {
	ctx := context.Background()

	// Create multiple user stories to test batch processing
	userStories := make([]*entities.UserStory, 5)
	for i := 0; i < 5; i++ {
		story := fixtures.ValidUserStory1()
		story.Row = i + 1
		userStories[i] = story
	}

	mockFileRepo := &mocks.MockFileRepository{
		ReadFileFunc: func(ctx context.Context, filePath string) ([]*entities.UserStory, error) {
			return userStories, nil
		},
	}

	createCallCount := 0
	mockJiraRepo := &mocks.MockJiraRepository{
		CreateUserStoryFunc: func(ctx context.Context, story *entities.UserStory, rowNumber int) (*entities.ProcessResult, error) {
			createCallCount++
			result := entities.NewProcessResult(story.Row)
			result.Success = true
			result.IssueKey = "PROJ-" + string(rune(100+story.Row))
			return result, nil
		},
	}

	mockFeatureRepo := &mocks.MockFeatureManager{}

	useCase := NewProcessFilesUseCase(mockFileRepo, mockJiraRepo, mockFeatureRepo)
	result, err := useCase.Execute(ctx, "test.csv", "PROJ", false)

	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
		return
	}

	if createCallCount != 5 {
		t.Errorf("CreateUserStory called %d times, want 5", createCallCount)
	}

	if result.ProcessedRows != 5 {
		t.Errorf("ProcessedRows = %v, want 5", result.ProcessedRows)
	}

	if result.SuccessfulRows != 5 {
		t.Errorf("SuccessfulRows = %v, want 5", result.SuccessfulRows)
	}
}

func TestProcessFilesUseCase_Execute_MixedResults(t *testing.T) {
	ctx := context.Background()

	userStories := []*entities.UserStory{
		fixtures.ValidUserStory1(), // Will succeed
		fixtures.ValidUserStory2(), // Will fail
	}
	userStories[0].Row = 1
	userStories[1].Row = 2

	mockFileRepo := &mocks.MockFileRepository{
		ReadFileFunc: func(ctx context.Context, filePath string) ([]*entities.UserStory, error) {
			return userStories, nil
		},
	}

	mockJiraRepo := &mocks.MockJiraRepository{
		CreateUserStoryFunc: func(ctx context.Context, story *entities.UserStory, rowNumber int) (*entities.ProcessResult, error) {
			result := entities.NewProcessResult(story.Row)
			if story.Row == 1 {
				result.Success = true
				result.IssueKey = "PROJ-123"
			} else {
				result.Success = false
				result.ErrorMessage = "Failed to create issue"
			}
			return result, nil
		},
	}

	mockFeatureRepo := &mocks.MockFeatureManager{}

	useCase := NewProcessFilesUseCase(mockFileRepo, mockJiraRepo, mockFeatureRepo)
	result, err := useCase.Execute(ctx, "test.csv", "PROJ", false)

	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
		return
	}

	if result.SuccessfulRows != 1 {
		t.Errorf("SuccessfulRows = %v, want 1", result.SuccessfulRows)
	}

	if result.ErrorRows != 1 {
		t.Errorf("ErrorRows = %v, want 1", result.ErrorRows)
	}

	if len(result.Results) != 2 {
		t.Errorf("Results length = %v, want 2", len(result.Results))
	}
}

func TestProcessFilesUseCase_Execute_FileValidationErrors(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name            string
		filePath        string
		projectKey      string
		dryRun          bool
		validateError   error
		readFileError   error
		moveFileError   error
		wantError       bool
		wantErrorString string
	}{
		{
			name:            "dry run with file validation error",
			filePath:        "invalid.csv",
			projectKey:      "PROJ",
			dryRun:          true,
			validateError:   errors.New("file has invalid format"),
			wantError:       true,
			wantErrorString: "file validation failed",
		},
		{
			name:            "file read error after validation",
			filePath:        "valid.csv",
			projectKey:      "PROJ",
			dryRun:          false,
			readFileError:   errors.New("permission denied"),
			wantError:       true,
			wantErrorString: "error reading file",
		},
		{
			name:          "move file error after successful processing",
			filePath:      "valid.csv",
			projectKey:    "PROJ",
			dryRun:        false,
			moveFileError: errors.New("target directory not writable"),
			wantError:     false, // Should not fail completely, just add warning
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileRepo := &mocks.MockFileRepository{
				ValidateFileFunc: func(ctx context.Context, filePath string) error {
					if tt.dryRun && tt.validateError != nil {
						return tt.validateError
					}
					return nil
				},
				ReadFileFunc: func(ctx context.Context, filePath string) ([]*entities.UserStory, error) {
					if tt.readFileError != nil {
						return nil, tt.readFileError
					}
					if !tt.dryRun && tt.validateError == nil {
						return []*entities.UserStory{fixtures.ValidUserStory1()}, nil
					}
					return []*entities.UserStory{}, nil
				},
				MoveToProcessedFunc: func(ctx context.Context, filePath string) error {
					return tt.moveFileError
				},
			}

			jiraRepo := &mocks.MockJiraRepository{
				TestConnectionFunc: func(ctx context.Context) error {
					return nil
				},
				ValidateProjectFunc: func(ctx context.Context, projectKey string) error {
					return nil
				},
				ValidateSubtaskIssueTypeFunc: func(ctx context.Context, projectKey string) error {
					return nil
				},
				ValidateFeatureIssueTypeFunc: func(ctx context.Context) error {
					return nil
				},
				CreateUserStoryFunc: func(ctx context.Context, story *entities.UserStory, rowNumber int) (*entities.ProcessResult, error) {
					return fixtures.SuccessProcessResult(), nil
				},
			}

			featureRepo := &mocks.MockFeatureManager{}

			useCase := NewProcessFilesUseCase(fileRepo, jiraRepo, featureRepo)
			result, err := useCase.Execute(ctx, tt.filePath, tt.projectKey, tt.dryRun)

			if tt.wantError {
				if err == nil {
					t.Error("Execute() expected error but got none")
				} else if !strings.Contains(err.Error(), tt.wantErrorString) {
					t.Errorf("Execute() error = %v, want error containing %v", err, tt.wantErrorString)
				}
			} else {
				if err != nil {
					t.Errorf("Execute() unexpected error = %v", err)
				}
				if tt.moveFileError != nil && result != nil && !strings.Contains(strings.Join(result.Errors, " "), "could not move file") {
					t.Error("Execute() expected move file warning in result.Errors")
				}
			}
		})
	}
}

func TestProcessFilesUseCase_ProcessAllFiles_ErrorPaths(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name             string
		inputDir         string
		projectKey       string
		dryRun           bool
		files            []string
		getFilesError    error
		executeError     error
		validationError  error
		wantError        bool
		wantErrorString  string
		wantResultsCount int
	}{
		{
			name:            "validation error in non-dry-run",
			inputDir:        "/input",
			projectKey:      "PROJ",
			dryRun:          false,
			validationError: errors.New("jira connection failed"),
			wantError:       true,
			wantErrorString: "jira connection failed",
		},
		{
			name:            "get pending files error",
			inputDir:        "/nonexistent",
			projectKey:      "PROJ",
			dryRun:          true,
			getFilesError:   errors.New("directory not found"),
			wantError:       true,
			wantErrorString: "error getting pending files",
		},
		{
			name:            "no files found",
			inputDir:        "/empty",
			projectKey:      "PROJ",
			dryRun:          true,
			files:           []string{},
			wantError:       true,
			wantErrorString: "no files found",
		},
		{
			name:             "single file processing error creates error result",
			inputDir:         "/input",
			projectKey:       "PROJ",
			dryRun:           true,
			files:            []string{"file1.csv"},
			executeError:     errors.New("file processing failed"),
			wantError:        false, // Should not fail, but create error result
			wantResultsCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileRepo := &mocks.MockFileRepository{
				GetPendingFilesFunc: func(ctx context.Context, inputDir string) ([]string, error) {
					if tt.getFilesError != nil {
						return nil, tt.getFilesError
					}
					return tt.files, nil
				},
				ValidateFileFunc: func(ctx context.Context, filePath string) error {
					if tt.executeError != nil && len(tt.files) > 0 {
						return tt.executeError
					}
					return nil
				},
				ReadFileFunc: func(ctx context.Context, filePath string) ([]*entities.UserStory, error) {
					return []*entities.UserStory{fixtures.ValidUserStory1()}, nil
				},
			}

			jiraRepo := &mocks.MockJiraRepository{
				TestConnectionFunc: func(ctx context.Context) error {
					if !tt.dryRun && tt.validationError != nil {
						return tt.validationError
					}
					return nil
				},
				ValidateProjectFunc: func(ctx context.Context, projectKey string) error {
					return nil
				},
				ValidateSubtaskIssueTypeFunc: func(ctx context.Context, projectKey string) error {
					return nil
				},
				ValidateFeatureIssueTypeFunc: func(ctx context.Context) error {
					return nil
				},
			}

			featureRepo := &mocks.MockFeatureManager{}

			useCase := NewProcessFilesUseCase(fileRepo, jiraRepo, featureRepo)
			results, err := useCase.ProcessAllFiles(ctx, tt.inputDir, tt.projectKey, tt.dryRun)

			if tt.wantError {
				if err == nil {
					t.Error("ProcessAllFiles() expected error but got none")
				} else if !strings.Contains(err.Error(), tt.wantErrorString) {
					t.Errorf("ProcessAllFiles() error = %v, want error containing %v", err, tt.wantErrorString)
				}
			} else {
				if err != nil {
					t.Errorf("ProcessAllFiles() unexpected error = %v", err)
				}
				if tt.wantResultsCount > 0 {
					if len(results) != tt.wantResultsCount {
						t.Errorf("ProcessAllFiles() results count = %d, want %d", len(results), tt.wantResultsCount)
					}
					// Verify error result was created for failed file
					if tt.executeError != nil && len(results) > 0 && len(results[0].Errors) == 0 {
						t.Error("ProcessAllFiles() expected error in result but got none")
					}
				}
			}
		})
	}
}

func TestProcessFilesUseCase_ProcessAllFiles(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name             string
		inputDir         string
		projectKey       string
		dryRun           bool
		pendingFiles     []string
		filesError       error
		validationError  error
		wantError        bool
		wantResultsCount int
	}{
		{
			name:             "successful processing multiple files",
			inputDir:         "/input",
			projectKey:       "PROJ",
			dryRun:           false,
			pendingFiles:     []string{"/input/file1.csv", "/input/file2.csv"},
			filesError:       nil,
			validationError:  nil,
			wantError:        false,
			wantResultsCount: 2,
		},
		{
			name:             "dry run multiple files",
			inputDir:         "/input",
			projectKey:       "PROJ",
			dryRun:           true,
			pendingFiles:     []string{"/input/file1.csv"},
			filesError:       nil,
			validationError:  nil,
			wantError:        false,
			wantResultsCount: 1,
		},
		{
			name:             "no files found",
			inputDir:         "/empty",
			projectKey:       "PROJ",
			dryRun:           false,
			pendingFiles:     []string{},
			filesError:       nil,
			validationError:  nil,
			wantError:        true,
			wantResultsCount: 0,
		},
		{
			name:             "error getting files",
			inputDir:         "/error",
			projectKey:       "PROJ",
			dryRun:           false,
			pendingFiles:     nil,
			filesError:       errors.New("directory not found"),
			validationError:  nil,
			wantError:        true,
			wantResultsCount: 0,
		},
		{
			name:             "validation error",
			inputDir:         "/input",
			projectKey:       "INVALID",
			dryRun:           false,
			pendingFiles:     []string{"/input/file1.csv"},
			filesError:       nil,
			validationError:  errors.New("invalid project"),
			wantError:        true,
			wantResultsCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockFileRepo := &mocks.MockFileRepository{
				GetPendingFilesFunc: func(ctx context.Context, inputDir string) ([]string, error) {
					if tt.filesError != nil {
						return nil, tt.filesError
					}
					return tt.pendingFiles, nil
				},
				ReadFileFunc: func(ctx context.Context, filePath string) ([]*entities.UserStory, error) {
					return []*entities.UserStory{fixtures.ValidUserStory1()}, nil
				},
				MoveToProcessedFunc: func(ctx context.Context, filePath string) error {
					return nil
				},
			}

			mockJiraRepo := &mocks.MockJiraRepository{
				TestConnectionFunc: func(ctx context.Context) error {
					return tt.validationError
				},
				ValidateProjectFunc: func(ctx context.Context, projectKey string) error {
					return tt.validationError
				},
				ValidateSubtaskIssueTypeFunc: func(ctx context.Context, projectKey string) error {
					return tt.validationError
				},
				ValidateFeatureIssueTypeFunc: func(ctx context.Context) error {
					return tt.validationError
				},
				CreateUserStoryFunc: func(ctx context.Context, story *entities.UserStory, rowNumber int) (*entities.ProcessResult, error) {
					result := entities.NewProcessResult(rowNumber)
					result.Success = true
					result.IssueKey = "PROJ-123"
					return result, nil
				},
			}

			mockFeatureRepo := &mocks.MockFeatureManager{}

			useCase := NewProcessFilesUseCase(mockFileRepo, mockJiraRepo, mockFeatureRepo)

			// Execute
			results, err := useCase.ProcessAllFiles(ctx, tt.inputDir, tt.projectKey, tt.dryRun)

			// Verify error
			if tt.wantError && err == nil {
				t.Errorf("ProcessAllFiles() error = nil, wantError = true")
				return
			}
			if !tt.wantError && err != nil {
				t.Errorf("ProcessAllFiles() error = %v, wantError = false", err)
				return
			}

			// Verify results count
			if len(results) != tt.wantResultsCount {
				t.Errorf("ProcessAllFiles() results count = %v, want %v", len(results), tt.wantResultsCount)
			}
		})
	}
}

func TestProcessFilesUseCase_validateInputs(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name                  string
		projectKey            string
		connectionError       error
		projectError          error
		subtaskTypeError      error
		featureTypeError      error
		wantError             bool
		expectedErrorContains string
	}{
		{
			name:       "all validations pass",
			projectKey: "PROJ",
			wantError:  false,
		},
		{
			name:                  "connection fails",
			projectKey:            "PROJ",
			connectionError:       errors.New("connection timeout"),
			wantError:             true,
			expectedErrorContains: "jira connection failed",
		},
		{
			name:                  "project validation fails",
			projectKey:            "INVALID",
			projectError:          errors.New("project not found"),
			wantError:             true,
			expectedErrorContains: "project validation failed",
		},
		{
			name:                  "subtask type validation fails",
			projectKey:            "PROJ",
			subtaskTypeError:      errors.New("subtask type not found"),
			wantError:             true,
			expectedErrorContains: "subtask type validation failed",
		},
		{
			name:                  "feature type validation fails",
			projectKey:            "PROJ",
			featureTypeError:      errors.New("feature type not found"),
			wantError:             true,
			expectedErrorContains: "feature type validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockJiraRepo := &mocks.MockJiraRepository{
				TestConnectionFunc: func(ctx context.Context) error {
					return tt.connectionError
				},
				ValidateProjectFunc: func(ctx context.Context, projectKey string) error {
					return tt.projectError
				},
				ValidateSubtaskIssueTypeFunc: func(ctx context.Context, projectKey string) error {
					return tt.subtaskTypeError
				},
				ValidateFeatureIssueTypeFunc: func(ctx context.Context) error {
					return tt.featureTypeError
				},
			}

			useCase := &ProcessFilesUseCase{jiraRepo: mockJiraRepo}

			err := useCase.validateInputs(ctx, tt.projectKey)

			if tt.wantError && err == nil {
				t.Errorf("validateInputs() error = nil, wantError = true")
				return
			}
			if !tt.wantError && err != nil {
				t.Errorf("validateInputs() error = %v, wantError = false", err)
				return
			}

			if tt.wantError && !contains(err.Error(), tt.expectedErrorContains) {
				t.Errorf("validateInputs() error = %v, want error containing %v", err.Error(), tt.expectedErrorContains)
			}
		})
	}
}

func TestProcessFilesUseCase_processUserStory_DryRun(t *testing.T) {
	ctx := context.Background()

	// Test dry run with subtasks
	story := fixtures.ValidUserStory1()
	story.Subtareas = []string{"Subtarea 1", "Subtarea 2"}

	useCase := &ProcessFilesUseCase{}
	result := useCase.processUserStory(ctx, story, "PROJ", 1, true)

	if !result.Success {
		t.Errorf("processUserStory() dry run should always succeed")
	}

	expectedKey := "DRY-RUN-1"
	if result.IssueKey != expectedKey {
		t.Errorf("processUserStory() dry run IssueKey = %v, want %v", result.IssueKey, expectedKey)
	}

	expectedURL := "https://dry-run.example.com/browse/DRY-RUN-1"
	if result.IssueURL != expectedURL {
		t.Errorf("processUserStory() dry run IssueURL = %v, want %v", result.IssueURL, expectedURL)
	}

	// Should have created subtask results
	if len(result.Subtareas) != 2 {
		t.Errorf("processUserStory() dry run should create %d subtask results, got %d", 2, len(result.Subtareas))
	}

	// Check first subtask
	if len(result.Subtareas) > 0 {
		firstSubtask := result.Subtareas[0]
		if !firstSubtask.Success {
			t.Errorf("processUserStory() dry run subtask should succeed")
		}
		if firstSubtask.IssueKey != "DRY-SUB-1-1" {
			t.Errorf("processUserStory() dry run subtask key = %v, want DRY-SUB-1-1", firstSubtask.IssueKey)
		}
	}
}

func TestProcessFilesUseCase_processUserStory_Production(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		jiraError   error
		wantSuccess bool
	}{
		{
			name:        "successful jira creation",
			jiraError:   nil,
			wantSuccess: true,
		},
		{
			name:        "jira creation fails",
			jiraError:   errors.New("jira API error"),
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			story := fixtures.ValidUserStory1()

			mockJiraRepo := &mocks.MockJiraRepository{
				CreateUserStoryFunc: func(ctx context.Context, story *entities.UserStory, rowNumber int) (*entities.ProcessResult, error) {
					if tt.jiraError != nil {
						return nil, tt.jiraError
					}
					result := entities.NewProcessResult(rowNumber)
					result.Success = true
					result.IssueKey = "PROJ-123"
					return result, nil
				},
			}

			useCase := &ProcessFilesUseCase{jiraRepo: mockJiraRepo}
			result := useCase.processUserStory(ctx, story, "PROJ", 1, false)

			if result.Success != tt.wantSuccess {
				t.Errorf("processUserStory() Success = %v, want %v", result.Success, tt.wantSuccess)
			}

			if tt.jiraError != nil && result.ErrorMessage == "" {
				t.Errorf("processUserStory() should set ErrorMessage when jira fails")
			}
		})
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
