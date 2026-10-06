package service

import (
	"context"

	"github.com/kristopher1027/devflow-ai/internal/integration/github"
	"github.com/kristopher1027/devflow-ai/internal/repository"
)

type RepositoryIngestionService interface {
	Ingest(
		ctx context.Context,
		snapshotID string,
	) error
}

type RepositoryIngestionServiceImpl struct {
	repositoryRepo repository.RepositoryRepository
	snapshotRepo   repository.RepositorySnapshotRepository
	fileRepo       repository.RepositoryFileRepository
	githubClient   github.RepositoryClient
}

func NewRepositoryIngestionService(
	repositoryRepo repository.RepositoryRepository,
	snapshotRepo repository.RepositorySnapshotRepository,
	fileRepo repository.RepositoryFileRepository,
	githubClient github.RepositoryClient,
) RepositoryIngestionService {
	return &RepositoryIngestionServiceImpl{
		repositoryRepo: repositoryRepo,
		snapshotRepo:   snapshotRepo,
		fileRepo:       fileRepo,
		githubClient:   githubClient,
	}
}

func (s *RepositoryIngestionServiceImpl) Ingest(
	ctx context.Context,
	snapshotID string,
) error {
	return nil
}

var _ RepositoryIngestionService = (*RepositoryIngestionServiceImpl)(nil)
