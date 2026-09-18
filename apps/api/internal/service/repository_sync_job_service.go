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

type RepositorySyncJobServiceImpl struct {
	jobRepository     repository.RepositorySyncJobRepository
	repositoryRepo    repository.RepositoryRepository
	projectRepo       repository.ProjectRepository
	workspaceMemberRepo repository.WorkspaceMemberRepository
}

func NewRepositorySyncJobService(
	jobRepository repository.RepositorySyncJobRepository,
	repositoryRepo repository.RepositoryRepository,
	projectRepo repository.ProjectRepository,
	workspaceMemberRepo repository.WorkspaceMemberRepository,
) *RepositorySyncJobServiceImpl {
	return &RepositorySyncJobServiceImpl{
		jobRepository:       jobRepository,
		repositoryRepo:      repositoryRepo,
		projectRepo:         projectRepo,
		workspaceMemberRepo: workspaceMemberRepo,
	}
}

func (s *RepositorySyncJobServiceImpl) FindByID(
	ctx context.Context,
	requesterID string,
	id string,
) (*domain.RepositorySyncJob, error) {
	job, err := s.jobRepository.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	repo, err := s.repositoryRepo.FindByID(
		ctx,
		job.RepositoryID,
	)
	if err != nil {
		return nil, err
	}

	project, err := s.projectRepo.FindByID(
		ctx,
		repo.ProjectID,
	)
	if err != nil {
		return nil, err
	}

	_, err = s.workspaceMemberRepo.Find(
		ctx,
		project.WorkspaceID,
		requesterID,
	)
	if err != nil {
		if errors.Is(
			err,
			repository.ErrWorkspaceMemberNotFound,
		) {
			return nil, ErrRepositorySyncJobUnauthorized
		}

		return nil, err
	}

	return job, nil
}

var _ RepositorySyncJobService = (*RepositorySyncJobServiceImpl)(nil)
