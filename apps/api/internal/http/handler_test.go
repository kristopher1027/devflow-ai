package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/kristopher1027/devflow-ai/internal/domain"
	"github.com/kristopher1027/devflow-ai/internal/repository"
	"github.com/kristopher1027/devflow-ai/internal/service"
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

type fakeRegistrationService struct {
	user *domain.User
	err  error
}

func (f *fakeRegistrationService) Register(
	ctx context.Context,
	email string,
	password string,
) (*domain.User, error) {
	return f.user, f.err
}

func TestRegisterUser(t *testing.T) {
	createdAt := time.Now()
	updatedAt := createdAt

	expectedUser := &domain.User{
		ID:        "user-123",
		Email:     "user@example.com",
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	service := &fakeRegistrationService{
		user: expectedUser,
	}

	handler := NewRegistrationHandler(service)

	body := `{
		"email": "user@example.com",
		"password": "correct-password"
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/register",
		strings.NewReader(body),
	)

	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	handler.Register(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusCreated,
			recorder.Code,
		)
	}

	if recorder.Header().Get("Content-Type") != "application/json" {
		t.Fatalf(
			"expected application/json content type, got %q",
			recorder.Header().Get("Content-Type"),
		)
	}

	var response domain.User

	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response.ID != expectedUser.ID {
		t.Fatalf(
			"expected ID %s, got %s",
			expectedUser.ID,
			response.ID,
		)
	}

	if response.Email != expectedUser.Email {
		t.Fatalf(
			"expected email %s, got %s",
			expectedUser.Email,
			response.Email,
		)
	}
}

func TestRegisterUserInvalidJSON(t *testing.T) {
	service := &fakeRegistrationService{}

	handler := NewRegistrationHandler(service)

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/register",
		strings.NewReader(`{"email":`),
	)

	recorder := httptest.NewRecorder()

	handler.Register(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}

	if recorder.Body.String() != "invalid request body\n" {
		t.Fatalf(
			"expected body %q, got %q",
			"invalid request body\n",
			recorder.Body.String(),
		)
	}
}

func TestRegisterUserMissingEmail(t *testing.T) {
	service := &fakeRegistrationService{
		err: service.ErrEmailRequired,
	}

	handler := NewRegistrationHandler(service)

	body := `{
		"password": "correct-password"
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/register",
		strings.NewReader(body),
	)

	recorder := httptest.NewRecorder()

	handler.Register(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}

	if recorder.Body.String() != "email is required\n" {
		t.Fatalf(
			"expected body %q, got %q",
			"email is required\n",
			recorder.Body.String(),
		)
	}
}

func TestRegisterUserMissingPassword(t *testing.T) {
	service := &fakeRegistrationService{
		err: service.ErrPasswordRequired,
	}

	handler := NewRegistrationHandler(service)

	body := `{
		"email": "user@example.com"
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/register",
		strings.NewReader(body),
	)

	recorder := httptest.NewRecorder()

	handler.Register(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}

	if recorder.Body.String() != "password is required\n" {
		t.Fatalf(
			"expected body %q, got %q",
			"password is required\n",
			recorder.Body.String(),
		)
	}
}

func TestRegisterUserDuplicateEmail(t *testing.T) {
	service := &fakeRegistrationService{
		err: repository.ErrEmailAlreadyExists,
	}

	handler := NewRegistrationHandler(service)

	body := `{
		"email": "existing@example.com",
		"password": "correct-password"
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/register",
		strings.NewReader(body),
	)

	recorder := httptest.NewRecorder()

	handler.Register(recorder, req)

	if recorder.Code != http.StatusConflict {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusConflict,
			recorder.Code,
		)
	}

	if recorder.Body.String() != "email already exists\n" {
		t.Fatalf(
			"expected body %q, got %q",
			"email already exists\n",
			recorder.Body.String(),
		)
	}
}

func TestRegisterUserInternalError(t *testing.T) {
	service := &fakeRegistrationService{
		err: errors.New("database connection failed"),
	}

	handler := NewRegistrationHandler(service)

	body := `{
		"email": "user@example.com",
		"password": "correct-password"
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/register",
		strings.NewReader(body),
	)

	recorder := httptest.NewRecorder()

	handler.Register(recorder, req)

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
