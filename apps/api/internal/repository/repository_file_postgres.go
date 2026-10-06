package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/kristopher1027/devflow-ai/internal/database"
	"github.com/kristopher1027/devflow-ai/internal/domain"
)

type PostgresRepositoryFileRepository struct {
	db *database.DB
}

func NewRepositoryFileRepository(
	db *database.DB,
) RepositoryFileRepository {
	return &PostgresRepositoryFileRepository{
		db: db,
	}
}

func (r *PostgresRepositoryFileRepository) Create(
	ctx context.Context,
	file *domain.RepositoryFile,
) error {
	_, err := r.db.Pool.Exec(
		ctx,
		`
		INSERT INTO repository_files (
			id,
			snapshot_id,
			path,
			size_bytes,
			language,
			content,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		`,
		file.ID,
		file.SnapshotID,
		file.Path,
		file.SizeBytes,
		file.Language,
		file.Content,
		file.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create repository file: %w", err)
	}

	return nil
}

func (r *PostgresRepositoryFileRepository) FindByID(
	ctx context.Context,
	id string,
) (*domain.RepositoryFile, error) {
	file := &domain.RepositoryFile{}

	err := r.db.Pool.QueryRow(
		ctx,
		`
		SELECT
			id,
			snapshot_id,
			path,
			size_bytes,
			language,
			content,
			created_at
		FROM repository_files
		WHERE id = $1
		`,
		id,
	).Scan(
		&file.ID,
		&file.SnapshotID,
		&file.Path,
		&file.SizeBytes,
		&file.Language,
		&file.Content,
		&file.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRepositoryFileNotFound
		}

		return nil, fmt.Errorf(
			"find repository file: %w",
			err,
		)
	}

	return file, nil
}

func (r *PostgresRepositoryFileRepository) ListBySnapshotID(
	ctx context.Context,
	snapshotID string,
) ([]*domain.RepositoryFile, error) {
	rows, err := r.db.Pool.Query(
		ctx,
		`
		SELECT
			id,
			snapshot_id,
			path,
			size_bytes,
			language,
			content,
			created_at
		FROM repository_files
		WHERE snapshot_id = $1
		ORDER BY path ASC
		`,
		snapshotID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"list repository files: %w",
			err,
		)
	}
	defer rows.Close()

	var files []*domain.RepositoryFile

	for rows.Next() {
		file := &domain.RepositoryFile{}

		if err := rows.Scan(
			&file.ID,
			&file.SnapshotID,
			&file.Path,
			&file.SizeBytes,
			&file.Language,
			&file.Content,
			&file.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf(
				"scan repository file: %w",
				err,
			)
		}

		files = append(files, file)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate repository files: %w",
			err,
		)
	}

	return files, nil
}

var _ RepositoryFileRepository = (*PostgresRepositoryFileRepository)(nil)
