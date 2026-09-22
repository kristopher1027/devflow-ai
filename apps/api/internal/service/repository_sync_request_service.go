package service

import (
	"context"
	"errors"

	"github.com/kristopher1027/devflow-ai/internal/domain"
	"github.com/kristopher1027/devflow-ai/internal/repository"
)

var (
	ErrRepositorySyncRequestUnauthorized = errors.New(
		"unauthorized to request repository sync",
	)
)

type RepositorySyncRequestService interface {
	RequestSync(
		ctx context.Context,
		requesterID string,
		repositoryID string,
	) (*domain.RepositorySyncJob, error)
}

type RepositorySyncRequestServiceImpl struct {
	repositories repository.RepositoryRepository
	projects     repository.ProjectRepository
	members      repository.WorkspaceMemberRepository
	queue        RepositorySyncQueue
}

func NewRepositorySyncRequestService(
	repositories repository.RepositoryRepository,
	projects repository.ProjectRepository,
	members repository.WorkspaceMemberRepository,
	queue RepositorySyncQueue,
) RepositorySyncRequestService {
	return &RepositorySyncRequestServiceImpl{
		repositories: repositories,
		projects:     projects,
		members:      members,
		queue:        queue,
	}
}

func (s *RepositorySyncRequestServiceImpl) RequestSync(
	ctx context.Context,
	requesterID string,
	repositoryID string,
) (*domain.RepositorySyncJob, error) {
	repo, err := s.repositories.FindByID(ctx, repositoryID)
	if err != nil {
		return nil, err
	}

	project, err := s.projects.FindByID(ctx, repo.ProjectID)
	if err != nil {
		return nil, err
	}

	member, err := s.members.Find(
		ctx,
		project.WorkspaceID,
		requesterID,
	)
	if err != nil {
		if errors.Is(err, repository.ErrWorkspaceMemberNotFound) {
			return nil, ErrRepositorySyncRequestUnauthorized
		}

		return nil, err
	}

	switch member.Role {
	case WorkspaceMemberRoleOwner,
		WorkspaceMemberRoleAdmin,
		WorkspaceMemberRoleMember:
		// Authorized.

	default:
		return nil, ErrRepositorySyncRequestUnauthorized
	}

	return s.queue.Enqueue(
		ctx,
		RepositorySyncRequest{
			RepositoryID: repositoryID,
		},
	)
}
