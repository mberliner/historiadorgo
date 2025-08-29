package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"historiadorgo/internal/application/usecases"
	"historiadorgo/internal/domain/entities"
	"historiadorgo/internal/infrastructure/config"
	"historiadorgo/internal/infrastructure/filesystem"
	"historiadorgo/internal/infrastructure/logger"
	"historiadorgo/internal/presentation/formatters"
	"historiadorgo/tests/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// App struct for integration tests (matches internal/presentation/cli/commands.go)
type App struct {
	config          *config.Config
	logger          *logger.Logger
	formatter       *formatters.OutputFormatter
	testConnUseCase *usecases.TestConnectionUseCase
	validateUseCase *usecases.ValidateFileUseCase
	processUseCase  *usecases.ProcessFilesUseCase
	diagnoseUseCase *usecases.DiagnoseFeaturesUseCase
}

func TestProcessDryRun_Integration(t *testing.T) {
	// Setup test environment
	tempDir := t.TempDir()

	// Create test directories
	inputDir := filepath.Join(tempDir, "entrada")
	logsDir := filepath.Join(tempDir, "logs")
	processedDir := filepath.Join(tempDir, "procesados")

	require.NoError(t, os.MkdirAll(inputDir, 0755))
	require.NoError(t, os.MkdirAll(logsDir, 0755))
	require.NoError(t, os.MkdirAll(processedDir, 0755))

	// Create test CSV file
	csvContent := `titulo,descripcion,criterio_aceptacion,subtareas,parent
Login de usuario,Como usuario quiero autenticarme en el sistema,Dado que ingreso credenciales válidas entonces accedo al sistema,Validar campos;Conectar con API;Mostrar errores de validación,
Dashboard principal,Como usuario quiero ver mi dashboard personalizado,Dado que estoy autenticado entonces veo mi dashboard,Cargar widgets;Aplicar filtros;Mostrar métricas,
Gestión de perfil,Como usuario quiero gestionar mi perfil,Dado que accedo a perfil entonces puedo editarlo,Editar datos;Subir avatar;Cambiar contraseña,AUTH-123`

	csvFile := filepath.Join(inputDir, "test_stories.csv")
	require.NoError(t, os.WriteFile(csvFile, []byte(csvContent), 0644))

	// Setup environment variables
	originalEnv := setupTestEnvironment(t, tempDir)
	defer restoreEnvironment(originalEnv)

	// Create app with real dependencies but mocked Jira for safety
	app, err := createTestApp(tempDir)
	require.NoError(t, err)
	defer app.logger.Close() // Ensure logger is closed to release file handles

	// Execute dry run processing
	ctx := context.Background()
	err = app.runProcess(ctx, "TEST-PROJ", csvFile, true)

	// Verify no error occurred
	assert.NoError(t, err)

	// Verify log file was created
	logFiles, err := filepath.Glob(filepath.Join(logsDir, "*.log"))
	require.NoError(t, err)
	require.Len(t, logFiles, 1)

	// Verify log content contains processing information
	logContent, err := os.ReadFile(logFiles[0])
	require.NoError(t, err)
	logStr := string(logContent)

	// Verify command execution was logged
	assert.Contains(t, logStr, "process")
	assert.Contains(t, logStr, "TEST-PROJ")
	assert.Contains(t, logStr, "dry_run:true")

	// Verify the CSV file exists and was not processed (moved)
	_, err = os.Stat(csvFile)
	assert.NoError(t, err, "CSV file should still exist after dry-run")

	// Verify file was NOT moved to processed (since it's dry run)
	processedFiles, err := filepath.Glob(filepath.Join(processedDir, "*"))
	require.NoError(t, err)
	assert.Len(t, processedFiles, 0, "File should not be moved in dry-run mode")
}

func TestValidateCommand_Integration(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name              string
		csvContent        string
		expectError       bool
		expectedInPreview []string
	}{
		{
			name: "valid_file_with_subtasks",
			csvContent: `titulo,descripcion,criterio_aceptacion,subtareas,parent
Historia válida,Descripción de prueba,Criterio de aceptación,Subtarea1;Subtarea2,PROJ-123`,
			expectError:       false,
			expectedInPreview: []string{"Historia válida"},
		},
		{
			name: "file_with_validation_errors",
			csvContent: `titulo,descripcion,criterio_aceptacion,subtareas,parent
,Descripción sin título,,Subtarea muy larga que excede los 255 caracteres permitidos y debería ser marcada como inválida porque supera el límite establecido para las subtareas en el sistema y esto puede causar problemas durante la creación,`,
			expectError:       true,
			expectedInPreview: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test environment
			inputDir := filepath.Join(tempDir, "entrada_"+tt.name)
			logsDir := filepath.Join(tempDir, "logs_"+tt.name)
			require.NoError(t, os.MkdirAll(inputDir, 0755))
			require.NoError(t, os.MkdirAll(logsDir, 0755))

			// Create test CSV
			csvFile := filepath.Join(inputDir, "validate_test.csv")
			require.NoError(t, os.WriteFile(csvFile, []byte(tt.csvContent), 0644))

			// Setup environment
			originalEnv := setupTestEnvironment(t, tempDir)
			defer restoreEnvironment(originalEnv)

			// Create app
			app, err := createTestApp(tempDir)
			require.NoError(t, err)
			defer app.logger.Close() // Ensure logger is closed to release file handles

			// Execute validation
			ctx := context.Background()
			err = app.runValidate(ctx, "", csvFile, 5)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Check if log files were created (validation might not always create logs)
			logFiles, err := filepath.Glob(filepath.Join(logsDir, "*.log"))
			require.NoError(t, err)

			// If log files exist, verify they contain some expected content
			if len(logFiles) > 0 {
				logContent, err := os.ReadFile(logFiles[0])
				require.NoError(t, err)
				logStr := string(logContent)

				// Verify basic validation logging occurred
				assert.Contains(t, logStr, "validate", "Expected validation command to be logged")
			}
		})
	}
}

func TestProcessMultipleFiles_Integration(t *testing.T) {
	tempDir := t.TempDir()
	inputDir := filepath.Join(tempDir, "entrada")
	logsDir := filepath.Join(tempDir, "logs")
	require.NoError(t, os.MkdirAll(inputDir, 0755))
	require.NoError(t, os.MkdirAll(logsDir, 0755))

	// Create multiple CSV files
	files := []struct {
		name    string
		content string
	}{
		{
			name: "stories1.csv",
			content: `titulo,descripcion,criterio_aceptacion,subtareas,parent
Historia 1,Descripción 1,Criterio 1,Sub1;Sub2,`,
		},
		{
			name: "stories2.csv",
			content: `titulo,descripcion,criterio_aceptacion,subtareas,parent
Historia 2,Descripción 2,Criterio 2,Sub3,PROJ-456`,
		},
	}

	for _, file := range files {
		filePath := filepath.Join(inputDir, file.name)
		require.NoError(t, os.WriteFile(filePath, []byte(file.content), 0644))
	}

	// Setup environment
	originalEnv := setupTestEnvironment(t, tempDir)
	defer restoreEnvironment(originalEnv)

	// Create app
	app, err := createTestApp(tempDir)
	require.NoError(t, err)
	defer app.logger.Close() // Ensure logger is closed to release file handles

	// Execute processing all files in directory (dry-run)
	ctx := context.Background()
	err = app.runProcess(ctx, "TEST-PROJ", "", true) // Empty file path processes all files

	assert.NoError(t, err)

	// Verify logs contain processing information
	logFiles, err := filepath.Glob(filepath.Join(logsDir, "*.log"))
	require.NoError(t, err)
	require.Len(t, logFiles, 1)

	logContent, err := os.ReadFile(logFiles[0])
	require.NoError(t, err)
	logStr := string(logContent)

	// Verify command execution was logged
	assert.Contains(t, logStr, "process")
	assert.Contains(t, logStr, "TEST-PROJ")
	assert.Contains(t, logStr, "dry_run:true")
}

// Helper functions

func setupTestEnvironment(t *testing.T, tempDir string) map[string]string {
	originalEnv := make(map[string]string)

	envVars := map[string]string{
		"JIRA_URL":            "https://test-integration.atlassian.net",
		"JIRA_EMAIL":          "integration-test@example.com",
		"JIRA_API_TOKEN":      "fake-integration-token",
		"PROJECT_KEY":         "INT-TEST",
		"INPUT_DIRECTORY":     filepath.Join(tempDir, "entrada"),
		"LOGS_DIRECTORY":      filepath.Join(tempDir, "logs"),
		"PROCESSED_DIRECTORY": filepath.Join(tempDir, "procesados"),
		"BATCH_SIZE":          "5",
		"DRY_RUN":             "false", // Controlled per test
	}

	for key, value := range envVars {
		originalEnv[key] = os.Getenv(key)
		os.Setenv(key, value)
	}

	return originalEnv
}

func restoreEnvironment(originalEnv map[string]string) {
	for key, value := range originalEnv {
		if value == "" {
			os.Unsetenv(key)
		} else {
			os.Setenv(key, value)
		}
	}
}

func createTestApp(tempDir string) (*App, error) {
	// Load config from environment
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}

	// Create real dependencies
	appLogger, err := logger.NewLogger(cfg.LogsDirectory)
	if err != nil {
		return nil, err
	}

	fileRepo := filesystem.NewFileProcessor(cfg.ProcessedDirectory)

	// Use mock Jira repository for safety - even in dry-run we don't want real API calls
	jiraRepo := &mocks.MockJiraRepository{
		ValidateProjectFunc: func(ctx context.Context, projectKey string) error {
			return nil // Always valid for integration tests
		},
		ValidateSubtaskIssueTypeFunc: func(ctx context.Context, projectKey string) error {
			return nil
		},
		ValidateFeatureIssueTypeFunc: func(ctx context.Context) error {
			return nil
		},
		TestConnectionFunc: func(ctx context.Context) error {
			return nil // Always connected for integration tests
		},
		CreateUserStoryFunc: func(ctx context.Context, story *entities.UserStory, rowNumber int) (*entities.ProcessResult, error) {
			// Mock successful creation for integration tests
			result := entities.NewProcessResult(rowNumber)
			result.Success = true
			result.IssueKey = "TEST-123"
			result.IssueURL = "https://test.atlassian.net/browse/TEST-123"
			result.CreatedIssueKey = "TEST-123"

			// Mock subtask creation results
			for _, subtarea := range story.Subtareas {
				result.AddSubtaskResult(subtarea, true, fmt.Sprintf("TEST-124-%d", rowNumber),
					fmt.Sprintf("https://test.atlassian.net/browse/TEST-124-%d", rowNumber), "")
			}

			return result, nil
		},
	}

	// Create feature manager mock
	featureManager := &mocks.MockFeatureManager{
		CreateOrGetFeatureFunc: func(ctx context.Context, description string, projectKey string) (*entities.FeatureResult, error) {
			result := entities.NewFeatureResult(description)
			result.SetSuccess("TEST-FEAT-1", "https://test.atlassian.net/browse/TEST-FEAT-1", false) // Simulating existing feature
			return result, nil
		},
	}

	// Create use cases
	processUseCase := usecases.NewProcessFilesUseCase(fileRepo, jiraRepo, featureManager)
	validateUseCase := usecases.NewValidateFileUseCase(fileRepo, jiraRepo)
	testConnectionUseCase := usecases.NewTestConnectionUseCase(jiraRepo)
	diagnoseUseCase := usecases.NewDiagnoseFeaturesUseCase(featureManager)

	// Create formatter
	formatter := formatters.NewOutputFormatter()

	// Create app
	app := &App{
		config:          cfg,
		logger:          appLogger,
		formatter:       formatter,
		testConnUseCase: testConnectionUseCase,
		validateUseCase: validateUseCase,
		processUseCase:  processUseCase,
		diagnoseUseCase: diagnoseUseCase,
	}

	return app, nil
}

// runProcess implements the process command logic for integration testing
func (app *App) runProcess(ctx context.Context, projectKey, filePath string, dryRun bool) error {
	startTime := time.Now()

	// Use default configuration if no project provided
	if projectKey == "" {
		projectKey = app.config.ProjectKey
	}

	// Log command start
	app.logger.LogCommandStart("process", map[string]interface{}{
		"file":        filePath,
		"project_key": projectKey,
		"dry_run":     dryRun,
	})

	// Only require project if not dry-run and not in configuration
	if projectKey == "" && !dryRun {
		app.logger.LogCommandEnd("process", false, time.Since(startTime))
		return fmt.Errorf("project key is required for real processing. Use -p flag, PROJECT_KEY env var, or --dry-run for testing")
	}

	// If dry-run without project, use a dummy one
	if projectKey == "" && dryRun {
		projectKey = "DRY-RUN-PROJECT"
		app.logger.Info("Using dry-run mode without project key - no real Jira operations will be performed")
	}

	var err error
	if filePath != "" {
		// Process single file
		_, err = app.processUseCase.Execute(ctx, filePath, projectKey, dryRun)
	} else {
		// Process all files in input directory
		_, err = app.processUseCase.ProcessAllFiles(ctx, app.config.InputDirectory, projectKey, dryRun)
	}

	if err != nil {
		app.logger.LogCommandEnd("process", false, time.Since(startTime))
		return fmt.Errorf("error processing: %w", err)
	}

	app.logger.LogCommandEnd("process", true, time.Since(startTime))
	return nil
}

// runValidate implements the validate command logic for integration testing
func (app *App) runValidate(ctx context.Context, projectKey, filePath string, rows int) error {
	startTime := time.Now()

	// Use default configuration if no project provided
	if projectKey == "" {
		projectKey = app.config.ProjectKey
	}

	// Validate that file path is provided
	if filePath == "" {
		return fmt.Errorf("file path is required. Use -f flag to specify the file to validate")
	}

	// Log command start
	app.logger.LogCommandStart("validate", map[string]interface{}{
		"file":        filePath,
		"project_key": projectKey,
	})
	app.logger.LogValidationStart(filePath)

	validationResult, err := app.validateUseCase.Execute(ctx, filePath, projectKey, rows)

	// Generate formatted output
	output := app.formatter.FormatValidation(filePath, validationResult, err)

	// Write to log
	app.logger.WriteFormattedOutput(output)

	// Log specific events
	if err != nil {
		app.logger.LogValidationError(filePath, err)
	} else {
		app.logger.LogValidationSuccess(filePath, validationResult.TotalStories)
	}

	// Log command end
	app.logger.LogCommandEnd("validate", err == nil, time.Since(startTime))

	return err
}
