package http

import (
	"context"
	"errors"
	"net/http"

	"github.com/kristopher1027/devflow-ai/internal/auth"
	"github.com/kristopher1027/devflow-ai/internal/service"
)

type contextKey string

const userIDContextKey contextKey = "userID"

type AuthMiddleware struct {
	authService service.AuthService
}

func NewAuthMiddleware(
	authService service.AuthService,
) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
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

		session, err := m.authService.AuthenticateSession(
			r.Context(),
			tokenHash,
		)
		if err != nil {
			if errors.Is(
				err,
				service.ErrSessionExpired,
			) {
				http.Error(
					w,
					"session expired",
					http.StatusUnauthorized,
				)
				return
			}

			http.Error(
				w,
				"authentication required",
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

func UserIDFromContext(
	ctx context.Context,
) (string, bool) {
	userID, ok := ctx.Value(userIDContextKey).(string)

	return userID, ok
}
