package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kristopher1027/devflow-ai/internal/domain"
	"github.com/kristopher1027/devflow-ai/internal/repository"
)

type fakeGitHubRepositoryImportJobRepository struct {
	job       *domain.GitHubRepositoryImportJob
	findErr   error
	findCalls int
}

func (f *fakeGitHubRepositoryImportJobRepository) Create(
	ctx context.Context,
	job *domain.GitHubRepositoryImportJob,
) error {
	return nil
}


func (f *fakeGitHubRepositoryImportJobRepository) FindByID(
	ctx context.Context,
	id string,
) (*domain.GitHubRepositoryImportJob, error) {
	f.findCalls++

	if f.findErr != nil {
		return nil, f.findErr
	}

	return f.job, nil
}

func (f *fakeGitHubRepositoryImportJobRepository) MarkRunning(
	ctx context.Context,
	id string,
	attempts int,
) error {
	return nil
}

func (f *fakeGitHubRepositoryImportJobRepository) MarkSucceeded(
	ctx context.Context,
	id string,
	attempts int,
) error {
	return nil
}

func (f *fakeGitHubRepositoryImportJobRepository) MarkFailed(
	ctx context.Context,
	id string,
	attempts int,
	code string,
	message string,
) error {
	return nil
}

type fakeProjectRepositoryForImportJob struct {
	project *domain.Project
	findErr error
}

func (f *fakeProjectRepositoryForImportJob) Create(
	ctx context.Context,
	project *domain.Project,
) error {
	return nil
}

func (f *fakeProjectRepositoryForImportJob) FindByID(
	ctx context.Context,
	id string,
) (*domain.Project, error) {
	if f.findErr != nil {
		return nil, f.findErr
	}

	return f.project, nil
}

func (f *fakeProjectRepositoryForImportJob) ListByWorkspaceID(
	ctx context.Context,
	workspaceID string,
) ([]*domain.Project, error) {
	return nil, nil
}

func (f *fakeProjectRepositoryForImportJob) Delete(
	ctx context.Context,
	id string,
) error {
	return nil
}

type fakeWorkspaceMemberRepositoryForImportJob struct {
	member   *domain.WorkspaceMember
	findErr  error
	findArgs struct {
		workspaceID string
		userID      string
	}
}

func (f *fakeWorkspaceMemberRepositoryForImportJob) Create(
	ctx context.Context,
	member *domain.WorkspaceMember,
) error {
	return nil
}

func (f *fakeWorkspaceMemberRepositoryForImportJob) Find(
	ctx context.Context,
	workspaceID string,
	userID string,
) (*domain.WorkspaceMember, error) {
	f.findArgs.workspaceID = workspaceID
	f.findArgs.userID = userID

	if f.findErr != nil {
		return nil, f.findErr
	}

	return f.member, nil
}

func (f *fakeWorkspaceMemberRepositoryForImportJob) ListByWorkspaceID(
	ctx context.Context,
	workspaceID string,
) ([]*domain.WorkspaceMember, error) {
	return nil, nil
}

func (f *fakeWorkspaceMemberRepositoryForImportJob) UpdateRole(
	ctx context.Context,
	workspaceID string,
	userID string,
	role string,
) error {
	return nil
}

func (f *fakeWorkspaceMemberRepositoryForImportJob) Delete(
	ctx context.Context,
	workspaceID string,
	userID string,
) error {
	return nil
}

func TestGitHubRepositoryImportJobServiceFindByID(t *testing.T) {
	ctx := context.Background()

	job := &domain.GitHubRepositoryImportJob{
		ID:          "job-123",
		RequesterID: "user-123",
		ProjectID:   "project-123",
		Status:      domain.GitHubRepositoryImportJobStatusRunning,
		Attempts:    1,
	}

	project := &domain.Project{
		ID:          "project-123",
		WorkspaceID: "workspace-123",
	}

	member := &domain.WorkspaceMember{
		WorkspaceID: "workspace-123",
		UserID:      "user-123",
		Role:        WorkspaceMemberRoleMember,
	}

	jobRepository := &fakeGitHubRepositoryImportJobRepository{
		job: job,
	}

	projectRepository := &fakeProjectRepositoryForImportJob{
		project: project,
	}

	memberRepository := &fakeWorkspaceMemberRepositoryForImportJob{
		member: member,
	}

	service := NewGitHubRepositoryImportJobService(
		jobRepository,
		projectRepository,
		memberRepository,
	)

	result, err := service.FindByID(
		ctx,
		"user-123",
		"job-123",
	)

	require.NoError(t, err)
	require.Equal(t, job, result)
	require.Equal(t, 1, jobRepository.findCalls)
	require.Equal(t, "workspace-123", memberRepository.findArgs.workspaceID)
	require.Equal(t, "user-123", memberRepository.findArgs.userID)
}

func TestGitHubRepositoryImportJobServiceFindByIDNotFound(t *testing.T) {
	expectedErr := repository.ErrGitHubRepositoryImportJobNotFound

	jobRepository := &fakeGitHubRepositoryImportJobRepository{
		findErr: expectedErr,
	}

	projectRepository := &fakeProjectRepositoryForImportJob{}

	memberRepository := &fakeWorkspaceMemberRepositoryForImportJob{}

	service := NewGitHubRepositoryImportJobService(
		jobRepository,
		projectRepository,
		memberRepository,
	)

	_, err := service.FindByID(
		context.Background(),
		"user-123",
		"job-123",
	)

	require.ErrorIs(t, err, expectedErr)
}

func TestGitHubRepositoryImportJobServiceFindByIDUnauthorized(t *testing.T) {
	job := &domain.GitHubRepositoryImportJob{
		ID:        "job-123",
		ProjectID: "project-123",
	}

	project := &domain.Project{
		ID:          "project-123",
		WorkspaceID: "workspace-123",
	}

	jobRepository := &fakeGitHubRepositoryImportJobRepository{
		job: job,
	}

	projectRepository := &fakeProjectRepositoryForImportJob{
		project: project,
	}

	memberRepository := &fakeWorkspaceMemberRepositoryForImportJob{
		findErr: repository.ErrWorkspaceMemberNotFound,
	}

	service := NewGitHubRepositoryImportJobService(
		jobRepository,
		projectRepository,
		memberRepository,
	)

	_, err := service.FindByID(
		context.Background(),
		"user-456",
		"job-123",
	)

	require.ErrorIs(
		t,
		err,
		ErrGitHubRepositoryImportJobUnauthorized,
	)
}

func TestGitHubRepositoryImportJobServiceFindByIDProjectError(t *testing.T) {
	expectedErr := errors.New("project repository failure")

	jobRepository := &fakeGitHubRepositoryImportJobRepository{
		job: &domain.GitHubRepositoryImportJob{
			ID:        "job-123",
			ProjectID: "project-123",
		},
	}

	projectRepository := &fakeProjectRepositoryForImportJob{
		findErr: expectedErr,
	}

	memberRepository := &fakeWorkspaceMemberRepositoryForImportJob{}

	service := NewGitHubRepositoryImportJobService(
		jobRepository,
		projectRepository,
		memberRepository,
	)

	_, err := service.FindByID(
		context.Background(),
		"user-123",
		"job-123",
	)

	require.ErrorIs(t, err, expectedErr)
}

func TestGitHubRepositoryImportJobServiceFindByIDMemberError(t *testing.T) {
	expectedErr := errors.New("membership repository failure")

	jobRepository := &fakeGitHubRepositoryImportJobRepository{
		job: &domain.GitHubRepositoryImportJob{
			ID:        "job-123",
			ProjectID: "project-123",
		},
	}

	projectRepository := &fakeProjectRepositoryForImportJob{
		project: &domain.Project{
			ID:          "project-123",
			WorkspaceID: "workspace-123",
		},
	}

	memberRepository := &fakeWorkspaceMemberRepositoryForImportJob{
		findErr: expectedErr,
	}

	service := NewGitHubRepositoryImportJobService(
		jobRepository,
		projectRepository,
		memberRepository,
	)

	_, err := service.FindByID(
		context.Background(),
		"user-123",
		"job-123",
	)

	require.ErrorIs(t, err, expectedErr)
}
