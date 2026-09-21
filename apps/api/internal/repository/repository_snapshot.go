package repository

import (
	"context"
	"errors"

	"github.com/kristopher1027/devflow-ai/internal/domain"
)

var ErrRepositorySnapshotNotFound = errors.New(
	"repository snapshot not found",
)

type RepositorySnapshotRepository interface {
	Create(
		ctx context.Context,
		snapshot *domain.RepositorySnapshot,
	) error

	FindByID(
		ctx context.Context,
		id string,
	) (*domain.RepositorySnapshot, error)

	FindByRepositoryIDAndCommitSHA(
		ctx context.Context,
		repositoryID string,
		commitSHA string,
	) (*domain.RepositorySnapshot, error)

	ListByRepositoryID(
		ctx context.Context,
		repositoryID string,
	) ([]*domain.RepositorySnapshot, error)
}
