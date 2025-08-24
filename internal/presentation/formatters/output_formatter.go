package formatters

import (
	"fmt"
	"strings"
	"time"

	"historiadorgo/internal/application/usecases"
	"historiadorgo/internal/domain/entities"
)

type OutputFormatter struct{}

func NewOutputFormatter() *OutputFormatter {
	return &OutputFormatter{}
}

func (of *OutputFormatter) FormatBatchResult(result *entities.BatchResult) string {
	var output strings.Builder

	output.WriteString(of.formatHeader(result))
	output.WriteString(of.formatSummary(result))

	if result.HasValidationErrors() {
		output.WriteString(of.formatValidationErrors(result))
	}

	if len(result.Results) > 0 {
		output.WriteString(of.formatProcessResults(result))
	}

	if result.HasErrors() {
		output.WriteString(of.formatErrors(result))
	}

	output.WriteString(of.formatFooter(result))

	return output.String()
}

func (of *OutputFormatter) FormatMultipleBatchResults(results []*entities.BatchResult) string {
	var output strings.Builder

	output.WriteString("=== RESUMEN GENERAL ===\n\n")

	totalFiles := len(results)
	totalProcessed := 0
	totalSuccessful := 0
	totalErrors := 0

	for _, result := range results {
		totalProcessed += result.ProcessedRows
		totalSuccessful += result.SuccessfulRows
		totalErrors += result.ErrorRows
	}

	output.WriteString(fmt.Sprintf("Archivos procesados: %d\n", totalFiles))
	output.WriteString(fmt.Sprintf("Historias procesadas: %d\n", totalProcessed))
	output.WriteString(fmt.Sprintf("[OK] Historias exitosas: %d\n", totalSuccessful))
	output.WriteString(fmt.Sprintf("[ERROR] Historias con errores: %d\n", totalErrors))

	if totalProcessed > 0 {
		successRate := float64(totalSuccessful) / float64(totalProcessed) * 100
		output.WriteString(fmt.Sprintf("Tasa de exito: %.1f%%\n", successRate))
	}

	output.WriteString("\n" + strings.Repeat("=", 50) + "\n\n")

	for i, result := range results {
		output.WriteString(fmt.Sprintf("=== ARCHIVO %d/%d ===\n", i+1, totalFiles))
		output.WriteString(of.FormatBatchResult(result))
		output.WriteString("\n")
	}

	return output.String()
}

func (of *OutputFormatter) FormatConnectionTest(err error) string {
	if err != nil {
		return fmt.Sprintf("[ERROR] Prueba de conexion fallida: %v\n", err)
	}
	return "[OK] Conexion con Jira exitosa\n"
}

func (of *OutputFormatter) FormatValidation(filePath string, validationResult *usecases.ValidationResult, err error) string {
	var output strings.Builder

	output.WriteString("=== VALIDACION DE ARCHIVO ===\n\n")
	output.WriteString(fmt.Sprintf("Archivo: %s\n", filePath))

	if err != nil {
		output.WriteString(fmt.Sprintf("[ERROR] Validacion fallida: %v\n", err))
		return output.String()
	}

	output.WriteString("[OK] Validacion exitosa\n\n")

	if validationResult != nil {
		output.WriteString("=== ESTADISTICAS ===\n")
		output.WriteString(fmt.Sprintf("Total de historias: %d\n", validationResult.TotalStories))
		output.WriteString(fmt.Sprintf("Con subtareas: %d\n", validationResult.WithSubtasks))
		output.WriteString(fmt.Sprintf("Total subtareas: %d\n", validationResult.TotalSubtasks))
		output.WriteString(fmt.Sprintf("Con parent: %d\n", validationResult.WithParent))

		if validationResult.InvalidSubtasks > 0 {
			output.WriteString(fmt.Sprintf("[WARNING] Subtareas invalidas: %d\n", validationResult.InvalidSubtasks))
		}

		output.WriteString("\n")

		if validationResult.Preview != "" {
			output.WriteString("=== PREVIEW (primeras 5 filas) ===\n")
			output.WriteString(validationResult.Preview)
			output.WriteString("\n\n")
		}

		output.WriteString("=== VALIDACIONES REALIZADAS ===\n")
		output.WriteString("[OK] Formato de archivo valido\n")
		output.WriteString("[OK] Columnas requeridas presentes\n")
		output.WriteString("[OK] Datos estructurales validos\n")
		if validationResult.InvalidSubtasks == 0 {
			output.WriteString("[OK] Todas las subtareas son validas\n")
		}
	}

	output.WriteString(strings.Repeat("=", 50) + "\n")

	return output.String()
}

func (of *OutputFormatter) FormatDiagnosis(requiredFields []string) string {
	var output strings.Builder

	output.WriteString("=== DIAGNÓSTICO DE FEATURES ===\n\n")

	if len(requiredFields) == 0 {
		output.WriteString("✅ No se requieren campos adicionales para crear Features\n")
	} else {
		output.WriteString("⚠️  Se requieren los siguientes campos para crear Features:\n\n")
		for _, field := range requiredFields {
			output.WriteString(fmt.Sprintf("  • %s\n", field))
		}
		output.WriteString("\n💡 Configura estos campos en FEATURE_REQUIRED_FIELDS como JSON en tu .env\n")
		output.WriteString("   Ejemplo: FEATURE_REQUIRED_FIELDS='{\"customfield_10100\":{\"value\":\"Medium\"}}'\n")
	}

	return output.String()
}

func (of *OutputFormatter) FormatDiagnosisNoProject() string {
	var output strings.Builder

	output.WriteString("=== DIAGNÓSTICO DE FEATURES ===\n\n")
	output.WriteString("ℹ️  Para diagnosticar configuración de Features se requiere un proyecto\n\n")
	output.WriteString("Opciones:\n")
	output.WriteString("  • Usar flag: historiador diagnose -p PROYECTO\n")
	output.WriteString("  • Configurar en .env: PROJECT_KEY=PROYECTO\n\n")
	output.WriteString("El diagnóstico verificará qué campos son obligatorios para crear Features automáticamente.\n")

	return output.String()
}

func (of *OutputFormatter) formatHeader(result *entities.BatchResult) string {
	var output strings.Builder

	output.WriteString("=== PROCESAMIENTO DE ARCHIVO ===\n\n")
	output.WriteString(fmt.Sprintf("Archivo: %s\n", result.FileName))

	if result.DryRun {
		output.WriteString("MODO DE PRUEBA (DRY-RUN)\n")
	}

	output.WriteString(fmt.Sprintf("Inicio: %s\n", result.StartTime.Format("2006-01-02 15:04:05")))
	output.WriteString(fmt.Sprintf("Duracion: %v\n", result.Duration.Round(time.Millisecond)))
	output.WriteString("\n")

	return output.String()
}

func (of *OutputFormatter) formatSummary(result *entities.BatchResult) string {
	var output strings.Builder

	output.WriteString("=== RESUMEN ===\n")
	output.WriteString(fmt.Sprintf("Total de filas: %d\n", result.TotalRows))
	output.WriteString(fmt.Sprintf("Filas procesadas: %d\n", result.ProcessedRows))
	output.WriteString(fmt.Sprintf("[OK] Exitosas: %d\n", result.SuccessfulRows))
	output.WriteString(fmt.Sprintf("[ERROR] Con errores: %d\n", result.ErrorRows))

	if result.SkippedRows > 0 {
		output.WriteString(fmt.Sprintf("Saltadas: %d\n", result.SkippedRows))
	}

	if result.ProcessedRows > 0 {
		successRate := result.GetSuccessRate()
		output.WriteString(fmt.Sprintf("Tasa de exito: %.1f%%\n", successRate))
	}

	output.WriteString("\n")

	return output.String()
}

func (of *OutputFormatter) formatProcessResults(result *entities.BatchResult) string {
	var output strings.Builder

	output.WriteString("=== DETALLE DE PROCESAMIENTO ===\n")

	for _, processResult := range result.Results {
		if processResult.Success {
			output.WriteString(fmt.Sprintf("[OK] Fila %d: %s\n", processResult.RowNumber, processResult.IssueKey))

			if len(processResult.Subtareas) > 0 {
				output.WriteString(of.formatSubtasks(processResult.Subtareas))
			}
		} else {
			output.WriteString(fmt.Sprintf("[ERROR] Fila %d: %s\n", processResult.RowNumber, processResult.ErrorMessage))
		}
	}

	output.WriteString("\n")

	return output.String()
}

func (of *OutputFormatter) formatSubtasks(subtasks []*entities.SubtaskResult) string {
	var output strings.Builder

	for _, subtask := range subtasks {
		if subtask.Success {
			output.WriteString(fmt.Sprintf("   [OK] Subtarea: %s (%s)\n", subtask.Description, subtask.IssueKey))
		} else {
			output.WriteString(fmt.Sprintf("   [ERROR] Subtarea fallida: %s - %s\n", subtask.Description, subtask.Error))
		}
	}

	return output.String()
}

func (of *OutputFormatter) formatValidationErrors(result *entities.BatchResult) string {
	var output strings.Builder

	output.WriteString("=== ERRORES DE VALIDACION ===\n")

	for _, error := range result.ValidationErrors {
		output.WriteString(fmt.Sprintf("[WARNING] %s\n", error))
	}

	output.WriteString("\n")

	return output.String()
}

func (of *OutputFormatter) formatErrors(result *entities.BatchResult) string {
	var output strings.Builder

	output.WriteString("=== ERRORES ===\n")

	for _, error := range result.Errors {
		output.WriteString(fmt.Sprintf("[ERROR] %s\n", error))
	}

	output.WriteString("\n")

	return output.String()
}

func (of *OutputFormatter) formatFooter(result *entities.BatchResult) string {
	var output strings.Builder

	output.WriteString(strings.Repeat("=", 50) + "\n")

	if result.IsSuccessful() {
		output.WriteString("[OK] Procesamiento completado exitosamente\n")

		issues := result.GetProcessedIssues()
		if len(issues) > 0 {
			output.WriteString(fmt.Sprintf("Issues creados: %s\n", strings.Join(issues, ", ")))
		}
	} else if result.HasErrors() {
		output.WriteString("[ERROR] Procesamiento completado con errores\n")
	} else {
		output.WriteString("[WARNING] No se procesaron historias\n")
	}

	return output.String()
}
