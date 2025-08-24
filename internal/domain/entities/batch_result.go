package entities

import "time"

type BatchResult struct {
	FileName         string           `json:"file_name"`
	TotalRows        int              `json:"total_rows"`
	ProcessedRows    int              `json:"processed_rows"`
	SuccessfulRows   int              `json:"successful_rows"`
	ErrorRows        int              `json:"error_rows"`
	SkippedRows      int              `json:"skipped_rows"`
	StartTime        time.Time        `json:"start_time"`
	EndTime          time.Time        `json:"end_time"`
	Duration         time.Duration    `json:"duration"`
	Results          []*ProcessResult `json:"results"`
	Errors           []string         `json:"errors"`
	ValidationErrors []string         `json:"validation_errors"`
	DryRun           bool             `json:"dry_run"`
}

func NewBatchResult(fileName string, totalRows int, dryRun bool) *BatchResult {
	return &BatchResult{
		FileName:         fileName,
		TotalRows:        totalRows,
		StartTime:        time.Now(),
		Results:          make([]*ProcessResult, 0),
		Errors:           make([]string, 0),
		ValidationErrors: make([]string, 0),
		DryRun:           dryRun,
	}
}

func (br *BatchResult) AddResult(result *ProcessResult) {
	br.Results = append(br.Results, result)
	br.ProcessedRows++

	if result.Success {
		br.SuccessfulRows++
	} else {
		br.ErrorRows++
	}
}

func (br *BatchResult) AddError(error string) {
	br.Errors = append(br.Errors, error)
}

func (br *BatchResult) AddValidationError(error string) {
	br.ValidationErrors = append(br.ValidationErrors, error)
}

func (br *BatchResult) Finish() {
	br.EndTime = time.Now()
	br.Duration = br.EndTime.Sub(br.StartTime)
}

func (br *BatchResult) HasErrors() bool {
	return len(br.Errors) > 0 || br.ErrorRows > 0
}

func (br *BatchResult) HasValidationErrors() bool {
	return len(br.ValidationErrors) > 0
}

func (br *BatchResult) IsSuccessful() bool {
	return !br.HasErrors() && !br.HasValidationErrors() && br.SuccessfulRows > 0
}

func (br *BatchResult) GetSuccessRate() float64 {
	if br.ProcessedRows == 0 {
		return 0.0
	}
	return float64(br.SuccessfulRows) / float64(br.ProcessedRows) * 100
}

func (br *BatchResult) GetProcessedIssues() []string {
	var issues []string
	for _, result := range br.Results {
		if result.Success && result.IssueKey != "" {
			issues = append(issues, result.IssueKey)
		}
	}
	return issues
}
