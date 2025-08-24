package usecases

import (
	"context"
	"fmt"
	"historiadorgo/internal/domain/entities"
	"historiadorgo/internal/domain/repositories"
	"strings"
)

type ValidateFileUseCase struct {
	fileRepo repositories.FileRepository
	jiraRepo repositories.JiraRepository
}

type ValidationResult struct {
	TotalStories    int
	WithSubtasks    int
	TotalSubtasks   int
	WithParent      int
	InvalidSubtasks int
	Preview         string
}

func NewValidateFileUseCase(fileRepo repositories.FileRepository, jiraRepo repositories.JiraRepository) *ValidateFileUseCase {
	return &ValidateFileUseCase{
		fileRepo: fileRepo,
		jiraRepo: jiraRepo,
	}
}

func (uc *ValidateFileUseCase) Execute(ctx context.Context, filePath, projectKey string, rows int) (*ValidationResult, error) {
	if err := uc.fileRepo.ValidateFile(ctx, filePath); err != nil {
		return nil, err
	}

	stories, err := uc.fileRepo.ReadFile(ctx, filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file for statistics: %w", err)
	}

	result := uc.generateStatistics(stories)

	if projectKey != "" {
		if err := uc.jiraRepo.ValidateProject(ctx, projectKey); err != nil {
			return result, err
		}

		if err := uc.jiraRepo.ValidateSubtaskIssueType(ctx, projectKey); err != nil {
			return result, err
		}

		if err := uc.jiraRepo.ValidateFeatureIssueType(ctx); err != nil {
			return result, err
		}
	}

	return result, nil
}

func (uc *ValidateFileUseCase) generateStatistics(stories []*entities.UserStory) *ValidationResult {
	result := &ValidationResult{
		TotalStories: len(stories),
	}

	for _, story := range stories {
		if story.HasSubtareas() {
			result.WithSubtasks++
			result.TotalSubtasks += len(story.Subtareas)

			for _, subtarea := range story.Subtareas {
				if strings.TrimSpace(subtarea) == "" || len(subtarea) > 255 {
					result.InvalidSubtasks++
				}
			}
		}

		if story.HasParent() {
			result.WithParent++
		}
	}

	if len(stories) > 0 {
		result.Preview = uc.generatePreview(stories, 5)
	}

	return result
}

func (uc *ValidateFileUseCase) generatePreview(stories []*entities.UserStory, maxRows int) string {
	var preview strings.Builder

	preview.WriteString(fmt.Sprintf("%-30s %-50s %-20s %-15s\n", "TITULO", "DESCRIPCION", "SUBTAREAS", "PARENT"))
	preview.WriteString(strings.Repeat("-", 115) + "\n")

	for i, story := range stories {
		if i >= maxRows {
			break
		}

		titulo := story.Titulo
		if len(titulo) > 28 {
			titulo = titulo[:25] + "..."
		}

		descripcion := story.Descripcion
		if len(descripcion) > 48 {
			descripcion = descripcion[:45] + "..."
		}

		subtareas := ""
		if story.HasSubtareas() {
			subtareas = fmt.Sprintf("%d subtareas", len(story.Subtareas))
		}

		parent := story.Parent
		if len(parent) > 13 {
			parent = parent[:10] + "..."
		}

		preview.WriteString(fmt.Sprintf("%-30s %-50s %-20s %-15s\n", titulo, descripcion, subtareas, parent))
	}

	if len(stories) > maxRows {
		preview.WriteString(fmt.Sprintf("\n... y %d historias mas\n", len(stories)-maxRows))
	}

	return preview.String()
}
