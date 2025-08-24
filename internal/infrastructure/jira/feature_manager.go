package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"historiadorgo/internal/domain/entities"
	"historiadorgo/internal/infrastructure/config"
)

type FeatureManager struct {
	jiraClient *JiraClient
	config     *config.Config
}

type JiraSearchResponse struct {
	Issues []JiraIssue `json:"issues"`
	Total  int         `json:"total"`
}

func NewFeatureManager(jiraClient *JiraClient, cfg *config.Config) *FeatureManager {
	return &FeatureManager{
		jiraClient: jiraClient,
		config:     cfg,
	}
}

func (fm *FeatureManager) CreateOrGetFeature(ctx context.Context, description, projectKey string) (*entities.FeatureResult, error) {
	result := entities.NewFeatureResult(description)

	if fm.isJiraKey(description) {
		if err := fm.jiraClient.ValidateParentIssue(ctx, description); err != nil {
			result.SetError(fmt.Sprintf("Parent issue validation failed: %v", err))
			return result, nil
		}
		result.SetExisting(description)
		return result, nil
	}

	normalizedDesc := fm.normalizeDescription(description)
	result.SetNormalizedDescription(normalizedDesc)

	existingKey, err := fm.SearchExistingFeature(ctx, description, projectKey)
	if err != nil {
		result.SetError(fmt.Sprintf("Error searching existing features: %v", err))
		return result, nil
	}

	if existingKey != "" {
		result.SetExisting(existingKey)
		return result, nil
	}

	issuePayload := fm.buildFeaturePayload(description, projectKey)

	issue, err := fm.jiraClient.createIssue(ctx, issuePayload)
	if err != nil {
		result.SetError(fmt.Sprintf("Error creating feature: %v", err))
		return result, nil
	}

	issueURL := fmt.Sprintf("%s/browse/%s", fm.jiraClient.baseURL, issue.Key)
	result.SetSuccess(issue.Key, issueURL, true)

	return result, nil
}

func (fm *FeatureManager) SearchExistingFeature(ctx context.Context, description, projectKey string) (string, error) {
	normalizedDesc := fm.normalizeDescription(description)

	jql := fmt.Sprintf(
		`project = "%s" AND issuetype = "%s" AND summary ~ "%s"`,
		projectKey,
		fm.config.FeatureIssueType,
		fm.escapeJQLString(normalizedDesc),
	)

	searchURL := fmt.Sprintf("/rest/api/3/search?jql=%s&fields=key,summary", url.QueryEscape(jql))

	req, err := fm.jiraClient.createRequest(ctx, "GET", searchURL, nil)
	if err != nil {
		return "", fmt.Errorf("error creating search request: %w", err)
	}

	resp, err := fm.jiraClient.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error executing search: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("search failed with status: %d", resp.StatusCode)
	}

	var searchResp JiraSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return "", fmt.Errorf("error decoding search response: %w", err)
	}

	for _, issue := range searchResp.Issues {
		if summary, ok := issue.Fields["summary"].(string); ok {
			existingNormalized := fm.normalizeDescription(summary)
			if fm.isSimilarDescription(normalizedDesc, existingNormalized) {
				return issue.Key, nil
			}
		}
	}

	return "", nil
}

func (fm *FeatureManager) ValidateFeatureRequiredFields(ctx context.Context, projectKey string) ([]string, error) {
	endpoint := fmt.Sprintf("/rest/api/3/issue/createmeta?projectKeys=%s&expand=projects.issuetypes.fields", projectKey)

	req, err := fm.jiraClient.createRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := fm.jiraClient.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error getting create meta: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error getting create meta: status %d", resp.StatusCode)
	}

	var createMeta map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&createMeta); err != nil {
		return nil, fmt.Errorf("error decoding create meta: %w", err)
	}

	projects, ok := createMeta["projects"].([]interface{})
	if !ok || len(projects) == 0 {
		return nil, fmt.Errorf("no projects found in create meta")
	}

	project := projects[0].(map[string]interface{})
	issueTypes, ok := project["issuetypes"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("no issue types found")
	}

	for _, issueTypeData := range issueTypes {
		issueType := issueTypeData.(map[string]interface{})
		if name, ok := issueType["name"].(string); ok && name == fm.config.FeatureIssueType {
			fields, ok := issueType["fields"].(map[string]interface{})
			if !ok {
				return []string{}, nil
			}

			var requiredFields []string
			for fieldKey, fieldData := range fields {
				if fieldMap, ok := fieldData.(map[string]interface{}); ok {
					if required, ok := fieldMap["required"].(bool); ok && required {
						if fieldKey != "project" && fieldKey != "issuetype" && fieldKey != "summary" && fieldKey != "description" {
							if name, ok := fieldMap["name"].(string); ok {
								requiredFields = append(requiredFields, fmt.Sprintf("%s (%s)", name, fieldKey))
							} else {
								requiredFields = append(requiredFields, fieldKey)
							}
						}
					}
				}
			}
			return requiredFields, nil
		}
	}

	return nil, fmt.Errorf("feature issue type '%s' not found", fm.config.FeatureIssueType)
}

func (fm *FeatureManager) buildFeaturePayload(description, projectKey string) map[string]interface{} {
	fields := map[string]interface{}{
		"project": map[string]interface{}{
			"key": projectKey,
		},
		"summary":     description,
		"description": CreateDescriptionADF(fmt.Sprintf("Feature creado automáticamente: %s", description)),
		"issuetype": map[string]interface{}{
			"name": fm.config.FeatureIssueType,
		},
	}

	if fm.config.FeatureRequiredFields != "" {
		var additionalFields map[string]interface{}
		if err := json.Unmarshal([]byte(fm.config.FeatureRequiredFields), &additionalFields); err == nil {
			for key, value := range additionalFields {
				fields[key] = value
			}
		}
	}

	return map[string]interface{}{
		"fields": fields,
	}
}

func (fm *FeatureManager) normalizeDescription(description string) string {
	desc := strings.ToLower(description)
	desc = strings.TrimSpace(desc)

	re := regexp.MustCompile(`[^\w\s]`)
	desc = re.ReplaceAllString(desc, "")

	re = regexp.MustCompile(`\s+`)
	desc = re.ReplaceAllString(desc, " ")

	return desc
}

func (fm *FeatureManager) isSimilarDescription(desc1, desc2 string) bool {
	words1 := strings.Fields(desc1)
	words2 := strings.Fields(desc2)

	if len(words1) == 0 || len(words2) == 0 {
		return desc1 == desc2
	}

	commonWords := 0
	wordMap := make(map[string]bool)

	for _, word := range words1 {
		if len(word) > 2 {
			wordMap[word] = true
		}
	}

	for _, word := range words2 {
		if len(word) > 2 && wordMap[word] {
			commonWords++
		}
	}

	totalWords := len(words1)
	if len(words2) > totalWords {
		totalWords = len(words2)
	}

	similarity := float64(commonWords) / float64(totalWords)
	return similarity >= 0.7
}

func (fm *FeatureManager) escapeJQLString(str string) string {
	str = strings.ReplaceAll(str, `"`, `\"`)
	str = strings.ReplaceAll(str, `\`, `\\`)
	return str
}

func (fm *FeatureManager) isJiraKey(str string) bool {
	re := regexp.MustCompile(`^[A-Z]+-\d+$`)
	return re.MatchString(str)
}
