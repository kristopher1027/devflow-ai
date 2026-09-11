package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kristopher1027/devflow-ai/internal/domain"
	"github.com/kristopher1027/devflow-ai/internal/service"
)

type fakeAuthService struct {
	session *domain.Session
	err     error
}

func (f *fakeAuthService) AuthenticateSession(
	ctx context.Context,
	tokenHash string,
) (*domain.Session, error) {
	return f.session, f.err
}

func (f *fakeAuthService) Logout(
	ctx context.Context,
	tokenHash string,
) error {
	return nil
}

func TestAuthMiddlewareNoCookie(t *testing.T) {
	authService := &fakeAuthService{}

	middleware := NewAuthMiddleware(authService)

	next := http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		t.Fatal("next handler should not be called")
	})

	handler := middleware.RequireAuth(next)

	req := httptest.NewRequest(
		http.MethodGet,
		"/protected",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			rec.Code,
		)
	}

	if rec.Body.String() != "authentication required\n" {
		t.Fatalf(
			"expected body %q, got %q",
			"authentication required\n",
			rec.Body.String(),
		)
	}
}

func TestAuthMiddlewareInvalidSession(t *testing.T) {
	authService := &fakeAuthService{
		err: errors.New("session not found"),
	}

	middleware := NewAuthMiddleware(authService)

	next := http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		t.Fatal("next handler should not be called")
	})

	handler := middleware.RequireAuth(next)

	req := httptest.NewRequest(
		http.MethodGet,
		"/protected",
		nil,
	)

	req.AddCookie(&http.Cookie{
		Name:  "devflow_session",
		Value: "invalid-token",
	})

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			rec.Code,
		)
	}
}

func TestAuthMiddlewareRepositoryError(t *testing.T) {
	authService := &fakeAuthService{
		err: errors.New("database connection failed"),
	}

	middleware := NewAuthMiddleware(authService)

	next := http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		t.Fatal("next handler should not be called")
	})

	handler := middleware.RequireAuth(next)

	req := httptest.NewRequest(
		http.MethodGet,
		"/protected",
		nil,
	)

	req.AddCookie(&http.Cookie{
		Name:  "devflow_session",
		Value: "valid-looking-token",
	})

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			rec.Code,
		)
	}

	if rec.Body.String() != "authentication required\n" {
		t.Fatalf(
			"expected body %q, got %q",
			"authentication required\n",
			rec.Body.String(),
		)
	}
}

func TestAuthMiddlewareExpiredSession(t *testing.T) {
	authService := &fakeAuthService{
		err: service.ErrSessionExpired,
	}

	middleware := NewAuthMiddleware(authService)

	next := http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		t.Fatal("next handler should not be called")
	})

	handler := middleware.RequireAuth(next)

	req := httptest.NewRequest(
		http.MethodGet,
		"/protected",
		nil,
	)

	req.AddCookie(&http.Cookie{
		Name:  "devflow_session",
		Value: "expired-token",
	})

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			rec.Code,
		)
	}

	if rec.Body.String() != "session expired\n" {
		t.Fatalf(
			"expected body %q, got %q",
			"session expired\n",
			rec.Body.String(),
		)
	}
}

func TestAuthMiddlewareValidSession(t *testing.T) {
	authService := &fakeAuthService{
		session: &domain.Session{
			ID:        "session-123",
			UserID:    "user-123",
			ExpiresAt: time.Now().Add(time.Hour),
		},
	}

	middleware := NewAuthMiddleware(authService)

	next := http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		userID, ok := UserIDFromContext(r.Context())
		if !ok {
			t.Fatal("expected user ID in request context")
		}

		if userID != "user-123" {
			t.Fatalf(
				"expected user ID %q, got %q",
				"user-123",
				userID,
			)
		}

		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.RequireAuth(next)

	req := httptest.NewRequest(
		http.MethodGet,
		"/protected",
		nil,
	)

	req.AddCookie(&http.Cookie{
		Name:  "devflow_session",
		Value: "valid-token",
	})

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}
}
