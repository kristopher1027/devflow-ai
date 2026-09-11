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

type fakeWorkspaceService struct {
	workspace  *domain.Workspace
	workspaces []*domain.Workspace
	err        error

	gotOwnerID string
	gotName    string
	gotID      string
}

func (f *fakeWorkspaceService) Create(
	ctx context.Context,
	ownerID string,
	name string,
) (*domain.Workspace, error) {
	f.gotOwnerID = ownerID
	f.gotName = name

	if f.err != nil {
		return nil, f.err
	}

	return f.workspace, nil
}

func (f *fakeWorkspaceService) FindByID(
	ctx context.Context,
	ownerID string,
	id string,
) (*domain.Workspace, error) {
	f.gotOwnerID = ownerID
	f.gotID = id

	if f.err != nil {
		return nil, f.err
	}

	return f.workspace, nil
}

func (f *fakeWorkspaceService) ListByOwnerID(
	ctx context.Context,
	ownerID string,
) ([]*domain.Workspace, error) {
	f.gotOwnerID = ownerID

	if f.err != nil {
		return nil, f.err
	}

	return f.workspaces, nil
}

func (f *fakeWorkspaceService) Delete(
	ctx context.Context,
	ownerID string,
	id string,
) error {
	f.gotID = id

	return f.err
}

func authenticatedRequest(
	method string,
	target string,
	body string,
	userID string,
) *http.Request {
	req := httptest.NewRequest(
		method,
		target,
		strings.NewReader(body),
	)

	ctx := context.WithValue(
		req.Context(),
		userIDContextKey,
		userID,
	)

	return req.WithContext(ctx)
}

func TestWorkspaceHandlerCreate(t *testing.T) {
	now := time.Now()

	workspace := &domain.Workspace{
		ID:        "workspace-123",
		OwnerID:   "user-123",
		Name:      "My Workspace",
		CreatedAt: now,
		UpdatedAt: now,
	}

	service := &fakeWorkspaceService{
		workspace: workspace,
	}

	handler := NewWorkspaceHandler(service)

	req := authenticatedRequest(
		http.MethodPost,
		"/workspaces",
		`{"name":"My Workspace"}`,
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

	if service.gotOwnerID != "user-123" {
		t.Fatalf(
			"expected owner ID %q, got %q",
			"user-123",
			service.gotOwnerID,
		)
	}

	if service.gotName != "My Workspace" {
		t.Fatalf(
			"expected name %q, got %q",
			"My Workspace",
			service.gotName,
		)
	}
}

func TestWorkspaceHandlerCreateRequiresAuth(t *testing.T) {
	service := &fakeWorkspaceService{}

	handler := NewWorkspaceHandler(service)

	req := httptest.NewRequest(
		http.MethodPost,
		"/workspaces",
		strings.NewReader(`{"name":"My Workspace"}`),
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

func TestWorkspaceHandlerCreateInvalidBody(t *testing.T) {
	service := &fakeWorkspaceService{}

	handler := NewWorkspaceHandler(service)

	req := authenticatedRequest(
		http.MethodPost,
		"/workspaces",
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

func TestWorkspaceHandlerCreateNameRequired(t *testing.T) {
	service := &fakeWorkspaceService{
		err: service.ErrWorkspaceNameRequired,
	}

	handler := NewWorkspaceHandler(service)

	req := authenticatedRequest(
		http.MethodPost,
		"/workspaces",
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

func TestWorkspaceHandlerList(t *testing.T) {
	now := time.Now()

	workspaces := []*domain.Workspace{
		{
			ID:        "workspace-123",
			OwnerID:   "user-123",
			Name:      "My Workspace",
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	workspaceService := &fakeWorkspaceService{
		workspaces: workspaces,
	}

	handler := NewWorkspaceHandler(workspaceService)

	req := authenticatedRequest(
		http.MethodGet,
		"/workspaces",
		"",
		"user-123",
	)

	rec := httptest.NewRecorder()

	handler.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}

	if workspaceService.gotOwnerID != "user-123" {
		t.Fatalf(
			"expected owner ID %q, got %q",
			"user-123",
			workspaceService.gotOwnerID,
		)
	}
}

func TestWorkspaceHandlerGetByID(t *testing.T) {
	now := time.Now()

	workspaceService := &fakeWorkspaceService{
		workspace: &domain.Workspace{
			ID:        "workspace-123",
			OwnerID:   "user-123",
			Name:      "My Workspace",
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	handler := NewWorkspaceHandler(workspaceService)

	req := authenticatedRequest(
		http.MethodGet,
		"/workspaces/workspace-123",
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

	if workspaceService.gotID != "workspace-123" {
		t.Fatalf(
			"expected workspace ID %q, got %q",
			"workspace-123",
			workspaceService.gotID,
		)
	}
}

func TestWorkspaceHandlerGetByIDNotFound(t *testing.T) {
	workspaceService := &fakeWorkspaceService{
		err: repository.ErrWorkspaceNotFound,
	}

	handler := NewWorkspaceHandler(workspaceService)

	req := authenticatedRequest(
		http.MethodGet,
		"/workspaces/missing",
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

func TestWorkspaceHandlerDelete(t *testing.T) {
	workspaceService := &fakeWorkspaceService{}

	handler := NewWorkspaceHandler(workspaceService)

	req := authenticatedRequest(
		http.MethodDelete,
		"/workspaces/workspace-123",
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

	if workspaceService.gotID != "workspace-123" {
		t.Fatalf(
			"expected workspace ID %q, got %q",
			"workspace-123",
			workspaceService.gotID,
		)
	}
}

func TestWorkspaceHandlerDeleteNotFound(t *testing.T) {
	workspaceService := &fakeWorkspaceService{
		err: repository.ErrWorkspaceNotFound,
	}

	handler := NewWorkspaceHandler(workspaceService)

	req := authenticatedRequest(
		http.MethodDelete,
		"/workspaces/missing",
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

func TestWorkspaceHandlerInternalError(t *testing.T) {
	workspaceService := &fakeWorkspaceService{
		err: errors.New("database failure"),
	}

	handler := NewWorkspaceHandler(workspaceService)

	req := authenticatedRequest(
		http.MethodGet,
		"/workspaces/workspace-123",
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
