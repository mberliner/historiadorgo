package usecases

import (
	"context"
	"errors"
	"testing"

	"historiadorgo/tests/mocks"
)

func TestTestConnectionUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name                string
		mockConnectionError error
		wantError           bool
	}{
		{
			name:                "successful connection",
			mockConnectionError: nil,
			wantError:           false,
		},
		{
			name:                "connection failed",
			mockConnectionError: errors.New("connection timeout"),
			wantError:           true,
		},
		{
			name:                "authentication failed",
			mockConnectionError: errors.New("401 Unauthorized"),
			wantError:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mockJiraRepo := &mocks.MockJiraRepository{
				TestConnectionFunc: func(ctx context.Context) error {
					return tt.mockConnectionError
				},
			}

			// Create use case
			useCase := NewTestConnectionUseCase(mockJiraRepo)

			// Execute
			err := useCase.Execute(ctx)

			// Verify
			if tt.wantError && err == nil {
				t.Errorf("Execute() error = nil, wantError = true")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Execute() error = %v, wantError = false", err)
			}
		})
	}
}
