package http

import (
	"context"
	"net/http"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kristopher1027/devflow-ai/internal/domain"
	"github.com/kristopher1027/devflow-ai/internal/repository"
)

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodGet,
		"/health",
		nil,
	)

	recorder := httptest.NewRecorder()

	HealthHandler(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			recorder.Code,
		)
	}

	expected := `{"service":"devflow-api","status":"ok"}`

	if recorder.Body.String() != expected+"\n" {
		t.Fatalf(
			"expected body %q, got %q",
			expected+"\n",
			recorder.Body.String(),
		)
	}
}

type fakeUserService struct {
	user *domain.User
	err  error
}

func (f *fakeUserService) FindUserByEmail(
	ctx context.Context,
	email string,
) (*domain.User, error) {
	return f.user, f.err
}
func TestGetUser(t *testing.T) {
	expectedUser := &domain.User{
		ID:        "user-123",
		Email:     "test@example.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	userService := &fakeUserService{
		user: expectedUser,
	}

	handler := NewUserHandler(userService)

	req := httptest.NewRequest(
		http.MethodGet,
		"/users?email=test@example.com",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.GetUser(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			recorder.Code,
		)
	}

	if recorder.Header().Get("Content-Type") != "application/json" {
		t.Fatalf(
			"expected application/json content type, got %q",
			recorder.Header().Get("Content-Type"),
		)
	}
}

func TestGetUserMissingEmail(t *testing.T) {
	userService := &fakeUserService{}

	handler := NewUserHandler(userService)

	req := httptest.NewRequest(
		http.MethodGet,
		"/users",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.GetUser(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}
func TestGetUserNotFound(t *testing.T) {
	userService := &fakeUserService{
		err: repository.ErrUserNotFound,
	}

	handler := NewUserHandler(userService)

	req := httptest.NewRequest(
		http.MethodGet,
		"/users?email=missing@example.com",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.GetUser(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNotFound,
			recorder.Code,
		)
	}

	if recorder.Body.String() != "user not found\n" {
		t.Fatalf(
			"expected body %q, got %q",
			"user not found\n",
			recorder.Body.String(),
		)
	}
}
func TestGetUserInternalError(t *testing.T) {
	userService := &fakeUserService{
		err: errors.New("database connection failed"),
	}

	handler := NewUserHandler(userService)

	req := httptest.NewRequest(
		http.MethodGet,
		"/users?email=test@example.com",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.GetUser(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusInternalServerError,
			recorder.Code,
		)
	}

	if recorder.Body.String() != "internal server error\n" {
		t.Fatalf(
			"expected body %q, got %q",
			"internal server error\n",
			recorder.Body.String(),
		)
	}
}
