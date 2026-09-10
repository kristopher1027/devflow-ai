package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kristopher1027/devflow-ai/internal/service"
)

type fakeLoginService struct {
	token string
	err   error
}

func (f *fakeLoginService) Login(
	ctx context.Context,
	email string,
	password string,
) (string, error) {
	return f.token, f.err
}

func TestLoginUser(t *testing.T) {
	loginService := &fakeLoginService{
		token: "session-token",
	}

	handler := NewLoginHandler(
		loginService,
		false,
	)

	body := `{
		"email": "user@example.com",
		"password": "correct-password"
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/login",
		strings.NewReader(body),
	)

	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	handler.Login(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			recorder.Code,
		)
	}

	cookies := recorder.Result().Cookies()

	if len(cookies) != 1 {
		t.Fatalf(
			"expected 1 cookie, got %d",
			len(cookies),
		)
	}

	cookie := cookies[0]

	if cookie.Name != "devflow_session" {
		t.Fatalf(
			"expected cookie name devflow_session, got %s",
			cookie.Name,
		)
	}

	if cookie.Value != "session-token" {
		t.Fatalf(
			"expected session token, got %s",
			cookie.Value,
		)
	}

	if !cookie.HttpOnly {
		t.Fatal("expected cookie to be HttpOnly")
	}

	if cookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf(
			"expected SameSite=Lax, got %v",
			cookie.SameSite,
		)
	}

	if cookie.Path != "/" {
		t.Fatalf(
			"expected cookie path /, got %s",
			cookie.Path,
		)
	}

	if cookie.Secure {
		t.Fatal(
			"expected Secure=false for the HTTP test environment",
		)
	}

	if strings.Contains(recorder.Body.String(), "session-token") {
		t.Fatal(
			"session token must not be returned in response body",
		)
	}
}

func TestLoginUserInvalidJSON(t *testing.T) {
	loginService := &fakeLoginService{}

	handler := NewLoginHandler(
		loginService,
		false,
	)
	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/login",
		strings.NewReader(`{"email":`),
	)

	recorder := httptest.NewRecorder()

	handler.Login(recorder, req)

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

func TestLoginUserInvalidCredentials(t *testing.T) {
	loginService := &fakeLoginService{
		err: service.ErrInvalidCredentials,
	}

	handler := NewLoginHandler(
		loginService,
		false,
	)

	body := `{
		"email": "user@example.com",
		"password": "wrong-password"
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/login",
		strings.NewReader(body),
	)

	recorder := httptest.NewRecorder()

	handler.Login(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			recorder.Code,
		)
	}

	if recorder.Body.String() != "invalid credentials\n" {
		t.Fatalf(
			"expected body %q, got %q",
			"invalid credentials\n",
			recorder.Body.String(),
		)
	}
}

func TestLoginUserInternalError(t *testing.T) {
	loginService := &fakeLoginService{
		err: errors.New("database connection failed"),
	}

	handler := NewLoginHandler(
		loginService,
		false,
	)

	body := `{
		"email": "user@example.com",
		"password": "correct-password"
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/login",
		strings.NewReader(body),
	)

	recorder := httptest.NewRecorder()

	handler.Login(recorder, req)

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

func TestLoginUserSecureCookie(t *testing.T) {
	loginService := &fakeLoginService{
		token: "session-token",
	}

	handler := NewLoginHandler(
		loginService,
		true,
	)

	body := `{
		"email": "user@example.com",
		"password": "correct-password"
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/login",
		strings.NewReader(body),
	)

	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	handler.Login(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			recorder.Code,
		)
	}

	cookies := recorder.Result().Cookies()

	if len(cookies) != 1 {
		t.Fatalf(
			"expected 1 cookie, got %d",
			len(cookies),
		)
	}

	cookie := cookies[0]

	if !cookie.Secure {
		t.Fatal("expected cookie to be Secure")
	}

	if !cookie.HttpOnly {
		t.Fatal("expected cookie to be HttpOnly")
	}

	if cookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf(
			"expected SameSite=Lax, got %v",
			cookie.SameSite,
		)
	}
}
