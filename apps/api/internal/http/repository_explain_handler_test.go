package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kristopher1027/devflow-ai/internal/integration/anthropic"
	"github.com/kristopher1027/devflow-ai/internal/repository"
	"github.com/kristopher1027/devflow-ai/internal/service"
)

type explainTestService struct {
	result *service.RepositoryExplanation
	err    error

	gotRequesterID  string
	gotRepositoryID string
	calls           int
}

func (f *explainTestService) Explain(
	ctx context.Context,
	requesterID string,
	repositoryID string,
) (*service.RepositoryExplanation, error) {
	f.calls++
	f.gotRequesterID = requesterID
	f.gotRepositoryID = repositoryID
	return f.result, f.err
}

type explainTestAuth struct {
	allow bool
}

func (a explainTestAuth) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !a.allow {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), userIDContextKey, "user-1")
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func explainRequest(method string, path string, withUser bool) *http.Request {
	request := httptest.NewRequest(method, path, nil)
	if withUser {
		request = request.WithContext(
			context.WithValue(request.Context(), userIDContextKey, "user-1"),
		)
	}
	return request
}

func TestRepositoryExplainHandlerReturnsExplanation(t *testing.T) {
	svc := &explainTestService{result: &service.RepositoryExplanation{
		SnapshotID: "snap-1", CommitSHA: "abc123", Text: "## What it does",
	}}
	recorder := httptest.NewRecorder()

	NewRepositoryExplainHandler(svc).Explain(
		recorder,
		explainRequest(http.MethodPost, "/repositories/repo-1/explain", true),
	)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.JSONEq(t,
		`{"snapshot_id":"snap-1","commit_sha":"abc123","explanation":"## What it does"}`,
		recorder.Body.String(),
	)
	require.Equal(t, "user-1", svc.gotRequesterID)
	require.Equal(t, "repo-1", svc.gotRepositoryID)
}

func TestRepositoryExplainHandlerRequiresUser(t *testing.T) {
	svc := &explainTestService{}
	recorder := httptest.NewRecorder()

	NewRepositoryExplainHandler(svc).Explain(
		recorder,
		explainRequest(http.MethodPost, "/repositories/repo-1/explain", false),
	)

	require.Equal(t, http.StatusUnauthorized, recorder.Code)
	require.Zero(t, svc.calls)
}

func TestRepositoryExplainHandlerRequiresRepositoryID(t *testing.T) {
	svc := &explainTestService{}
	recorder := httptest.NewRecorder()

	NewRepositoryExplainHandler(svc).Explain(
		recorder,
		explainRequest(http.MethodPost, "/repositories//explain", true),
	)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	require.Zero(t, svc.calls)
}

func TestRepositoryExplainHandlerMapsErrors(t *testing.T) {
	cases := map[string]struct {
		err    error
		status int
	}{
		"forbidden":      {service.ErrRepositoryExplainUnauthorized, http.StatusForbidden},
		"repo missing":   {repository.ErrRepositoryNotFound, http.StatusNotFound},
		"project gone":   {repository.ErrProjectNotFound, http.StatusNotFound},
		"not synced":     {service.ErrRepositoryExplainNoSnapshot, http.StatusConflict},
		"no files":       {service.ErrRepositoryExplainNoFiles, http.StatusConflict},
		"ai not set up":  {anthropic.ErrAPIKeyRequired, http.StatusServiceUnavailable},
		"provider error": {fmt.Errorf("%w: 500", anthropic.ErrUnexpectedStatus), http.StatusBadGateway},
		"empty reply":    {anthropic.ErrEmptyResponse, http.StatusBadGateway},
		"unknown":        {errors.New("boom secret detail"), http.StatusInternalServerError},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			recorder := httptest.NewRecorder()

			NewRepositoryExplainHandler(&explainTestService{err: tc.err}).Explain(
				recorder,
				explainRequest(http.MethodPost, "/repositories/repo-1/explain", true),
			)

			require.Equal(t, tc.status, recorder.Code)
			require.NotContains(t, recorder.Body.String(), "secret detail")
		})
	}
}

func TestWithRepositoryExplainRoutesOnlyExplainPosts(t *testing.T) {
	svc := &explainTestService{result: &service.RepositoryExplanation{Text: "ok"}}
	handler := NewRepositoryExplainHandler(svc)

	fallbackCalls := 0
	fallback := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fallbackCalls++
		w.WriteHeader(http.StatusTeapot)
	})

	routed := WithRepositoryExplain(fallback, explainTestAuth{allow: true}, handler)

	explain := httptest.NewRecorder()
	routed.ServeHTTP(explain, httptest.NewRequest(http.MethodPost, "/repositories/repo-1/explain", nil))
	require.Equal(t, http.StatusOK, explain.Code)
	require.Equal(t, 1, svc.calls)
	require.Zero(t, fallbackCalls)

	for _, tc := range []struct{ method, path string }{
		{http.MethodGet, "/repositories/repo-1/explain"},
		{http.MethodPost, "/repositories/repo-1/sync"},
		{http.MethodPost, "/repositories/a/b/explain"},
		{http.MethodGet, "/health"},
	} {
		recorder := httptest.NewRecorder()
		routed.ServeHTTP(recorder, httptest.NewRequest(tc.method, tc.path, nil))
		require.Equal(t, http.StatusTeapot, recorder.Code, tc.method+" "+tc.path)
	}
	require.Equal(t, 4, fallbackCalls)
	require.Equal(t, 1, svc.calls)
}

func TestWithRepositoryExplainRequiresAuthentication(t *testing.T) {
	svc := &explainTestService{}
	routed := WithRepositoryExplain(
		http.NotFoundHandler(),
		explainTestAuth{allow: false},
		NewRepositoryExplainHandler(svc),
	)
	recorder := httptest.NewRecorder()

	routed.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/repositories/repo-1/explain", nil))

	require.Equal(t, http.StatusUnauthorized, recorder.Code)
	require.Zero(t, svc.calls)
}
