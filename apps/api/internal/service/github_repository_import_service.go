package service

import (
	"context"
	"errors"

	"github.com/kristopher1027/devflow-ai/internal/domain"
	githubintegration "github.com/kristopher1027/devflow-ai/internal/integration/github"
	"github.com/kristopher1027/devflow-ai/internal/repository"
)

type GitHubRepositoryImportService interface {
	ImportRepositories(
		ctx context.Context,
		requesterID string,
		projectID string,
	) ([]*domain.Repository, error)
}

type GitHubRepositoryImportServiceImpl struct {
	projects     repository.ProjectRepository
	connections  GitHubConnectionService
	github       githubintegration.Client
	repositories RepositoryService
}

func NewGitHubRepositoryImportService(
	projects repository.ProjectRepository,
	connections GitHubConnectionService,
	github githubintegration.Client,
	repositories RepositoryService,
) GitHubRepositoryImportService {
	return &GitHubRepositoryImportServiceImpl{
		projects:     projects,
		connections:  connections,
		github:       github,
		repositories: repositories,
	}
}

func (s *GitHubRepositoryImportServiceImpl) ImportRepositories(
	ctx context.Context,
	requesterID string,
	projectID string,
) ([]*domain.Repository, error) {
	project, err := s.projects.FindByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	connection, err := s.connections.FindByWorkspaceID(
		ctx,
		requesterID,
		project.WorkspaceID,
	)
	if err != nil {
		return nil, err
	}

	githubRepositories, err := s.github.ListRepositories(
		ctx,
		connection.InstallationID,
	)
	if err != nil {
		return nil, err
	}

	imported := make([]*domain.Repository, 0, len(githubRepositories))
	for _, githubRepository := range githubRepositories {
		created, err := s.repositories.Create(
			ctx,
			requesterID,
			projectID,
			"github",
			githubRepository.ExternalID,
			githubRepository.Owner,
			githubRepository.Name,
			githubRepository.FullName,
			githubRepository.DefaultBranch,
			githubRepository.HTMLURL,
			githubRepository.CloneURL,
			githubRepository.IsPrivate,
		)
		if errors.Is(err, ErrRepositoryAlreadyExists) {
			continue
		}
		if err != nil {
			return nil, err
		}

		imported = append(imported, created)
	}

	return imported, nil
}
