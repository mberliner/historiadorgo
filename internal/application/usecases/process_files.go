package usecases

import (
	"context"
	"fmt"
	"historiadorgo/internal/domain/entities"
	"historiadorgo/internal/domain/repositories"
	"path/filepath"
)

type ProcessFilesUseCase struct {
	fileRepo    repositories.FileRepository
	jiraRepo    repositories.JiraRepository
	featureRepo repositories.FeatureManager
}

func NewProcessFilesUseCase(
	fileRepo repositories.FileRepository,
	jiraRepo repositories.JiraRepository,
	featureRepo repositories.FeatureManager,
) *ProcessFilesUseCase {
	return &ProcessFilesUseCase{
		fileRepo:    fileRepo,
		jiraRepo:    jiraRepo,
		featureRepo: featureRepo,
	}
}

func (uc *ProcessFilesUseCase) Execute(ctx context.Context, filePath, projectKey string, dryRun bool) (*entities.BatchResult, error) {
	// Solo validar inputs si no es dry-run
	if !dryRun {
		if err := uc.validateInputs(ctx, projectKey); err != nil {
			return nil, err
		}
	} else {
		// En dry-run, solo validar archivo
		if err := uc.fileRepo.ValidateFile(ctx, filePath); err != nil {
			return nil, fmt.Errorf("file validation failed: %w", err)
		}
	}

	stories, err := uc.fileRepo.ReadFile(ctx, filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	fileName := filepath.Base(filePath)
	batchResult := entities.NewBatchResult(fileName, len(stories), dryRun)

	for i, story := range stories {
		rowNumber := i + 2
		result := uc.processUserStory(ctx, story, projectKey, rowNumber, dryRun)
		batchResult.AddResult(result)
	}

	batchResult.Finish()

	if !dryRun && batchResult.SuccessfulRows > 0 {
		if err := uc.fileRepo.MoveToProcessed(ctx, filePath); err != nil {
			batchResult.AddError(fmt.Sprintf("Warning: could not move file to processed: %v", err))
		}
	}

	return batchResult, nil
}

func (uc *ProcessFilesUseCase) ProcessAllFiles(ctx context.Context, inputDir, projectKey string, dryRun bool) ([]*entities.BatchResult, error) {
	// Solo validar inputs si no es dry-run
	if !dryRun {
		if err := uc.validateInputs(ctx, projectKey); err != nil {
			return nil, err
		}
	}

	files, err := uc.fileRepo.GetPendingFiles(ctx, inputDir)
	if err != nil {
		return nil, fmt.Errorf("error getting pending files: %w", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no files found in %s", inputDir)
	}

	var results []*entities.BatchResult
	for _, file := range files {
		result, err := uc.Execute(ctx, file, projectKey, dryRun)
		if err != nil {
			result = entities.NewBatchResult(filepath.Base(file), 0, dryRun)
			result.AddError(fmt.Sprintf("Error processing file: %v", err))
			result.Finish()
		}
		results = append(results, result)
	}

	return results, nil
}

func (uc *ProcessFilesUseCase) validateInputs(ctx context.Context, projectKey string) error {
	if err := uc.jiraRepo.TestConnection(ctx); err != nil {
		return fmt.Errorf("jira connection failed: %w", err)
	}

	if err := uc.jiraRepo.ValidateProject(ctx, projectKey); err != nil {
		return fmt.Errorf("project validation failed: %w", err)
	}

	if err := uc.jiraRepo.ValidateSubtaskIssueType(ctx, projectKey); err != nil {
		return fmt.Errorf("subtask type validation failed: %w", err)
	}

	if err := uc.jiraRepo.ValidateFeatureIssueType(ctx); err != nil {
		return fmt.Errorf("feature type validation failed: %w", err)
	}

	return nil
}

func (uc *ProcessFilesUseCase) processUserStory(ctx context.Context, story *entities.UserStory, projectKey string, rowNumber int, dryRun bool) *entities.ProcessResult {
	result := entities.NewProcessResult(rowNumber)

	if dryRun {
		result.Success = true
		result.IssueKey = fmt.Sprintf("DRY-RUN-%d", rowNumber)
		result.IssueURL = fmt.Sprintf("https://dry-run.example.com/browse/DRY-RUN-%d", rowNumber)

		// Simular subtareas en dry-run
		if story.HasSubtareas() {
			for i, subtarea := range story.GetValidSubtareas() {
				subtaskKey := fmt.Sprintf("DRY-SUB-%d-%d", rowNumber, i+1)
				subtaskURL := fmt.Sprintf("https://dry-run.example.com/browse/%s", subtaskKey)
				result.AddSubtaskResult(subtarea, true, subtaskKey, subtaskURL, "")
			}
		}

		return result
	}

	// Handle feature creation/resolution if story has parent
	if story.HasParent() {
		featureResult, err := uc.featureRepo.CreateOrGetFeature(ctx, story.Parent, projectKey)
		if err != nil {
			result.Success = false
			result.ErrorMessage = fmt.Sprintf("feature handling failed: %v", err)
			return result
		}

		// Update story with the resolved parent key
		if featureResult.Success && featureResult.IssueKey != "" {
			story.Parent = featureResult.IssueKey
		} else if !featureResult.Success {
			result.Success = false
			result.ErrorMessage = fmt.Sprintf("feature creation failed: %s", featureResult.ErrorMessage)
			return result
		}
	}

	processResult, err := uc.jiraRepo.CreateUserStory(ctx, story, rowNumber)
	if err != nil {
		result.Success = false
		result.ErrorMessage = err.Error()
		return result
	}

	return processResult
}
