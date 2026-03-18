package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"
)

const (
	releaseURL   = "https://api.github.com/repos/lucasvidela94/github-notifier/releases/latest"
	checkEvery   = 6 * time.Hour
	binaryName   = "github-notifier"
)

type githubRelease struct {
	TagName string `json:"tag_name"`
}

// CheckResult holds the outcome of an update check.
type CheckResult struct {
	Available bool
	Current   string
	Latest    string
}

// Check compares the running version against the latest GitHub release.
func Check(currentVersion string) (*CheckResult, error) {
	req, err := http.NewRequest("GET", releaseURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("update check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("update check got status %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("update check parse error: %w", err)
	}

	return &CheckResult{
		Available: release.TagName != "" && release.TagName != currentVersion,
		Current:   currentVersion,
		Latest:    release.TagName,
	}, nil
}

// Apply downloads the latest binary and replaces the current one, then restarts via systemd.
func Apply(version string) error {
	platform := fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)
	url := fmt.Sprintf(
		"https://github.com/lucasvidela94/github-notifier/releases/download/%s/%s_%s",
		version, binaryName, platform,
	)

	// Download to temp file
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download got status %d", resp.StatusCode)
	}

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot find own executable: %w", err)
	}

	tmp, err := os.CreateTemp("", "github-notifier-update-*")
	if err != nil {
		return fmt.Errorf("cannot create temp file: %w", err)
	}
	tmpPath := tmp.Name()

	if _, err := io.Copy(tmp, resp.Body); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("download write failed: %w", err)
	}
	tmp.Close()

	if err := os.Chmod(tmpPath, 0o755); err != nil {
		os.Remove(tmpPath)
		return err
	}

	// Atomic replace: rename over existing binary
	if err := os.Rename(tmpPath, exePath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("replace binary failed: %w", err)
	}

	log.Printf("Updated to %s, restarting service...", version)

	// Restart via systemd (non-blocking, the service manager handles it)
	_ = exec.Command("systemctl", "--user", "restart", "github-notifier.service").Start()
	return nil
}

// StartBackgroundCheck periodically checks for updates and calls onUpdate when one is found.
func StartBackgroundCheck(currentVersion string, onUpdate func(result *CheckResult)) {
	go func() {
		// First check after a short delay (let the app start up)
		time.Sleep(30 * time.Second)

		for {
			result, err := Check(currentVersion)
			if err != nil {
				log.Printf("Update check error: %v", err)
			} else if result.Available {
				log.Printf("Update available: %s -> %s", result.Current, result.Latest)
				onUpdate(result)
			}
			time.Sleep(checkEvery)
		}
	}()
}
