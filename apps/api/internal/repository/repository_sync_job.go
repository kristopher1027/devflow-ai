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

	// ClaimNext atomically claims the oldest pending job or a running job
	// whose lease has expired. It returns nil when no job is currently claimable.
	ClaimNext(
		ctx context.Context,
		workerID string,
		leaseDuration time.Duration,
	) (*domain.RepositorySyncJob, error)

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
