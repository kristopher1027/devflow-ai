package service

import (
	"context"
	"errors"

	"github.com/kristopher1027/devflow-ai/internal/domain"
	"github.com/kristopher1027/devflow-ai/internal/integration/github"
	"github.com/kristopher1027/devflow-ai/internal/repository"
)

type RepositorySyncService interface {
	Sync(
		ctx context.Context,
		repositoryID string,
	) (*domain.RepositorySnapshot, error)
}

var (
	ErrRepositorySyncRepositoryNotFound = errors.New(
		"repository to sync not found",
	)

	ErrRepositorySyncBranchRequired = errors.New(
		"repository sync branch is required",
	)

	ErrRepositorySyncCommitNotFound = errors.New(
		"repository sync commit not found",
	)
	ErrRepositorySyncOwnerRequired = errors.New(
		"repository sync owner is required",
	)

	ErrRepositorySyncNameRequired = errors.New(
		"repository sync name is required",
	)
)

type RepositorySyncServiceImpl struct {
	repositoryRepo repository.RepositoryRepository
	snapshotRepo   repository.RepositorySnapshotRepository
	githubClient   github.RepositoryClient
}

func NewRepositorySyncService(
	repositoryRepo repository.RepositoryRepository,
	snapshotRepo repository.RepositorySnapshotRepository,
	githubClient github.RepositoryClient,
) RepositorySyncService {
	return &RepositorySyncServiceImpl{
		repositoryRepo: repositoryRepo,
		snapshotRepo:   snapshotRepo,
		githubClient:   githubClient,
	}
}

func (s *RepositorySyncServiceImpl) Sync(
	ctx context.Context,
	repositoryID string,
) (*domain.RepositorySnapshot, error) {
	repo, err := s.repositoryRepo.FindByID(ctx, repositoryID)
	if err != nil {
		if errors.Is(err, repository.ErrRepositoryNotFound) {
			return nil, ErrRepositorySyncRepositoryNotFound
		}

		return nil, err
	}

	if repo.DefaultBranch == "" {
		return nil, ErrRepositorySyncBranchRequired
	}

	if repo.Owner == "" {
		return nil, ErrRepositorySyncOwnerRequired
	}

	if repo.Name == "" {
		return nil, ErrRepositorySyncNameRequired
	}

	commitSHA, err := s.githubClient.GetLatestCommitSHA(
		ctx,
		repo.Owner,
		repo.Name,
		repo.DefaultBranch,
	)
	if err != nil {
		return nil, err
	}

	if commitSHA == "" {
		return nil, ErrRepositorySyncCommitNotFound
	}

	return nil, nil
}
