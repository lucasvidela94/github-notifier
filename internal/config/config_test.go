package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoad_Success(t *testing.T) {
	// Guardar valores originales
	origToken := os.Getenv("GITHUB_TOKEN")
	origUser := os.Getenv("GITHUB_USER")
	origPoll := os.Getenv("POLL_INTERVAL_SECONDS")
	origDB := os.Getenv("DB_PATH")
	defer func() {
		os.Setenv("GITHUB_TOKEN", origToken)
		os.Setenv("GITHUB_USER", origUser)
		os.Setenv("POLL_INTERVAL_SECONDS", origPoll)
		os.Setenv("DB_PATH", origDB)
	}()

	// Setear valores de test
	os.Setenv("GITHUB_TOKEN", "test-token")
	os.Setenv("GITHUB_USER", "testuser")
	os.Setenv("POLL_INTERVAL_SECONDS", "30")
	os.Setenv("DB_PATH", "/tmp/test.db")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.GitHubToken != "test-token" {
		t.Errorf("GitHubToken = %q, want 'test-token'", cfg.GitHubToken)
	}
	if cfg.GitHubUser != "testuser" {
		t.Errorf("GitHubUser = %q, want 'testuser'", cfg.GitHubUser)
	}
	if cfg.PollInterval != 30*time.Second {
		t.Errorf("PollInterval = %v, want 30s", cfg.PollInterval)
	}
	if cfg.DBPath != "/tmp/test.db" {
		t.Errorf("DBPath = %q, want '/tmp/test.db'", cfg.DBPath)
	}
}

func TestLoad_Defaults(t *testing.T) {
	// Guardar valores originales
	origToken := os.Getenv("GITHUB_TOKEN")
	origUser := os.Getenv("GITHUB_USER")
	origPoll := os.Getenv("POLL_INTERVAL_SECONDS")
	origDB := os.Getenv("DB_PATH")
	defer func() {
		os.Setenv("GITHUB_TOKEN", origToken)
		os.Setenv("GITHUB_USER", origUser)
		os.Setenv("POLL_INTERVAL_SECONDS", origPoll)
		os.Setenv("DB_PATH", origDB)
	}()

	// Solo setear lo requerido
	os.Setenv("GITHUB_TOKEN", "test-token")
	os.Setenv("GITHUB_USER", "testuser")
	os.Unsetenv("POLL_INTERVAL_SECONDS")
	os.Unsetenv("DB_PATH")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verificar defaults
	if cfg.PollInterval != 60*time.Second {
		t.Errorf("PollInterval default = %v, want 60s", cfg.PollInterval)
	}

	home, _ := os.UserHomeDir()
	expectedDB := filepath.Join(home, ".config/github-notifier/notifications.db")
	if cfg.DBPath != expectedDB {
		t.Errorf("DBPath default = %q, want %q", cfg.DBPath, expectedDB)
	}
}

func TestLoad_MissingToken(t *testing.T) {
	// Guardar valores originales
	origToken := os.Getenv("GITHUB_TOKEN")
	origUser := os.Getenv("GITHUB_USER")
	defer func() {
		os.Setenv("GITHUB_TOKEN", origToken)
		os.Setenv("GITHUB_USER", origUser)
	}()

	os.Unsetenv("GITHUB_TOKEN")
	os.Setenv("GITHUB_USER", "testuser")

	_, err := Load()
	if err == nil {
		t.Error("expected error when GITHUB_TOKEN is missing")
	}
}

func TestLoad_MissingUser(t *testing.T) {
	// Guardar valores originales
	origToken := os.Getenv("GITHUB_TOKEN")
	origUser := os.Getenv("GITHUB_USER")
	defer func() {
		os.Setenv("GITHUB_TOKEN", origToken)
		os.Setenv("GITHUB_USER", origUser)
	}()

	os.Setenv("GITHUB_TOKEN", "test-token")
	os.Unsetenv("GITHUB_USER")

	_, err := Load()
	if err == nil {
		t.Error("expected error when GITHUB_USER is missing")
	}
}

func TestLoad_InvalidPollInterval(t *testing.T) {
	// Guardar valores originales
	origToken := os.Getenv("GITHUB_TOKEN")
	origUser := os.Getenv("GITHUB_USER")
	origPoll := os.Getenv("POLL_INTERVAL_SECONDS")
	defer func() {
		os.Setenv("GITHUB_TOKEN", origToken)
		os.Setenv("GITHUB_USER", origUser)
		os.Setenv("POLL_INTERVAL_SECONDS", origPoll)
	}()

	os.Setenv("GITHUB_TOKEN", "test-token")
	os.Setenv("GITHUB_USER", "testuser")
	os.Setenv("POLL_INTERVAL_SECONDS", "invalid")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Debería usar el default cuando el valor es inválido
	if cfg.PollInterval != 60*time.Second {
		t.Errorf("PollInterval with invalid value = %v, want 60s", cfg.PollInterval)
	}
}

func TestLoad_ZeroPollInterval(t *testing.T) {
	// Guardar valores originales
	origToken := os.Getenv("GITHUB_TOKEN")
	origUser := os.Getenv("GITHUB_USER")
	origPoll := os.Getenv("POLL_INTERVAL_SECONDS")
	defer func() {
		os.Setenv("GITHUB_TOKEN", origToken)
		os.Setenv("GITHUB_USER", origUser)
		os.Setenv("POLL_INTERVAL_SECONDS", origPoll)
	}()

	os.Setenv("GITHUB_TOKEN", "test-token")
	os.Setenv("GITHUB_USER", "testuser")
	os.Setenv("POLL_INTERVAL_SECONDS", "0")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Debería usar el default cuando es 0
	if cfg.PollInterval != 60*time.Second {
		t.Errorf("PollInterval with zero = %v, want 60s", cfg.PollInterval)
	}
}
