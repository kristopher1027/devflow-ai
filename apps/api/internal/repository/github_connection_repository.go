package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/kristopher1027/devflow-ai/internal/database"
	"github.com/kristopher1027/devflow-ai/internal/domain"
)

var ErrGitHubConnectionNotFound = errors.New("github connection not found")

type GitHubConnectionRepository interface {
	Create(
		ctx context.Context,
		connection *domain.GitHubConnection,
	) error

	FindByWorkspaceID(
		ctx context.Context,
		workspaceID string,
	) (*domain.GitHubConnection, error)

	FindByID(
		ctx context.Context,
		id string,
	) (*domain.GitHubConnection, error)

	FindByInstallationID(
		ctx context.Context,
		installationID string,
	) (*domain.GitHubConnection, error)

	UpdateStatus(
		ctx context.Context,
		id string,
		status string,
	) error
}

type PostgresGitHubConnectionRepository struct {
	db *database.DB
}

func NewGitHubConnectionRepository(
	db *database.DB,
) GitHubConnectionRepository {
	return &PostgresGitHubConnectionRepository{db: db}
}

func (r *PostgresGitHubConnectionRepository) Create(
	ctx context.Context,
	connection *domain.GitHubConnection,
) error {
	query := `
		INSERT INTO github_connections (
			id,
			workspace_id,
			installation_id,
			account_login,
			status,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.Pool.Exec(
		ctx,
		query,
		connection.ID,
		connection.WorkspaceID,
		connection.InstallationID,
		connection.AccountLogin,
		connection.Status,
		connection.CreatedAt,
		connection.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create github connection: %w", err)
	}

	return nil
}

func (r *PostgresGitHubConnectionRepository) FindByWorkspaceID(
	ctx context.Context,
	workspaceID string,
) (*domain.GitHubConnection, error) {
	return r.findOne(
		ctx,
		`WHERE workspace_id = $1`,
		workspaceID,
		"find github connection by workspace ID",
	)
}

func (r *PostgresGitHubConnectionRepository) FindByID(
	ctx context.Context,
	id string,
) (*domain.GitHubConnection, error) {
	return r.findOne(
		ctx,
		`WHERE id = $1`,
		id,
		"find github connection by ID",
	)
}

func (r *PostgresGitHubConnectionRepository) FindByInstallationID(
	ctx context.Context,
	installationID string,
) (*domain.GitHubConnection, error) {
	return r.findOne(
		ctx,
		`WHERE installation_id = $1`,
		installationID,
		"find github connection by installation ID",
	)
}

func (r *PostgresGitHubConnectionRepository) findOne(
	ctx context.Context,
	condition string,
	value string,
	operation string,
) (*domain.GitHubConnection, error) {
	query := `
		SELECT
			id,
			workspace_id,
			installation_id,
			account_login,
			status,
			created_at,
			updated_at
		FROM github_connections
		` + condition

	var connection domain.GitHubConnection
	err := r.db.Pool.QueryRow(ctx, query, value).Scan(
		&connection.ID,
		&connection.WorkspaceID,
		&connection.InstallationID,
		&connection.AccountLogin,
		&connection.Status,
		&connection.CreatedAt,
		&connection.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrGitHubConnectionNotFound
		}

		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	return &connection, nil
}

func (r *PostgresGitHubConnectionRepository) UpdateStatus(
	ctx context.Context,
	id string,
	status string,
) error {
	switch status {
	case domain.GitHubConnectionStatusPending,
		domain.GitHubConnectionStatusActive,
		domain.GitHubConnectionStatusDisconnected,
		domain.GitHubConnectionStatusError:
	default:
		return domain.ErrGitHubConnectionStatusInvalid
	}

	result, err := r.db.Pool.Exec(
		ctx,
		`UPDATE github_connections
		 SET status = $1, updated_at = NOW()
		 WHERE id = $2`,
		status,
		id,
	)
	if err != nil {
		return fmt.Errorf("update github connection status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrGitHubConnectionNotFound
	}

	return nil
}
