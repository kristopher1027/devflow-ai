package repository

import (
	"context"
	"errors"

	"github.com/kristopher1027/devflow-ai/internal/domain"
)

var ErrRepositorySyncJobNotFound = errors.New(
	"repository sync job not found",
)

type RepositorySyncJobRepository interface {
	Create(
		ctx context.Context,
		job *domain.RepositorySyncJob,
	) error

	FindByID(
		ctx context.Context,
		id string,
	) (*domain.RepositorySyncJob, error)

	ListByRepositoryID(
		ctx context.Context,
		repositoryID string,
	) ([]*domain.RepositorySyncJob, error)

	MarkRunning(
		ctx context.Context,
		id string,
		attempts int,
	) error

	MarkSucceeded(
		ctx context.Context,
		id string,
		attempts int,
	) error

	MarkFailed(
		ctx context.Context,
		id string,
		attempts int,
		code string,
		message string,
	) error
}
