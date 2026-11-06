package repository

import (
	"context"
	"errors"
	"time"

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

	// FailInterrupted marks pending/running jobs updated before the cutoff
	// as failed and returns the number of jobs changed.
	FailInterrupted(
		ctx context.Context,
		updatedBefore time.Time,
		code string,
		message string,
	) (int64, error)
}
