package service

import (
	"context"
	"github.com/google/uuid"
	"github.com/kristopher1027/devflow-ai/internal/auth"
	"github.com/kristopher1027/devflow-ai/internal/domain"
	"github.com/kristopher1027/devflow-ai/internal/repository"
	"time"
)

const sessionLifetime = 7 * 24 * time.Hour

type SessionService interface {
	CreateSession(
		ctx context.Context,
		userID string,
	) (string, error)
}

type SessionServiceImpl struct {
	sessions repository.SessionRepository
}

func NewSessionService(
	sessions repository.SessionRepository,
) SessionService {
	return &SessionServiceImpl{
		sessions: sessions,
	}
}

func (s *SessionServiceImpl) CreateSession(
	ctx context.Context,
	userID string,
) (string, error) {
	token, err := auth.GenerateSessionToken()
	if err != nil {
		return "", err
	}

	tokenHash := auth.HashSessionToken(token)

	now := time.Now()

	session := &domain.Session{
		ID:         uuid.NewString(),
		UserID:     userID,
		TokenHash:  tokenHash,
		ExpiresAt:  now.Add(sessionLifetime),
		CreatedAt:  now,
		LastSeenAt: now,
	}

	if err := s.sessions.Create(ctx, session); err != nil {
		return "", err
	}

	return token, nil
}
