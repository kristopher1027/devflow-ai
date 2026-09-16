package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kristopher1027/devflow-ai/internal/domain"
	"github.com/kristopher1027/devflow-ai/internal/repository"
)

type fakeProjectRepository struct {
	project  *domain.Project
	projects []*domain.Project
	err      error

	gotProjectID   string
	gotWorkspaceID string
	gotName        string
	gotDescription *string
	gotCreatedBy   string
}

func (f *fakeProjectRepository) Create(
	ctx context.Context,
	project *domain.Project,
) error {
	f.gotProjectID = project.ID
	f.gotWorkspaceID = project.WorkspaceID
	f.gotName = project.Name
	f.gotDescription = project.Description
	f.gotCreatedBy = project.CreatedBy

	if f.err != nil {
		return f.err
	}

	f.project = project

	return nil
}

func (f *fakeProjectRepository) FindByID(
	ctx context.Context,
	id string,
) (*domain.Project, error) {
	f.gotProjectID = id

	if f.err != nil {
		return nil, f.err
	}

	return f.project, nil
}

func (f *fakeProjectRepository) ListByWorkspaceID(
	ctx context.Context,
	workspaceID string,
) ([]*domain.Project, error) {
	f.gotWorkspaceID = workspaceID

	if f.err != nil {
		return nil, f.err
	}

	return f.projects, nil
}

func (f *fakeProjectRepository) Delete(
	ctx context.Context,
	id string,
) error {
	f.gotProjectID = id

	return f.err
}

type fakeProjectMemberRepository struct {
	member *domain.WorkspaceMember
	err    error

	membersByUserID map[string]*domain.WorkspaceMember

	gotWorkspaceID string
	gotUserID      string
}

func (f *fakeProjectMemberRepository) Create(
	ctx context.Context,
	member *domain.WorkspaceMember,
) error {
	return nil
}

func (f *fakeProjectMemberRepository) Find(
	ctx context.Context,
	workspaceID string,
	userID string,
) (*domain.WorkspaceMember, error) {
	f.gotWorkspaceID = workspaceID
	f.gotUserID = userID

	if f.err != nil {
		return nil, f.err
	}

	if f.membersByUserID != nil {
		member, ok := f.membersByUserID[userID]
		if !ok {
			return nil, repository.ErrWorkspaceMemberNotFound
		}

		return member, nil
	}

	return f.member, nil
}

func (f *fakeProjectMemberRepository) ListByWorkspaceID(
	ctx context.Context,
	workspaceID string,
) ([]*domain.WorkspaceMember, error) {
	return nil, nil
}

func (f *fakeProjectMemberRepository) UpdateRole(
	ctx context.Context,
	workspaceID string,
	userID string,
	role string,
) error {
	return nil
}

func (f *fakeProjectMemberRepository) Delete(
	ctx context.Context,
	workspaceID string,
	userID string,
) error {
	return nil
}

func testProject() *domain.Project {
	now := time.Now()

	return &domain.Project{
		ID:          "project-123",
		WorkspaceID: "workspace-123",
		Name:        "Test Project",
		CreatedBy:   "user-123",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func testProjectMember(
	userID string,
	role string,
) *domain.WorkspaceMember {
	return &domain.WorkspaceMember{
		WorkspaceID: "workspace-123",
		UserID:      userID,
		Role:        role,
		CreatedAt:   time.Now(),
	}
}

func TestProjectServiceCreate(t *testing.T) {
	projectRepo := &fakeProjectRepository{}

	memberRepo := &fakeProjectMemberRepository{
		member: testProjectMember(
			"user-123",
			WorkspaceMemberRoleMember,
		),
	}

	service := NewProjectService(
		projectRepo,
		memberRepo,
	)

	description := "A developer productivity project"

	project, err := service.Create(
		context.Background(),
		"user-123",
		"workspace-123",
		"Test Project",
		&description,
	)

	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	if project == nil {
		t.Fatal("expected project, got nil")
	}

	if project.ID == "" {
		t.Fatal("expected project ID to be generated")
	}

	if project.WorkspaceID != "workspace-123" {
		t.Fatalf(
			"expected workspace ID workspace-123, got %s",
			project.WorkspaceID,
		)
	}

	if project.Name != "Test Project" {
		t.Fatalf(
			"expected project name Test Project, got %s",
			project.Name,
		)
	}

	if project.CreatedBy != "user-123" {
		t.Fatalf(
			"expected created by user-123, got %s",
			project.CreatedBy,
		)
	}

	if project.Description == nil {
		t.Fatal("expected description")
	}

	if *project.Description != description {
		t.Fatalf(
			"expected description %q, got %q",
			description,
			*project.Description,
		)
	}

	if projectRepo.gotWorkspaceID != "workspace-123" {
		t.Fatalf(
			"expected repository workspace ID workspace-123, got %s",
			projectRepo.gotWorkspaceID,
		)
	}

	if projectRepo.gotCreatedBy != "user-123" {
		t.Fatalf(
			"expected repository created by user-123, got %s",
			projectRepo.gotCreatedBy,
		)
	}
}

func TestProjectServiceCreateNameRequired(t *testing.T) {
	projectRepo := &fakeProjectRepository{}

	memberRepo := &fakeProjectMemberRepository{
		member: testProjectMember(
			"user-123",
			WorkspaceMemberRoleMember,
		),
	}

	service := NewProjectService(
		projectRepo,
		memberRepo,
	)

	_, err := service.Create(
		context.Background(),
		"user-123",
		"workspace-123",
		"   ",
		nil,
	)

	if !errors.Is(err, ErrProjectNameRequired) {
		t.Fatalf(
			"expected ErrProjectNameRequired, got %v",
			err,
		)
	}
}

func TestProjectServiceCreateUnauthorized(t *testing.T) {
	projectRepo := &fakeProjectRepository{}

	memberRepo := &fakeProjectMemberRepository{
		err: repository.ErrWorkspaceMemberNotFound,
	}

	service := NewProjectService(
		projectRepo,
		memberRepo,
	)

	_, err := service.Create(
		context.Background(),
		"user-999",
		"workspace-123",
		"Test Project",
		nil,
	)

	if !errors.Is(err, ErrProjectUnauthorized) {
		t.Fatalf(
			"expected ErrProjectUnauthorized, got %v",
			err,
		)
	}
}

func TestProjectServiceCreateRepositoryError(t *testing.T) {
	repositoryErr := errors.New("database failure")

	projectRepo := &fakeProjectRepository{
		err: repositoryErr,
	}

	memberRepo := &fakeProjectMemberRepository{
		member: testProjectMember(
			"user-123",
			WorkspaceMemberRoleMember,
		),
	}

	service := NewProjectService(
		projectRepo,
		memberRepo,
	)

	_, err := service.Create(
		context.Background(),
		"user-123",
		"workspace-123",
		"Test Project",
		nil,
	)

	if !errors.Is(err, repositoryErr) {
		t.Fatalf(
			"expected repository error, got %v",
			err,
		)
	}
}

func TestProjectServiceFindByID(t *testing.T) {
	projectRepo := &fakeProjectRepository{
		project: testProject(),
	}

	memberRepo := &fakeProjectMemberRepository{
		member: testProjectMember(
			"user-123",
			WorkspaceMemberRoleMember,
		),
	}

	service := NewProjectService(
		projectRepo,
		memberRepo,
	)

	project, err := service.FindByID(
		context.Background(),
		"user-123",
		"project-123",
	)

	if err != nil {
		t.Fatalf("find project: %v", err)
	}

	if project.ID != "project-123" {
		t.Fatalf(
			"expected project ID project-123, got %s",
			project.ID,
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

func TestProjectServiceFindByIDUnauthorized(t *testing.T) {
	projectRepo := &fakeProjectRepository{
		project: testProject(),
	}

	memberRepo := &fakeProjectMemberRepository{
		err: repository.ErrWorkspaceMemberNotFound,
	}

	service := NewProjectService(
		projectRepo,
		memberRepo,
	)

	_, err := service.FindByID(
		context.Background(),
		"user-999",
		"project-123",
	)

	if !errors.Is(err, ErrProjectUnauthorized) {
		t.Fatalf(
			"expected ErrProjectUnauthorized, got %v",
			err,
		)
	}
}

func TestProjectServiceFindByIDNotFound(t *testing.T) {
	projectRepo := &fakeProjectRepository{
		err: repository.ErrProjectNotFound,
	}

	memberRepo := &fakeProjectMemberRepository{}

	service := NewProjectService(
		projectRepo,
		memberRepo,
	)

	_, err := service.FindByID(
		context.Background(),
		"user-123",
		"project-123",
	)

	if !errors.Is(err, repository.ErrProjectNotFound) {
		t.Fatalf(
			"expected ErrProjectNotFound, got %v",
			err,
		)
	}
}

func TestProjectServiceListByWorkspaceID(t *testing.T) {
	projects := []*domain.Project{
		testProject(),
		{
			ID:          "project-456",
			WorkspaceID: "workspace-123",
			Name:        "Second Project",
			CreatedBy:   "user-123",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	projectRepo := &fakeProjectRepository{
		projects: projects,
	}

	memberRepo := &fakeProjectMemberRepository{
		member: testProjectMember(
			"user-123",
			WorkspaceMemberRoleMember,
		),
	}

	service := NewProjectService(
		projectRepo,
		memberRepo,
	)

	result, err := service.ListByWorkspaceID(
		context.Background(),
		"user-123",
		"workspace-123",
	)

	if err != nil {
		t.Fatalf("list projects: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf(
			"expected 2 projects, got %d",
			len(result),
		)
	}

	if projectRepo.gotWorkspaceID != "workspace-123" {
		t.Fatalf(
			"expected workspace ID workspace-123, got %s",
			projectRepo.gotWorkspaceID,
		)
	}
}

func TestProjectServiceListByWorkspaceIDUnauthorized(t *testing.T) {
	projectRepo := &fakeProjectRepository{}

	memberRepo := &fakeProjectMemberRepository{
		err: repository.ErrWorkspaceMemberNotFound,
	}

	service := NewProjectService(
		projectRepo,
		memberRepo,
	)

	_, err := service.ListByWorkspaceID(
		context.Background(),
		"user-999",
		"workspace-123",
	)

	if !errors.Is(err, ErrProjectUnauthorized) {
		t.Fatalf(
			"expected ErrProjectUnauthorized, got %v",
			err,
		)
	}
}

func TestProjectServiceDelete(t *testing.T) {
	projectRepo := &fakeProjectRepository{
		project: testProject(),
	}

	memberRepo := &fakeProjectMemberRepository{
		member: testProjectMember(
			"user-123",
			WorkspaceMemberRoleMember,
		),
	}

	service := NewProjectService(
		projectRepo,
		memberRepo,
	)

	err := service.Delete(
		context.Background(),
		"user-123",
		"project-123",
	)

	if err != nil {
		t.Fatalf("delete project: %v", err)
	}

	if projectRepo.gotProjectID != "project-123" {
		t.Fatalf(
			"expected project ID project-123, got %s",
			projectRepo.gotProjectID,
		)
	}
}

func TestProjectServiceDeleteUnauthorized(t *testing.T) {
	projectRepo := &fakeProjectRepository{
		project: testProject(),
	}

	memberRepo := &fakeProjectMemberRepository{
		err: repository.ErrWorkspaceMemberNotFound,
	}

	service := NewProjectService(
		projectRepo,
		memberRepo,
	)

	err := service.Delete(
		context.Background(),
		"user-999",
		"project-123",
	)

	if !errors.Is(err, ErrProjectUnauthorized) {
		t.Fatalf(
			"expected ErrProjectUnauthorized, got %v",
			err,
		)
	}
}

func TestProjectServiceDeleteNotFound(t *testing.T) {
	projectRepo := &fakeProjectRepository{
		err: repository.ErrProjectNotFound,
	}

	memberRepo := &fakeProjectMemberRepository{}

	service := NewProjectService(
		projectRepo,
		memberRepo,
	)

	err := service.Delete(
		context.Background(),
		"user-123",
		"project-123",
	)

	if !errors.Is(err, repository.ErrProjectNotFound) {
		t.Fatalf(
			"expected ErrProjectNotFound, got %v",
			err,
		)
	}
}
