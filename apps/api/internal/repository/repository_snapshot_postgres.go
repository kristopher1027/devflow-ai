package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/kristopher1027/devflow-ai/internal/database"
	"github.com/kristopher1027/devflow-ai/internal/domain"
)

type PostgresRepositorySnapshotRepository struct {
	db *database.DB
}

func NewRepositorySnapshotRepository(
	db *database.DB,
) RepositorySnapshotRepository {
	return &PostgresRepositorySnapshotRepository{
		db: db,
	}
}

func (r *PostgresRepositorySnapshotRepository) Create(
	ctx context.Context,
	snapshot *domain.RepositorySnapshot,
) error {
	_, err := r.db.Pool.Exec(
		ctx,
		`
		INSERT INTO repository_snapshots (
			id,
			repository_id,
			commit_sha,
			branch,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5)
		`,
		snapshot.ID,
		snapshot.RepositoryID,
		snapshot.CommitSHA,
		snapshot.Branch,
		snapshot.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create repository snapshot: %w", err)
	}

	return nil
}

func (r *PostgresRepositorySnapshotRepository) FindByID(
	ctx context.Context,
	id string,
) (*domain.RepositorySnapshot, error) {
	snapshot := &domain.RepositorySnapshot{}

	err := r.db.Pool.QueryRow(
		ctx,
		`
		SELECT
			id,
			repository_id,
			commit_sha,
			branch,
			created_at
		FROM repository_snapshots
		WHERE id = $1
		`,
		id,
	).Scan(
		&snapshot.ID,
		&snapshot.RepositoryID,
		&snapshot.CommitSHA,
		&snapshot.Branch,
		&snapshot.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRepositorySnapshotNotFound
		}

		return nil, fmt.Errorf("find repository snapshot: %w", err)
	}

	return snapshot, nil
}

func (r *PostgresRepositorySnapshotRepository) FindByRepositoryIDAndCommitSHA(
	ctx context.Context,
	repositoryID string,
	commitSHA string,
) (*domain.RepositorySnapshot, error) {
	snapshot := &domain.RepositorySnapshot{}

	err := r.db.Pool.QueryRow(
		ctx,
		`
		SELECT
			id,
			repository_id,
			commit_sha,
			branch,
			created_at
		FROM repository_snapshots
		WHERE repository_id = $1
		  AND commit_sha = $2
		`,
		repositoryID,
		commitSHA,
	).Scan(
		&snapshot.ID,
		&snapshot.RepositoryID,
		&snapshot.CommitSHA,
		&snapshot.Branch,
		&snapshot.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRepositorySnapshotNotFound
		}

		return nil, fmt.Errorf(
			"find repository snapshot by commit: %w",
			err,
		)
	}

	return snapshot, nil
}

func (r *PostgresRepositorySnapshotRepository) ListByRepositoryID(
	ctx context.Context,
	repositoryID string,
) ([]*domain.RepositorySnapshot, error) {
	rows, err := r.db.Pool.Query(
		ctx,
		`
		SELECT
			id,
			repository_id,
			commit_sha,
			branch,
			created_at
		FROM repository_snapshots
		WHERE repository_id = $1
		ORDER BY created_at DESC
		`,
		repositoryID,
	)
	if err != nil {
		return nil, fmt.Errorf("list repository snapshots: %w", err)
	}
	defer rows.Close()

	var snapshots []*domain.RepositorySnapshot

	for rows.Next() {
		snapshot := &domain.RepositorySnapshot{}

		if err := rows.Scan(
			&snapshot.ID,
			&snapshot.RepositoryID,
			&snapshot.CommitSHA,
			&snapshot.Branch,
			&snapshot.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf(
				"scan repository snapshot: %w",
				err,
			)
		}

		snapshots = append(snapshots, snapshot)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate repository snapshots: %w",
			err,
		)
	}

	return snapshots, nil
}

var _ RepositorySnapshotRepository = (*PostgresRepositorySnapshotRepository)(nil)
