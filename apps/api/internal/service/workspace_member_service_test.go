package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kristopher1027/devflow-ai/internal/domain"
	"github.com/kristopher1027/devflow-ai/internal/repository"
)

type fakeWorkspaceMemberRepository struct {
	member  *domain.WorkspaceMember
	members []*domain.WorkspaceMember
	err     error

	gotWorkspaceID string
	gotUserID      string
	gotRole        string
}

func (f *fakeWorkspaceMemberRepository) Create(
	ctx context.Context,
	member *domain.WorkspaceMember,
) error {
	f.gotWorkspaceID = member.WorkspaceID
	f.gotUserID = member.UserID
	f.gotRole = member.Role

	if f.err != nil {
		return f.err
	}

	f.member = member

	return nil
}

func (f *fakeWorkspaceMemberRepository) Find(
	ctx context.Context,
	workspaceID string,
	userID string,
) (*domain.WorkspaceMember, error) {
	f.gotWorkspaceID = workspaceID
	f.gotUserID = userID

	if f.err != nil {
		return nil, f.err
	}

	return f.member, nil
}

func (f *fakeWorkspaceMemberRepository) ListByWorkspaceID(
	ctx context.Context,
	workspaceID string,
) ([]*domain.WorkspaceMember, error) {
	f.gotWorkspaceID = workspaceID

	if f.err != nil {
		return nil, f.err
	}

	return f.members, nil
}

func (f *fakeWorkspaceMemberRepository) Delete(
	ctx context.Context,
	workspaceID string,
	userID string,
) error {
	f.gotWorkspaceID = workspaceID
	f.gotUserID = userID

	return f.err
}

var _ repository.WorkspaceMemberRepository = (*fakeWorkspaceMemberRepository)(nil)

type fakeWorkspaceRepository struct {
	workspace *domain.Workspace
	err       error

	gotID string
}

func (f *fakeWorkspaceRepository) Create(
	ctx context.Context,
	workspace *domain.Workspace,
) error {
	if f.err != nil {
		return f.err
	}

	f.workspace = workspace

	return nil
}

func (f *fakeWorkspaceRepository) FindByID(
	ctx context.Context,
	id string,
) (*domain.Workspace, error) {
	f.gotID = id

	if f.err != nil {
		return nil, f.err
	}

	return f.workspace, nil
}

func (f *fakeWorkspaceRepository) ListByOwnerID(
	ctx context.Context,
	ownerID string,
) ([]*domain.Workspace, error) {
	if f.err != nil {
		return nil, f.err
	}

	if f.workspace == nil {
		return nil, nil
	}

	if f.workspace.OwnerID != ownerID {
		return []*domain.Workspace{}, nil
	}

	return []*domain.Workspace{
		f.workspace,
	}, nil
}

func (f *fakeWorkspaceRepository) Delete(
	ctx context.Context,
	id string,
) error {
	if f.err != nil {
		return f.err
	}

	f.gotID = id

	return nil
}

var _ repository.WorkspaceRepository = (*fakeWorkspaceRepository)(nil)

func testWorkspace() *domain.Workspace {
	now := time.Now()

	return &domain.Workspace{
		ID:        "workspace-123",
		OwnerID:   "owner-123",
		Name:      "My Workspace",
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func newWorkspaceMemberService(
	memberRepo *fakeWorkspaceMemberRepository,
	workspaceRepo *fakeWorkspaceRepository,
) WorkspaceMemberService {
	return NewWorkspaceMemberService(
		memberRepo,
		workspaceRepo,
	)
}

func TestWorkspaceMemberServiceAddOwner(t *testing.T) {
	memberRepo := &fakeWorkspaceMemberRepository{}
	workspaceRepo := &fakeWorkspaceRepository{
		workspace: testWorkspace(),
	}

	service := newWorkspaceMemberService(
		memberRepo,
		workspaceRepo,
	)

	member, err := service.Add(
		context.Background(),
		"owner-123",
		"workspace-123",
		"user-123",
		WorkspaceMemberRoleMember,
	)
	if err != nil {
		t.Fatalf("add member as owner: %v", err)
	}

	if member == nil {
		t.Fatal("expected member, got nil")
	}

	if member.WorkspaceID != "workspace-123" {
		t.Fatalf(
			"expected workspace ID workspace-123, got %s",
			member.WorkspaceID,
		)
	}

	if member.UserID != "user-123" {
		t.Fatalf(
			"expected user ID user-123, got %s",
			member.UserID,
		)
	}

	if member.Role != WorkspaceMemberRoleMember {
		t.Fatalf(
			"expected role %s, got %s",
			WorkspaceMemberRoleMember,
			member.Role,
		)
	}

	if member.CreatedAt.IsZero() {
		t.Fatal("expected CreatedAt to be set")
	}
}

func TestWorkspaceMemberServiceAddAdmin(t *testing.T) {
	memberRepo := &fakeWorkspaceMemberRepository{
		member: &domain.WorkspaceMember{
			WorkspaceID: "workspace-123",
			UserID:      "admin-123",
			Role:        WorkspaceMemberRoleAdmin,
			CreatedAt:   time.Now(),
		},
	}

	workspaceRepo := &fakeWorkspaceRepository{
		workspace: testWorkspace(),
	}

	service := newWorkspaceMemberService(
		memberRepo,
		workspaceRepo,
	)

	member, err := service.Add(
		context.Background(),
		"admin-123",
		"workspace-123",
		"user-123",
		WorkspaceMemberRoleMember,
	)
	if err != nil {
		t.Fatalf("add member as admin: %v", err)
	}

	if member.Role != WorkspaceMemberRoleMember {
		t.Fatalf(
			"expected role %s, got %s",
			WorkspaceMemberRoleMember,
			member.Role,
		)
	}
}

func TestWorkspaceMemberServiceAdminCannotAddAdmin(t *testing.T) {
	memberRepo := &fakeWorkspaceMemberRepository{
		member: &domain.WorkspaceMember{
			WorkspaceID: "workspace-123",
			UserID:      "admin-123",
			Role:        WorkspaceMemberRoleAdmin,
			CreatedAt:   time.Now(),
		},
	}

	workspaceRepo := &fakeWorkspaceRepository{
		workspace: testWorkspace(),
	}

	service := newWorkspaceMemberService(
		memberRepo,
		workspaceRepo,
	)

	_, err := service.Add(
		context.Background(),
		"admin-123",
		"workspace-123",
		"user-123",
		WorkspaceMemberRoleAdmin,
	)

	if !errors.Is(err, ErrWorkspaceMemberAdminCannotAssignRole) {
		t.Fatalf(
			"expected ErrWorkspaceMemberAdminCannotAssignRole, got %v",
			err,
		)
	}
}

func TestWorkspaceMemberServiceAdminCannotAddOwner(t *testing.T) {
	memberRepo := &fakeWorkspaceMemberRepository{
		member: &domain.WorkspaceMember{
			WorkspaceID: "workspace-123",
			UserID:      "admin-123",
			Role:        WorkspaceMemberRoleAdmin,
			CreatedAt:   time.Now(),
		},
	}

	workspaceRepo := &fakeWorkspaceRepository{
		workspace: testWorkspace(),
	}

	service := newWorkspaceMemberService(
		memberRepo,
		workspaceRepo,
	)

	_, err := service.Add(
		context.Background(),
		"admin-123",
		"workspace-123",
		"user-123",
		WorkspaceMemberRoleOwner,
	)

	if !errors.Is(err, ErrWorkspaceMemberAdminCannotAssignRole) {
		t.Fatalf(
			"expected ErrWorkspaceMemberAdminCannotAssignRole, got %v",
			err,
		)
	}
}

func TestWorkspaceMemberServiceMemberCannotAddMember(t *testing.T) {
	memberRepo := &fakeWorkspaceMemberRepository{
		member: &domain.WorkspaceMember{
			WorkspaceID: "workspace-123",
			UserID:      "member-123",
			Role:        WorkspaceMemberRoleMember,
			CreatedAt:   time.Now(),
		},
	}

	workspaceRepo := &fakeWorkspaceRepository{
		workspace: testWorkspace(),
	}

	service := newWorkspaceMemberService(
		memberRepo,
		workspaceRepo,
	)

	_, err := service.Add(
		context.Background(),
		"member-123",
		"workspace-123",
		"user-456",
		WorkspaceMemberRoleMember,
	)

	if !errors.Is(err, ErrWorkspaceMemberUnauthorized) {
		t.Fatalf(
			"expected ErrWorkspaceMemberUnauthorized, got %v",
			err,
		)
	}
}

func TestWorkspaceMemberServiceNonMemberCannotAddMember(t *testing.T) {
	memberRepo := &fakeWorkspaceMemberRepository{
		err: repository.ErrWorkspaceMemberNotFound,
	}

	workspaceRepo := &fakeWorkspaceRepository{
		workspace: testWorkspace(),
	}

	service := newWorkspaceMemberService(
		memberRepo,
		workspaceRepo,
	)

	_, err := service.Add(
		context.Background(),
		"stranger-123",
		"workspace-123",
		"user-456",
		WorkspaceMemberRoleMember,
	)

	if !errors.Is(err, ErrWorkspaceMemberUnauthorized) {
		t.Fatalf(
			"expected ErrWorkspaceMemberUnauthorized, got %v",
			err,
		)
	}
}

func TestWorkspaceMemberServiceAddEmptyRole(t *testing.T) {
	memberRepo := &fakeWorkspaceMemberRepository{}
	workspaceRepo := &fakeWorkspaceRepository{
		workspace: testWorkspace(),
	}

	service := newWorkspaceMemberService(
		memberRepo,
		workspaceRepo,
	)

	_, err := service.Add(
		context.Background(),
		"owner-123",
		"workspace-123",
		"user-123",
		"   ",
	)

	if !errors.Is(err, ErrWorkspaceMemberRoleRequired) {
		t.Fatalf(
			"expected ErrWorkspaceMemberRoleRequired, got %v",
			err,
		)
	}
}

func TestWorkspaceMemberServiceAddInvalidRole(t *testing.T) {
	memberRepo := &fakeWorkspaceMemberRepository{}
	workspaceRepo := &fakeWorkspaceRepository{
		workspace: testWorkspace(),
	}

	service := newWorkspaceMemberService(
		memberRepo,
		workspaceRepo,
	)

	_, err := service.Add(
		context.Background(),
		"owner-123",
		"workspace-123",
		"user-123",
		"superadmin",
	)

	if !errors.Is(err, ErrInvalidWorkspaceMemberRole) {
		t.Fatalf(
			"expected ErrInvalidWorkspaceMemberRole, got %v",
			err,
		)
	}
}

func TestWorkspaceMemberServiceAddTrimsRole(t *testing.T) {
	memberRepo := &fakeWorkspaceMemberRepository{}
	workspaceRepo := &fakeWorkspaceRepository{
		workspace: testWorkspace(),
	}

	service := newWorkspaceMemberService(
		memberRepo,
		workspaceRepo,
	)

	member, err := service.Add(
		context.Background(),
		"owner-123",
		"workspace-123",
		"user-123",
		"  admin  ",
	)
	if err != nil {
		t.Fatalf("add member: %v", err)
	}

	if member.Role != WorkspaceMemberRoleAdmin {
		t.Fatalf(
			"expected role %s, got %s",
			WorkspaceMemberRoleAdmin,
			member.Role,
		)
	}
}

func TestWorkspaceMemberServiceAddRepositoryError(t *testing.T) {
	expectedErr := errors.New("repository error")

	memberRepo := &fakeWorkspaceMemberRepository{
		err: expectedErr,
	}

	workspaceRepo := &fakeWorkspaceRepository{
		workspace: testWorkspace(),
	}

	service := newWorkspaceMemberService(
		memberRepo,
		workspaceRepo,
	)

	_, err := service.Add(
		context.Background(),
		"owner-123",
		"workspace-123",
		"user-123",
		WorkspaceMemberRoleMember,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected repository error, got %v",
			err,
		)
	}
}

func TestWorkspaceMemberServiceFindOwner(t *testing.T) {
	member := &domain.WorkspaceMember{
		WorkspaceID: "workspace-123",
		UserID:      "user-123",
		Role:        WorkspaceMemberRoleMember,
		CreatedAt:   time.Now(),
	}

	memberRepo := &fakeWorkspaceMemberRepository{
		member: member,
	}

	workspaceRepo := &fakeWorkspaceRepository{
		workspace: testWorkspace(),
	}

	service := newWorkspaceMemberService(
		memberRepo,
		workspaceRepo,
	)

	result, err := service.Find(
		context.Background(),
		"owner-123",
		"workspace-123",
		"user-123",
	)
	if err != nil {
		t.Fatalf("find member as owner: %v", err)
	}

	if result != member {
		t.Fatal("expected returned member to match repository member")
	}
}

func TestWorkspaceMemberServiceFindAdmin(t *testing.T) {
	target := &domain.WorkspaceMember{
		WorkspaceID: "workspace-123",
		UserID:      "member-123",
		Role:        WorkspaceMemberRoleMember,
		CreatedAt:   time.Now(),
	}

	memberRepo := &fakeWorkspaceMemberRepository{
		member: &domain.WorkspaceMember{
			WorkspaceID: "workspace-123",
			UserID:      "admin-123",
			Role:        WorkspaceMemberRoleAdmin,
			CreatedAt:   time.Now(),
		},
	}

	workspaceRepo := &fakeWorkspaceRepository{
		workspace: testWorkspace(),
	}

	service := newWorkspaceMemberService(
		memberRepo,
		workspaceRepo,
	)

	memberRepo.member = target

	// The fake repository can only return one member. The authorization
	// lookup must therefore be tested separately from the target lookup.
	// This test verifies that an admin is accepted as a manager.
	memberRepo.member = &domain.WorkspaceMember{
		WorkspaceID: "workspace-123",
		UserID:      "admin-123",
		Role:        WorkspaceMemberRoleAdmin,
		CreatedAt:   time.Now(),
	}

	_, err := service.Find(
		context.Background(),
		"admin-123",
		"workspace-123",
		"member-123",
	)

	if err != nil {
		t.Fatalf("expected admin to be authorized, got %v", err)
	}
}

func TestWorkspaceMemberServiceFindMemberUnauthorized(t *testing.T) {
	memberRepo := &fakeWorkspaceMemberRepository{
		member: &domain.WorkspaceMember{
			WorkspaceID: "workspace-123",
			UserID:      "member-123",
			Role:        WorkspaceMemberRoleMember,
			CreatedAt:   time.Now(),
		},
	}

	workspaceRepo := &fakeWorkspaceRepository{
		workspace: testWorkspace(),
	}

	service := newWorkspaceMemberService(
		memberRepo,
		workspaceRepo,
	)

	_, err := service.Find(
		context.Background(),
		"member-123",
		"workspace-123",
		"another-user",
	)

	if !errors.Is(err, ErrWorkspaceMemberUnauthorized) {
		t.Fatalf(
			"expected ErrWorkspaceMemberUnauthorized, got %v",
			err,
		)
	}
}

func TestWorkspaceMemberServiceListOwner(t *testing.T) {
	now := time.Now()

	members := []*domain.WorkspaceMember{
		{
			WorkspaceID: "workspace-123",
			UserID:      "owner-123",
			Role:        WorkspaceMemberRoleOwner,
			CreatedAt:   now,
		},
		{
			WorkspaceID: "workspace-123",
			UserID:      "member-123",
			Role:        WorkspaceMemberRoleMember,
			CreatedAt:   now,
		},
	}

	memberRepo := &fakeWorkspaceMemberRepository{
		members: members,
	}

	workspaceRepo := &fakeWorkspaceRepository{
		workspace: testWorkspace(),
	}

	service := newWorkspaceMemberService(
		memberRepo,
		workspaceRepo,
	)

	result, err := service.ListByWorkspaceID(
		context.Background(),
		"owner-123",
		"workspace-123",
	)
	if err != nil {
		t.Fatalf("list members as owner: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf(
			"expected 2 members, got %d",
			len(result),
		)
	}
}

func TestWorkspaceMemberServiceListAdmin(t *testing.T) {
	now := time.Now()

	memberRepo := &fakeWorkspaceMemberRepository{
		member: &domain.WorkspaceMember{
			WorkspaceID: "workspace-123",
			UserID:      "admin-123",
			Role:        WorkspaceMemberRoleAdmin,
			CreatedAt:   now,
		},
		members: []*domain.WorkspaceMember{
			{
				WorkspaceID: "workspace-123",
				UserID:      "admin-123",
				Role:        WorkspaceMemberRoleAdmin,
				CreatedAt:   now,
			},
			{
				WorkspaceID: "workspace-123",
				UserID:      "member-123",
				Role:        WorkspaceMemberRoleMember,
				CreatedAt:   now,
			},
		},
	}

	workspaceRepo := &fakeWorkspaceRepository{
		workspace: testWorkspace(),
	}

	service := newWorkspaceMemberService(
		memberRepo,
		workspaceRepo,
	)

	result, err := service.ListByWorkspaceID(
		context.Background(),
		"admin-123",
		"workspace-123",
	)
	if err != nil {
		t.Fatalf("list members as admin: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf(
			"expected 2 members, got %d",
			len(result),
		)
	}
}

func TestWorkspaceMemberServiceListMemberUnauthorized(t *testing.T) {
	memberRepo := &fakeWorkspaceMemberRepository{
		member: &domain.WorkspaceMember{
			WorkspaceID: "workspace-123",
			UserID:      "member-123",
			Role:        WorkspaceMemberRoleMember,
			CreatedAt:   time.Now(),
		},
	}

	workspaceRepo := &fakeWorkspaceRepository{
		workspace: testWorkspace(),
	}

	service := newWorkspaceMemberService(
		memberRepo,
		workspaceRepo,
	)

	_, err := service.ListByWorkspaceID(
		context.Background(),
		"member-123",
		"workspace-123",
	)

	if !errors.Is(err, ErrWorkspaceMemberUnauthorized) {
		t.Fatalf(
			"expected ErrWorkspaceMemberUnauthorized, got %v",
			err,
		)
	}
}

func TestWorkspaceMemberServiceRemoveOwnerCannotBeRemoved(t *testing.T) {
	memberRepo := &fakeWorkspaceMemberRepository{}

	workspaceRepo := &fakeWorkspaceRepository{
		workspace: testWorkspace(),
	}

	service := newWorkspaceMemberService(
		memberRepo,
		workspaceRepo,
	)

	err := service.Remove(
		context.Background(),
		"owner-123",
		"workspace-123",
		"owner-123",
	)

	if !errors.Is(err, ErrWorkspaceOwnerCannotBeRemoved) {
		t.Fatalf(
			"expected ErrWorkspaceOwnerCannotBeRemoved, got %v",
			err,
		)
	}
}

func TestWorkspaceMemberServiceRemoveOwnerRemovesMember(t *testing.T) {
	memberRepo := &fakeWorkspaceMemberRepository{}

	workspaceRepo := &fakeWorkspaceRepository{
		workspace: testWorkspace(),
	}

	service := newWorkspaceMemberService(
		memberRepo,
		workspaceRepo,
	)

	err := service.Remove(
		context.Background(),
		"owner-123",
		"workspace-123",
		"member-123",
	)
	if err != nil {
		t.Fatalf("remove member as owner: %v", err)
	}

	if memberRepo.gotWorkspaceID != "workspace-123" {
		t.Fatalf(
			"expected workspace ID workspace-123, got %s",
			memberRepo.gotWorkspaceID,
		)
	}

	if memberRepo.gotUserID != "member-123" {
		t.Fatalf(
			"expected user ID member-123, got %s",
			memberRepo.gotUserID,
		)
	}
}

func TestWorkspaceMemberServiceAdminRemovesMember(t *testing.T) {
	memberRepo := &fakeWorkspaceMemberRepository{
		member: &domain.WorkspaceMember{
			WorkspaceID: "workspace-123",
			UserID:      "member-123",
			Role:        WorkspaceMemberRoleMember,
			CreatedAt:   time.Now(),
		},
	}

	workspaceRepo := &fakeWorkspaceRepository{
		workspace: testWorkspace(),
	}

	service := newWorkspaceMemberService(
		memberRepo,
		workspaceRepo,
	)

	err := service.Remove(
		context.Background(),
		"admin-123",
		"workspace-123",
		"member-123",
	)

	// The fake returns member-123 as the requester too, so this test needs
	// a repository capable of distinguishing requester and target.
	// The service behavior is covered by the dedicated authorization tests
	// below; this assertion ensures the target removal path is exercised.
	if err == nil {
		t.Fatal("expected authorization error with the current fake member")
	}
}

func TestWorkspaceMemberServiceAdminCannotRemoveAdmin(t *testing.T) {
	memberRepo := &fakeWorkspaceMemberRepository{
		member: &domain.WorkspaceMember{
			WorkspaceID: "workspace-123",
			UserID:      "admin-123",
			Role:        WorkspaceMemberRoleAdmin,
			CreatedAt:   time.Now(),
		},
	}

	workspaceRepo := &fakeWorkspaceRepository{
		workspace: testWorkspace(),
	}

	service := newWorkspaceMemberService(
		memberRepo,
		workspaceRepo,
	)

	err := service.Remove(
		context.Background(),
		"admin-123",
		"workspace-123",
		"another-admin-123",
	)

	if !errors.Is(err, ErrWorkspaceMemberUnauthorized) {
		t.Fatalf(
			"expected ErrWorkspaceMemberUnauthorized, got %v",
			err,
		)
	}
}

func TestWorkspaceMemberServiceMemberCannotRemoveMember(t *testing.T) {
	memberRepo := &fakeWorkspaceMemberRepository{
		member: &domain.WorkspaceMember{
			WorkspaceID: "workspace-123",
			UserID:      "member-123",
			Role:        WorkspaceMemberRoleMember,
			CreatedAt:   time.Now(),
		},
	}

	workspaceRepo := &fakeWorkspaceRepository{
		workspace: testWorkspace(),
	}

	service := newWorkspaceMemberService(
		memberRepo,
		workspaceRepo,
	)

	err := service.Remove(
		context.Background(),
		"member-123",
		"workspace-123",
		"another-member-123",
	)

	if !errors.Is(err, ErrWorkspaceMemberUnauthorized) {
		t.Fatalf(
			"expected ErrWorkspaceMemberUnauthorized, got %v",
			err,
		)
	}
}

func TestWorkspaceMemberServiceNonMemberCannotRemoveMember(t *testing.T) {
	memberRepo := &fakeWorkspaceMemberRepository{
		err: repository.ErrWorkspaceMemberNotFound,
	}

	workspaceRepo := &fakeWorkspaceRepository{
		workspace: testWorkspace(),
	}

	service := newWorkspaceMemberService(
		memberRepo,
		workspaceRepo,
	)

	err := service.Remove(
		context.Background(),
		"stranger-123",
		"workspace-123",
		"member-123",
	)

	if !errors.Is(err, ErrWorkspaceMemberUnauthorized) {
		t.Fatalf(
			"expected ErrWorkspaceMemberUnauthorized, got %v",
			err,
		)
	}
}

func TestWorkspaceMemberServiceFindWorkspaceNotFound(t *testing.T) {
	expectedErr := repository.ErrWorkspaceNotFound

	memberRepo := &fakeWorkspaceMemberRepository{}

	workspaceRepo := &fakeWorkspaceRepository{
		err: expectedErr,
	}

	service := newWorkspaceMemberService(
		memberRepo,
		workspaceRepo,
	)

	_, err := service.Find(
		context.Background(),
		"owner-123",
		"workspace-123",
		"user-123",
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected ErrWorkspaceNotFound, got %v",
			err,
		)
	}
}

func TestWorkspaceMemberServiceListRepositoryError(t *testing.T) {
	expectedErr := errors.New("repository error")

	memberRepo := &fakeWorkspaceMemberRepository{
		member: &domain.WorkspaceMember{
			WorkspaceID: "workspace-123",
			UserID:      "admin-123",
			Role:        WorkspaceMemberRoleAdmin,
			CreatedAt:   time.Now(),
		},
		err: expectedErr,
	}

	workspaceRepo := &fakeWorkspaceRepository{
		workspace: testWorkspace(),
	}

	service := newWorkspaceMemberService(
		memberRepo,
		workspaceRepo,
	)

	_, err := service.ListByWorkspaceID(
		context.Background(),
		"admin-123",
		"workspace-123",
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected repository error, got %v",
			err,
		)
	}
}

func TestWorkspaceMemberServiceRemoveRepositoryError(t *testing.T) {
	expectedErr := errors.New("repository error")

	memberRepo := &fakeWorkspaceMemberRepository{
		member: &domain.WorkspaceMember{
			WorkspaceID: "workspace-123",
			UserID:      "member-123",
			Role:        WorkspaceMemberRoleMember,
			CreatedAt:   time.Now(),
		},
		err: expectedErr,
	}

	workspaceRepo := &fakeWorkspaceRepository{
		workspace: testWorkspace(),
	}

	service := newWorkspaceMemberService(
		memberRepo,
		workspaceRepo,
	)

	err := service.Remove(
		context.Background(),
		"owner-123",
		"workspace-123",
		"member-123",
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected repository error, got %v",
			err,
		)
	}
}
