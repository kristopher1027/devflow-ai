package service

import (
	"context"
	"errors"

	"github.com/kristopher1027/devflow-ai/internal/auth"
	"github.com/kristopher1027/devflow-ai/internal/repository"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type LoginService interface {
	Login(
		ctx context.Context,
		email string,
		password string,
	) (string, error)
}

type LoginServiceImpl struct {
	users    repository.UserRepository
	sessions SessionService
}

func NewLoginService(
	users repository.UserRepository,
	sessions SessionService,
) LoginService {
	return &LoginServiceImpl{
		users:    users,
		sessions: sessions,
	}
}

func (s *LoginServiceImpl) Login(
	ctx context.Context,
	email string,
	password string,
) (string, error) {
	if email == "" || password == "" {
		return "", ErrInvalidCredentials
	}

	credentials, err := s.users.FindCredentialsByEmail(
		ctx,
		email,
	)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return "", ErrInvalidCredentials
		}

		return "", err
	}

	if credentials.PasswordHash == "" {
		return "", ErrInvalidCredentials
	}

	if err := auth.VerifyPassword(
		password,
		credentials.PasswordHash,
	); err != nil {
		return "", ErrInvalidCredentials
	}

	token, err := s.sessions.CreateSession(
		ctx,
		credentials.UserID,
	)
	if err != nil {
		return "", err
	}

	return token, nil
}
