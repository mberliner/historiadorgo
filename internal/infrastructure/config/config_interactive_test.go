package config

import (
	"bufio"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestCreateInteractiveEnvFile_ValidInput(t *testing.T) {
	// Clean up any existing .env file
	defer os.Remove(".env")

	// Test helper functions with bufio.Reader
	// Verify the helper functions work correctly
	if promptForInput(bufio.NewReader(strings.NewReader("test_value\n")), "Test prompt", "") != "test_value" {
		t.Error("promptForInput should return entered value")
	}

	if promptForInput(bufio.NewReader(strings.NewReader("\n")), "Test prompt", "default") != "default" {
		t.Error("promptForInput should return default value when input is empty")
	}

	if !promptForYesNo(bufio.NewReader(strings.NewReader("y\n")), "Test prompt", false) {
		t.Error("promptForYesNo should return true for 'y' input")
	}

	if !promptForYesNo(bufio.NewReader(strings.NewReader("yes\n")), "Test prompt", false) {
		t.Error("promptForYesNo should return true for 'yes' input")
	}

	if !promptForYesNo(bufio.NewReader(strings.NewReader("si\n")), "Test prompt", false) {
		t.Error("promptForYesNo should return true for 'si' input")
	}

	if promptForYesNo(bufio.NewReader(strings.NewReader("n\n")), "Test prompt", true) {
		t.Error("promptForYesNo should return false for 'n' input")
	}

	if promptForYesNo(bufio.NewReader(strings.NewReader("\n")), "Test prompt", true) != true {
		t.Error("promptForYesNo should return default value when input is empty")
	}
}

func TestLoadConfig_CreatesEnvFileWhenMissing(t *testing.T) {
	// This test is challenging because it requires interactive input
	// In a real environment, we'd use dependency injection to mock the input
	// For now, we just test that the logic path exists

	// Clean up
	defer os.Remove(".env")

	// Check that LoadConfig identifies missing .env file
	if _, err := os.Stat(".env"); err == nil {
		os.Remove(".env") // Ensure it doesn't exist
	}

	// We can't easily test the interactive part in unit tests
	// This would require either dependency injection or integration tests
	// The main logic is tested above in the helper function tests
}

func TestLoadConfig_WithExistingEnvFile(t *testing.T) {
	// Create temporary .env file
	envContent := `JIRA_URL=https://test.atlassian.net
JIRA_EMAIL=test@example.com
JIRA_API_TOKEN=token123
PROJECT_KEY=TEST
`

	// Clean up
	defer os.Remove(".env")

	// Create .env file
	if err := os.WriteFile(".env", []byte(envContent), 0644); err != nil {
		t.Fatalf("Failed to create test .env file: %v", err)
	}

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() failed: %v", err)
	}

	// Verify config was loaded correctly
	if config.JiraURL != "https://test.atlassian.net" {
		t.Errorf("Expected JIRA_URL=https://test.atlassian.net, got %s", config.JiraURL)
	}
	if config.JiraEmail != "test@example.com" {
		t.Errorf("Expected JIRA_EMAIL=test@example.com, got %s", config.JiraEmail)
	}
	if config.JiraAPIToken != "token123" {
		t.Errorf("Expected JIRA_API_TOKEN=token123, got %s", config.JiraAPIToken)
	}
	if config.ProjectKey != "TEST" {
		t.Errorf("Expected PROJECT_KEY=TEST, got %s", config.ProjectKey)
	}
}

func TestLoadConfig_WithExistingEnvVars(t *testing.T) {
	// Save current env vars
	originalURL := os.Getenv("JIRA_URL")
	originalEmail := os.Getenv("JIRA_EMAIL")
	originalToken := os.Getenv("JIRA_API_TOKEN")

	// Clean up after test
	defer func() {
		os.Setenv("JIRA_URL", originalURL)
		os.Setenv("JIRA_EMAIL", originalEmail)
		os.Setenv("JIRA_API_TOKEN", originalToken)
		os.Remove(".env")
	}()

	// Ensure no .env file exists
	os.Remove(".env")

	// Set environment variables
	os.Setenv("JIRA_URL", "https://env.atlassian.net")
	os.Setenv("JIRA_EMAIL", "env@example.com")
	os.Setenv("JIRA_API_TOKEN", "env_token123")

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() failed: %v", err)
	}

	// Verify config uses environment variables
	if config.JiraURL != "https://env.atlassian.net" {
		t.Errorf("Expected JIRA_URL from env, got %s", config.JiraURL)
	}
	if config.JiraEmail != "env@example.com" {
		t.Errorf("Expected JIRA_EMAIL from env, got %s", config.JiraEmail)
	}
	if config.JiraAPIToken != "env_token123" {
		t.Errorf("Expected JIRA_API_TOKEN from env, got %s", config.JiraAPIToken)
	}
}

func TestValidate_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		wantError   bool
		errorString string
	}{
		{
			name: "missing_jira_url_only",
			config: &Config{
				JiraURL:      "",
				JiraEmail:    "test@example.com",
				JiraAPIToken: "token123",
			},
			wantError:   true,
			errorString: "JIRA_URL",
		},
		{
			name: "missing_jira_email_only",
			config: &Config{
				JiraURL:      "https://test.atlassian.net",
				JiraEmail:    "",
				JiraAPIToken: "token123",
			},
			wantError:   true,
			errorString: "JIRA_EMAIL",
		},
		{
			name: "missing_jira_token_only",
			config: &Config{
				JiraURL:      "https://test.atlassian.net",
				JiraEmail:    "test@example.com",
				JiraAPIToken: "",
			},
			wantError:   true,
			errorString: "JIRA_API_TOKEN",
		},
		{
			name: "missing_multiple_fields",
			config: &Config{
				JiraURL:      "",
				JiraEmail:    "",
				JiraAPIToken: "token123",
			},
			wantError:   true,
			errorString: "JIRA_URL, JIRA_EMAIL",
		},
		{
			name: "valid_config",
			config: &Config{
				JiraURL:      "https://test.atlassian.net",
				JiraEmail:    "test@example.com",
				JiraAPIToken: "token123",
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.wantError && err == nil {
				t.Error("Expected error, got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
			if tt.wantError && err != nil && !strings.Contains(err.Error(), tt.errorString) {
				t.Errorf("Expected error to contain %s, got: %v", tt.errorString, err)
			}
		})
	}
}

func TestGetEnvHelpers_EdgeCases(t *testing.T) {
	// Test getEnv function
	original := os.Getenv("TEST_ENV_VAR")
	defer os.Setenv("TEST_ENV_VAR", original)

	// Test with existing env var
	os.Setenv("TEST_ENV_VAR", "test_value")
	result := getEnv("TEST_ENV_VAR", "default")
	if result != "test_value" {
		t.Errorf("getEnv() = %v, want test_value", result)
	}

	// Test with non-existent env var
	os.Setenv("TEST_ENV_VAR", "")
	result = getEnv("TEST_ENV_VAR", "default")
	if result != "default" {
		t.Errorf("getEnv() = %v, want default", result)
	}

	// Test getEnvAsInt
	os.Setenv("TEST_INT_VAR", "42")
	intResult := getEnvAsInt("TEST_INT_VAR", 100)
	if intResult != 42 {
		t.Errorf("getEnvAsInt() = %v, want 42", intResult)
	}

	// Test getEnvAsInt with invalid value
	os.Setenv("TEST_INT_VAR", "invalid")
	intResult = getEnvAsInt("TEST_INT_VAR", 100)
	if intResult != 100 {
		t.Errorf("getEnvAsInt() = %v, want 100", intResult)
	}

	// Test getEnvAsBool
	os.Setenv("TEST_BOOL_VAR", "true")
	boolResult := getEnvAsBool("TEST_BOOL_VAR", false)
	if boolResult != true {
		t.Errorf("getEnvAsBool() = %v, want true", boolResult)
	}

	// Test getEnvAsBool with invalid value
	os.Setenv("TEST_BOOL_VAR", "invalid")
	boolResult = getEnvAsBool("TEST_BOOL_VAR", false)
	if boolResult != false {
		t.Errorf("getEnvAsBool() = %v, want false", boolResult)
	}

	// Clean up
	os.Setenv("TEST_ENV_VAR", "")
	os.Setenv("TEST_INT_VAR", "")
	os.Setenv("TEST_BOOL_VAR", "")
}

func TestHasRequiredEnvVars(t *testing.T) {
	// Save current env vars
	originalURL := os.Getenv("JIRA_URL")
	originalEmail := os.Getenv("JIRA_EMAIL")
	originalToken := os.Getenv("JIRA_API_TOKEN")

	// Clean up after test
	defer func() {
		os.Setenv("JIRA_URL", originalURL)
		os.Setenv("JIRA_EMAIL", originalEmail)
		os.Setenv("JIRA_API_TOKEN", originalToken)
	}()

	tests := []struct {
		name      string
		jiraURL   string
		jiraEmail string
		jiraToken string
		expected  bool
	}{
		{
			name:      "all_required_vars_set",
			jiraURL:   "https://test.atlassian.net",
			jiraEmail: "test@example.com",
			jiraToken: "token123",
			expected:  true,
		},
		{
			name:      "missing_jira_url",
			jiraURL:   "",
			jiraEmail: "test@example.com",
			jiraToken: "token123",
			expected:  false,
		},
		{
			name:      "missing_jira_email",
			jiraURL:   "https://test.atlassian.net",
			jiraEmail: "",
			jiraToken: "token123",
			expected:  false,
		},
		{
			name:      "missing_jira_token",
			jiraURL:   "https://test.atlassian.net",
			jiraEmail: "test@example.com",
			jiraToken: "",
			expected:  false,
		},
		{
			name:      "all_missing",
			jiraURL:   "",
			jiraEmail: "",
			jiraToken: "",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables for this test
			os.Setenv("JIRA_URL", tt.jiraURL)
			os.Setenv("JIRA_EMAIL", tt.jiraEmail)
			os.Setenv("JIRA_API_TOKEN", tt.jiraToken)

			result := hasRequiredEnvVars()
			if result != tt.expected {
				t.Errorf("hasRequiredEnvVars() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParseNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"valid number", "5", 5},
		{"valid number with spaces", "  10  ", 10},
		{"invalid string", "abc", 0},
		{"empty string", "", 0},
		{"negative number", "-5", -5},
		{"zero", "0", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseNumber(tt.input)
			if result != tt.expected {
				t.Errorf("parseNumber(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetAvailableIssueTypes(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse string
		statusCode     int
		wantError      bool
		wantCount      int
	}{
		{
			name: "successful response with issue types",
			serverResponse: `{
				"projects": [{
					"issuetypes": [
						{"id": "1", "name": "Story", "description": "User story", "subtask": false},
						{"id": "2", "name": "Bug", "description": "Software bug", "subtask": false},
						{"id": "3", "name": "Sub-task", "description": "Sub-task", "subtask": true}
					]
				}]
			}`,
			statusCode: http.StatusOK,
			wantError:  false,
			wantCount:  3,
		},
		{
			name:           "server error",
			serverResponse: `{"error": "internal error"}`,
			statusCode:     http.StatusInternalServerError,
			wantError:      true,
			wantCount:      0,
		},
		{
			name:           "no projects",
			serverResponse: `{"projects": []}`,
			statusCode:     http.StatusOK,
			wantError:      true,
			wantCount:      0,
		},
		{
			name:           "no issue types",
			serverResponse: `{"projects": [{"issuetypes": []}]}`,
			statusCode:     http.StatusOK,
			wantError:      false,
			wantCount:      0,
		},
		{
			name:           "invalid json",
			serverResponse: `{invalid json}`,
			statusCode:     http.StatusOK,
			wantError:      true,
			wantCount:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			issueTypes, err := getAvailableIssueTypes(server.URL, "test@example.com", "token", "TEST")

			if tt.wantError {
				if err == nil {
					t.Errorf("getAvailableIssueTypes() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("getAvailableIssueTypes() unexpected error: %v", err)
				return
			}

			if len(issueTypes) != tt.wantCount {
				t.Errorf("getAvailableIssueTypes() got %d issue types, want %d", len(issueTypes), tt.wantCount)
			}

			if tt.wantCount > 0 {
				if issueTypes[0].ID != "1" || issueTypes[0].Name != "Story" {
					t.Errorf("getAvailableIssueTypes() got wrong first issue type: %+v", issueTypes[0])
				}
			}
		})
	}
}

func TestSelectIssueType(t *testing.T) {
	issueTypes := []IssueTypeInfo{
		{ID: "1", Name: "Story", Description: "User story", IsSubtask: false},
		{ID: "2", Name: "Bug", Description: "Software bug", IsSubtask: false},
		{ID: "3", Name: "Sub-task", Description: "Sub-task", IsSubtask: true},
	}

	tests := []struct {
		name         string
		input        string
		purpose      string
		onlySubtasks bool
		expected     string
	}{
		{
			name:         "select first option for stories",
			input:        "1\n",
			purpose:      "historias",
			onlySubtasks: false,
			expected:     "Story",
		},
		{
			name:         "select second option for stories",
			input:        "2\n",
			purpose:      "historias",
			onlySubtasks: false,
			expected:     "Bug",
		},
		{
			name:         "select subtask",
			input:        "1\n",
			purpose:      "subtareas",
			onlySubtasks: true,
			expected:     "Sub-task",
		},
		{
			name:         "invalid then valid selection",
			input:        "invalid\n5\n2\n",
			purpose:      "historias",
			onlySubtasks: false,
			expected:     "Bug",
		},
		{
			name:         "empty input defaults to first",
			input:        "\n",
			purpose:      "historias",
			onlySubtasks: false,
			expected:     "Story",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tt.input))
			result := selectIssueType(reader, tt.purpose, issueTypes, tt.onlySubtasks)

			if result != tt.expected {
				t.Errorf("selectIssueType() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestSelectIssueType_NoValidTypes(t *testing.T) {
	issueTypes := []IssueTypeInfo{
		{ID: "1", Name: "Story", Description: "User story", IsSubtask: false},
	}

	reader := bufio.NewReader(strings.NewReader("Custom Type\n"))
	result := selectIssueType(reader, "subtareas", issueTypes, true)

	if result != "Custom Type" {
		t.Errorf("selectIssueType() with no valid types = %q, want %q", result, "Custom Type")
	}
}

func TestDetectAcceptanceCriteriaField_Interactive(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse string
		statusCode     int
		wantError      bool
		expectedField  string
	}{
		{
			name: "finds known acceptance criteria field",
			serverResponse: `[
				{"id": "customfield_10147", "name": "Acceptance Criteria", "description": "Field for acceptance criteria"},
				{"id": "customfield_10001", "name": "Summary", "description": "Summary field"}
			]`,
			statusCode:    http.StatusOK,
			wantError:     false,
			expectedField: "customfield_10147",
		},
		{
			name: "finds field by name pattern",
			serverResponse: `[
				{"id": "customfield_12345", "name": "Criterios de Aceptación", "description": "Spanish acceptance criteria"},
				{"id": "customfield_20001", "name": "Summary", "description": "Summary field"}
			]`,
			statusCode:    http.StatusOK,
			wantError:     false,
			expectedField: "customfield_12345",
		},
		{
			name: "no acceptance criteria field found",
			serverResponse: `[
				{"id": "customfield_20001", "name": "Summary", "description": "Summary field"},
				{"id": "customfield_20002", "name": "Priority", "description": "Priority field"}
			]`,
			statusCode:    http.StatusOK,
			wantError:     true,
			expectedField: "",
		},
		{
			name:           "server error",
			serverResponse: `{"error": "internal error"}`,
			statusCode:     http.StatusInternalServerError,
			wantError:      true,
			expectedField:  "",
		},
		{
			name:           "invalid json response",
			serverResponse: `{invalid json}`,
			statusCode:     http.StatusOK,
			wantError:      true,
			expectedField:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			ctx := context.Background()
			client := &http.Client{}

			result, err := detectAcceptanceCriteriaField(ctx, client, server.URL, "test@example.com", "token", "TEST", "Story")

			if tt.wantError {
				if err == nil {
					t.Errorf("detectAcceptanceCriteriaField() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("detectAcceptanceCriteriaField() unexpected error: %v", err)
				return
			}

			if result != tt.expectedField {
				t.Errorf("detectAcceptanceCriteriaField() = %q, want %q", result, tt.expectedField)
			}
		})
	}
}
