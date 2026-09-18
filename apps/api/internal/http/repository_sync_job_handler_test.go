package http

import (
	"context"
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

type fakeRepositorySyncJobService struct {
	job       *domain.RepositorySyncJob
	err       error
	findCalls int
	gotUserID string
	gotJobID  string
}

func (f *fakeRepositorySyncJobService) FindByID(
	ctx context.Context,
	requesterID string,
	id string,
) (*domain.RepositorySyncJob, error) {
	f.findCalls++
	f.gotUserID = requesterID
	f.gotJobID = id

	if f.err != nil {
		return nil, f.err
	}

	return f.job, nil
}

func testHTTPRepositorySyncJob() *domain.RepositorySyncJob {
	now := time.Now()

	return &domain.RepositorySyncJob{
		ID:           "sync-job-123",
		RepositoryID: "repository-123",
		Status:       domain.RepositorySyncJobStatusSucceeded,
		Attempts:     1,
		CreatedAt:    now,
		UpdatedAt:    now,
		CompletedAt:  &now,
	}
}

func authenticatedSyncJobRequest(
	method string,
	path string,
	userID string,
) *http.Request {
	req := httptest.NewRequest(method, path, nil)

	ctx := context.WithValue(
		req.Context(),
		userIDContextKey,
		userID,
	)

	return req.WithContext(ctx)
}

func TestRepositorySyncJobHandlerGetByID(t *testing.T) {
	job := testHTTPRepositorySyncJob()

	fakeService := &fakeRepositorySyncJobService{
		job: job,
	}

	handler := NewRepositorySyncJobHandler(fakeService)

	req := authenticatedSyncJobRequest(
		http.MethodGet,
		"/repositories/sync-jobs/sync-job-123",
		"user-123",
	)

	rec := httptest.NewRecorder()

	handler.GetByID(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	if fakeService.findCalls != 1 {
		t.Fatalf("FindByID calls = %d, want 1", fakeService.findCalls)
	}

	if fakeService.gotUserID != "user-123" {
		t.Fatalf(
			"requester ID = %q, want %q",
			fakeService.gotUserID,
			"user-123",
		)
	}

	if fakeService.gotJobID != "sync-job-123" {
		t.Fatalf(
			"job ID = %q, want %q",
			fakeService.gotJobID,
			"sync-job-123",
		)
	}

	if !strings.Contains(
		rec.Header().Get("Content-Type"),
		"application/json",
	) {
		t.Fatalf("Content-Type = %q, want application/json",
			rec.Header().Get("Content-Type"))
	}

	if !strings.Contains(rec.Body.String(), `"ID":"sync-job-123"`) {
		t.Fatalf("response body does not contain job ID: %s", rec.Body.String())
	}
}

func TestRepositorySyncJobHandlerGetByIDMissingID(t *testing.T) {
	fakeService := &fakeRepositorySyncJobService{}

	handler := NewRepositorySyncJobHandler(fakeService)

	req := authenticatedSyncJobRequest(
		http.MethodGet,
		"/repositories/sync-jobs/",
		"user-123",
	)

	rec := httptest.NewRecorder()

	handler.GetByID(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"status = %d, want %d",
			rec.Code,
			http.StatusBadRequest,
		)
	}

	if fakeService.findCalls != 0 {
		t.Fatalf("FindByID should not be called")
	}
}

func TestRepositorySyncJobHandlerGetByIDUnauthorized(t *testing.T) {
	fakeService := &fakeRepositorySyncJobService{
		job: testHTTPRepositorySyncJob(),
	}

	handler := NewRepositorySyncJobHandler(fakeService)

	req := httptest.NewRequest(
		http.MethodGet,
		"/repositories/sync-jobs/sync-job-123",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.GetByID(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf(
			"status = %d, want %d",
			rec.Code,
			http.StatusUnauthorized,
		)
	}

	if fakeService.findCalls != 0 {
		t.Fatalf("FindByID should not be called")
	}
}

func TestRepositorySyncJobHandlerGetByIDForbidden(t *testing.T) {
	fakeService := &fakeRepositorySyncJobService{
		err: service.ErrRepositorySyncJobUnauthorized,
	}

	handler := NewRepositorySyncJobHandler(fakeService)

	req := authenticatedSyncJobRequest(
		http.MethodGet,
		"/repositories/sync-jobs/sync-job-123",
		"user-123",
	)

	rec := httptest.NewRecorder()

	handler.GetByID(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf(
			"status = %d, want %d",
			rec.Code,
			http.StatusForbidden,
		)
	}
}

func TestRepositorySyncJobHandlerGetByIDNotFound(t *testing.T) {
	fakeService := &fakeRepositorySyncJobService{
		err: repository.ErrRepositorySyncJobNotFound,
	}

	handler := NewRepositorySyncJobHandler(fakeService)

	req := authenticatedSyncJobRequest(
		http.MethodGet,
		"/repositories/sync-jobs/sync-job-123",
		"user-123",
	)

	rec := httptest.NewRecorder()

	handler.GetByID(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf(
			"status = %d, want %d",
			rec.Code,
			http.StatusNotFound,
		)
	}
}

func TestRepositorySyncJobHandlerGetByIDInternalError(t *testing.T) {
	expectedErr := errors.New("database failure")

	fakeService := &fakeRepositorySyncJobService{
		err: expectedErr,
	}

	handler := NewRepositorySyncJobHandler(fakeService)

	req := authenticatedSyncJobRequest(
		http.MethodGet,
		"/repositories/sync-jobs/sync-job-123",
		"user-123",
	)

	rec := httptest.NewRecorder()

	handler.GetByID(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"status = %d, want %d",
			rec.Code,
			http.StatusInternalServerError,
		)
	}
}

var _ service.RepositorySyncJobService = (*fakeRepositorySyncJobService)(nil)
