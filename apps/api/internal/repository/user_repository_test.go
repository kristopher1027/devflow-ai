package repository

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/kristopher1027/devflow-ai/internal/auth"
	"github.com/kristopher1027/devflow-ai/internal/database"
	"github.com/kristopher1027/devflow-ai/internal/domain"
)

func TestUserRepositoryFindByEmail(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")

	if databaseURL == "" {
		t.Fatal("DATABASE_URL is required")
	}

	ctx := context.Background()

	db, err := database.Connect(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect to database: %v", err)
	}

	repo := NewUserRepository(db)

	email := "test-" + uuid.NewString() + "@example.com"
	userID := uuid.New()

	_, err = db.Pool.Exec(
		ctx,
		`
		INSERT INTO users (id, email)
		VALUES ($1, $2)
		`,
		userID,
		email,
	)
	if err != nil {
		db.Close()
		t.Fatalf("insert test user: %v", err)
	}

	t.Cleanup(func() {
		_, err := db.Pool.Exec(
			context.Background(),
			"DELETE FROM users WHERE id = $1",
			userID,
		)

		if err != nil {
			t.Errorf("cleanup test user: %v", err)
		}

		db.Close()
	})

	user, err := repo.FindByEmail(ctx, email)
	if err != nil {
		t.Fatalf("find user: %v", err)
	}

	if user.ID != userID.String() {
		t.Fatalf(
			"expected user ID %s, got %s",
			userID,
			user.ID,
		)
	}

	if user.Email != email {
		t.Fatalf(
			"expected email %s, got %s",
			email,
			user.Email,
		)
	}
}
func TestUserRepositoryFindByEmailNotFound(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")

	if databaseURL == "" {
		t.Fatal("DATABASE_URL is required")
	}

	ctx := context.Background()

	db, err := database.Connect(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect to database: %v", err)
	}

	repo := NewUserRepository(db)

	t.Cleanup(func() {
		db.Close()
	})

	_, err = repo.FindByEmail(
		ctx,
		"does-not-exist@example.com",
	)

	if err != ErrUserNotFound {
		t.Fatalf(
			"expected ErrUserNotFound, got %v",
			err,
		)
	}
}
func TestUserRepositoryFindCredentialsByEmail(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Fatal("DATABASE_URL is required")
	}

	ctx := context.Background()

	db, err := database.Connect(
		ctx,
		os.Getenv("DATABASE_URL"),
	)
	if err != nil {
		t.Fatalf("connect database: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	repo := NewUserRepository(db)

	userID := uuid.NewString()
	email := fmt.Sprintf(
		"test-%s@example.com",
		uuid.NewString(),
	)

	passwordHash, err := auth.HashPassword(
		"correct-password",
	)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	_, err = db.Pool.Exec(
		ctx,
		`
		INSERT INTO users (
			id,
			email,
			password_hash
		)
		VALUES ($1, $2, $3)
		`,
		userID,
		email,
		passwordHash,
	)
	if err != nil {
		t.Fatalf("insert test user: %v", err)
	}

	t.Cleanup(func() {
		_, err := db.Pool.Exec(
			context.Background(),
			`DELETE FROM users WHERE id = $1`,
			userID,
		)
		if err != nil {
			t.Errorf("cleanup test user: %v", err)
		}
	})

	credentials, err := repo.FindCredentialsByEmail(
		ctx,
		email,
	)
	if err != nil {
		t.Fatalf(
			"find credentials by email: %v",
			err,
		)
	}

	if credentials.UserID != userID {
		t.Fatalf(
			"expected user ID %s, got %s",
			userID,
			credentials.UserID,
		)
	}

	if credentials.PasswordHash != passwordHash {
		t.Fatal("expected password hash to match")
	}
}

func TestUserRepositoryCreate(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Fatal("DATABASE_URL is required")
	}

	ctx := context.Background()

	db, err := database.Connect(
		ctx,
		os.Getenv("DATABASE_URL"),
	)
	if err != nil {
		t.Fatalf("connect database: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	repo := NewUserRepository(db)

	userID := uuid.NewString()
	email := fmt.Sprintf(
		"test-%s@example.com",
		uuid.NewString(),
	)

	passwordHash, err := auth.HashPassword(
		"correct-password",
	)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	now := time.Now()

	user := &domain.User{
		ID:        userID,
		Email:     email,
		CreatedAt: now,
		UpdatedAt: now,
	}

	t.Cleanup(func() {
		_, err := db.Pool.Exec(
			context.Background(),
			`DELETE FROM users WHERE id = $1`,
			userID,
		)
		if err != nil {
			t.Errorf("cleanup test user: %v", err)
		}
	})

	err = repo.Create(
		ctx,
		user,
		passwordHash,
	)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	var (
		storedEmail        string
		storedPasswordHash string
	)

	err = db.Pool.QueryRow(
		ctx,
		`
		SELECT email, password_hash
		FROM users
		WHERE id = $1
		`,
		userID,
	).Scan(
		&storedEmail,
		&storedPasswordHash,
	)
	if err != nil {
		t.Fatalf("query created user: %v", err)
	}

	if storedEmail != email {
		t.Fatalf(
			"expected email %s, got %s",
			email,
			storedEmail,
		)
	}

	if storedPasswordHash != passwordHash {
		t.Fatal("expected password hash to match")
	}
}
