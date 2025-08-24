package jira

import (
	"strings"
	"testing"
)

func TestNewADFDocument(t *testing.T) {
	doc := NewADFDocument()

	if doc == nil {
		t.Fatal("Expected ADF document to be created")
	}

	if doc.Type != "doc" {
		t.Errorf("Expected type to be 'doc', got %s", doc.Type)
	}

	if doc.Version != 1 {
		t.Errorf("Expected version to be 1, got %d", doc.Version)
	}

	if len(doc.Content) != 0 {
		t.Errorf("Expected empty content, got %d items", len(doc.Content))
	}
}

func TestADFDocument_AddParagraph(t *testing.T) {
	doc := NewADFDocument()
	text := "Test paragraph text"

	doc.AddParagraph(text)

	if len(doc.Content) != 1 {
		t.Errorf("Expected 1 content item, got %d", len(doc.Content))
	}

	paragraph := doc.Content[0]
	if paragraph.Type != "paragraph" {
		t.Errorf("Expected type 'paragraph', got %s", paragraph.Type)
	}

	if len(paragraph.Content) != 1 {
		t.Errorf("Expected 1 text item, got %d", len(paragraph.Content))
	}

	textItem := paragraph.Content[0]
	if textItem.Type != "text" {
		t.Errorf("Expected text type 'text', got %s", textItem.Type)
	}

	if textItem.Text != text {
		t.Errorf("Expected text '%s', got '%s'", text, textItem.Text)
	}
}

func TestADFDocument_AddBulletList(t *testing.T) {
	tests := []struct {
		name     string
		items    []string
		expected int
	}{
		{
			name:     "multiple_items",
			items:    []string{"Item 1", "Item 2", "Item 3"},
			expected: 3,
		},
		{
			name:     "items_with_spaces",
			items:    []string{" Item 1 ", "  Item 2  ", "Item 3"},
			expected: 3,
		},
		{
			name:     "items_with_empty",
			items:    []string{"Item 1", "", "Item 3", "   "},
			expected: 2, // Solo los no vacíos
		},
		{
			name:     "empty_list",
			items:    []string{},
			expected: 0,
		},
		{
			name:     "all_empty_items",
			items:    []string{"", "   ", "\t"},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := NewADFDocument()
			doc.AddBulletList(tt.items)

			if len(doc.Content) != tt.expected {
				t.Errorf("Expected %d content items, got %d", tt.expected, len(doc.Content))
			}

			// Verificar que cada item no vacío tiene bullet
			for i, item := range tt.items {
				if trimmed := item; len(trimmed) > 0 && trimmed != "   " && trimmed != "\t" {
					if i < len(doc.Content) {
						paragraph := doc.Content[i]
						if len(paragraph.Content) > 0 {
							text := paragraph.Content[0].Text
							if !containsBullet(text) {
								t.Errorf("Expected bullet in text '%s'", text)
							}
						}
					}
				}
			}
		})
	}
}

func TestCreateAcceptanceCriteriaADF(t *testing.T) {
	tests := []struct {
		name        string
		criteria    string
		expectedLen int
		isList      bool
	}{
		{
			name:        "empty_criteria",
			criteria:    "",
			expectedLen: 0,
			isList:      false,
		},
		{
			name:        "single_criteria",
			criteria:    "Single acceptance criteria",
			expectedLen: 1,
			isList:      false,
		},
		{
			name:        "multiple_criteria_semicolon",
			criteria:    "Criteria 1; Criteria 2; Criteria 3",
			expectedLen: 3,
			isList:      true,
		},
		{
			name:        "multiple_criteria_newlines",
			criteria:    "Criteria 1\nCriteria 2\nCriteria 3",
			expectedLen: 3,
			isList:      true,
		},
		{
			name:        "mixed_separators",
			criteria:    "Criteria 1; Criteria 2\nCriteria 3",
			expectedLen: 2, // Se prioriza ';'
			isList:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := CreateAcceptanceCriteriaADF(tt.criteria)

			if doc == nil {
				t.Fatal("Expected ADF document to be created")
			}

			if len(doc.Content) != tt.expectedLen {
				t.Errorf("Expected %d content items, got %d", tt.expectedLen, len(doc.Content))
			}

			if tt.isList && tt.expectedLen > 0 {
				// Verificar que los items tienen bullets
				for _, content := range doc.Content {
					if len(content.Content) > 0 {
						text := content.Content[0].Text
						if !containsBullet(text) {
							t.Errorf("Expected bullet in list item '%s'", text)
						}
					}
				}
			}
		})
	}
}

func TestCreateDescriptionWithCriteriaADF(t *testing.T) {
	tests := []struct {
		name           string
		description    string
		criteria       string
		expectedMinLen int
		hasSeparator   bool
	}{
		{
			name:           "both_empty",
			description:    "",
			criteria:       "",
			expectedMinLen: 0,
			hasSeparator:   false,
		},
		{
			name:           "description_only",
			description:    "Test description",
			criteria:       "",
			expectedMinLen: 1,
			hasSeparator:   false,
		},
		{
			name:           "criteria_only",
			description:    "",
			criteria:       "Test criteria",
			expectedMinLen: 3, // separador vacío + header + criterio
			hasSeparator:   true,
		},
		{
			name:           "both_present",
			description:    "Test description",
			criteria:       "Test criteria",
			expectedMinLen: 4, // descripción + separador + header + criterio
			hasSeparator:   true,
		},
		{
			name:           "multiple_criteria",
			description:    "Test description",
			criteria:       "Criteria 1; Criteria 2",
			expectedMinLen: 5, // descripción + separador + header + 2 criterios
			hasSeparator:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := CreateDescriptionWithCriteriaADF(tt.description, tt.criteria)

			if doc == nil {
				t.Fatal("Expected ADF document to be created")
			}

			if len(doc.Content) < tt.expectedMinLen {
				t.Errorf("Expected at least %d content items, got %d", tt.expectedMinLen, len(doc.Content))
			}

			if tt.hasSeparator && len(doc.Content) >= 3 {
				// Verificar que hay un separador y header
				found := false
				for _, content := range doc.Content {
					if len(content.Content) > 0 {
						text := content.Content[0].Text
						if text == "--- Criterios de Aceptación ---" {
							found = true
							break
						}
					}
				}
				if !found {
					t.Error("Expected to find criteria header separator")
				}
			}
		})
	}
}

func TestCreateDescriptionADF(t *testing.T) {
	tests := []struct {
		name        string
		description string
		expectedLen int
	}{
		{
			name:        "empty_description",
			description: "",
			expectedLen: 0,
		},
		{
			name:        "simple_description",
			description: "Simple test description",
			expectedLen: 1,
		},
		{
			name:        "description_with_spaces",
			description: "  Description with spaces  ",
			expectedLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := CreateDescriptionADF(tt.description)

			if doc == nil {
				t.Fatal("Expected ADF document to be created")
			}

			if len(doc.Content) != tt.expectedLen {
				t.Errorf("Expected %d content items, got %d", tt.expectedLen, len(doc.Content))
			}

			if tt.expectedLen > 0 {
				paragraph := doc.Content[0]
				if paragraph.Type != "paragraph" {
					t.Errorf("Expected paragraph type, got %s", paragraph.Type)
				}

				if len(paragraph.Content) > 0 {
					text := paragraph.Content[0].Text
					if text != tt.description {
						t.Errorf("Expected text '%s', got '%s'", tt.description, text)
					}
				}
			}
		})
	}
}

func TestSplitCriteria(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "semicolon_separated",
			input:    "Criteria 1; Criteria 2; Criteria 3",
			expected: []string{"Criteria 1", "Criteria 2", "Criteria 3"},
		},
		{
			name:     "newline_separated",
			input:    "Criteria 1\nCriteria 2\nCriteria 3",
			expected: []string{"Criteria 1", "Criteria 2", "Criteria 3"},
		},
		{
			name:     "mixed_spaces",
			input:    " Criteria 1 ; Criteria 2 ; Criteria 3 ",
			expected: []string{"Criteria 1", "Criteria 2", "Criteria 3"},
		},
		{
			name:     "empty_parts",
			input:    "Criteria 1;; Criteria 2; ; Criteria 3",
			expected: []string{"Criteria 1", "Criteria 2", "Criteria 3"},
		},
		{
			name:     "single_criterion",
			input:    "Single criterion",
			expected: []string{"Single criterion"},
		},
		{
			name:     "empty_input",
			input:    "",
			expected: []string{""},
		},
		{
			name:     "only_spaces",
			input:    "   ",
			expected: []string{""},
		},
		{
			name:     "semicolon_priority",
			input:    "Item 1; Item 2\nItem 3",
			expected: []string{"Item 1", "Item 2\nItem 3"}, // Los ; tienen prioridad
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitCriteria(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d items, got %d", len(tt.expected), len(result))
			}

			for i, expected := range tt.expected {
				if i < len(result) {
					if result[i] != expected {
						t.Errorf("Expected item %d to be '%s', got '%s'", i, expected, result[i])
					}
				}
			}
		})
	}
}

// Helper function para verificar si un texto contiene bullet
func containsBullet(text string) bool {
	return strings.HasPrefix(text, "• ")
}
