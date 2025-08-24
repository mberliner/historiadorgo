package formatters

import (
	"errors"
	"strings"
	"testing"
	"time"

	"historiadorgo/internal/application/usecases"
	"historiadorgo/internal/domain/entities"
)

func TestOutputFormatter_FormatBatchResult(t *testing.T) {
	formatter := NewOutputFormatter()

	// Create sample batch result
	batchResult := entities.NewBatchResult("test.csv", 5, false)

	// Add some results (this will update counters automatically)
	successResult := entities.NewProcessResult(1)
	successResult.Success = true
	successResult.IssueKey = "PROJ-123"
	batchResult.AddResult(successResult)

	errorResult := entities.NewProcessResult(2)
	errorResult.Success = false
	errorResult.ErrorMessage = "Test error"
	batchResult.AddResult(errorResult)

	batchResult.Finish()

	output := formatter.FormatBatchResult(batchResult)

	// Verify output contains expected sections
	expectedSections := []string{
		"=== PROCESAMIENTO DE ARCHIVO ===",
		"Archivo: test.csv",
		"=== RESUMEN ===",
		"Total de filas: 5",
		"Filas procesadas: 2",    // 2 results added
		"[OK] Exitosas: 1",       // 1 successful
		"[ERROR] Con errores: 1", // 1 error
		"=== DETALLE DE PROCESAMIENTO ===",
		"[OK] Fila 1: PROJ-123",
		"[ERROR] Fila 2: Test error",
	}

	for _, section := range expectedSections {
		if !strings.Contains(output, section) {
			t.Errorf("Output should contain %q, got: %s", section, output)
		}
	}
}

func TestOutputFormatter_FormatBatchResult_DryRun(t *testing.T) {
	formatter := NewOutputFormatter()

	batchResult := entities.NewBatchResult("test.csv", 2, true) // dry run = true
	batchResult.Finish()

	output := formatter.FormatBatchResult(batchResult)

	if !strings.Contains(output, "MODO DE PRUEBA (DRY-RUN)") {
		t.Errorf("Output should contain dry run indicator, got: %s", output)
	}
}

func TestOutputFormatter_FormatMultipleBatchResults(t *testing.T) {
	formatter := NewOutputFormatter()

	// Create multiple batch results
	result1 := entities.NewBatchResult("file1.csv", 3, false)
	result1.SuccessfulRows = 2
	result1.ErrorRows = 1
	result1.Finish()

	result2 := entities.NewBatchResult("file2.csv", 2, false)
	result2.SuccessfulRows = 2
	result2.ErrorRows = 0
	result2.Finish()

	results := []*entities.BatchResult{result1, result2}
	output := formatter.FormatMultipleBatchResults(results)

	expectedSections := []string{
		"=== RESUMEN GENERAL ===",
		"Archivos procesados: 2",
		"Historias procesadas: 0",          // No results were added
		"[OK] Historias exitosas: 4",       // Manual count from SuccessfulRows
		"[ERROR] Historias con errores: 1", // Manual count from ErrorRows
	}

	for _, section := range expectedSections {
		if !strings.Contains(output, section) {
			t.Errorf("Output should contain %q, got: %s", section, output)
		}
	}
}

func TestOutputFormatter_FormatValidation(t *testing.T) {
	formatter := NewOutputFormatter()

	validationResult := &usecases.ValidationResult{
		TotalStories:    3,
		WithSubtasks:    2,
		TotalSubtasks:   5,
		WithParent:      1,
		InvalidSubtasks: 1,
		Preview:         "Sample preview",
	}

	output := formatter.FormatValidation("test.csv", validationResult, nil)

	expectedSections := []string{
		"=== VALIDACION DE ARCHIVO ===",
		"Archivo: test.csv",
		"[OK] Validacion exitosa",
		"Total de historias: 3",
		"Con subtareas: 2",
		"Total subtareas: 5",
		"Con parent: 1",
		"Subtareas invalidas: 1",
		"Sample preview",
	}

	for _, section := range expectedSections {
		if !strings.Contains(output, section) {
			t.Errorf("Output should contain %q, got: %s", section, output)
		}
	}
}

func TestOutputFormatter_FormatValidation_WithError(t *testing.T) {
	formatter := NewOutputFormatter()

	err := errors.New("validation failed")
	output := formatter.FormatValidation("test.csv", nil, err)

	expectedSections := []string{
		"=== VALIDACION DE ARCHIVO ===",
		"Archivo: test.csv",
		"[ERROR] Validacion fallida: validation failed",
	}

	for _, section := range expectedSections {
		if !strings.Contains(output, section) {
			t.Errorf("Output should contain %q, got: %s", section, output)
		}
	}
}

func TestOutputFormatter_FormatConnectionTest(t *testing.T) {
	formatter := NewOutputFormatter()

	// Test successful connection
	output := formatter.FormatConnectionTest(nil)

	expectedSections := []string{
		"[OK] Conexion con Jira exitosa",
	}

	for _, section := range expectedSections {
		if !strings.Contains(output, section) {
			t.Errorf("Output should contain %q, got: %s", section, output)
		}
	}

	// Test failed connection
	err := errors.New("connection failed")
	output = formatter.FormatConnectionTest(err)

	expectedSections = []string{
		"[ERROR] Prueba de conexion fallida: connection failed",
	}

	for _, section := range expectedSections {
		if !strings.Contains(output, section) {
			t.Errorf("Output should contain %q, got: %s", section, output)
		}
	}
}

func TestOutputFormatter_FormatDiagnosis(t *testing.T) {
	formatter := NewOutputFormatter()

	// Test with required fields
	requiredFields := []string{"customfield_10001", "reporter"}
	output := formatter.FormatDiagnosis(requiredFields)

	expectedSections := []string{
		"=== DIAGNÓSTICO DE FEATURES ===",
		"⚠️  Se requieren los siguientes campos para crear Features:",
		"• customfield_10001",
		"• reporter",
		"💡 Configura estos campos en FEATURE_REQUIRED_FIELDS",
	}

	for _, section := range expectedSections {
		if !strings.Contains(output, section) {
			t.Errorf("Output should contain %q, got: %s", section, output)
		}
	}

	// Test with no required fields
	output = formatter.FormatDiagnosis([]string{})

	if !strings.Contains(output, "✅ No se requieren campos adicionales") {
		t.Errorf("Output should contain no additional fields message, got: %s", output)
	}
}

func TestOutputFormatter_FormatDiagnosisNoProject(t *testing.T) {
	formatter := NewOutputFormatter()

	output := formatter.FormatDiagnosisNoProject()

	expectedSections := []string{
		"=== DIAGNÓSTICO DE FEATURES ===",
		"ℹ️  Para diagnosticar configuración de Features se requiere un proyecto",
		"Opciones:",
		"• Usar flag: historiador diagnose -p PROYECTO",
		"• Configurar en .env: PROJECT_KEY=PROYECTO",
	}

	for _, section := range expectedSections {
		if !strings.Contains(output, section) {
			t.Errorf("Output should contain %q, got: %s", section, output)
		}
	}
}

func TestOutputFormatter_FormatHeader(t *testing.T) {
	formatter := NewOutputFormatter()

	batchResult := entities.NewBatchResult("test-file.csv", 10, false)
	batchResult.Duration = 2500 * time.Millisecond

	// Use reflection or test through FormatBatchResult since formatHeader is private
	output := formatter.FormatBatchResult(batchResult)

	expectedSections := []string{
		"=== PROCESAMIENTO DE ARCHIVO ===",
		"Archivo: test-file.csv",
		"Duracion: 2.5s", // Should be rounded to nearest second/millisecond
	}

	for _, section := range expectedSections {
		if !strings.Contains(output, section) {
			t.Errorf("Output should contain %q, got: %s", section, output)
		}
	}
}

func TestOutputFormatter_ProcessResults_WithSubtasks(t *testing.T) {
	formatter := NewOutputFormatter()

	batchResult := entities.NewBatchResult("test.csv", 1, false)

	// Create result with subtasks
	result := entities.NewProcessResult(1)
	result.Success = true
	result.IssueKey = "PROJ-123"
	result.AddSubtaskResult("Subtask 1", true, "PROJ-124", "url1", "")
	result.AddSubtaskResult("Subtask 2", false, "", "", "Error message")

	batchResult.AddResult(result)
	batchResult.Finish()

	output := formatter.FormatBatchResult(batchResult)

	expectedSections := []string{
		"[OK] Fila 1: PROJ-123",
		"   [OK] Subtarea: Subtask 1 (PROJ-124)",
		"   [ERROR] Subtarea fallida: Subtask 2 - Error message",
	}

	for _, section := range expectedSections {
		if !strings.Contains(output, section) {
			t.Errorf("Output should contain %q, got: %s", section, output)
		}
	}
}
