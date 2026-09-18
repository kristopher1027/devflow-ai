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
	jobs             repository.RepositorySyncJobRepository
	repositories     repository.RepositoryRepository
	projects         repository.ProjectRepository
	workspaceMembers repository.WorkspaceMemberRepository
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

func (s *repositorySyncJobService) FindByID(
	ctx context.Context,
	requesterID string,
	id string,
) (*domain.RepositorySyncJob, error) {
	job, err := s.jobs.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	repo, err := s.repositories.FindByID(ctx, job.RepositoryID)
	if err != nil {
		return nil, err
	}

	project, err := s.projects.FindByID(ctx, repo.ProjectID)
	if err != nil {
		return nil, err
	}

	_, err = s.workspaceMembers.Find(
		ctx,
		project.WorkspaceID,
		requesterID,
	)
	if err != nil {
		if errors.Is(err, repository.ErrWorkspaceMemberNotFound) {
			return nil, ErrRepositorySyncJobUnauthorized
		}

		return nil, err
	}

	return job, nil
}
