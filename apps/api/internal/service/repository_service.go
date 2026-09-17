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
	ErrRepositoryUnauthorized = errors.New(
		"unauthorized to manage repository",
	)
	ErrRepositoryProviderUnsupported = errors.New(
		"repository provider unsupported",
	)
	ErrRepositoryExternalIDRequired = errors.New(
		"repository external ID is required",
	)
	ErrRepositoryOwnerRequired = errors.New(
		"repository owner is required",
	)
	ErrRepositoryNameRequired = errors.New(
		"repository name is required",
	)
	ErrRepositoryFullNameRequired = errors.New(
		"repository full name is required",
	)
	ErrRepositoryDefaultBranchRequired = errors.New(
		"repository default branch is required",
	)
	ErrRepositoryHTMLURLRequired = errors.New(
		"repository HTML URL is required",
	)
	ErrRepositoryCloneURLRequired = errors.New(
		"repository clone URL is required",
	)
	ErrRepositoryAlreadyExists = errors.New(
		"repository already exists",
	)
)

type RepositoryService interface {
	Create(
		ctx context.Context,
		requesterID string,
		projectID string,
		provider string,
		externalID string,
		owner string,
		name string,
		fullName string,
		defaultBranch string,
		htmlURL string,
		cloneURL string,
		isPrivate bool,
	) (*domain.Repository, error)

	FindByID(
		ctx context.Context,
		requesterID string,
		id string,
	) (*domain.Repository, error)

	ListByProjectID(
		ctx context.Context,
		requesterID string,
		projectID string,
	) ([]*domain.Repository, error)

	Delete(
		ctx context.Context,
		requesterID string,
		id string,
	) error
}
type RepositoryServiceImpl struct {
	repositories repository.RepositoryRepository
	projects     repository.ProjectRepository
	members      repository.WorkspaceMemberRepository
}

func NewRepositoryService(
	repositories repository.RepositoryRepository,
	projects repository.ProjectRepository,
	members repository.WorkspaceMemberRepository,
) *RepositoryServiceImpl {
	return &RepositoryServiceImpl{
		repositories: repositories,
		projects:     projects,
		members:      members,
	}
}

func (s *RepositoryServiceImpl) authorizeProjectMember(
	ctx context.Context,
	requesterID string,
	projectID string,
) (*domain.Project, error) {
	project, err := s.projects.FindByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	member, err := s.members.Find(ctx, project.WorkspaceID, requesterID)
	if err != nil {
		if errors.Is(err, repository.ErrWorkspaceMemberNotFound) {
			return nil, ErrRepositoryUnauthorized
		}

		return nil, err
	}

	switch member.Role {
	case WorkspaceMemberRoleOwner,
		WorkspaceMemberRoleAdmin,
		WorkspaceMemberRoleMember:
		return project, nil

	default:
		return nil, ErrRepositoryUnauthorized
	}
}

func (s *RepositoryServiceImpl) Create(
	ctx context.Context,
	requesterID string,
	projectID string,
	provider string,
	externalID string,
	owner string,
	name string,
	fullName string,
	defaultBranch string,
	htmlURL string,
	cloneURL string,
	isPrivate bool,
) (*domain.Repository, error) {
	provider = strings.TrimSpace(provider)
	externalID = strings.TrimSpace(externalID)
	owner = strings.TrimSpace(owner)
	name = strings.TrimSpace(name)
	fullName = strings.TrimSpace(fullName)
	defaultBranch = strings.TrimSpace(defaultBranch)
	htmlURL = strings.TrimSpace(htmlURL)
	cloneURL = strings.TrimSpace(cloneURL)

	if provider != "github" {
		return nil, ErrRepositoryProviderUnsupported
	}

	switch {
	case externalID == "":
		return nil, ErrRepositoryExternalIDRequired
	case owner == "":
		return nil, ErrRepositoryOwnerRequired
	case name == "":
		return nil, ErrRepositoryNameRequired
	case fullName == "":
		return nil, ErrRepositoryFullNameRequired
	case defaultBranch == "":
		return nil, ErrRepositoryDefaultBranchRequired
	case htmlURL == "":
		return nil, ErrRepositoryHTMLURLRequired
	case cloneURL == "":
		return nil, ErrRepositoryCloneURLRequired
	}

	if _, err := s.authorizeProjectMember(ctx, requesterID, projectID); err != nil {
		return nil, err
	}

	_, err := s.repositories.FindByProviderExternalID(
		ctx,
		provider,
		externalID,
	)
	switch {
	case err == nil:
		return nil, ErrRepositoryAlreadyExists
	case !errors.Is(err, repository.ErrRepositoryNotFound):
		return nil, err
	}

	now := time.Now()
	newRepository := &domain.Repository{
		ID:            uuid.NewString(),
		ProjectID:     projectID,
		Provider:      provider,
		ExternalID:    externalID,
		Owner:         owner,
		Name:          name,
		FullName:      fullName,
		DefaultBranch: defaultBranch,
		HTMLURL:       htmlURL,
		CloneURL:      cloneURL,
		IsPrivate:     isPrivate,
		SyncStatus:    "pending",
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.repositories.Create(ctx, newRepository); err != nil {
		return nil, err
	}

	return newRepository, nil
}
