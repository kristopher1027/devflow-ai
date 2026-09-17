package service

import (
	"context"
	"errors"

	"github.com/kristopher1027/devflow-ai/internal/domain"
	"github.com/kristopher1027/devflow-ai/internal/repository"
)

var (
	ErrGitHubRepositoryImportJobUnauthorized = errors.New(
		"unauthorized to view github repository import job",
	)
)

type GitHubRepositoryImportJobService interface {
	FindByID(
		ctx context.Context,
		requesterID string,
		id string,
	) (*domain.GitHubRepositoryImportJob, error)
}

type GitHubRepositoryImportJobServiceImpl struct {
	jobs    repository.GitHubRepositoryImportJobRepository
	projects repository.ProjectRepository
	members repository.WorkspaceMemberRepository
}

func NewGitHubRepositoryImportJobService(
	jobs repository.GitHubRepositoryImportJobRepository,
	projects repository.ProjectRepository,
	members repository.WorkspaceMemberRepository,
) GitHubRepositoryImportJobService {
	return &GitHubRepositoryImportJobServiceImpl{
		jobs:     jobs,
		projects: projects,
		members:  members,
	}
}

func (s *GitHubRepositoryImportJobServiceImpl) FindByID(
	ctx context.Context,
	requesterID string,
	id string,
) (*domain.GitHubRepositoryImportJob, error) {
	job, err := s.jobs.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	project, err := s.projects.FindByID(ctx, job.ProjectID)
	if err != nil {
		return nil, err
	}

	member, err := s.members.Find(
		ctx,
		project.WorkspaceID,
		requesterID,
	)
	if err != nil {
		if errors.Is(
			err,
			repository.ErrWorkspaceMemberNotFound,
		) {
			return nil, ErrGitHubRepositoryImportJobUnauthorized
		}

		return nil, err
	}

	switch member.Role {
	case WorkspaceMemberRoleOwner,
		WorkspaceMemberRoleAdmin,
		WorkspaceMemberRoleMember:
		return job, nil

	default:
		return nil, ErrGitHubRepositoryImportJobUnauthorized
	}
}
