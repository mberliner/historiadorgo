package usecases

import (
	"context"
	"errors"
	"testing"

	"historiadorgo/tests/mocks"
)

func TestDiagnoseFeaturesUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name               string
		projectKey         string
		mockRequiredFields []string
		mockError          error
		wantError          bool
		wantFields         []string
	}{
		{
			name:               "successful diagnosis with required fields",
			projectKey:         "PROJ",
			mockRequiredFields: []string{"customfield_10001", "reporter"},
			mockError:          nil,
			wantError:          false,
			wantFields:         []string{"customfield_10001", "reporter"},
		},
		{
			name:               "successful diagnosis with no required fields",
			projectKey:         "PROJ",
			mockRequiredFields: []string{},
			mockError:          nil,
			wantError:          false,
			wantFields:         []string{},
		},
		{
			name:               "project not found",
			projectKey:         "INVALID",
			mockRequiredFields: nil,
			mockError:          errors.New("project not found"),
			wantError:          true,
			wantFields:         nil,
		},
		{
			name:               "jira connection error",
			projectKey:         "PROJ",
			mockRequiredFields: nil,
			mockError:          errors.New("connection failed"),
			wantError:          true,
			wantFields:         nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mockFeatureRepo := &mocks.MockFeatureManager{
				ValidateFeatureRequiredFieldsFunc: func(ctx context.Context, projectKey string) ([]string, error) {
					if tt.mockError != nil {
						return nil, tt.mockError
					}
					return tt.mockRequiredFields, nil
				},
			}

			// Create use case
			useCase := NewDiagnoseFeaturesUseCase(mockFeatureRepo)

			// Execute
			fields, err := useCase.Execute(ctx, tt.projectKey)

			// Verify error
			if tt.wantError && err == nil {
				t.Errorf("Execute() error = nil, wantError = true")
				return
			}
			if !tt.wantError && err != nil {
				t.Errorf("Execute() error = %v, wantError = false", err)
				return
			}

			// Verify fields
			if !tt.wantError {
				if len(fields) != len(tt.wantFields) {
					t.Errorf("Execute() fields length = %v, want %v", len(fields), len(tt.wantFields))
					return
				}

				for i, field := range fields {
					if i < len(tt.wantFields) && field != tt.wantFields[i] {
						t.Errorf("Execute() fields[%d] = %v, want %v", i, field, tt.wantFields[i])
					}
				}
			}
		})
	}
}
