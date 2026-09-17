package domain

import (
	"errors"
	"strings"
	"time"
)

const (
	GitHubConnectionStatusPending      = "pending"
	GitHubConnectionStatusActive       = "active"
	GitHubConnectionStatusDisconnected = "disconnected"
	GitHubConnectionStatusError        = "error"
)

var (
	ErrGitHubConnectionWorkspaceRequired    = errors.New("github connection workspace is required")
	ErrGitHubConnectionInstallationRequired = errors.New("github connection installation ID is required")
	ErrGitHubConnectionAccountRequired      = errors.New("github connection account login is required")
	ErrGitHubConnectionStatusInvalid        = errors.New("github connection status is invalid")
)

type GitHubConnection struct {
	ID             string
	WorkspaceID    string
	InstallationID string
	AccountLogin   string
	Status         string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (c *GitHubConnection) Validate() error {
	if strings.TrimSpace(c.WorkspaceID) == "" {
		return ErrGitHubConnectionWorkspaceRequired
	}

	if strings.TrimSpace(c.InstallationID) == "" {
		return ErrGitHubConnectionInstallationRequired
	}

	if strings.TrimSpace(c.AccountLogin) == "" {
		return ErrGitHubConnectionAccountRequired
	}

	switch c.Status {
	case GitHubConnectionStatusPending,
		GitHubConnectionStatusActive,
		GitHubConnectionStatusDisconnected,
		GitHubConnectionStatusError:
		return nil
	default:
		return ErrGitHubConnectionStatusInvalid
	}
}
