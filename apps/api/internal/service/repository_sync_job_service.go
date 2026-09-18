package service

import (
	"context"
	"errors"

	"github.com/kristopher1027/devflow-ai/internal/domain"
	"github.com/kristopher1027/devflow-ai/internal/repository"
)

var ErrRepositorySyncJobUnauthorized = errors.New(
	"unauthorized to view repository sync job",
)

type RepositorySyncJobService interface {
	FindByID(
		ctx context.Context,
		requesterID string,
		id string,
	) (*domain.RepositorySyncJob, error)
}

type repositorySyncJobService struct {
	jobs              repository.RepositorySyncJobRepository
	repositories      repository.RepositoryRepository
	projects          repository.ProjectRepository
	workspaceMembers  repository.WorkspaceMemberRepository
}

func NewRepositorySyncJobService(
	jobs repository.RepositorySyncJobRepository,
	repositories repository.RepositoryRepository,
	projects repository.ProjectRepository,
	workspaceMembers repository.WorkspaceMemberRepository,
) RepositorySyncJobService {
	return &repositorySyncJobService{
		jobs:             jobs,
		repositories:     repositories,
		projects:         projects,
		workspaceMembers: workspaceMembers,
	}
}
