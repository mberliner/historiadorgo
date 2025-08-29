package config

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDetectAcceptanceCriteriaField(t *testing.T) {
	tests := []struct {
		name          string
		responseBody  string
		statusCode    int
		expectedField string
		expectedError bool
	}{
		{
			name:       "success with acceptance criteria field",
			statusCode: 200,
			responseBody: `{
				"projects": [{
					"issuetypes": [{
						"name": "Story",
						"fields": {
							"customfield_10147": {
								"name": "Acceptance Criteria",
								"required": false
							},
							"summary": {
								"name": "Summary",
								"required": true
							}
						}
					}]
				}]
			}`,
			expectedField: "customfield_10147",
			expectedError: false,
		},
		{
			name:       "success with criterio field in Spanish",
			statusCode: 200,
			responseBody: `{
				"projects": [{
					"issuetypes": [{
						"name": "Story",
						"fields": {
							"customfield_10200": {
								"name": "Criterio de Aceptación",
								"required": false
							}
						}
					}]
				}]
			}`,
			expectedField: "customfield_10200",
			expectedError: false,
		},
		{
			name:       "field not found",
			statusCode: 200,
			responseBody: `{
				"projects": [{
					"issuetypes": [{
						"name": "Story",
						"fields": {
							"summary": {
								"name": "Summary",
								"required": true
							}
						}
					}]
				}]
			}`,
			expectedField: "",
			expectedError: true,
		},
		{
			name:          "HTTP error",
			statusCode:    401,
			responseBody:  `{"error": "Unauthorized"}`,
			expectedField: "",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client := &http.Client{Timeout: 5 * time.Second}
			ctx := context.Background()

			field, err := detectAcceptanceCriteriaField(ctx, client, server.URL, "test@example.com", "token", "TEST", "Story")

			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if field != tt.expectedField {
					t.Errorf("expected field %q, got %q", tt.expectedField, field)
				}
			}
		})
	}
}

func TestDetectFeatureRequiredFields(t *testing.T) {
	tests := []struct {
		name           string
		responseBody   string
		statusCode     int
		expectedFields string
		expectedError  bool
	}{
		{
			name:       "success with required field",
			statusCode: 200,
			responseBody: `{
				"projects": [{
					"issuetypes": [{
						"name": "Feature",
						"fields": {
							"customfield_11493": {
								"name": "Backlog",
								"required": true,
								"allowedValues": [
									{"id": "54672", "value": "Product Backlog"}
								]
							},
							"summary": {
								"name": "Summary",
								"required": true
							}
						}
					}]
				}]
			}`,
			expectedFields: `{"customfield_11493":{"id":"54672"}}`,
			expectedError:  false,
		},
		{
			name:       "no required fields",
			statusCode: 200,
			responseBody: `{
				"projects": [{
					"issuetypes": [{
						"name": "Feature",
						"fields": {
							"summary": {
								"name": "Summary",
								"required": true
							}
						}
					}]
				}]
			}`,
			expectedFields: "{}",
			expectedError:  false,
		},
		{
			name:           "HTTP error",
			statusCode:     404,
			responseBody:   `{"error": "Not Found"}`,
			expectedFields: "",
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client := &http.Client{Timeout: 5 * time.Second}
			ctx := context.Background()

			fields, err := detectFeatureRequiredFields(ctx, client, server.URL, "test@example.com", "token", "TEST", "Feature")

			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if fields != tt.expectedFields {
					t.Errorf("expected fields %q, got %q", tt.expectedFields, fields)
				}
			}
		})
	}
}

func TestDetectJiraConfiguration(t *testing.T) {
	// Mock server that returns both acceptance criteria and feature fields
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("issuetypeNames") == "Story" {
			// Story response with acceptance criteria field
			w.Write([]byte(`{
				"projects": [{
					"issuetypes": [{
						"name": "Story",
						"fields": {
							"customfield_10147": {
								"name": "Acceptance Criteria",
								"required": false
							}
						}
					}]
				}]
			}`))
		} else if r.URL.Query().Get("issuetypeNames") == "Feature" {
			// Feature response with required fields
			w.Write([]byte(`{
				"projects": [{
					"issuetypes": [{
						"name": "Feature",
						"fields": {
							"customfield_11493": {
								"name": "Backlog",
								"required": true,
								"allowedValues": [
									{"id": "54672", "value": "Product Backlog"}
								]
							}
						}
					}]
				}]
			}`))
		}
	}))
	defer server.Close()

	config, err := DetectJiraConfiguration(server.URL, "test@example.com", "token", "TEST", "Story", "Feature")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if config.AcceptanceCriteriaField != "customfield_10147" {
		t.Errorf("expected acceptance criteria field %q, got %q", "customfield_10147", config.AcceptanceCriteriaField)
	}

	expectedFeatureFields := `{"customfield_11493":{"id":"54672"}}`
	if config.FeatureRequiredFields != expectedFeatureFields {
		t.Errorf("expected feature fields %q, got %q", expectedFeatureFields, config.FeatureRequiredFields)
	}
}
