package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/kristopher1027/devflow-ai/internal/database"
	"github.com/kristopher1027/devflow-ai/internal/domain"
)

var ErrWorkspaceNotFound = errors.New("workspace not found")

type WorkspaceRepository interface {
	Create(
		ctx context.Context,
		workspace *domain.Workspace,
	) error

	FindByID(
		ctx context.Context,
		id string,
	) (*domain.Workspace, error)

	ListByOwnerID(
		ctx context.Context,
		ownerID string,
	) ([]*domain.Workspace, error)

	Delete(
		ctx context.Context,
		id string,
	) error
}

type PostgresWorkspaceRepository struct {
	db *database.DB
}

func NewWorkspaceRepository(
	db *database.DB,
) WorkspaceRepository {
	return &PostgresWorkspaceRepository{
		db: db,
	}
}

func (r *PostgresWorkspaceRepository) Create(
	ctx context.Context,
	workspace *domain.Workspace,
) error {
	query := `
		INSERT INTO workspaces (
			id,
			owner_id,
			name,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.Pool.Exec(
		ctx,
		query,
		workspace.ID,
		workspace.OwnerID,
		workspace.Name,
		workspace.CreatedAt,
		workspace.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("create workspace: %w", err)
	}

	return nil
}

func (r *PostgresWorkspaceRepository) FindByID(
	ctx context.Context,
	id string,
) (*domain.Workspace, error) {
	query := `
		SELECT
			id,
			owner_id,
			name,
			created_at,
			updated_at
		FROM workspaces
		WHERE id = $1
	`

	var workspace domain.Workspace

	err := r.db.Pool.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&workspace.ID,
		&workspace.OwnerID,
		&workspace.Name,
		&workspace.CreatedAt,
		&workspace.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrWorkspaceNotFound
		}

		return nil, fmt.Errorf(
			"find workspace by ID: %w",
			err,
		)
	}

	return &workspace, nil
}

func (r *PostgresWorkspaceRepository) ListByOwnerID(
	ctx context.Context,
	ownerID string,
) ([]*domain.Workspace, error) {
	query := `
		SELECT
			id,
			owner_id,
			name,
			created_at,
			updated_at
		FROM workspaces
		WHERE owner_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.Pool.Query(
		ctx,
		query,
		ownerID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"list workspaces by owner ID: %w",
			err,
		)
	}
	defer rows.Close()

	workspaces := make([]*domain.Workspace, 0)

	for rows.Next() {
		var workspace domain.Workspace

		if err := rows.Scan(
			&workspace.ID,
			&workspace.OwnerID,
			&workspace.Name,
			&workspace.CreatedAt,
			&workspace.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf(
				"scan workspace: %w",
				err,
			)
		}

		workspaces = append(workspaces, &workspace)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate workspaces: %w",
			err,
		)
	}

	return workspaces, nil
}

func (r *PostgresWorkspaceRepository) Delete(
	ctx context.Context,
	id string,
) error {
	query := `
		DELETE FROM workspaces
		WHERE id = $1
	`

	result, err := r.db.Pool.Exec(
		ctx,
		query,
		id,
	)
	if err != nil {
		return fmt.Errorf("delete workspace: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrWorkspaceNotFound
	}

	return nil
}
