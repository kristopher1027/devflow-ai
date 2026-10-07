package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/kristopher1027/devflow-ai/internal/database"
	"github.com/kristopher1027/devflow-ai/internal/domain"
)

type PostgresRepositoryChunkRepository struct {
	db *database.DB
}

func NewRepositoryChunkRepository(
	db *database.DB,
) RepositoryChunkRepository {
	return &PostgresRepositoryChunkRepository{db: db}
}

func (r *PostgresRepositoryChunkRepository) Create(
	ctx context.Context,
	chunk *domain.RepositoryChunk,
) error {
	_, err := r.db.Pool.Exec(
		ctx,
		`
		INSERT INTO repository_chunks (
			id,
			file_id,
			chunk_index,
			start_line,
			end_line,
			character_count,
			content,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`,
		chunk.ID,
		chunk.FileID,
		chunk.ChunkIndex,
		chunk.StartLine,
		chunk.EndLine,
		chunk.CharacterCount,
		chunk.Content,
		chunk.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create repository chunk: %w", err)
	}
	return nil
}


func (r *PostgresRepositoryChunkRepository) ReplaceByFileID(
	ctx context.Context,
	fileID string,
	chunks []*domain.RepositoryChunk,
) error {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin replace repository chunks: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, "DELETE FROM repository_chunks WHERE file_id = $1", fileID); err != nil {
		return fmt.Errorf("delete repository chunks for file %q: %w", fileID, err)
	}

	for _, chunk := range chunks {
		if _, err := tx.Exec(ctx, `
			INSERT INTO repository_chunks (
				id, file_id, chunk_index, start_line, end_line,
				character_count, content, created_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`, chunk.ID, chunk.FileID, chunk.ChunkIndex, chunk.StartLine,
			chunk.EndLine, chunk.CharacterCount, chunk.Content, chunk.CreatedAt); err != nil {
			return fmt.Errorf(
				"insert repository chunk for file %q at index %d: %w",
				fileID, chunk.ChunkIndex, err,
			)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit replace repository chunks: %w", err)
	}
	return nil
}

func (r *PostgresRepositoryChunkRepository) FindByID(
	ctx context.Context,
	id string,
) (*domain.RepositoryChunk, error) {
	chunk := &domain.RepositoryChunk{}

	err := r.db.Pool.QueryRow(
		ctx,
		`
		SELECT
			id,
			file_id,
			chunk_index,
			start_line,
			end_line,
			character_count,
			content,
			created_at
		FROM repository_chunks
		WHERE id = $1
		`,
		id,
	).Scan(
		&chunk.ID,
		&chunk.FileID,
		&chunk.ChunkIndex,
		&chunk.StartLine,
		&chunk.EndLine,
		&chunk.CharacterCount,
		&chunk.Content,
		&chunk.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRepositoryChunkNotFound
		}
		return nil, fmt.Errorf("find repository chunk: %w", err)
	}

	return chunk, nil
}

func (r *PostgresRepositoryChunkRepository) ListByFileID(
	ctx context.Context,
	fileID string,
) ([]*domain.RepositoryChunk, error) {
	rows, err := r.db.Pool.Query(
		ctx,
		`
		SELECT
			id,
			file_id,
			chunk_index,
			start_line,
			end_line,
			character_count,
			content,
			created_at
		FROM repository_chunks
		WHERE file_id = $1
		ORDER BY chunk_index ASC
		`,
		fileID,
	)
	if err != nil {
		return nil, fmt.Errorf("list repository chunks: %w", err)
	}
	defer rows.Close()

	var chunks []*domain.RepositoryChunk
	for rows.Next() {
		chunk := &domain.RepositoryChunk{}
		if err := rows.Scan(
			&chunk.ID,
			&chunk.FileID,
			&chunk.ChunkIndex,
			&chunk.StartLine,
			&chunk.EndLine,
			&chunk.CharacterCount,
			&chunk.Content,
			&chunk.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan repository chunk: %w", err)
		}
		chunks = append(chunks, chunk)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate repository chunks: %w", err)
	}

	return chunks, nil
}

var _ RepositoryChunkRepository = (*PostgresRepositoryChunkRepository)(nil)
