package repository

import (
	"context"
	"errors"

	"github.com/kristopher1027/devflow-ai/internal/domain"
)

var ErrRepositoryFileNotFound = errors.New(
	"repository file not found",
)

type RepositoryFileRepository interface {
	Create(
		ctx context.Context,
		file *domain.RepositoryFile,
	) error

	FindByID(
		ctx context.Context,
		id string,
	) (*domain.RepositoryFile, error)

	ListBySnapshotID(
		ctx context.Context,
		snapshotID string,
	) ([]*domain.RepositoryFile, error)
}
