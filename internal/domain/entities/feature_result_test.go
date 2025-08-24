package entities

import (
	"testing"
	"time"
)

func TestNewFeatureResult(t *testing.T) {
	description := "Test feature description"
	before := time.Now()

	result := NewFeatureResult(description)

	after := time.Now()

	if result.Description != description {
		t.Errorf("Description = %v, want %v", result.Description, description)
	}

	if result.Success {
		t.Errorf("Success = %v, want false", result.Success)
	}

	if result.WasCreated {
		t.Errorf("WasCreated = %v, want false", result.WasCreated)
	}

	if result.Timestamp.Before(before) || result.Timestamp.After(after) {
		t.Errorf("Timestamp not within expected range")
	}
}

func TestFeatureResult_SetSuccess(t *testing.T) {
	tests := []struct {
		name       string
		issueKey   string
		issueURL   string
		wasCreated bool
	}{
		{
			name:       "new feature created",
			issueKey:   "PROJ-123",
			issueURL:   "https://test.atlassian.net/browse/PROJ-123",
			wasCreated: true,
		},
		{
			name:       "existing feature found",
			issueKey:   "PROJ-456",
			issueURL:   "https://test.atlassian.net/browse/PROJ-456",
			wasCreated: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewFeatureResult("Test description")

			result.SetSuccess(tt.issueKey, tt.issueURL, tt.wasCreated)

			if !result.Success {
				t.Errorf("Success = %v, want true", result.Success)
			}

			if result.IssueKey != tt.issueKey {
				t.Errorf("IssueKey = %v, want %v", result.IssueKey, tt.issueKey)
			}

			if result.IssueURL != tt.issueURL {
				t.Errorf("IssueURL = %v, want %v", result.IssueURL, tt.issueURL)
			}

			if result.WasCreated != tt.wasCreated {
				t.Errorf("WasCreated = %v, want %v", result.WasCreated, tt.wasCreated)
			}
		})
	}
}

func TestFeatureResult_SetExisting(t *testing.T) {
	result := NewFeatureResult("Test description")
	existingKey := "PROJ-789"

	result.SetExisting(existingKey)

	if !result.Success {
		t.Errorf("Success = %v, want true", result.Success)
	}

	if result.ExistingKey != existingKey {
		t.Errorf("ExistingKey = %v, want %v", result.ExistingKey, existingKey)
	}

	if result.IssueKey != existingKey {
		t.Errorf("IssueKey = %v, want %v", result.IssueKey, existingKey)
	}

	if result.WasCreated {
		t.Errorf("WasCreated = %v, want false", result.WasCreated)
	}
}

func TestFeatureResult_SetError(t *testing.T) {
	result := NewFeatureResult("Test description")
	errorMessage := "Failed to create feature"

	result.SetError(errorMessage)

	if result.Success {
		t.Errorf("Success = %v, want false", result.Success)
	}

	if result.ErrorMessage != errorMessage {
		t.Errorf("ErrorMessage = %v, want %v", result.ErrorMessage, errorMessage)
	}
}

func TestFeatureResult_SetNormalizedDescription(t *testing.T) {
	result := NewFeatureResult("Original Description")
	normalized := "normalized description"

	result.SetNormalizedDescription(normalized)

	if result.NormalizedDesc != normalized {
		t.Errorf("NormalizedDesc = %v, want %v", result.NormalizedDesc, normalized)
	}
}

func TestFeatureResult_StateTransitions(t *testing.T) {
	// Test that we can transition between states
	result := NewFeatureResult("Test description")

	// Start as failure
	result.SetError("Initial error")
	if result.Success {
		t.Errorf("Expected failure state")
	}

	// Transition to success
	result.SetSuccess("PROJ-123", "url", true)
	if !result.Success {
		t.Errorf("Expected success state")
	}

	// Transition to existing
	result.SetExisting("PROJ-456")
	if !result.Success || result.WasCreated {
		t.Errorf("Expected existing state")
	}
}
