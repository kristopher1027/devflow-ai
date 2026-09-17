package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kristopher1027/devflow-ai/internal/domain"
	"github.com/kristopher1027/devflow-ai/internal/repository"
)

type fakeRepositoryRepository struct {
	repository  *domain.Repository
	findResult  *domain.Repository
	listResult  []*domain.Repository
	findErr     error
	findByIDErr error
	listErr     error
	createErr   error
	deleteErr   error

	gotRepository   *domain.Repository
	gotProvider     string
	gotExternalID   string
	gotRepositoryID string
	gotProjectID    string
}

func (f *fakeRepositoryRepository) Create(
	ctx context.Context,
	repository *domain.Repository,
) error {
	f.gotRepository = repository

	if f.createErr != nil {
		return f.createErr
	}

	f.repository = repository

	return nil
}

func (f *fakeRepositoryRepository) FindByID(
	ctx context.Context,
	id string,
) (*domain.Repository, error) {
	f.gotRepositoryID = id

	return f.repository, f.findByIDErr
}

func (f *fakeRepositoryRepository) ListByProjectID(
	ctx context.Context,
	projectID string,
) ([]*domain.Repository, error) {
	f.gotProjectID = projectID

	return f.listResult, f.listErr
}

func (f *fakeRepositoryRepository) FindByProviderExternalID(
	ctx context.Context,
	provider string,
	externalID string,
) (*domain.Repository, error) {
	f.gotProvider = provider
	f.gotExternalID = externalID

	return f.findResult, f.findErr
}

func (f *fakeRepositoryRepository) Delete(
	ctx context.Context,
	id string,
) error {
	f.gotRepositoryID = id

	return f.deleteErr
}

func testRepositoryService(t *testing.T) (
	*RepositoryServiceImpl,
	*fakeRepositoryRepository,
) {
	t.Helper()

	repositoryRepo := &fakeRepositoryRepository{
		findErr: repository.ErrRepositoryNotFound,
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

	return NewRepositoryService(
		repositoryRepo,
		projectRepo,
		memberRepo,
	), repositoryRepo
}

func TestRepositoryServiceCreate(t *testing.T) {
	service, repositoryRepo := testRepositoryService(t)

	createdBefore := time.Now()
	repository, err := service.Create(
		context.Background(),
		"user-123",
		"project-123",
		"github",
		"12345",
		"octocat",
		"hello-world",
		"octocat/hello-world",
		"main",
		"https://github.com/octocat/hello-world",
		"https://github.com/octocat/hello-world.git",
		true,
	)
	createdAfter := time.Now()

	if err != nil {
		t.Fatalf("create repository: %v", err)
	}

	if repository == nil {
		t.Fatal("expected repository, got nil")
	}

	if repository.ID == "" {
		t.Fatal("expected repository ID to be generated")
	}

	if repository.ProjectID != "project-123" {
		t.Fatalf(
			"expected project ID project-123, got %s",
			repository.ProjectID,
		)
	}

	if repository.Provider != "github" {
		t.Fatalf(
			"expected provider github, got %s",
			repository.Provider,
		)
	}

	if repository.ExternalID != "12345" {
		t.Fatalf(
			"expected external ID 12345, got %s",
			repository.ExternalID,
		)
	}

	if repository.Owner != "octocat" {
		t.Fatalf(
			"expected owner octocat, got %s",
			repository.Owner,
		)
	}

	if repository.Name != "hello-world" {
		t.Fatalf(
			"expected name hello-world, got %s",
			repository.Name,
		)
	}

	if repository.FullName != "octocat/hello-world" {
		t.Fatalf(
			"expected full name octocat/hello-world, got %s",
			repository.FullName,
		)
	}

	if repository.DefaultBranch != "main" {
		t.Fatalf(
			"expected default branch main, got %s",
			repository.DefaultBranch,
		)
	}

	if repository.HTMLURL != "https://github.com/octocat/hello-world" {
		t.Fatalf(
			"expected HTML URL, got %s",
			repository.HTMLURL,
		)
	}

	if repository.CloneURL != "https://github.com/octocat/hello-world.git" {
		t.Fatalf(
			"expected clone URL, got %s",
			repository.CloneURL,
		)
	}

	if !repository.IsPrivate {
		t.Fatal("expected private repository")
	}

	if repository.SyncStatus != "pending" {
		t.Fatalf(
			"expected sync status pending, got %s",
			repository.SyncStatus,
		)
	}

	if repository.LastSyncedAt != nil {
		t.Fatal("expected last synced at to be nil")
	}

	if repository.CreatedAt.Before(createdBefore) || repository.CreatedAt.After(createdAfter) {
		t.Fatalf("expected created timestamp between test timestamps, got %v", repository.CreatedAt)
	}

	if repository.UpdatedAt.Before(createdBefore) || repository.UpdatedAt.After(createdAfter) {
		t.Fatalf("expected updated timestamp between test timestamps, got %v", repository.UpdatedAt)
	}

	if repositoryRepo.gotRepository != repository {
		t.Fatal("expected repository model to be passed to repository layer")
	}

	if repositoryRepo.gotProvider != "github" {
		t.Fatalf(
			"expected lookup provider github, got %s",
			repositoryRepo.gotProvider,
		)
	}

	if repositoryRepo.gotExternalID != "12345" {
		t.Fatalf(
			"expected lookup external ID 12345, got %s",
			repositoryRepo.gotExternalID,
		)
	}
}

func TestRepositoryServiceCreateProviderUnsupported(t *testing.T) {
	service, _ := testRepositoryService(t)

	_, err := service.Create(
		context.Background(),
		"user-123",
		"project-123",
		"gitlab",
		"12345",
		"octocat",
		"hello-world",
		"octocat/hello-world",
		"main",
		"https://gitlab.com/octocat/hello-world",
		"https://gitlab.com/octocat/hello-world.git",
		false,
	)

	if !errors.Is(err, ErrRepositoryProviderUnsupported) {
		t.Fatalf(
			"expected ErrRepositoryProviderUnsupported, got %v",
			err,
		)
	}
}

func TestRepositoryServiceCreateRequiredFields(t *testing.T) {
	tests := []struct {
		name string
		call func(*RepositoryServiceImpl) error
		want error
	}{
		{
			name: "external ID",
			call: func(service *RepositoryServiceImpl) error {
				_, err := service.Create(
					context.Background(),
					"user-123",
					"project-123",
					"github",
					"   ",
					"octocat",
					"hello-world",
					"octocat/hello-world",
					"main",
					"https://github.com/octocat/hello-world",
					"https://github.com/octocat/hello-world.git",
					false,
				)
				return err
			},
			want: ErrRepositoryExternalIDRequired,
		},
		{
			name: "owner",
			call: func(service *RepositoryServiceImpl) error {
				_, err := service.Create(
					context.Background(),
					"user-123",
					"project-123",
					"github",
					"12345",
					"   ",
					"hello-world",
					"octocat/hello-world",
					"main",
					"https://github.com/octocat/hello-world",
					"https://github.com/octocat/hello-world.git",
					false,
				)
				return err
			},
			want: ErrRepositoryOwnerRequired,
		},
		{
			name: "name",
			call: func(service *RepositoryServiceImpl) error {
				_, err := service.Create(
					context.Background(),
					"user-123",
					"project-123",
					"github",
					"12345",
					"octocat",
					"   ",
					"octocat/hello-world",
					"main",
					"https://github.com/octocat/hello-world",
					"https://github.com/octocat/hello-world.git",
					false,
				)
				return err
			},
			want: ErrRepositoryNameRequired,
		},
		{
			name: "full name",
			call: func(service *RepositoryServiceImpl) error {
				_, err := service.Create(
					context.Background(),
					"user-123",
					"project-123",
					"github",
					"12345",
					"octocat",
					"hello-world",
					"   ",
					"main",
					"https://github.com/octocat/hello-world",
					"https://github.com/octocat/hello-world.git",
					false,
				)
				return err
			},
			want: ErrRepositoryFullNameRequired,
		},
		{
			name: "default branch",
			call: func(service *RepositoryServiceImpl) error {
				_, err := service.Create(
					context.Background(),
					"user-123",
					"project-123",
					"github",
					"12345",
					"octocat",
					"hello-world",
					"octocat/hello-world",
					"   ",
					"https://github.com/octocat/hello-world",
					"https://github.com/octocat/hello-world.git",
					false,
				)
				return err
			},
			want: ErrRepositoryDefaultBranchRequired,
		},
		{
			name: "HTML URL",
			call: func(service *RepositoryServiceImpl) error {
				_, err := service.Create(
					context.Background(),
					"user-123",
					"project-123",
					"github",
					"12345",
					"octocat",
					"hello-world",
					"octocat/hello-world",
					"main",
					"   ",
					"https://github.com/octocat/hello-world.git",
					false,
				)
				return err
			},
			want: ErrRepositoryHTMLURLRequired,
		},
		{
			name: "clone URL",
			call: func(service *RepositoryServiceImpl) error {
				_, err := service.Create(
					context.Background(),
					"user-123",
					"project-123",
					"github",
					"12345",
					"octocat",
					"hello-world",
					"octocat/hello-world",
					"main",
					"https://github.com/octocat/hello-world",
					"   ",
					false,
				)
				return err
			},
			want: ErrRepositoryCloneURLRequired,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service, _ := testRepositoryService(t)

			if err := test.call(service); !errors.Is(err, test.want) {
				t.Fatalf("expected %v, got %v", test.want, err)
			}
		})
	}
}

func TestRepositoryServiceCreateUnauthorized(t *testing.T) {
	repositoryRepo := &fakeRepositoryRepository{
		findErr: repository.ErrRepositoryNotFound,
	}
	projectRepo := &fakeProjectRepository{
		project: testProject(),
	}
	memberRepo := &fakeProjectMemberRepository{
		err: repository.ErrWorkspaceMemberNotFound,
	}
	service := NewRepositoryService(repositoryRepo, projectRepo, memberRepo)

	_, err := service.Create(
		context.Background(),
		"user-999",
		"project-123",
		"github",
		"12345",
		"octocat",
		"hello-world",
		"octocat/hello-world",
		"main",
		"https://github.com/octocat/hello-world",
		"https://github.com/octocat/hello-world.git",
		false,
	)

	if !errors.Is(err, ErrRepositoryUnauthorized) {
		t.Fatalf("expected ErrRepositoryUnauthorized, got %v", err)
	}
}

func TestRepositoryServiceCreateDuplicate(t *testing.T) {
	service, repositoryRepo := testRepositoryService(t)
	repositoryRepo.findResult = &domain.Repository{ID: "existing-repository"}
	repositoryRepo.findErr = nil

	_, err := service.Create(
		context.Background(),
		"user-123",
		"project-123",
		"github",
		"12345",
		"octocat",
		"hello-world",
		"octocat/hello-world",
		"main",
		"https://github.com/octocat/hello-world",
		"https://github.com/octocat/hello-world.git",
		false,
	)

	if !errors.Is(err, ErrRepositoryAlreadyExists) {
		t.Fatalf("expected ErrRepositoryAlreadyExists, got %v", err)
	}
}

func TestRepositoryServiceCreateRepositoryError(t *testing.T) {
	repositoryErr := errors.New("database failure")
	service, repositoryRepo := testRepositoryService(t)
	repositoryRepo.findErr = repositoryErr

	_, err := service.Create(
		context.Background(),
		"user-123",
		"project-123",
		"github",
		"12345",
		"octocat",
		"hello-world",
		"octocat/hello-world",
		"main",
		"https://github.com/octocat/hello-world",
		"https://github.com/octocat/hello-world.git",
		false,
	)

	if !errors.Is(err, repositoryErr) {
		t.Fatalf("expected repository error, got %v", err)
	}
}

func TestRepositoryServiceFindByID(t *testing.T) {
	expectedRepository := &domain.Repository{
		ID:        "repository-123",
		ProjectID: "project-123",
	}
	repositoryRepo := &fakeRepositoryRepository{
		repository: expectedRepository,
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
	service := NewRepositoryService(repositoryRepo, projectRepo, memberRepo)

	repository, err := service.FindByID(
		context.Background(),
		"user-123",
		"repository-123",
	)

	if err != nil {
		t.Fatalf("find repository: %v", err)
	}

	if repository != expectedRepository {
		t.Fatal("expected repository returned from repository layer")
	}

	if repositoryRepo.gotRepositoryID != "repository-123" {
		t.Fatalf(
			"expected repository ID repository-123, got %s",
			repositoryRepo.gotRepositoryID,
		)
	}

	if projectRepo.gotProjectID != "project-123" {
		t.Fatalf(
			"expected project ID project-123, got %s",
			projectRepo.gotProjectID,
		)
	}

	if memberRepo.gotWorkspaceID != "workspace-123" {
		t.Fatalf(
			"expected workspace ID workspace-123, got %s",
			memberRepo.gotWorkspaceID,
		)
	}

	if memberRepo.gotUserID != "user-123" {
		t.Fatalf(
			"expected user ID user-123, got %s",
			memberRepo.gotUserID,
		)
	}
}

func TestRepositoryServiceFindByIDUnauthorized(t *testing.T) {
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
		err: repository.ErrWorkspaceMemberNotFound,
	}
	service := NewRepositoryService(repositoryRepo, projectRepo, memberRepo)

	_, err := service.FindByID(
		context.Background(),
		"user-999",
		"repository-123",
	)

	if !errors.Is(err, ErrRepositoryUnauthorized) {
		t.Fatalf("expected ErrRepositoryUnauthorized, got %v", err)
	}
}

func TestRepositoryServiceFindByIDNotFound(t *testing.T) {
	repositoryRepo := &fakeRepositoryRepository{
		findByIDErr: repository.ErrRepositoryNotFound,
	}
	service := NewRepositoryService(
		repositoryRepo,
		&fakeProjectRepository{},
		&fakeProjectMemberRepository{},
	)

	_, err := service.FindByID(
		context.Background(),
		"user-123",
		"repository-123",
	)

	if !errors.Is(err, repository.ErrRepositoryNotFound) {
		t.Fatalf(
			"expected repository not found error, got %v",
			err,
		)
	}
}

func TestRepositoryServiceFindByIDRepositoryError(t *testing.T) {
	repositoryErr := errors.New("database failure")
	repositoryRepo := &fakeRepositoryRepository{
		findByIDErr: repositoryErr,
	}
	service := NewRepositoryService(
		repositoryRepo,
		&fakeProjectRepository{},
		&fakeProjectMemberRepository{},
	)

	_, err := service.FindByID(
		context.Background(),
		"user-123",
		"repository-123",
	)

	if !errors.Is(err, repositoryErr) {
		t.Fatalf("expected repository error, got %v", err)
	}
}

func TestRepositoryServiceListByProjectID(t *testing.T) {
	expectedRepositories := []*domain.Repository{
		{
			ID:        "repository-123",
			ProjectID: "project-123",
		},
		{
			ID:        "repository-456",
			ProjectID: "project-123",
		},
	}
	repositoryRepo := &fakeRepositoryRepository{
		listResult: expectedRepositories,
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
	service := NewRepositoryService(repositoryRepo, projectRepo, memberRepo)

	repositories, err := service.ListByProjectID(
		context.Background(),
		"user-123",
		"project-123",
	)

	if err != nil {
		t.Fatalf("list repositories: %v", err)
	}

	if len(repositories) != len(expectedRepositories) {
		t.Fatalf(
			"expected %d repositories, got %d",
			len(expectedRepositories),
			len(repositories),
		)
	}

	for index, expectedRepository := range expectedRepositories {
		if repositories[index] != expectedRepository {
			t.Fatalf("expected repository at index %d to be returned", index)
		}
	}

	if projectRepo.gotProjectID != "project-123" {
		t.Fatalf(
			"expected project ID project-123, got %s",
			projectRepo.gotProjectID,
		)
	}

	if memberRepo.gotWorkspaceID != "workspace-123" {
		t.Fatalf(
			"expected workspace ID workspace-123, got %s",
			memberRepo.gotWorkspaceID,
		)
	}

	if memberRepo.gotUserID != "user-123" {
		t.Fatalf(
			"expected user ID user-123, got %s",
			memberRepo.gotUserID,
		)
	}

	if repositoryRepo.gotProjectID != "project-123" {
		t.Fatalf(
			"expected repository project ID project-123, got %s",
			repositoryRepo.gotProjectID,
		)
	}
}

func TestRepositoryServiceListByProjectIDUnauthorized(t *testing.T) {
	repositoryRepo := &fakeRepositoryRepository{}
	projectRepo := &fakeProjectRepository{
		project: testProject(),
	}
	memberRepo := &fakeProjectMemberRepository{
		err: repository.ErrWorkspaceMemberNotFound,
	}
	service := NewRepositoryService(repositoryRepo, projectRepo, memberRepo)

	_, err := service.ListByProjectID(
		context.Background(),
		"user-999",
		"project-123",
	)

	if !errors.Is(err, ErrRepositoryUnauthorized) {
		t.Fatalf("expected ErrRepositoryUnauthorized, got %v", err)
	}

	if repositoryRepo.gotProjectID != "" {
		t.Fatal("expected repository list not to be called")
	}
}

func TestRepositoryServiceListByProjectIDRepositoryError(t *testing.T) {
	repositoryErr := errors.New("database failure")
	repositoryRepo := &fakeRepositoryRepository{
		listErr: repositoryErr,
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
	service := NewRepositoryService(repositoryRepo, projectRepo, memberRepo)

	_, err := service.ListByProjectID(
		context.Background(),
		"user-123",
		"project-123",
	)

	if !errors.Is(err, repositoryErr) {
		t.Fatalf("expected repository error, got %v", err)
	}
}

func TestRepositoryServiceDelete(t *testing.T) {
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
	service := NewRepositoryService(repositoryRepo, projectRepo, memberRepo)

	err := service.Delete(
		context.Background(),
		"user-123",
		"repository-123",
	)

	if err != nil {
		t.Fatalf("delete repository: %v", err)
	}

	if repositoryRepo.gotRepositoryID != "repository-123" {
		t.Fatalf(
			"expected repository ID repository-123, got %s",
			repositoryRepo.gotRepositoryID,
		)
	}

	if memberRepo.gotWorkspaceID != "workspace-123" {
		t.Fatalf(
			"expected workspace ID workspace-123, got %s",
			memberRepo.gotWorkspaceID,
		)
	}

	if memberRepo.gotUserID != "user-123" {
		t.Fatalf(
			"expected user ID user-123, got %s",
			memberRepo.gotUserID,
		)
	}
}

func TestRepositoryServiceDeleteUnauthorized(t *testing.T) {
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
		err: repository.ErrWorkspaceMemberNotFound,
	}
	service := NewRepositoryService(repositoryRepo, projectRepo, memberRepo)

	err := service.Delete(
		context.Background(),
		"user-999",
		"repository-123",
	)

	if !errors.Is(err, ErrRepositoryUnauthorized) {
		t.Fatalf("expected ErrRepositoryUnauthorized, got %v", err)
	}
}

func TestRepositoryServiceDeleteNotFound(t *testing.T) {
	repositoryRepo := &fakeRepositoryRepository{
		findByIDErr: repository.ErrRepositoryNotFound,
	}
	service := NewRepositoryService(
		repositoryRepo,
		&fakeProjectRepository{},
		&fakeProjectMemberRepository{},
	)

	err := service.Delete(
		context.Background(),
		"user-123",
		"repository-123",
	)

	if !errors.Is(err, repository.ErrRepositoryNotFound) {
		t.Fatalf(
			"expected repository not found error, got %v",
			err,
		)
	}
}

func TestRepositoryServiceDeleteRepositoryError(t *testing.T) {
	repositoryErr := errors.New("database failure")
	repositoryRepo := &fakeRepositoryRepository{
		repository: &domain.Repository{
			ID:        "repository-123",
			ProjectID: "project-123",
		},
		deleteErr: repositoryErr,
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
	service := NewRepositoryService(repositoryRepo, projectRepo, memberRepo)

	err := service.Delete(
		context.Background(),
		"user-123",
		"repository-123",
	)

	if !errors.Is(err, repositoryErr) {
		t.Fatalf("expected repository error, got %v", err)
	}
}
