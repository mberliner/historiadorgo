package filesystem

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"historiadorgo/internal/domain/entities"
)

func TestNewFileProcessor(t *testing.T) {
	processedDir := "/test/processed"
	fp := NewFileProcessor(processedDir)

	if fp == nil {
		t.Fatal("Expected FileProcessor to be created")
	}

	if fp.processedDir != processedDir {
		t.Errorf("Expected processedDir to be %s, got %s", processedDir, fp.processedDir)
	}

	if fp.validator == nil {
		t.Error("Expected validator to be initialized")
	}
}

func TestFileProcessor_ValidateFile(t *testing.T) {
	// Crear directorio temporal para tests
	tempDir := t.TempDir()
	fp := NewFileProcessor(tempDir)

	tests := []struct {
		name        string
		setupFile   func() string
		expectedErr string
	}{
		{
			name: "file_does_not_exist",
			setupFile: func() string {
				return filepath.Join(tempDir, "nonexistent.csv")
			},
			expectedErr: "file does not exist",
		},
		{
			name: "unsupported_file_format",
			setupFile: func() string {
				filePath := filepath.Join(tempDir, "test.txt")
				os.WriteFile(filePath, []byte("some content"), 0644)
				return filePath
			},
			expectedErr: "unsupported file format",
		},
		{
			name: "valid_csv_file",
			setupFile: func() string {
				content := `titulo,descripcion,criterio_aceptacion,subtareas,parent
Test Story,Test Description,Test Criteria,Task 1;Task 2,PROJ-123`
				filePath := filepath.Join(tempDir, "valid.csv")
				os.WriteFile(filePath, []byte(content), 0644)
				return filePath
			},
			expectedErr: "",
		},
		{
			name: "empty_csv_file",
			setupFile: func() string {
				content := `titulo,descripcion,criterio_aceptacion,subtareas,parent`
				filePath := filepath.Join(tempDir, "empty.csv")
				os.WriteFile(filePath, []byte(content), 0644)
				return filePath
			},
			expectedErr: "file contains no valid stories",
		},
		{
			name: "csv_with_invalid_story",
			setupFile: func() string {
				content := `titulo,descripcion,criterio_aceptacion,subtareas,parent
,Test Description,Test Criteria,Task 1,PROJ-123`
				filePath := filepath.Join(tempDir, "invalid.csv")
				os.WriteFile(filePath, []byte(content), 0644)
				return filePath
			},
			expectedErr: "file contains no valid stories",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := tt.setupFile()
			defer os.RemoveAll(filePath)

			err := fp.ValidateFile(context.Background(), filePath)

			if tt.expectedErr == "" {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected error containing '%s', got no error", tt.expectedErr)
				} else if !strings.Contains(err.Error(), tt.expectedErr) {
					t.Errorf("Expected error containing '%s', got: %v", tt.expectedErr, err)
				}
			}
		})
	}
}

func TestFileProcessor_ReadFile(t *testing.T) {
	tempDir := t.TempDir()
	fp := NewFileProcessor(tempDir)

	tests := []struct {
		name          string
		setupFile     func() string
		expectedCount int
		expectedErr   string
		validateStory func(*entities.UserStory) bool
	}{
		{
			name: "valid_csv_file",
			setupFile: func() string {
				content := `titulo,descripcion,criterio_aceptacion,subtareas,parent
Story 1,Description 1,Criteria 1,Task 1;Task 2,PROJ-123
Story 2,Description 2,Criteria 2,,`
				filePath := filepath.Join(tempDir, "test.csv")
				os.WriteFile(filePath, []byte(content), 0644)
				return filePath
			},
			expectedCount: 2,
			expectedErr:   "",
			validateStory: func(story *entities.UserStory) bool {
				return story.Titulo != "" && story.Descripcion != "" && story.CriterioAceptacion != ""
			},
		},
		{
			name: "csv_with_subtasks_newlines",
			setupFile: func() string {
				content := `titulo,descripcion,criterio_aceptacion,subtareas,parent
"Story 1","Description 1","Criteria 1","Task 1
Task 2
Task 3",PROJ-123`
				filePath := filepath.Join(tempDir, "subtasks.csv")
				os.WriteFile(filePath, []byte(content), 0644)
				return filePath
			},
			expectedCount: 1,
			expectedErr:   "",
			validateStory: func(story *entities.UserStory) bool {
				return len(story.Subtareas) == 3
			},
		},
		{
			name: "unsupported_format",
			setupFile: func() string {
				filePath := filepath.Join(tempDir, "test.txt")
				os.WriteFile(filePath, []byte("content"), 0644)
				return filePath
			},
			expectedCount: 0,
			expectedErr:   "unsupported file format",
		},
		{
			name: "malformed_csv",
			setupFile: func() string {
				content := `titulo,descripcion,criterio_aceptacion
"Story 1,"Description 1","Criteria 1"`
				filePath := filepath.Join(tempDir, "malformed.csv")
				os.WriteFile(filePath, []byte(content), 0644)
				return filePath
			},
			expectedCount: 0,
			expectedErr:   "error parsing CSV",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := tt.setupFile()
			defer os.RemoveAll(filePath)

			stories, err := fp.ReadFile(context.Background(), filePath)

			if tt.expectedErr == "" {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				if len(stories) != tt.expectedCount {
					t.Errorf("Expected %d stories, got %d", tt.expectedCount, len(stories))
				}
				if tt.validateStory != nil && len(stories) > 0 {
					if !tt.validateStory(stories[0]) {
						t.Error("Story validation failed")
					}
				}
			} else {
				if err == nil {
					t.Errorf("Expected error containing '%s', got no error", tt.expectedErr)
				} else if !strings.Contains(err.Error(), tt.expectedErr) {
					t.Errorf("Expected error containing '%s', got: %v", tt.expectedErr, err)
				}
			}
		})
	}
}

func TestFileProcessor_MoveToProcessed(t *testing.T) {
	tempDir := t.TempDir()
	processedDir := filepath.Join(tempDir, "processed")
	fp := NewFileProcessor(processedDir)

	// Crear archivo fuente
	sourceFile := filepath.Join(tempDir, "source.csv")
	content := "test content"
	err := os.WriteFile(sourceFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Mover archivo
	err = fp.MoveToProcessed(context.Background(), sourceFile)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verificar que el directorio processed fue creado
	if _, err := os.Stat(processedDir); os.IsNotExist(err) {
		t.Error("Expected processed directory to be created")
	}

	// Verificar que el archivo fue movido
	destFile := filepath.Join(processedDir, "source.csv")
	if _, err := os.Stat(destFile); os.IsNotExist(err) {
		t.Error("Expected file to be moved to processed directory")
	}

	// Verificar que el archivo original ya no existe
	if _, err := os.Stat(sourceFile); !os.IsNotExist(err) {
		t.Error("Expected original file to be removed")
	}

	// Verificar contenido
	movedContent, err := os.ReadFile(destFile)
	if err != nil {
		t.Errorf("Failed to read moved file: %v", err)
	}
	if string(movedContent) != content {
		t.Errorf("Expected content '%s', got '%s'", content, string(movedContent))
	}
}

func TestFileProcessor_MoveToProcessed_ErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func() (string, string) // retorna (sourceFile, processedDir)
		expectError bool
	}{
		{
			name: "nonexistent_source_file",
			setupFunc: func() (string, string) {
				tempDir := t.TempDir()
				processedDir := filepath.Join(tempDir, "processed")
				sourceFile := filepath.Join(tempDir, "nonexistent.csv")
				return sourceFile, processedDir
			},
			expectError: true,
		},
		{
			name: "invalid_processed_directory",
			setupFunc: func() (string, string) {
				tempDir := t.TempDir()
				// Crear archivo fuente válido
				sourceFile := filepath.Join(tempDir, "source.csv")
				os.WriteFile(sourceFile, []byte("content"), 0644)

				// Usar directorio inválido (archivo existente como directorio)
				invalidDir := filepath.Join(tempDir, "invalid_dir")
				os.WriteFile(invalidDir, []byte("not a directory"), 0644)
				processedDir := filepath.Join(invalidDir, "processed")

				return sourceFile, processedDir
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sourceFile, processedDir := tt.setupFunc()
			fp := NewFileProcessor(processedDir)

			err := fp.MoveToProcessed(context.Background(), sourceFile)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			}
		})
	}
}

func TestFileProcessor_GetPendingFiles(t *testing.T) {
	tempDir := t.TempDir()
	fp := NewFileProcessor(tempDir)

	// Crear estructura de archivos
	files := []struct {
		path       string
		shouldFind bool
	}{
		{"test1.csv", true},
		{"test2.xlsx", true},
		{"test3.xls", true},
		{"test4.txt", false},
		{"test5.pdf", false},
		{"subdir/test6.csv", true},
		{"subdir/test7.doc", false},
	}

	for _, file := range files {
		fullPath := filepath.Join(tempDir, file.path)
		err := os.MkdirAll(filepath.Dir(fullPath), 0755)
		if err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		err = os.WriteFile(fullPath, []byte("content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create file %s: %v", file.path, err)
		}
	}

	pendingFiles, err := fp.GetPendingFiles(context.Background(), tempDir)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	expectedCount := 0
	for _, file := range files {
		if file.shouldFind {
			expectedCount++
		}
	}

	if len(pendingFiles) != expectedCount {
		t.Errorf("Expected %d files, got %d", expectedCount, len(pendingFiles))
	}

	// Verificar que solo se encontraron archivos válidos
	for _, foundFile := range pendingFiles {
		ext := strings.ToLower(filepath.Ext(foundFile))
		if ext != ".csv" && ext != ".xlsx" && ext != ".xls" {
			t.Errorf("Found unexpected file with extension %s: %s", ext, foundFile)
		}
	}
}

func TestFileProcessor_GetPendingFiles_ErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func() string // retorna inputDir
		expectError bool
	}{
		{
			name: "nonexistent_directory",
			setupFunc: func() string {
				return "/nonexistent/directory"
			},
			expectError: true,
		},
		{
			name: "file_instead_of_directory",
			setupFunc: func() string {
				tempDir := t.TempDir()
				filePath := filepath.Join(tempDir, "test.csv")
				os.WriteFile(filePath, []byte("titulo,descripcion\nTest,Description"), 0644)
				return filePath
			},
			expectError: false, // filepath.Walk no falla con archivos individuales
		},
		{
			name: "empty_directory",
			setupFunc: func() string {
				return t.TempDir()
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputDir := tt.setupFunc()
			fp := NewFileProcessor(t.TempDir())

			files, err := fp.GetPendingFiles(context.Background(), inputDir)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				if tt.name == "empty_directory" && len(files) != 0 {
					t.Errorf("Expected 0 files in empty directory, got %d", len(files))
				}
			}
		})
	}
}

func TestFileProcessor_mapColumns(t *testing.T) {
	fp := NewFileProcessor("/test")

	tests := []struct {
		name     string
		header   []string
		expected map[string]int
	}{
		{
			name:   "standard_headers",
			header: []string{"titulo", "descripcion", "criterio_aceptacion", "subtareas", "parent"},
			expected: map[string]int{
				"titulo":              0,
				"descripcion":         1,
				"criterio_aceptacion": 2,
				"subtareas":           3,
				"parent":              4,
			},
		},
		{
			name:   "mixed_case_headers",
			header: []string{"TITULO", "Descripcion", "Criterio_Aceptacion", "SUBTAREAS", "Parent"},
			expected: map[string]int{
				"titulo":              0,
				"descripcion":         1,
				"criterio_aceptacion": 2,
				"subtareas":           3,
				"parent":              4,
			},
		},
		{
			name:   "headers_with_spaces",
			header: []string{" titulo ", " descripcion ", " criterio_aceptacion ", " subtareas ", " parent "},
			expected: map[string]int{
				"titulo":              0,
				"descripcion":         1,
				"criterio_aceptacion": 2,
				"subtareas":           3,
				"parent":              4,
			},
		},
		{
			name:   "missing_headers",
			header: []string{"titulo", "other_column", "descripcion"},
			expected: map[string]int{
				"titulo":      0,
				"descripcion": 2,
			},
		},
		{
			name:     "empty_headers",
			header:   []string{},
			expected: map[string]int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fp.mapColumns(tt.header)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d mapped columns, got %d", len(tt.expected), len(result))
			}

			for key, expectedIndex := range tt.expected {
				if actualIndex, exists := result[key]; !exists {
					t.Errorf("Expected column '%s' to be mapped", key)
				} else if actualIndex != expectedIndex {
					t.Errorf("Expected column '%s' to be at index %d, got %d", key, expectedIndex, actualIndex)
				}
			}
		})
	}
}

func TestFileProcessor_parseExcelRow(t *testing.T) {
	fp := NewFileProcessor("/test")

	columnMap := map[string]int{
		"titulo":              0,
		"descripcion":         1,
		"criterio_aceptacion": 2,
		"subtareas":           3,
		"parent":              4,
	}

	tests := []struct {
		name     string
		row      []string
		expected *CSVRecord
	}{
		{
			name: "complete_row",
			row:  []string{"Test Title", "Test Description", "Test Criteria", "Task 1;Task 2", "PROJ-123"},
			expected: &CSVRecord{
				Titulo:             "Test Title",
				Descripcion:        "Test Description",
				CriterioAceptacion: "Test Criteria",
				Subtareas:          "Task 1;Task 2",
				Parent:             "PROJ-123",
			},
		},
		{
			name: "partial_row",
			row:  []string{"Test Title", "Test Description"},
			expected: &CSVRecord{
				Titulo:             "Test Title",
				Descripcion:        "Test Description",
				CriterioAceptacion: "",
				Subtareas:          "",
				Parent:             "",
			},
		},
		{
			name: "row_with_spaces",
			row:  []string{" Test Title ", " Test Description ", " Test Criteria "},
			expected: &CSVRecord{
				Titulo:             "Test Title",
				Descripcion:        "Test Description",
				CriterioAceptacion: "Test Criteria",
				Subtareas:          "",
				Parent:             "",
			},
		},
		{
			name: "empty_row",
			row:  []string{},
			expected: &CSVRecord{
				Titulo:             "",
				Descripcion:        "",
				CriterioAceptacion: "",
				Subtareas:          "",
				Parent:             "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fp.parseExcelRow(tt.row, columnMap)

			if result.Titulo != tt.expected.Titulo {
				t.Errorf("Expected Titulo '%s', got '%s'", tt.expected.Titulo, result.Titulo)
			}
			if result.Descripcion != tt.expected.Descripcion {
				t.Errorf("Expected Descripcion '%s', got '%s'", tt.expected.Descripcion, result.Descripcion)
			}
			if result.CriterioAceptacion != tt.expected.CriterioAceptacion {
				t.Errorf("Expected CriterioAceptacion '%s', got '%s'", tt.expected.CriterioAceptacion, result.CriterioAceptacion)
			}
			if result.Subtareas != tt.expected.Subtareas {
				t.Errorf("Expected Subtareas '%s', got '%s'", tt.expected.Subtareas, result.Subtareas)
			}
			if result.Parent != tt.expected.Parent {
				t.Errorf("Expected Parent '%s', got '%s'", tt.expected.Parent, result.Parent)
			}
		})
	}
}

func TestFileProcessor_ValidateFile_FieldValidation(t *testing.T) {
	tempDir := t.TempDir()
	fp := NewFileProcessor(tempDir)

	tests := []struct {
		name        string
		content     string
		expectedErr string
	}{
		{
			name:        "file_with_only_headers",
			content:     `titulo,descripcion,criterio_aceptacion,subtareas,parent`,
			expectedErr: "file contains no valid stories",
		},
		{
			name: "valid_story_one_required_field",
			content: `titulo,descripcion,criterio_aceptacion,subtareas,parent
Test Title,Test Description,Test Criteria,,`,
			expectedErr: "",
		},
		{
			name: "xls_extension_parse_error",
			content: `titulo,descripcion,criterio_aceptacion,subtareas,parent
Test Story,Test Description,Test Criteria,Task 1,PROJ-123`,
			expectedErr: "error reading file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extension := ".csv"
			if tt.name == "xls_extension_parse_error" {
				extension = ".xls"
			}

			filePath := filepath.Join(tempDir, "test"+extension)
			err := os.WriteFile(filePath, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			err = fp.ValidateFile(context.Background(), filePath)

			if tt.expectedErr == "" {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.expectedErr)
				} else if !strings.Contains(err.Error(), tt.expectedErr) {
					t.Errorf("Expected error containing '%s', got: %v", tt.expectedErr, err)
				}
			}
		})
	}
}

func TestFileProcessor_readCSV_ErrorPaths(t *testing.T) {
	tempDir := t.TempDir()
	fp := NewFileProcessor(tempDir)

	tests := []struct {
		name        string
		setupFile   func() string
		expectedErr string
	}{
		{
			name: "file_open_error",
			setupFile: func() string {
				return "/nonexistent/file.csv"
			},
			expectedErr: "error opening CSV file",
		},
		{
			name: "malformed_csv_data",
			setupFile: func() string {
				content := `titulo,descripcion,criterio_aceptacion
"Unclosed quote,Bad data,More bad data`
				filePath := filepath.Join(tempDir, "malformed.csv")
				os.WriteFile(filePath, []byte(content), 0644)
				return filePath
			},
			expectedErr: "error parsing CSV",
		},
		{
			name: "csv_with_missing_required_fields",
			setupFile: func() string {
				content := `titulo,descripcion,criterio_aceptacion,subtareas,parent
Test Story,,Test Criteria,Task 1,PROJ-123
,Test Description,Test Criteria,Task 2,PROJ-124
Test Story 2,Test Description 2,,Task 3,PROJ-125`
				filePath := filepath.Join(tempDir, "missing_fields.csv")
				os.WriteFile(filePath, []byte(content), 0644)
				return filePath
			},
			expectedErr: "", // Should return empty stories array, not error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := tt.setupFile()

			stories, err := fp.readCSV(filePath)

			if tt.expectedErr == "" {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				// For missing fields test, should return 0 stories
				if tt.name == "csv_with_missing_required_fields" && len(stories) != 0 {
					t.Errorf("Expected 0 stories for missing fields, got %d", len(stories))
				}
			} else {
				if err == nil {
					t.Errorf("Expected error containing '%s', got no error", tt.expectedErr)
				} else if !strings.Contains(err.Error(), tt.expectedErr) {
					t.Errorf("Expected error containing '%s', got: %v", tt.expectedErr, err)
				}
			}
		})
	}
}

func TestFileProcessor_readExcel_ErrorPaths(t *testing.T) {
	tempDir := t.TempDir()
	fp := NewFileProcessor(tempDir)

	tests := []struct {
		name        string
		setupFile   func() string
		expectedErr string
	}{
		{
			name: "file_open_error",
			setupFile: func() string {
				return "/nonexistent/file.xlsx"
			},
			expectedErr: "error opening Excel file",
		},
		{
			name: "excel_file_with_insufficient_rows",
			setupFile: func() string {
				// Create a valid CSV file with .xlsx extension to trigger Excel reading
				content := `titulo`
				filePath := filepath.Join(tempDir, "insufficient.xlsx")
				os.WriteFile(filePath, []byte(content), 0644)
				return filePath
			},
			expectedErr: "error opening Excel file",
		},
		{
			name: "excel_with_validation_error",
			setupFile: func() string {
				// This will fail when trying to read as Excel file
				content := `titulo,descripcion,criterio_aceptacion
Invalid Excel Data`
				filePath := filepath.Join(tempDir, "invalid.xlsx")
				os.WriteFile(filePath, []byte(content), 0644)
				return filePath
			},
			expectedErr: "error opening Excel file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := tt.setupFile()

			stories, err := fp.readExcel(filePath)

			if tt.expectedErr == "" {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected error containing '%s', got no error", tt.expectedErr)
				} else if !strings.Contains(err.Error(), tt.expectedErr) {
					t.Errorf("Expected error containing '%s', got: %v", tt.expectedErr, err)
				}
			}
			if err != nil && stories != nil {
				t.Error("Expected nil stories when error occurs")
			}
		})
	}
}
