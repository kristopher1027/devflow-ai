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
	ErrProjectNameRequired = errors.New("project name is required")

	ErrProjectUnauthorized = errors.New(
		"unauthorized to manage project",
	)
)

type ProjectService interface {
	Create(
		ctx context.Context,
		requesterID string,
		workspaceID string,
		name string,
		description *string,
	) (*domain.Project, error)

	FindByID(
		ctx context.Context,
		requesterID string,
		id string,
	) (*domain.Project, error)

	ListByWorkspaceID(
		ctx context.Context,
		requesterID string,
		workspaceID string,
	) ([]*domain.Project, error)

	Delete(
		ctx context.Context,
		requesterID string,
		id string,
	) error
}

type ProjectServiceImpl struct {
	projects repository.ProjectRepository
	members  repository.WorkspaceMemberRepository
}

func NewProjectService(
	projects repository.ProjectRepository,
	members repository.WorkspaceMemberRepository,
) ProjectService {
	return &ProjectServiceImpl{
		projects: projects,
		members:  members,
	}
}

func (s *ProjectServiceImpl) authorizeWorkspaceMember(
	ctx context.Context,
	requesterID string,
	workspaceID string,
) error {
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
			return ErrProjectUnauthorized
		}

		return err
	}

	switch member.Role {
	case WorkspaceMemberRoleOwner,
		WorkspaceMemberRoleAdmin,
		WorkspaceMemberRoleMember:
		return nil

	default:
		return ErrProjectUnauthorized
	}
}

func (s *ProjectServiceImpl) Create(
	ctx context.Context,
	requesterID string,
	workspaceID string,
	name string,
	description *string,
) (*domain.Project, error) {
	name = strings.TrimSpace(name)

	if name == "" {
		return nil, ErrProjectNameRequired
	}

	if err := s.authorizeWorkspaceMember(
		ctx,
		requesterID,
		workspaceID,
	); err != nil {
		return nil, err
	}

	now := time.Now()

	project := &domain.Project{
		ID:          uuid.NewString(),
		WorkspaceID: workspaceID,
		Name:        name,
		Description: description,
		CreatedBy:   requesterID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.projects.Create(ctx, project); err != nil {
		return nil, err
	}

	return project, nil
}

func (s *ProjectServiceImpl) FindByID(
	ctx context.Context,
	requesterID string,
	id string,
) (*domain.Project, error) {
	project, err := s.projects.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := s.authorizeWorkspaceMember(
		ctx,
		requesterID,
		project.WorkspaceID,
	); err != nil {
		return nil, err
	}

	return project, nil
}

func (s *ProjectServiceImpl) ListByWorkspaceID(
	ctx context.Context,
	requesterID string,
	workspaceID string,
) ([]*domain.Project, error) {
	if err := s.authorizeWorkspaceMember(
		ctx,
		requesterID,
		workspaceID,
	); err != nil {
		return nil, err
	}

	return s.projects.ListByWorkspaceID(
		ctx,
		workspaceID,
	)
}

func (s *ProjectServiceImpl) Delete(
	ctx context.Context,
	requesterID string,
	id string,
) error {
	project, err := s.projects.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.authorizeWorkspaceMember(
		ctx,
		requesterID,
		project.WorkspaceID,
	); err != nil {
		return err
	}

	return s.projects.Delete(ctx, id)
}
