package service

import (
	"context"
	"errors"
	"testing"

	"github.com/kristopher1027/devflow-ai/internal/domain"
	"github.com/kristopher1027/devflow-ai/internal/repository"
)

type fakeRepositorySyncQueue struct {
	job        *domain.RepositorySyncJob
	err        error
	gotRequest RepositorySyncRequest
}

func (f *fakeRepositorySyncQueue) Enqueue(
	ctx context.Context,
	request RepositorySyncRequest,
) (*domain.RepositorySyncJob, error) {
	f.gotRequest = request

	if f.err != nil {
		return nil, f.err
	}

	return f.job, nil
}

func testRepositorySyncRequestService(
	t *testing.T,
) (
	RepositorySyncRequestService,
	*fakeRepositoryRepository,
	*fakeProjectRepository,
	*fakeProjectMemberRepository,
	*fakeRepositorySyncQueue,
) {
	t.Helper()

	repositoryRepo := &fakeRepositoryRepository{
		repository: &domain.Repository{
			ID:        "repository-123",
			ProjectID: "project-123",
		},
	}

	projectRepo := &fakeProjectRepository{
		project: testProject(),
	}

	memberRepo := &fakeProjectMemberRepository{
		member: testProjectMember(
			"user-123",
			WorkspaceMemberRoleMember,
		),
	}

	queue := &fakeRepositorySyncQueue{
		job: &domain.RepositorySyncJob{
			ID:           "job-123",
			RepositoryID: "repository-123",
			Status:       domain.RepositorySyncJobStatusPending,
		},
	}

	service := NewRepositorySyncRequestService(
		repositoryRepo,
		projectRepo,
		memberRepo,
		queue,
	)

	return service, repositoryRepo, projectRepo, memberRepo, queue
}

func TestRepositorySyncRequestServiceRequestSync(t *testing.T) {
	service, repositoryRepo, _, _, queue := testRepositorySyncRequestService(t)

	job, err := service.RequestSync(
		context.Background(),
		"user-123",
		"repository-123",
	)
	if err != nil {
		t.Fatalf("request sync: %v", err)
	}

	if job == nil {
		t.Fatal("expected job, got nil")
	}

	if job.ID != "job-123" {
		t.Fatalf("expected job ID job-123, got %s", job.ID)
	}

	if repositoryRepo.gotRepositoryID != "repository-123" {
		t.Fatalf(
			"expected repository ID repository-123, got %s",
			repositoryRepo.gotRepositoryID,
		)
	}

	if queue.gotRequest.RepositoryID != "repository-123" {
		t.Fatalf(
			"expected queued repository ID repository-123, got %s",
			queue.gotRequest.RepositoryID,
		)
	}
}

func TestRepositorySyncRequestServiceRequestSyncRepositoryNotFound(
	t *testing.T,
) {
	service, repositoryRepo, _, _, _ := testRepositorySyncRequestService(t)

	repositoryRepo.repository = nil
	repositoryRepo.findByIDErr = repository.ErrRepositoryNotFound

	_, err := service.RequestSync(
		context.Background(),
		"user-123",
		"repository-123",
	)

	if !errors.Is(err, repository.ErrRepositoryNotFound) {
		t.Fatalf(
			"expected ErrRepositoryNotFound, got %v",
			err,
		)
	}
}

func TestRepositorySyncRequestServiceRequestSyncProjectError(
	t *testing.T,
) {
	service, _, projectRepo, _, _ := testRepositorySyncRequestService(t)

	wantErr := errors.New("project lookup failed")
	projectRepo.err = wantErr

	_, err := service.RequestSync(
		context.Background(),
		"user-123",
		"repository-123",
	)

	if !errors.Is(err, wantErr) {
		t.Fatalf(
			"expected project lookup error, got %v",
			err,
		)
	}
}

func TestRepositorySyncRequestServiceRequestSyncUnauthorized(
	t *testing.T,
) {
	service, _, _, memberRepo, _ := testRepositorySyncRequestService(t)

	memberRepo.err = repository.ErrWorkspaceMemberNotFound

	_, err := service.RequestSync(
		context.Background(),
		"user-123",
		"repository-123",
	)

	if !errors.Is(err, ErrRepositorySyncRequestUnauthorized) {
		t.Fatalf(
			"expected ErrRepositorySyncRequestUnauthorized, got %v",
			err,
		)
	}
}

func TestRepositorySyncRequestServiceRequestSyncViewerDenied(
	t *testing.T,
) {
	service, _, _, memberRepo, _ := testRepositorySyncRequestService(t)

	memberRepo.member = testProjectMember(
		"user-123",
		"viewer",
	)

	_, err := service.RequestSync(
		context.Background(),
		"user-123",
		"repository-123",
	)

	if !errors.Is(err, ErrRepositorySyncRequestUnauthorized) {
		t.Fatalf(
			"expected ErrRepositorySyncRequestUnauthorized, got %v",
			err,
		)
	}
}

func TestRepositorySyncRequestServiceRequestSyncQueueError(
	t *testing.T,
) {
	service, _, _, _, queue := testRepositorySyncRequestService(t)

	wantErr := errors.New("queue unavailable")
	queue.err = wantErr

	_, err := service.RequestSync(
		context.Background(),
		"user-123",
		"repository-123",
	)

	if !errors.Is(err, wantErr) {
		t.Fatalf(
			"expected queue error, got %v",
			err,
		)
	}
}
func TestRepositorySyncRequestServiceRequestSyncAllowedRoles(t *testing.T) {
	roles := []string{
		WorkspaceMemberRoleOwner,
		WorkspaceMemberRoleAdmin,
		WorkspaceMemberRoleMember,
	}

	for _, role := range roles {
		t.Run(role, func(t *testing.T) {
			service, _, _, memberRepo, queue :=
				testRepositorySyncRequestService(t)

			memberRepo.member = testProjectMember(
				"user-123",
				role,
			)

			job, err := service.RequestSync(
				context.Background(),
				"user-123",
				"repository-123",
			)
			if err != nil {
				t.Fatalf(
					"expected role %s to be allowed, got %v",
					role,
					err,
				)
			}

			if job == nil {
				t.Fatal("expected job, got nil")
			}

			if queue.gotRequest.RepositoryID != "repository-123" {
				t.Fatalf(
					"expected repository ID repository-123, got %s",
					queue.gotRequest.RepositoryID,
				)
			}
		})
	}
}
