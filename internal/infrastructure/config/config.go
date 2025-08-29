package config

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

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
	
	// Get issue types dynamically from Jira
	var storyType, subtaskType, featureType string
	if projectKey != "" {
		fmt.Println()
		fmt.Println("CONSULTANDO TIPOS DE ISSUE EN JIRA...")
		fmt.Println("=====================================")
		fmt.Println()
		
		issueTypes, err := getAvailableIssueTypes(jiraURL, jiraEmail, jiraToken, projectKey)
		if err != nil {
			fmt.Printf("⚠ No se pudieron obtener los tipos de issue desde Jira: %v\n", err)
			fmt.Println("Usando valores por defecto...")
			storyType = promptForInput(reader, "Tipo de issue para historias", "Story")
			subtaskType = promptForInput(reader, "Tipo de issue para subtareas", "Sub-task")
			featureType = promptForInput(reader, "Tipo de issue para Features", "Epic")
		} else {
			storyType = selectIssueType(reader, "historias", issueTypes, false)
			subtaskType = selectIssueType(reader, "subtareas", issueTypes, true)
			featureType = selectIssueType(reader, "Features/Epics", issueTypes, false)
		}
	} else {
		// Fallback to manual input if no project key
		storyType = promptForInput(reader, "Tipo de issue para historias", "Story")
		subtaskType = promptForInput(reader, "Tipo de issue para subtareas", "Sub-task")
		featureType = promptForInput(reader, "Tipo de issue para Features", "Epic")
	}

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

	// Auto-detect Jira configuration if project is provided
	var acceptanceCriteriaField string
	var featureRequiredFields string

	if projectKey != "" {
		fmt.Println()
		fmt.Println("DETECTANDO CONFIGURACION DE JIRA...")
		fmt.Println("===================================")
		fmt.Println()

		if autoConfig, err := DetectJiraConfiguration(jiraURL, jiraEmail, jiraToken, projectKey, storyType, featureType); err == nil {
			acceptanceCriteriaField = autoConfig.AcceptanceCriteriaField
			featureRequiredFields = autoConfig.FeatureRequiredFields

			fmt.Printf("✓ Campo de criterios de aceptación detectado: %s\n", acceptanceCriteriaField)
			fmt.Printf("✓ Campos obligatorios para Features detectados\n")
		} else {
			fmt.Printf("⚠ No se pudo detectar configuración automáticamente: %v\n", err)
		}
	}

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

# Configuracion de campos Jira (detectados automaticamente)
ACCEPTANCE_CRITERIA_FIELD=%s
FEATURE_REQUIRED_FIELDS=%s

# Configuracion de directorios
INPUT_DIRECTORY=%s
LOGS_DIRECTORY=%s
PROCESSED_DIRECTORY=%s

# Configuracion avanzada
ROLLBACK_ON_SUBTASK_FAILURE=%t
BATCH_SIZE=10
DRY_RUN=false
`, jiraURL, jiraEmail, jiraToken, projectKey, storyType, subtaskType, featureType,
		acceptanceCriteriaField, featureRequiredFields,
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

// AutoDetectedConfig contains the automatically detected Jira configuration
type AutoDetectedConfig struct {
	AcceptanceCriteriaField string
	FeatureRequiredFields   string
}

// DetectJiraConfiguration automatically detects Jira field configuration
func DetectJiraConfiguration(jiraURL, jiraEmail, jiraToken, projectKey, storyType, featureType string) (*AutoDetectedConfig, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := &http.Client{Timeout: 30 * time.Second}
	baseURL := strings.TrimSuffix(jiraURL, "/")

	config := &AutoDetectedConfig{}

	// Detect acceptance criteria field
	acceptanceCriteriaField, err := detectAcceptanceCriteriaField(ctx, client, baseURL, jiraEmail, jiraToken, projectKey, storyType)
	if err == nil {
		config.AcceptanceCriteriaField = acceptanceCriteriaField
	}

	// Detect feature required fields
	featureRequiredFields, err := detectFeatureRequiredFields(ctx, client, baseURL, jiraEmail, jiraToken, projectKey, featureType)
	if err == nil {
		config.FeatureRequiredFields = featureRequiredFields
	}

	return config, nil
}

// detectAcceptanceCriteriaField detects the acceptance criteria custom field
func detectAcceptanceCriteriaField(ctx context.Context, client *http.Client, baseURL, email, token, projectKey, storyType string) (string, error) {
	// Get create meta for Story issue type to find acceptance criteria field
	endpoint := fmt.Sprintf("%s/rest/api/3/issue/createmeta?projectKeys=%s&issuetypeNames=%s&expand=projects.issuetypes.fields", baseURL, projectKey, storyType)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return "", err
	}

	req.SetBasicAuth(email, token)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get create meta: status %d", resp.StatusCode)
	}

	var createMeta map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&createMeta); err != nil {
		return "", err
	}

	projects, ok := createMeta["projects"].([]interface{})
	if !ok || len(projects) == 0 {
		return "", fmt.Errorf("no projects found")
	}

	project := projects[0].(map[string]interface{})
	issueTypes, ok := project["issuetypes"].([]interface{})
	if !ok {
		return "", fmt.Errorf("no issue types found")
	}

	for _, issueTypeData := range issueTypes {
		issueType := issueTypeData.(map[string]interface{})
		if name, ok := issueType["name"].(string); ok && name == storyType {
			fields, ok := issueType["fields"].(map[string]interface{})
			if !ok {
				return "", fmt.Errorf("no fields found")
			}

			// Look for common acceptance criteria field names/patterns
			acceptancePatterns := []string{
				"acceptance", "criterio", "criteria", "aceptacion", "aceptacion",
			}

			for fieldKey, fieldData := range fields {
				if fieldMap, ok := fieldData.(map[string]interface{}); ok {
					if fieldName, ok := fieldMap["name"].(string); ok {
						fieldNameLower := strings.ToLower(fieldName)
						for _, pattern := range acceptancePatterns {
							if strings.Contains(fieldNameLower, pattern) {
								return fieldKey, nil
							}
						}
					}
				}
			}
		}
	}

	return "", fmt.Errorf("acceptance criteria field not found")
}

// detectFeatureRequiredFields detects required fields for Feature/Epic issue type
func detectFeatureRequiredFields(ctx context.Context, client *http.Client, baseURL, email, token, projectKey, featureType string) (string, error) {
	endpoint := fmt.Sprintf("%s/rest/api/3/issue/createmeta?projectKeys=%s&issuetypeNames=%s&expand=projects.issuetypes.fields", baseURL, projectKey, featureType)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return "", err
	}

	req.SetBasicAuth(email, token)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get create meta: status %d", resp.StatusCode)
	}

	var createMeta map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&createMeta); err != nil {
		return "", err
	}

	projects, ok := createMeta["projects"].([]interface{})
	if !ok || len(projects) == 0 {
		return "", fmt.Errorf("no projects found")
	}

	project := projects[0].(map[string]interface{})
	issueTypes, ok := project["issuetypes"].([]interface{})
	if !ok {
		return "", fmt.Errorf("no issue types found")
	}

	for _, issueTypeData := range issueTypes {
		issueType := issueTypeData.(map[string]interface{})
		if name, ok := issueType["name"].(string); ok && name == featureType {
			fields, ok := issueType["fields"].(map[string]interface{})
			if !ok {
				return "{}", nil
			}

			requiredFields := make(map[string]interface{})

			for fieldKey, fieldData := range fields {
				if fieldMap, ok := fieldData.(map[string]interface{}); ok {
					if required, ok := fieldMap["required"].(bool); ok && required {
						// Skip standard fields
						if fieldKey != "project" && fieldKey != "issuetype" && fieldKey != "summary" && fieldKey != "description" {
							// Check if field has allowedValues (select field)
							if allowedValues, ok := fieldMap["allowedValues"].([]interface{}); ok && len(allowedValues) > 0 {
								// Take first allowed value as default
								if firstValue, ok := allowedValues[0].(map[string]interface{}); ok {
									if id, ok := firstValue["id"].(string); ok {
										requiredFields[fieldKey] = map[string]string{"id": id}
									}
								}
							}
						}
					}
				}
			}

			if len(requiredFields) > 0 {
				if jsonData, err := json.Marshal(requiredFields); err == nil {
					return string(jsonData), nil
				}
			}

			return "{}", nil
		}
	}

	return "", fmt.Errorf("feature issue type '%s' not found", featureType)
}

// IssueTypeInfo contains information about a Jira issue type
type IssueTypeInfo struct {
	ID          string
	Name        string
	Description string
	IsSubtask   bool
}

// getAvailableIssueTypes fetches available issue types from Jira for a project
func getAvailableIssueTypes(jiraURL, email, token, projectKey string) ([]IssueTypeInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := &http.Client{Timeout: 30 * time.Second}
	baseURL := strings.TrimSuffix(jiraURL, "/")

	// Get project issue types from createmeta API
	endpoint := fmt.Sprintf("%s/rest/api/3/issue/createmeta?projectKeys=%s&expand=projects.issuetypes", baseURL, projectKey)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(email, token)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get issue types: status %d", resp.StatusCode)
	}

	var createMeta map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&createMeta); err != nil {
		return nil, err
	}

	projects, ok := createMeta["projects"].([]interface{})
	if !ok || len(projects) == 0 {
		return nil, fmt.Errorf("no projects found")
	}

	project := projects[0].(map[string]interface{})
	issueTypes, ok := project["issuetypes"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("no issue types found")
	}

	var result []IssueTypeInfo
	for _, issueTypeData := range issueTypes {
		issueType := issueTypeData.(map[string]interface{})
		
		info := IssueTypeInfo{}
		if id, ok := issueType["id"].(string); ok {
			info.ID = id
		}
		if name, ok := issueType["name"].(string); ok {
			info.Name = name
		}
		if desc, ok := issueType["description"].(string); ok {
			info.Description = desc
		}
		if subtask, ok := issueType["subtask"].(bool); ok {
			info.IsSubtask = subtask
		}

		result = append(result, info)
	}

	return result, nil
}

// selectIssueType allows user to select an issue type from available options
func selectIssueType(reader *bufio.Reader, purpose string, issueTypes []IssueTypeInfo, onlySubtasks bool) string {
	// Filter issue types based on purpose
	var filtered []IssueTypeInfo
	for _, issueType := range issueTypes {
		if onlySubtasks {
			if issueType.IsSubtask {
				filtered = append(filtered, issueType)
			}
		} else {
			if !issueType.IsSubtask {
				filtered = append(filtered, issueType)
			}
		}
	}

	if len(filtered) == 0 {
		fmt.Printf("No se encontraron tipos de issue válidos para %s\n", purpose)
		return promptForInput(reader, fmt.Sprintf("Ingrese manualmente el tipo para %s", purpose), "")
	}

	fmt.Printf("Tipos de issue disponibles para %s:\n", purpose)
	fmt.Println()
	
	for i, issueType := range filtered {
		description := issueType.Description
		if description == "" {
			description = "Sin descripción"
		}
		fmt.Printf("  %d. %s - %s\n", i+1, issueType.Name, description)
	}
	fmt.Println()

	for {
		input := promptForInput(reader, fmt.Sprintf("Seleccione el número para %s (1-%d)", purpose, len(filtered)), "1")
		
		if input == "" {
			return filtered[0].Name
		}

		if num := parseNumber(input); num >= 1 && num <= len(filtered) {
			selected := filtered[num-1]
			fmt.Printf("✓ Seleccionado: %s\n", selected.Name)
			return selected.Name
		}

		fmt.Printf("Por favor ingrese un número entre 1 y %d\n", len(filtered))
	}
}

// parseNumber converts string to number, returns 0 if invalid
func parseNumber(str string) int {
	if num, err := strconv.Atoi(strings.TrimSpace(str)); err == nil {
		return num
	}
	return 0
}
