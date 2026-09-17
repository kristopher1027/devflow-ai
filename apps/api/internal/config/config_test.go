package config

import (
	"errors"
	"testing"
)

func TestGitHubAppConfigValidate(t *testing.T) {
	config := GitHubAppConfig{
		AppID:      "123456",
		PrivateKey: "private-key",
	}

	if err := config.Validate(); err != nil {
		t.Fatalf("validate github app config: %v", err)
	}
}

func TestGitHubAppConfigValidateRequiredFields(t *testing.T) {
	tests := []struct {
		name   string
		config GitHubAppConfig
		want   error
	}{
		{
			name: "app ID",
			config: GitHubAppConfig{
				PrivateKey: "private-key",
			},
			want: ErrGitHubAppIDRequired,
		},
		{
			name: "private key",
			config: GitHubAppConfig{
				AppID: "123456",
			},
			want: ErrGitHubAppPrivateKeyRequired,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.config.Validate(); !errors.Is(err, test.want) {
				t.Fatalf("expected %v, got %v", test.want, err)
			}
		})
	}
}

func TestLoadGitHubAppConfig(t *testing.T) {
	t.Setenv("GITHUB_APP_ID", "123456")
	t.Setenv("GITHUB_APP_PRIVATE_KEY", "private-key")

	loaded := Load()

	if loaded.GitHubApp.AppID != "123456" {
		t.Fatalf("expected app ID 123456, got %s", loaded.GitHubApp.AppID)
	}

	if loaded.GitHubApp.PrivateKey != "private-key" {
		t.Fatal("expected private key to be loaded")
	}
}
