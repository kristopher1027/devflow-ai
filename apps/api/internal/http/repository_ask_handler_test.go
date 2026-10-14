package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kristopher1027/devflow-ai/internal/domain"
	"github.com/kristopher1027/devflow-ai/internal/integration/anthropic"
	"github.com/kristopher1027/devflow-ai/internal/repository"
	"github.com/kristopher1027/devflow-ai/internal/service"
)

type askTestService struct {
	result *service.RepositoryAnswer
	err    error

	gotRequesterID  string
	gotRepositoryID string
	gotQuestion     string
	calls           int
}

func (f *askTestService) Ask(
	ctx context.Context,
	requesterID string,
	repositoryID string,
	question string,
) (*service.RepositoryAnswer, error) {
	f.calls++
	f.gotRequesterID = requesterID
	f.gotRepositoryID = repositoryID
	f.gotQuestion = question
	return f.result, f.err
}

func askRequest(method string, path string, body string, withUser bool) *http.Request {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	if withUser {
		request = request.WithContext(
			context.WithValue(request.Context(), userIDContextKey, "user-1"),
		)
	}
	return request
}

func TestRepositoryAskHandlerReturnsAnswerWithSources(t *testing.T) {
	svc := &askTestService{result: &service.RepositoryAnswer{
		SnapshotID: "snap-1",
		CommitSHA:  "abc123",
		Answer:     "Sessions are checked in auth.go.",
		Sources: []*domain.RepositoryAIContextSource{
			{FilePath: "auth.go", StartLine: 1, EndLine: 20},
			nil,
		},
	}}
	recorder := httptest.NewRecorder()

	NewRepositoryAskHandler(svc).Ask(
		recorder,
		askRequest(http.MethodPost, "/repositories/repo-1/ask",
			`{"question":"how are sessions checked?"}`, true),
	)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.JSONEq(t, `{
		"snapshot_id":"snap-1","commit_sha":"abc123",
		"answer":"Sessions are checked in auth.go.",
		"sources":[{"file_path":"auth.go","start_line":1,"end_line":20}]
	}`, recorder.Body.String())
	require.Equal(t, "user-1", svc.gotRequesterID)
	require.Equal(t, "repo-1", svc.gotRepositoryID)
	require.Equal(t, "how are sessions checked?", svc.gotQuestion)
}

func TestRepositoryAskHandlerRejectsBadRequests(t *testing.T) {
	cases := map[string]struct {
		path     string
		body     string
		withUser bool
		status   int
	}{
		"no user":      {"/repositories/repo-1/ask", `{"question":"x"}`, false, http.StatusUnauthorized},
		"no repo id":   {"/repositories//ask", `{"question":"x"}`, true, http.StatusBadRequest},
		"invalid json": {"/repositories/repo-1/ask", `not json`, true, http.StatusBadRequest},
		"empty body":   {"/repositories/repo-1/ask", ``, true, http.StatusBadRequest},
		"too large": {
			"/repositories/repo-1/ask",
			`{"question":"` + strings.Repeat("a", maxAskRequestBytes) + `"}`,
			true, http.StatusBadRequest,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			svc := &askTestService{}
			recorder := httptest.NewRecorder()

			NewRepositoryAskHandler(svc).Ask(
				recorder,
				askRequest(http.MethodPost, tc.path, tc.body, tc.withUser),
			)

			require.Equal(t, tc.status, recorder.Code)
			require.Zero(t, svc.calls)
		})
	}
}

func TestRepositoryAskHandlerMapsErrors(t *testing.T) {
	cases := map[string]struct {
		err    error
		status int
	}{
		"question required": {service.ErrRepositoryAskQuestionRequired, http.StatusBadRequest},
		"question too long": {service.ErrRepositoryAskQuestionTooLong, http.StatusBadRequest},
		"forbidden":         {service.ErrRepositoryAskUnauthorized, http.StatusForbidden},
		"not synced":        {service.ErrRepositoryAskNoSnapshot, http.StatusConflict},
		"no matches":        {service.ErrRepositoryAskNoMatches, http.StatusNotFound},
		"repo missing":      {repository.ErrRepositoryNotFound, http.StatusNotFound},
		"ai not set up":     {anthropic.ErrAPIKeyRequired, http.StatusServiceUnavailable},
		"provider error":    {fmt.Errorf("%w: 500", anthropic.ErrUnexpectedStatus), http.StatusBadGateway},
		"unknown":           {errors.New("boom secret detail"), http.StatusInternalServerError},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			recorder := httptest.NewRecorder()

			NewRepositoryAskHandler(&askTestService{err: tc.err}).Ask(
				recorder,
				askRequest(http.MethodPost, "/repositories/repo-1/ask",
					`{"question":"login flow"}`, true),
			)

			require.Equal(t, tc.status, recorder.Code)
			require.NotContains(t, recorder.Body.String(), "secret detail")
		})
	}
}

func TestWithRepositoryAskRoutesOnlyAskPosts(t *testing.T) {
	svc := &askTestService{result: &service.RepositoryAnswer{Answer: "ok"}}

	fallbackCalls := 0
	fallback := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fallbackCalls++
		w.WriteHeader(http.StatusTeapot)
	})

	routed := WithRepositoryAsk(
		fallback, explainTestAuth{allow: true}, NewRepositoryAskHandler(svc),
	)

	asked := httptest.NewRecorder()
	routed.ServeHTTP(asked, httptest.NewRequest(
		http.MethodPost, "/repositories/repo-1/ask",
		strings.NewReader(`{"question":"login"}`),
	))
	require.Equal(t, http.StatusOK, asked.Code)
	require.Equal(t, 1, svc.calls)
	require.Zero(t, fallbackCalls)

	for _, tc := range []struct{ method, path string }{
		{http.MethodGet, "/repositories/repo-1/ask"},
		{http.MethodPost, "/repositories/repo-1/explain"},
		{http.MethodPost, "/repositories/a/b/ask"},
		{http.MethodGet, "/health"},
	} {
		recorder := httptest.NewRecorder()
		routed.ServeHTTP(recorder, httptest.NewRequest(tc.method, tc.path, nil))
		require.Equal(t, http.StatusTeapot, recorder.Code, tc.method+" "+tc.path)
	}
	require.Equal(t, 4, fallbackCalls)
	require.Equal(t, 1, svc.calls)
}

func TestWithRepositoryAskRequiresAuthentication(t *testing.T) {
	svc := &askTestService{}
	routed := WithRepositoryAsk(
		http.NotFoundHandler(), explainTestAuth{allow: false}, NewRepositoryAskHandler(svc),
	)
	recorder := httptest.NewRecorder()

	routed.ServeHTTP(recorder, httptest.NewRequest(
		http.MethodPost, "/repositories/repo-1/ask",
		strings.NewReader(`{"question":"login"}`),
	))

	require.Equal(t, http.StatusUnauthorized, recorder.Code)
	require.Zero(t, svc.calls)
}
