package repository

import (
	"context"
	"errors"

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
			id,
			repository_id,
			status,
			attempts,
			failure_code,
			failure_message,
			created_at,
			updated_at,
			completed_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.Pool.Exec(
		ctx,
		query,
		job.ID,
		job.RepositoryID,
		job.Status,
		job.Attempts,
		job.FailureCode,
		job.FailureMessage,
		job.CreatedAt,
		job.UpdatedAt,
		job.CompletedAt,
	)

	return err
}

func (r *PostgresRepositorySyncJobRepository) FindByID(
	ctx context.Context,
	id string,
) (*domain.RepositorySyncJob, error) {
	const query = `
		SELECT
			id,
			repository_id,
			status,
			attempts,
			failure_code,
			failure_message,
			created_at,
			updated_at,
			completed_at
		FROM repository_sync_jobs
		WHERE id = $1
	`

	job := &domain.RepositorySyncJob{}

	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&job.ID,
		&job.RepositoryID,
		&job.Status,
		&job.Attempts,
		&job.FailureCode,
		&job.FailureMessage,
		&job.CreatedAt,
		&job.UpdatedAt,
		&job.CompletedAt,
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
		SELECT
			id,
			repository_id,
			status,
			attempts,
			failure_code,
			failure_message,
			created_at,
			updated_at,
			completed_at
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
			&job.ID,
			&job.RepositoryID,
			&job.Status,
			&job.Attempts,
			&job.FailureCode,
			&job.FailureMessage,
			&job.CreatedAt,
			&job.UpdatedAt,
			&job.CompletedAt,
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
		SET
			status = $2,
			attempts = $3,
			updated_at = now()
		WHERE id = $1
	`

	result, err := r.db.Pool.Exec(
		ctx,
		query,
		id,
		domain.RepositorySyncJobStatusRunning,
		attempts,
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
		SET
			status = $2,
			attempts = $3,
			failure_code = NULL,
			failure_message = NULL,
			updated_at = now(),
			completed_at = now()
		WHERE id = $1
	`

	result, err := r.db.Pool.Exec(
		ctx,
		query,
		id,
		domain.RepositorySyncJobStatusSucceeded,
		attempts,
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
		SET
			status = $2,
			attempts = $3,
			failure_code = $4,
			failure_message = $5,
			updated_at = now(),
			completed_at = now()
		WHERE id = $1
	`

	result, err := r.db.Pool.Exec(
		ctx,
		query,
		id,
		domain.RepositorySyncJobStatusFailed,
		attempts,
		code,
		message,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrRepositorySyncJobNotFound
	}

	return nil
}
