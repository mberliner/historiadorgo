package entities

import (
	"testing"
)

func TestNewUserStory(t *testing.T) {
	tests := []struct {
		name               string
		titulo             string
		descripcion        string
		criterioAceptacion string
		subtareasRaw       string
		parent             string
		wantSubtareas      []string
		wantHasSubtareas   bool
		wantHasParent      bool
	}{
		{
			name:               "valid user story with subtasks",
			titulo:             "Test Story",
			descripcion:        "Test description",
			criterioAceptacion: "Test criteria",
			subtareasRaw:       "Task 1;Task 2;Task 3",
			parent:             "PROJ-100",
			wantSubtareas:      []string{"Task 1", "Task 2", "Task 3"},
			wantHasSubtareas:   true,
			wantHasParent:      true,
		},
		{
			name:               "story without subtasks",
			titulo:             "Simple Story",
			descripcion:        "Simple description",
			criterioAceptacion: "Simple criteria",
			subtareasRaw:       "",
			parent:             "",
			wantSubtareas:      nil,
			wantHasSubtareas:   false,
			wantHasParent:      false,
		},
		{
			name:               "subtasks with newlines",
			titulo:             "Story with newline subtasks",
			descripcion:        "Test description",
			criterioAceptacion: "Test criteria",
			subtareasRaw:       "Task 1\nTask 2\nTask 3",
			parent:             "",
			wantSubtareas:      []string{"Task 1", "Task 2", "Task 3"},
			wantHasSubtareas:   true,
			wantHasParent:      false,
		},
		{
			name:               "mixed separators",
			titulo:             "Mixed separators story",
			descripcion:        "Test description",
			criterioAceptacion: "Test criteria",
			subtareasRaw:       "Task 1;Task 2\nTask 3;Task 4",
			parent:             "Feature Description",
			wantSubtareas:      []string{"Task 1", "Task 2", "Task 3", "Task 4"},
			wantHasSubtareas:   true,
			wantHasParent:      true,
		},
		{
			name:               "subtasks with empty spaces",
			titulo:             "Story with spaces",
			descripcion:        "Test description",
			criterioAceptacion: "Test criteria",
			subtareasRaw:       "  Task 1  ; ; Task 2  \n  \n Task 3  ",
			parent:             "",
			wantSubtareas:      []string{"Task 1", "Task 2", "Task 3"},
			wantHasSubtareas:   true,
			wantHasParent:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			story := NewUserStory(
				tt.titulo,
				tt.descripcion,
				tt.criterioAceptacion,
				tt.subtareasRaw,
				tt.parent,
			)

			if story.Titulo != tt.titulo {
				t.Errorf("titulo = %v, want %v", story.Titulo, tt.titulo)
			}
			if story.Descripcion != tt.descripcion {
				t.Errorf("descripcion = %v, want %v", story.Descripcion, tt.descripcion)
			}
			if story.CriterioAceptacion != tt.criterioAceptacion {
				t.Errorf("criterioAceptacion = %v, want %v", story.CriterioAceptacion, tt.criterioAceptacion)
			}
			if story.Parent != tt.parent {
				t.Errorf("parent = %v, want %v", story.Parent, tt.parent)
			}

			if len(story.Subtareas) != len(tt.wantSubtareas) {
				t.Errorf("subtareas length = %v, want %v", len(story.Subtareas), len(tt.wantSubtareas))
			}
			for i, subtarea := range story.Subtareas {
				if i < len(tt.wantSubtareas) && subtarea != tt.wantSubtareas[i] {
					t.Errorf("subtarea[%d] = %v, want %v", i, subtarea, tt.wantSubtareas[i])
				}
			}

			if story.HasSubtareas() != tt.wantHasSubtareas {
				t.Errorf("HasSubtareas() = %v, want %v", story.HasSubtareas(), tt.wantHasSubtareas)
			}
			if story.HasParent() != tt.wantHasParent {
				t.Errorf("HasParent() = %v, want %v", story.HasParent(), tt.wantHasParent)
			}
		})
	}
}

func TestUserStory_GetValidSubtareas(t *testing.T) {
	tests := []struct {
		name      string
		subtareas []string
		want      []string
	}{
		{
			name:      "all valid subtasks",
			subtareas: []string{"Task 1", "Task 2", "Task 3"},
			want:      []string{"Task 1", "Task 2", "Task 3"},
		},
		{
			name:      "mixed valid and invalid",
			subtareas: []string{"Valid task", "", "Another valid", string(make([]byte, 256))},
			want:      []string{"Valid task", "Another valid"},
		},
		{
			name:      "empty subtasks",
			subtareas: []string{},
			want:      []string{},
		},
		{
			name:      "nil subtasks",
			subtareas: nil,
			want:      []string{},
		},
		{
			name:      "all invalid subtasks",
			subtareas: []string{"", string(make([]byte, 256)), ""},
			want:      []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			story := &UserStory{
				Titulo:             "Test",
				Descripcion:        "Test",
				CriterioAceptacion: "Test",
				Subtareas:          tt.subtareas,
			}

			got := story.GetValidSubtareas()
			if len(got) != len(tt.want) {
				t.Errorf("GetValidSubtareas() length = %v, want %v", len(got), len(tt.want))
				return
			}

			for i, validSubtarea := range got {
				if i < len(tt.want) && validSubtarea != tt.want[i] {
					t.Errorf("GetValidSubtareas()[%d] = %v, want %v", i, validSubtarea, tt.want[i])
				}
			}
		})
	}
}

func TestUserStory_ParseSubtareas(t *testing.T) {
	story := &UserStory{}

	tests := []struct {
		name         string
		subtareasRaw string
		want         []string
	}{
		{
			name:         "semicolon separated",
			subtareasRaw: "Task 1;Task 2;Task 3",
			want:         []string{"Task 1", "Task 2", "Task 3"},
		},
		{
			name:         "newline separated",
			subtareasRaw: "Task 1\nTask 2\nTask 3",
			want:         []string{"Task 1", "Task 2", "Task 3"},
		},
		{
			name:         "empty string",
			subtareasRaw: "",
			want:         nil,
		},
		{
			name:         "only whitespace",
			subtareasRaw: "   \n  ; \n ",
			want:         nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset subtareas for each test
			story.Subtareas = nil

			// Use reflection to call private method parseSubtareas
			// Since Go doesn't have reflection for private methods in tests,
			// we'll create a new story to test this functionality
			testStory := NewUserStory("Test", "Test", "Test", tt.subtareasRaw, "")

			if len(testStory.Subtareas) != len(tt.want) {
				t.Errorf("parseSubtareas() length = %v, want %v", len(testStory.Subtareas), len(tt.want))
				return
			}

			for i, subtarea := range testStory.Subtareas {
				if i < len(tt.want) && subtarea != tt.want[i] {
					t.Errorf("parseSubtareas()[%d] = %v, want %v", i, subtarea, tt.want[i])
				}
			}
		})
	}
}
