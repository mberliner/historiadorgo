package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	JiraURL                  string
	JiraEmail                string
	JiraAPIToken             string
	ProjectKey               string
	DefaultIssueType         string
	SubtaskIssueType         string
	FeatureIssueType         string
	BatchSize                int
	DryRun                   bool
	AcceptanceCriteriaField  string
	InputDirectory           string
	LogsDirectory            string
	ProcessedDirectory       string
	RollbackOnSubtaskFailure bool
	FeatureRequiredFields    string
}

func LoadConfig() (*Config, error) {
	// Try to load .env file if it exists
	if err := godotenv.Load(); err != nil {
		// .env file doesn't exist or can't be loaded
		// Check if we have the required environment variables already set
		if !hasRequiredEnvVars() {
			fmt.Println("Archivo .env no encontrado")
			fmt.Println("Iniciando configuracion interactiva...")
			fmt.Println()
			
			if err := CreateInteractiveEnvFile(); err != nil {
				return nil, fmt.Errorf("error creating .env file: %w", err)
			}
			
			fmt.Println("Archivo .env creado exitosamente")
			fmt.Println()
			
			// Load the newly created .env file
			if err := godotenv.Load(); err != nil {
				return nil, fmt.Errorf("error loading created .env file: %w", err)
			}
		}
	}

	config := &Config{
		JiraURL:                  getEnv("JIRA_URL", ""),
		JiraEmail:                getEnv("JIRA_EMAIL", ""),
		JiraAPIToken:             getEnv("JIRA_API_TOKEN", ""),
		ProjectKey:               getEnv("PROJECT_KEY", ""),
		DefaultIssueType:         getEnv("DEFAULT_ISSUE_TYPE", "Story"),
		SubtaskIssueType:         getEnv("SUBTASK_ISSUE_TYPE", "Sub-task"),
		FeatureIssueType:         getEnv("FEATURE_ISSUE_TYPE", "Feature"),
		BatchSize:                getEnvAsInt("BATCH_SIZE", 10),
		DryRun:                   getEnvAsBool("DRY_RUN", false),
		AcceptanceCriteriaField:  getEnv("ACCEPTANCE_CRITERIA_FIELD", ""),
		InputDirectory:           getEnv("INPUT_DIRECTORY", "entrada"),
		LogsDirectory:            getEnv("LOGS_DIRECTORY", "logs"),
		ProcessedDirectory:       getEnv("PROCESSED_DIRECTORY", "procesados"),
		RollbackOnSubtaskFailure: getEnvAsBool("ROLLBACK_ON_SUBTASK_FAILURE", false),
		FeatureRequiredFields:    getEnv("FEATURE_REQUIRED_FIELDS", ""),
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

func (c *Config) Validate() error {
	var missing []string

	if c.JiraURL == "" {
		missing = append(missing, "JIRA_URL")
	}
	if c.JiraEmail == "" {
		missing = append(missing, "JIRA_EMAIL")
	}
	if c.JiraAPIToken == "" {
		missing = append(missing, "JIRA_API_TOKEN")
	}
	// PROJECT_KEY ya no es obligatorio - se puede pasar por flag o usar para dry-run

	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

// CreateInteractiveEnvFile creates a .env file by prompting user for configuration
func CreateInteractiveEnvFile() error {
	reader := bufio.NewReader(os.Stdin)
	
	fmt.Println("CONFIGURACION DE JIRA")
	fmt.Println("=====================")
	fmt.Println()
	
	// Collect Jira configuration
	jiraURL := promptForInput(reader, "URL de Jira (ej: https://company.atlassian.net)", "")
	if jiraURL == "" {
		return fmt.Errorf("JIRA_URL es requerido")
	}
	
	jiraEmail := promptForInput(reader, "Email de Jira", "")
	if jiraEmail == "" {
		return fmt.Errorf("JIRA_EMAIL es requerido")
	}
	
	fmt.Println("API Token de Jira:")
	fmt.Println("  Obten tu token en: https://id.atlassian.com/manage-profile/security/api-tokens")
	jiraToken := promptForInput(reader, "  Token", "")
	if jiraToken == "" {
		return fmt.Errorf("JIRA_API_TOKEN es requerido")
	}
	
	fmt.Println()
	fmt.Println("CONFIGURACION DEL PROYECTO")
	fmt.Println("==========================")
	fmt.Println()
	
	// Project configuration
	projectKey := promptForInput(reader, "Clave del proyecto por defecto (ej: MYPROJ)", "")
	storyType := promptForInput(reader, "Tipo de issue para historias", "Story")
	subtaskType := promptForInput(reader, "Tipo de issue para subtareas", "Sub-task")
	featureType := promptForInput(reader, "Tipo de issue para Features", "Epic")
	
	fmt.Println()
	fmt.Println("CONFIGURACION DE DIRECTORIOS")
	fmt.Println("============================")
	fmt.Println()
	
	// Directory configuration
	inputDir := promptForInput(reader, "Directorio de entrada", "entrada")
	logsDir := promptForInput(reader, "Directorio de logs", "logs")
	processedDir := promptForInput(reader, "Directorio de procesados", "procesados")
	
	fmt.Println()
	fmt.Println("CONFIGURACION AVANZADA")
	fmt.Println("======================")
	fmt.Println()
	
	rollback := promptForYesNo(reader, "Hacer rollback si fallan subtareas?", false)
	
	// Create .env content
	envContent := fmt.Sprintf(`# Configuracion de Jira
JIRA_URL=%s
JIRA_EMAIL=%s
JIRA_API_TOKEN=%s

# Configuracion del proyecto
PROJECT_KEY=%s
DEFAULT_ISSUE_TYPE=%s
SUBTASK_ISSUE_TYPE=%s
FEATURE_ISSUE_TYPE=%s

# Configuracion de directorios
INPUT_DIRECTORY=%s
LOGS_DIRECTORY=%s
PROCESSED_DIRECTORY=%s

# Configuracion avanzada
ROLLBACK_ON_SUBTASK_FAILURE=%t
FEATURE_REQUIRED_FIELDS=""
ACCEPTANCE_CRITERIA_FIELD=""
BATCH_SIZE=10
DRY_RUN=false
`, jiraURL, jiraEmail, jiraToken, projectKey, storyType, subtaskType, featureType,
		inputDir, logsDir, processedDir, rollback)
	
	// Write .env file
	if err := os.WriteFile(".env", []byte(envContent), 0644); err != nil {
		return fmt.Errorf("error writing .env file: %w", err)
	}
	
	// Create directories if they don't exist
	dirs := []string{inputDir, logsDir, processedDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("Warning: Could not create directory %s: %v\n", dir, err)
		}
	}
	
	return nil
}

// promptForInput prompts the user for input with a default value
func promptForInput(reader *bufio.Reader, prompt, defaultValue string) string {
	if defaultValue != "" {
		fmt.Printf("%s [%s]: ", prompt, defaultValue)
	} else {
		fmt.Printf("%s: ", prompt)
	}
	
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	
	if input == "" && defaultValue != "" {
		return defaultValue
	}
	
	return input
}

// promptForYesNo prompts the user for a yes/no input
func promptForYesNo(reader *bufio.Reader, prompt string, defaultValue bool) bool {
	defaultStr := "n"
	if defaultValue {
		defaultStr = "y"
	}
	
	fmt.Printf("%s [%s]: ", prompt, defaultStr)
	
	input, _ := reader.ReadString('\n')
	input = strings.ToLower(strings.TrimSpace(input))
	
	if input == "" {
		return defaultValue
	}
	
	return input == "y" || input == "yes" || input == "si"
}

// hasRequiredEnvVars checks if all required environment variables are already set
func hasRequiredEnvVars() bool {
	requiredVars := []string{"JIRA_URL", "JIRA_EMAIL", "JIRA_API_TOKEN"}
	
	for _, envVar := range requiredVars {
		if os.Getenv(envVar) == "" {
			return false
		}
	}
	
	return true
}
