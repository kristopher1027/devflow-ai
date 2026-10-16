package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kristopher1027/devflow-ai/internal/auth"
)

func TestWithLogoutHandlesOnlyPostAuthLogout(t *testing.T) {
	service := &fakeLogoutService{}
	fallbackCalls := 0
	fallback := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fallbackCalls++
		w.WriteHeader(http.StatusTeapot)
	})
	routed := WithLogout(fallback, NewLogoutHandler(service, false))

	request := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	request.AddCookie(&http.Cookie{Name: "devflow_session", Value: "token-1"})
	recorder := httptest.NewRecorder()

	routed.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusNoContent, recorder.Code)
	require.Equal(t, auth.HashSessionToken("token-1"), service.tokenHash)
	require.Contains(t, recorder.Header().Get("Set-Cookie"), "devflow_session=")
	require.Contains(t, recorder.Header().Get("Set-Cookie"), "Max-Age=0")
	require.Zero(t, fallbackCalls)

	for _, tc := range []struct{ method, path string }{
		{http.MethodGet, "/auth/logout"},
		{http.MethodPost, "/auth/login"},
		{http.MethodPost, "/auth/logout/extra"},
		{http.MethodGet, "/health"},
	} {
		recorder := httptest.NewRecorder()
		routed.ServeHTTP(recorder, httptest.NewRequest(tc.method, tc.path, nil))
		require.Equal(t, http.StatusTeapot, recorder.Code, tc.method+" "+tc.path)
	}
	require.Equal(t, 4, fallbackCalls)
}
