package jira

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"historiadorgo/internal/infrastructure/config"
)

func createTestFeatureManager() (*FeatureManager, *httptest.Server) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Default handler - puede ser sobrescrito en tests específicos
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"key": "TEST-123"}`))
	}))

	cfg := &config.Config{
		JiraURL:          server.URL,
		JiraEmail:        "test@example.com",
		JiraAPIToken:     "test-token",
		FeatureIssueType: "Feature",
	}

	jiraClient := NewJiraClient(cfg)
	featureManager := NewFeatureManager(jiraClient, cfg)

	return featureManager, server
}

func TestNewFeatureManager(t *testing.T) {
	cfg := createTestConfig()
	jiraClient := NewJiraClient(cfg)
	fm := NewFeatureManager(jiraClient, cfg)

	if fm == nil {
		t.Fatal("Expected FeatureManager to be created")
	}

	if fm.jiraClient != jiraClient {
		t.Error("Expected jiraClient to be set")
	}

	if fm.config != cfg {
		t.Error("Expected config to be set")
	}
}

func TestFeatureManager_CreateOrGetFeature_JiraKey(t *testing.T) {
	fm, server := createTestFeatureManager()
	defer server.Close()

	// Mock para validar parent issue
	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/rest/api/3/issue/TEST-123") {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"key": "TEST-123"}`))
		}
	})

	result, err := fm.CreateOrGetFeature(context.Background(), "TEST-123", "PROJ")

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result.ExistingKey == "" {
		t.Error("Expected result to have existing key")
	}

	if result.IssueKey != "TEST-123" {
		t.Errorf("Expected key to be TEST-123, got %s", result.IssueKey)
	}
}

func TestFeatureManager_CreateOrGetFeature_InvalidJiraKey(t *testing.T) {
	fm, server := createTestFeatureManager()
	defer server.Close()

	// Mock para retornar 404 cuando se valida el parent issue
	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/rest/api/3/issue/TEST-999") {
			w.WriteHeader(http.StatusNotFound)
		}
	})

	result, err := fm.CreateOrGetFeature(context.Background(), "TEST-999", "PROJ")

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result.Success {
		t.Error("Expected result to not be successful")
	}

	if !strings.Contains(result.ErrorMessage, "Parent issue validation failed") {
		t.Errorf("Expected validation error, got: %s", result.ErrorMessage)
	}
}

func TestFeatureManager_CreateOrGetFeature_ExistingFeature(t *testing.T) {
	fm, server := createTestFeatureManager()
	defer server.Close()

	description := "Test Feature Description"

	// Mock para búsqueda que encuentra feature existente
	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/rest/api/3/search") {
			response := JiraSearchResponse{
				Issues: []JiraIssue{
					{
						Key: "PROJ-456",
						Fields: map[string]interface{}{
							"summary": "Test Feature Description",
						},
					},
				},
				Total: 1,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}
	})

	result, err := fm.CreateOrGetFeature(context.Background(), description, "PROJ")

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result.ExistingKey == "" {
		t.Error("Expected result to have existing key")
	}

	if result.IssueKey != "PROJ-456" {
		t.Errorf("Expected key to be PROJ-456, got %s", result.IssueKey)
	}
}

func TestFeatureManager_CreateOrGetFeature_CreateNew(t *testing.T) {
	fm, server := createTestFeatureManager()
	defer server.Close()

	description := "New Feature Description"

	// Mock para crear nuevo feature
	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/rest/api/3/search") {
			// No features found
			response := JiraSearchResponse{
				Issues: []JiraIssue{},
				Total:  0,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		} else if strings.Contains(r.URL.Path, "/rest/api/3/issue") && r.Method == "POST" {
			// Create new issue
			w.WriteHeader(http.StatusCreated)
			response := JiraCreateResponse{
				ID:   "10001",
				Key:  "PROJ-789",
				Self: server.URL + "/rest/api/3/issue/10001",
			}
			json.NewEncoder(w).Encode(response)
		}
	})

	result, err := fm.CreateOrGetFeature(context.Background(), description, "PROJ")

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected result to be successful, got error: %s", result.ErrorMessage)
	}

	if result.IssueKey != "PROJ-789" {
		t.Errorf("Expected key to be PROJ-789, got %s", result.IssueKey)
	}

	if !result.WasCreated {
		t.Error("Expected feature to be marked as created")
	}
}

func TestFeatureManager_SearchExistingFeature(t *testing.T) {
	tests := []struct {
		name         string
		description  string
		mockResponse JiraSearchResponse
		expectedKey  string
		expectError  bool
	}{
		{
			name:        "feature_found",
			description: "Test Feature",
			mockResponse: JiraSearchResponse{
				Issues: []JiraIssue{
					{
						Key: "PROJ-123",
						Fields: map[string]interface{}{
							"summary": "Test Feature",
						},
					},
				},
				Total: 1,
			},
			expectedKey: "PROJ-123",
			expectError: false,
		},
		{
			name:        "no_features_found",
			description: "Nonexistent Feature",
			mockResponse: JiraSearchResponse{
				Issues: []JiraIssue{},
				Total:  0,
			},
			expectedKey: "",
			expectError: false,
		},
		{
			name:        "similar_feature_found",
			description: "Test System Feature",
			mockResponse: JiraSearchResponse{
				Issues: []JiraIssue{
					{
						Key: "PROJ-456",
						Fields: map[string]interface{}{
							"summary": "Test System Feature Implementation",
						},
					},
				},
				Total: 1,
			},
			expectedKey: "PROJ-456", // Debería encontrarlo por similaridad
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create individual server for each subtest to avoid race conditions
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.Contains(r.URL.Path, "/rest/api/3/search") {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(tt.mockResponse)
				}
			}))
			defer server.Close()

			cfg := &config.Config{
				JiraURL:          server.URL,
				JiraEmail:        "test@example.com",
				JiraAPIToken:     "test-token",
				FeatureIssueType: "Feature",
			}

			jiraClient := NewJiraClient(cfg)
			fm := NewFeatureManager(jiraClient, cfg)

			key, err := fm.SearchExistingFeature(context.Background(), tt.description, "PROJ")

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				if key != tt.expectedKey {
					t.Errorf("Expected key '%s', got '%s'", tt.expectedKey, key)
				}
			}
		})
	}
}

func TestFeatureManager_ValidateFeatureRequiredFields(t *testing.T) {
	fm, server := createTestFeatureManager()
	defer server.Close()

	tests := []struct {
		name           string
		mockResponse   map[string]interface{}
		statusCode     int
		expectedFields int
		expectError    bool
	}{
		{
			name: "valid_metadata_with_required_fields",
			mockResponse: map[string]interface{}{
				"projects": []interface{}{
					map[string]interface{}{
						"issuetypes": []interface{}{
							map[string]interface{}{
								"name": "Feature",
								"fields": map[string]interface{}{
									"customfield_10001": map[string]interface{}{
										"required": true,
										"name":     "Epic Link",
									},
									"priority": map[string]interface{}{
										"required": true,
										"name":     "Priority",
									},
									"project": map[string]interface{}{
										"required": true,
										"name":     "Project",
									},
									"summary": map[string]interface{}{
										"required": true,
										"name":     "Summary",
									},
								},
							},
						},
					},
				},
			},
			statusCode:     http.StatusOK,
			expectedFields: 2, // Epic Link y Priority (project y summary se excluyen)
			expectError:    false,
		},
		{
			name: "no_required_fields",
			mockResponse: map[string]interface{}{
				"projects": []interface{}{
					map[string]interface{}{
						"issuetypes": []interface{}{
							map[string]interface{}{
								"name": "Feature",
								"fields": map[string]interface{}{
									"description": map[string]interface{}{
										"required": false,
										"name":     "Description",
									},
								},
							},
						},
					},
				},
			},
			statusCode:     http.StatusOK,
			expectedFields: 0,
			expectError:    false,
		},
		{
			name:           "server_error",
			mockResponse:   nil,
			statusCode:     http.StatusInternalServerError,
			expectedFields: 0,
			expectError:    true,
		},
		{
			name: "malformed_json_response",
			mockResponse: map[string]interface{}{
				"projects": "invalid_json_structure",
			},
			statusCode:     http.StatusOK,
			expectedFields: 0,
			expectError:    true,
		},
		{
			name: "empty_projects_array",
			mockResponse: map[string]interface{}{
				"projects": []interface{}{},
			},
			statusCode:     http.StatusOK,
			expectedFields: 0,
			expectError:    true,
		},
		{
			name: "no_projects_field",
			mockResponse: map[string]interface{}{
				"other_field": "value",
			},
			statusCode:     http.StatusOK,
			expectedFields: 0,
			expectError:    true,
		},
		{
			name: "project_without_issuetypes",
			mockResponse: map[string]interface{}{
				"projects": []interface{}{
					map[string]interface{}{
						"key":  "PROJ",
						"name": "Test Project",
					},
				},
			},
			statusCode:     http.StatusOK,
			expectedFields: 0,
			expectError:    true,
		},
		{
			name: "feature_issuetype_not_found",
			mockResponse: map[string]interface{}{
				"projects": []interface{}{
					map[string]interface{}{
						"issuetypes": []interface{}{
							map[string]interface{}{
								"name": "Story",
								"fields": map[string]interface{}{
									"priority": map[string]interface{}{
										"required": true,
										"name":     "Priority",
									},
								},
							},
							map[string]interface{}{
								"name": "Bug",
								"fields": map[string]interface{}{
									"severity": map[string]interface{}{
										"required": true,
										"name":     "Severity",
									},
								},
							},
						},
					},
				},
			},
			statusCode:     http.StatusOK,
			expectedFields: 0,
			expectError:    true,
		},
		{
			name: "feature_issuetype_without_fields",
			mockResponse: map[string]interface{}{
				"projects": []interface{}{
					map[string]interface{}{
						"issuetypes": []interface{}{
							map[string]interface{}{
								"name": "Feature",
							},
						},
					},
				},
			},
			statusCode:     http.StatusOK,
			expectedFields: 0,
			expectError:    false,
		},
		{
			name: "fields_without_name",
			mockResponse: map[string]interface{}{
				"projects": []interface{}{
					map[string]interface{}{
						"issuetypes": []interface{}{
							map[string]interface{}{
								"name": "Feature",
								"fields": map[string]interface{}{
									"customfield_10001": map[string]interface{}{
										"required": true,
									},
									"customfield_10002": map[string]interface{}{
										"required": true,
										"name":     "Named Field",
									},
								},
							},
						},
					},
				},
			},
			statusCode:     http.StatusOK,
			expectedFields: 2, // uno sin name, otro con name
			expectError:    false,
		},
		{
			name: "mixed_required_and_optional_fields_excluding_defaults",
			mockResponse: map[string]interface{}{
				"projects": []interface{}{
					map[string]interface{}{
						"issuetypes": []interface{}{
							map[string]interface{}{
								"name": "Feature",
								"fields": map[string]interface{}{
									"customfield_10001": map[string]interface{}{
										"required": true,
										"name":     "Epic Link",
									},
									"customfield_10002": map[string]interface{}{
										"required": false,
										"name":     "Optional Field",
									},
									"project": map[string]interface{}{
										"required": true,
										"name":     "Project",
									},
									"issuetype": map[string]interface{}{
										"required": true,
										"name":     "Issue Type",
									},
									"summary": map[string]interface{}{
										"required": true,
										"name":     "Summary",
									},
									"description": map[string]interface{}{
										"required": true,
										"name":     "Description",
									},
									"customfield_10003": map[string]interface{}{
										"required": true,
										"name":     "Another Required Field",
									},
								},
							},
						},
					},
				},
			},
			statusCode:     http.StatusOK,
			expectedFields: 2, // Epic Link y Another Required Field (project, issuetype, summary, description se excluyen)
			expectError:    false,
		},
		{
			name: "field_data_not_map",
			mockResponse: map[string]interface{}{
				"projects": []interface{}{
					map[string]interface{}{
						"issuetypes": []interface{}{
							map[string]interface{}{
								"name": "Feature",
								"fields": map[string]interface{}{
									"customfield_10001": "invalid_field_data",
									"customfield_10002": map[string]interface{}{
										"required": true,
										"name":     "Valid Field",
									},
								},
							},
						},
					},
				},
			},
			statusCode:     http.StatusOK,
			expectedFields: 1, // solo el válido
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.Contains(r.URL.Path, "/rest/api/3/issue/createmeta") {
					w.WriteHeader(tt.statusCode)
					if tt.mockResponse != nil {
						w.Header().Set("Content-Type", "application/json")
						json.NewEncoder(w).Encode(tt.mockResponse)
					}
				}
			})

			fields, err := fm.ValidateFeatureRequiredFields(context.Background(), "PROJ")

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				if len(fields) != tt.expectedFields {
					t.Errorf("Expected %d fields, got %d", tt.expectedFields, len(fields))
				}
			}
		})
	}
}

func TestFeatureManager_ValidateFeatureRequiredFields_NetworkErrors(t *testing.T) {
	fm, server := createTestFeatureManager()
	defer server.Close()

	t.Run("invalid_json_response", func(t *testing.T) {
		server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.URL.Path, "/rest/api/3/issue/createmeta") {
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{invalid json`))
			}
		})

		fields, err := fm.ValidateFeatureRequiredFields(context.Background(), "PROJ")

		if err == nil {
			t.Error("Expected error for invalid JSON, got none")
		}
		if fields != nil {
			t.Error("Expected nil fields on error")
		}
	})

	t.Run("context_cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		fields, err := fm.ValidateFeatureRequiredFields(ctx, "PROJ")

		if err == nil {
			t.Error("Expected error for cancelled context, got none")
		}
		if fields != nil {
			t.Error("Expected nil fields on error")
		}
	})
}

func TestFeatureManager_ValidateFeatureRequiredFields_EdgeCases(t *testing.T) {
	fm, server := createTestFeatureManager()
	defer server.Close()

	tests := []struct {
		name           string
		mockResponse   map[string]interface{}
		expectedFields int
	}{
		{
			name: "required_field_with_no_bool_value",
			mockResponse: map[string]interface{}{
				"projects": []interface{}{
					map[string]interface{}{
						"issuetypes": []interface{}{
							map[string]interface{}{
								"name": "Feature",
								"fields": map[string]interface{}{
									"customfield_10001": map[string]interface{}{
										"required": "true", // string instead of bool
										"name":     "Should be ignored",
									},
									"customfield_10002": map[string]interface{}{
										"required": true,
										"name":     "Should be included",
									},
								},
							},
						},
					},
				},
			},
			expectedFields: 1, // solo el que tiene required: true (bool)
		},
		{
			name: "issue_type_name_type_assertion_failure",
			mockResponse: map[string]interface{}{
				"projects": []interface{}{
					map[string]interface{}{
						"issuetypes": []interface{}{
							map[string]interface{}{
								"name": 123, // number instead of string
								"fields": map[string]interface{}{
									"customfield_10001": map[string]interface{}{
										"required": true,
										"name":     "Should be ignored",
									},
								},
							},
							map[string]interface{}{
								"name": "Feature",
								"fields": map[string]interface{}{
									"customfield_10002": map[string]interface{}{
										"required": true,
										"name":     "Should be included",
									},
								},
							},
						},
					},
				},
			},
			expectedFields: 1, // solo del issuetype válido
		},
		{
			name: "field_name_type_assertion_failure",
			mockResponse: map[string]interface{}{
				"projects": []interface{}{
					map[string]interface{}{
						"issuetypes": []interface{}{
							map[string]interface{}{
								"name": "Feature",
								"fields": map[string]interface{}{
									"customfield_10001": map[string]interface{}{
										"required": true,
										"name":     123, // number instead of string
									},
									"customfield_10002": map[string]interface{}{
										"required": true,
										"name":     "Valid Name",
									},
								},
							},
						},
					},
				},
			},
			expectedFields: 2, // uno con fieldKey, otro con name válido
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.Contains(r.URL.Path, "/rest/api/3/issue/createmeta") {
					w.WriteHeader(http.StatusOK)
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(tt.mockResponse)
				}
			})

			fields, err := fm.ValidateFeatureRequiredFields(context.Background(), "PROJ")

			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
			if len(fields) != tt.expectedFields {
				t.Errorf("Expected %d fields, got %d. Fields: %v", tt.expectedFields, len(fields), fields)
			}
		})
	}
}

func TestFeatureManager_normalizeDescription(t *testing.T) {
	fm, server := createTestFeatureManager()
	defer server.Close()

	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "Simple Feature",
			expected: "simple feature",
		},
		{
			input:    "Feature with @#$%^& special chars!",
			expected: "feature with special chars",
		},
		{
			input:    "  Multiple    spaces   ",
			expected: "multiple spaces",
		},
		{
			input:    "Mixed-Case_Feature123",
			expected: "mixedcase_feature123",
		},
		{
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := fm.normalizeDescription(tt.input)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestFeatureManager_isSimilarDescription(t *testing.T) {
	fm, server := createTestFeatureManager()
	defer server.Close()

	tests := []struct {
		desc1    string
		desc2    string
		expected bool
	}{
		{
			desc1:    "user authentication system",
			desc2:    "user authentication system implementation",
			expected: true, // Alta similaridad
		},
		{
			desc1:    "payment processing feature",
			desc2:    "payment gateway integration",
			expected: false, // Solo comparten "payment" pero no alcanza el 70%
		},
		{
			desc1:    "user management",
			desc2:    "file upload system",
			expected: false, // Sin palabras comunes relevantes
		},
		{
			desc1:    "",
			desc2:    "",
			expected: true, // Ambos vacíos
		},
		{
			desc1:    "test",
			desc2:    "",
			expected: false, // Uno vacío
		},
		{
			desc1:    "a b c",
			desc2:    "x y z",
			expected: false, // Palabras muy cortas (≤2 chars)
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc1+"_vs_"+tt.desc2, func(t *testing.T) {
			result := fm.isSimilarDescription(tt.desc1, tt.desc2)
			if result != tt.expected {
				t.Errorf("Expected %v for '%s' vs '%s', got %v", tt.expected, tt.desc1, tt.desc2, result)
			}
		})
	}
}

func TestFeatureManager_escapeJQLString(t *testing.T) {
	fm, server := createTestFeatureManager()
	defer server.Close()

	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    `Simple text`,
			expected: `Simple text`,
		},
		{
			input:    `Text with "quotes"`,
			expected: `Text with \\"quotes\\"`,
		},
		{
			input:    `Text with \backslash`,
			expected: `Text with \\backslash`,
		},
		{
			input:    `Text with "quotes" and \backslash`,
			expected: `Text with \\"quotes\\" and \\backslash`,
		},
		{
			input:    ``,
			expected: ``,
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := fm.escapeJQLString(tt.input)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestFeatureManager_isJiraKey(t *testing.T) {
	fm, server := createTestFeatureManager()
	defer server.Close()

	tests := []struct {
		input    string
		expected bool
	}{
		{"PROJ-123", true},
		{"ABC-456", true},
		{"X-1", true},
		{"proj-123", false}, // lowercase
		{"PROJ", false},     // no number
		{"123-PROJ", false}, // number first
		{"PROJ-", false},    // no number after dash
		{"PROJ-abc", false}, // letters after dash
		{"", false},         // empty
		{"INVALID", false},  // no dash
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := fm.isJiraKey(tt.input)
			if result != tt.expected {
				t.Errorf("Expected isJiraKey(%s) to be %v, got %v", tt.input, tt.expected, result)
			}
		})
	}
}

func TestFeatureManager_buildFeaturePayload(t *testing.T) {
	tests := []struct {
		name               string
		description        string
		projectKey         string
		requiredFields     string
		expectedFieldCount int
	}{
		{
			name:               "basic_payload",
			description:        "Test Feature",
			projectKey:         "PROJ",
			requiredFields:     "",
			expectedFieldCount: 4, // project, summary, description, issuetype
		},
		{
			name:               "payload_with_additional_fields",
			description:        "Test Feature",
			projectKey:         "PROJ",
			requiredFields:     `{"priority": {"name": "High"}, "labels": ["feature"]}`,
			expectedFieldCount: 6, // basic + priority + labels
		},
		{
			name:               "payload_with_invalid_json",
			description:        "Test Feature",
			projectKey:         "PROJ",
			requiredFields:     `{invalid json}`,
			expectedFieldCount: 4, // Should fallback to basic fields
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := createTestConfig()
			cfg.FeatureRequiredFields = tt.requiredFields
			jiraClient := NewJiraClient(cfg)
			fm := NewFeatureManager(jiraClient, cfg)

			payload := fm.buildFeaturePayload(tt.description, tt.projectKey)

			fields, ok := payload["fields"].(map[string]interface{})
			if !ok {
				t.Fatal("Expected fields to be a map")
			}

			if len(fields) != tt.expectedFieldCount {
				t.Errorf("Expected %d fields, got %d", tt.expectedFieldCount, len(fields))
			}

			// Verificar campos básicos
			if fields["summary"] != tt.description {
				t.Errorf("Expected summary to be '%s', got %v", tt.description, fields["summary"])
			}

			project, ok := fields["project"].(map[string]interface{})
			if !ok {
				t.Fatal("Expected project to be a map")
			}
			if project["key"] != tt.projectKey {
				t.Errorf("Expected project key to be '%s', got %v", tt.projectKey, project["key"])
			}

			issueType, ok := fields["issuetype"].(map[string]interface{})
			if !ok {
				t.Fatal("Expected issuetype to be a map")
			}
			if issueType["name"] != cfg.FeatureIssueType {
				t.Errorf("Expected issue type to be '%s', got %v", cfg.FeatureIssueType, issueType["name"])
			}
		})
	}
}

func TestFeatureManager_CreateOrGetFeature_SearchError(t *testing.T) {
	fm, server := createTestFeatureManager()
	defer server.Close()

	description := "Test Feature Description"

	// Mock para búsqueda que falla
	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/rest/api/3/search") {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Search failed"))
		}
	})

	result, err := fm.CreateOrGetFeature(context.Background(), description, "PROJ")

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result.Success {
		t.Error("Expected result to not be successful")
	}

	if !strings.Contains(result.ErrorMessage, "Error searching existing features") {
		t.Errorf("Expected search error, got: %s", result.ErrorMessage)
	}
}

func TestFeatureManager_CreateOrGetFeature_CreateError(t *testing.T) {
	fm, server := createTestFeatureManager()
	defer server.Close()

	description := "New Feature Description"

	// Mock para crear nuevo feature que falla
	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/rest/api/3/search") {
			// No features found
			response := JiraSearchResponse{
				Issues: []JiraIssue{},
				Total:  0,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		} else if strings.Contains(r.URL.Path, "/rest/api/3/issue") && r.Method == "POST" {
			// Create fails
			w.WriteHeader(http.StatusBadRequest)
			errorResp := JiraErrorResponse{
				ErrorMessages: []string{"Feature creation failed"},
			}
			json.NewEncoder(w).Encode(errorResp)
		}
	})

	result, err := fm.CreateOrGetFeature(context.Background(), description, "PROJ")

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result.Success {
		t.Error("Expected result to not be successful")
	}

	if !strings.Contains(result.ErrorMessage, "Error creating feature") {
		t.Errorf("Expected creation error, got: %s", result.ErrorMessage)
	}
}

func TestFeatureManager_SearchExistingFeature_ErrorPaths(t *testing.T) {
	tests := []struct {
		name        string
		serverFunc  func(w http.ResponseWriter, r *http.Request)
		expectError bool
	}{
		{
			name: "network_error",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				hj, ok := w.(http.Hijacker)
				if !ok {
					http.Error(w, "webserver doesn't support hijacking", http.StatusInternalServerError)
					return
				}
				conn, _, _ := hj.Hijack()
				conn.Close()
			},
			expectError: true,
		},
		{
			name: "invalid_status_code",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Server error"))
			},
			expectError: true,
		},
		{
			name: "invalid_json_response",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("invalid json"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create individual server for each subtest to avoid race conditions
			server := httptest.NewServer(http.HandlerFunc(tt.serverFunc))
			defer server.Close()

			cfg := &config.Config{
				JiraURL:          server.URL,
				JiraEmail:        "test@example.com",
				JiraAPIToken:     "test-token",
				FeatureIssueType: "Feature",
			}

			jiraClient := NewJiraClient(cfg)
			fm := NewFeatureManager(jiraClient, cfg)

			key, err := fm.SearchExistingFeature(context.Background(), "Test", "PROJ")
			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				if key != "" {
					t.Errorf("Expected empty key on error, got: %s", key)
				}
			}
		})
	}
}

func TestFeatureManager_SearchExistingFeature_EdgeCases(t *testing.T) {
	fm, server := createTestFeatureManager()
	defer server.Close()

	tests := []struct {
		name        string
		description string
		mockIssues  []JiraIssue
		expectedKey string
	}{
		{
			name:        "issue_without_summary_field",
			description: "Test Feature",
			mockIssues: []JiraIssue{
				{
					Key: "PROJ-123",
					Fields: map[string]interface{}{
						"description": "Some description without summary",
					},
				},
			},
			expectedKey: "", // Should not match without summary
		},
		{
			name:        "issue_with_non_string_summary",
			description: "Test Feature",
			mockIssues: []JiraIssue{
				{
					Key: "PROJ-123",
					Fields: map[string]interface{}{
						"summary": 123, // number instead of string
					},
				},
			},
			expectedKey: "", // Should not match with non-string summary
		},
		{
			name:        "multiple_issues_first_matches",
			description: "Test Feature System Implementation",
			mockIssues: []JiraIssue{
				{
					Key: "PROJ-123",
					Fields: map[string]interface{}{
						"summary": "Test Feature System Development", // Should match due to similarity
					},
				},
				{
					Key: "PROJ-124",
					Fields: map[string]interface{}{
						"summary": "Different Feature",
					},
				},
			},
			expectedKey: "PROJ-123", // Should return first match
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.Contains(r.URL.Path, "/rest/api/3/search") {
					response := JiraSearchResponse{
						Issues: tt.mockIssues,
						Total:  len(tt.mockIssues),
					}
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(response)
				}
			})

			key, err := fm.SearchExistingFeature(context.Background(), tt.description, "PROJ")
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
			if key != tt.expectedKey {
				t.Errorf("Expected key '%s', got '%s'", tt.expectedKey, key)
			}
		})
	}
}
