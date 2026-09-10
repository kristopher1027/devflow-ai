package service

import (
	"context"
	"errors"
	"testing"

	"github.com/kristopher1027/devflow-ai/internal/auth"
	"github.com/kristopher1027/devflow-ai/internal/domain"
	"github.com/kristopher1027/devflow-ai/internal/repository"
)

type fakeLoginUserRepository struct {
	credentials *domain.UserCredentials
	err         error
}

func (f *fakeLoginUserRepository) Create(
	ctx context.Context,
	user *domain.User,
	passwordHash string,
) error {
	return nil
}

func (f *fakeLoginUserRepository) FindByEmail(
	ctx context.Context,
	email string,
) (*domain.User, error) {
	return nil, repository.ErrUserNotFound
}

func (f *fakeLoginUserRepository) FindCredentialsByEmail(
	ctx context.Context,
	email string,
) (*domain.UserCredentials, error) {
	return f.credentials, f.err
}

type fakeLoginSessionService struct {
	token  string
	err    error
	userID string
}

func (f *fakeLoginSessionService) CreateSession(
	ctx context.Context,
	userID string,
) (string, error) {
	f.userID = userID

	if f.err != nil {
		return "", f.err
	}

	return f.token, nil
}

func TestLoginServiceLogin(t *testing.T) {
	password := "correct-password"

	passwordHash, err := hashTestPassword(password)
	if err != nil {
		t.Fatalf("hash test password: %v", err)
	}

	repo := &fakeLoginUserRepository{
		credentials: &domain.UserCredentials{
			UserID:       "user-123",
			PasswordHash: passwordHash,
		},
	}

	sessions := &fakeLoginSessionService{
		token: "session-token",
	}

	service := NewLoginService(repo, sessions)

	token, err := service.Login(
		context.Background(),
		"user@example.com",
		password,
	)

	if err != nil {
		t.Fatalf("login: %v", err)
	}

	if token != "session-token" {
		t.Fatalf(
			"expected session-token, got %s",
			token,
		)
	}

	if sessions.userID != "user-123" {
		t.Fatalf(
			"expected user ID user-123, got %s",
			sessions.userID,
		)
	}
}

func TestLoginServiceWrongPassword(t *testing.T) {
	passwordHash, err := hashTestPassword("correct-password")
	if err != nil {
		t.Fatalf("hash test password: %v", err)
	}

	repo := &fakeLoginUserRepository{
		credentials: &domain.UserCredentials{
			UserID:       "user-123",
			PasswordHash: passwordHash,
		},
	}

	sessions := &fakeLoginSessionService{
		token: "session-token",
	}

	service := NewLoginService(repo, sessions)

	_, err = service.Login(
		context.Background(),
		"user@example.com",
		"wrong-password",
	)

	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf(
			"expected ErrInvalidCredentials, got %v",
			err,
		)
	}
}

func TestLoginServiceUserNotFound(t *testing.T) {
	repo := &fakeLoginUserRepository{
		err: repository.ErrUserNotFound,
	}

	sessions := &fakeLoginSessionService{
		token: "session-token",
	}

	service := NewLoginService(repo, sessions)

	_, err := service.Login(
		context.Background(),
		"missing@example.com",
		"correct-password",
	)

	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf(
			"expected ErrInvalidCredentials, got %v",
			err,
		)
	}
}

func TestLoginServiceEmptyEmail(t *testing.T) {
	repo := &fakeLoginUserRepository{}

	sessions := &fakeLoginSessionService{}

	service := NewLoginService(repo, sessions)

	_, err := service.Login(
		context.Background(),
		"",
		"correct-password",
	)

	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf(
			"expected ErrInvalidCredentials, got %v",
			err,
		)
	}
}

func TestLoginServiceEmptyPassword(t *testing.T) {
	repo := &fakeLoginUserRepository{}

	sessions := &fakeLoginSessionService{}

	service := NewLoginService(repo, sessions)

	_, err := service.Login(
		context.Background(),
		"user@example.com",
		"",
	)

	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf(
			"expected ErrInvalidCredentials, got %v",
			err,
		)
	}
}

func TestLoginServiceMissingPasswordHash(t *testing.T) {
	repo := &fakeLoginUserRepository{
		credentials: &domain.UserCredentials{
			UserID:       "user-123",
			PasswordHash: "",
		},
	}

	sessions := &fakeLoginSessionService{
		token: "session-token",
	}

	service := NewLoginService(repo, sessions)

	_, err := service.Login(
		context.Background(),
		"user@example.com",
		"correct-password",
	)

	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf(
			"expected ErrInvalidCredentials, got %v",
			err,
		)
	}
}

func TestLoginServiceSessionCreationError(t *testing.T) {
	passwordHash, err := hashTestPassword("correct-password")
	if err != nil {
		t.Fatalf("hash test password: %v", err)
	}

	expectedErr := errors.New("session creation failed")

	repo := &fakeLoginUserRepository{
		credentials: &domain.UserCredentials{
			UserID:       "user-123",
			PasswordHash: passwordHash,
		},
	}

	sessions := &fakeLoginSessionService{
		err: expectedErr,
	}

	service := NewLoginService(repo, sessions)

	_, err = service.Login(
		context.Background(),
		"user@example.com",
		"correct-password",
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected session error, got %v",
			err,
		)
	}
}

func hashTestPassword(password string) (string, error) {
	return auth.HashPassword(password)
}
