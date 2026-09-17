package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kristopher1027/devflow-ai/internal/domain"
	"github.com/kristopher1027/devflow-ai/internal/repository"
	"github.com/kristopher1027/devflow-ai/internal/service"
)

type fakeGitHubRepositoryImportJobService struct {
	job *domain.GitHubRepositoryImportJob
	err error

	gotRequesterID string
	gotJobID       string
}

func (f *fakeGitHubRepositoryImportJobService) FindByID(
	ctx context.Context,
	requesterID string,
	id string,
) (*domain.GitHubRepositoryImportJob, error) {
	f.gotRequesterID = requesterID
	f.gotJobID = id

	if f.err != nil {
		return nil, f.err
	}

	return f.job, nil
}

func TestGitHubRepositoryImportJobHandlerGetByID(t *testing.T) {
	job := &domain.GitHubRepositoryImportJob{
		ID:          "job-123",
		RequesterID: "user-123",
		ProjectID:   "project-123",
		Status:      domain.GitHubRepositoryImportJobStatusRunning,
		Attempts:    1,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	jobService := &fakeGitHubRepositoryImportJobService{
		job: job,
	}

	handler := NewGitHubRepositoryImportJobHandler(jobService)

	req := authenticatedRequest(
		http.MethodGet,
		"/github/repository-import-jobs/job-123",
		"",
		"user-123",
	)
	rec := httptest.NewRecorder()

	handler.GetByID(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}

	if jobService.gotRequesterID != "user-123" {
		t.Fatalf(
			"expected requester ID user-123, got %s",
			jobService.gotRequesterID,
		)
	}

	if jobService.gotJobID != "job-123" {
		t.Fatalf(
			"expected job ID job-123, got %s",
			jobService.gotJobID,
		)
	}
}

func TestGitHubRepositoryImportJobHandlerGetByIDRequiresAuth(t *testing.T) {
	handler := NewGitHubRepositoryImportJobHandler(
		&fakeGitHubRepositoryImportJobService{},
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/github/repository-import-jobs/job-123",
		nil,
	)
	rec := httptest.NewRecorder()

	handler.GetByID(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			rec.Code,
		)
	}
}

func TestGitHubRepositoryImportJobHandlerGetByIDNotFound(t *testing.T) {
	handler := NewGitHubRepositoryImportJobHandler(
		&fakeGitHubRepositoryImportJobService{
			err: repository.ErrGitHubRepositoryImportJobNotFound,
		},
	)

	req := authenticatedRequest(
		http.MethodGet,
		"/github/repository-import-jobs/job-123",
		"",
		"user-123",
	)
	rec := httptest.NewRecorder()

	handler.GetByID(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNotFound,
			rec.Code,
		)
	}
}

func TestGitHubRepositoryImportJobHandlerGetByIDUnauthorized(t *testing.T) {
	handler := NewGitHubRepositoryImportJobHandler(
		&fakeGitHubRepositoryImportJobService{
			err: service.ErrGitHubRepositoryImportJobUnauthorized,
		},
	)

	req := authenticatedRequest(
		http.MethodGet,
		"/github/repository-import-jobs/job-123",
		"",
		"user-123",
	)
	rec := httptest.NewRecorder()

	handler.GetByID(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusForbidden,
			rec.Code,
		)
	}
}

func TestGitHubRepositoryImportJobHandlerGetByIDUnexpectedError(t *testing.T) {
	handler := NewGitHubRepositoryImportJobHandler(
		&fakeGitHubRepositoryImportJobService{
			err: errors.New("database failure"),
		},
	)

	req := authenticatedRequest(
		http.MethodGet,
		"/github/repository-import-jobs/job-123",
		"",
		"user-123",
	)
	rec := httptest.NewRecorder()

	handler.GetByID(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusInternalServerError,
			rec.Code,
		)
	}
}
