package entities

import (
	"testing"
	"time"
)

func TestNewProcessResult(t *testing.T) {
	rowNumber := 5
	before := time.Now()

	result := NewProcessResult(rowNumber)
	after := time.Now()

	if result.RowNumber != rowNumber {
		t.Errorf("RowNumber = %v, want %v", result.RowNumber, rowNumber)
	}

	if result.Timestamp.Before(before) || result.Timestamp.After(after) {
		t.Errorf("Timestamp should be between %v and %v, got %v", before, after, result.Timestamp)
	}

	if result.Subtareas == nil {
		t.Error("Subtareas should be initialized, got nil")
	}

	if len(result.Subtareas) != 0 {
		t.Errorf("Subtareas length = %v, want 0", len(result.Subtareas))
	}

	if result.Success {
		t.Error("Success should be false by default")
	}
}

func TestProcessResult_AddSubtaskResult(t *testing.T) {
	result := NewProcessResult(1)

	tests := []struct {
		name        string
		description string
		success     bool
		issueKey    string
		issueURL    string
		errorMsg    string
		wantStatus  ProcessStatus
	}{
		{
			name:        "successful subtask",
			description: "Test subtask",
			success:     true,
			issueKey:    "PROJ-123",
			issueURL:    "https://jira.example.com/browse/PROJ-123",
			errorMsg:    "",
			wantStatus:  StatusSuccess,
		},
		{
			name:        "failed subtask",
			description: "Failed subtask",
			success:     false,
			issueKey:    "",
			issueURL:    "",
			errorMsg:    "Connection error",
			wantStatus:  StatusError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initialCount := len(result.Subtareas)

			result.AddSubtaskResult(tt.description, tt.success, tt.issueKey, tt.issueURL, tt.errorMsg)

			if len(result.Subtareas) != initialCount+1 {
				t.Errorf("Subtareas count = %v, want %v", len(result.Subtareas), initialCount+1)
			}

			subtask := result.Subtareas[len(result.Subtareas)-1]

			if subtask.Description != tt.description {
				t.Errorf("Description = %v, want %v", subtask.Description, tt.description)
			}
			if subtask.Success != tt.success {
				t.Errorf("Success = %v, want %v", subtask.Success, tt.success)
			}
			if subtask.IssueKey != tt.issueKey {
				t.Errorf("IssueKey = %v, want %v", subtask.IssueKey, tt.issueKey)
			}
			if subtask.IssueURL != tt.issueURL {
				t.Errorf("IssueURL = %v, want %v", subtask.IssueURL, tt.issueURL)
			}
			if subtask.Error != tt.errorMsg {
				t.Errorf("Error = %v, want %v", subtask.Error, tt.errorMsg)
			}
			if subtask.Status != tt.wantStatus {
				t.Errorf("Status = %v, want %v", subtask.Status, tt.wantStatus)
			}
		})
	}
}

func TestProcessResult_GetSuccessfulSubtasks(t *testing.T) {
	result := NewProcessResult(1)

	// Add mixed successful and failed subtasks
	result.AddSubtaskResult("Success 1", true, "PROJ-1", "url1", "")
	result.AddSubtaskResult("Failed 1", false, "", "", "error1")
	result.AddSubtaskResult("Success 2", true, "PROJ-2", "url2", "")
	result.AddSubtaskResult("Failed 2", false, "", "", "error2")

	successful := result.GetSuccessfulSubtasks()

	if len(successful) != 2 {
		t.Errorf("GetSuccessfulSubtasks() count = %v, want 2", len(successful))
	}

	for i, subtask := range successful {
		if !subtask.Success {
			t.Errorf("successful[%d].Success = false, want true", i)
		}
		if subtask.Status != StatusSuccess {
			t.Errorf("successful[%d].Status = %v, want %v", i, subtask.Status, StatusSuccess)
		}
	}
}

func TestProcessResult_GetFailedSubtasks(t *testing.T) {
	result := NewProcessResult(1)

	// Add mixed successful and failed subtasks
	result.AddSubtaskResult("Success 1", true, "PROJ-1", "url1", "")
	result.AddSubtaskResult("Failed 1", false, "", "", "error1")
	result.AddSubtaskResult("Success 2", true, "PROJ-2", "url2", "")
	result.AddSubtaskResult("Failed 2", false, "", "", "error2")

	failed := result.GetFailedSubtasks()

	if len(failed) != 2 {
		t.Errorf("GetFailedSubtasks() count = %v, want 2", len(failed))
	}

	for i, subtask := range failed {
		if subtask.Success {
			t.Errorf("failed[%d].Success = true, want false", i)
		}
		if subtask.Status != StatusError {
			t.Errorf("failed[%d].Status = %v, want %v", i, subtask.Status, StatusError)
		}
	}
}

func TestProcessResult_AllSubtasksFailed(t *testing.T) {
	tests := []struct {
		name     string
		subtasks []struct {
			success bool
		}
		want bool
	}{
		{
			name:     "no subtasks",
			subtasks: []struct{ success bool }{},
			want:     false,
		},
		{
			name: "all failed",
			subtasks: []struct{ success bool }{
				{false}, {false}, {false},
			},
			want: true,
		},
		{
			name: "mixed success and failure",
			subtasks: []struct{ success bool }{
				{false}, {true}, {false},
			},
			want: false,
		},
		{
			name: "all successful",
			subtasks: []struct{ success bool }{
				{true}, {true}, {true},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewProcessResult(1)

			for i, subtask := range tt.subtasks {
				result.AddSubtaskResult(
					"task",
					subtask.success,
					"", "", "",
				)
				_ = i // avoid unused variable
			}

			got := result.AllSubtasksFailed()
			if got != tt.want {
				t.Errorf("AllSubtasksFailed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessResult_HasAnySubtaskSuccess(t *testing.T) {
	tests := []struct {
		name     string
		subtasks []struct {
			success bool
		}
		want bool
	}{
		{
			name:     "no subtasks",
			subtasks: []struct{ success bool }{},
			want:     false,
		},
		{
			name: "all failed",
			subtasks: []struct{ success bool }{
				{false}, {false}, {false},
			},
			want: false,
		},
		{
			name: "mixed success and failure",
			subtasks: []struct{ success bool }{
				{false}, {true}, {false},
			},
			want: true,
		},
		{
			name: "all successful",
			subtasks: []struct{ success bool }{
				{true}, {true}, {true},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewProcessResult(1)

			for i, subtask := range tt.subtasks {
				result.AddSubtaskResult(
					"task",
					subtask.success,
					"", "", "",
				)
				_ = i // avoid unused variable
			}

			got := result.HasAnySubtaskSuccess()
			if got != tt.want {
				t.Errorf("HasAnySubtaskSuccess() = %v, want %v", got, tt.want)
			}
		})
	}
}
