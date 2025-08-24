package config

import (
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
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Warning: .env file not found: %v\n", err)
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
