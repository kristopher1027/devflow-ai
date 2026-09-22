package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

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
	repositoryRepo       repository.RepositoryRepository
	projectRepo          repository.ProjectRepository
	githubConnectionRepo repository.GitHubConnectionRepository
	snapshotRepo         repository.RepositorySnapshotRepository
	githubClient         github.RepositoryClient
}

func NewRepositorySyncService(
	repositoryRepo repository.RepositoryRepository,
	projectRepo repository.ProjectRepository,
	githubConnectionRepo repository.GitHubConnectionRepository,
	snapshotRepo repository.RepositorySnapshotRepository,
	githubClient github.RepositoryClient,
) RepositorySyncService {
	return &RepositorySyncServiceImpl{
		repositoryRepo:       repositoryRepo,
		projectRepo:          projectRepo,
		githubConnectionRepo: githubConnectionRepo,
		snapshotRepo:         snapshotRepo,
		githubClient:         githubClient,
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

	project, err := s.projectRepo.FindByID(ctx, repo.ProjectID)
	if err != nil {
		return nil, err
	}
	connection, err := s.githubConnectionRepo.FindByWorkspaceID(
		ctx,
		project.WorkspaceID,
	)
	if err != nil {
		return nil, err
	}
	commitSHA, err := s.githubClient.GetLatestCommitSHA(
		ctx,
		connection.InstallationID,
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
	existingSnapshot, err := s.snapshotRepo.FindByRepositoryIDAndCommitSHA(
		ctx,
		repositoryID,
		commitSHA,
	)
	if err == nil {
		return existingSnapshot, nil
	}

	if !errors.Is(err, repository.ErrRepositorySnapshotNotFound) {
		return nil, err
	}

	snapshot := &domain.RepositorySnapshot{
		ID:           uuid.NewString(),
		RepositoryID: repo.ID,
		CommitSHA:    commitSHA,
		Branch:       repo.DefaultBranch,
		CreatedAt:    time.Now().UTC(),
	}

	if err := s.snapshotRepo.Create(ctx, snapshot); err != nil {
		return nil, err
	}

	syncedAt := time.Now().UTC()

	if err := s.repositoryRepo.UpdateSyncStatus(
		ctx,
		repositoryID,
		"synced",
		&syncedAt,
	); err != nil {
		return nil, err
	}

	return snapshot, nil
}
