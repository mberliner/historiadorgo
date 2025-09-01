package jira

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"historiadorgo/internal/domain/entities"
	"historiadorgo/internal/infrastructure/config"
)

func createTestConfig() *config.Config {
	return &config.Config{
		JiraURL:                 "https://test.atlassian.net",
		JiraEmail:               "test@example.com",
		JiraAPIToken:            "test-token",
		DefaultIssueType:        "Story",
		SubtaskIssueType:        "Sub-task",
		FeatureIssueType:        "Feature",
		AcceptanceCriteriaField: "customfield_10001",
	}
}

func TestNewJiraClient(t *testing.T) {
	cfg := createTestConfig()
	client := NewJiraClient(cfg)

	if client == nil {
		t.Fatal("Expected JiraClient to be created")
	}

	if client.config != cfg {
		t.Error("Expected config to be set")
	}

	if client.httpClient == nil {
		t.Error("Expected httpClient to be initialized")
	}

	expectedBaseURL := "https://test.atlassian.net"
	if client.baseURL != expectedBaseURL {
		t.Errorf("Expected baseURL to be %s, got %s", expectedBaseURL, client.baseURL)
	}
}

func TestNewJiraClient_TrimsSlash(t *testing.T) {
	cfg := createTestConfig()
	cfg.JiraURL = "https://test.atlassian.net/"
	client := NewJiraClient(cfg)

	expectedBaseURL := "https://test.atlassian.net"
	if client.baseURL != expectedBaseURL {
		t.Errorf("Expected trailing slash to be trimmed, got %s", client.baseURL)
	}
}

func TestJiraClient_TestConnection(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		responseBody  string
		expectedError string
	}{
		{
			name:          "successful_connection",
			statusCode:    http.StatusOK,
			responseBody:  `{"accountId": "test", "displayName": "Test User"}`,
			expectedError: "",
		},
		{
			name:          "authentication_failed",
			statusCode:    http.StatusUnauthorized,
			responseBody:  `{"errorMessage": "Unauthorized"}`,
			expectedError: "authentication failed: status 401",
		},
		{
			name:          "forbidden_access",
			statusCode:    http.StatusForbidden,
			responseBody:  `{"errorMessage": "Forbidden"}`,
			expectedError: "authentication failed: status 403",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/rest/api/3/myself" {
					t.Errorf("Expected path /rest/api/3/myself, got %s", r.URL.Path)
				}
				if r.Method != "GET" {
					t.Errorf("Expected GET method, got %s", r.Method)
				}

				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			cfg := createTestConfig()
			cfg.JiraURL = server.URL
			client := NewJiraClient(cfg)

			err := client.TestConnection(context.Background())

			if tt.expectedError == "" {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected error containing '%s', got no error", tt.expectedError)
				} else if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing '%s', got: %v", tt.expectedError, err)
				}
			}
		})
	}
}

func TestJiraClient_ValidateProject(t *testing.T) {
	tests := []struct {
		name          string
		projectKey    string
		statusCode    int
		responseBody  string
		expectedError string
	}{
		{
			name:          "valid_project",
			projectKey:    "TEST",
			statusCode:    http.StatusOK,
			responseBody:  `{"key": "TEST", "name": "Test Project"}`,
			expectedError: "",
		},
		{
			name:          "project_not_found",
			projectKey:    "MISSING",
			statusCode:    http.StatusNotFound,
			responseBody:  `{"errorMessage": "Project not found"}`,
			expectedError: "project 'MISSING' not found",
		},
		{
			name:          "server_error",
			projectKey:    "TEST",
			statusCode:    http.StatusInternalServerError,
			responseBody:  `{"errorMessage": "Internal error"}`,
			expectedError: "error validating project: status 500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := "/rest/api/3/project/" + tt.projectKey
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
				}

				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			cfg := createTestConfig()
			cfg.JiraURL = server.URL
			client := NewJiraClient(cfg)

			err := client.ValidateProject(context.Background(), tt.projectKey)

			if tt.expectedError == "" {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected error containing '%s', got no error", tt.expectedError)
				} else if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing '%s', got: %v", tt.expectedError, err)
				}
			}
		})
	}
}

func TestJiraClient_GetIssueTypes(t *testing.T) {
	issueTypes := []map[string]interface{}{
		{
			"id":      "1",
			"name":    "Story",
			"subtask": false,
		},
		{
			"id":      "2",
			"name":    "Sub-task",
			"subtask": true,
		},
		{
			"id":      "3",
			"name":    "Feature",
			"subtask": false,
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/api/3/issuetype" {
			t.Errorf("Expected path /rest/api/3/issuetype, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(issueTypes)
	}))
	defer server.Close()

	cfg := createTestConfig()
	cfg.JiraURL = server.URL
	client := NewJiraClient(cfg)

	result, err := client.GetIssueTypes(context.Background())
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if len(result) != len(issueTypes) {
		t.Errorf("Expected %d issue types, got %d", len(issueTypes), len(result))
	}

	// Verificar el primer issue type
	if result[0]["name"] != "Story" {
		t.Errorf("Expected first issue type to be Story, got %v", result[0]["name"])
	}
}

func TestJiraClient_ValidateSubtaskIssueType(t *testing.T) {
	tests := []struct {
		name          string
		issueTypes    []map[string]interface{}
		expectedError string
	}{
		{
			name: "valid_subtask_type",
			issueTypes: []map[string]interface{}{
				{"name": "Story", "subtask": false},
				{"name": "Sub-task", "subtask": true},
			},
			expectedError: "",
		},
		{
			name: "subtask_type_not_found",
			issueTypes: []map[string]interface{}{
				{"name": "Story", "subtask": false},
				{"name": "Bug", "subtask": false},
			},
			expectedError: "subtask issue type 'Sub-task' not found",
		},
		{
			name: "subtask_type_not_subtask",
			issueTypes: []map[string]interface{}{
				{"name": "Story", "subtask": false},
				{"name": "Sub-task", "subtask": false},
			},
			expectedError: "subtask issue type 'Sub-task' not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(tt.issueTypes)
			}))
			defer server.Close()

			cfg := createTestConfig()
			cfg.JiraURL = server.URL
			client := NewJiraClient(cfg)

			err := client.ValidateSubtaskIssueType(context.Background(), "TEST")

			if tt.expectedError == "" {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected error containing '%s', got no error", tt.expectedError)
				} else if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing '%s', got: %v", tt.expectedError, err)
				}
			}
		})
	}
}

func TestJiraClient_ValidateFeatureIssueType(t *testing.T) {
	tests := []struct {
		name          string
		issueTypes    []map[string]interface{}
		expectedError string
	}{
		{
			name: "valid_feature_type",
			issueTypes: []map[string]interface{}{
				{"name": "Story", "subtask": false},
				{"name": "Feature", "subtask": false},
			},
			expectedError: "",
		},
		{
			name: "feature_type_not_found",
			issueTypes: []map[string]interface{}{
				{"name": "Story", "subtask": false},
				{"name": "Bug", "subtask": false},
			},
			expectedError: "feature issue type 'Feature' not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(tt.issueTypes)
			}))
			defer server.Close()

			cfg := createTestConfig()
			cfg.JiraURL = server.URL
			client := NewJiraClient(cfg)

			err := client.ValidateFeatureIssueType(context.Background())

			if tt.expectedError == "" {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected error containing '%s', got no error", tt.expectedError)
				} else if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing '%s', got: %v", tt.expectedError, err)
				}
			}
		})
	}
}

func TestJiraClient_ValidateParentIssue(t *testing.T) {
	tests := []struct {
		name          string
		issueKey      string
		statusCode    int
		expectedError string
	}{
		{
			name:          "valid_parent_issue",
			issueKey:      "TEST-123",
			statusCode:    http.StatusOK,
			expectedError: "",
		},
		{
			name:          "parent_issue_not_found",
			issueKey:      "TEST-999",
			statusCode:    http.StatusNotFound,
			expectedError: "parent issue 'TEST-999' not found",
		},
		{
			name:          "server_error",
			issueKey:      "TEST-123",
			statusCode:    http.StatusInternalServerError,
			expectedError: "error validating parent issue: status 500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := "/rest/api/3/issue/" + tt.issueKey
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
				}

				w.WriteHeader(tt.statusCode)
				if tt.statusCode == http.StatusOK {
					w.Write([]byte(`{"key": "` + tt.issueKey + `"}`))
				}
			}))
			defer server.Close()

			cfg := createTestConfig()
			cfg.JiraURL = server.URL
			client := NewJiraClient(cfg)

			err := client.ValidateParentIssue(context.Background(), tt.issueKey)

			if tt.expectedError == "" {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected error containing '%s', got no error", tt.expectedError)
				} else if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing '%s', got: %v", tt.expectedError, err)
				}
			}
		})
	}
}

func TestJiraClient_CreateUserStory(t *testing.T) {
	tests := []struct {
		name         string
		story        *entities.UserStory
		statusCode   int
		responseBody string
		expectError  bool
	}{
		{
			name: "successful_story_creation",
			story: entities.NewUserStory(
				"Test Story",
				"Test Description",
				"Test Criteria",
				"",
				"",
			),
			statusCode:   http.StatusCreated,
			responseBody: `{"id": "10001", "key": "TEST-123", "self": "https://test.atlassian.net/rest/api/3/issue/10001"}`,
			expectError:  false,
		},
		{
			name: "story_with_subtasks",
			story: entities.NewUserStory(
				"Test Story",
				"Test Description",
				"Test Criteria",
				"Task 1;Task 2",
				"",
			),
			statusCode:   http.StatusCreated,
			responseBody: `{"id": "10002", "key": "TEST-124", "self": "https://test.atlassian.net/rest/api/3/issue/10002"}`,
			expectError:  false,
		},
		{
			name: "story_creation_failed",
			story: entities.NewUserStory(
				"Test Story",
				"Test Description",
				"Test Criteria",
				"",
				"",
			),
			statusCode:   http.StatusBadRequest,
			responseBody: `{"errorMessages": ["Invalid request"]}`,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issueCreationCount := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/rest/api/3/issue" && r.Method == "POST" {
					issueCreationCount++

					if issueCreationCount == 1 {
						// Primera llamada - crear historia principal
						w.WriteHeader(tt.statusCode)
						w.Write([]byte(tt.responseBody))
					} else {
						// Llamadas siguientes - crear subtareas
						w.WriteHeader(http.StatusCreated)
						subtaskKey := "TEST-SUB" + string(rune(issueCreationCount+123))
						response := `{"id": "` + string(rune(issueCreationCount+10000)) + `", "key": "` + subtaskKey + `"}`
						w.Write([]byte(response))
					}
				}
			}))
			defer server.Close()

			cfg := createTestConfig()
			cfg.JiraURL = server.URL
			cfg.ProjectKey = "TEST"
			client := NewJiraClient(cfg)

			result, err := client.CreateUserStory(context.Background(), tt.story, 1)

			if tt.expectError {
				if result.Success {
					t.Error("Expected result to be unsuccessful")
				}
				if result.ErrorMessage == "" {
					t.Error("Expected error message to be set")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				if !result.Success {
					t.Errorf("Expected result to be successful, got error: %s", result.ErrorMessage)
				}
				if result.IssueKey == "" {
					t.Error("Expected issue key to be set")
				}
				if result.IssueURL == "" {
					t.Error("Expected issue URL to be set")
				}

				// Verificar subtareas si las hay
				if tt.story.HasSubtareas() {
					expectedSubtasks := len(tt.story.GetValidSubtareas())
					if len(result.Subtareas) != expectedSubtasks {
						t.Errorf("Expected %d subtask results, got %d", expectedSubtasks, len(result.Subtareas))
					}
				}
			}
		})
	}
}

func TestJiraClient_isJiraKey(t *testing.T) {
	cfg := createTestConfig()
	client := NewJiraClient(cfg)

	tests := []struct {
		input    string
		expected bool
	}{
		{"TEST-123", true},
		{"PROJ-456", true},
		{"A-1", true},
		{"ABC-999", true},
		{"test-123", false}, // lowercase
		{"TEST", false},     // no number
		{"123-TEST", false}, // number first
		{"TEST-", false},    // no number after dash
		{"TEST-abc", false}, // letters after dash
		{"", false},         // empty
		{"INVALID", false},  // no dash
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := client.isJiraKey(tt.input)
			if result != tt.expected {
				t.Errorf("Expected isJiraKey(%s) to be %v, got %v", tt.input, tt.expected, result)
			}
		})
	}
}

func TestJiraClient_buildIssuePayload(t *testing.T) {
	cfg := createTestConfig()
	client := NewJiraClient(cfg)

	story := entities.NewUserStory(
		"Test Story",
		"Test Description",
		"Test Criteria",
		"Task 1;Task 2",
		"TEST-123",
	)

	payload := client.buildIssuePayload(story, "PROJ")

	fields, ok := payload["fields"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected fields to be a map")
	}

	// Verificar project
	project, ok := fields["project"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected project to be a map")
	}
	if project["key"] != "PROJ" {
		t.Errorf("Expected project key to be PROJ, got %v", project["key"])
	}

	// Verificar summary
	if fields["summary"] != story.Titulo {
		t.Errorf("Expected summary to be %s, got %v", story.Titulo, fields["summary"])
	}

	// Verificar issue type
	issueType, ok := fields["issuetype"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected issuetype to be a map")
	}
	if issueType["name"] != cfg.DefaultIssueType {
		t.Errorf("Expected issue type to be %s, got %v", cfg.DefaultIssueType, issueType["name"])
	}

	// Verificar parent (debe estar presente porque es una clave Jira válida)
	parent, exists := fields["parent"]
	if !exists {
		t.Error("Expected parent field to be present")
	} else {
		parentMap, ok := parent.(map[string]interface{})
		if !ok {
			t.Fatal("Expected parent to be a map")
		}
		if parentMap["key"] != "TEST-123" {
			t.Errorf("Expected parent key to be TEST-123, got %v", parentMap["key"])
		}
	}

	// Verificar que description está presente (ADF format)
	if _, exists := fields["description"]; !exists {
		t.Error("Expected description field to be present")
	}

	// Verificar campo de criterios de aceptación personalizado
	if _, exists := fields[cfg.AcceptanceCriteriaField]; !exists {
		t.Error("Expected acceptance criteria field to be present")
	}
}

func TestJiraClient_buildSubtaskPayload(t *testing.T) {
	cfg := createTestConfig()
	client := NewJiraClient(cfg)

	description := "Test subtask description"
	parentKey := "TEST-123"
	projectKey := "PROJ"

	payload := client.buildSubtaskPayload(description, parentKey, projectKey)

	fields, ok := payload["fields"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected fields to be a map")
	}

	// Verificar project
	project, ok := fields["project"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected project to be a map")
	}
	if project["key"] != projectKey {
		t.Errorf("Expected project key to be %s, got %v", projectKey, project["key"])
	}

	// Verificar summary
	if fields["summary"] != description {
		t.Errorf("Expected summary to be %s, got %v", description, fields["summary"])
	}

	// Verificar issue type
	issueType, ok := fields["issuetype"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected issuetype to be a map")
	}
	if issueType["name"] != cfg.SubtaskIssueType {
		t.Errorf("Expected issue type to be %s, got %v", cfg.SubtaskIssueType, issueType["name"])
	}

	// Verificar parent
	parent, ok := fields["parent"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected parent to be a map")
	}
	if parent["key"] != parentKey {
		t.Errorf("Expected parent key to be %s, got %v", parentKey, parent["key"])
	}

	// Verificar que description está presente
	if _, exists := fields["description"]; !exists {
		t.Error("Expected description field to be present")
	}
}

func TestJiraClient_GetIssueTypes_ErrorCases(t *testing.T) {
	tests := []struct {
		name          string
		serverFunc    func(w http.ResponseWriter, r *http.Request)
		expectedError string
	}{
		{
			name: "HTTP error response",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			expectedError: "error getting issue types: status 500",
		},
		{
			name: "Invalid JSON response",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("invalid json"))
			},
			expectedError: "error decoding response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverFunc))
			defer server.Close()

			cfg := createTestConfig()
			cfg.JiraURL = server.URL
			client := NewJiraClient(cfg)

			_, err := client.GetIssueTypes(context.Background())
			if err == nil {
				t.Error("Expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.expectedError) {
				t.Errorf("Expected error to contain %q, got: %v", tt.expectedError, err)
			}
		})
	}
}

func TestJiraClient_createIssue_ComprehensiveTests(t *testing.T) {
	tests := []struct {
		name          string
		payload       map[string]interface{}
		serverFunc    func(w http.ResponseWriter, r *http.Request)
		expectedError string
		expectSuccess bool
	}{
		{
			name: "successful_issue_creation",
			payload: map[string]interface{}{
				"fields": map[string]interface{}{
					"project":   map[string]interface{}{"key": "TEST"},
					"summary":   "Test Issue",
					"issuetype": map[string]interface{}{"name": "Story"},
				},
			},
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "POST" || r.URL.Path != "/rest/api/3/issue" {
					t.Errorf("Expected POST to /rest/api/3/issue, got %s %s", r.Method, r.URL.Path)
				}
				w.WriteHeader(http.StatusCreated)
				response := JiraCreateResponse{
					ID:   "10001",
					Key:  "TEST-123",
					Self: "https://test.atlassian.net/rest/api/3/issue/10001",
				}
				json.NewEncoder(w).Encode(response)
			},
			expectSuccess: true,
		},
		{
			name: "json_marshal_error",
			payload: map[string]interface{}{
				"fields": map[string]interface{}{
					"invalid": make(chan int), // channels can't be marshaled to JSON
				},
			},
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				// Should not reach here
				t.Error("Server should not be called for marshal error")
			},
			expectedError: "error marshaling payload",
			expectSuccess: false,
		},
		{
			name: "http_request_error",
			payload: map[string]interface{}{
				"fields": map[string]interface{}{
					"project": map[string]interface{}{"key": "TEST"},
				},
			},
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				// Simulate connection error by closing immediately
				hj, ok := w.(http.Hijacker)
				if !ok {
					http.Error(w, "webserver doesn't support hijacking", http.StatusInternalServerError)
					return
				}
				conn, _, _ := hj.Hijack()
				conn.Close()
			},
			expectedError: "error creating issue",
			expectSuccess: false,
		},
		{
			name: "bad_request_with_jira_error_response",
			payload: map[string]interface{}{
				"fields": map[string]interface{}{
					"project": map[string]interface{}{"key": "TEST"},
				},
			},
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				errorResp := JiraErrorResponse{
					ErrorMessages: []string{"The project key is invalid"},
					Errors: map[string]string{
						"project": "Project does not exist",
						"summary": "Summary is required",
					},
				}
				json.NewEncoder(w).Encode(errorResp)
			},
			expectedError: "jira error: The project key is invalid; project: Project does not exist; summary: Summary is required",
			expectSuccess: false,
		},
		{
			name: "bad_request_with_invalid_error_response",
			payload: map[string]interface{}{
				"fields": map[string]interface{}{
					"project": map[string]interface{}{"key": "TEST"},
				},
			},
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Invalid JSON error response"))
			},
			expectedError: "error creating issue: status 400",
			expectSuccess: false,
		},
		{
			name: "server_internal_error",
			payload: map[string]interface{}{
				"fields": map[string]interface{}{
					"project": map[string]interface{}{"key": "TEST"},
				},
			},
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Internal server error"))
			},
			expectedError: "error creating issue: status 500",
			expectSuccess: false,
		},
		{
			name: "invalid_success_response_json",
			payload: map[string]interface{}{
				"fields": map[string]interface{}{
					"project": map[string]interface{}{"key": "TEST"},
				},
			},
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusCreated)
				w.Write([]byte("invalid json"))
			},
			expectedError: "error parsing response",
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverFunc))
			defer server.Close()

			cfg := createTestConfig()
			cfg.JiraURL = server.URL
			client := NewJiraClient(cfg)

			result, err := client.createIssue(context.Background(), tt.payload)

			if tt.expectSuccess {
				if err != nil {
					t.Errorf("Expected success, got error: %v", err)
				}
				if result == nil {
					t.Error("Expected non-nil result for successful creation")
				}
				if result != nil && result.Key != "TEST-123" {
					t.Errorf("Expected key TEST-123, got %s", result.Key)
				}
			} else {
				if err == nil {
					t.Error("Expected error, got success")
				}
				if tt.expectedError != "" && !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error to contain %q, got: %v", tt.expectedError, err)
				}
			}
		})
	}
}

func TestJiraClient_createSubtasks_ComprehensiveTests(t *testing.T) {
	tests := []struct {
		name            string
		story           *entities.UserStory
		parentKey       string
		serverFunc      func(w http.ResponseWriter, r *http.Request)
		expectedCalls   int
		expectedSuccess int
		expectedFailed  int
	}{
		{
			name: "successful_subtasks_creation",
			story: &entities.UserStory{
				Titulo:      "Test Story",
				Descripcion: "Test Description",
				Subtareas:   []string{"Subtask 1", "Subtask 2", "Subtask 3"},
			},
			parentKey: "TEST-100",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "POST" || r.URL.Path != "/rest/api/3/issue" {
					t.Errorf("Expected POST to /rest/api/3/issue, got %s %s", r.Method, r.URL.Path)
				}
				w.WriteHeader(http.StatusCreated)
				response := JiraCreateResponse{
					ID:   "10001",
					Key:  "TEST-101",
					Self: "https://test.atlassian.net/rest/api/3/issue/10001",
				}
				json.NewEncoder(w).Encode(response)
			},
			expectedCalls:   3,
			expectedSuccess: 3,
			expectedFailed:  0,
		},
		{
			name: "mixed_success_and_failure",
			story: &entities.UserStory{
				Titulo:      "Test Story",
				Descripcion: "Test Description",
				Subtareas:   []string{"Subtask 1", "Subtask 2"},
			},
			parentKey: "TEST-100",
			serverFunc: func() func(w http.ResponseWriter, r *http.Request) {
				callCount := 0
				return func(w http.ResponseWriter, r *http.Request) {
					callCount++
					if callCount == 1 {
						// First call succeeds
						w.WriteHeader(http.StatusCreated)
						response := JiraCreateResponse{
							ID:   "10001",
							Key:  "TEST-101",
							Self: "https://test.atlassian.net/rest/api/3/issue/10001",
						}
						json.NewEncoder(w).Encode(response)
					} else {
						// Second call fails
						w.WriteHeader(http.StatusBadRequest)
						errorResp := JiraErrorResponse{
							ErrorMessages: []string{"Invalid subtask"},
						}
						json.NewEncoder(w).Encode(errorResp)
					}
				}
			}(),
			expectedCalls:   2,
			expectedSuccess: 1,
			expectedFailed:  1,
		},
		{
			name: "no_valid_subtasks",
			story: &entities.UserStory{
				Titulo:      "Test Story",
				Descripcion: "Test Description",
				Subtareas:   []string{}, // No subtasks
			},
			parentKey: "TEST-100",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				t.Error("Server should not be called when there are no subtasks")
			},
			expectedCalls:   0,
			expectedSuccess: 0,
			expectedFailed:  0,
		},
		{
			name: "all_subtasks_fail",
			story: &entities.UserStory{
				Titulo:      "Test Story",
				Descripcion: "Test Description",
				Subtareas:   []string{"Subtask 1", "Subtask 2"},
			},
			parentKey: "TEST-100",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Server error"))
			},
			expectedCalls:   2,
			expectedSuccess: 0,
			expectedFailed:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				callCount++
				tt.serverFunc(w, r)
			}))
			defer server.Close()

			cfg := createTestConfig()
			cfg.JiraURL = server.URL
			client := NewJiraClient(cfg)

			result := entities.NewProcessResult(1)
			client.createSubtasks(context.Background(), tt.story, tt.parentKey, result)

			// Verify number of calls made
			if callCount != tt.expectedCalls {
				t.Errorf("Expected %d server calls, got %d", tt.expectedCalls, callCount)
			}

			// Verify subtask results
			successfulSubtasks := result.GetSuccessfulSubtasks()
			failedSubtasks := result.GetFailedSubtasks()

			if len(successfulSubtasks) != tt.expectedSuccess {
				t.Errorf("Expected %d successful subtasks, got %d", tt.expectedSuccess, len(successfulSubtasks))
			}

			if len(failedSubtasks) != tt.expectedFailed {
				t.Errorf("Expected %d failed subtasks, got %d", tt.expectedFailed, len(failedSubtasks))
			}

			// Verify successful subtasks have proper URLs
			for _, subtask := range successfulSubtasks {
				if subtask.IssueURL == "" {
					t.Error("Successful subtask should have issue URL")
				}
				if !strings.Contains(subtask.IssueURL, server.URL) {
					t.Errorf("Issue URL should contain server URL. Got: %s", subtask.IssueURL)
				}
			}

			// Verify failed subtasks have error messages
			for _, subtask := range failedSubtasks {
				if subtask.Error == "" {
					t.Error("Failed subtask should have error message")
				}
			}
		})
	}
}

func TestJiraClient_TestConnection_NetworkErrors(t *testing.T) {
	tests := []struct {
		name          string
		serverFunc    func(w http.ResponseWriter, r *http.Request)
		expectedError string
	}{
		{
			name: "connection_timeout",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				// Simulate network timeout by hijacking connection
				hj, ok := w.(http.Hijacker)
				if !ok {
					http.Error(w, "webserver doesn't support hijacking", http.StatusInternalServerError)
					return
				}
				conn, _, _ := hj.Hijack()
				conn.Close()
			},
			expectedError: "connection failed",
		},
		{
			name: "request_creation_error",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				// This test should not reach the server
				t.Error("Server should not be called for request creation error")
			},
			expectedError: "error creating request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "request_creation_error" {
				// Test with invalid URL to cause request creation error
				cfg := createTestConfig()
				cfg.JiraURL = "://invalid-url"
				client := NewJiraClient(cfg)

				err := client.TestConnection(context.Background())
				if err == nil {
					t.Error("Expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error to contain %q, got: %v", tt.expectedError, err)
				}
			} else {
				server := httptest.NewServer(http.HandlerFunc(tt.serverFunc))
				defer server.Close()

				cfg := createTestConfig()
				cfg.JiraURL = server.URL
				client := NewJiraClient(cfg)

				err := client.TestConnection(context.Background())
				if err == nil {
					t.Error("Expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error to contain %q, got: %v", tt.expectedError, err)
				}
			}
		})
	}
}

func TestJiraClient_ValidateProject_NetworkErrors(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc     func() (*JiraClient, string) // returns client and project key
		expectedError string
	}{
		{
			name: "request_creation_error",
			setupFunc: func() (*JiraClient, string) {
				cfg := createTestConfig()
				cfg.JiraURL = "://invalid-url"
				client := NewJiraClient(cfg)
				return client, "TEST"
			},
			expectedError: "error creating request",
		},
		{
			name: "network_error",
			setupFunc: func() (*JiraClient, string) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					hj, ok := w.(http.Hijacker)
					if !ok {
						http.Error(w, "webserver doesn't support hijacking", http.StatusInternalServerError)
						return
					}
					conn, _, _ := hj.Hijack()
					conn.Close()
				}))
				defer server.Close()

				cfg := createTestConfig()
				cfg.JiraURL = server.URL
				client := NewJiraClient(cfg)
				return client, "TEST"
			},
			expectedError: "error validating project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, projectKey := tt.setupFunc()

			err := client.ValidateProject(context.Background(), projectKey)
			if err == nil {
				t.Error("Expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.expectedError) {
				t.Errorf("Expected error to contain %q, got: %v", tt.expectedError, err)
			}
		})
	}
}

func TestJiraClient_ValidateSubtaskIssueType_ErrorPaths(t *testing.T) {
	tests := []struct {
		name          string
		serverFunc    func(w http.ResponseWriter, r *http.Request)
		expectedError string
	}{
		{
			name: "get_issue_types_error",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Server error"))
			},
			expectedError: "error getting issue types",
		},
		{
			name: "subtask_name_matches_but_not_subtask",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				issueTypes := []map[string]interface{}{
					{
						"name":    "Sub-task",
						"subtask": false, // Wrong: should be true for subtasks
					},
				}
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(issueTypes)
			},
			expectedError: "subtask issue type 'Sub-task' not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverFunc))
			defer server.Close()

			cfg := createTestConfig()
			cfg.JiraURL = server.URL
			client := NewJiraClient(cfg)

			err := client.ValidateSubtaskIssueType(context.Background(), "TEST")
			if err == nil {
				t.Error("Expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.expectedError) {
				t.Errorf("Expected error to contain %q, got: %v", tt.expectedError, err)
			}
		})
	}
}

func TestJiraClient_ValidateFeatureIssueType_ErrorPaths(t *testing.T) {
	tests := []struct {
		name          string
		serverFunc    func(w http.ResponseWriter, r *http.Request)
		expectedError string
	}{
		{
			name: "get_issue_types_error",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Server error"))
			},
			expectedError: "error getting issue types",
		},
		{
			name: "feature_type_not_found_in_response",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				issueTypes := []map[string]interface{}{
					{
						"name":    "Story",
						"subtask": false,
					},
					{
						"name":    "Bug",
						"subtask": false,
					},
				}
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(issueTypes)
			},
			expectedError: "feature issue type 'Feature' not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverFunc))
			defer server.Close()

			cfg := createTestConfig()
			cfg.JiraURL = server.URL
			client := NewJiraClient(cfg)

			err := client.ValidateFeatureIssueType(context.Background())
			if err == nil {
				t.Error("Expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.expectedError) {
				t.Errorf("Expected error to contain %q, got: %v", tt.expectedError, err)
			}
		})
	}
}

func TestJiraClient_buildIssuePayload_VariousScenarios(t *testing.T) {
	tests := []struct {
		name                        string
		acceptanceCriteriaField     string
		parent                      string
		expectedHasAcceptanceField  bool
		expectedHasParent           bool
		expectedDescriptionIncludes string
	}{
		{
			name:                        "no_acceptance_criteria_field",
			acceptanceCriteriaField:     "",
			parent:                      "",
			expectedHasAcceptanceField:  false,
			expectedHasParent:           false,
			expectedDescriptionIncludes: "criteria", // Should include criteria in description
		},
		{
			name:                        "with_acceptance_field_no_parent",
			acceptanceCriteriaField:     "customfield_10001",
			parent:                      "",
			expectedHasAcceptanceField:  true,
			expectedHasParent:           false,
			expectedDescriptionIncludes: "description", // Should have separate description
		},
		{
			name:                        "with_non_jira_key_parent",
			acceptanceCriteriaField:     "customfield_10001",
			parent:                      "Some Feature Description",
			expectedHasAcceptanceField:  true,
			expectedHasParent:           false,
			expectedDescriptionIncludes: "description",
		},
		{
			name:                        "with_jira_key_parent",
			acceptanceCriteriaField:     "customfield_10001",
			parent:                      "PROJ-123",
			expectedHasAcceptanceField:  true,
			expectedHasParent:           true,
			expectedDescriptionIncludes: "description",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := createTestConfig()
			cfg.AcceptanceCriteriaField = tt.acceptanceCriteriaField
			client := NewJiraClient(cfg)

			story := entities.NewUserStory(
				"Test Story",
				"Test Description",
				"Test Criteria",
				"",
				tt.parent,
			)

			payload := client.buildIssuePayload(story, "PROJ")

			fields, ok := payload["fields"].(map[string]interface{})
			if !ok {
				t.Fatal("Expected fields to be a map")
			}

			// Check acceptance criteria field
			if tt.expectedHasAcceptanceField {
				if _, exists := fields[tt.acceptanceCriteriaField]; !exists {
					t.Errorf("Expected acceptance criteria field %s to be present", tt.acceptanceCriteriaField)
				}
			} else {
				if tt.acceptanceCriteriaField != "" {
					if _, exists := fields[tt.acceptanceCriteriaField]; exists {
						t.Error("Should not have acceptance criteria field when not configured")
					}
				}
			}

			// Check parent field
			if tt.expectedHasParent {
				if _, exists := fields["parent"]; !exists {
					t.Error("Expected parent field to be present")
				}
			} else {
				if _, exists := fields["parent"]; exists {
					t.Error("Should not have parent field")
				}
			}

			// Verify description is present
			if _, exists := fields["description"]; !exists {
				t.Error("Expected description field to be present")
			}
		})
	}
}
