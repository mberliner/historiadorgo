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

func TestValidateFileUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name                  string
		filePath              string
		projectKey            string
		rows                  int
		mockFileValidateError error
		mockFileReadError     error
		mockFileReadStories   func() interface{}
		mockJiraProjectError  error
		mockJiraSubtaskError  error
		mockJiraFeatureError  error
		wantError             bool
		wantTotalStories      int
		wantWithSubtasks      int
		wantTotalSubtasks     int
		wantWithParent        int
		wantInvalidSubtasks   int
		wantPreviewContains   string
	}{
		{
			name:                  "successful validation with project",
			filePath:              "test.csv",
			projectKey:            "PROJ",
			rows:                  5,
			mockFileValidateError: nil,
			mockFileReadError:     nil,
			mockFileReadStories:   func() interface{} { return fixtures.GetSampleUserStories() },
			mockJiraProjectError:  nil,
			mockJiraSubtaskError:  nil,
			mockJiraFeatureError:  nil,
			wantError:             false,
			wantTotalStories:      4,
			wantWithSubtasks:      3,
			wantTotalSubtasks:     8, // 3+3+0+2
			wantWithParent:        3,
			wantInvalidSubtasks:   1, // Only very long subtask (empty ones are filtered out)
			wantPreviewContains:   "Login de usuario",
		},
		{
			name:                  "successful validation without project",
			filePath:              "test.csv",
			projectKey:            "",
			rows:                  5,
			mockFileValidateError: nil,
			mockFileReadError:     nil,
			mockFileReadStories:   func() interface{} { return fixtures.GetSingleUserStory() },
			wantError:             false,
			wantTotalStories:      1,
			wantWithSubtasks:      1,
			wantTotalSubtasks:     1,
			wantWithParent:        0,
			wantInvalidSubtasks:   0,
			wantPreviewContains:   "Historia única",
		},
		{
			name:                  "file validation error",
			filePath:              "invalid.csv",
			projectKey:            "",
			rows:                  5,
			mockFileValidateError: errors.New("file does not exist"),
			wantError:             true,
		},
		{
			name:                  "file read error",
			filePath:              "test.csv",
			projectKey:            "",
			rows:                  5,
			mockFileValidateError: nil,
			mockFileReadError:     errors.New("cannot read file"),
			wantError:             true,
		},
		{
			name:                  "jira project validation error",
			filePath:              "test.csv",
			projectKey:            "INVALID",
			rows:                  5,
			mockFileValidateError: nil,
			mockFileReadError:     nil,
			mockFileReadStories:   func() interface{} { return fixtures.GetSingleUserStory() },
			mockJiraProjectError:  errors.New("project not found"),
			wantError:             true,
			wantTotalStories:      1, // Should still return file statistics
			wantWithSubtasks:      1,
			wantTotalSubtasks:     1,
		},
		{
			name:                  "jira subtask type validation error",
			filePath:              "test.csv",
			projectKey:            "PROJ",
			rows:                  5,
			mockFileValidateError: nil,
			mockFileReadError:     nil,
			mockFileReadStories:   func() interface{} { return fixtures.GetSingleUserStory() },
			mockJiraProjectError:  nil,
			mockJiraSubtaskError:  errors.New("subtask type not found"),
			wantError:             true,
			wantTotalStories:      1,
			wantWithSubtasks:      1,
			wantTotalSubtasks:     1,
		},
		{
			name:                  "jira feature type validation error",
			filePath:              "test.csv",
			projectKey:            "PROJ",
			rows:                  5,
			mockFileValidateError: nil,
			mockFileReadError:     nil,
			mockFileReadStories:   func() interface{} { return fixtures.GetSingleUserStory() },
			mockJiraProjectError:  nil,
			mockJiraSubtaskError:  nil,
			mockJiraFeatureError:  errors.New("feature type not found"),
			wantError:             true,
			wantTotalStories:      1,
			wantWithSubtasks:      1,
			wantTotalSubtasks:     1,
		},
		{
			name:                  "empty file",
			filePath:              "empty.csv",
			projectKey:            "",
			rows:                  5,
			mockFileValidateError: nil,
			mockFileReadError:     nil,
			mockFileReadStories:   func() interface{} { return fixtures.GetEmptyUserStories() },
			wantError:             false,
			wantTotalStories:      0,
			wantWithSubtasks:      0,
			wantTotalSubtasks:     0,
			wantWithParent:        0,
			wantInvalidSubtasks:   0,
			wantPreviewContains:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockFileRepo := &mocks.MockFileRepository{
				ValidateFileFunc: func(ctx context.Context, filePath string) error {
					return tt.mockFileValidateError
				},
				ReadFileFunc: func(ctx context.Context, filePath string) ([]*entities.UserStory, error) {
					if tt.mockFileReadError != nil {
						return nil, tt.mockFileReadError
					}
					if tt.mockFileReadStories != nil {
						return tt.mockFileReadStories().([]*entities.UserStory), nil
					}
					return nil, nil
				},
			}

			mockJiraRepo := &mocks.MockJiraRepository{
				ValidateProjectFunc: func(ctx context.Context, projectKey string) error {
					return tt.mockJiraProjectError
				},
				ValidateSubtaskIssueTypeFunc: func(ctx context.Context, projectKey string) error {
					return tt.mockJiraSubtaskError
				},
				ValidateFeatureIssueTypeFunc: func(ctx context.Context) error {
					return tt.mockJiraFeatureError
				},
			}

			// Create use case
			useCase := NewValidateFileUseCase(mockFileRepo, mockJiraRepo)

			// Execute
			result, err := useCase.Execute(ctx, tt.filePath, tt.projectKey, tt.rows)

			// Verify error expectation
			if tt.wantError && err == nil {
				t.Errorf("Execute() error = nil, wantError = true")
				return
			}
			if !tt.wantError && err != nil {
				t.Errorf("Execute() error = %v, wantError = false", err)
				return
			}

			// If we expect an error, result might still be returned for some cases
			if result != nil {
				if result.TotalStories != tt.wantTotalStories {
					t.Errorf("TotalStories = %v, want %v", result.TotalStories, tt.wantTotalStories)
				}
				if result.WithSubtasks != tt.wantWithSubtasks {
					t.Errorf("WithSubtasks = %v, want %v", result.WithSubtasks, tt.wantWithSubtasks)
				}
				if result.TotalSubtasks != tt.wantTotalSubtasks {
					t.Errorf("TotalSubtasks = %v, want %v", result.TotalSubtasks, tt.wantTotalSubtasks)
				}
				if result.WithParent != tt.wantWithParent {
					t.Errorf("WithParent = %v, want %v", result.WithParent, tt.wantWithParent)
				}
				if result.InvalidSubtasks != tt.wantInvalidSubtasks {
					t.Errorf("InvalidSubtasks = %v, want %v", result.InvalidSubtasks, tt.wantInvalidSubtasks)
				}
				if tt.wantPreviewContains != "" && !strings.Contains(result.Preview, tt.wantPreviewContains) {
					t.Errorf("Preview should contain %q, got %q", tt.wantPreviewContains, result.Preview)
				}
			}
		})
	}
}

func TestValidateFileUseCase_Statistics(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name                string
		stories             []*entities.UserStory
		wantTotalStories    int
		wantWithSubtasks    int
		wantTotalSubtasks   int
		wantWithParent      int
		wantInvalidSubtasks int
		wantPreviewContains string
	}{
		{
			name:                "sample stories statistics",
			stories:             fixtures.GetSampleUserStories(),
			wantTotalStories:    4,
			wantWithSubtasks:    3,
			wantTotalSubtasks:   8, // 3+3+0+2
			wantWithParent:      3,
			wantInvalidSubtasks: 1, // Only very long subtask (empty ones are filtered out)
			wantPreviewContains: "Login de usuario",
		},
		{
			name:                "single story statistics",
			stories:             fixtures.GetSingleUserStory(),
			wantTotalStories:    1,
			wantWithSubtasks:    1,
			wantTotalSubtasks:   1,
			wantWithParent:      0,
			wantInvalidSubtasks: 0,
			wantPreviewContains: "Historia única",
		},
		{
			name:                "empty stories",
			stories:             fixtures.GetEmptyUserStories(),
			wantTotalStories:    0,
			wantWithSubtasks:    0,
			wantTotalSubtasks:   0,
			wantWithParent:      0,
			wantInvalidSubtasks: 0,
			wantPreviewContains: "",
		},
		{
			name:                "mixed parents stories",
			stories:             fixtures.GetSampleUserStoriesWithMixedParents(),
			wantTotalStories:    3,
			wantWithSubtasks:    2,
			wantTotalSubtasks:   4, // 2+2+0
			wantWithParent:      2,
			wantInvalidSubtasks: 0,
			wantPreviewContains: "Historia con key existente",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks that return our test stories
			mockFileRepo := &mocks.MockFileRepository{
				ValidateFileFunc: func(ctx context.Context, filePath string) error {
					return nil
				},
				ReadFileFunc: func(ctx context.Context, filePath string) ([]*entities.UserStory, error) {
					return tt.stories, nil
				},
			}

			mockJiraRepo := &mocks.MockJiraRepository{}

			useCase := NewValidateFileUseCase(mockFileRepo, mockJiraRepo)

			// Execute validation to get statistics
			result, err := useCase.Execute(ctx, "test.csv", "", 5)

			if err != nil {
				t.Errorf("Execute() error = %v, want nil", err)
				return
			}

			if result.TotalStories != tt.wantTotalStories {
				t.Errorf("TotalStories = %v, want %v", result.TotalStories, tt.wantTotalStories)
			}
			if result.WithSubtasks != tt.wantWithSubtasks {
				t.Errorf("WithSubtasks = %v, want %v", result.WithSubtasks, tt.wantWithSubtasks)
			}
			if result.TotalSubtasks != tt.wantTotalSubtasks {
				t.Errorf("TotalSubtasks = %v, want %v", result.TotalSubtasks, tt.wantTotalSubtasks)
			}
			if result.WithParent != tt.wantWithParent {
				t.Errorf("WithParent = %v, want %v", result.WithParent, tt.wantWithParent)
			}
			if result.InvalidSubtasks != tt.wantInvalidSubtasks {
				t.Errorf("InvalidSubtasks = %v, want %v", result.InvalidSubtasks, tt.wantInvalidSubtasks)
			}

			if tt.wantPreviewContains != "" {
				if !strings.Contains(result.Preview, tt.wantPreviewContains) {
					t.Errorf("Preview should contain %q, got: %q", tt.wantPreviewContains, result.Preview)
				}

				// Verify preview has proper headers
				if !strings.Contains(result.Preview, "TITULO") || !strings.Contains(result.Preview, "DESCRIPCION") {
					t.Errorf("Preview should contain headers, got: %q", result.Preview)
				}
			}
		})
	}
}

func TestValidateFileUseCase_GeneratePreview_EdgeCases(t *testing.T) {
	mockFileRepo := &mocks.MockFileRepository{}
	mockJiraRepo := &mocks.MockJiraRepository{}
	uc := NewValidateFileUseCase(mockFileRepo, mockJiraRepo)

	tests := []struct {
		name     string
		stories  []*entities.UserStory
		maxRows  int
		expected []string // strings that should be in the preview
	}{
		{
			name: "long_titulo_truncation",
			stories: []*entities.UserStory{
				{
					Titulo:      "Este es un título muy largo que debería ser truncado porque excede el límite",
					Descripcion: "Descripción",
					Subtareas:   []string{"Subtarea1", "Subtarea2"},
					Parent:      "PROJ-123",
				},
			},
			maxRows:  5,
			expected: []string{"Este es un título muy la...", "Descripción"},
		},
		{
			name: "long_descripcion_truncation",
			stories: []*entities.UserStory{
				{
					Titulo:      "Título corto",
					Descripcion: "Esta es una descripción muy larga que definitivamente debería ser truncada porque excede significativamente el límite de caracteres",
					Subtareas:   []string{"Subtarea1"},
					Parent:      "PROJ-123",
				},
			},
			maxRows:  5,
			expected: []string{"Título corto", "Esta es una descripción muy larga que defini..."},
		},
		{
			name: "many_subtasks_truncation",
			stories: []*entities.UserStory{
				{
					Titulo:      "Título",
					Descripcion: "Descripción",
					Subtareas:   []string{"Sub1", "Sub2", "Sub3", "Sub4", "Sub5"},
					Parent:      "PROJ-123",
				},
			},
			maxRows:  5,
			expected: []string{"5 subtareas"},
		},
		{
			name: "long_parent_truncation",
			stories: []*entities.UserStory{
				{
					Titulo:      "Título",
					Descripcion: "Descripción", 
					Subtareas:   []string{"Subtarea1"},
					Parent:      "PROYECTO-CON-NOMBRE-MUY-LARGO-123",
				},
			},
			maxRows:  5,
			expected: []string{"PROYECTO-C..."},
		},
		{
			name: "max_rows_limitation",
			stories: []*entities.UserStory{
				{Titulo: "Historia1", Descripcion: "Desc1", Subtareas: []string{"Sub1"}, Parent: "PROJ-1"},
				{Titulo: "Historia2", Descripcion: "Desc2", Subtareas: []string{"Sub2"}, Parent: "PROJ-2"},
				{Titulo: "Historia3", Descripcion: "Desc3", Subtareas: []string{"Sub3"}, Parent: "PROJ-3"},
			},
			maxRows:  2,
			expected: []string{"Historia1", "Historia2"}, // Historia3 shouldn't appear
		},
		{
			name: "empty_stories",
			stories: []*entities.UserStory{},
			maxRows: 5,
			expected: []string{"TITULO", "DESCRIPCION", "SUBTAREAS", "PARENT"}, // Just headers
		},
		{
			name: "zero_max_rows",
			stories: []*entities.UserStory{
				{Titulo: "Historia1", Descripcion: "Desc1", Subtareas: []string{"Sub1"}, Parent: "PROJ-1"},
			},
			maxRows:  0,
			expected: []string{"TITULO", "DESCRIPCION"}, // Just headers, no data rows
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			preview := uc.generatePreview(tt.stories, tt.maxRows)
			
			for _, expectedString := range tt.expected {
				if !strings.Contains(preview, expectedString) {
					t.Errorf("Expected preview to contain %q, but it didn't. Preview: %s", expectedString, preview)
				}
			}
			
			// For max_rows_limitation test, ensure Historia3 doesn't appear
			if tt.name == "max_rows_limitation" {
				if strings.Contains(preview, "Historia3") {
					t.Errorf("Preview should not contain Historia3 due to maxRows limit")
				}
			}
			
			// For zero_max_rows test, ensure no data rows appear
			if tt.name == "zero_max_rows" {
				if strings.Contains(preview, "Historia1") {
					t.Errorf("Preview should not contain Historia1 due to zero maxRows")
				}
			}
		})
	}
}
