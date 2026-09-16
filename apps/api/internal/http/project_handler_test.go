package http

import (
	"context"
	//"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/kristopher1027/devflow-ai/internal/domain"
	"github.com/kristopher1027/devflow-ai/internal/repository"
	"github.com/kristopher1027/devflow-ai/internal/service"
)

type fakeProjectService struct {
	project  *domain.Project
	projects []*domain.Project
	err      error

	gotRequesterID string
	gotWorkspaceID string
	gotProjectID   string
	gotName        string
	gotDescription *string
}

func (f *fakeProjectService) Create(
	ctx context.Context,
	requesterID string,
	workspaceID string,
	name string,
	description *string,
) (*domain.Project, error) {
	f.gotRequesterID = requesterID
	f.gotWorkspaceID = workspaceID
	f.gotName = name
	f.gotDescription = description

	if f.err != nil {
		return nil, f.err
	}

	return f.project, nil
}

func (f *fakeProjectService) FindByID(
	ctx context.Context,
	requesterID string,
	id string,
) (*domain.Project, error) {
	f.gotRequesterID = requesterID
	f.gotProjectID = id

	if f.err != nil {
		return nil, f.err
	}

	return f.project, nil
}

func (f *fakeProjectService) ListByWorkspaceID(
	ctx context.Context,
	requesterID string,
	workspaceID string,
) ([]*domain.Project, error) {
	f.gotRequesterID = requesterID
	f.gotWorkspaceID = workspaceID

	if f.err != nil {
		return nil, f.err
	}

	return f.projects, nil
}

func (f *fakeProjectService) Delete(
	ctx context.Context,
	requesterID string,
	id string,
) error {
	f.gotRequesterID = requesterID
	f.gotProjectID = id

	return f.err
}

func TestProjectHandlerCreate(t *testing.T) {
	now := time.Now()

	description := "Developer productivity project"

	projectService := &fakeProjectService{
		project: &domain.Project{
			ID:          "project-123",
			WorkspaceID: "workspace-123",
			Name:        "DevFlow",
			Description: &description,
			CreatedBy:   "user-123",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	handler := NewProjectHandler(projectService)

	req := authenticatedRequest(
		http.MethodPost,
		"/workspaces/workspace-123/projects",
		`{"name":"DevFlow","description":"Developer productivity project"}`,
		"user-123",
	)

	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusCreated,
			rec.Code,
		)
	}

	if projectService.gotRequesterID != "user-123" {
		t.Fatalf(
			"expected requester ID %q, got %q",
			"user-123",
			projectService.gotRequesterID,
		)
	}

	if projectService.gotWorkspaceID != "workspace-123" {
		t.Fatalf(
			"expected workspace ID %q, got %q",
			"workspace-123",
			projectService.gotWorkspaceID,
		)
	}

	if projectService.gotName != "DevFlow" {
		t.Fatalf(
			"expected project name %q, got %q",
			"DevFlow",
			projectService.gotName,
		)
	}

	if projectService.gotDescription == nil {
		t.Fatal("expected description")
	}

	if *projectService.gotDescription != description {
		t.Fatalf(
			"expected description %q, got %q",
			description,
			*projectService.gotDescription,
		)
	}
}

func TestProjectHandlerCreateRequiresAuth(t *testing.T) {
	projectService := &fakeProjectService{}

	handler := NewProjectHandler(projectService)

	req := httptest.NewRequest(
		http.MethodPost,
		"/workspaces/workspace-123/projects",
		strings.NewReader(`{"name":"DevFlow"}`),
	)

	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			rec.Code,
		)
	}
}

func TestProjectHandlerCreateInvalidBody(t *testing.T) {
	projectService := &fakeProjectService{}

	handler := NewProjectHandler(projectService)

	req := authenticatedRequest(
		http.MethodPost,
		"/workspaces/workspace-123/projects",
		`invalid-json`,
		"user-123",
	)

	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestProjectHandlerCreateNameRequired(t *testing.T) {
	projectService := &fakeProjectService{
		err: service.ErrProjectNameRequired,
	}

	handler := NewProjectHandler(projectService)

	req := authenticatedRequest(
		http.MethodPost,
		"/workspaces/workspace-123/projects",
		`{"name":""}`,
		"user-123",
	)

	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestProjectHandlerCreateUnauthorized(t *testing.T) {
	projectService := &fakeProjectService{
		err: service.ErrProjectUnauthorized,
	}

	handler := NewProjectHandler(projectService)

	req := authenticatedRequest(
		http.MethodPost,
		"/workspaces/workspace-123/projects",
		`{"name":"DevFlow"}`,
		"user-123",
	)

	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusForbidden,
			rec.Code,
		)
	}
}

func TestProjectHandlerListByWorkspaceID(t *testing.T) {
	now := time.Now()

	projectService := &fakeProjectService{
		projects: []*domain.Project{
			{
				ID:          "project-123",
				WorkspaceID: "workspace-123",
				Name:        "DevFlow",
				CreatedBy:   "user-123",
				CreatedAt:   now,
				UpdatedAt:   now,
			},
		},
	}

	handler := NewProjectHandler(projectService)

	req := authenticatedRequest(
		http.MethodGet,
		"/workspaces/workspace-123/projects",
		"",
		"user-123",
	)

	rec := httptest.NewRecorder()

	handler.ListByWorkspaceID(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}

	if projectService.gotRequesterID != "user-123" {
		t.Fatalf(
			"expected requester ID %q, got %q",
			"user-123",
			projectService.gotRequesterID,
		)
	}

	if projectService.gotWorkspaceID != "workspace-123" {
		t.Fatalf(
			"expected workspace ID %q, got %q",
			"workspace-123",
			projectService.gotWorkspaceID,
		)
	}
}

func TestProjectHandlerListRequiresAuth(t *testing.T) {
	projectService := &fakeProjectService{}

	handler := NewProjectHandler(projectService)

	req := httptest.NewRequest(
		http.MethodGet,
		"/workspaces/workspace-123/projects",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ListByWorkspaceID(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			rec.Code,
		)
	}
}

func TestProjectHandlerListUnauthorized(t *testing.T) {
	projectService := &fakeProjectService{
		err: service.ErrProjectUnauthorized,
	}

	handler := NewProjectHandler(projectService)

	req := authenticatedRequest(
		http.MethodGet,
		"/workspaces/workspace-123/projects",
		"",
		"user-123",
	)

	rec := httptest.NewRecorder()

	handler.ListByWorkspaceID(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusForbidden,
			rec.Code,
		)
	}
}

func TestProjectHandlerGetByID(t *testing.T) {
	now := time.Now()

	projectService := &fakeProjectService{
		project: &domain.Project{
			ID:          "project-123",
			WorkspaceID: "workspace-123",
			Name:        "DevFlow",
			CreatedBy:   "user-123",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	handler := NewProjectHandler(projectService)

	req := authenticatedRequest(
		http.MethodGet,
		"/projects/project-123",
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

	if projectService.gotRequesterID != "user-123" {
		t.Fatalf(
			"expected requester ID %q, got %q",
			"user-123",
			projectService.gotRequesterID,
		)
	}

	if projectService.gotProjectID != "project-123" {
		t.Fatalf(
			"expected project ID %q, got %q",
			"project-123",
			projectService.gotProjectID,
		)
	}
}

func TestProjectHandlerGetByIDRequiresAuth(t *testing.T) {
	projectService := &fakeProjectService{}

	handler := NewProjectHandler(projectService)

	req := httptest.NewRequest(
		http.MethodGet,
		"/projects/project-123",
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

func TestProjectHandlerGetByIDNotFound(t *testing.T) {
	projectService := &fakeProjectService{
		err: repository.ErrProjectNotFound,
	}

	handler := NewProjectHandler(projectService)

	req := authenticatedRequest(
		http.MethodGet,
		"/projects/project-123",
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

func TestProjectHandlerGetByIDUnauthorized(t *testing.T) {
	projectService := &fakeProjectService{
		err: service.ErrProjectUnauthorized,
	}

	handler := NewProjectHandler(projectService)

	req := authenticatedRequest(
		http.MethodGet,
		"/projects/project-123",
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

func TestProjectHandlerDelete(t *testing.T) {
	projectService := &fakeProjectService{}

	handler := NewProjectHandler(projectService)

	req := authenticatedRequest(
		http.MethodDelete,
		"/projects/project-123",
		"",
		"user-123",
	)

	rec := httptest.NewRecorder()

	handler.Delete(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNoContent,
			rec.Code,
		)
	}

	if projectService.gotRequesterID != "user-123" {
		t.Fatalf(
			"expected requester ID %q, got %q",
			"user-123",
			projectService.gotRequesterID,
		)
	}

	if projectService.gotProjectID != "project-123" {
		t.Fatalf(
			"expected project ID %q, got %q",
			"project-123",
			projectService.gotProjectID,
		)
	}
}

func TestProjectHandlerDeleteRequiresAuth(t *testing.T) {
	projectService := &fakeProjectService{}

	handler := NewProjectHandler(projectService)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/projects/project-123",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.Delete(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			rec.Code,
		)
	}
}

func TestProjectHandlerDeleteNotFound(t *testing.T) {
	projectService := &fakeProjectService{
		err: repository.ErrProjectNotFound,
	}

	handler := NewProjectHandler(projectService)

	req := authenticatedRequest(
		http.MethodDelete,
		"/projects/project-123",
		"",
		"user-123",
	)

	rec := httptest.NewRecorder()

	handler.Delete(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNotFound,
			rec.Code,
		)
	}
}

func TestProjectHandlerDeleteUnauthorized(t *testing.T) {
	projectService := &fakeProjectService{
		err: service.ErrProjectUnauthorized,
	}

	handler := NewProjectHandler(projectService)

	req := authenticatedRequest(
		http.MethodDelete,
		"/projects/project-123",
		"",
		"user-123",
	)

	rec := httptest.NewRecorder()

	handler.Delete(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusForbidden,
			rec.Code,
		)
	}
}
