package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kristopher1027/devflow-ai/internal/domain"
	"github.com/kristopher1027/devflow-ai/internal/service"
)

type fakeGitHubConnectionService struct {
	connection *domain.GitHubConnection
	err        error

	gotRequesterID    string
	gotWorkspaceID    string
	gotConnectionID   string
	gotInstallationID string
	gotAccountLogin   string
	gotStatus         string
}

func (f *fakeGitHubConnectionService) Create(
	ctx context.Context,
	requesterID string,
	workspaceID string,
	installationID string,
	accountLogin string,
) (*domain.GitHubConnection, error) {
	f.gotRequesterID = requesterID
	f.gotWorkspaceID = workspaceID
	f.gotInstallationID = installationID
	f.gotAccountLogin = accountLogin

	if f.err != nil {
		return nil, f.err
	}

	return f.connection, nil
}

func (f *fakeGitHubConnectionService) FindByWorkspaceID(
	ctx context.Context,
	requesterID string,
	workspaceID string,
) (*domain.GitHubConnection, error) {
	f.gotRequesterID = requesterID
	f.gotWorkspaceID = workspaceID

	if f.err != nil {
		return nil, f.err
	}

	return f.connection, nil
}

func (f *fakeGitHubConnectionService) UpdateStatus(
	ctx context.Context,
	requesterID string,
	id string,
	status string,
) error {
	f.gotRequesterID = requesterID
	f.gotConnectionID = id
	f.gotStatus = status

	return f.err
}

func TestGitHubConnectionHandlerCreate(t *testing.T) {
	connectionService := &fakeGitHubConnectionService{
		connection: &domain.GitHubConnection{
			ID:          "connection-123",
			WorkspaceID: "workspace-123",
			Status:      domain.GitHubConnectionStatusPending,
		},
	}
	handler := NewGitHubConnectionHandler(connectionService)
	req := authenticatedRequest(
		http.MethodPost,
		"/workspaces/workspace-123/github",
		`{"installation_id":"installation-123","account_login":"devflow-org"}`,
		"user-123",
	)
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	if connectionService.gotRequesterID != "user-123" {
		t.Fatalf("expected requester ID user-123, got %s", connectionService.gotRequesterID)
	}

	if connectionService.gotWorkspaceID != "workspace-123" {
		t.Fatalf("expected workspace ID workspace-123, got %s", connectionService.gotWorkspaceID)
	}

	if connectionService.gotInstallationID != "installation-123" {
		t.Fatalf("expected installation ID, got %s", connectionService.gotInstallationID)
	}
}

func TestGitHubConnectionHandlerRequiresAuth(t *testing.T) {
	handler := NewGitHubConnectionHandler(&fakeGitHubConnectionService{})
	req := httptest.NewRequest(
		http.MethodGet,
		"/workspaces/workspace-123/github",
		nil,
	)
	rec := httptest.NewRecorder()

	handler.GetByWorkspaceID(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestGitHubConnectionHandlerGetByWorkspaceID(t *testing.T) {
	connectionService := &fakeGitHubConnectionService{
		connection: &domain.GitHubConnection{
			ID:          "connection-123",
			WorkspaceID: "workspace-123",
			Status:      domain.GitHubConnectionStatusActive,
		},
	}
	handler := NewGitHubConnectionHandler(connectionService)
	req := authenticatedRequest(
		http.MethodGet,
		"/workspaces/workspace-123/github",
		"",
		"user-123",
	)
	rec := httptest.NewRecorder()

	handler.GetByWorkspaceID(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if connectionService.gotWorkspaceID != "workspace-123" {
		t.Fatalf("expected workspace ID workspace-123, got %s", connectionService.gotWorkspaceID)
	}
}

func TestGitHubConnectionHandlerUpdateStatus(t *testing.T) {
	connectionService := &fakeGitHubConnectionService{}
	handler := NewGitHubConnectionHandler(connectionService)
	req := authenticatedRequest(
		http.MethodPatch,
		"/workspaces/workspace-123/github",
		`{"id":"connection-123","status":"active"}`,
		"user-123",
	)
	rec := httptest.NewRecorder()

	handler.UpdateStatus(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rec.Code)
	}

	if connectionService.gotConnectionID != "connection-123" {
		t.Fatalf("expected connection ID connection-123, got %s", connectionService.gotConnectionID)
	}

	if connectionService.gotStatus != "active" {
		t.Fatalf("expected status active, got %s", connectionService.gotStatus)
	}
}

func TestGitHubConnectionHandlerCreateValidationError(t *testing.T) {
	handler := NewGitHubConnectionHandler(&fakeGitHubConnectionService{
		err: domain.ErrGitHubConnectionInstallationRequired,
	})
	req := authenticatedRequest(
		http.MethodPost,
		"/workspaces/workspace-123/github",
		`{"account_login":"devflow-org"}`,
		"user-123",
	)
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestGitHubConnectionHandlerUnauthorized(t *testing.T) {
	handler := NewGitHubConnectionHandler(&fakeGitHubConnectionService{
		err: service.ErrGitHubConnectionUnauthorized,
	})
	req := authenticatedRequest(
		http.MethodGet,
		"/workspaces/workspace-123/github",
		"",
		"user-123",
	)
	rec := httptest.NewRecorder()

	handler.GetByWorkspaceID(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
}

func TestGitHubConnectionHandlerInvalidTransition(t *testing.T) {
	handler := NewGitHubConnectionHandler(&fakeGitHubConnectionService{
		err: service.ErrGitHubConnectionStatusTransitionInvalid,
	})
	req := authenticatedRequest(
		http.MethodPatch,
		"/workspaces/workspace-123/github",
		`{"id":"connection-123","status":"pending"}`,
		"user-123",
	)
	rec := httptest.NewRecorder()

	handler.UpdateStatus(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestGitHubConnectionHandlerUnexpectedError(t *testing.T) {
	handler := NewGitHubConnectionHandler(&fakeGitHubConnectionService{
		err: errors.New("database failure"),
	})
	req := authenticatedRequest(
		http.MethodGet,
		"/workspaces/workspace-123/github",
		"",
		"user-123",
	)
	rec := httptest.NewRecorder()

	handler.GetByWorkspaceID(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}
