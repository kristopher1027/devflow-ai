package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
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
