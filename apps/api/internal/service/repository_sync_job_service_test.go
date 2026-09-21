package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kristopher1027/devflow-ai/internal/domain"
	"github.com/kristopher1027/devflow-ai/internal/repository"
)

type fakeRepositorySyncJobRepository struct {
	job       *domain.RepositorySyncJob
	findErr   error
	findCalls int
	gotJobID  string
}

func (f *fakeRepositorySyncJobRepository) Create(
	ctx context.Context,
	job *domain.RepositorySyncJob,
) error {
	return nil
}

func (f *fakeRepositorySyncJobRepository) FindByID(
	ctx context.Context,
	id string,
) (*domain.RepositorySyncJob, error) {
	f.findCalls++
	f.gotJobID = id

	if f.findErr != nil {
		return nil, f.findErr
	}

	return f.job, nil
}

func (f *fakeRepositorySyncJobRepository) ListByRepositoryID(
	ctx context.Context,
	repositoryID string,
) ([]*domain.RepositorySyncJob, error) {
	return nil, nil
}

func (f *fakeRepositorySyncJobRepository) MarkRunning(
	ctx context.Context,
	id string,
	attempts int,
) error {
	return nil
}

func (f *fakeRepositorySyncJobRepository) MarkSucceeded(
	ctx context.Context,
	id string,
	attempts int,
) error {
	return nil
}

func (f *fakeRepositorySyncJobRepository) MarkFailed(
	ctx context.Context,
	id string,
	attempts int,
	code string,
	message string,
) error {
	return nil
}

type fakeSyncRepositoryRepository struct {
	repository *domain.Repository
	findErr    error
	findCalls  int
	gotID      string
}

func (f *fakeSyncRepositoryRepository) Create(
	ctx context.Context,
	repo *domain.Repository,
) error {
	return nil
}

func (f *fakeSyncRepositoryRepository) FindByID(
	ctx context.Context,
	id string,
) (*domain.Repository, error) {
	f.findCalls++
	f.gotID = id

	if f.findErr != nil {
		return nil, f.findErr
	}

	return f.repository, nil
}

func (f *fakeSyncRepositoryRepository) ListByProjectID(
	ctx context.Context,
	projectID string,
) ([]*domain.Repository, error) {
	return nil, nil
}

func (f *fakeSyncRepositoryRepository) FindByProviderExternalID(
	ctx context.Context,
	provider string,
	externalID string,
) (*domain.Repository, error) {
	return nil, nil
}
func (f *fakeSyncRepositoryRepository) UpdateSyncStatus(
	ctx context.Context,
	id string,
	status string,
	lastSyncedAt *time.Time,
) error {
	return nil
}
func (f *fakeSyncRepositoryRepository) Delete(
	ctx context.Context,
	id string,
) error {
	return nil
}

type fakeSyncProjectRepository struct {
	project   *domain.Project
	findErr   error
	findCalls int
	gotID     string
}

func (f *fakeSyncProjectRepository) Create(
	ctx context.Context,
	project *domain.Project,
) error {
	return nil
}

func (f *fakeSyncProjectRepository) FindByID(
	ctx context.Context,
	id string,
) (*domain.Project, error) {
	f.findCalls++
	f.gotID = id

	if f.findErr != nil {
		return nil, f.findErr
	}

	return f.project, nil
}

func (f *fakeSyncProjectRepository) ListByWorkspaceID(
	ctx context.Context,
	workspaceID string,
) ([]*domain.Project, error) {
	return nil, nil
}

func (f *fakeSyncProjectRepository) Delete(
	ctx context.Context,
	id string,
) error {
	return nil
}

type fakeSyncWorkspaceMemberRepository struct {
	err       error
	findCalls int
	gotUserID string
	gotWSID   string
}

func (f *fakeSyncWorkspaceMemberRepository) Create(
	ctx context.Context,
	member *domain.WorkspaceMember,
) error {
	return nil
}

func (f *fakeSyncWorkspaceMemberRepository) Find(
	ctx context.Context,
	workspaceID string,
	userID string,
) (*domain.WorkspaceMember, error) {
	f.findCalls++
	f.gotWSID = workspaceID
	f.gotUserID = userID

	if f.err != nil {
		return nil, f.err
	}

	return &domain.WorkspaceMember{
		WorkspaceID: workspaceID,
		UserID:      userID,
		Role:        WorkspaceMemberRoleMember,
	}, nil
}

func (f *fakeSyncWorkspaceMemberRepository) ListByWorkspaceID(
	ctx context.Context,
	workspaceID string,
) ([]*domain.WorkspaceMember, error) {
	return nil, nil
}

func (f *fakeSyncWorkspaceMemberRepository) UpdateRole(
	ctx context.Context,
	workspaceID string,
	userID string,
	role string,
) error {
	return nil
}

func (f *fakeSyncWorkspaceMemberRepository) Delete(
	ctx context.Context,
	workspaceID string,
	userID string,
) error {
	return nil
}

func testRepositorySyncJob() *domain.RepositorySyncJob {
	now := time.Now()

	return &domain.RepositorySyncJob{
		ID:           "sync-job-123",
		RepositoryID: "repository-123",
		Status:       domain.RepositorySyncJobStatusPending,
		Attempts:     0,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func testSyncRepository() *domain.Repository {
	return &domain.Repository{
		ID:            "repository-123",
		ProjectID:     "project-123",
		Provider:      "github",
		ExternalID:    "12345",
		Owner:         "test-owner",
		Name:          "test-repo",
		FullName:      "test-owner/test-repo",
		DefaultBranch: "main",
		HTMLURL:       "https://github.com/test-owner/test-repo",
		CloneURL:      "https://github.com/test-owner/test-repo.git",
	}
}

func testSyncProject() *domain.Project {
	return &domain.Project{
		ID:          "project-123",
		WorkspaceID: "workspace-123",
		Name:        "Test Project",
		CreatedBy:   "user-123",
	}
}

func TestRepositorySyncJobServiceFindByID(t *testing.T) {
	job := testRepositorySyncJob()
	repo := testSyncRepository()
	project := testSyncProject()

	jobRepo := &fakeRepositorySyncJobRepository{
		job: job,
	}
	repoRepo := &fakeSyncRepositoryRepository{
		repository: repo,
	}
	projectRepo := &fakeSyncProjectRepository{
		project: project,
	}
	memberRepo := &fakeSyncWorkspaceMemberRepository{}

	service := NewRepositorySyncJobService(
		jobRepo,
		repoRepo,
		projectRepo,
		memberRepo,
	)

	got, err := service.FindByID(
		context.Background(),
		"user-123",
		"sync-job-123",
	)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}

	if got != job {
		t.Fatalf("FindByID() returned unexpected job")
	}

	if jobRepo.gotJobID != "sync-job-123" {
		t.Fatalf("job repository received ID %q", jobRepo.gotJobID)
	}

	if repoRepo.gotID != "repository-123" {
		t.Fatalf("repository repository received ID %q", repoRepo.gotID)
	}

	if projectRepo.gotID != "project-123" {
		t.Fatalf("project repository received ID %q", projectRepo.gotID)
	}

	if memberRepo.gotWSID != "workspace-123" {
		t.Fatalf("member repository received workspace ID %q", memberRepo.gotWSID)
	}

	if memberRepo.gotUserID != "user-123" {
		t.Fatalf("member repository received user ID %q", memberRepo.gotUserID)
	}
}

func TestRepositorySyncJobServiceFindByIDUnauthorized(t *testing.T) {
	jobRepo := &fakeRepositorySyncJobRepository{
		job: testRepositorySyncJob(),
	}
	repoRepo := &fakeSyncRepositoryRepository{
		repository: testSyncRepository(),
	}
	projectRepo := &fakeSyncProjectRepository{
		project: testSyncProject(),
	}
	memberRepo := &fakeSyncWorkspaceMemberRepository{
		err: repository.ErrWorkspaceMemberNotFound,
	}

	service := NewRepositorySyncJobService(
		jobRepo,
		repoRepo,
		projectRepo,
		memberRepo,
	)

	_, err := service.FindByID(
		context.Background(),
		"user-unauthorized",
		"sync-job-123",
	)

	if !errors.Is(err, ErrRepositorySyncJobUnauthorized) {
		t.Fatalf(
			"FindByID() error = %v, want %v",
			err,
			ErrRepositorySyncJobUnauthorized,
		)
	}
}

func TestRepositorySyncJobServiceFindByIDJobNotFound(t *testing.T) {
	jobRepo := &fakeRepositorySyncJobRepository{
		findErr: repository.ErrRepositorySyncJobNotFound,
	}
	repoRepo := &fakeSyncRepositoryRepository{}
	projectRepo := &fakeSyncProjectRepository{}
	memberRepo := &fakeSyncWorkspaceMemberRepository{}

	service := NewRepositorySyncJobService(
		jobRepo,
		repoRepo,
		projectRepo,
		memberRepo,
	)

	_, err := service.FindByID(
		context.Background(),
		"user-123",
		"sync-job-123",
	)

	if !errors.Is(err, repository.ErrRepositorySyncJobNotFound) {
		t.Fatalf(
			"FindByID() error = %v, want %v",
			err,
			repository.ErrRepositorySyncJobNotFound,
		)
	}

	if repoRepo.findCalls != 0 {
		t.Fatalf("repository lookup should not occur")
	}

	if projectRepo.findCalls != 0 {
		t.Fatalf("project lookup should not occur")
	}

	if memberRepo.findCalls != 0 {
		t.Fatalf("member lookup should not occur")
	}
}

func TestRepositorySyncJobServiceFindByIDRepositoryError(t *testing.T) {
	expectedErr := errors.New("repository lookup failed")

	jobRepo := &fakeRepositorySyncJobRepository{
		job: testRepositorySyncJob(),
	}
	repoRepo := &fakeSyncRepositoryRepository{
		findErr: expectedErr,
	}
	projectRepo := &fakeSyncProjectRepository{
		project: testSyncProject(),
	}
	memberRepo := &fakeSyncWorkspaceMemberRepository{}

	service := NewRepositorySyncJobService(
		jobRepo,
		repoRepo,
		projectRepo,
		memberRepo,
	)

	_, err := service.FindByID(
		context.Background(),
		"user-123",
		"sync-job-123",
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("FindByID() error = %v, want %v", err, expectedErr)
	}

	if projectRepo.findCalls != 0 {
		t.Fatalf("project lookup should not occur")
	}

	if memberRepo.findCalls != 0 {
		t.Fatalf("member lookup should not occur")
	}
}

func TestRepositorySyncJobServiceFindByIDProjectError(t *testing.T) {
	expectedErr := errors.New("project lookup failed")

	jobRepo := &fakeRepositorySyncJobRepository{
		job: testRepositorySyncJob(),
	}
	repoRepo := &fakeSyncRepositoryRepository{
		repository: testSyncRepository(),
	}
	projectRepo := &fakeSyncProjectRepository{
		findErr: expectedErr,
	}
	memberRepo := &fakeSyncWorkspaceMemberRepository{}

	service := NewRepositorySyncJobService(
		jobRepo,
		repoRepo,
		projectRepo,
		memberRepo,
	)

	_, err := service.FindByID(
		context.Background(),
		"user-123",
		"sync-job-123",
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("FindByID() error = %v, want %v", err, expectedErr)
	}

	if memberRepo.findCalls != 0 {
		t.Fatalf("member lookup should not occur")
	}
}

func TestRepositorySyncJobServiceFindByIDMemberError(t *testing.T) {
	expectedErr := errors.New("membership lookup failed")

	jobRepo := &fakeRepositorySyncJobRepository{
		job: testRepositorySyncJob(),
	}
	repoRepo := &fakeSyncRepositoryRepository{
		repository: testSyncRepository(),
	}
	projectRepo := &fakeSyncProjectRepository{
		project: testSyncProject(),
	}
	memberRepo := &fakeSyncWorkspaceMemberRepository{
		err: expectedErr,
	}

	service := NewRepositorySyncJobService(
		jobRepo,
		repoRepo,
		projectRepo,
		memberRepo,
	)

	_, err := service.FindByID(
		context.Background(),
		"user-123",
		"sync-job-123",
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("FindByID() error = %v, want %v", err, expectedErr)
	}
}

var _ RepositorySyncJobService = (*RepositorySyncJobServiceImpl)(nil)
var _ repository.RepositorySyncJobRepository = (*fakeRepositorySyncJobRepository)(nil)
var _ repository.RepositoryRepository = (*fakeSyncRepositoryRepository)(nil)
var _ repository.ProjectRepository = (*fakeSyncProjectRepository)(nil)
var _ repository.WorkspaceMemberRepository = (*fakeSyncWorkspaceMemberRepository)(nil)
