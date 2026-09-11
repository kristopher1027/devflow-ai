
package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/kristopher1027/devflow-ai/internal/domain"
	"github.com/kristopher1027/devflow-ai/internal/repository"
)

var (
	ErrWorkspaceNameRequired = errors.New("workspace name is required")
)

type WorkspaceService interface {
	Create(
		ctx context.Context,
		ownerID string,
		name string,
	) (*domain.Workspace, error)

	FindByID(
		ctx context.Context,
		ownerID string,
		id string,
	) (*domain.Workspace, error)

	ListByOwnerID(
		ctx context.Context,
		ownerID string,
	) ([]*domain.Workspace, error)

	Delete(
		ctx context.Context,
		ownerID string,
		id string,
	) error
}

type WorkspaceServiceImpl struct {
	workspaces repository.WorkspaceRepository
	members    repository.WorkspaceMemberRepository
}

func NewWorkspaceService(
	workspaces repository.WorkspaceRepository,
) WorkspaceService {
	return &WorkspaceServiceImpl{
		workspaces: workspaces,
	}
}

func NewWorkspaceServiceWithMembers(
	workspaces repository.WorkspaceRepository,
	members repository.WorkspaceMemberRepository,
) WorkspaceService {
	return &WorkspaceServiceImpl{
		workspaces: workspaces,
		members:    members,
	}
}

func (s *WorkspaceServiceImpl) Create(
	ctx context.Context,
	ownerID string,
	name string,
) (*domain.Workspace, error) {
	name = strings.TrimSpace(name)

	if name == "" {
		return nil, ErrWorkspaceNameRequired
	}

	now := time.Now()

	workspace := &domain.Workspace{
		ID:        uuid.NewString(),
		OwnerID:   ownerID,
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.workspaces.Create(ctx, workspace); err != nil {
		return nil, err
	}

	if s.members != nil {
		member := &domain.WorkspaceMember{
			WorkspaceID: workspace.ID,
			UserID:      ownerID,
			Role:        WorkspaceMemberRoleOwner,
			CreatedAt:   now,
		}

		if err := s.members.Create(ctx, member); err != nil {
			return nil, err
		}
	}

	return workspace, nil
}

func (s *WorkspaceServiceImpl) FindByID(
	ctx context.Context,
	ownerID string,
	id string,
) (*domain.Workspace, error) {
	workspace, err := s.workspaces.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if workspace.OwnerID != ownerID {
		return nil, repository.ErrWorkspaceNotFound
	}

	return workspace, nil
}

func (s *WorkspaceServiceImpl) ListByOwnerID(
	ctx context.Context,
	ownerID string,
) ([]*domain.Workspace, error) {
	return s.workspaces.ListByOwnerID(ctx, ownerID)
}

func (s *WorkspaceServiceImpl) Delete(
	ctx context.Context,
	ownerID string,
	id string,
) error {
	workspace, err := s.workspaces.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if workspace.OwnerID != ownerID {
		return repository.ErrWorkspaceNotFound
	}

	return s.workspaces.Delete(ctx, id)
}
