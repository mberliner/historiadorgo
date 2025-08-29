package config

import (
	"os"
	"strings"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		wantError   bool
		wantJiraURL string
	}{
		{
			name: "valid config",
			envVars: map[string]string{
				"JIRA_URL":       "https://test.atlassian.net",
				"JIRA_EMAIL":     "test@example.com",
				"JIRA_API_TOKEN": "test-token",
				"PROJECT_KEY":    "TEST",
			},
			wantError:   false,
			wantJiraURL: "https://test.atlassian.net",
		},
		{
			name: "missing required env vars",
			envVars: map[string]string{
				"PROJECT_KEY": "TEST",
			},
			wantError: true,
		},
		{
			name: "missing jira url",
			envVars: map[string]string{
				"JIRA_EMAIL":     "test@example.com",
				"JIRA_API_TOKEN": "test-token",
			},
			wantError: true,
		},
		{
			name: "missing jira email",
			envVars: map[string]string{
				"JIRA_URL":       "https://test.atlassian.net",
				"JIRA_API_TOKEN": "test-token",
			},
			wantError: true,
		},
		{
			name: "missing jira token",
			envVars: map[string]string{
				"JIRA_URL":   "https://test.atlassian.net",
				"JIRA_EMAIL": "test@example.com",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			clearEnv()

			// Set test environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			config, err := LoadConfig()

			if tt.wantError && err == nil {
				t.Errorf("LoadConfig() error = nil, wantError = true")
				return
			}
			if !tt.wantError && err != nil {
				t.Errorf("LoadConfig() error = %v, wantError = false", err)
				return
			}

			if !tt.wantError && config != nil {
				if config.JiraURL != tt.wantJiraURL {
					t.Errorf("JiraURL = %v, want %v", config.JiraURL, tt.wantJiraURL)
				}
			}

			// Clean up
			clearEnv()
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		wantError bool
	}{
		{
			name: "valid config",
			config: &Config{
				JiraURL:      "https://test.atlassian.net",
				JiraEmail:    "test@example.com",
				JiraAPIToken: "test-token",
			},
			wantError: false,
		},
		{
			name: "missing jira url",
			config: &Config{
				JiraEmail:    "test@example.com",
				JiraAPIToken: "test-token",
			},
			wantError: true,
		},
		{
			name: "missing jira email",
			config: &Config{
				JiraURL:      "https://test.atlassian.net",
				JiraAPIToken: "test-token",
			},
			wantError: true,
		},
		{
			name: "missing jira token",
			config: &Config{
				JiraURL:   "https://test.atlassian.net",
				JiraEmail: "test@example.com",
			},
			wantError: true,
		},
		{
			name:      "all missing",
			config:    &Config{},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.wantError && err == nil {
				t.Errorf("Validate() error = nil, wantError = true")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Validate() error = %v, wantError = false", err)
			}
		})
	}
}

func TestGetEnvHelpers(t *testing.T) {
	// Test getEnv
	os.Setenv("TEST_STRING", "test_value")
	if got := getEnv("TEST_STRING", "default"); got != "test_value" {
		t.Errorf("getEnv() = %v, want test_value", got)
	}

	if got := getEnv("NON_EXISTENT", "default"); got != "default" {
		t.Errorf("getEnv() = %v, want default", got)
	}

	// Test getEnvAsInt
	os.Setenv("TEST_INT", "42")
	if got := getEnvAsInt("TEST_INT", 0); got != 42 {
		t.Errorf("getEnvAsInt() = %v, want 42", got)
	}

	if got := getEnvAsInt("NON_EXISTENT_INT", 10); got != 10 {
		t.Errorf("getEnvAsInt() = %v, want 10", got)
	}

	os.Setenv("INVALID_INT", "not_a_number")
	if got := getEnvAsInt("INVALID_INT", 5); got != 5 {
		t.Errorf("getEnvAsInt() with invalid int = %v, want 5", got)
	}

	// Test getEnvAsBool
	os.Setenv("TEST_BOOL_TRUE", "true")
	if got := getEnvAsBool("TEST_BOOL_TRUE", false); got != true {
		t.Errorf("getEnvAsBool() = %v, want true", got)
	}

	os.Setenv("TEST_BOOL_FALSE", "false")
	if got := getEnvAsBool("TEST_BOOL_FALSE", true); got != false {
		t.Errorf("getEnvAsBool() = %v, want false", got)
	}

	if got := getEnvAsBool("NON_EXISTENT_BOOL", true); got != true {
		t.Errorf("getEnvAsBool() = %v, want true", got)
	}

	os.Setenv("INVALID_BOOL", "not_a_bool")
	if got := getEnvAsBool("INVALID_BOOL", false); got != false {
		t.Errorf("getEnvAsBool() with invalid bool = %v, want false", got)
	}

	// Clean up
	clearEnv()
}

func TestDefaultValues(t *testing.T) {
	// Clear environment and test defaults
	clearEnv()

	// Set only required values
	os.Setenv("JIRA_URL", "https://test.atlassian.net")
	os.Setenv("JIRA_EMAIL", "test@example.com")
	os.Setenv("JIRA_API_TOKEN", "test-token")

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	// Test default values
	if config.DefaultIssueType != "Story" {
		t.Errorf("DefaultIssueType = %v, want Story", config.DefaultIssueType)
	}
	if config.SubtaskIssueType != "Sub-task" {
		t.Errorf("SubtaskIssueType = %v, want Sub-task", config.SubtaskIssueType)
	}
	if config.FeatureIssueType != "Feature" {
		t.Errorf("FeatureIssueType = %v, want Feature", config.FeatureIssueType)
	}
	if config.BatchSize != 10 {
		t.Errorf("BatchSize = %v, want 10", config.BatchSize)
	}
	if config.DryRun != false {
		t.Errorf("DryRun = %v, want false", config.DryRun)
	}
	if config.InputDirectory != "entrada" {
		t.Errorf("InputDirectory = %v, want entrada", config.InputDirectory)
	}
	if config.LogsDirectory != "logs" {
		t.Errorf("LogsDirectory = %v, want logs", config.LogsDirectory)
	}
	if config.ProcessedDirectory != "procesados" {
		t.Errorf("ProcessedDirectory = %v, want procesados", config.ProcessedDirectory)
	}
	if config.RollbackOnSubtaskFailure != false {
		t.Errorf("RollbackOnSubtaskFailure = %v, want false", config.RollbackOnSubtaskFailure)
	}

	clearEnv()
}

func TestCreateInteractiveEnvFile_InputValidation(t *testing.T) {
	// Clean up any existing .env file
	defer os.Remove(".env")

	tests := []struct {
		name        string
		inputLines  []string
		wantError   bool
		errorString string
	}{
		{
			name:        "missing_jira_url",
			inputLines:  []string{"", "test@example.com", "token123"},
			wantError:   true,
			errorString: "JIRA_URL es requerido",
		},
		{
			name:        "missing_jira_email",
			inputLines:  []string{"https://test.atlassian.net", "", "token123"},
			wantError:   true,
			errorString: "JIRA_EMAIL es requerido",
		},
		{
			name:        "missing_jira_token",
			inputLines:  []string{"https://test.atlassian.net", "test@example.com", ""},
			wantError:   true,
			errorString: "JIRA_API_TOKEN es requerido",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This function would need significant refactoring to be unit testable
			// as it directly reads from stdin. For now, we test the validation logic
			// The CreateInteractiveEnvFile function should be refactored to accept
			// a reader interface for better testability
			t.Skip("Function needs refactoring for better testability")
		})
	}
}

func TestLoadConfig_EnvFileScenarios(t *testing.T) {
	// Test scenarios where .env file exists vs doesn't exist
	tests := []struct {
		name           string
		envFileExists  bool
		envFileContent string
		envVars        map[string]string
		wantError      bool
		wantJiraURL    string
	}{
		{
			name:          "env_file_exists_valid",
			envFileExists: true,
			envFileContent: `JIRA_URL=https://file.atlassian.net
JIRA_EMAIL=file@example.com
JIRA_API_TOKEN=file-token`,
			wantError:   false,
			wantJiraURL: "https://file.atlassian.net",
		},
		{
			name:           "env_file_exists_but_invalid",
			envFileExists:  true,
			envFileContent: `JIRA_URL=https://file.atlassian.net`,
			wantError:      true,
		},
		{
			name:          "no_env_file_but_env_vars_set",
			envFileExists: false,
			envVars: map[string]string{
				"JIRA_URL":       "https://env.atlassian.net",
				"JIRA_EMAIL":     "env@example.com",
				"JIRA_API_TOKEN": "env-token",
			},
			wantError:   false,
			wantJiraURL: "https://env.atlassian.net",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up
			clearEnv()
			defer os.Remove(".env")

			// Set up env file if needed
			if tt.envFileExists {
				if err := os.WriteFile(".env", []byte(tt.envFileContent), 0644); err != nil {
					t.Fatalf("Failed to create test .env file: %v", err)
				}
			}

			// Set environment variables if provided
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			config, err := LoadConfig()

			if tt.wantError && err == nil {
				t.Errorf("LoadConfig() error = nil, wantError = true")
				return
			}
			if !tt.wantError && err != nil {
				t.Errorf("LoadConfig() error = %v, wantError = false", err)
				return
			}

			if !tt.wantError && config != nil {
				if config.JiraURL != tt.wantJiraURL {
					t.Errorf("JiraURL = %v, want %v", config.JiraURL, tt.wantJiraURL)
				}
			}

			clearEnv()
		})
	}
}

func TestLoadConfig_EnvOverrides(t *testing.T) {
	// Test that environment variables can override config values
	tests := []struct {
		name          string
		envVars       map[string]string
		wantBatchSize int
		wantDryRun    bool
		wantInputDir  string
	}{
		{
			name: "custom_batch_size_and_dry_run",
			envVars: map[string]string{
				"JIRA_URL":        "https://test.atlassian.net",
				"JIRA_EMAIL":      "test@example.com",
				"JIRA_API_TOKEN":  "test-token",
				"BATCH_SIZE":      "25",
				"DRY_RUN":         "true",
				"INPUT_DIRECTORY": "custom_input",
			},
			wantBatchSize: 25,
			wantDryRun:    true,
			wantInputDir:  "custom_input",
		},
		{
			name: "invalid_batch_size_uses_default",
			envVars: map[string]string{
				"JIRA_URL":       "https://test.atlassian.net",
				"JIRA_EMAIL":     "test@example.com",
				"JIRA_API_TOKEN": "test-token",
				"BATCH_SIZE":     "invalid",
				"DRY_RUN":        "invalid",
			},
			wantBatchSize: 10,        // default
			wantDryRun:    false,     // default
			wantInputDir:  "entrada", // default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnv()
			defer clearEnv()

			// Set test environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			config, err := LoadConfig()
			if err != nil {
				t.Fatalf("LoadConfig() error = %v", err)
			}

			if config.BatchSize != tt.wantBatchSize {
				t.Errorf("BatchSize = %v, want %v", config.BatchSize, tt.wantBatchSize)
			}
			if config.DryRun != tt.wantDryRun {
				t.Errorf("DryRun = %v, want %v", config.DryRun, tt.wantDryRun)
			}
			if config.InputDirectory != tt.wantInputDir {
				t.Errorf("InputDirectory = %v, want %v", config.InputDirectory, tt.wantInputDir)
			}
		})
	}
}

func TestConfig_ValidateExtended(t *testing.T) {
	// Test additional validation scenarios
	tests := []struct {
		name          string
		config        *Config
		wantError     bool
		errorContains string
	}{
		{
			name: "valid_config_with_optional_fields",
			config: &Config{
				JiraURL:                  "https://test.atlassian.net",
				JiraEmail:                "test@example.com",
				JiraAPIToken:             "test-token",
				ProjectKey:               "TEST",
				DefaultIssueType:         "Story",
				SubtaskIssueType:         "Sub-task",
				FeatureIssueType:         "Epic",
				BatchSize:                10,
				DryRun:                   false,
				AcceptanceCriteriaField:  "customfield_10001",
				InputDirectory:           "entrada",
				LogsDirectory:            "logs",
				ProcessedDirectory:       "procesados",
				RollbackOnSubtaskFailure: true,
				FeatureRequiredFields:    "summary,description",
			},
			wantError: false,
		},
		{
			name: "missing_multiple_required_fields",
			config: &Config{
				JiraAPIToken: "test-token",
			},
			wantError:     true,
			errorContains: "JIRA_URL, JIRA_EMAIL",
		},
		{
			name: "empty_strings_treated_as_missing",
			config: &Config{
				JiraURL:      "",
				JiraEmail:    "",
				JiraAPIToken: "",
			},
			wantError:     true,
			errorContains: "JIRA_URL, JIRA_EMAIL, JIRA_API_TOKEN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.wantError && err == nil {
				t.Errorf("Validate() error = nil, wantError = true")
				return
			}
			if !tt.wantError && err != nil {
				t.Errorf("Validate() error = %v, wantError = false", err)
				return
			}

			if tt.wantError && err != nil && tt.errorContains != "" {
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Validate() error = %v, should contain %v", err.Error(), tt.errorContains)
				}
			}
		})
	}
}

func TestHasRequiredEnvVars_Coverage(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected bool
	}{
		{
			name: "all_vars_present",
			envVars: map[string]string{
				"JIRA_URL":       "https://test.atlassian.net",
				"JIRA_EMAIL":     "test@example.com",
				"JIRA_API_TOKEN": "test-token",
			},
			expected: true,
		},
		{
			name: "missing_first_var",
			envVars: map[string]string{
				"JIRA_EMAIL":     "test@example.com",
				"JIRA_API_TOKEN": "test-token",
			},
			expected: false,
		},
		{
			name: "missing_middle_var",
			envVars: map[string]string{
				"JIRA_URL":       "https://test.atlassian.net",
				"JIRA_API_TOKEN": "test-token",
			},
			expected: false,
		},
		{
			name: "missing_last_var",
			envVars: map[string]string{
				"JIRA_URL":   "https://test.atlassian.net",
				"JIRA_EMAIL": "test@example.com",
			},
			expected: false,
		},
		{
			name:     "all_vars_missing",
			envVars:  map[string]string{},
			expected: false,
		},
		{
			name: "empty_string_values",
			envVars: map[string]string{
				"JIRA_URL":       "",
				"JIRA_EMAIL":     "test@example.com",
				"JIRA_API_TOKEN": "test-token",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnv()
			defer clearEnv()

			// Set test environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			result := hasRequiredEnvVars()
			if result != tt.expected {
				t.Errorf("hasRequiredEnvVars() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Helper function to clear environment variables used in tests
func clearEnv() {
	envVars := []string{
		"JIRA_URL", "JIRA_EMAIL", "JIRA_API_TOKEN", "PROJECT_KEY",
		"DEFAULT_ISSUE_TYPE", "SUBTASK_ISSUE_TYPE", "FEATURE_ISSUE_TYPE",
		"BATCH_SIZE", "DRY_RUN", "ACCEPTANCE_CRITERIA_FIELD",
		"INPUT_DIRECTORY", "LOGS_DIRECTORY", "PROCESSED_DIRECTORY",
		"ROLLBACK_ON_SUBTASK_FAILURE", "FEATURE_REQUIRED_FIELDS",
		"TEST_STRING", "TEST_INT", "INVALID_INT", "TEST_BOOL_TRUE",
		"TEST_BOOL_FALSE", "INVALID_BOOL",
	}

	for _, envVar := range envVars {
		os.Unsetenv(envVar)
	}
}
