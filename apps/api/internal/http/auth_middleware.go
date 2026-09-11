package http

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/kristopher1027/devflow-ai/internal/auth"
	"github.com/kristopher1027/devflow-ai/internal/repository"
)

type contextKey string

const userIDContextKey contextKey = "userID"

type AuthMiddleware struct {
	sessions repository.SessionRepository
}

func NewAuthMiddleware(
	sessions repository.SessionRepository,
) *AuthMiddleware {
	return &AuthMiddleware{
		sessions: sessions,
	}
}

func (m *AuthMiddleware) RequireAuth(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		cookie, err := r.Cookie("devflow_session")
		if err != nil {
			http.Error(
				w,
				"authentication required",
				http.StatusUnauthorized,
			)
			return
		}

		tokenHash := auth.HashSessionToken(cookie.Value)

		session, err := m.sessions.FindByTokenHash(
			r.Context(),
			tokenHash,
		)
		if err != nil {
			if errors.Is(err, repository.ErrSessionNotFound) {
				http.Error(
					w,
					"authentication required",
					http.StatusUnauthorized,
				)
				return
			}

			http.Error(
				w,
				"internal server error",
				http.StatusInternalServerError,
			)
			return
		}

		if time.Now().After(session.ExpiresAt) {
			http.Error(
				w,
				"session expired",
				http.StatusUnauthorized,
			)
			return
		}

		ctx := context.WithValue(
			r.Context(),
			userIDContextKey,
			session.UserID,
		)

		next.ServeHTTP(
			w,
			r.WithContext(ctx),
		)
	})
}
func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDContextKey).(string)

	return userID, ok
}
