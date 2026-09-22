package config

import (
	"os"
	"strings"
)

type Config struct {
	AppPort            string
	DatabaseURL        string
	GitHubClientID     string
	GitHubClientSecret string
	SessionSecret      string
	BaseURL            string
	Env                string
	GitHubToken        string
	ChromeBin          string
}

func Load() *Config {
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = os.Getenv("PORT")
	}
	if port == "" {
		port = "8080"
	}
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@127.0.0.1:5437/agentsight?sslmode=disable"
	}

	sessionSecret := os.Getenv("SESSION_SECRET")
	if sessionSecret == "" {
		sessionSecret = "agentsight-default-secret-key-min-32-chars-long"
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = os.Getenv("ENV")
	}
	if env == "" {
		env = "development"
	}

	chromeBin := os.Getenv("CHROME_BIN")
	if chromeBin == "" {
		chromeBin = "/usr/bin/chromium-browser"
	}

	return &Config{
		AppPort:            port,
		DatabaseURL:        dbURL,
		GitHubClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		GitHubClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		SessionSecret:      sessionSecret,
		BaseURL:            baseURL,
		Env:                env,
		GitHubToken:        os.Getenv("GITHUB_TOKEN"),
		ChromeBin:          chromeBin,
	}
}
