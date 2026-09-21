package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"time"

	"github.com/kristopher1027/devflow-ai/internal/database"
	"github.com/kristopher1027/devflow-ai/internal/domain"
)

var ErrRepositoryNotFound = errors.New("repository not found")

type RepositoryRepository interface {
	Create(
		ctx context.Context,
		repository *domain.Repository,
	) error

	FindByID(
		ctx context.Context,
		id string,
	) (*domain.Repository, error)

	ListByProjectID(
		ctx context.Context,
		projectID string,
	) ([]*domain.Repository, error)

	FindByProviderExternalID(
		ctx context.Context,
		provider string,
		externalID string,
	) (*domain.Repository, error)
	UpdateSyncStatus(
		ctx context.Context,
		id string,
		status string,
		lastSyncedAt *time.Time,
	) error

	Delete(
		ctx context.Context,
		id string,
	) error
}

type PostgresRepositoryRepository struct {
	db *database.DB
}

func NewRepositoryRepository(
	db *database.DB,
) RepositoryRepository {
	return &PostgresRepositoryRepository{
		db: db,
	}
}

func (r *PostgresRepositoryRepository) Create(
	ctx context.Context,
	repository *domain.Repository,
) error {
	query := `
		INSERT INTO repositories (
			id,
			project_id,
			provider,
			external_id,
			owner,
			name,
			full_name,
			default_branch,
			html_url,
			clone_url,
			is_private,
			sync_status,
			last_synced_at,
			created_at,
			updated_at
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9,
			$10,
			$11,
			$12,
			$13,
			$14,
			$15
		)
	`

	_, err := r.db.Pool.Exec(
		ctx,
		query,
		repository.ID,
		repository.ProjectID,
		repository.Provider,
		repository.ExternalID,
		repository.Owner,
		repository.Name,
		repository.FullName,
		repository.DefaultBranch,
		repository.HTMLURL,
		repository.CloneURL,
		repository.IsPrivate,
		repository.SyncStatus,
		repository.LastSyncedAt,
		repository.CreatedAt,
		repository.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("create repository: %w", err)
	}

	return nil
}

func (r *PostgresRepositoryRepository) FindByID(
	ctx context.Context,
	id string,
) (*domain.Repository, error) {
	query := `
		SELECT
			id,
			project_id,
			provider,
			external_id,
			owner,
			name,
			full_name,
			default_branch,
			html_url,
			clone_url,
			is_private,
			sync_status,
			last_synced_at,
			created_at,
			updated_at
		FROM repositories
		WHERE id = $1
	`

	var repository domain.Repository

	err := r.db.Pool.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&repository.ID,
		&repository.ProjectID,
		&repository.Provider,
		&repository.ExternalID,
		&repository.Owner,
		&repository.Name,
		&repository.FullName,
		&repository.DefaultBranch,
		&repository.HTMLURL,
		&repository.CloneURL,
		&repository.IsPrivate,
		&repository.SyncStatus,
		&repository.LastSyncedAt,
		&repository.CreatedAt,
		&repository.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRepositoryNotFound
		}

		return nil, fmt.Errorf(
			"find repository by ID: %w",
			err,
		)
	}

	return &repository, nil
}

func (r *PostgresRepositoryRepository) ListByProjectID(
	ctx context.Context,
	projectID string,
) ([]*domain.Repository, error) {
	query := `
		SELECT
			id,
			project_id,
			provider,
			external_id,
			owner,
			name,
			full_name,
			default_branch,
			html_url,
			clone_url,
			is_private,
			sync_status,
			last_synced_at,
			created_at,
			updated_at
		FROM repositories
		WHERE project_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.Pool.Query(
		ctx,
		query,
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"list repositories by project ID: %w",
			err,
		)
	}
	defer rows.Close()

	repositories := make([]*domain.Repository, 0)

	for rows.Next() {
		var repository domain.Repository

		if err := rows.Scan(
			&repository.ID,
			&repository.ProjectID,
			&repository.Provider,
			&repository.ExternalID,
			&repository.Owner,
			&repository.Name,
			&repository.FullName,
			&repository.DefaultBranch,
			&repository.HTMLURL,
			&repository.CloneURL,
			&repository.IsPrivate,
			&repository.SyncStatus,
			&repository.LastSyncedAt,
			&repository.CreatedAt,
			&repository.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf(
				"scan repository: %w",
				err,
			)
		}

		repositories = append(repositories, &repository)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate repositories: %w",
			err,
		)
	}

	return repositories, nil
}

func (r *PostgresRepositoryRepository) FindByProviderExternalID(
	ctx context.Context,
	provider string,
	externalID string,
) (*domain.Repository, error) {
	query := `
		SELECT
			id,
			project_id,
			provider,
			external_id,
			owner,
			name,
			full_name,
			default_branch,
			html_url,
			clone_url,
			is_private,
			sync_status,
			last_synced_at,
			created_at,
			updated_at
		FROM repositories
		WHERE provider = $1
		  AND external_id = $2
	`

	var repository domain.Repository

	err := r.db.Pool.QueryRow(
		ctx,
		query,
		provider,
		externalID,
	).Scan(
		&repository.ID,
		&repository.ProjectID,
		&repository.Provider,
		&repository.ExternalID,
		&repository.Owner,
		&repository.Name,
		&repository.FullName,
		&repository.DefaultBranch,
		&repository.HTMLURL,
		&repository.CloneURL,
		&repository.IsPrivate,
		&repository.SyncStatus,
		&repository.LastSyncedAt,
		&repository.CreatedAt,
		&repository.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRepositoryNotFound
		}

		return nil, fmt.Errorf(
			"find repository by provider and external ID: %w",
			err,
		)
	}

	return &repository, nil
}
func (r *PostgresRepositoryRepository) UpdateSyncStatus(
	ctx context.Context,
	id string,
	status string,
	lastSyncedAt *time.Time,
) error {
	query := `
		UPDATE repositories
		SET
			sync_status = $1,
			last_synced_at = $2,
			updated_at = now()
		WHERE id = $3
	`

	commandTag, err := r.db.Pool.Exec(
		ctx,
		query,
		status,
		lastSyncedAt,
		id,
	)
	if err != nil {
		return fmt.Errorf(
			"update repository sync status: %w",
			err,
		)
	}

	if commandTag.RowsAffected() == 0 {
		return ErrRepositoryNotFound
	}

	return nil
}

func (r *PostgresRepositoryRepository) Delete(
	ctx context.Context,
	id string,
) error {
	query := `
		DELETE FROM repositories
		WHERE id = $1
	`

	result, err := r.db.Pool.Exec(
		ctx,
		query,
		id,
	)
	if err != nil {
		return fmt.Errorf("delete repository: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrRepositoryNotFound
	}

	return nil
}
