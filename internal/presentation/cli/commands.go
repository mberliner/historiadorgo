package cli

import (
	"context"
	"fmt"
	"time"

	"historiadorgo/internal/application/usecases"
	"historiadorgo/internal/domain/entities"
	"historiadorgo/internal/infrastructure/config"
	"historiadorgo/internal/infrastructure/filesystem"
	"historiadorgo/internal/infrastructure/jira"
	"historiadorgo/internal/infrastructure/logger"
	"historiadorgo/internal/presentation/formatters"

	"github.com/spf13/cobra"
)

type App struct {
	config          *config.Config
	logger          *logger.Logger
	formatter       *formatters.OutputFormatter
	testConnUseCase *usecases.TestConnectionUseCase
	validateUseCase *usecases.ValidateFileUseCase
	processUseCase  *usecases.ProcessFilesUseCase
	diagnoseUseCase *usecases.DiagnoseFeaturesUseCase
}

func NewApp() (*App, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("error loading config: %w", err)
	}

	appLogger, err := logger.NewLogger(cfg.LogsDirectory)
	if err != nil {
		return nil, fmt.Errorf("error creating logger: %w", err)
	}

	jiraClient := jira.NewJiraClient(cfg)
	fileProcessor := filesystem.NewFileProcessor(cfg.ProcessedDirectory)
	featureManager := jira.NewFeatureManager(jiraClient, cfg)
	formatter := formatters.NewOutputFormatter()

	return &App{
		config:          cfg,
		logger:          appLogger,
		formatter:       formatter,
		testConnUseCase: usecases.NewTestConnectionUseCase(jiraClient),
		validateUseCase: usecases.NewValidateFileUseCase(fileProcessor, jiraClient),
		processUseCase:  usecases.NewProcessFilesUseCase(fileProcessor, jiraClient, featureManager),
		diagnoseUseCase: usecases.NewDiagnoseFeaturesUseCase(featureManager),
	}, nil
}

func NewRootCmd() *cobra.Command {
	var (
		projectKey string
		filePath   string
		dryRun     bool
		batchSize  int
		logLevel   string
	)

	rootCmd := &cobra.Command{
		Use:   "historiador",
		Short: "Jira Batch Importer - Crea historias de usuario desde archivos Excel/CSV",
		Long: `Aplicación CLI para crear historias de usuario en Jira desde archivos Excel/CSV 
con gestión automática de subtareas y Features.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := NewApp()
			if err != nil {
				return err
			}

			app.logger.SetLevel(logLevel)

			return app.runProcess(cmd.Context(), projectKey, filePath, dryRun)
		},
	}

	rootCmd.PersistentFlags().StringVarP(&projectKey, "project", "p", "", "Key del proyecto en Jira (ej: MYPROJ)")
	rootCmd.PersistentFlags().StringVarP(&filePath, "file", "f", "", "Archivo Excel o CSV específico")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Modo de prueba sin crear issues")
	rootCmd.PersistentFlags().IntVarP(&batchSize, "batch-size", "b", 10, "Tamaño del lote de procesamiento")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "INFO", "Nivel de log (DEBUG, INFO, WARN, ERROR)")

	return rootCmd
}

func NewProcessCmd() *cobra.Command {
	var (
		projectKey string
		filePath   string
		dryRun     bool
		batchSize  int
	)

	cmd := &cobra.Command{
		Use:   "process",
		Short: "Procesa archivos Excel/CSV para crear historias en Jira",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := NewApp()
			if err != nil {
				return err
			}

			return app.runProcess(cmd.Context(), projectKey, filePath, dryRun)
		},
	}

	cmd.Flags().StringVarP(&projectKey, "project", "p", "", "Key del proyecto en Jira (ej: MYPROJ)")
	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Archivo Excel o CSV específico")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Modo de prueba sin crear issues")
	cmd.Flags().IntVarP(&batchSize, "batch-size", "b", 10, "Tamaño del lote de procesamiento")

	return cmd
}

func NewValidateCmd() *cobra.Command {
	var (
		projectKey string
		filePath   string
		rows       int
	)

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Valida un archivo sin crear issues en Jira",
		RunE: func(cmd *cobra.Command, args []string) error {
			logLevel, _ := cmd.Flags().GetString("log-level")

			app, err := NewApp()
			if err != nil {
				return err
			}

			app.logger.SetLevel(logLevel)

			return app.runValidate(cmd.Context(), projectKey, filePath, rows)
		},
	}

	cmd.Flags().StringVarP(&projectKey, "project", "p", "", "Key del proyecto en Jira")
	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Archivo Excel o CSV a validar")
	cmd.Flags().IntVarP(&rows, "rows", "r", 5, "Número de filas a mostrar en preview")

	return cmd
}

func NewTestConnectionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test-connection",
		Short: "Prueba la conexión con Jira",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := NewApp()
			if err != nil {
				return err
			}

			return app.runTestConnection(cmd.Context())
		},
	}

	return cmd
}

func NewDiagnoseCmd() *cobra.Command {
	var projectKey string

	cmd := &cobra.Command{
		Use:   "diagnose",
		Short: "Diagnostica los campos requeridos para Features en el proyecto",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := NewApp()
			if err != nil {
				return err
			}

			return app.runDiagnose(cmd.Context(), projectKey)
		},
	}

	cmd.Flags().StringVarP(&projectKey, "project", "p", "", "Key del proyecto en Jira")

	return cmd
}

func (app *App) runProcess(ctx context.Context, projectKey, filePath string, dryRun bool) error {
	startTime := time.Now()

	// Usar configuración por defecto si no se proporciona proyecto
	if projectKey == "" {
		projectKey = app.config.ProjectKey
	}

	// Log inicio de comando
	app.logger.LogCommandStart("process", map[string]interface{}{
		"file":        filePath,
		"project_key": projectKey,
		"dry_run":     dryRun,
	})

	// Solo requerir proyecto si no es dry-run y no está en configuración
	if projectKey == "" && !dryRun {
		app.logger.LogCommandEnd("process", false, time.Since(startTime))
		return fmt.Errorf("project key is required for real processing. Use -p flag, PROJECT_KEY env var, or --dry-run for testing")
	}

	// Si es dry-run sin proyecto, usar uno ficticio
	if projectKey == "" && dryRun {
		projectKey = "DRY-RUN-PROJECT"
		app.logger.Info("Using dry-run mode without project key - no real Jira operations will be performed")
	}

	var results []*entities.BatchResult
	var err error

	if filePath != "" {
		var result *entities.BatchResult
		result, err = app.processUseCase.Execute(ctx, filePath, projectKey, dryRun)
		if err != nil {
			app.logger.LogCommandEnd("process", false, time.Since(startTime))
			return fmt.Errorf("error processing file: %w", err)
		}
		results = []*entities.BatchResult{result}
	} else {
		results, err = app.processUseCase.ProcessAllFiles(ctx, app.config.InputDirectory, projectKey, dryRun)
		if err != nil {
			app.logger.LogCommandEnd("process", false, time.Since(startTime))
			return fmt.Errorf("error processing files: %w", err)
		}
	}

	// Generar salida formateada
	var output string
	if len(results) == 1 {
		output = app.formatter.FormatBatchResult(results[0])
	} else {
		output = app.formatter.FormatMultipleBatchResults(results)
	}

	// Mostrar en consola
	fmt.Print(output)

	// Escribir al log
	app.logger.WriteFormattedOutput(output)

	// Log fin de comando
	app.logger.LogCommandEnd("process", true, time.Since(startTime))

	return nil
}

func (app *App) runValidate(ctx context.Context, projectKey, filePath string, rows int) error {
	startTime := time.Now()

	// Usar configuración por defecto si no se proporciona proyecto
	if projectKey == "" {
		projectKey = app.config.ProjectKey
	}

	// Validar que se proporcione archivo
	if filePath == "" {
		return fmt.Errorf("file path is required. Use -f flag to specify the file to validate")
	}

	// Log inicio de comando
	app.logger.LogCommandStart("validate", map[string]interface{}{
		"file":        filePath,
		"project_key": projectKey,
	})
	app.logger.LogValidationStart(filePath)

	validationResult, err := app.validateUseCase.Execute(ctx, filePath, projectKey, rows)

	// Generar salida formateada
	output := app.formatter.FormatValidation(filePath, validationResult, err)

	// Mostrar en consola
	fmt.Print(output)

	// Escribir al log
	app.logger.WriteFormattedOutput(output)

	// Log eventos específicos
	if err != nil {
		app.logger.LogValidationError(filePath, err)
	} else {
		app.logger.LogValidationSuccess(filePath, validationResult.TotalStories)
	}

	// Log fin de comando
	app.logger.LogCommandEnd("validate", err == nil, time.Since(startTime))

	return err
}

func (app *App) runTestConnection(ctx context.Context) error {
	startTime := time.Now()

	// Log inicio de comando
	app.logger.LogCommandStart("test-connection", map[string]interface{}{})

	err := app.testConnUseCase.Execute(ctx)

	// Generar salida formateada
	output := app.formatter.FormatConnectionTest(err)

	// Mostrar en consola
	fmt.Print(output)

	// Escribir al log
	app.logger.WriteFormattedOutput(output)

	// Log eventos específicos
	app.logger.LogConnectionTest(err == nil, err)

	// Log fin de comando
	app.logger.LogCommandEnd("test-connection", err == nil, time.Since(startTime))

	return err
}

func (app *App) runDiagnose(ctx context.Context, projectKey string) error {
	startTime := time.Now()

	// Usar configuración por defecto si no se proporciona proyecto
	if projectKey == "" {
		projectKey = app.config.ProjectKey
	}

	// Log inicio de comando
	app.logger.LogCommandStart("diagnose", map[string]interface{}{
		"project_key": projectKey,
	})

	// Si no hay proyecto disponible, mostrar mensaje informativo
	if projectKey == "" {
		app.logger.LogCommandEnd("diagnose", false, time.Since(startTime))
		output := app.formatter.FormatDiagnosisNoProject()
		fmt.Print(output)
		app.logger.WriteFormattedOutput(output)
		return nil
	}

	requiredFields, err := app.diagnoseUseCase.Execute(ctx, projectKey)
	if err != nil {
		app.logger.LogCommandEnd("diagnose", false, time.Since(startTime))
		return fmt.Errorf("error diagnosing features: %w", err)
	}

	// Generar salida formateada
	output := app.formatter.FormatDiagnosis(requiredFields)

	// Mostrar en consola
	fmt.Print(output)

	// Escribir al log
	app.logger.WriteFormattedOutput(output)

	// Log fin de comando
	app.logger.LogCommandEnd("diagnose", true, time.Since(startTime))

	return nil
}

func SetupCommands() *cobra.Command {
	rootCmd := NewRootCmd()

	rootCmd.AddCommand(NewProcessCmd())
	rootCmd.AddCommand(NewValidateCmd())
	rootCmd.AddCommand(NewTestConnectionCmd())
	rootCmd.AddCommand(NewDiagnoseCmd())

	return rootCmd
}
