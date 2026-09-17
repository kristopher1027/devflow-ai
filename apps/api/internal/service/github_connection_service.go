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
	ErrGitHubConnectionUnauthorized = errors.New(
		"unauthorized to manage github connection",
	)
	ErrGitHubConnectionAlreadyExists = errors.New(
		"github connection already exists",
	)
	ErrGitHubConnectionStatusTransitionInvalid = errors.New(
		"github connection status transition is invalid",
	)
)

type GitHubConnectionService interface {
	Create(
		ctx context.Context,
		requesterID string,
		workspaceID string,
		installationID string,
		accountLogin string,
	) (*domain.GitHubConnection, error)

	FindByWorkspaceID(
		ctx context.Context,
		requesterID string,
		workspaceID string,
	) (*domain.GitHubConnection, error)

	UpdateStatus(
		ctx context.Context,
		requesterID string,
		id string,
		status string,
	) error
}

type GitHubConnectionServiceImpl struct {
	connections repository.GitHubConnectionRepository
	workspaces  repository.WorkspaceRepository
	members     repository.WorkspaceMemberRepository
}

func NewGitHubConnectionService(
	connections repository.GitHubConnectionRepository,
	workspaces repository.WorkspaceRepository,
	members repository.WorkspaceMemberRepository,
) GitHubConnectionService {
	return &GitHubConnectionServiceImpl{
		connections: connections,
		workspaces:  workspaces,
		members:     members,
	}
}

func (s *GitHubConnectionServiceImpl) Create(
	ctx context.Context,
	requesterID string,
	workspaceID string,
	installationID string,
	accountLogin string,
) (*domain.GitHubConnection, error) {
	installationID = strings.TrimSpace(installationID)
	accountLogin = strings.TrimSpace(accountLogin)

	connection := &domain.GitHubConnection{
		ID:             uuid.NewString(),
		WorkspaceID:    workspaceID,
		InstallationID: installationID,
		AccountLogin:   accountLogin,
		Status:         domain.GitHubConnectionStatusPending,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	if err := connection.Validate(); err != nil {
		return nil, err
	}

	if err := s.authorizeManager(ctx, requesterID, workspaceID); err != nil {
		return nil, err
	}

	_, err := s.connections.FindByWorkspaceID(ctx, workspaceID)
	if err == nil {
		return nil, ErrGitHubConnectionAlreadyExists
	}
	if !errors.Is(err, repository.ErrGitHubConnectionNotFound) {
		return nil, err
	}

	if err := s.connections.Create(ctx, connection); err != nil {
		return nil, err
	}

	return connection, nil
}

func (s *GitHubConnectionServiceImpl) FindByWorkspaceID(
	ctx context.Context,
	requesterID string,
	workspaceID string,
) (*domain.GitHubConnection, error) {
	if err := s.authorizeMember(ctx, requesterID, workspaceID); err != nil {
		return nil, err
	}

	return s.connections.FindByWorkspaceID(ctx, workspaceID)
}

func (s *GitHubConnectionServiceImpl) UpdateStatus(
	ctx context.Context,
	requesterID string,
	id string,
	status string,
) error {
	connection, err := s.connections.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.authorizeManager(ctx, requesterID, connection.WorkspaceID); err != nil {
		return err
	}

	if !isValidGitHubConnectionTransition(connection.Status, status) {
		return ErrGitHubConnectionStatusTransitionInvalid
	}

	return s.connections.UpdateStatus(ctx, id, status)
}

func (s *GitHubConnectionServiceImpl) authorizeMember(
	ctx context.Context,
	requesterID string,
	workspaceID string,
) error {
	workspace, err := s.workspaces.FindByID(ctx, workspaceID)
	if err != nil {
		return err
	}

	if workspace.OwnerID == requesterID {
		return nil
	}

	member, err := s.members.Find(ctx, workspaceID, requesterID)
	if err != nil {
		if errors.Is(err, repository.ErrWorkspaceMemberNotFound) {
			return ErrGitHubConnectionUnauthorized
		}
		return err
	}

	switch member.Role {
	case WorkspaceMemberRoleOwner,
		WorkspaceMemberRoleAdmin,
		WorkspaceMemberRoleMember:
		return nil
	default:
		return ErrGitHubConnectionUnauthorized
	}
}

func (s *GitHubConnectionServiceImpl) authorizeManager(
	ctx context.Context,
	requesterID string,
	workspaceID string,
) error {
	workspace, err := s.workspaces.FindByID(ctx, workspaceID)
	if err != nil {
		return err
	}

	if workspace.OwnerID == requesterID {
		return nil
	}

	member, err := s.members.Find(ctx, workspaceID, requesterID)
	if err != nil {
		if errors.Is(err, repository.ErrWorkspaceMemberNotFound) {
			return ErrGitHubConnectionUnauthorized
		}
		return err
	}

	if member.Role != WorkspaceMemberRoleAdmin &&
		member.Role != WorkspaceMemberRoleOwner {
		return ErrGitHubConnectionUnauthorized
	}

	return nil
}

func isValidGitHubConnectionTransition(from string, to string) bool {
	if from == to {
		return true
	}

	switch from {
	case domain.GitHubConnectionStatusPending:
		return to == domain.GitHubConnectionStatusActive ||
			to == domain.GitHubConnectionStatusError
	case domain.GitHubConnectionStatusActive:
		return to == domain.GitHubConnectionStatusDisconnected ||
			to == domain.GitHubConnectionStatusError
	case domain.GitHubConnectionStatusDisconnected,
		domain.GitHubConnectionStatusError:
		return to == domain.GitHubConnectionStatusPending
	default:
		return false
	}
}
