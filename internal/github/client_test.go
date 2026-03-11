package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPIToWebURL(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{
			"https://api.github.com/repos/alice/myrepo/pulls/42",
			"https://github.com/alice/myrepo/pull/42",
		},
		{
			"https://api.github.com/repos/org/project/pulls/1",
			"https://github.com/org/project/pull/1",
		},
		{"", ""},
	}

	for _, c := range cases {
		got := apiToWebURL(c.in)
		if got != c.want {
			t.Errorf("apiToWebURL(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestResolveCommentURL(t *testing.T) {
	expected := "https://github.com/alice/repo/pull/5#issuecomment-999"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"html_url": expected})
	}))
	defer srv.Close()

	client := New("fake-token")
	got := client.resolveCommentURL(context.Background(), srv.URL+"/comment/999")
	if got != expected {
		t.Errorf("resolveCommentURL = %q, want %q", got, expected)
	}
}

func TestResolveCommentURL_Empty(t *testing.T) {
	client := New("fake-token")
	got := client.resolveCommentURL(context.Background(), "")
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestResolveCommentURL_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := New("fake-token")
	got := client.resolveCommentURL(context.Background(), srv.URL+"/fail")
	if got != "" {
		t.Errorf("expected empty string on error, got %q", got)
	}
}

func TestPRCommentReasons(t *testing.T) {
	shouldInclude := []string{"comment", "mention", "team_mention", "review_requested", "author"}
	shouldExclude := []string{"subscribed", "ci_activity", "state_change", ""}

	for _, r := range shouldInclude {
		if !prCommentReasons[r] {
			t.Errorf("reason %q should be included but is not", r)
		}
	}
	for _, r := range shouldExclude {
		if prCommentReasons[r] {
			t.Errorf("reason %q should be excluded but is included", r)
		}
	}
}
