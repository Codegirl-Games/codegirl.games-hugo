package config

import (
	"fmt"
	"os"
)

// Config holds all environment-based configuration for the CMS server.
type Config struct {
	GitHubToken       string
	GitHubOwner       string
	GitHubRepo        string
	GitHubBranch      string
	SessionSecret     string
	AdminUsername     string
	AdminPasswordHash string
	Addr              string
	CookieSecure      bool
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	cfg := &Config{
		GitHubToken:       os.Getenv("GITHUB_TOKEN"),
		GitHubOwner:       os.Getenv("GITHUB_OWNER"),
		GitHubRepo:        os.Getenv("GITHUB_REPO"),
		GitHubBranch:      envOrDefault("GITHUB_BRANCH", "master"),
		SessionSecret:     os.Getenv("SESSION_SECRET"),
		AdminUsername:     os.Getenv("ADMIN_USERNAME"),
		AdminPasswordHash: os.Getenv("ADMIN_PASSWORD_HASH"),
		Addr:              envOrDefault("ADDR", ":8080"),
		CookieSecure:      envOrDefault("COOKIE_SECURE", "true") == "true",
	}

	if cfg.GitHubToken == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN is required")
	}
	if cfg.GitHubOwner == "" {
		return nil, fmt.Errorf("GITHUB_OWNER is required")
	}
	if cfg.GitHubRepo == "" {
		return nil, fmt.Errorf("GITHUB_REPO is required")
	}
	if cfg.SessionSecret == "" {
		return nil, fmt.Errorf("SESSION_SECRET is required")
	}
	if cfg.AdminUsername == "" {
		return nil, fmt.Errorf("ADMIN_USERNAME is required")
	}
	if cfg.AdminPasswordHash == "" {
		return nil, fmt.Errorf("ADMIN_PASSWORD_HASH is required")
	}

	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
