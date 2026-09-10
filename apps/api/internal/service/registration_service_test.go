package service

import (
	"context"
	"errors"
	"testing"

	"github.com/kristopher1027/devflow-ai/internal/domain"
	"github.com/kristopher1027/devflow-ai/internal/repository"
)

type fakeRegistrationUserRepository struct {
	createdUser         *domain.User
	createdPasswordHash string
	createErr           error
}

func (f *fakeRegistrationUserRepository) Create(
	ctx context.Context,
	user *domain.User,
	passwordHash string,
) error {
	f.createdUser = user
	f.createdPasswordHash = passwordHash

	return f.createErr
}

func (f *fakeRegistrationUserRepository) FindByEmail(
	ctx context.Context,
	email string,
) (*domain.User, error) {
	return nil, repository.ErrUserNotFound
}

func (f *fakeRegistrationUserRepository) FindCredentialsByEmail(
	ctx context.Context,
	email string,
) (*domain.UserCredentials, error) {
	return nil, repository.ErrUserNotFound
}

func TestRegistrationServiceRegister(t *testing.T) {
	repo := &fakeRegistrationUserRepository{}

	service := NewRegistrationService(repo)

	user, err := service.Register(
		context.Background(),
		"user@example.com",
		"correct-password",
	)

	if err != nil {
		t.Fatalf("register user: %v", err)
	}

	if user == nil {
		t.Fatal("expected user, got nil")
	}

	if user.Email != "user@example.com" {
		t.Fatalf(
			"expected email user@example.com, got %s",
			user.Email,
		)
	}

	if user.ID == "" {
		t.Fatal("expected user ID")
	}

	if repo.createdUser == nil {
		t.Fatal("expected user to be created")
	}

	if repo.createdPasswordHash == "" {
		t.Fatal("expected password hash to be created")
	}

	if repo.createdPasswordHash == "correct-password" {
		t.Fatal("password must not be stored as plaintext")
	}
}

func TestRegistrationServiceRegisterEmptyEmail(t *testing.T) {
	repo := &fakeRegistrationUserRepository{}

	service := NewRegistrationService(repo)

	_, err := service.Register(
		context.Background(),
		"",
		"correct-password",
	)

	if err == nil {
		t.Fatal("expected error for empty email")
	}

	if repo.createdUser != nil {
		t.Fatal("user should not be created")
	}
}

func TestRegistrationServiceRegisterEmptyPassword(t *testing.T) {
	repo := &fakeRegistrationUserRepository{}

	service := NewRegistrationService(repo)

	_, err := service.Register(
		context.Background(),
		"user@example.com",
		"",
	)

	if err == nil {
		t.Fatal("expected error for empty password")
	}

	if repo.createdUser != nil {
		t.Fatal("user should not be created")
	}
}

func TestRegistrationServiceRegisterRepositoryError(t *testing.T) {
	expectedErr := errors.New("database error")

	repo := &fakeRegistrationUserRepository{
		createErr: expectedErr,
	}

	service := NewRegistrationService(repo)

	_, err := service.Register(
		context.Background(),
		"user@example.com",
		"correct-password",
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected repository error, got %v",
			err,
		)
	}
}
func TestRegistrationServiceRegisterDuplicateEmail(t *testing.T) {
	repo := &fakeRegistrationUserRepository{
		createErr: repository.ErrEmailAlreadyExists,
	}

	service := NewRegistrationService(repo)

	_, err := service.Register(
		context.Background(),
		"user@example.com",
		"correct-password",
	)

	if !errors.Is(err, repository.ErrEmailAlreadyExists) {
		t.Fatalf(
			"expected ErrEmailAlreadyExists, got %v",
			err,
		)
	}
}
