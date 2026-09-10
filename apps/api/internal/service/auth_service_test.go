package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kristopher1027/devflow-ai/internal/domain"
	"github.com/kristopher1027/devflow-ai/internal/repository"
)

type fakeSessionRepository struct {
	session *domain.Session
	err     error
}

func (f *fakeSessionRepository) FindByTokenHash(
	ctx context.Context,
	tokenHash string,
) (*domain.Session, error) {
	return f.session, f.err
}

func (f *fakeSessionRepository) Create(
	ctx context.Context,
	session *domain.Session,
) error {
	return nil
}

func TestAuthServiceAuthenticateSession(t *testing.T) {
	session := &domain.Session{
		ID:         "session-123",
		UserID:     "user-123",
		TokenHash:  "token-hash",
		ExpiresAt:  time.Now().Add(1 * time.Hour),
		CreatedAt:  time.Now(),
		LastSeenAt: time.Now(),
	}

	repo := &fakeSessionRepository{
		session: session,
	}

	authService := NewAuthService(repo)

	result, err := authService.AuthenticateSession(
		context.Background(),
		"token-hash",
	)
	if err != nil {
		t.Fatalf("authenticate session: %v", err)
	}

	if result != session {
		t.Fatalf(
			"expected session %v, got %v",
			session,
			result,
		)
	}
}
func TestAuthServiceAuthenticateSessionExpired(t *testing.T) {
	session := &domain.Session{
		ID:         "session-123",
		UserID:     "user-123",
		TokenHash:  "token-hash",
		ExpiresAt:  time.Now().Add(-1 * time.Hour),
		CreatedAt:  time.Now().Add(-2 * time.Hour),
		LastSeenAt: time.Now().Add(-1 * time.Hour),
	}

	repo := &fakeSessionRepository{
		session: session,
	}

	authService := NewAuthService(repo)

	_, err := authService.AuthenticateSession(
		context.Background(),
		"token-hash",
	)

	if !errors.Is(err, ErrSessionExpired) {
		t.Fatalf(
			"expected ErrSessionExpired, got %v",
			err,
		)
	}
}
func TestAuthServiceAuthenticateSessionNotFound(t *testing.T) {
	repo := &fakeSessionRepository{
		err: repository.ErrSessionNotFound,
	}

	authService := NewAuthService(repo)

	_, err := authService.AuthenticateSession(
		context.Background(),
		"missing-token-hash",
	)

	if !errors.Is(err, repository.ErrSessionNotFound) {
		t.Fatalf(
			"expected ErrSessionNotFound, got %v",
			err,
		)
	}
}