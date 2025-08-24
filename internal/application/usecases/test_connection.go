package usecases

import (
	"context"
	"historiadorgo/internal/domain/repositories"
)

type TestConnectionUseCase struct {
	jiraRepo repositories.JiraRepository
}

func NewTestConnectionUseCase(jiraRepo repositories.JiraRepository) *TestConnectionUseCase {
	return &TestConnectionUseCase{
		jiraRepo: jiraRepo,
	}
}

func (uc *TestConnectionUseCase) Execute(ctx context.Context) error {
	return uc.jiraRepo.TestConnection(ctx)
}
