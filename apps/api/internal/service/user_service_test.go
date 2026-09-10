package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kristopher1027/devflow-ai/internal/domain"
	"github.com/kristopher1027/devflow-ai/internal/repository"
)

type fakeUserRepository struct {
	user *domain.User
	err  error
}

func (f *fakeUserRepository) Create(
	ctx context.Context,
	user *domain.User,
	passwordHash string,
) error {
	return nil
}

func (f *fakeUserRepository) FindByEmail(
	ctx context.Context,
	email string,
) (*domain.User, error) {
	return f.user, f.err
}

func TestUserServiceFindUserByEmail(t *testing.T) {
	expectedUser := &domain.User{
		ID:        "user-123",
		Email:     "test@example.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	repo := &fakeUserRepository{
		user: expectedUser,
	}

	service := NewUserService(repo)

	user, err := service.FindUserByEmail(
		context.Background(),
		"test@example.com",
	)
	if err != nil {
		t.Fatalf("find user: %v", err)
	}

	if user != expectedUser {
		t.Fatalf("expected user %v, got %v", expectedUser, user)
	}
}

func TestUserServiceFindUserByEmailNotFound(t *testing.T) {
	repo := &fakeUserRepository{
		err: repository.ErrUserNotFound,
	}

	service := NewUserService(repo)

	_, err := service.FindUserByEmail(
		context.Background(),
		"missing@example.com",
	)

	if !errors.Is(err, repository.ErrUserNotFound) {
		t.Fatalf(
			"expected ErrUserNotFound, got %v",
			err,
		)
	}
}
func (f *fakeUserRepository) FindCredentialsByEmail(
	ctx context.Context,
	email string,
) (*domain.UserCredentials, error) {
	return nil, nil
}
