package jira

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"historiadorgo/internal/domain/entities"
	"historiadorgo/internal/infrastructure/config"
)

type JiraClient struct {
	config     *config.Config
	httpClient *http.Client
	baseURL    string
}

type JiraIssue struct {
	ID     string                 `json:"id"`
	Key    string                 `json:"key"`
	Self   string                 `json:"self"`
	Fields map[string]interface{} `json:"fields"`
}

type JiraCreateResponse struct {
	ID   string `json:"id"`
	Key  string `json:"key"`
	Self string `json:"self"`
}

type JiraErrorResponse struct {
	ErrorMessages   []string          `json:"errorMessages"`
	Errors          map[string]string `json:"errors"`
	WarningMessages []string          `json:"warningMessages"`
}

func NewJiraClient(cfg *config.Config) *JiraClient {
	return &JiraClient{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: strings.TrimSuffix(cfg.JiraURL, "/"),
	}
}

func (jc *JiraClient) TestConnection(ctx context.Context) error {
	req, err := jc.createRequest(ctx, "GET", "/rest/api/3/myself", nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	resp, err := jc.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("authentication failed: status %d", resp.StatusCode)
	}

	return nil
}

func (jc *JiraClient) ValidateProject(ctx context.Context, projectKey string) error {
	endpoint := fmt.Sprintf("/rest/api/3/project/%s", projectKey)
	req, err := jc.createRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	resp, err := jc.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error validating project: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("project '%s' not found", projectKey)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error validating project: status %d", resp.StatusCode)
	}

	return nil
}

func (jc *JiraClient) ValidateSubtaskIssueType(ctx context.Context, projectKey string) error {
	issueTypes, err := jc.GetIssueTypes(ctx)
	if err != nil {
		return fmt.Errorf("error getting issue types: %w", err)
	}

	for _, issueType := range issueTypes {
		if name, ok := issueType["name"].(string); ok && name == jc.config.SubtaskIssueType {
			if subtask, ok := issueType["subtask"].(bool); ok && subtask {
				return nil
			}
		}
	}

	return fmt.Errorf("subtask issue type '%s' not found", jc.config.SubtaskIssueType)
}

func (jc *JiraClient) ValidateFeatureIssueType(ctx context.Context) error {
	issueTypes, err := jc.GetIssueTypes(ctx)
	if err != nil {
		return fmt.Errorf("error getting issue types: %w", err)
	}

	for _, issueType := range issueTypes {
		if name, ok := issueType["name"].(string); ok && name == jc.config.FeatureIssueType {
			return nil
		}
	}

	return fmt.Errorf("feature issue type '%s' not found", jc.config.FeatureIssueType)
}

func (jc *JiraClient) ValidateParentIssue(ctx context.Context, issueKey string) error {
	endpoint := fmt.Sprintf("/rest/api/3/issue/%s", issueKey)
	req, err := jc.createRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	resp, err := jc.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error validating parent issue: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("parent issue '%s' not found", issueKey)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error validating parent issue: status %d", resp.StatusCode)
	}

	return nil
}

func (jc *JiraClient) CreateUserStory(ctx context.Context, story *entities.UserStory, rowNumber int) (*entities.ProcessResult, error) {
	result := entities.NewProcessResult(rowNumber)

	issuePayload := jc.buildIssuePayload(story, jc.config.ProjectKey)

	issue, err := jc.createIssue(ctx, issuePayload)
	if err != nil {
		result.Success = false
		result.ErrorMessage = err.Error()
		return result, nil
	}

	result.Success = true
	result.IssueKey = issue.Key
	result.IssueURL = fmt.Sprintf("%s/browse/%s", jc.baseURL, issue.Key)

	if story.HasSubtareas() {
		jc.createSubtasks(ctx, story, issue.Key, result)
	}

	return result, nil
}

func (jc *JiraClient) GetIssueTypes(ctx context.Context) ([]map[string]interface{}, error) {
	req, err := jc.createRequest(ctx, "GET", "/rest/api/3/issuetype", nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := jc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error getting issue types: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error getting issue types: status %d", resp.StatusCode)
	}

	var issueTypes []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&issueTypes); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return issueTypes, nil
}

func (jc *JiraClient) createIssue(ctx context.Context, payload map[string]interface{}) (*JiraCreateResponse, error) {
	reqBody, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("error marshaling payload: %w", err)
	}

	req, err := jc.createRequest(ctx, "POST", "/rest/api/3/issue", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := jc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error creating issue: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		var errorResp JiraErrorResponse
		if err := json.Unmarshal(body, &errorResp); err == nil {
			errorMsg := strings.Join(errorResp.ErrorMessages, "; ")
			for field, msg := range errorResp.Errors {
				errorMsg += fmt.Sprintf("; %s: %s", field, msg)
			}
			return nil, fmt.Errorf("jira error: %s", errorMsg)
		}
		return nil, fmt.Errorf("error creating issue: status %d, body: %s", resp.StatusCode, string(body))
	}

	var createResp JiraCreateResponse
	if err := json.Unmarshal(body, &createResp); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	return &createResp, nil
}

func (jc *JiraClient) createSubtasks(ctx context.Context, story *entities.UserStory, parentKey string, result *entities.ProcessResult) {
	validSubtasks := story.GetValidSubtareas()

	for _, subtaskDesc := range validSubtasks {
		subtaskPayload := jc.buildSubtaskPayload(subtaskDesc, parentKey, jc.config.ProjectKey)

		subtask, err := jc.createIssue(ctx, subtaskPayload)
		if err != nil {
			result.AddSubtaskResult(subtaskDesc, false, "", "", err.Error())
			continue
		}

		subtaskURL := fmt.Sprintf("%s/browse/%s", jc.baseURL, subtask.Key)
		result.AddSubtaskResult(subtaskDesc, true, subtask.Key, subtaskURL, "")
	}
}

func (jc *JiraClient) buildIssuePayload(story *entities.UserStory, projectKey string) map[string]interface{} {
	fields := map[string]interface{}{
		"project": map[string]interface{}{
			"key": projectKey,
		},
		"summary": story.Titulo,
		"issuetype": map[string]interface{}{
			"name": jc.config.DefaultIssueType,
		},
	}

	// Usar ADF para descripción y criterios de aceptación
	if jc.config.AcceptanceCriteriaField != "" {
		// Si hay campo personalizado para criterios, usar descripción simple y criterios en campo separado
		fields["description"] = CreateDescriptionADF(story.Descripcion)
		fields[jc.config.AcceptanceCriteriaField] = CreateAcceptanceCriteriaADF(story.CriterioAceptacion)
	} else {
		// Si no hay campo personalizado, incluir criterios en la descripción
		fields["description"] = CreateDescriptionWithCriteriaADF(story.Descripcion, story.CriterioAceptacion)
	}

	if story.HasParent() && jc.isJiraKey(story.Parent) {
		fields["parent"] = map[string]interface{}{
			"key": story.Parent,
		}
	}

	return map[string]interface{}{
		"fields": fields,
	}
}

func (jc *JiraClient) buildSubtaskPayload(description, parentKey, projectKey string) map[string]interface{} {
	return map[string]interface{}{
		"fields": map[string]interface{}{
			"project": map[string]interface{}{
				"key": projectKey,
			},
			"summary":     description,
			"description": CreateDescriptionADF(description),
			"issuetype": map[string]interface{}{
				"name": jc.config.SubtaskIssueType,
			},
			"parent": map[string]interface{}{
				"key": parentKey,
			},
		},
	}
}

func (jc *JiraClient) createRequest(ctx context.Context, method, endpoint string, body io.Reader) (*http.Request, error) {
	fullURL := jc.baseURL + endpoint

	req, err := http.NewRequestWithContext(ctx, method, fullURL, body)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(jc.config.JiraEmail, jc.config.JiraAPIToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	return req, nil
}

func (jc *JiraClient) isJiraKey(str string) bool {
	re := regexp.MustCompile(`^[A-Z]+-\d+$`)
	return re.MatchString(str)
}
