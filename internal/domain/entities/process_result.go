package entities

import "time"

type ProcessStatus string

const (
	StatusPending ProcessStatus = "pending"
	StatusSuccess ProcessStatus = "success"
	StatusError   ProcessStatus = "error"
	StatusSkipped ProcessStatus = "skipped"
)

type ProcessResult struct {
	Success         bool             `json:"success"`
	IssueKey        string           `json:"issue_key,omitempty"`
	IssueURL        string           `json:"issue_url,omitempty"`
	ErrorMessage    string           `json:"error_message,omitempty"`
	RowNumber       int              `json:"row_number,omitempty"`
	Timestamp       time.Time        `json:"timestamp"`
	Subtareas       []*SubtaskResult `json:"subtareas,omitempty"`
	FeatureKey      string           `json:"feature_key,omitempty"`
	CreatedIssueKey string           `json:"created_issue_key,omitempty"`
}

type SubtaskResult struct {
	Description string        `json:"description"`
	Success     bool          `json:"success"`
	IssueKey    string        `json:"issue_key,omitempty"`
	IssueURL    string        `json:"issue_url,omitempty"`
	Error       string        `json:"error,omitempty"`
	Status      ProcessStatus `json:"status"`
}

func NewProcessResult(rowNumber int) *ProcessResult {
	return &ProcessResult{
		RowNumber: rowNumber,
		Timestamp: time.Now(),
		Subtareas: make([]*SubtaskResult, 0),
	}
}

func (pr *ProcessResult) AddSubtaskResult(description string, success bool, issueKey, issueURL, errorMsg string) {
	status := StatusSuccess
	if !success {
		status = StatusError
	}

	subtask := &SubtaskResult{
		Description: description,
		Success:     success,
		IssueKey:    issueKey,
		IssueURL:    issueURL,
		Error:       errorMsg,
		Status:      status,
	}

	pr.Subtareas = append(pr.Subtareas, subtask)
}

func (pr *ProcessResult) GetSuccessfulSubtasks() []*SubtaskResult {
	var successful []*SubtaskResult
	for _, subtask := range pr.Subtareas {
		if subtask.Success {
			successful = append(successful, subtask)
		}
	}
	return successful
}

func (pr *ProcessResult) GetFailedSubtasks() []*SubtaskResult {
	var failed []*SubtaskResult
	for _, subtask := range pr.Subtareas {
		if !subtask.Success {
			failed = append(failed, subtask)
		}
	}
	return failed
}

func (pr *ProcessResult) AllSubtasksFailed() bool {
	if len(pr.Subtareas) == 0 {
		return false
	}
	return len(pr.GetFailedSubtasks()) == len(pr.Subtareas)
}

func (pr *ProcessResult) HasAnySubtaskSuccess() bool {
	return len(pr.GetSuccessfulSubtasks()) > 0
}
