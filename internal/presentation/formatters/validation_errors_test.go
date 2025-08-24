package formatters

import (
	"strings"
	"testing"

	"historiadorgo/internal/domain/entities"
)

func TestOutputFormatter_formatValidationErrors(t *testing.T) {
	formatter := NewOutputFormatter()

	// Crear BatchResult con errores de validación
	result := entities.NewBatchResult("test-file.csv", 3, false)
	result.AddValidationError("Fila 1: Título requerido")
	result.AddValidationError("Fila 3: Descripción demasiado larga")

	// Como formatValidationErrors es un método privado, vamos a probarlo a través de FormatBatchResult
	output := formatter.FormatBatchResult(result)

	// Verificar que contiene la sección de errores de validación
	if !strings.Contains(output, "ERRORES DE VALIDACION") {
		t.Error("Output should contain validation errors section")
	}

	if !strings.Contains(output, "Fila 1: Título requerido") {
		t.Error("Output should contain first validation error")
	}

	if !strings.Contains(output, "Fila 3: Descripción demasiado larga") {
		t.Error("Output should contain second validation error")
	}
}

func TestOutputFormatter_BatchResultWithValidationErrors(t *testing.T) {
	formatter := NewOutputFormatter()

	// Crear BatchResult con errores de validación y resultados de proceso
	result := entities.NewBatchResult("test-file.csv", 2, false)
	result.AddValidationError("Campo requerido faltante")
	result.AddValidationError("Formato inválido")

	// Agregar también un resultado exitoso
	processResult := entities.NewProcessResult(1)
	processResult.Success = true
	processResult.IssueKey = "PROJ-123"
	processResult.IssueURL = "https://test.com/PROJ-123"
	result.AddResult(processResult)

	result.Finish()

	output := formatter.FormatBatchResult(result)

	// Verificar que incluye ambas secciones
	expectedSections := []string{
		"ERRORES DE VALIDACION",
		"Campo requerido faltante",
		"Formato inválido",
		"DETALLE DE PROCESAMIENTO",
		"[OK] Fila 1: PROJ-123",
	}

	for _, section := range expectedSections {
		if !strings.Contains(output, section) {
			t.Errorf("Output should contain %q", section)
		}
	}
}

func TestOutputFormatter_formatValidationErrorsOnly(t *testing.T) {
	formatter := NewOutputFormatter()

	// Crear BatchResult solo con errores de validación
	result := entities.NewBatchResult("test-file.csv", 3, false)
	result.AddValidationError("Error de validación 1")
	result.AddValidationError("Error de validación 2")
	result.AddValidationError("Error de validación 3")

	result.Finish()

	output := formatter.FormatBatchResult(result)

	// Verificar estructura del output
	if !strings.Contains(output, "ERRORES DE VALIDACION") {
		t.Error("Output should contain validation errors header")
	}

	if !strings.Contains(output, "[WARNING] Error de validación 1") {
		t.Error("Output should format validation errors with WARNING prefix")
	}

	if !strings.Contains(output, "[WARNING] Error de validación 2") {
		t.Error("Output should format all validation errors")
	}

	if !strings.Contains(output, "[WARNING] Error de validación 3") {
		t.Error("Output should format all validation errors")
	}

	// Verificar que no contiene sección de resultados si no hay resultados de proceso
	if strings.Contains(output, "DETALLE DE PROCESAMIENTO") {
		t.Error("Output should not contain process results section when there are no process results")
	}
}
