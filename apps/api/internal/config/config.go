package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port         string
	DatabaseURL  string
	CookieSecure bool
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
	}
}
