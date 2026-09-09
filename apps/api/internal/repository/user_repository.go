package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/kristopher1027/devflow-ai/internal/database"
	"github.com/kristopher1027/devflow-ai/internal/domain"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository interface {
	FindByEmail(
		ctx context.Context,
		email string,
	) (*domain.User, error)
}

type PostgresUserRepository struct {
	db *database.DB
}

func NewUserRepository(db *database.DB) UserRepository {
	return &PostgresUserRepository{
		db: db,
	}
}

func (r *PostgresUserRepository) FindByEmail(
	ctx context.Context,
	email string,
) (*domain.User, error) {
	query := `
		SELECT id, email, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user domain.User

	err := r.db.Pool.QueryRow(
		ctx,
		query,
		email,
	).Scan(
		&user.ID,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}

		return nil, fmt.Errorf("find user by email: %w", err)
	}

	return &user, nil
}
