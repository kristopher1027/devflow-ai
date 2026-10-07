package repository

import (
	"context"
	"errors"

	"github.com/kristopher1027/devflow-ai/internal/domain"
)

var ErrRepositoryChunkNotFound = errors.New(
	"repository chunk not found",
)

type RepositoryChunkRepository interface {
	Create(
		ctx context.Context,
		chunk *domain.RepositoryChunk,
	) error

	FindByID(
		ctx context.Context,
		id string,
	) (*domain.RepositoryChunk, error)

	ListByFileID(
		ctx context.Context,
		fileID string,
	) ([]*domain.RepositoryChunk, error)

	ReplaceByFileID(
		ctx context.Context,
		fileID string,
		chunks []*domain.RepositoryChunk,
	) error

	SearchBySnapshotID(
		ctx context.Context,
		snapshotID string,
		query string,
		limit int,
	) ([]*domain.RepositoryChunkSearchResult, error)
}
