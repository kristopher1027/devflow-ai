
package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/kristopher1027/devflow-ai/internal/database"
	"github.com/kristopher1027/devflow-ai/internal/domain"
)

var ErrSessionNotFound = errors.New("session not found")

type SessionRepository interface {
	Create(
		ctx context.Context,
		session *domain.Session,
	) error

	FindByTokenHash(
		ctx context.Context,
		tokenHash string,
	) (*domain.Session, error)

	DeleteByTokenHash(
		ctx context.Context,
		tokenHash string,
	) error
}

type PostgresSessionRepository struct {
	db *database.DB
}

func NewSessionRepository(db *database.DB) SessionRepository {
	return &PostgresSessionRepository{
		db: db,
	}
}

func (r *PostgresSessionRepository) Create(
	ctx context.Context,
	session *domain.Session,
) error {
	query := `
		INSERT INTO sessions (
			id,
			user_id,
			token_hash,
			expires_at,
			created_at,
			last_seen_at
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.Pool.Exec(
		ctx,
		query,
		session.ID,
		session.UserID,
		session.TokenHash,
		session.ExpiresAt,
		session.CreatedAt,
		session.LastSeenAt,
	)

	return err
}

func (r *PostgresSessionRepository) FindByTokenHash(
	ctx context.Context,
	tokenHash string,
) (*domain.Session, error) {
	query := `
		SELECT
			id,
			user_id,
			token_hash,
			expires_at,
			created_at,
			last_seen_at
		FROM sessions
		WHERE token_hash = $1
	`

	var session domain.Session

	err := r.db.Pool.QueryRow(
		ctx,
		query,
		tokenHash,
	).Scan(
		&session.ID,
		&session.UserID,
		&session.TokenHash,
		&session.ExpiresAt,
		&session.CreatedAt,
		&session.LastSeenAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSessionNotFound
		}

		return nil, err
	}

	return &session, nil
}

func (r *PostgresSessionRepository) DeleteByTokenHash(
	ctx context.Context,
	tokenHash string,
) error {
	query := `
		DELETE FROM sessions
		WHERE token_hash = $1
	`

	_, err := r.db.Pool.Exec(
		ctx,
		query,
		tokenHash,
	)

	return err
}
