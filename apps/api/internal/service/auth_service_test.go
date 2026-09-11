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
	session          *domain.Session
	findErr          error
	createErr        error
	deleteErr        error
	deletedTokenHash string
}

func (f *fakeSessionRepository) FindByTokenHash(
	ctx context.Context,
	tokenHash string,
) (*domain.Session, error) {
	return f.session, f.findErr
}

func (f *fakeSessionRepository) Create(
	ctx context.Context,
	session *domain.Session,
) error {
	return f.createErr
}

func (f *fakeSessionRepository) DeleteByTokenHash(
	ctx context.Context,
	tokenHash string,
) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}

	f.deletedTokenHash = tokenHash

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
		findErr: repository.ErrSessionNotFound,
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

func TestAuthServiceLogout(t *testing.T) {
	repo := &fakeSessionRepository{}

	authService := NewAuthService(repo)

	tokenHash := "test-token-hash"

	err := authService.Logout(
		context.Background(),
		tokenHash,
	)
	if err != nil {
		t.Fatalf("logout: %v", err)
	}

	if repo.deletedTokenHash != tokenHash {
		t.Fatalf(
			"expected deleted token hash %s, got %s",
			tokenHash,
			repo.deletedTokenHash,
		)
	}
}

func TestAuthServiceLogoutRepositoryError(t *testing.T) {
	expectedErr := errors.New("delete session failed")

	repo := &fakeSessionRepository{
		deleteErr: expectedErr,
	}

	authService := NewAuthService(repo)

	err := authService.Logout(
		context.Background(),
		"test-token-hash",
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected repository error %v, got %v",
			expectedErr,
			err,
		)
	}
}
