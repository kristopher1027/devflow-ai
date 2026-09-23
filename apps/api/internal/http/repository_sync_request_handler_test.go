package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kristopher1027/devflow-ai/internal/domain"
	"github.com/kristopher1027/devflow-ai/internal/service"
)

type fakeRepositorySyncRequestService struct {
	job *domain.RepositorySyncJob
	err error

	gotRequesterID  string
	gotRepositoryID string
}

func (f *fakeRepositorySyncRequestService) RequestSync(
	ctx context.Context,
	requesterID string,
	repositoryID string,
) (*domain.RepositorySyncJob, error) {
	f.gotRequesterID = requesterID
	f.gotRepositoryID = repositoryID

	if f.err != nil {
		return nil, f.err
	}

	return f.job, nil
}

func TestRepositorySyncRequestHandlerRequestSync(t *testing.T) {
	syncService := &fakeRepositorySyncRequestService{
		job: &domain.RepositorySyncJob{
			ID:           "job-123",
			RepositoryID: "repository-123",
			Status:       domain.RepositorySyncJobStatusPending,
		},
	}

	handler := NewRepositorySyncRequestHandler(syncService)

	request := httptest.NewRequest(
		http.MethodPost,
		"/repositories/repository-123/sync",
		nil,
	)

	request = request.WithContext(
		context.WithValue(
			request.Context(),
			userIDContextKey,
			"user-123",
		),
	)

	recorder := httptest.NewRecorder()

	handler.RequestSync(recorder, request)

	if recorder.Code != http.StatusAccepted {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusAccepted,
			recorder.Code,
		)
	}

	if syncService.gotRequesterID != "user-123" {
		t.Fatalf(
			"expected requester ID user-123, got %s",
			syncService.gotRequesterID,
		)
	}

	if syncService.gotRepositoryID != "repository-123" {
		t.Fatalf(
			"expected repository ID repository-123, got %s",
			syncService.gotRepositoryID,
		)
	}

	if !strings.Contains(
		recorder.Body.String(),
		`"ID":"job-123"`,
	) {
		t.Fatalf(
			"expected response to contain job ID, got %s",
			recorder.Body.String(),
		)
	}
}

func TestRepositorySyncRequestHandlerUnauthorized(t *testing.T) {
	syncService := &fakeRepositorySyncRequestService{
		job: &domain.RepositorySyncJob{
			ID:           "job-123",
			RepositoryID: "repository-123",
			Status:       domain.RepositorySyncJobStatusPending,
		},
	}

	handler := NewRepositorySyncRequestHandler(syncService)

	request := httptest.NewRequest(
		http.MethodPost,
		"/repositories/repository-123/sync",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.RequestSync(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			recorder.Code,
		)
	}
}

func TestRepositorySyncRequestHandlerForbidden(t *testing.T) {
	syncService := &fakeRepositorySyncRequestService{
		err: service.ErrRepositorySyncRequestUnauthorized,
	}

	handler := NewRepositorySyncRequestHandler(syncService)

	request := httptest.NewRequest(
		http.MethodPost,
		"/repositories/repository-123/sync",
		nil,
	)

	request = request.WithContext(
		context.WithValue(
			request.Context(),
			userIDContextKey,
			"user-123",
		),
	)

	recorder := httptest.NewRecorder()

	handler.RequestSync(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusForbidden,
			recorder.Code,
		)
	}
}

func TestRepositorySyncRequestHandlerRepositoryNotFound(t *testing.T) {
	syncService := &fakeRepositorySyncRequestService{
		err: errors.New("repository not found"),
	}

	handler := NewRepositorySyncRequestHandler(syncService)

	request := httptest.NewRequest(
		http.MethodPost,
		"/repositories/repository-123/sync",
		nil,
	)

	request = request.WithContext(
		context.WithValue(
			request.Context(),
			userIDContextKey,
			"user-123",
		),
	)

	recorder := httptest.NewRecorder()

	handler.RequestSync(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusInternalServerError,
			recorder.Code,
		)
	}
}
