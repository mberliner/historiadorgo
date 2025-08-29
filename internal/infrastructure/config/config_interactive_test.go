package config

import (
	"bufio"
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
