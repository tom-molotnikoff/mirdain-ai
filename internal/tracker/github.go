package tracker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// ErrUnauthorized is returned by GitHubTracker when the token is missing or
// rejected by the GitHub API.
var ErrUnauthorized = errors.New("GitHub API: unauthorized — check MIRDAIN_GITHUB_TOKEN")

// ghIssueResponse mirrors the GitHub REST API issue object. The pull_request
// field is present on PRs but absent on plain issues; we use it to filter PRs.
type ghIssueResponse struct {
	NodeID      string    `json:"node_id"`
	Number      int       `json:"number"`
	Title       string    `json:"title"`
	Body        string    `json:"body"`
	Labels      []ghLabel `json:"labels"`
	PullRequest *struct{} `json:"pull_request"`
}

type ghLabel struct {
	Name string `json:"name"`
}

// GitHubTracker implements Tracker against the GitHub REST API using a PAT.
// TODO(#3): replace PAT with GitHub App installation token.
type GitHubTracker struct {
	token  string
	owner  string
	repo   string
	client *http.Client
}

// NewGitHub creates a GitHubTracker for the given repoID.
// repoID must be in the form "github.com/owner/repo".
// Token validation is deferred to the first API call so the server can boot
// without a GitHub dependency.
func NewGitHub(token, repoID string) (*GitHubTracker, error) {
	trimmed := strings.TrimPrefix(repoID, "github.com/")
	parts := strings.SplitN(trimmed, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, fmt.Errorf("invalid repo ID %q: expected format github.com/owner/repo", repoID)
	}
	return &GitHubTracker{
		token: token,
		owner: parts[0],
		repo:  parts[1],
		client: &http.Client{Timeout: 15 * time.Second},
	}, nil
}

func (g *GitHubTracker) doGet(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+g.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	return g.client.Do(req)
}

// ListIssues returns open issues (not PRs) that carry the given label.
// It uses per_page=100 without pagination (acceptable for v1 single-user scale).
// TODO: handle Link header pagination if needed at scale.
func (g *GitHubTracker) ListIssues(ctx context.Context, label string) ([]Issue, error) {
	if g.token == "" {
		return nil, ErrUnauthorized
	}

	url := fmt.Sprintf(
		"https://api.github.com/repos/%s/%s/issues?labels=%s&state=open&per_page=100",
		g.owner, g.repo, label,
	)

	resp, err := g.doGet(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("GitHub API request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrUnauthorized
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned HTTP %d", resp.StatusCode)
	}

	var raw []ghIssueResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decoding GitHub API response: %w", err)
	}

	var issues []Issue
	for _, r := range raw {
		// GitHub's /issues endpoint includes pull requests; skip them.
		if r.PullRequest != nil {
			continue
		}
		labels := make([]string, len(r.Labels))
		for i, l := range r.Labels {
			labels[i] = l.Name
		}
		issues = append(issues, Issue{
			ID:     r.NodeID,
			Number: r.Number,
			Title:  r.Title,
			Labels: labels,
			Body:   r.Body,
		})
	}
	return issues, nil
}

// GetIssue returns the issue with the given node ID.
// TODO(#20): implement.
func (g *GitHubTracker) GetIssue(ctx context.Context, id string) (Issue, error) {
	return Issue{}, fmt.Errorf("GetIssue: not implemented (TODO #20)")
}

// AddComment posts a comment on the given issue.
// TODO(#20): implement.
func (g *GitHubTracker) AddComment(ctx context.Context, issueID, body string) (string, error) {
	return "", fmt.Errorf("AddComment: not implemented (TODO #20)")
}
