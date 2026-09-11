package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kristopher1027/devflow-ai/internal/domain"
	"github.com/kristopher1027/devflow-ai/internal/repository"
)

type fakeAuthSessionRepository struct {
	session *domain.Session
	err     error
}

func (f *fakeAuthSessionRepository) Create(
	ctx context.Context,
	session *domain.Session,
) error {
	return nil
}

func (f *fakeAuthSessionRepository) FindByTokenHash(
	ctx context.Context,
	tokenHash string,
) (*domain.Session, error) {
	return f.session, f.err
}

func TestAuthMiddlewareNoCookie(t *testing.T) {
	repo := &fakeAuthSessionRepository{}

	middleware := NewAuthMiddleware(repo)

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
	repo := &fakeAuthSessionRepository{
		err: repository.ErrSessionNotFound,
	}

	middleware := NewAuthMiddleware(repo)

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
	repo := &fakeAuthSessionRepository{
		err: errors.New("database connection failed"),
	}

	middleware := NewAuthMiddleware(repo)

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

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusInternalServerError,
			rec.Code,
		)
	}

	if rec.Body.String() != "internal server error\n" {
		t.Fatalf(
			"expected body %q, got %q",
			"internal server error\n",
			rec.Body.String(),
		)
	}
}

func TestAuthMiddlewareExpiredSession(t *testing.T) {
	repo := &fakeAuthSessionRepository{
		session: &domain.Session{
			ID:        "session-123",
			UserID:    "user-123",
			ExpiresAt: time.Now().Add(-time.Hour),
		},
	}

	middleware := NewAuthMiddleware(repo)

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
	repo := &fakeAuthSessionRepository{
		session: &domain.Session{
			ID:        "session-123",
			UserID:    "user-123",
			ExpiresAt: time.Now().Add(time.Hour),
		},
	}

	middleware := NewAuthMiddleware(repo)

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
