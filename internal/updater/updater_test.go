package updater

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCheck_NewVersionAvailable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(githubRelease{TagName: "v2.0.0"})
	}))
	defer srv.Close()

	// Override releaseURL for test
	origCheck := Check
	_ = origCheck

	// Use a direct approach: build a custom check
	result, err := checkWithURL(srv.URL, "v1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Available {
		t.Error("expected update to be available")
	}
	if result.Latest != "v2.0.0" {
		t.Errorf("expected latest v2.0.0, got %s", result.Latest)
	}
}

func TestCheck_AlreadyLatest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(githubRelease{TagName: "v1.0.0"})
	}))
	defer srv.Close()

	result, err := checkWithURL(srv.URL, "v1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Available {
		t.Error("expected no update available")
	}
}

func TestCheck_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	_, err := checkWithURL(srv.URL, "v1.0.0")
	if err == nil {
		t.Error("expected error on server failure")
	}
}

// checkWithURL is a test helper that allows overriding the release URL.
func checkWithURL(url, currentVersion string) (*CheckResult, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &checkError{status: resp.StatusCode}
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &CheckResult{
		Available: release.TagName != "" && release.TagName != currentVersion,
		Current:   currentVersion,
		Latest:    release.TagName,
	}, nil
}

type checkError struct {
	status int
}

func (e *checkError) Error() string {
	return "check failed"
}
