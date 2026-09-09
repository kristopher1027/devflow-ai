package repository

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"

	"github.com/kristopher1027/devflow-ai/internal/database"
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
