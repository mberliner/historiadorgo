package entities

import "time"

type FeatureResult struct {
	Description    string    `json:"description"`
	Success        bool      `json:"success"`
	IssueKey       string    `json:"issue_key,omitempty"`
	IssueURL       string    `json:"issue_url,omitempty"`
	ErrorMessage   string    `json:"error_message,omitempty"`
	WasCreated     bool      `json:"was_created"`
	ExistingKey    string    `json:"existing_key,omitempty"`
	Timestamp      time.Time `json:"timestamp"`
	NormalizedDesc string    `json:"normalized_description,omitempty"`
}

func NewFeatureResult(description string) *FeatureResult {
	return &FeatureResult{
		Description: description,
		Timestamp:   time.Now(),
	}
}

func (fr *FeatureResult) SetSuccess(issueKey, issueURL string, wasCreated bool) {
	fr.Success = true
	fr.IssueKey = issueKey
	fr.IssueURL = issueURL
	fr.WasCreated = wasCreated
}

func (fr *FeatureResult) SetExisting(existingKey string) {
	fr.Success = true
	fr.ExistingKey = existingKey
	fr.IssueKey = existingKey
	fr.WasCreated = false
}

func (fr *FeatureResult) SetError(errorMessage string) {
	fr.Success = false
	fr.ErrorMessage = errorMessage
}

func (fr *FeatureResult) SetNormalizedDescription(normalized string) {
	fr.NormalizedDesc = normalized
}
