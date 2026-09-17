package domain

import (
	"errors"
	"testing"
)

func validGitHubConnection() GitHubConnection {
	return GitHubConnection{
		WorkspaceID:    "workspace-123",
		InstallationID: "installation-123",
		AccountLogin:   "devflow-org",
		Status:         GitHubConnectionStatusPending,
	}
}

func TestGitHubConnectionValidateStatuses(t *testing.T) {
	statuses := []string{
		GitHubConnectionStatusPending,
		GitHubConnectionStatusActive,
		GitHubConnectionStatusDisconnected,
		GitHubConnectionStatusError,
	}

	for _, status := range statuses {
		t.Run(status, func(t *testing.T) {
			connection := validGitHubConnection()
			connection.Status = status

			if err := connection.Validate(); err != nil {
				t.Fatalf("validate status %q: %v", status, err)
			}
		})
	}
}

func TestGitHubConnectionValidateRequiredFields(t *testing.T) {
	tests := []struct {
		name string
		set  func(*GitHubConnection)
		want error
	}{
		{
			name: "workspace",
			set: func(connection *GitHubConnection) {
				connection.WorkspaceID = "   "
			},
			want: ErrGitHubConnectionWorkspaceRequired,
		},
		{
			name: "installation",
			set: func(connection *GitHubConnection) {
				connection.InstallationID = "   "
			},
			want: ErrGitHubConnectionInstallationRequired,
		},
		{
			name: "account",
			set: func(connection *GitHubConnection) {
				connection.AccountLogin = "   "
			},
			want: ErrGitHubConnectionAccountRequired,
		},
		{
			name: "status",
			set: func(connection *GitHubConnection) {
				connection.Status = "unknown"
			},
			want: ErrGitHubConnectionStatusInvalid,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			connection := validGitHubConnection()
			test.set(&connection)

			if err := connection.Validate(); !errors.Is(err, test.want) {
				t.Fatalf("expected %v, got %v", test.want, err)
			}
		})
	}
}
