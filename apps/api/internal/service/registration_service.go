package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/kristopher1027/devflow-ai/internal/auth"
	"github.com/kristopher1027/devflow-ai/internal/domain"
	"github.com/kristopher1027/devflow-ai/internal/repository"
)

var (
	ErrEmailRequired    = errors.New("email is required")
	ErrPasswordRequired = errors.New("password is required")
)

type RegistrationService interface {
	Register(
		ctx context.Context,
		email string,
		password string,
	) (*domain.User, error)
}

type RegistrationServiceImpl struct {
	users repository.UserRepository
}

func NewRegistrationService(
	users repository.UserRepository,
) RegistrationService {
	return &RegistrationServiceImpl{
		users: users,
	}
}

func (s *RegistrationServiceImpl) Register(
	ctx context.Context,
	email string,
	password string,
) (*domain.User, error) {
	if email == "" {
		return nil, ErrEmailRequired
	}

	if password == "" {
		return nil, ErrPasswordRequired
	}

	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	user := &domain.User{
		ID:        uuid.NewString(),
		Email:     email,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.users.Create(
		ctx,
		user,
		passwordHash,
	); err != nil {
		return nil, err
	}

	return user, nil
}
