package service

import (
	"context"
	"errors"
	"testing"

	"github.com/kristopher1027/devflow-ai/internal/domain"
	"github.com/kristopher1027/devflow-ai/internal/repository"
)

type fakeGitHubConnectionRepository struct {
	connection  *domain.GitHubConnection
	findErr     error
	findByIDErr error
	createErr   error
	updateErr   error

	gotWorkspaceID    string
	gotInstallationID string
	gotConnectionID   string
	gotStatus         string
}

func (f *fakeGitHubConnectionRepository) Create(
	ctx context.Context,
	connection *domain.GitHubConnection,
) error {
	if f.createErr != nil {
		return f.createErr
	}

	f.connection = connection
	return nil
}

func (f *fakeGitHubConnectionRepository) FindByWorkspaceID(
	ctx context.Context,
	workspaceID string,
) (*domain.GitHubConnection, error) {
	f.gotWorkspaceID = workspaceID
	return f.connection, f.findErr
}

func (f *fakeGitHubConnectionRepository) FindByID(
	ctx context.Context,
	id string,
) (*domain.GitHubConnection, error) {
	f.gotConnectionID = id
	if f.findByIDErr != nil {
		return nil, f.findByIDErr
	}
	return f.connection, nil
}

func (f *fakeGitHubConnectionRepository) FindByInstallationID(
	ctx context.Context,
	installationID string,
) (*domain.GitHubConnection, error) {
	f.gotInstallationID = installationID
	return f.connection, f.findErr
}

func (f *fakeGitHubConnectionRepository) UpdateStatus(
	ctx context.Context,
	id string,
	status string,
) error {
	f.gotConnectionID = id
	f.gotStatus = status
	return f.updateErr
}

func TestGitHubConnectionServiceCreate(t *testing.T) {
	connectionRepo := &fakeGitHubConnectionRepository{
		findErr: repository.ErrGitHubConnectionNotFound,
	}
	workspaceRepo := &fakeWorkspaceRepository{
		workspace: testWorkspace(),
	}
	memberRepo := &fakeWorkspaceMemberRepository{}
	service := NewGitHubConnectionService(
		connectionRepo,
		workspaceRepo,
		memberRepo,
	)

	connection, err := service.Create(
		context.Background(),
		"owner-123",
		"workspace-123",
		" installation-123 ",
		" devflow-org ",
	)

	if err != nil {
		t.Fatalf("create github connection: %v", err)
	}

	if connection.ID == "" {
		t.Fatal("expected generated connection ID")
	}

	if connection.WorkspaceID != "workspace-123" {
		t.Fatalf("expected workspace ID workspace-123, got %s", connection.WorkspaceID)
	}

	if connection.InstallationID != "installation-123" {
		t.Fatalf("expected trimmed installation ID, got %s", connection.InstallationID)
	}

	if connection.AccountLogin != "devflow-org" {
		t.Fatalf("expected trimmed account login, got %s", connection.AccountLogin)
	}

	if connection.Status != domain.GitHubConnectionStatusPending {
		t.Fatalf("expected pending status, got %s", connection.Status)
	}

	if connectionRepo.connection != connection {
		t.Fatal("expected connection passed to repository layer")
	}
}

func TestGitHubConnectionServiceCreateUnauthorizedMember(t *testing.T) {
	connectionRepo := &fakeGitHubConnectionRepository{
		findErr: repository.ErrGitHubConnectionNotFound,
	}
	workspaceRepo := &fakeWorkspaceRepository{
		workspace: testWorkspace(),
	}
	memberRepo := &fakeWorkspaceMemberRepository{
		member: &domain.WorkspaceMember{
			WorkspaceID: "workspace-123",
			UserID:      "member-123",
			Role:        WorkspaceMemberRoleMember,
		},
	}
	service := NewGitHubConnectionService(
		connectionRepo,
		workspaceRepo,
		memberRepo,
	)

	_, err := service.Create(
		context.Background(),
		"member-123",
		"workspace-123",
		"installation-123",
		"devflow-org",
	)

	if !errors.Is(err, ErrGitHubConnectionUnauthorized) {
		t.Fatalf("expected unauthorized error, got %v", err)
	}
}

func TestGitHubConnectionServiceCreateDuplicate(t *testing.T) {
	connectionRepo := &fakeGitHubConnectionRepository{
		connection: &domain.GitHubConnection{ID: "connection-123"},
	}
	workspaceRepo := &fakeWorkspaceRepository{
		workspace: testWorkspace(),
	}
	service := NewGitHubConnectionService(
		connectionRepo,
		workspaceRepo,
		&fakeWorkspaceMemberRepository{},
	)

	_, err := service.Create(
		context.Background(),
		"owner-123",
		"workspace-123",
		"installation-123",
		"devflow-org",
	)

	if !errors.Is(err, ErrGitHubConnectionAlreadyExists) {
		t.Fatalf("expected duplicate error, got %v", err)
	}
}

func TestGitHubConnectionServiceFindByWorkspaceIDMember(t *testing.T) {
	connection := &domain.GitHubConnection{
		ID:          "connection-123",
		WorkspaceID: "workspace-123",
		Status:      domain.GitHubConnectionStatusActive,
	}
	connectionRepo := &fakeGitHubConnectionRepository{
		connection: connection,
	}
	workspaceRepo := &fakeWorkspaceRepository{
		workspace: testWorkspace(),
	}
	memberRepo := &fakeWorkspaceMemberRepository{
		member: &domain.WorkspaceMember{
			WorkspaceID: "workspace-123",
			UserID:      "member-123",
			Role:        WorkspaceMemberRoleMember,
		},
	}
	service := NewGitHubConnectionService(
		connectionRepo,
		workspaceRepo,
		memberRepo,
	)

	result, err := service.FindByWorkspaceID(
		context.Background(),
		"member-123",
		"workspace-123",
	)

	if err != nil {
		t.Fatalf("find github connection: %v", err)
	}

	if result != connection {
		t.Fatal("expected connection returned from repository layer")
	}

	if connectionRepo.gotWorkspaceID != "workspace-123" {
		t.Fatalf("expected workspace lookup, got %s", connectionRepo.gotWorkspaceID)
	}
}

func TestGitHubConnectionServiceUpdateStatus(t *testing.T) {
	connection := &domain.GitHubConnection{
		ID:          "connection-123",
		WorkspaceID: "workspace-123",
		Status:      domain.GitHubConnectionStatusPending,
	}
	connectionRepo := &fakeGitHubConnectionRepository{
		connection: connection,
	}
	service := NewGitHubConnectionService(
		connectionRepo,
		&fakeWorkspaceRepository{workspace: testWorkspace()},
		&fakeWorkspaceMemberRepository{},
	)

	err := service.UpdateStatus(
		context.Background(),
		"owner-123",
		"connection-123",
		domain.GitHubConnectionStatusActive,
	)

	if err != nil {
		t.Fatalf("update github connection status: %v", err)
	}

	if connectionRepo.gotConnectionID != "connection-123" {
		t.Fatalf("expected connection ID, got %s", connectionRepo.gotConnectionID)
	}

	if connectionRepo.gotStatus != domain.GitHubConnectionStatusActive {
		t.Fatalf("expected active status, got %s", connectionRepo.gotStatus)
	}
}

func TestGitHubConnectionServiceUpdateStatusInvalidTransition(t *testing.T) {
	connectionRepo := &fakeGitHubConnectionRepository{
		connection: &domain.GitHubConnection{
			ID:          "connection-123",
			WorkspaceID: "workspace-123",
			Status:      domain.GitHubConnectionStatusActive,
		},
	}
	service := NewGitHubConnectionService(
		connectionRepo,
		&fakeWorkspaceRepository{workspace: testWorkspace()},
		&fakeWorkspaceMemberRepository{},
	)

	err := service.UpdateStatus(
		context.Background(),
		"owner-123",
		"connection-123",
		domain.GitHubConnectionStatusPending,
	)

	if !errors.Is(err, ErrGitHubConnectionStatusTransitionInvalid) {
		t.Fatalf("expected invalid transition error, got %v", err)
	}

	if connectionRepo.gotStatus != "" {
		t.Fatal("expected repository status update not to be called")
	}
}

func TestGitHubConnectionServiceUpdateStatusUnauthorized(t *testing.T) {
	connectionRepo := &fakeGitHubConnectionRepository{
		connection: &domain.GitHubConnection{
			ID:          "connection-123",
			WorkspaceID: "workspace-123",
			Status:      domain.GitHubConnectionStatusPending,
		},
	}
	service := NewGitHubConnectionService(
		connectionRepo,
		&fakeWorkspaceRepository{workspace: testWorkspace()},
		&fakeWorkspaceMemberRepository{
			member: &domain.WorkspaceMember{
				WorkspaceID: "workspace-123",
				UserID:      "member-123",
				Role:        WorkspaceMemberRoleMember,
			},
		},
	)

	err := service.UpdateStatus(
		context.Background(),
		"member-123",
		"connection-123",
		domain.GitHubConnectionStatusActive,
	)

	if !errors.Is(err, ErrGitHubConnectionUnauthorized) {
		t.Fatalf("expected unauthorized error, got %v", err)
	}
}
