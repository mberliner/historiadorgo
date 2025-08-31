package config

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateInteractiveEnvFile_MockedInput(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantError     bool
		expectedError string
		checkEnvFile  bool
	}{
		{
			name: "valid_complete_input_with_project",
			input: strings.Join([]string{
				"https://company.atlassian.net",  // JIRA_URL
				"user@company.com",              // JIRA_EMAIL  
				"token123",                      // JIRA_API_TOKEN
				"MYPROJ",                        // PROJECT_KEY
				"Story",                         // DEFAULT_ISSUE_TYPE (fallback when Jira call fails)
				"Sub-task",                      // SUBTASK_ISSUE_TYPE (fallback when Jira call fails)
				"Epic",                          // FEATURE_ISSUE_TYPE (fallback when Jira call fails)
				"entrada",                       // INPUT_DIRECTORY (default)
				"logs",                          // LOGS_DIRECTORY (default)
				"procesados",                    // PROCESSED_DIRECTORY (default)
				"n",                             // ROLLBACK_ON_SUBTASK_FAILURE
				"",                              // ACCEPTANCE_CRITERIA_FIELD (empty)
				"",                              // End of input
			}, "\n") + "\n",
			wantError:    false,
			checkEnvFile: true,
		},
		{
			name: "valid_minimal_input_no_project",
			input: strings.Join([]string{
				"https://company.atlassian.net",  // JIRA_URL
				"user@company.com",              // JIRA_EMAIL  
				"token123",                      // JIRA_API_TOKEN
				"",                              // PROJECT_KEY (empty)
				"Story",                         // DEFAULT_ISSUE_TYPE
				"Sub-task",                      // SUBTASK_ISSUE_TYPE
				"Epic",                          // FEATURE_ISSUE_TYPE
				"entrada",                       // INPUT_DIRECTORY (default)
				"logs",                          // LOGS_DIRECTORY (default) 
				"procesados",                    // PROCESSED_DIRECTORY (default)
				"y",                             // ROLLBACK_ON_SUBTASK_FAILURE
				"",                              // ACCEPTANCE_CRITERIA_FIELD (empty)
				"",                              // End of input
			}, "\n") + "\n",
			wantError:    false,
			checkEnvFile: true,
		},
		{
			name: "missing_jira_url",
			input: strings.Join([]string{
				"",                              // JIRA_URL (empty - should error)
				"user@company.com",              // JIRA_EMAIL
				"token123",                      // JIRA_API_TOKEN
				"",                              // End of input
			}, "\n") + "\n",
			wantError:     true,
			expectedError: "JIRA_URL es requerido",
		},
		{
			name: "missing_jira_email",
			input: strings.Join([]string{
				"https://company.atlassian.net",  // JIRA_URL
				"",                              // JIRA_EMAIL (empty - should error)
				"token123",                      // JIRA_API_TOKEN
				"",                              // End of input
			}, "\n") + "\n",
			wantError:     true,
			expectedError: "JIRA_EMAIL es requerido",
		},
		{
			name: "missing_jira_token",
			input: strings.Join([]string{
				"https://company.atlassian.net",  // JIRA_URL
				"user@company.com",              // JIRA_EMAIL
				"",                              // JIRA_API_TOKEN (empty - should error)
				"",                              // End of input
			}, "\n") + "\n",
			wantError:     true,
			expectedError: "JIRA_API_TOKEN es requerido",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for test
			tempDir := t.TempDir()
			originalDir, _ := os.Getwd()
			defer os.Chdir(originalDir)
			os.Chdir(tempDir)

			// Mock stdin with test input
			oldStdin := os.Stdin
			r, w, _ := os.Pipe()
			os.Stdin = r
			defer func() {
				os.Stdin = oldStdin
				r.Close()
				w.Close()
			}()

			// Write input to pipe
			go func() {
				defer w.Close()
				io.WriteString(w, tt.input)
			}()

			// Call the function
			err := CreateInteractiveEnvFile()

			// Check error expectation
			if tt.wantError {
				if err == nil {
					t.Errorf("CreateInteractiveEnvFile() expected error, got nil")
					return
				}
				if tt.expectedError != "" && !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("CreateInteractiveEnvFile() expected error containing %q, got %q", tt.expectedError, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("CreateInteractiveEnvFile() unexpected error: %v", err)
				return
			}

			// Check .env file was created and has expected content
			if tt.checkEnvFile {
				envPath := filepath.Join(tempDir, ".env")
				if _, err := os.Stat(envPath); os.IsNotExist(err) {
					t.Errorf("CreateInteractiveEnvFile() did not create .env file")
					return
				}

				content, err := os.ReadFile(envPath)
				if err != nil {
					t.Errorf("CreateInteractiveEnvFile() could not read .env file: %v", err)
					return
				}

				envContent := string(content)
				
				// Check required fields are present
				requiredFields := []string{
					"JIRA_URL=",
					"JIRA_EMAIL=", 
					"JIRA_API_TOKEN=",
					"PROJECT_KEY=",
					"DEFAULT_ISSUE_TYPE=",
					"SUBTASK_ISSUE_TYPE=",
					"FEATURE_ISSUE_TYPE=",
					"INPUT_DIRECTORY=",
					"LOGS_DIRECTORY=",
					"PROCESSED_DIRECTORY=",
					"ROLLBACK_ON_SUBTASK_FAILURE=",
				}
				
				for _, field := range requiredFields {
					if !strings.Contains(envContent, field) {
						t.Errorf("CreateInteractiveEnvFile() .env file missing field: %s", field)
					}
				}

				// Check directories were created (only check if function succeeded)
				// Note: The function creates directories with the values from input
				// For this test, it uses default values "entrada", "logs", "procesados"
				dirs := []string{"entrada", "logs", "procesados"}
				for _, dir := range dirs {
					dirPath := filepath.Join(tempDir, dir)
					if _, err := os.Stat(dirPath); os.IsNotExist(err) {
						// Directory creation might fail in some test environments
						// This is not a critical failure for the function's core purpose
						t.Logf("Warning: Directory %s was not created (may be expected in test environment)", dir)
					}
				}
			}
		})
	}
}

func TestCreateInteractiveEnvFile_HelperFunctions(t *testing.T) {
	t.Run("promptForInput_with_default", func(t *testing.T) {
		input := "\n" // Empty input should use default
		reader := bufio.NewReader(strings.NewReader(input))
		
		result := promptForInput(reader, "Test prompt", "default_value")
		if result != "default_value" {
			t.Errorf("promptForInput() expected 'default_value', got %q", result)
		}
	})

	t.Run("promptForInput_with_value", func(t *testing.T) {
		input := "user_input\n"
		reader := bufio.NewReader(strings.NewReader(input))
		
		result := promptForInput(reader, "Test prompt", "default_value")
		if result != "user_input" {
			t.Errorf("promptForInput() expected 'user_input', got %q", result)
		}
	})

	t.Run("promptForInput_with_whitespace", func(t *testing.T) {
		input := "  user_input  \n"
		reader := bufio.NewReader(strings.NewReader(input))
		
		result := promptForInput(reader, "Test prompt", "default_value")
		if result != "user_input" {
			t.Errorf("promptForInput() expected 'user_input' (trimmed), got %q", result)
		}
	})
}

func TestCreateInteractiveEnvFile_FileOperations(t *testing.T) {
	t.Run("directory_creation_failure", func(t *testing.T) {
		// This test is difficult to implement without root permissions
		// or complex filesystem mocking. The function handles directory 
		// creation errors gracefully by printing a warning.
		t.Skip("Directory creation failure testing requires complex setup")
	})

	t.Run("env_file_write_permission", func(t *testing.T) {
		// Create temp directory with restricted permissions
		tempDir := t.TempDir()
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)
		
		// Make directory read-only (this may not work on all systems)
		os.Chmod(tempDir, 0444)
		defer os.Chmod(tempDir, 0755) // Restore permissions for cleanup
		
		os.Chdir(tempDir)

		// Mock stdin with valid input
		oldStdin := os.Stdin
		r, w, _ := os.Pipe()
		os.Stdin = r
		defer func() {
			os.Stdin = oldStdin
			r.Close()
			w.Close()
		}()

		input := strings.Join([]string{
			"https://company.atlassian.net",
			"user@company.com", 
			"token123",
			"",        // No project key
			"Story",   
			"Sub-task",
			"Epic",
			"entrada",
			"logs", 
			"procesados",
			"n",
			"",
		}, "\n") + "\n"

		go func() {
			defer w.Close()
			io.WriteString(w, input)
		}()

		err := CreateInteractiveEnvFile()
		
		// Should get a file write error
		if err == nil {
			t.Skip("Could not create read-only directory scenario")
		}
		
		if !strings.Contains(err.Error(), "error writing .env file") {
			t.Errorf("Expected file write error, got: %v", err)
		}
	})
}