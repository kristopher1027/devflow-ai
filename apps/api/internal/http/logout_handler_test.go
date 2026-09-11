package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kristopher1027/devflow-ai/internal/auth"
	"github.com/kristopher1027/devflow-ai/internal/domain"
)

type fakeLogoutService struct {
	tokenHash string
	err       error
}

func (f *fakeLogoutService) AuthenticateSession(
	ctx context.Context,
	tokenHash string,
) (*domain.Session, error) {
	return nil, nil
}

func (f *fakeLogoutService) Logout(
	ctx context.Context,
	tokenHash string,
) error {
	f.tokenHash = tokenHash

	return f.err
}

func TestLogoutUser(t *testing.T) {
	service := &fakeLogoutService{}

	handler := NewLogoutHandler(
		service,
		false,
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/logout",
		nil,
	)

	req.AddCookie(&http.Cookie{
		Name:  "devflow_session",
		Value: "session-token",
	})

	recorder := httptest.NewRecorder()

	handler.Logout(recorder, req)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNoContent,
			recorder.Code,
		)
	}

	expectedHash := auth.HashSessionToken("session-token")

	if service.tokenHash != expectedHash {
		t.Fatalf(
			"expected token hash %q, got %q",
			expectedHash,
			service.tokenHash,
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

	if cookie.Value != "" {
		t.Fatalf(
			"expected empty cookie value, got %q",
			cookie.Value,
		)
	}

	if cookie.MaxAge >= 0 {
		t.Fatalf(
			"expected cookie MaxAge < 0, got %d",
			cookie.MaxAge,
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
}

func TestLogoutUserWithoutCookie(t *testing.T) {
	service := &fakeLogoutService{}

	handler := NewLogoutHandler(
		service,
		false,
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/logout",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.Logout(recorder, req)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNoContent,
			recorder.Code,
		)
	}

	if service.tokenHash != "" {
		t.Fatalf(
			"expected service not to be called, got token hash %q",
			service.tokenHash,
		)
	}
}

func TestLogoutUserServiceError(t *testing.T) {
	service := &fakeLogoutService{
		err: errors.New("database connection failed"),
	}

	handler := NewLogoutHandler(
		service,
		false,
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/logout",
		nil,
	)

	req.AddCookie(&http.Cookie{
		Name:  "devflow_session",
		Value: "session-token",
	})

	recorder := httptest.NewRecorder()

	handler.Logout(recorder, req)

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

func TestLogoutUserSecureCookie(t *testing.T) {
	service := &fakeLogoutService{}

	handler := NewLogoutHandler(
		service,
		true,
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/logout",
		nil,
	)

	req.AddCookie(&http.Cookie{
		Name:  "devflow_session",
		Value: "session-token",
	})

	recorder := httptest.NewRecorder()

	handler.Logout(recorder, req)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNoContent,
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

	if !cookies[0].Secure {
		t.Fatal("expected cookie to be Secure")
	}
}
