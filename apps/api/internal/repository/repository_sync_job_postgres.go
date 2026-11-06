package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/kristopher1027/devflow-ai/internal/database"
	"github.com/kristopher1027/devflow-ai/internal/domain"
)

type PostgresRepositorySyncJobRepository struct {
	db *database.DB
}

func NewRepositorySyncJobRepository(
	db *database.DB,
) *PostgresRepositorySyncJobRepository {
	return &PostgresRepositorySyncJobRepository{
		db: db,
	}
}

func (r *PostgresRepositorySyncJobRepository) Create(
	ctx context.Context,
	job *domain.RepositorySyncJob,
) error {
	const query = `
		INSERT INTO repository_sync_jobs (
			id, repository_id, status, attempts, failure_code,
			failure_message, created_at, updated_at, completed_at,
			lease_owner, lease_expires_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.db.Pool.Exec(
		ctx, query, job.ID, job.RepositoryID, job.Status, job.Attempts,
		job.FailureCode, job.FailureMessage, job.CreatedAt, job.UpdatedAt,
		job.CompletedAt,
	)

	return err
}

func (r *PostgresRepositorySyncJobRepository) FindByID(
	ctx context.Context,
	id string,
) (*domain.RepositorySyncJob, error) {
	const query = `
		SELECT id, repository_id, status, attempts, failure_code,
		       failure_message, created_at, updated_at, completed_at,
		       lease_owner, lease_expires_at
		FROM repository_sync_jobs
		WHERE id = $1
	`

	job := &domain.RepositorySyncJob{}

	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&job.ID, &job.RepositoryID, &job.Status, &job.Attempts,
		&job.FailureCode, &job.FailureMessage, &job.CreatedAt,
		&job.UpdatedAt, &job.CompletedAt,
		&job.LeaseOwner, &job.LeaseExpiresAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrRepositorySyncJobNotFound
	}
	if err != nil {
		return nil, err
	}

	return job, nil
}

func (r *PostgresRepositorySyncJobRepository) ListByRepositoryID(
	ctx context.Context,
	repositoryID string,
) ([]*domain.RepositorySyncJob, error) {
	const query = `
		SELECT id, repository_id, status, attempts, failure_code,
		       failure_message, created_at, updated_at, completed_at
		FROM repository_sync_jobs
		WHERE repository_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Pool.Query(ctx, query, repositoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []*domain.RepositorySyncJob
	for rows.Next() {
		job := &domain.RepositorySyncJob{}
		if err := rows.Scan(
			&job.ID, &job.RepositoryID, &job.Status, &job.Attempts,
			&job.FailureCode, &job.FailureMessage, &job.CreatedAt,
			&job.UpdatedAt, &job.CompletedAt,
		); err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return jobs, nil
}

func (r *PostgresRepositorySyncJobRepository) MarkRunning(
	ctx context.Context,
	id string,
	attempts int,
) error {
	const query = `
		UPDATE repository_sync_jobs
		SET status = $2, attempts = $3, updated_at = now()
		WHERE id = $1
	`

	result, err := r.db.Pool.Exec(
		ctx, query, id, domain.RepositorySyncJobStatusRunning, attempts,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrRepositorySyncJobNotFound
	}
	return nil
}

func (r *PostgresRepositorySyncJobRepository) MarkSucceeded(
	ctx context.Context,
	id string,
	attempts int,
) error {
	const query = `
		UPDATE repository_sync_jobs
		SET status = $2, attempts = $3, failure_code = NULL,
		    failure_message = NULL, updated_at = now(), completed_at = now(),
		    lease_owner = NULL, lease_expires_at = NULL
		WHERE id = $1
	`

	result, err := r.db.Pool.Exec(
		ctx, query, id, domain.RepositorySyncJobStatusSucceeded, attempts,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrRepositorySyncJobNotFound
	}
	return nil
}

func (r *PostgresRepositorySyncJobRepository) MarkFailed(
	ctx context.Context,
	id string,
	attempts int,
	code string,
	message string,
) error {
	const query = `
		UPDATE repository_sync_jobs
		SET status = $2, attempts = $3, failure_code = $4,
		    failure_message = $5, updated_at = now(), completed_at = now(),
		    lease_owner = NULL, lease_expires_at = NULL
		WHERE id = $1
	`

	result, err := r.db.Pool.Exec(
		ctx, query, id, domain.RepositorySyncJobStatusFailed, attempts,
		code, message,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrRepositorySyncJobNotFound
	}
	return nil
}

// FailInterrupted marks only pending or running jobs older than the startup
// cutoff as failed. Completed jobs and jobs created by this process are safe.
func (r *PostgresRepositorySyncJobRepository) FailInterrupted(
	ctx context.Context,
	updatedBefore time.Time,
	code string,
	message string,
) (int64, error) {
	result, err := r.db.Pool.Exec(ctx, `
		UPDATE repository_sync_jobs
		SET status = $1, failure_code = $2, failure_message = $3,
		    updated_at = NOW(), completed_at = NOW(),
		    lease_owner = NULL, lease_expires_at = NULL
		WHERE status IN ($4, $5) AND updated_at < $6
	`,
		domain.RepositorySyncJobStatusFailed, code, message,
		domain.RepositorySyncJobStatusPending,
		domain.RepositorySyncJobStatusRunning,
		updatedBefore,
	)
	if err != nil {
		return 0, fmt.Errorf("fail interrupted repository sync jobs: %w", err)
	}
	return result.RowsAffected(), nil
}


// ClaimNext uses PostgreSQL row locking to ensure that concurrent workers
// cannot claim the same job. Expired running jobs are eligible for recovery.
func (r *PostgresRepositorySyncJobRepository) ClaimNext(
	ctx context.Context,
	workerID string,
	leaseDuration time.Duration,
) (*domain.RepositorySyncJob, error) {
	if workerID == "" {
		return nil, errors.New("repository sync worker ID is required")
	}
	if leaseDuration <= 0 {
		return nil, errors.New("repository sync job lease duration must be positive")
	}

	const query = `
		WITH candidate AS (
			SELECT id
			FROM repository_sync_jobs
			WHERE status = $3
			   OR (
					status = $4
					AND (
						lease_expires_at IS NULL
						OR lease_expires_at <= NOW()
					)
			   )
			ORDER BY created_at ASC, id ASC
			LIMIT 1
			FOR UPDATE SKIP LOCKED
		)
		UPDATE repository_sync_jobs AS job
		SET status = $4,
		    attempts = job.attempts + 1,
		    lease_owner = $1,
		    lease_expires_at = NOW() + $2::interval,
		    updated_at = NOW()
		FROM candidate
		WHERE job.id = candidate.id
		RETURNING job.id, job.repository_id, job.status, job.attempts,
		          job.failure_code, job.failure_message, job.created_at,
		          job.updated_at, job.completed_at, job.lease_owner,
		          job.lease_expires_at
	`

	job := &domain.RepositorySyncJob{}
	err := r.db.Pool.QueryRow(
		ctx,
		query,
		workerID,
		leaseDuration.String(),
		domain.RepositorySyncJobStatusPending,
		domain.RepositorySyncJobStatusRunning,
	).Scan(
		&job.ID,
		&job.RepositoryID,
		&job.Status,
		&job.Attempts,
		&job.FailureCode,
		&job.FailureMessage,
		&job.CreatedAt,
		&job.UpdatedAt,
		&job.CompletedAt,
		&job.LeaseOwner,
		&job.LeaseExpiresAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("claim repository sync job: %w", err)
	}

	return job, nil
}
