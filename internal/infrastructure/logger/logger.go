package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

type Logger struct {
	*logrus.Logger
	logFile *os.File
}

func NewLogger(logsDir string) (*Logger, error) {
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return nil, fmt.Errorf("error creating logs directory: %w", err)
	}

	logFileName := fmt.Sprintf("historiador_%s.log", time.Now().Format("20060102_150405"))
	logFilePath := filepath.Join(logsDir, logFileName)

	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("error creating log file: %w", err)
	}

	logger := logrus.New()

	// Solo estructurados van a logrus
	logger.SetOutput(logFile)

	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		DisableColors:   true, // Sin colores en archivo
	})

	logger.SetLevel(logrus.InfoLevel)

	return &Logger{
		Logger:  logger,
		logFile: logFile,
	}, nil
}

func (l *Logger) Close() error {
	if l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}

func (l *Logger) SetLevel(level string) {
	switch level {
	case "DEBUG":
		l.Logger.SetLevel(logrus.DebugLevel)
	case "INFO":
		l.Logger.SetLevel(logrus.InfoLevel)
	case "WARNING", "WARN":
		l.Logger.SetLevel(logrus.WarnLevel)
	case "ERROR":
		l.Logger.SetLevel(logrus.ErrorLevel)
	default:
		l.Logger.SetLevel(logrus.InfoLevel)
	}
}

func (l *Logger) LogValidationStart(filePath string) {
	l.WithFields(logrus.Fields{
		"action": "validation_start",
		"file":   filePath,
	}).Info("Iniciando validacion de archivo")
}

func (l *Logger) LogValidationSuccess(filePath string, totalStories int) {
	l.WithFields(logrus.Fields{
		"action":        "validation_success",
		"file":          filePath,
		"total_stories": totalStories,
	}).Info("Validacion exitosa")
}

func (l *Logger) LogValidationError(filePath string, err error) {
	l.WithFields(logrus.Fields{
		"action": "validation_error",
		"file":   filePath,
		"error":  err.Error(),
	}).Error("Error en validacion")
}

func (l *Logger) LogConnectionTest(success bool, err error) {
	if success {
		l.WithFields(logrus.Fields{
			"action": "connection_test",
			"result": "success",
		}).Info("Conexion con Jira exitosa")
	} else {
		l.WithFields(logrus.Fields{
			"action": "connection_test",
			"result": "error",
			"error":  err.Error(),
		}).Error("Error en conexion con Jira")
	}
}

func (l *Logger) LogProcessStart(filePath, projectKey string, dryRun bool) {
	l.WithFields(logrus.Fields{
		"action":      "process_start",
		"file":        filePath,
		"project_key": projectKey,
		"dry_run":     dryRun,
	}).Info("Iniciando procesamiento de archivo")
}

func (l *Logger) LogProcessEnd(filePath string, successful, failed int, duration time.Duration) {
	l.WithFields(logrus.Fields{
		"action":     "process_end",
		"file":       filePath,
		"successful": successful,
		"failed":     failed,
		"duration":   duration.String(),
	}).Info("Procesamiento completado")
}

func (l *Logger) LogIssueCreated(issueKey, issueType string, rowNumber int) {
	l.WithFields(logrus.Fields{
		"action":     "issue_created",
		"issue_key":  issueKey,
		"issue_type": issueType,
		"row":        rowNumber,
	}).Info("Issue creado exitosamente")
}

func (l *Logger) LogIssueError(issueType string, rowNumber int, err error) {
	l.WithFields(logrus.Fields{
		"action":     "issue_error",
		"issue_type": issueType,
		"row":        rowNumber,
		"error":      err.Error(),
	}).Error("Error creando issue")
}

func (l *Logger) LogSubtaskCreated(parentKey, subtaskKey, description string) {
	l.WithFields(logrus.Fields{
		"action":      "subtask_created",
		"parent_key":  parentKey,
		"subtask_key": subtaskKey,
		"description": description,
	}).Debug("Subtarea creada")
}

func (l *Logger) LogSubtaskError(parentKey, description string, err error) {
	l.WithFields(logrus.Fields{
		"action":      "subtask_error",
		"parent_key":  parentKey,
		"description": description,
		"error":       err.Error(),
	}).Warn("Error creando subtarea")
}

func (l *Logger) LogFeatureCreated(featureKey, description string, wasCreated bool) {
	action := "feature_found"
	if wasCreated {
		action = "feature_created"
	}

	l.WithFields(logrus.Fields{
		"action":      action,
		"feature_key": featureKey,
		"description": description,
	}).Info(fmt.Sprintf("Feature %s", action))
}

// Métodos para escribir salida formateada al log
func (l *Logger) WriteFormattedOutput(output string) {
	if l.logFile != nil {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		formattedOutput := fmt.Sprintf("\n=== SALIDA COMANDO [%s] ===\n%s=== FIN SALIDA ===\n\n", timestamp, output)
		l.logFile.WriteString(formattedOutput)
	}
}

func (l *Logger) LogCommandStart(command string, args map[string]interface{}) {
	l.WithFields(logrus.Fields{
		"action":  "command_start",
		"command": command,
		"args":    args,
	}).Info(fmt.Sprintf("Ejecutando comando: %s", command))
}

func (l *Logger) LogCommandEnd(command string, success bool, duration time.Duration) {
	status := "success"
	if !success {
		status = "error"
	}

	l.WithFields(logrus.Fields{
		"action":   "command_end",
		"command":  command,
		"status":   status,
		"duration": duration.String(),
	}).Info(fmt.Sprintf("Comando completado: %s [%s]", command, status))
}
