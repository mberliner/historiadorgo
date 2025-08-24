package filesystem

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/xuri/excelize/v2"
)

func createTestExcelFile(filePath string, headers []string, rows [][]string) error {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Sheet1"

	// Escribir headers
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, header)
	}

	// Escribir datos
	for rowIdx, row := range rows {
		for colIdx, value := range row {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			f.SetCellValue(sheetName, cell, value)
		}
	}

	return f.SaveAs(filePath)
}

func TestFileProcessor_ReadExcel_ValidFile(t *testing.T) {
	tempDir := t.TempDir()
	fp := NewFileProcessor(tempDir)

	// Crear archivo Excel de prueba
	excelPath := filepath.Join(tempDir, "test.xlsx")
	headers := []string{"titulo", "descripcion", "criterio_aceptacion", "subtareas", "parent"}
	rows := [][]string{
		{"Story 1", "Description 1", "Criteria 1", "Task 1;Task 2", "PROJ-123"},
		{"Story 2", "Description 2", "Criteria 2", "", ""},
		{"Story 3", "Description 3", "Criteria 3", "Task A\nTask B", "Feature Description"},
	}

	err := createTestExcelFile(excelPath, headers, rows)
	if err != nil {
		t.Fatalf("Failed to create test Excel file: %v", err)
	}

	stories, err := fp.ReadFile(context.Background(), excelPath)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	expectedCount := 3
	if len(stories) != expectedCount {
		t.Errorf("Expected %d stories, got %d", expectedCount, len(stories))
	}

	// Verificar primera historia
	if stories[0].Titulo != "Story 1" {
		t.Errorf("Expected title 'Story 1', got '%s'", stories[0].Titulo)
	}

	if len(stories[0].Subtareas) != 2 {
		t.Errorf("Expected 2 subtasks, got %d", len(stories[0].Subtareas))
	}

	if stories[0].Parent != "PROJ-123" {
		t.Errorf("Expected parent 'PROJ-123', got '%s'", stories[0].Parent)
	}

	// Verificar tercera historia con subtareas separadas por newline
	if len(stories[2].Subtareas) != 2 {
		t.Errorf("Expected 2 subtasks for story 3, got %d", len(stories[2].Subtareas))
	}
}

func TestFileProcessor_ReadExcel_InvalidHeaders(t *testing.T) {
	tempDir := t.TempDir()
	fp := NewFileProcessor(tempDir)

	// Crear archivo Excel con headers incorrectos
	excelPath := filepath.Join(tempDir, "invalid_headers.xlsx")
	headers := []string{"wrong_header", "another_wrong", "invalid"}
	rows := [][]string{
		{"Value 1", "Value 2", "Value 3"},
	}

	err := createTestExcelFile(excelPath, headers, rows)
	if err != nil {
		t.Fatalf("Failed to create test Excel file: %v", err)
	}

	stories, err := fp.ReadFile(context.Background(), excelPath)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Debería retornar 0 historias porque no encuentra las columnas requeridas
	if len(stories) != 0 {
		t.Errorf("Expected 0 stories with invalid headers, got %d", len(stories))
	}
}

func TestFileProcessor_ReadExcel_EmptyFile(t *testing.T) {
	tempDir := t.TempDir()
	fp := NewFileProcessor(tempDir)

	// Crear archivo Excel vacío (solo headers)
	excelPath := filepath.Join(tempDir, "empty.xlsx")
	headers := []string{"titulo", "descripcion", "criterio_aceptacion"}
	rows := [][]string{} // Sin datos

	err := createTestExcelFile(excelPath, headers, rows)
	if err != nil {
		t.Fatalf("Failed to create test Excel file: %v", err)
	}

	_, err = fp.ReadFile(context.Background(), excelPath)
	if err == nil {
		t.Error("Expected error for empty file, got none")
	}

	if !containsString(err.Error(), "must have at least a header row and one data row") {
		t.Errorf("Expected row count error, got: %v", err)
	}
}

func TestFileProcessor_ReadExcel_ValidationErrors(t *testing.T) {
	tempDir := t.TempDir()
	fp := NewFileProcessor(tempDir)

	// Crear archivo Excel con datos que fallan validación
	excelPath := filepath.Join(tempDir, "validation_error.xlsx")
	headers := []string{"titulo", "descripcion", "criterio_aceptacion"}
	rows := [][]string{
		{"", "Description without title", "Criteria"}, // Título vacío - fallará validación
		{"Valid Title", "Valid Description", "Valid Criteria"},
	}

	err := createTestExcelFile(excelPath, headers, rows)
	if err != nil {
		t.Fatalf("Failed to create test Excel file: %v", err)
	}

	stories, err := fp.ReadFile(context.Background(), excelPath)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// La validación ocurre al procesar las historias, no al leer el archivo
	// Verificar que se obtuvieron historias pero algunas pueden fallar validación interna
	if len(stories) == 0 {
		t.Error("Expected at least one story to be returned")
	}
}

func TestFileProcessor_ReadExcel_PartialData(t *testing.T) {
	tempDir := t.TempDir()
	fp := NewFileProcessor(tempDir)

	// Crear archivo Excel con filas de diferentes longitudes
	excelPath := filepath.Join(tempDir, "partial_data.xlsx")
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Sheet1"

	// Headers completos
	headers := []string{"titulo", "descripcion", "criterio_aceptacion", "subtareas", "parent"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, header)
	}

	// Fila con datos parciales (solo primeras 3 columnas)
	f.SetCellValue(sheetName, "A2", "Story 1")
	f.SetCellValue(sheetName, "B2", "Description 1")
	f.SetCellValue(sheetName, "C2", "Criteria 1")
	// D2 y E2 quedan vacías

	// Fila completa
	f.SetCellValue(sheetName, "A3", "Story 2")
	f.SetCellValue(sheetName, "B3", "Description 2")
	f.SetCellValue(sheetName, "C3", "Criteria 2")
	f.SetCellValue(sheetName, "D3", "Task 1")
	f.SetCellValue(sheetName, "E3", "PROJ-123")

	err := f.SaveAs(excelPath)
	if err != nil {
		t.Fatalf("Failed to create test Excel file: %v", err)
	}

	stories, err := fp.ReadFile(context.Background(), excelPath)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if len(stories) != 2 {
		t.Errorf("Expected 2 stories, got %d", len(stories))
	}

	// Primera historia sin subtareas ni parent
	if len(stories[0].Subtareas) != 0 {
		t.Errorf("Expected no subtasks for first story, got %d", len(stories[0].Subtareas))
	}

	if stories[0].Parent != "" {
		t.Errorf("Expected empty parent for first story, got '%s'", stories[0].Parent)
	}

	// Segunda historia con subtareas y parent
	if len(stories[1].Subtareas) != 1 {
		t.Errorf("Expected 1 subtask for second story, got %d", len(stories[1].Subtareas))
	}

	if stories[1].Parent != "PROJ-123" {
		t.Errorf("Expected parent 'PROJ-123' for second story, got '%s'", stories[1].Parent)
	}
}

func TestFileProcessor_ReadExcel_NonExistentFile(t *testing.T) {
	tempDir := t.TempDir()
	fp := NewFileProcessor(tempDir)

	nonExistentPath := filepath.Join(tempDir, "nonexistent.xlsx")

	_, err := fp.ReadFile(context.Background(), nonExistentPath)
	if err == nil {
		t.Error("Expected error for non-existent file, got none")
	}

	if !containsString(err.Error(), "error opening Excel file") {
		t.Errorf("Expected file opening error, got: %v", err)
	}
}

func TestFileProcessor_ReadExcel_CorruptedFile(t *testing.T) {
	tempDir := t.TempDir()
	fp := NewFileProcessor(tempDir)

	// Crear archivo corrupto (no es Excel válido)
	corruptPath := filepath.Join(tempDir, "corrupt.xlsx")
	err := os.WriteFile(corruptPath, []byte("This is not an Excel file"), 0644)
	if err != nil {
		t.Fatalf("Failed to create corrupt file: %v", err)
	}

	_, err = fp.ReadFile(context.Background(), corruptPath)
	if err == nil {
		t.Error("Expected error for corrupted file, got none")
	}

	if !containsString(err.Error(), "error opening Excel file") {
		t.Errorf("Expected file opening error, got: %v", err)
	}
}

func TestFileProcessor_ReadExcel_NoDataRows(t *testing.T) {
	tempDir := t.TempDir()
	fp := NewFileProcessor(tempDir)

	// Crear archivo Excel con solo header pero sin datos
	excelPath := filepath.Join(tempDir, "header_only.xlsx")
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Sheet1"
	f.SetCellValue(sheetName, "A1", "titulo")
	f.SetCellValue(sheetName, "B1", "descripcion")

	err := f.SaveAs(excelPath)
	if err != nil {
		t.Fatalf("Failed to create test Excel file: %v", err)
	}

	_, err = fp.ReadFile(context.Background(), excelPath)
	if err == nil {
		t.Error("Expected error for file with insufficient rows, got none")
	}

	if !containsString(err.Error(), "must have at least a header row and one data row") {
		t.Errorf("Expected insufficient rows error, got: %v", err)
	}
}

// Helper function para verificar si una cadena contiene otra
func containsString(haystack, needle string) bool {
	return len(needle) == 0 || (len(haystack) >= len(needle) &&
		haystack[:len(needle)] == needle) ||
		(len(haystack) > len(needle) &&
			containsString(haystack[1:], needle))
}

// Implementación simple de contains para strings
func contains(s, substr string) bool {
	return len(substr) == 0 || len(s) >= len(substr) && (s == substr || s[0:len(substr)] == substr || (len(s) > len(substr) && contains(s[1:], substr)))
}
