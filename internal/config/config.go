package config

import (
	"errors"
	"os"
	"strconv"
	"time"
)

type Config struct {
	GitHubToken  string
	GitHubUser   string
	PollInterval time.Duration
	DBPath       string
}

func Load() (*Config, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, errors.New("GITHUB_TOKEN environment variable is required")
	}

	user := os.Getenv("GITHUB_USER")
	if user == "" {
		return nil, errors.New("GITHUB_USER environment variable is required")
	}

	pollSecs := 60
	if s := os.Getenv("POLL_INTERVAL_SECONDS"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			pollSecs = n
		}
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		home, _ := os.UserHomeDir()
		dbPath = home + "/.config/github-notifier/notifications.db"
	}

	return &Config{
		GitHubToken:  token,
		GitHubUser:   user,
		PollInterval: time.Duration(pollSecs) * time.Second,
		DBPath:       dbPath,
	}, nil
}
