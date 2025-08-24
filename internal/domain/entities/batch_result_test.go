package entities

import (
	"testing"
	"time"
)

func TestNewBatchResult(t *testing.T) {
	fileName := "test.csv"
	totalRows := 10
	dryRun := true
	before := time.Now()

	result := NewBatchResult(fileName, totalRows, dryRun)
	after := time.Now()

	if result.FileName != fileName {
		t.Errorf("FileName = %v, want %v", result.FileName, fileName)
	}
	if result.TotalRows != totalRows {
		t.Errorf("TotalRows = %v, want %v", result.TotalRows, totalRows)
	}
	if result.DryRun != dryRun {
		t.Errorf("DryRun = %v, want %v", result.DryRun, dryRun)
	}

	if result.StartTime.Before(before) || result.StartTime.After(after) {
		t.Errorf("StartTime should be between %v and %v, got %v", before, after, result.StartTime)
	}

	if result.Results == nil {
		t.Error("Results should be initialized, got nil")
	}
	if result.Errors == nil {
		t.Error("Errors should be initialized, got nil")
	}
	if result.ValidationErrors == nil {
		t.Error("ValidationErrors should be initialized, got nil")
	}

	if result.ProcessedRows != 0 {
		t.Errorf("ProcessedRows = %v, want 0", result.ProcessedRows)
	}
	if result.SuccessfulRows != 0 {
		t.Errorf("SuccessfulRows = %v, want 0", result.SuccessfulRows)
	}
	if result.ErrorRows != 0 {
		t.Errorf("ErrorRows = %v, want 0", result.ErrorRows)
	}
}

func TestBatchResult_AddResult(t *testing.T) {
	batchResult := NewBatchResult("test.csv", 10, false)

	tests := []struct {
		name          string
		resultSuccess bool
		wantProcessed int
		wantSuccess   int
		wantErrors    int
	}{
		{
			name:          "successful result",
			resultSuccess: true,
			wantProcessed: 1,
			wantSuccess:   1,
			wantErrors:    0,
		},
		{
			name:          "failed result",
			resultSuccess: false,
			wantProcessed: 2,
			wantSuccess:   1,
			wantErrors:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processResult := NewProcessResult(1)
			processResult.Success = tt.resultSuccess

			batchResult.AddResult(processResult)

			if batchResult.ProcessedRows != tt.wantProcessed {
				t.Errorf("ProcessedRows = %v, want %v", batchResult.ProcessedRows, tt.wantProcessed)
			}
			if batchResult.SuccessfulRows != tt.wantSuccess {
				t.Errorf("SuccessfulRows = %v, want %v", batchResult.SuccessfulRows, tt.wantSuccess)
			}
			if batchResult.ErrorRows != tt.wantErrors {
				t.Errorf("ErrorRows = %v, want %v", batchResult.ErrorRows, tt.wantErrors)
			}
		})
	}
}

func TestBatchResult_AddError(t *testing.T) {
	batchResult := NewBatchResult("test.csv", 10, false)

	errors := []string{"Error 1", "Error 2", "Error 3"}

	for i, err := range errors {
		batchResult.AddError(err)

		if len(batchResult.Errors) != i+1 {
			t.Errorf("Errors count = %v, want %v", len(batchResult.Errors), i+1)
		}
		if batchResult.Errors[i] != err {
			t.Errorf("Errors[%d] = %v, want %v", i, batchResult.Errors[i], err)
		}
	}
}

func TestBatchResult_AddValidationError(t *testing.T) {
	batchResult := NewBatchResult("test.csv", 10, false)

	validationErrors := []string{"Validation Error 1", "Validation Error 2"}

	for i, err := range validationErrors {
		batchResult.AddValidationError(err)

		if len(batchResult.ValidationErrors) != i+1 {
			t.Errorf("ValidationErrors count = %v, want %v", len(batchResult.ValidationErrors), i+1)
		}
		if batchResult.ValidationErrors[i] != err {
			t.Errorf("ValidationErrors[%d] = %v, want %v", i, batchResult.ValidationErrors[i], err)
		}
	}
}

func TestBatchResult_Finish(t *testing.T) {
	batchResult := NewBatchResult("test.csv", 10, false)
	startTime := batchResult.StartTime

	// Simulate some processing time
	time.Sleep(10 * time.Millisecond)

	before := time.Now()
	batchResult.Finish()
	after := time.Now()

	if batchResult.EndTime.Before(before) || batchResult.EndTime.After(after) {
		t.Errorf("EndTime should be between %v and %v, got %v", before, after, batchResult.EndTime)
	}

	if batchResult.Duration <= 0 {
		t.Errorf("Duration should be positive, got %v", batchResult.Duration)
	}

	expectedDuration := batchResult.EndTime.Sub(startTime)
	if batchResult.Duration != expectedDuration {
		t.Errorf("Duration = %v, want %v", batchResult.Duration, expectedDuration)
	}
}

func TestBatchResult_HasErrors(t *testing.T) {
	tests := []struct {
		name           string
		addGeneralErr  bool
		addErrorResult bool
		want           bool
	}{
		{
			name:           "no errors",
			addGeneralErr:  false,
			addErrorResult: false,
			want:           false,
		},
		{
			name:           "has general error",
			addGeneralErr:  true,
			addErrorResult: false,
			want:           true,
		},
		{
			name:           "has error result",
			addGeneralErr:  false,
			addErrorResult: true,
			want:           true,
		},
		{
			name:           "has both errors",
			addGeneralErr:  true,
			addErrorResult: true,
			want:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			batchResult := NewBatchResult("test.csv", 10, false)

			if tt.addGeneralErr {
				batchResult.AddError("General error")
			}

			if tt.addErrorResult {
				errorResult := NewProcessResult(1)
				errorResult.Success = false
				batchResult.AddResult(errorResult)
			}

			got := batchResult.HasErrors()
			if got != tt.want {
				t.Errorf("HasErrors() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBatchResult_HasValidationErrors(t *testing.T) {
	batchResult := NewBatchResult("test.csv", 10, false)

	if batchResult.HasValidationErrors() {
		t.Error("HasValidationErrors() should be false initially")
	}

	batchResult.AddValidationError("Validation error")

	if !batchResult.HasValidationErrors() {
		t.Error("HasValidationErrors() should be true after adding validation error")
	}
}

func TestBatchResult_IsSuccessful(t *testing.T) {
	tests := []struct {
		name                string
		hasGeneralErrors    bool
		hasValidationErrors bool
		hasSuccessfulRows   bool
		want                bool
	}{
		{
			name:                "successful batch",
			hasGeneralErrors:    false,
			hasValidationErrors: false,
			hasSuccessfulRows:   true,
			want:                true,
		},
		{
			name:                "has general errors",
			hasGeneralErrors:    true,
			hasValidationErrors: false,
			hasSuccessfulRows:   true,
			want:                false,
		},
		{
			name:                "has validation errors",
			hasGeneralErrors:    false,
			hasValidationErrors: true,
			hasSuccessfulRows:   true,
			want:                false,
		},
		{
			name:                "no successful rows",
			hasGeneralErrors:    false,
			hasValidationErrors: false,
			hasSuccessfulRows:   false,
			want:                false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			batchResult := NewBatchResult("test.csv", 10, false)

			if tt.hasGeneralErrors {
				batchResult.AddError("General error")
			}

			if tt.hasValidationErrors {
				batchResult.AddValidationError("Validation error")
			}

			if tt.hasSuccessfulRows {
				successResult := NewProcessResult(1)
				successResult.Success = true
				batchResult.AddResult(successResult)
			}

			got := batchResult.IsSuccessful()
			if got != tt.want {
				t.Errorf("IsSuccessful() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBatchResult_GetSuccessRate(t *testing.T) {
	tests := []struct {
		name            string
		successfulRows  int
		totalProcessed  int
		wantSuccessRate float64
	}{
		{
			name:            "no processed rows",
			successfulRows:  0,
			totalProcessed:  0,
			wantSuccessRate: 0.0,
		},
		{
			name:            "100% success",
			successfulRows:  5,
			totalProcessed:  5,
			wantSuccessRate: 100.0,
		},
		{
			name:            "50% success",
			successfulRows:  3,
			totalProcessed:  6,
			wantSuccessRate: 50.0,
		},
		{
			name:            "0% success",
			successfulRows:  0,
			totalProcessed:  3,
			wantSuccessRate: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			batchResult := NewBatchResult("test.csv", 10, false)

			// Add successful results
			for i := 0; i < tt.successfulRows; i++ {
				successResult := NewProcessResult(i + 1)
				successResult.Success = true
				batchResult.AddResult(successResult)
			}

			// Add failed results
			for i := 0; i < tt.totalProcessed-tt.successfulRows; i++ {
				failedResult := NewProcessResult(tt.successfulRows + i + 1)
				failedResult.Success = false
				batchResult.AddResult(failedResult)
			}

			got := batchResult.GetSuccessRate()
			if got != tt.wantSuccessRate {
				t.Errorf("GetSuccessRate() = %v, want %v", got, tt.wantSuccessRate)
			}
		})
	}
}

func TestBatchResult_GetProcessedIssues(t *testing.T) {
	batchResult := NewBatchResult("test.csv", 10, false)

	// Add successful result with issue key
	successResult1 := NewProcessResult(1)
	successResult1.Success = true
	successResult1.IssueKey = "PROJ-123"
	batchResult.AddResult(successResult1)

	// Add failed result (should not be included)
	failedResult := NewProcessResult(2)
	failedResult.Success = false
	failedResult.IssueKey = "PROJ-124"
	batchResult.AddResult(failedResult)

	// Add successful result without issue key (should not be included)
	successResult2 := NewProcessResult(3)
	successResult2.Success = true
	successResult2.IssueKey = ""
	batchResult.AddResult(successResult2)

	// Add another successful result with issue key
	successResult3 := NewProcessResult(4)
	successResult3.Success = true
	successResult3.IssueKey = "PROJ-125"
	batchResult.AddResult(successResult3)

	issues := batchResult.GetProcessedIssues()

	expectedIssues := []string{"PROJ-123", "PROJ-125"}
	if len(issues) != len(expectedIssues) {
		t.Errorf("GetProcessedIssues() count = %v, want %v", len(issues), len(expectedIssues))
	}

	for i, issue := range issues {
		if i < len(expectedIssues) && issue != expectedIssues[i] {
			t.Errorf("GetProcessedIssues()[%d] = %v, want %v", i, issue, expectedIssues[i])
		}
	}
}
