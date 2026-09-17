package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kristopher1027/devflow-ai/internal/domain"
	"github.com/kristopher1027/devflow-ai/internal/service"
)

type fakeGitHubRepositoryImportQueue struct {
	err            error
	gotRequesterID string
	gotProjectID   string
}

func (f *fakeGitHubRepositoryImportQueue) Enqueue(
	ctx context.Context,
	request service.GitHubRepositoryImportRequest,
) (*domain.GitHubRepositoryImportJob, error) {
	f.gotRequesterID = request.RequesterID
	f.gotProjectID = request.ProjectID
	if f.err != nil {
		return nil, f.err
	}
	return &domain.GitHubRepositoryImportJob{ID: "job-123"}, nil
}

func TestGitHubRepositoryImportHandlerImport(t *testing.T) {
	importQueue := &fakeGitHubRepositoryImportQueue{}
	handler := NewGitHubRepositoryImportHandler(importQueue)
	req := authenticatedRequest(
		http.MethodPost,
		"/projects/project-123/repositories/import",
		"",
		"user-123",
	)
	rec := httptest.NewRecorder()

	handler.Import(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d", http.StatusAccepted, rec.Code)
	}
	if importQueue.gotRequesterID != "user-123" || importQueue.gotProjectID != "project-123" {
		t.Fatalf("unexpected import arguments: %+v", importQueue)
	}
}

func TestGitHubRepositoryImportHandlerRequiresAuth(t *testing.T) {
	handler := NewGitHubRepositoryImportHandler(&fakeGitHubRepositoryImportQueue{})
	req := httptest.NewRequest(
		http.MethodPost,
		"/projects/project-123/repositories/import",
		nil,
	)
	rec := httptest.NewRecorder()

	handler.Import(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestGitHubRepositoryImportHandlerQueueFailure(t *testing.T) {
	handler := NewGitHubRepositoryImportHandler(&fakeGitHubRepositoryImportQueue{
		err: context.Canceled,
	})
	req := authenticatedRequest(
		http.MethodPost,
		"/projects/project-123/repositories/import",
		"",
		"user-123",
	)
	rec := httptest.NewRecorder()

	handler.Import(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, rec.Code)
	}
}
