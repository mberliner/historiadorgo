package logger

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestNewLogger(t *testing.T) {
	tempDir := t.TempDir()

	logger, err := NewLogger(tempDir)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	defer logger.Close()

	if logger.Logger == nil {
		t.Error("Expected Logger to be initialized")
	}

	if logger.logFile == nil {
		t.Error("Expected logFile to be initialized")
	}

	// Verificar que el directorio fue creado
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Error("Expected logs directory to be created")
	}

	// Verificar que el archivo de log fue creado
	files, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read log directory: %v", err)
	}

	found := false
	for _, file := range files {
		if strings.HasPrefix(file.Name(), "historiador_") && strings.HasSuffix(file.Name(), ".log") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected log file to be created")
	}
}

func TestNewLogger_InvalidDirectory(t *testing.T) {
	// Intentar crear logger en un directorio que no se puede crear
	invalidDir := "/root/invalid/path/that/cannot/be/created"

	_, err := NewLogger(invalidDir)
	if err == nil {
		t.Error("Expected error when creating logger with invalid directory")
	}

	if !strings.Contains(err.Error(), "error creating logs directory") {
		t.Errorf("Expected error about creating logs directory, got: %v", err)
	}
}

func TestLogger_Close(t *testing.T) {
	tempDir := t.TempDir()

	logger, err := NewLogger(tempDir)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	err = logger.Close()
	if err != nil {
		t.Errorf("Expected no error when closing logger, got: %v", err)
	}

	// Nota: Cerrar un archivo ya cerrado puede generar error en algunos sistemas
	// Este comportamiento es normal y esperado
}

func TestLogger_Close_NilFile(t *testing.T) {
	logger := &Logger{
		Logger:  logrus.New(),
		logFile: nil,
	}

	err := logger.Close()
	if err != nil {
		t.Errorf("Expected no error when closing logger with nil file, got: %v", err)
	}
}

func TestLogger_SetLevel(t *testing.T) {
	tempDir := t.TempDir()

	logger, err := NewLogger(tempDir)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	tests := []struct {
		level    string
		expected logrus.Level
	}{
		{"DEBUG", logrus.DebugLevel},
		{"INFO", logrus.InfoLevel},
		{"WARNING", logrus.WarnLevel},
		{"WARN", logrus.WarnLevel},
		{"ERROR", logrus.ErrorLevel},
		{"INVALID", logrus.InfoLevel}, // Default
		{"", logrus.InfoLevel},        // Default
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			logger.SetLevel(tt.level)

			if logger.Logger.GetLevel() != tt.expected {
				t.Errorf("Expected level %v, got %v", tt.expected, logger.Logger.GetLevel())
			}
		})
	}
}

func TestLogger_LogValidationStart(t *testing.T) {
	tempDir := t.TempDir()

	logger, err := NewLogger(tempDir)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	filePath := "/test/file.csv"
	logger.LogValidationStart(filePath)

	// Verificar que se escribió al archivo
	logContent := readLogFile(t, tempDir)
	if !strings.Contains(logContent, "validation_start") {
		t.Error("Expected log to contain 'validation_start'")
	}
	if !strings.Contains(logContent, filePath) {
		t.Error("Expected log to contain file path")
	}
}

func TestLogger_LogValidationSuccess(t *testing.T) {
	tempDir := t.TempDir()

	logger, err := NewLogger(tempDir)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	filePath := "/test/file.csv"
	totalStories := 5
	logger.LogValidationSuccess(filePath, totalStories)

	logContent := readLogFile(t, tempDir)
	if !strings.Contains(logContent, "validation_success") {
		t.Error("Expected log to contain 'validation_success'")
	}
	if !strings.Contains(logContent, fmt.Sprintf("total_stories=%d", totalStories)) {
		t.Error("Expected log to contain total stories count")
	}
}

func TestLogger_LogValidationError(t *testing.T) {
	tempDir := t.TempDir()

	logger, err := NewLogger(tempDir)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	filePath := "/test/file.csv"
	testError := errors.New("validation failed")
	logger.LogValidationError(filePath, testError)

	logContent := readLogFile(t, tempDir)
	if !strings.Contains(logContent, "validation_error") {
		t.Error("Expected log to contain 'validation_error'")
	}
	if !strings.Contains(logContent, testError.Error()) {
		t.Error("Expected log to contain error message")
	}
}

func TestLogger_LogConnectionTest(t *testing.T) {
	tempDir := t.TempDir()

	logger, err := NewLogger(tempDir)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	tests := []struct {
		name    string
		success bool
		err     error
	}{
		{
			name:    "successful_connection",
			success: true,
			err:     nil,
		},
		{
			name:    "failed_connection",
			success: false,
			err:     errors.New("connection failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger.LogConnectionTest(tt.success, tt.err)

			logContent := readLogFile(t, tempDir)
			if !strings.Contains(logContent, "connection_test") {
				t.Error("Expected log to contain 'connection_test'")
			}

			if tt.success {
				if !strings.Contains(logContent, "result=success") {
					t.Error("Expected log to contain 'result=success'")
				}
			} else {
				if !strings.Contains(logContent, "result=error") {
					t.Error("Expected log to contain 'result=error'")
				}
				if !strings.Contains(logContent, tt.err.Error()) {
					t.Error("Expected log to contain error message")
				}
			}
		})
	}
}

func TestLogger_LogProcessStart(t *testing.T) {
	tempDir := t.TempDir()

	logger, err := NewLogger(tempDir)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	filePath := "/test/file.csv"
	projectKey := "TEST"
	dryRun := true

	logger.LogProcessStart(filePath, projectKey, dryRun)

	logContent := readLogFile(t, tempDir)
	if !strings.Contains(logContent, "process_start") {
		t.Error("Expected log to contain 'process_start'")
	}
	if !strings.Contains(logContent, filePath) {
		t.Error("Expected log to contain file path")
	}
	if !strings.Contains(logContent, projectKey) {
		t.Error("Expected log to contain project key")
	}
	if !strings.Contains(logContent, "dry_run=true") {
		t.Error("Expected log to contain dry_run flag")
	}
}

func TestLogger_LogProcessEnd(t *testing.T) {
	tempDir := t.TempDir()

	logger, err := NewLogger(tempDir)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	filePath := "/test/file.csv"
	successful := 3
	failed := 1
	duration := 2 * time.Second

	logger.LogProcessEnd(filePath, successful, failed, duration)

	logContent := readLogFile(t, tempDir)
	if !strings.Contains(logContent, "process_end") {
		t.Error("Expected log to contain 'process_end'")
	}
	if !strings.Contains(logContent, fmt.Sprintf("successful=%d", successful)) {
		t.Error("Expected log to contain successful count")
	}
	if !strings.Contains(logContent, fmt.Sprintf("failed=%d", failed)) {
		t.Error("Expected log to contain failed count")
	}
}

func TestLogger_LogIssueCreated(t *testing.T) {
	tempDir := t.TempDir()

	logger, err := NewLogger(tempDir)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	issueKey := "TEST-123"
	issueType := "Story"
	rowNumber := 5

	logger.LogIssueCreated(issueKey, issueType, rowNumber)

	logContent := readLogFile(t, tempDir)
	if !strings.Contains(logContent, "issue_created") {
		t.Error("Expected log to contain 'issue_created'")
	}
	if !strings.Contains(logContent, issueKey) {
		t.Error("Expected log to contain issue key")
	}
	if !strings.Contains(logContent, issueType) {
		t.Error("Expected log to contain issue type")
	}
}

func TestLogger_LogIssueError(t *testing.T) {
	tempDir := t.TempDir()

	logger, err := NewLogger(tempDir)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	issueType := "Story"
	rowNumber := 5
	testError := errors.New("creation failed")

	logger.LogIssueError(issueType, rowNumber, testError)

	logContent := readLogFile(t, tempDir)
	if !strings.Contains(logContent, "issue_error") {
		t.Error("Expected log to contain 'issue_error'")
	}
	if !strings.Contains(logContent, testError.Error()) {
		t.Error("Expected log to contain error message")
	}
}

func TestLogger_LogFeatureCreated(t *testing.T) {
	tempDir := t.TempDir()

	logger, err := NewLogger(tempDir)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	tests := []struct {
		name           string
		featureKey     string
		description    string
		wasCreated     bool
		expectedAction string
	}{
		{
			name:           "feature_created",
			featureKey:     "TEST-456",
			description:    "New Feature",
			wasCreated:     true,
			expectedAction: "feature_created",
		},
		{
			name:           "feature_found",
			featureKey:     "TEST-789",
			description:    "Existing Feature",
			wasCreated:     false,
			expectedAction: "feature_found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger.LogFeatureCreated(tt.featureKey, tt.description, tt.wasCreated)

			logContent := readLogFile(t, tempDir)
			if !strings.Contains(logContent, tt.expectedAction) {
				t.Errorf("Expected log to contain '%s'", tt.expectedAction)
			}
			if !strings.Contains(logContent, tt.featureKey) {
				t.Error("Expected log to contain feature key")
			}
		})
	}
}

func TestLogger_WriteFormattedOutput(t *testing.T) {
	tempDir := t.TempDir()

	logger, err := NewLogger(tempDir)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	testOutput := "Test formatted output content"
	logger.WriteFormattedOutput(testOutput)

	logContent := readLogFile(t, tempDir)
	if !strings.Contains(logContent, "=== SALIDA COMANDO") {
		t.Error("Expected log to contain formatted output header")
	}
	if !strings.Contains(logContent, testOutput) {
		t.Error("Expected log to contain test output content")
	}
	if !strings.Contains(logContent, "=== FIN SALIDA ===") {
		t.Error("Expected log to contain formatted output footer")
	}
}

func TestLogger_WriteFormattedOutput_NilFile(t *testing.T) {
	logger := &Logger{
		Logger:  logrus.New(),
		logFile: nil,
	}

	// No debería hacer panic
	logger.WriteFormattedOutput("test output")
}

func TestLogger_LogCommandStart(t *testing.T) {
	tempDir := t.TempDir()

	logger, err := NewLogger(tempDir)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	command := "process"
	args := map[string]interface{}{
		"file":    "test.csv",
		"project": "TEST",
		"dry_run": true,
	}

	logger.LogCommandStart(command, args)

	logContent := readLogFile(t, tempDir)
	if !strings.Contains(logContent, "command_start") {
		t.Error("Expected log to contain 'command_start'")
	}
	if !strings.Contains(logContent, command) {
		t.Error("Expected log to contain command name")
	}
}

func TestLogger_LogCommandEnd(t *testing.T) {
	tempDir := t.TempDir()

	logger, err := NewLogger(tempDir)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	tests := []struct {
		name           string
		command        string
		success        bool
		duration       time.Duration
		expectedStatus string
	}{
		{
			name:           "successful_command",
			command:        "validate",
			success:        true,
			duration:       time.Second,
			expectedStatus: "success",
		},
		{
			name:           "failed_command",
			command:        "process",
			success:        false,
			duration:       2 * time.Second,
			expectedStatus: "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger.LogCommandEnd(tt.command, tt.success, tt.duration)

			logContent := readLogFile(t, tempDir)
			if !strings.Contains(logContent, "command_end") {
				t.Error("Expected log to contain 'command_end'")
			}
			if !strings.Contains(logContent, tt.command) {
				t.Error("Expected log to contain command name")
			}
			if !strings.Contains(logContent, fmt.Sprintf("status=%s", tt.expectedStatus)) {
				t.Errorf("Expected log to contain 'status=%s'", tt.expectedStatus)
			}
		})
	}
}

func TestLogger_LogSubtaskCreated(t *testing.T) {
	tempDir := t.TempDir()

	logger, err := NewLogger(tempDir)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Set debug level to capture debug logs
	logger.SetLevel("DEBUG")

	parentKey := "PROJ-123"
	subtaskKey := "PROJ-124"
	description := "Subtask description"

	logger.LogSubtaskCreated(parentKey, subtaskKey, description)

	logContent := readLogFile(t, tempDir)
	if !strings.Contains(logContent, "subtask_created") {
		t.Error("Expected log to contain 'subtask_created'")
	}
	if !strings.Contains(logContent, parentKey) {
		t.Error("Expected log to contain parent key")
	}
	if !strings.Contains(logContent, subtaskKey) {
		t.Error("Expected log to contain subtask key")
	}
	if !strings.Contains(logContent, description) {
		t.Error("Expected log to contain description")
	}
}

func TestLogger_LogSubtaskError(t *testing.T) {
	tempDir := t.TempDir()

	logger, err := NewLogger(tempDir)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	parentKey := "PROJ-123"
	description := "Failed subtask description"
	testError := errors.New("creation failed")

	logger.LogSubtaskError(parentKey, description, testError)

	logContent := readLogFile(t, tempDir)
	if !strings.Contains(logContent, "subtask_error") {
		t.Error("Expected log to contain 'subtask_error'")
	}
	if !strings.Contains(logContent, parentKey) {
		t.Error("Expected log to contain parent key")
	}
	if !strings.Contains(logContent, description) {
		t.Error("Expected log to contain description")
	}
	if !strings.Contains(logContent, testError.Error()) {
		t.Error("Expected log to contain error message")
	}
}

// Helper function para leer el contenido del archivo de log
func readLogFile(t *testing.T, logDir string) string {
	files, err := os.ReadDir(logDir)
	if err != nil {
		t.Fatalf("Failed to read log directory: %v", err)
	}

	for _, file := range files {
		if strings.HasPrefix(file.Name(), "historiador_") && strings.HasSuffix(file.Name(), ".log") {
			content, err := os.ReadFile(filepath.Join(logDir, file.Name()))
			if err != nil {
				t.Fatalf("Failed to read log file: %v", err)
			}
			return string(content)
		}
	}

	t.Fatal("No log file found")
	return ""
}
