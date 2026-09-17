package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/kristopher1027/devflow-ai/internal/database"
	"github.com/kristopher1027/devflow-ai/internal/domain"
)

var ErrGitHubRepositoryImportJobNotFound = errors.New("github repository import job not found")

type GitHubRepositoryImportJobRepository interface {
	Create(ctx context.Context, job *domain.GitHubRepositoryImportJob) error
	FindByID(ctx context.Context, id string) (*domain.GitHubRepositoryImportJob, error)
	MarkRunning(ctx context.Context, id string, attempts int) error
	MarkSucceeded(ctx context.Context, id string, attempts int) error
	MarkFailed(ctx context.Context, id string, attempts int, code string, message string) error
}

type PostgresGitHubRepositoryImportJobRepository struct {
	db *database.DB
}

func NewGitHubRepositoryImportJobRepository(db *database.DB) GitHubRepositoryImportJobRepository {
	return &PostgresGitHubRepositoryImportJobRepository{db: db}
}

func (r *PostgresGitHubRepositoryImportJobRepository) Create(
	ctx context.Context,
	job *domain.GitHubRepositoryImportJob,
) error {
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO github_repository_import_jobs (
			id, requester_id, project_id, status, attempts,
			failure_code, failure_message, created_at, updated_at,
			completed_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, job.ID, job.RequesterID, job.ProjectID, job.Status, job.Attempts,
		job.FailureCode, job.FailureMessage, job.CreatedAt, job.UpdatedAt, job.CompletedAt)
	if err != nil {
		return fmt.Errorf("create github repository import job: %w", err)
	}
	return nil
}

func (r *PostgresGitHubRepositoryImportJobRepository) FindByID(
	ctx context.Context,
	id string,
) (*domain.GitHubRepositoryImportJob, error) {
	var job domain.GitHubRepositoryImportJob
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, requester_id, project_id, status, attempts,
		       failure_code, failure_message, created_at, updated_at, completed_at
		FROM github_repository_import_jobs
		WHERE id = $1
	`, id).Scan(&job.ID, &job.RequesterID, &job.ProjectID, &job.Status,
		&job.Attempts, &job.FailureCode, &job.FailureMessage,
		&job.CreatedAt, &job.UpdatedAt, &job.CompletedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrGitHubRepositoryImportJobNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find github repository import job: %w", err)
	}
	return &job, nil
}

func (r *PostgresGitHubRepositoryImportJobRepository) MarkRunning(
	ctx context.Context,
	id string,
	attempts int,
) error {
	return r.update(ctx, id, domain.GitHubRepositoryImportJobStatusRunning, attempts, "", "", false)
}

func (r *PostgresGitHubRepositoryImportJobRepository) MarkSucceeded(
	ctx context.Context,
	id string,
	attempts int,
) error {
	return r.update(ctx, id, domain.GitHubRepositoryImportJobStatusSucceeded, attempts, "", "", true)
}

func (r *PostgresGitHubRepositoryImportJobRepository) MarkFailed(
	ctx context.Context,
	id string,
	attempts int,
	code string,
	message string,
) error {
	return r.update(ctx, id, domain.GitHubRepositoryImportJobStatusFailed, attempts, code, message, true)
}

func (r *PostgresGitHubRepositoryImportJobRepository) update(
	ctx context.Context,
	id string,
	status string,
	attempts int,
	code string,
	message string,
	completed bool,
) error {
	query := `UPDATE github_repository_import_jobs
		SET status = $1, attempts = $2, updated_at = NOW(),
		    failure_code = NULLIF($3, ''), failure_message = NULLIF($4, '')`
	if completed {
		query += `, completed_at = NOW()`
	}
	query += ` WHERE id = $5`
	args := []any{status, attempts, code, message, id}
	result, err := r.db.Pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update github repository import job: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrGitHubRepositoryImportJobNotFound
	}
	return nil
}
