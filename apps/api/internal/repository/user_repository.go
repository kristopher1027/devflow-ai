package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/kristopher1027/devflow-ai/internal/database"
	"github.com/kristopher1027/devflow-ai/internal/domain"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
)

type UserRepository interface {
	Create(
		ctx context.Context,
		user *domain.User,
		passwordHash string,
	) error

	FindByEmail(
		ctx context.Context,
		email string,
	) (*domain.User, error)

	FindCredentialsByEmail(
		ctx context.Context,
		email string,
	) (*domain.UserCredentials, error)
}

type PostgresUserRepository struct {
	db *database.DB
}

func NewUserRepository(db *database.DB) UserRepository {
	return &PostgresUserRepository{
		db: db,
	}
}

func (r *PostgresUserRepository) Create(
	ctx context.Context,
	user *domain.User,
	passwordHash string,
) error {
	query := `
		INSERT INTO users (
			id,
			email,
			password_hash,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.Pool.Exec(
		ctx,
		query,
		user.ID,
		user.Email,
		passwordHash,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) &&
			pgErr.Code == "23505" &&
			pgErr.ConstraintName == "users_email_key" {
			return ErrEmailAlreadyExists
		}

		return fmt.Errorf("create user: %w", err)
	}

	return nil
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

func (r *PostgresUserRepository) FindCredentialsByEmail(
	ctx context.Context,
	email string,
) (*domain.UserCredentials, error) {
	query := `
		SELECT id, password_hash
		FROM users
		WHERE email = $1
	`

	var credentials domain.UserCredentials

	err := r.db.Pool.QueryRow(
		ctx,
		query,
		email,
	).Scan(
		&credentials.UserID,
		&credentials.PasswordHash,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}

		return nil, fmt.Errorf(
			"find user credentials by email: %w",
			err,
		)
	}

	return &credentials, nil
}
