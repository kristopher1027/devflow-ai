package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/kristopher1027/devflow-ai/internal/domain"
	"github.com/kristopher1027/devflow-ai/internal/repository"
)

const (
	WorkspaceMemberRoleOwner  = "owner"
	WorkspaceMemberRoleAdmin  = "admin"
	WorkspaceMemberRoleMember = "member"
)

var (
	ErrWorkspaceMemberRoleRequired = errors.New(
		"workspace member role is required",
	)

	ErrInvalidWorkspaceMemberRole = errors.New(
		"invalid workspace member role",
	)

	ErrWorkspaceMemberUnauthorized = errors.New(
		"unauthorized to manage workspace members",
	)

	ErrWorkspaceOwnerCannotBeRemoved = errors.New(
		"workspace owner cannot be removed",
	)

	ErrWorkspaceMemberAdminCannotAssignRole = errors.New(
		"admins can only add members",
	)

	ErrWorkspaceMemberAdminCannotChangeRole = errors.New(
		"admins cannot change member roles",
	)
)

type WorkspaceMemberService interface {
	Add(
		ctx context.Context,
		requesterID string,
		workspaceID string,
		userID string,
		role string,
	) (*domain.WorkspaceMember, error)

	Find(
		ctx context.Context,
		requesterID string,
		workspaceID string,
		userID string,
	) (*domain.WorkspaceMember, error)

	ListByWorkspaceID(
		ctx context.Context,
		requesterID string,
		workspaceID string,
	) ([]*domain.WorkspaceMember, error)

	UpdateRole(
		ctx context.Context,
		requesterID string,
		workspaceID string,
		userID string,
		role string,
	) error

	Remove(
		ctx context.Context,
		requesterID string,
		workspaceID string,
		userID string,
	) error
}

type WorkspaceMemberServiceImpl struct {
	members repository.WorkspaceMemberRepository

	workspaces repository.WorkspaceRepository
}

func NewWorkspaceMemberService(
	members repository.WorkspaceMemberRepository,
	workspaces repository.WorkspaceRepository,
) WorkspaceMemberService {

	return &WorkspaceMemberServiceImpl{

		members: members,

		workspaces: workspaces,
	}

}

func isValidWorkspaceMemberRole(
	role string,
) bool {

	switch role {

	case WorkspaceMemberRoleOwner,
		WorkspaceMemberRoleAdmin,
		WorkspaceMemberRoleMember:

		return true

	default:

		return false
	}

}
func (s *WorkspaceMemberServiceImpl) authorizeManager(
	ctx context.Context,
	requesterID string,
	workspaceID string,
) (*domain.Workspace, *domain.WorkspaceMember, error) {

	workspace, err := s.workspaces.FindByID(
		ctx,
		workspaceID,
	)

	if err != nil {
		return nil, nil, err
	}

	// Workspace owner automatically has permission.
	if workspace.OwnerID == requesterID {

		return workspace,
			&domain.WorkspaceMember{
				WorkspaceID: workspaceID,
				UserID:      requesterID,
				Role:        WorkspaceMemberRoleOwner,
			},
			nil
	}

	member, err := s.members.Find(
		ctx,
		workspaceID,
		requesterID,
	)

	if err != nil {

		if errors.Is(
			err,
			repository.ErrWorkspaceMemberNotFound,
		) {

			return nil, nil,
				ErrWorkspaceMemberUnauthorized
		}

		return nil, nil, err
	}

	if member.Role != WorkspaceMemberRoleAdmin {

		return nil, nil,
			ErrWorkspaceMemberUnauthorized
	}

	return workspace, member, nil
}

func (s *WorkspaceMemberServiceImpl) Add(
	ctx context.Context,
	requesterID string,
	workspaceID string,
	userID string,
	role string,
) (*domain.WorkspaceMember, error) {

	role = strings.TrimSpace(role)

	if role == "" {

		return nil,
			ErrWorkspaceMemberRoleRequired
	}

	if !isValidWorkspaceMemberRole(role) {

		return nil,
			ErrInvalidWorkspaceMemberRole
	}

	_, requester, err :=
		s.authorizeManager(
			ctx,
			requesterID,
			workspaceID,
		)

	if err != nil {

		return nil, err
	}

	// Admins cannot create admins or owners.
	if requester.Role == WorkspaceMemberRoleAdmin &&
		role != WorkspaceMemberRoleMember {

		return nil,
			ErrWorkspaceMemberAdminCannotAssignRole
	}

	member := &domain.WorkspaceMember{

		WorkspaceID: workspaceID,

		UserID: userID,

		Role: role,

		CreatedAt: time.Now(),
	}

	err = s.members.Create(
		ctx,
		member,
	)

	if err != nil {

		return nil, err
	}

	return member, nil
}

func (s *WorkspaceMemberServiceImpl) Find(
	ctx context.Context,
	requesterID string,
	workspaceID string,
	userID string,
) (*domain.WorkspaceMember, error) {

	_, _, err :=
		s.authorizeManager(
			ctx,
			requesterID,
			workspaceID,
		)

	if err != nil {

		return nil, err
	}

	return s.members.Find(
		ctx,
		workspaceID,
		userID,
	)

}

func (s *WorkspaceMemberServiceImpl) ListByWorkspaceID(
	ctx context.Context,
	requesterID string,
	workspaceID string,
) ([]*domain.WorkspaceMember, error) {

	_, _, err :=
		s.authorizeManager(
			ctx,
			requesterID,
			workspaceID,
		)

	if err != nil {

		return nil, err
	}

	return s.members.ListByWorkspaceID(
		ctx,
		workspaceID,
	)

}

func (s *WorkspaceMemberServiceImpl) UpdateRole(
	ctx context.Context,
	requesterID string,
	workspaceID string,
	userID string,
	role string,
) error {

	role = strings.TrimSpace(role)

	if role == "" {

		return ErrWorkspaceMemberRoleRequired
	}

	if !isValidWorkspaceMemberRole(role) {

		return ErrInvalidWorkspaceMemberRole
	}

	workspace, requester, err :=
		s.authorizeManager(
			ctx,
			requesterID,
			workspaceID,
		)

	if err != nil {

		return err
	}

	// Owner role cannot be changed.
	if userID == workspace.OwnerID {

		return ErrWorkspaceOwnerCannotBeRemoved
	}

	target, err :=
		s.members.Find(
			ctx,
			workspaceID,
			userID,
		)

	if err != nil {

		return err
	}

	// Admin restrictions.
	if requester.Role == WorkspaceMemberRoleAdmin {

		// Admin cannot promote.
		if role != WorkspaceMemberRoleMember {

			return ErrWorkspaceMemberAdminCannotChangeRole
		}

		// Admin cannot modify another admin.
		if target.Role == WorkspaceMemberRoleAdmin {

			return ErrWorkspaceMemberAdminCannotChangeRole
		}
	}

	return s.members.UpdateRole(
		ctx,
		workspaceID,
		userID,
		role,
	)

}

func (s *WorkspaceMemberServiceImpl) Remove(
	ctx context.Context,
	requesterID string,
	workspaceID string,
	userID string,
) error {

	workspace, requester, err :=
		s.authorizeManager(
			ctx,
			requesterID,
			workspaceID,
		)

	if err != nil {

		return err
	}

	// Owner can never be removed.
	if userID == workspace.OwnerID {

		return ErrWorkspaceOwnerCannotBeRemoved
	}

	if requester.Role == WorkspaceMemberRoleAdmin {

		member, err :=
			s.members.Find(
				ctx,
				workspaceID,
				userID,
			)

		if err != nil {

			return err
		}

		// Admin can remove only normal members.
		if member.Role != WorkspaceMemberRoleMember {

			return ErrWorkspaceMemberUnauthorized
		}

	}

	return s.members.Delete(
		ctx,
		workspaceID,
		userID,
	)

}
