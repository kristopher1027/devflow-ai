package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kristopher1027/devflow-ai/internal/domain"
	"github.com/kristopher1027/devflow-ai/internal/repository"
	"github.com/kristopher1027/devflow-ai/internal/service"
)

type fakeRepositoryService struct {
	repository   *domain.Repository
	repositories []*domain.Repository
	err          error

	gotRequesterID  string
	gotProjectID    string
	gotRepositoryID string
	gotProvider     string
	gotExternalID   string
}

func (f *fakeRepositoryService) Create(
	ctx context.Context,
	requesterID string,
	projectID string,
	provider string,
	externalID string,
	owner string,
	name string,
	fullName string,
	defaultBranch string,
	htmlURL string,
	cloneURL string,
	isPrivate bool,
) (*domain.Repository, error) {
	f.gotRequesterID = requesterID
	f.gotProjectID = projectID
	f.gotProvider = provider
	f.gotExternalID = externalID

	if f.err != nil {
		return nil, f.err
	}

	return f.repository, nil
}

func (f *fakeRepositoryService) FindByID(
	ctx context.Context,
	requesterID string,
	id string,
) (*domain.Repository, error) {
	f.gotRequesterID = requesterID
	f.gotRepositoryID = id

	if f.err != nil {
		return nil, f.err
	}

	return f.repository, nil
}

func (f *fakeRepositoryService) ListByProjectID(
	ctx context.Context,
	requesterID string,
	projectID string,
) ([]*domain.Repository, error) {
	f.gotRequesterID = requesterID
	f.gotProjectID = projectID

	if f.err != nil {
		return nil, f.err
	}

	return f.repositories, nil
}

func (f *fakeRepositoryService) Delete(
	ctx context.Context,
	requesterID string,
	id string,
) error {
	f.gotRequesterID = requesterID
	f.gotRepositoryID = id

	return f.err
}

func TestRepositoryHandlerCreate(t *testing.T) {
	repositoryService := &fakeRepositoryService{
		repository: &domain.Repository{
			ID:        "repository-123",
			ProjectID: "project-123",
		},
	}
	handler := NewRepositoryHandler(repositoryService)

	req := authenticatedRequest(
		http.MethodPost,
		"/projects/project-123/repositories",
		`{"provider":"github","external_id":"12345","owner":"octocat","name":"hello-world","full_name":"octocat/hello-world","default_branch":"main","html_url":"https://github.com/octocat/hello-world","clone_url":"https://github.com/octocat/hello-world.git","is_private":true}`,
		"user-123",
	)
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	if repositoryService.gotRequesterID != "user-123" {
		t.Fatalf("expected requester ID user-123, got %s", repositoryService.gotRequesterID)
	}

	if repositoryService.gotProjectID != "project-123" {
		t.Fatalf("expected project ID project-123, got %s", repositoryService.gotProjectID)
	}

	if repositoryService.gotProvider != "github" {
		t.Fatalf("expected provider github, got %s", repositoryService.gotProvider)
	}

	if repositoryService.gotExternalID != "12345" {
		t.Fatalf("expected external ID 12345, got %s", repositoryService.gotExternalID)
	}
}

func TestRepositoryHandlerCreateRequiresAuth(t *testing.T) {
	handler := NewRepositoryHandler(&fakeRepositoryService{})
	req := httptest.NewRequest(
		http.MethodPost,
		"/projects/project-123/repositories",
		nil,
	)
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestRepositoryHandlerListByProjectID(t *testing.T) {
	repositoryService := &fakeRepositoryService{
		repositories: []*domain.Repository{
			{ID: "repository-123", ProjectID: "project-123"},
		},
	}
	handler := NewRepositoryHandler(repositoryService)
	req := authenticatedRequest(
		http.MethodGet,
		"/projects/project-123/repositories",
		"",
		"user-123",
	)
	rec := httptest.NewRecorder()

	handler.ListByProjectID(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if repositoryService.gotProjectID != "project-123" {
		t.Fatalf("expected project ID project-123, got %s", repositoryService.gotProjectID)
	}
}

func TestRepositoryHandlerGetByIDNotFound(t *testing.T) {
	handler := NewRepositoryHandler(&fakeRepositoryService{
		err: repository.ErrRepositoryNotFound,
	})
	req := authenticatedRequest(
		http.MethodGet,
		"/repositories/repository-123",
		"",
		"user-123",
	)
	rec := httptest.NewRecorder()

	handler.GetByID(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestRepositoryHandlerDelete(t *testing.T) {
	repositoryService := &fakeRepositoryService{}
	handler := NewRepositoryHandler(repositoryService)
	req := authenticatedRequest(
		http.MethodDelete,
		"/repositories/repository-123",
		"",
		"user-123",
	)
	rec := httptest.NewRecorder()

	handler.Delete(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rec.Code)
	}

	if repositoryService.gotRepositoryID != "repository-123" {
		t.Fatalf("expected repository ID repository-123, got %s", repositoryService.gotRepositoryID)
	}
}

func TestRepositoryHandlerDeleteUnauthorized(t *testing.T) {
	handler := NewRepositoryHandler(&fakeRepositoryService{
		err: service.ErrRepositoryUnauthorized,
	})
	req := authenticatedRequest(
		http.MethodDelete,
		"/repositories/repository-123",
		"",
		"user-123",
	)
	rec := httptest.NewRecorder()

	handler.Delete(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
}

func TestRepositoryHandlerCreateValidationError(t *testing.T) {
	handler := NewRepositoryHandler(&fakeRepositoryService{
		err: service.ErrRepositoryNameRequired,
	})
	req := authenticatedRequest(
		http.MethodPost,
		"/projects/project-123/repositories",
		`{}`,
		"user-123",
	)
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestRepositoryHandlerUnexpectedError(t *testing.T) {
	handler := NewRepositoryHandler(&fakeRepositoryService{
		err: errors.New("database failure"),
	})
	req := authenticatedRequest(
		http.MethodGet,
		"/repositories/repository-123",
		"",
		"user-123",
	)
	rec := httptest.NewRecorder()

	handler.GetByID(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}
