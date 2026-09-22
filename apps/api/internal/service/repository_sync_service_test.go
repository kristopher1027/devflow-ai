package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kristopher1027/devflow-ai/internal/domain"
	"github.com/kristopher1027/devflow-ai/internal/repository"
)

type fakeRepositorySyncRepository struct {
	repository      *domain.Repository
	err             error
	updateCalls     int
	updatedStatus   string
	updatedSyncedAt *time.Time
	updateErr       error
}

func (f *fakeRepositorySyncRepository) Create(
	ctx context.Context,
	repo *domain.Repository,
) error {
	return nil
}

func (f *fakeRepositorySyncRepository) FindByID(
	ctx context.Context,
	id string,
) (*domain.Repository, error) {
	return f.repository, f.err
}

func (f *fakeRepositorySyncRepository) ListByProjectID(
	ctx context.Context,
	projectID string,
) ([]*domain.Repository, error) {
	return nil, nil
}

func (f *fakeRepositorySyncRepository) FindByProviderExternalID(
	ctx context.Context,
	provider string,
	externalID string,
) (*domain.Repository, error) {
	return nil, nil
}

func (f *fakeRepositorySyncRepository) UpdateSyncStatus(
	ctx context.Context,
	id string,
	status string,
	lastSyncedAt *time.Time,
) error {
	f.updateCalls++
	f.updatedStatus = status
	f.updatedSyncedAt = lastSyncedAt

	return f.updateErr
}

func (f *fakeRepositorySyncRepository) Delete(
	ctx context.Context,
	id string,
) error {
	return nil
}

type fakeRepositorySnapshotRepository struct {
	findCalls       int
	createCalls     int
	snapshot        *domain.RepositorySnapshot
	findErr         error
	createErr       error
	createdSnapshot *domain.RepositorySnapshot
}

func (f *fakeRepositorySnapshotRepository) Create(
	ctx context.Context,
	snapshot *domain.RepositorySnapshot,
) error {
	f.createCalls++
	f.createdSnapshot = snapshot

	return f.createErr
}

func (f *fakeRepositorySnapshotRepository) FindByID(
	ctx context.Context,
	id string,
) (*domain.RepositorySnapshot, error) {
	f.findCalls++

	return f.snapshot, f.findErr
}

func (f *fakeRepositorySnapshotRepository) ListByRepositoryID(
	ctx context.Context,
	repositoryID string,
) ([]*domain.RepositorySnapshot, error) {
	return nil, nil
}

func (f *fakeRepositorySnapshotRepository) FindByRepositoryIDAndCommitSHA(
	ctx context.Context,
	repositoryID string,
	commitSHA string,
) (*domain.RepositorySnapshot, error) {
	f.findCalls++

	return f.snapshot, f.findErr
}

type fakeRepositoryClient struct {
	calls          int
	installationID string
	owner          string
	repositoryName string
	branch         string
	commitSHA      string
	err            error
}

func (f *fakeRepositoryClient) GetLatestCommitSHA(
	ctx context.Context,
	installationID string,
	owner string,
	repositoryName string,
	branch string,
) (string, error) {
	f.calls++
	f.installationID = installationID
	f.owner = owner
	f.repositoryName = repositoryName
	f.branch = branch

	return f.commitSHA, f.err
}

func TestRepositorySyncServiceRepositoryNotFound(t *testing.T) {
	repositoryRepo := &fakeRepositorySyncRepository{
		err: repository.ErrRepositoryNotFound,
	}

	snapshotRepo := &fakeRepositorySnapshotRepository{}

	githubClient := &fakeRepositoryClient{}

	service := NewRepositorySyncService(
		repositoryRepo,
		nil,
		nil,
		snapshotRepo,
		githubClient,
	)

	_, err := service.Sync(
		context.Background(),
		"missing-repository-id",
	)

	if !errors.Is(err, ErrRepositorySyncRepositoryNotFound) {
		t.Fatalf(
			"expected ErrRepositorySyncRepositoryNotFound, got %v",
			err,
		)
	}

	if githubClient.calls != 0 {
		t.Fatalf(
			"expected GitHub client not to be called, got %d calls",
			githubClient.calls,
		)
	}

	if snapshotRepo.findCalls != 0 {
		t.Fatalf(
			"expected snapshot repository not to be called, got %d calls",
			snapshotRepo.findCalls,
		)
	}

	if snapshotRepo.createCalls != 0 {
		t.Fatalf(
			"expected snapshot creation not to be called, got %d calls",
			snapshotRepo.createCalls,
		)
	}
}

func TestRepositorySyncServicePropagatesSnapshotCreateError(t *testing.T) {
	ctx := context.Background()

	expectedErr := errors.New("snapshot create failed")

	repo := &domain.Repository{
		ID:            "repository-1",
		ProjectID:     "project-1",
		Owner:         "octocat",
		Name:          "hello-world",
		DefaultBranch: "main",
	}

	repositoryRepo := &fakeRepositorySyncRepository{
		repository: repo,
	}

	projectRepo := &fakeProjectRepository{
		project: &domain.Project{
			ID:          "project-1",
			WorkspaceID: "workspace-1",
		},
	}

	githubConnectionRepo := &fakeGitHubConnectionRepository{
		connection: &domain.GitHubConnection{
			ID:             "connection-1",
			WorkspaceID:    "workspace-1",
			InstallationID: "installation-123",
			AccountLogin:   "devflow-test",
			Status:         domain.GitHubConnectionStatusActive,
		},
	}

	snapshotRepo := &fakeRepositorySnapshotRepository{
		findErr:   repository.ErrRepositorySnapshotNotFound,
		createErr: expectedErr,
	}

	githubClient := &fakeRepositoryClient{
		commitSHA: "abc123",
	}

	service := NewRepositorySyncService(
		repositoryRepo,
		projectRepo,
		githubConnectionRepo,
		snapshotRepo,
		githubClient,
	)

	snapshot, err := service.Sync(ctx, repo.ID)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected snapshot create error, got %v",
			err,
		)
	}

	if snapshot != nil {
		t.Fatalf("expected nil snapshot, got %#v", snapshot)
	}

	if snapshotRepo.createCalls != 1 {
		t.Fatalf(
			"expected 1 snapshot create call, got %d",
			snapshotRepo.createCalls,
		)
	}
}

func TestRepositorySyncServicePropagatesRepositoryError(t *testing.T) {
	expectedErr := errors.New("database unavailable")

	repositoryRepo := &fakeRepositorySyncRepository{
		err: expectedErr,
	}

	snapshotRepo := &fakeRepositorySnapshotRepository{}
	githubClient := &fakeRepositoryClient{}

	service := NewRepositorySyncService(
		repositoryRepo,
		nil,
		nil,
		snapshotRepo,
		githubClient,
	)

	_, err := service.Sync(
		context.Background(),
		"repository-id",
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected original repository error, got %v",
			err,
		)
	}

	if errors.Is(err, ErrRepositorySyncRepositoryNotFound) {
		t.Fatal("unexpected ErrRepositorySyncRepositoryNotFound")
	}

	if githubClient.calls != 0 {
		t.Fatalf(
			"expected GitHub client not to be called, got %d calls",
			githubClient.calls,
		)
	}
}

func TestRepositorySyncServiceRequiresDefaultBranch(t *testing.T) {
	repositoryRepo := &fakeRepositorySyncRepository{
		repository: &domain.Repository{
			ID:            "repository-id",
			ProjectID:     "project-1",
			Owner:         "devflow-test",
			Name:          "sync-test-repository",
			DefaultBranch: "",
		},
	}

	snapshotRepo := &fakeRepositorySnapshotRepository{}
	githubClient := &fakeRepositoryClient{}

	service := NewRepositorySyncService(
		repositoryRepo,
		nil,
		nil,
		snapshotRepo,
		githubClient,
	)

	_, err := service.Sync(
		context.Background(),
		"repository-id",
	)

	if !errors.Is(err, ErrRepositorySyncBranchRequired) {
		t.Fatalf(
			"expected ErrRepositorySyncBranchRequired, got %v",
			err,
		)
	}

	if githubClient.calls != 0 {
		t.Fatalf(
			"expected GitHub client not to be called, got %d calls",
			githubClient.calls,
		)
	}

	if snapshotRepo.findCalls != 0 {
		t.Fatalf(
			"expected snapshot repository not to be called, got %d calls",
			snapshotRepo.findCalls,
		)
	}
}

func TestRepositorySyncServiceRequiresOwner(t *testing.T) {
	repositoryRepo := &fakeRepositorySyncRepository{
		repository: &domain.Repository{
			ID:            "repository-id",
			ProjectID:     "project-1",
			Owner:         "",
			Name:          "sync-test-repository",
			DefaultBranch: "main",
		},
	}

	snapshotRepo := &fakeRepositorySnapshotRepository{}
	githubClient := &fakeRepositoryClient{}

	service := NewRepositorySyncService(
		repositoryRepo,
		nil,
		nil,
		snapshotRepo,
		githubClient,
	)

	_, err := service.Sync(
		context.Background(),
		"repository-id",
	)

	if !errors.Is(err, ErrRepositorySyncOwnerRequired) {
		t.Fatalf(
			"expected ErrRepositorySyncOwnerRequired, got %v",
			err,
		)
	}

	if githubClient.calls != 0 {
		t.Fatalf(
			"expected GitHub client not to be called, got %d calls",
			githubClient.calls,
		)
	}
}

func TestRepositorySyncServiceRequiresName(t *testing.T) {
	repositoryRepo := &fakeRepositorySyncRepository{
		repository: &domain.Repository{
			ID:            "repository-id",
			ProjectID:     "project-1",
			Owner:         "devflow-test",
			Name:          "",
			DefaultBranch: "main",
		},
	}

	snapshotRepo := &fakeRepositorySnapshotRepository{}
	githubClient := &fakeRepositoryClient{}

	service := NewRepositorySyncService(
		repositoryRepo,
		nil,
		nil,
		snapshotRepo,
		githubClient,
	)

	_, err := service.Sync(
		context.Background(),
		"repository-id",
	)

	if !errors.Is(err, ErrRepositorySyncNameRequired) {
		t.Fatalf(
			"expected ErrRepositorySyncNameRequired, got %v",
			err,
		)
	}

	if githubClient.calls != 0 {
		t.Fatalf(
			"expected GitHub client not to be called, got %d calls",
			githubClient.calls,
		)
	}
}

func TestRepositorySyncServiceGetsLatestCommit(t *testing.T) {
	repo := &domain.Repository{
		ID:            "repository-id",
		ProjectID:     "project-1",
		Owner:         "devflow-test",
		Name:          "sync-test-repository",
		DefaultBranch: "main",
	}

	repositoryRepo := &fakeRepositorySyncRepository{
		repository: repo,
	}

	projectRepo := &fakeProjectRepository{
		project: &domain.Project{
			ID:          "project-1",
			WorkspaceID: "workspace-1",
		},
	}

	githubConnectionRepo := &fakeGitHubConnectionRepository{
		connection: &domain.GitHubConnection{
			ID:             "connection-1",
			WorkspaceID:    "workspace-1",
			InstallationID: "installation-123",
			AccountLogin:   "devflow-test",
			Status:         domain.GitHubConnectionStatusActive,
		},
	}

	snapshotRepo := &fakeRepositorySnapshotRepository{}

	githubClient := &fakeRepositoryClient{
		commitSHA: "abc123",
	}

	service := NewRepositorySyncService(
		repositoryRepo,
		projectRepo,
		githubConnectionRepo,
		snapshotRepo,
		githubClient,
	)

	_, err := service.Sync(
		context.Background(),
		repo.ID,
	)

	if err != nil {
		t.Fatalf("sync failed: %v", err)
	}

	if githubClient.calls != 1 {
		t.Fatalf(
			"expected GitHub client to be called once, got %d calls",
			githubClient.calls,
		)
	}

	if githubClient.installationID != "installation-123" {
		t.Fatalf(
			"expected installation ID %q, got %q",
			"installation-123",
			githubClient.installationID,
		)
	}

	if githubClient.owner != "devflow-test" {
		t.Fatalf(
			"expected owner %q, got %q",
			"devflow-test",
			githubClient.owner,
		)
	}

	if githubClient.repositoryName != "sync-test-repository" {
		t.Fatalf(
			"expected repository name %q, got %q",
			"sync-test-repository",
			githubClient.repositoryName,
		)
	}

	if githubClient.branch != "main" {
		t.Fatalf(
			"expected branch %q, got %q",
			"main",
			githubClient.branch,
		)
	}
}

func TestRepositorySyncServicePropagatesGitHubError(t *testing.T) {
	expectedErr := errors.New("github unavailable")

	repo := &domain.Repository{
		ID:            "repository-id",
		ProjectID:     "project-1",
		Owner:         "devflow-test",
		Name:          "sync-test-repository",
		DefaultBranch: "main",
	}

	repositoryRepo := &fakeRepositorySyncRepository{
		repository: repo,
	}

	projectRepo := &fakeProjectRepository{
		project: &domain.Project{
			ID:          "project-1",
			WorkspaceID: "workspace-1",
		},
	}

	githubConnectionRepo := &fakeGitHubConnectionRepository{
		connection: &domain.GitHubConnection{
			ID:             "connection-1",
			WorkspaceID:    "workspace-1",
			InstallationID: "installation-123",
			AccountLogin:   "devflow-test",
			Status:         domain.GitHubConnectionStatusActive,
		},
	}

	snapshotRepo := &fakeRepositorySnapshotRepository{}

	githubClient := &fakeRepositoryClient{
		err: expectedErr,
	}

	service := NewRepositorySyncService(
		repositoryRepo,
		projectRepo,
		githubConnectionRepo,
		snapshotRepo,
		githubClient,
	)

	_, err := service.Sync(
		context.Background(),
		repo.ID,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected original GitHub error, got %v",
			err,
		)
	}

	if snapshotRepo.findCalls != 0 {
		t.Fatalf(
			"expected snapshot repository not to be called, got %d calls",
			snapshotRepo.findCalls,
		)
	}

	if snapshotRepo.createCalls != 0 {
		t.Fatalf(
			"expected snapshot creation not to be called, got %d calls",
			snapshotRepo.createCalls,
		)
	}
}

func TestRepositorySyncServiceReturnsExistingSnapshot(t *testing.T) {
	existingSnapshot := &domain.RepositorySnapshot{
		ID:           "snapshot-id",
		RepositoryID: "repository-id",
		CommitSHA:    "abc123",
		Branch:       "main",
	}

	repo := &domain.Repository{
		ID:            "repository-id",
		ProjectID:     "project-1",
		Owner:         "devflow-test",
		Name:          "sync-test-repository",
		DefaultBranch: "main",
	}

	repositoryRepo := &fakeRepositorySyncRepository{
		repository: repo,
	}

	projectRepo := &fakeProjectRepository{
		project: &domain.Project{
			ID:          "project-1",
			WorkspaceID: "workspace-1",
		},
	}

	githubConnectionRepo := &fakeGitHubConnectionRepository{
		connection: &domain.GitHubConnection{
			ID:             "connection-1",
			WorkspaceID:    "workspace-1",
			InstallationID: "installation-123",
			AccountLogin:   "devflow-test",
			Status:         domain.GitHubConnectionStatusActive,
		},
	}

	snapshotRepo := &fakeRepositorySnapshotRepository{
		snapshot: existingSnapshot,
	}

	githubClient := &fakeRepositoryClient{
		commitSHA: "abc123",
	}

	service := NewRepositorySyncService(
		repositoryRepo,
		projectRepo,
		githubConnectionRepo,
		snapshotRepo,
		githubClient,
	)

	result, err := service.Sync(
		context.Background(),
		repo.ID,
	)

	if err != nil {
		t.Fatalf("sync failed: %v", err)
	}

	if result != existingSnapshot {
		t.Fatal("expected existing snapshot to be returned")
	}

	if snapshotRepo.createCalls != 0 {
		t.Fatalf(
			"expected no snapshot creation, got %d calls",
			snapshotRepo.createCalls,
		)
	}

	if repositoryRepo.updateCalls != 0 {
		t.Fatalf(
			"expected repository status not to be updated, got %d calls",
			repositoryRepo.updateCalls,
		)
	}
}

func TestRepositorySyncServiceCreatesNewSnapshot(t *testing.T) {
	repo := &domain.Repository{
		ID:            "repository-id",
		ProjectID:     "project-1",
		Owner:         "devflow-test",
		Name:          "sync-test-repository",
		DefaultBranch: "main",
	}

	repositoryRepo := &fakeRepositorySyncRepository{
		repository: repo,
	}

	projectRepo := &fakeProjectRepository{
		project: &domain.Project{
			ID:          "project-1",
			WorkspaceID: "workspace-1",
		},
	}

	githubConnectionRepo := &fakeGitHubConnectionRepository{
		connection: &domain.GitHubConnection{
			ID:             "connection-1",
			WorkspaceID:    "workspace-1",
			InstallationID: "installation-123",
			AccountLogin:   "devflow-test",
			Status:         domain.GitHubConnectionStatusActive,
		},
	}

	snapshotRepo := &fakeRepositorySnapshotRepository{
		findErr: repository.ErrRepositorySnapshotNotFound,
	}

	githubClient := &fakeRepositoryClient{
		commitSHA: "abc123",
	}

	service := NewRepositorySyncService(
		repositoryRepo,
		projectRepo,
		githubConnectionRepo,
		snapshotRepo,
		githubClient,
	)

	result, err := service.Sync(
		context.Background(),
		repo.ID,
	)

	if err != nil {
		t.Fatalf("sync failed: %v", err)
	}

	if result == nil {
		t.Fatal("expected new snapshot, got nil")
	}

	if result.RepositoryID != repo.ID {
		t.Fatalf(
			"expected repository ID %q, got %q",
			repo.ID,
			result.RepositoryID,
		)
	}

	if result.CommitSHA != "abc123" {
		t.Fatalf(
			"expected commit SHA %q, got %q",
			"abc123",
			result.CommitSHA,
		)
	}

	if result.Branch != "main" {
		t.Fatalf(
			"expected branch %q, got %q",
			"main",
			result.Branch,
		)
	}

	if snapshotRepo.createCalls != 1 {
		t.Fatalf(
			"expected snapshot creation once, got %d calls",
			snapshotRepo.createCalls,
		)
	}

	if snapshotRepo.createdSnapshot == nil {
		t.Fatal("expected created snapshot to be captured")
	}
}

func TestRepositorySyncServiceUpdatesRepositoryStatusAfterSuccessfulSync(t *testing.T) {
	repo := &domain.Repository{
		ID:            "repository-id",
		ProjectID:     "project-1",
		Owner:         "devflow-test",
		Name:          "sync-test-repository",
		DefaultBranch: "main",
	}

	repositoryRepo := &fakeRepositorySyncRepository{
		repository: repo,
	}

	projectRepo := &fakeProjectRepository{
		project: &domain.Project{
			ID:          "project-1",
			WorkspaceID: "workspace-1",
		},
	}

	githubConnectionRepo := &fakeGitHubConnectionRepository{
		connection: &domain.GitHubConnection{
			ID:             "connection-1",
			WorkspaceID:    "workspace-1",
			InstallationID: "installation-123",
			AccountLogin:   "devflow-test",
			Status:         domain.GitHubConnectionStatusActive,
		},
	}

	snapshotRepo := &fakeRepositorySnapshotRepository{
		findErr: repository.ErrRepositorySnapshotNotFound,
	}

	githubClient := &fakeRepositoryClient{
		commitSHA: "abc123",
	}

	service := NewRepositorySyncService(
		repositoryRepo,
		projectRepo,
		githubConnectionRepo,
		snapshotRepo,
		githubClient,
	)

	_, err := service.Sync(
		context.Background(),
		repo.ID,
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if repositoryRepo.updateCalls != 1 {
		t.Fatalf(
			"expected UpdateSyncStatus to be called once, got %d",
			repositoryRepo.updateCalls,
		)
	}

	if repositoryRepo.updatedStatus != "synced" {
		t.Fatalf(
			"expected status %q, got %q",
			"synced",
			repositoryRepo.updatedStatus,
		)
	}

	if repositoryRepo.updatedSyncedAt == nil {
		t.Fatal("expected lastSyncedAt to be set")
	}
}

func TestRepositorySyncServicePropagatesUpdateSyncStatusError(t *testing.T) {
	updateErr := errors.New("update sync status failed")

	repo := &domain.Repository{
		ID:            "repository-id",
		ProjectID:     "project-1",
		Owner:         "devflow-test",
		Name:          "sync-test-repository",
		DefaultBranch: "main",
	}

	repositoryRepo := &fakeRepositorySyncRepository{
		repository: repo,
		updateErr:  updateErr,
	}

	projectRepo := &fakeProjectRepository{
		project: &domain.Project{
			ID:          "project-1",
			WorkspaceID: "workspace-1",
		},
	}

	githubConnectionRepo := &fakeGitHubConnectionRepository{
		connection: &domain.GitHubConnection{
			ID:             "connection-1",
			WorkspaceID:    "workspace-1",
			InstallationID: "installation-123",
			AccountLogin:   "devflow-test",
			Status:         domain.GitHubConnectionStatusActive,
		},
	}

	snapshotRepo := &fakeRepositorySnapshotRepository{
		findErr: repository.ErrRepositorySnapshotNotFound,
	}

	githubClient := &fakeRepositoryClient{
		commitSHA: "abc123",
	}

	service := NewRepositorySyncService(
		repositoryRepo,
		projectRepo,
		githubConnectionRepo,
		snapshotRepo,
		githubClient,
	)

	_, err := service.Sync(
		context.Background(),
		repo.ID,
	)

	if !errors.Is(err, updateErr) {
		t.Fatalf(
			"expected update sync status error, got %v",
			err,
		)
	}

	if repositoryRepo.updateCalls != 1 {
		t.Fatalf(
			"expected UpdateSyncStatus to be called once, got %d",
			repositoryRepo.updateCalls,
		)
	}
}
