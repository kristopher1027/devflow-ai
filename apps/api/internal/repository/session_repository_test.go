package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/kristopher1027/devflow-ai/internal/database"
	"github.com/kristopher1027/devflow-ai/internal/domain"
)

func TestSessionRepositoryCreate(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")

	if databaseURL == "" {
		t.Fatal("DATABASE_URL is required")
	}

	ctx := context.Background()

	db, err := database.Connect(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect to database: %v", err)
	}

	repo := NewSessionRepository(db)

	userID := uuid.New()
	sessionID := uuid.New()

	email := "session-test-" + uuid.NewString() + "@example.com"

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

	now := time.Now()

	session := &domain.Session{
		ID:         sessionID.String(),
		UserID:     userID.String(),
		TokenHash:  "test-token-hash-" + uuid.NewString(),
		ExpiresAt:  now.Add(24 * time.Hour),
		CreatedAt:  now,
		LastSeenAt: now,
	}

	err = repo.Create(ctx, session)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	var storedSession domain.Session

	err = db.Pool.QueryRow(
		ctx,
		`
		SELECT
			id,
			user_id,
			token_hash,
			expires_at,
			created_at,
			last_seen_at
		FROM sessions
		WHERE id = $1
		`,
		sessionID,
	).Scan(
		&storedSession.ID,
		&storedSession.UserID,
		&storedSession.TokenHash,
		&storedSession.ExpiresAt,
		&storedSession.CreatedAt,
		&storedSession.LastSeenAt,
	)

	if err != nil {
		t.Fatalf("query created session: %v", err)
	}

	if storedSession.ID != session.ID {
		t.Fatalf(
			"expected session ID %s, got %s",
			session.ID,
			storedSession.ID,
		)
	}

	if storedSession.UserID != session.UserID {
		t.Fatalf(
			"expected user ID %s, got %s",
			session.UserID,
			storedSession.UserID,
		)
	}

	if storedSession.TokenHash != session.TokenHash {
		t.Fatalf(
			"expected token hash %s, got %s",
			session.TokenHash,
			storedSession.TokenHash,
		)
	}
}

func TestSessionRepositoryFindByTokenHash(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")

	if databaseURL == "" {
		t.Fatal("DATABASE_URL is required")
	}

	ctx := context.Background()

	db, err := database.Connect(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect to database: %v", err)
	}

	repo := NewSessionRepository(db)

	userID := uuid.New()
	sessionID := uuid.New()

	email := "session-find-" + uuid.NewString() + "@example.com"
	tokenHash := "test-token-hash-" + uuid.NewString()

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

	now := time.Now()

	session := &domain.Session{
		ID:         sessionID.String(),
		UserID:     userID.String(),
		TokenHash:  tokenHash,
		ExpiresAt:  now.Add(24 * time.Hour),
		CreatedAt:  now,
		LastSeenAt: now,
	}

	err = repo.Create(ctx, session)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	foundSession, err := repo.FindByTokenHash(
		ctx,
		tokenHash,
	)
	if err != nil {
		t.Fatalf("find session: %v", err)
	}

	if foundSession.ID != session.ID {
		t.Fatalf(
			"expected session ID %s, got %s",
			session.ID,
			foundSession.ID,
		)
	}

	if foundSession.UserID != session.UserID {
		t.Fatalf(
			"expected user ID %s, got %s",
			session.UserID,
			foundSession.UserID,
		)
	}

	if foundSession.TokenHash != session.TokenHash {
		t.Fatalf(
			"expected token hash %s, got %s",
			session.TokenHash,
			foundSession.TokenHash,
		)
	}
}

func TestSessionRepositoryFindByTokenHashNotFound(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")

	if databaseURL == "" {
		t.Fatal("DATABASE_URL is required")
	}

	ctx := context.Background()

	db, err := database.Connect(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect to database: %v", err)
	}

	repo := NewSessionRepository(db)

	t.Cleanup(func() {
		db.Close()
	})

	_, err = repo.FindByTokenHash(
		ctx,
		"does-not-exist-"+uuid.NewString(),
	)

	if err != ErrSessionNotFound {
		t.Fatalf(
			"expected ErrSessionNotFound, got %v",
			err,
		)
	}
}

func TestSessionRepositoryDeleteByTokenHash(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")

	if databaseURL == "" {
		t.Fatal("DATABASE_URL is required")
	}

	ctx := context.Background()

	db, err := database.Connect(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect to database: %v", err)
	}

	repo := NewSessionRepository(db)

	userID := uuid.New()
	sessionID := uuid.New()
	tokenHash := "delete-test-token-" + uuid.NewString()
	email := "session-delete-" + uuid.NewString() + "@example.com"

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

	now := time.Now()

	session := &domain.Session{
		ID:         sessionID.String(),
		UserID:     userID.String(),
		TokenHash:  tokenHash,
		ExpiresAt:  now.Add(24 * time.Hour),
		CreatedAt:  now,
		LastSeenAt: now,
	}

	err = repo.Create(ctx, session)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	err = repo.DeleteByTokenHash(ctx, tokenHash)
	if err != nil {
		t.Fatalf("delete session: %v", err)
	}

	_, err = repo.FindByTokenHash(ctx, tokenHash)

	if err != ErrSessionNotFound {
		t.Fatalf(
			"expected ErrSessionNotFound after deletion, got %v",
			err,
		)
	}
}
