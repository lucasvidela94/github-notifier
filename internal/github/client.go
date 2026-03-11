package github

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	gogithub "github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"

	"github-notifier/internal/db"
)

// Reasons that indicate a comment or review activity on a PR.
var prCommentReasons = map[string]bool{
	"comment":         true,
	"mention":         true,
	"team_mention":    true,
	"review_requested": true,
	"author":          true,
}

type Client struct {
	token string
	gh    *gogithub.Client
	http  *http.Client
}

func New(token string) *Client {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)
	return &Client{
		token: token,
		gh:    gogithub.NewClient(tc),
		http:  tc,
	}
}

func (c *Client) FetchNotifications(ctx context.Context) ([]*db.Notification, error) {
	opts := &gogithub.NotificationListOptions{
		All:   false,
		Since: time.Now().Add(-48 * time.Hour),
	}

	raw, _, err := c.gh.Activity.ListNotifications(ctx, opts)
	if err != nil {
		return nil, err
	}

	var result []*db.Notification
	for _, n := range raw {
		// Only care about Pull Request notifications with comment-related reasons.
		if n.GetSubject().GetType() != "PullRequest" {
			continue
		}
		if !prCommentReasons[n.GetReason()] {
			continue
		}

		// Try to get a direct link to the specific comment; fall back to PR URL.
		commentURL := c.resolveCommentURL(ctx, n.GetSubject().GetLatestCommentURL())
		if commentURL == "" {
			commentURL = apiToWebURL(n.GetSubject().GetURL())
		}

		result = append(result, &db.Notification{
			ID:        n.GetID(),
			Repo:      n.GetRepository().GetFullName(),
			Title:     n.GetSubject().GetTitle(),
			Type:      n.GetSubject().GetType(),
			URL:       commentURL,
			Reason:    n.GetReason(),
			Unread:    n.GetUnread(),
			CreatedAt: time.Now(),
			UpdatedAt: n.GetUpdatedAt().Time,
		})
	}
	return result, nil
}

func (c *Client) MarkRead(ctx context.Context, id string) error {
	_, err := c.gh.Activity.MarkThreadRead(ctx, id)
	return err
}

// resolveCommentURL calls the GitHub API comment endpoint and returns its html_url,
// giving a direct link to the specific comment (with anchor).
func (c *Client) resolveCommentURL(ctx context.Context, apiURL string) string {
	if apiURL == "" {
		return ""
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.http.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return ""
	}
	defer resp.Body.Close()

	var payload struct {
		HTMLURL string `json:"html_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return ""
	}
	return payload.HTMLURL
}

// apiToWebURL converts a GitHub API PR URL to its browser URL.
//
//	https://api.github.com/repos/owner/repo/pulls/1
//	→ https://github.com/owner/repo/pull/1
func apiToWebURL(apiURL string) string {
	if apiURL == "" {
		return ""
	}
	url := strings.Replace(apiURL, "https://api.github.com/repos/", "https://github.com/", 1)
	url = strings.Replace(url, "/pulls/", "/pull/", 1)
	return url
}
