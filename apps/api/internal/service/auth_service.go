package service

import (
	"context"
	"errors"
	"time"

	"github.com/kristopher1027/devflow-ai/internal/domain"
	"github.com/kristopher1027/devflow-ai/internal/repository"
)

var ErrSessionExpired = errors.New("session expired")

type AuthService interface {
	AuthenticateSession(
		ctx context.Context,
		tokenHash string,
	) (*domain.Session, error)

	Logout(
		ctx context.Context,
		tokenHash string,
	) error
}

type AuthServiceImpl struct {
	sessions repository.SessionRepository
}

func NewAuthService(
	sessions repository.SessionRepository,
) AuthService {
	return &AuthServiceImpl{
		sessions: sessions,
	}
}

func (s *AuthServiceImpl) AuthenticateSession(
	ctx context.Context,
	tokenHash string,
) (*domain.Session, error) {
	session, err := s.sessions.FindByTokenHash(
		ctx,
		tokenHash,
	)
	if err != nil {
		return nil, err
	}

	if session.ExpiresAt.Before(time.Now()) {
		return nil, ErrSessionExpired
	}

	return session, nil
}

func (s *AuthServiceImpl) Logout(
	ctx context.Context,
	tokenHash string,
) error {
	return s.sessions.DeleteByTokenHash(
		ctx,
		tokenHash,
	)
}
