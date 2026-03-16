package prcomments

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

// Comment representa un comentario en un PR
type Comment struct {
	ID        int64
	PRNumber  int
	PRTitle   string
	Repo      string
	Body      string
	Author    string
	URL       string
	CreatedAt time.Time
	Type      string // "review_comment" o "issue_comment"
}

// Client consulta PRs abiertos y sus comentarios
type Client struct {
	gh    *github.Client
	token string
	user  string
}

// New crea un nuevo cliente
func New(token, user string) *Client {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)
	return &Client{
		gh:    github.NewClient(tc),
		token: token,
		user:  user,
	}
}

// FetchComments obtiene todos los comentarios de PRs abiertos del usuario
func (c *Client) FetchComments(ctx context.Context) ([]*Comment, error) {
	// Buscar PRs abiertos donde el usuario es autor
	query := fmt.Sprintf("is:pr is:open author:%s archived:false", c.user)
	result, _, err := c.gh.Search.Issues(ctx, query, nil)
	if err != nil {
		return nil, fmt.Errorf("error buscando PRs: %w", err)
	}

	var allComments []*Comment

	for _, issue := range result.Issues {
		// Extraer owner/repo del URL del repositorio
		repoURL := issue.GetRepositoryURL()
		owner, repo := parseRepoURL(repoURL)
		if owner == "" || repo == "" {
			continue
		}

		prNumber := issue.GetNumber()
		prTitle := issue.GetTitle()

		// Obtener review comments (comentarios en líneas de código)
		reviewComments, err := c.fetchReviewComments(ctx, owner, repo, prNumber, prTitle)
		if err != nil {
			// Log error pero continuar con otros PRs
			continue
		}
		allComments = append(allComments, reviewComments...)

		// Obtener issue comments (comentarios generales del PR)
		issueComments, err := c.fetchIssueComments(ctx, owner, repo, prNumber, prTitle)
		if err != nil {
			continue
		}
		allComments = append(allComments, issueComments...)
	}

	return allComments, nil
}

func (c *Client) fetchReviewComments(ctx context.Context, owner, repo string, prNumber int, prTitle string) ([]*Comment, error) {
	comments, _, err := c.gh.PullRequests.ListComments(ctx, owner, repo, prNumber, nil)
	if err != nil {
		return nil, err
	}

	var result []*Comment
	for _, cmt := range comments {
		// Ignorar comentarios del propio usuario
		if cmt.GetUser().GetLogin() == c.user {
			continue
		}

		result = append(result, &Comment{
			ID:        cmt.GetID(),
			PRNumber:  prNumber,
			PRTitle:   prTitle,
			Repo:      fmt.Sprintf("%s/%s", owner, repo),
			Body:      cmt.GetBody(),
			Author:    cmt.GetUser().GetLogin(),
			URL:       cmt.GetHTMLURL(),
			CreatedAt: cmt.GetCreatedAt().Time,
			Type:      "review_comment",
		})
	}

	return result, nil
}

func (c *Client) fetchIssueComments(ctx context.Context, owner, repo string, prNumber int, prTitle string) ([]*Comment, error) {
	// Los PRs son issues, entonces usamos la API de issues
	comments, _, err := c.gh.Issues.ListComments(ctx, owner, repo, prNumber, nil)
	if err != nil {
		return nil, err
	}

	var result []*Comment
	for _, cmt := range comments {
		// Ignorar comentarios del propio usuario
		if cmt.GetUser().GetLogin() == c.user {
			continue
		}

		result = append(result, &Comment{
			ID:        cmt.GetID(),
			PRNumber:  prNumber,
			PRTitle:   prTitle,
			Repo:      fmt.Sprintf("%s/%s", owner, repo),
			Body:      cmt.GetBody(),
			Author:    cmt.GetUser().GetLogin(),
			URL:       cmt.GetHTMLURL(),
			CreatedAt: cmt.GetCreatedAt().Time,
			Type:      "issue_comment",
		})
	}

	return result, nil
}

func parseRepoURL(url string) (owner, repo string) {
	// URL formato: https://api.github.com/repos/owner/repo
	// o: https://github.com/owner/repo
	var parts []string
	if len(url) > 29 && url[:29] == "https://api.github.com/repos/" {
		parts = splitURL(url[29:])
	} else if len(url) > 19 && url[:19] == "https://github.com/" {
		parts = splitURL(url[19:])
	}

	if len(parts) >= 2 {
		return parts[0], parts[1]
	}
	return "", ""
}

func splitURL(path string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(path); i++ {
		if path[i] == '/' {
			if i > start {
				parts = append(parts, path[start:i])
			}
			start = i + 1
		}
	}
	if start < len(path) {
		parts = append(parts, path[start:])
	}
	return parts
}
