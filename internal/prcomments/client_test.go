package prcomments

import (
	"testing"
)

func TestParseRepoURL_API(t *testing.T) {
	cases := []struct {
		url       string
		wantOwner string
		wantRepo  string
	}{
		{
			url:       "https://api.github.com/repos/ARPEM-LABS/ecommerce-bb-web",
			wantOwner: "ARPEM-LABS",
			wantRepo:  "ecommerce-bb-web",
		},
		{
			url:       "https://api.github.com/repos/latinmundo/lmx-frontend",
			wantOwner: "latinmundo",
			wantRepo:  "lmx-frontend",
		},
		{
			url:       "https://api.github.com/repos/epi-market/epi-web",
			wantOwner: "epi-market",
			wantRepo:  "epi-web",
		},
	}

	for _, c := range cases {
		gotOwner, gotRepo := parseRepoURL(c.url)
		if gotOwner != c.wantOwner {
			t.Errorf("parseRepoURL(%q) owner = %q, want %q", c.url, gotOwner, c.wantOwner)
		}
		if gotRepo != c.wantRepo {
			t.Errorf("parseRepoURL(%q) repo = %q, want %q", c.url, gotRepo, c.wantRepo)
		}
	}
}

func TestParseRepoURL_Web(t *testing.T) {
	cases := []struct {
		url       string
		wantOwner string
		wantRepo  string
	}{
		{
			url:       "https://github.com/ARPEM-LABS/ecommerce-bb-web",
			wantOwner: "ARPEM-LABS",
			wantRepo:  "ecommerce-bb-web",
		},
		{
			url:       "https://github.com/owner/repo",
			wantOwner: "owner",
			wantRepo:  "repo",
		},
	}

	for _, c := range cases {
		gotOwner, gotRepo := parseRepoURL(c.url)
		if gotOwner != c.wantOwner {
			t.Errorf("parseRepoURL(%q) owner = %q, want %q", c.url, gotOwner, c.wantOwner)
		}
		if gotRepo != c.wantRepo {
			t.Errorf("parseRepoURL(%q) repo = %q, want %q", c.url, gotRepo, c.wantRepo)
		}
	}
}

func TestParseRepoURL_Invalid(t *testing.T) {
	cases := []string{
		"",
		"not-a-url",
		"https://example.com/something",
		"https://github.com/",
	}

	for _, url := range cases {
		owner, repo := parseRepoURL(url)
		if owner != "" || repo != "" {
			t.Errorf("parseRepoURL(%q) should return empty, got owner=%q, repo=%q", url, owner, repo)
		}
	}
}

func TestSplitURL(t *testing.T) {
	cases := []struct {
		path string
		want []string
	}{
		{
			path: "owner/repo",
			want: []string{"owner", "repo"},
		},
		{
			path: "owner/repo/pulls/123",
			want: []string{"owner", "repo", "pulls", "123"},
		},
		{
			path: "single",
			want: []string{"single"},
		},
		{
			path: "",
			want: []string{},
		},
	}

	for _, c := range cases {
		got := splitURL(c.path)
		if len(got) != len(c.want) {
			t.Errorf("splitURL(%q) = %v, want %v", c.path, got, c.want)
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("splitURL(%q)[%d] = %q, want %q", c.path, i, got[i], c.want[i])
			}
		}
	}
}

func TestComment_Struct(t *testing.T) {
	c := &Comment{
		ID:       123,
		PRNumber: 456,
		PRTitle:  "Test PR",
		Repo:     "owner/repo",
		Body:     "Test comment",
		Author:   "testuser",
		URL:      "https://github.com/owner/repo/pull/456#issuecomment-123",
		Type:     "issue_comment",
	}

	if c.ID != 123 {
		t.Errorf("Comment.ID = %d, want 123", c.ID)
	}
	if c.Repo != "owner/repo" {
		t.Errorf("Comment.Repo = %q, want 'owner/repo'", c.Repo)
	}
}
