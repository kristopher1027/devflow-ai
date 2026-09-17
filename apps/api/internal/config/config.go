package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
)

var (
	ErrGitHubAppIDRequired         = errors.New("github app ID is required")
	ErrGitHubAppPrivateKeyRequired = errors.New("github app private key is required")
)

type GitHubAppConfig struct {
	AppID      string
	PrivateKey string
}

func (c GitHubAppConfig) Validate() error {
	if strings.TrimSpace(c.AppID) == "" {
		return ErrGitHubAppIDRequired
	}

	if strings.TrimSpace(c.PrivateKey) == "" {
		return ErrGitHubAppPrivateKeyRequired
	}

	return nil
}

type Config struct {
	Port         string
	DatabaseURL  string
	CookieSecure bool
	GitHubApp    GitHubAppConfig
}

func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	databaseURL := os.Getenv("DATABASE_URL")

	cookieSecure := false

	if value := os.Getenv("DEVFLOW_COOKIE_SECURE"); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err == nil {
			cookieSecure = parsed
		}
	}

	return Config{
		Port:         port,
		DatabaseURL:  databaseURL,
		CookieSecure: cookieSecure,
		GitHubApp: GitHubAppConfig{
			AppID:      os.Getenv("GITHUB_APP_ID"),
			PrivateKey: os.Getenv("GITHUB_APP_PRIVATE_KEY"),
		},
	}
}
