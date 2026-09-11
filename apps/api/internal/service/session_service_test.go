package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kristopher1027/devflow-ai/internal/domain"
)

type fakeCreateSessionRepository struct {
	createdSession *domain.Session
	createErr      error
}

func (f *fakeCreateSessionRepository) DeleteByTokenHash(
	ctx context.Context,
	tokenHash string,
) error {
	return nil
}

func (f *fakeCreateSessionRepository) Create(
	ctx context.Context,
	session *domain.Session,
) error {
	f.createdSession = session
	return f.createErr
}

func (f *fakeCreateSessionRepository) FindByTokenHash(
	ctx context.Context,
	tokenHash string,
) (*domain.Session, error) {
	return nil, nil
}

func TestSessionServiceCreateSession(t *testing.T) {
	repo := &fakeCreateSessionRepository{}

	sessionService := NewSessionService(repo)

	before := time.Now()

	token, err := sessionService.CreateSession(
		context.Background(),
		"user-123",
	)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	after := time.Now()

	if token == "" {
		t.Fatal("expected session token to be non-empty")
	}

	if repo.createdSession == nil {
		t.Fatal("expected session to be created")
	}

	session := repo.createdSession

	if session.ID == "" {
		t.Fatal("expected session ID to be non-empty")
	}

	if session.UserID != "user-123" {
		t.Fatalf(
			"expected user ID user-123, got %s",
			session.UserID,
		)
	}

	if session.TokenHash == "" {
		t.Fatal("expected token hash to be non-empty")
	}

	if session.TokenHash == token {
		t.Fatal("expected stored token to be hashed")
	}

	expectedExpiry := before.Add(sessionLifetime)

	if session.ExpiresAt.Before(
		expectedExpiry.Add(-time.Second),
	) || session.ExpiresAt.After(
		after.Add(sessionLifetime).Add(time.Second),
	) {
		t.Fatalf(
			"unexpected expiration time: %v",
			session.ExpiresAt,
		)
	}

	if session.CreatedAt.Before(before) ||
		session.CreatedAt.After(after) {
		t.Fatalf(
			"unexpected created time: %v",
			session.CreatedAt,
		)
	}

	if session.LastSeenAt.Before(before) ||
		session.LastSeenAt.After(after) {
		t.Fatalf(
			"unexpected last seen time: %v",
			session.LastSeenAt,
		)
	}
}

func TestSessionServiceCreateSessionRepositoryError(t *testing.T) {
	expectedErr := errors.New("database unavailable")

	repo := &fakeCreateSessionRepository{
		createErr: expectedErr,
	}

	sessionService := NewSessionService(repo)

	token, err := sessionService.CreateSession(
		context.Background(),
		"user-123",
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected repository error %v, got %v",
			expectedErr,
			err,
		)
	}

	if token != "" {
		t.Fatalf(
			"expected empty token on failure, got %q",
			token,
		)
	}

	if repo.createdSession == nil {
		t.Fatal("expected repository Create to be called")
	}
}
