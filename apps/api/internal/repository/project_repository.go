package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/kristopher1027/devflow-ai/internal/database"
	"github.com/kristopher1027/devflow-ai/internal/domain"
)

var ErrProjectNotFound = errors.New("project not found")

type ProjectRepository interface {
	Create(
		ctx context.Context,
		project *domain.Project,
	) error

	FindByID(
		ctx context.Context,
		id string,
	) (*domain.Project, error)

	ListByWorkspaceID(
		ctx context.Context,
		workspaceID string,
	) ([]*domain.Project, error)

	Delete(
		ctx context.Context,
		id string,
	) error
}

type PostgresProjectRepository struct {
	db *database.DB
}

func NewProjectRepository(
	db *database.DB,
) ProjectRepository {
	return &PostgresProjectRepository{
		db: db,
	}
}

func (r *PostgresProjectRepository) Create(
	ctx context.Context,
	project *domain.Project,
) error {
	query := `
		INSERT INTO projects (
			id,
			workspace_id,
			name,
			description,
			created_by,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.Pool.Exec(
		ctx,
		query,
		project.ID,
		project.WorkspaceID,
		project.Name,
		project.Description,
		project.CreatedBy,
		project.CreatedAt,
		project.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("create project: %w", err)
	}

	return nil
}

func (r *PostgresProjectRepository) FindByID(
	ctx context.Context,
	id string,
) (*domain.Project, error) {
	query := `
		SELECT
			id,
			workspace_id,
			name,
			description,
			created_by,
			created_at,
			updated_at
		FROM projects
		WHERE id = $1
	`

	var project domain.Project

	err := r.db.Pool.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&project.ID,
		&project.WorkspaceID,
		&project.Name,
		&project.Description,
		&project.CreatedBy,
		&project.CreatedAt,
		&project.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProjectNotFound
		}

		return nil, fmt.Errorf(
			"find project by ID: %w",
			err,
		)
	}

	return &project, nil
}

func (r *PostgresProjectRepository) ListByWorkspaceID(
	ctx context.Context,
	workspaceID string,
) ([]*domain.Project, error) {
	query := `
		SELECT
			id,
			workspace_id,
			name,
			description,
			created_by,
			created_at,
			updated_at
		FROM projects
		WHERE workspace_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.Pool.Query(
		ctx,
		query,
		workspaceID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"list projects by workspace ID: %w",
			err,
		)
	}
	defer rows.Close()

	projects := make([]*domain.Project, 0)

	for rows.Next() {
		var project domain.Project

		if err := rows.Scan(
			&project.ID,
			&project.WorkspaceID,
			&project.Name,
			&project.Description,
			&project.CreatedBy,
			&project.CreatedAt,
			&project.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf(
				"scan project: %w",
				err,
			)
		}

		projects = append(projects, &project)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate projects: %w",
			err,
		)
	}

	return projects, nil
}

func (r *PostgresProjectRepository) Delete(
	ctx context.Context,
	id string,
) error {
	query := `
		DELETE FROM projects
		WHERE id = $1
	`

	result, err := r.db.Pool.Exec(
		ctx,
		query,
		id,
	)
	if err != nil {
		return fmt.Errorf("delete project: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrProjectNotFound
	}

	return nil
}
