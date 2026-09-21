package service

import (
	"context"
	"errors"
	"testing"

	"github.com/kristopher1027/devflow-ai/internal/domain"
	"github.com/kristopher1027/devflow-ai/internal/repository"
)

type fakeRepositorySyncRepository struct {
	repository *domain.Repository
	err        error
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

func (f *fakeRepositorySyncRepository) Delete(
	ctx context.Context,
	id string,
) error {
	return nil
}

type fakeRepositorySnapshotRepository struct {
	findCalls      int
	createCalls    int
	snapshot       *domain.RepositorySnapshot
	findErr        error
	createErr      error
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
type fakeRepositoryClient struct {
	calls          int
	owner          string
	repositoryName string
	branch         string
	commitSHA      string
	err            error
}

func (f *fakeRepositoryClient) GetLatestCommitSHA(
	ctx context.Context,
	owner string,
	repositoryName string,
	branch string,
) (string, error) {
	f.calls++
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
func TestRepositorySyncServicePropagatesRepositoryError(t *testing.T) {
	expectedErr := errors.New("database unavailable")

	repositoryRepo := &fakeRepositorySyncRepository{
		err: expectedErr,
	}

	snapshotRepo := &fakeRepositorySnapshotRepository{}
	githubClient := &fakeRepositoryClient{}

	service := NewRepositorySyncService(
		repositoryRepo,
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
		t.Fatal(
			"unexpected ErrRepositorySyncRepositoryNotFound",
		)
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
			Owner:         "devflow-test",
			Name:          "sync-test-repository",
			DefaultBranch: "",
		},
	}

	snapshotRepo := &fakeRepositorySnapshotRepository{}
	githubClient := &fakeRepositoryClient{}

	service := NewRepositorySyncService(
		repositoryRepo,
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
			Owner:         "",
			Name:          "sync-test-repository",
			DefaultBranch: "main",
		},
	}

	snapshotRepo := &fakeRepositorySnapshotRepository{}
	githubClient := &fakeRepositoryClient{}

	service := NewRepositorySyncService(
		repositoryRepo,
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
			Owner:         "devflow-test",
			Name:          "",
			DefaultBranch: "main",
		},
	}

	snapshotRepo := &fakeRepositorySnapshotRepository{}
	githubClient := &fakeRepositoryClient{}

	service := NewRepositorySyncService(
		repositoryRepo,
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
	repositoryRepo := &fakeRepositorySyncRepository{
		repository: &domain.Repository{
			ID:            "repository-id",
			Owner:         "devflow-test",
			Name:          "sync-test-repository",
			DefaultBranch: "main",
		},
	}

	snapshotRepo := &fakeRepositorySnapshotRepository{}

	githubClient := &fakeRepositoryClient{
		commitSHA: "abc123",
	}

	service := NewRepositorySyncService(
		repositoryRepo,
		snapshotRepo,
		githubClient,
	)

	_, _ = service.Sync(
		context.Background(),
		"repository-id",
	)

	if githubClient.calls != 1 {
		t.Fatalf(
			"expected GitHub client to be called once, got %d calls",
			githubClient.calls,
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

	repositoryRepo := &fakeRepositorySyncRepository{
		repository: &domain.Repository{
			ID:            "repository-id",
			Owner:         "devflow-test",
			Name:          "sync-test-repository",
			DefaultBranch: "main",
		},
	}

	snapshotRepo := &fakeRepositorySnapshotRepository{}

	githubClient := &fakeRepositoryClient{
		err: expectedErr,
	}

	service := NewRepositorySyncService(
		repositoryRepo,
		snapshotRepo,
		githubClient,
	)

	_, err := service.Sync(
		context.Background(),
		"repository-id",
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
