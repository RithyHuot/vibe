package github

import (
	"context"
	"encoding/base64"
	"fmt"
	"os/exec"

	"github.com/rithyhuot/vibe/internal/models"
	"github.com/rithyhuot/vibe/internal/utils"
)

const (
	baseURL = "https://api.github.com"
)

// Client interface defines GitHub operations
type Client interface {
	CreatePR(ctx context.Context, req *models.PRCreateRequest) (*models.PullRequest, error)
	GetPR(ctx context.Context, prNumber int) (*models.PullRequest, error)
	UpdatePR(ctx context.Context, prNumber int, title, body *string) (*models.PullRequest, error)
	GetPRStatus(ctx context.Context, prNumber int) (*models.PRStatus, error)
	ListPRs(ctx context.Context, state string) ([]*models.PullRequest, error)
	AddComment(ctx context.Context, prNumber int, body string) error
	GetPRTemplate(ctx context.Context) (string, error)

	// Issue operations
	CreateIssue(ctx context.Context, req *models.IssueCreateRequest) (*models.Issue, error)
	GetIssue(ctx context.Context, issueNumber int, includeComments bool) (*models.Issue, error)
	UpdateIssue(ctx context.Context, issueNumber int, req *models.IssueUpdateRequest) (*models.Issue, error)
	ListIssues(ctx context.Context, state string) ([]*models.Issue, error)
	GetIssueTemplate(ctx context.Context) (string, error)
}

// HTTPClient implements the Client interface using go-github
type HTTPClient struct {
	httpClient *utils.HTTPClient
	token      string
	owner      string
	repo       string
	baseURL    string // Base URL for REST API (default: https://api.github.com)
	graphqlURL string // GraphQL URL (default: https://api.github.com/graphql)
}

// NewClient creates a new GitHub client (API-based)
func NewClient(token, owner, repo string) *HTTPClient {
	return &HTTPClient{
		httpClient: utils.NewHTTPClient(0).WithUserAgent("vibe"),
		token:      token,
		owner:      owner,
		repo:       repo,
		baseURL:    baseURL,
		graphqlURL: graphqlURL,
	}
}

// NewClientWithMode creates a GitHub client based on the specified mode
// mode can be "api", "cli", or "auto"
// Returns Client interface that can be either HTTPClient or CLIClient
func NewClientWithMode(mode, token, owner, repo string) (Client, error) {
	switch mode {
	case "cli":
		// Force CLI mode
		if !IsGHCLIAvailable() {
			return nil, fmt.Errorf("gh CLI is not available or not authenticated\n" +
				"To set up gh CLI, run: gh auth login\n" +
				"Alternatively, switch to API mode by setting github.mode: \"api\" in your config")
		}
		return NewCLIClient(owner, repo), nil

	case "api":
		// Force API mode
		if token == "" {
			return nil, fmt.Errorf("GitHub token is required for API mode\n" +
				"To set up a token, visit: https://github.com/settings/tokens\n" +
				"Alternatively, switch to CLI mode by setting github.mode: \"cli\" in your config")
		}
		return NewClient(token, owner, repo), nil

	case "auto":
		// Auto-detect: prefer CLI if available, fallback to API
		if IsGHCLIAvailable() {
			return NewCLIClient(owner, repo), nil
		}
		if token != "" {
			return NewClient(token, owner, repo), nil
		}
		return nil, fmt.Errorf("neither gh CLI nor GitHub token is available\n" +
			"To use gh CLI, run: gh auth login\n" +
			"To use API mode, add your token to config.yaml: github.token\n" +
			"Get a token from: https://github.com/settings/tokens")

	default:
		return nil, fmt.Errorf("invalid GitHub mode: %s (must be 'api', 'cli', or 'auto')", mode)
	}
}

// headers returns the common headers for GitHub API requests
func (c *HTTPClient) headers() map[string]string {
	return map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", c.token),
		"Accept":        "application/vnd.github.v3+json",
		"Content-Type":  "application/json",
	}
}

// CreatePR creates a new pull request
func (c *HTTPClient) CreatePR(ctx context.Context, req *models.PRCreateRequest) (*models.PullRequest, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/pulls", c.baseURL, c.owner, c.repo)

	payload := map[string]interface{}{
		"title": req.Title,
		"body":  req.Body,
		"head":  req.Head,
		"base":  req.Base,
		"draft": req.Draft,
	}

	var resp PRResponse
	err := c.httpClient.DoJSONRequest(ctx, "POST", url, payload, &resp, c.headers())
	if err != nil {
		return nil, fmt.Errorf("failed to create PR: %w", err)
	}

	return resp.ToPullRequest(), nil
}

// GetPR retrieves a pull request by number
func (c *HTTPClient) GetPR(ctx context.Context, prNumber int) (*models.PullRequest, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/pulls/%d", c.baseURL, c.owner, c.repo, prNumber)

	var resp PRResponse
	err := c.httpClient.DoJSONRequest(ctx, "GET", url, nil, &resp, c.headers())
	if err != nil {
		return nil, fmt.Errorf("failed to get PR: %w", err)
	}

	return resp.ToPullRequest(), nil
}

// UpdatePR updates a pull request
func (c *HTTPClient) UpdatePR(ctx context.Context, prNumber int, title, body *string) (*models.PullRequest, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/pulls/%d", c.baseURL, c.owner, c.repo, prNumber)

	payload := make(map[string]interface{})
	if title != nil {
		payload["title"] = *title
	}
	if body != nil {
		payload["body"] = *body
	}

	var resp PRResponse
	err := c.httpClient.DoJSONRequest(ctx, "PATCH", url, payload, &resp, c.headers())
	if err != nil {
		return nil, fmt.Errorf("failed to update PR: %w", err)
	}

	return resp.ToPullRequest(), nil
}

// GetPRStatus retrieves the status of a pull request including reviews and checks
func (c *HTTPClient) GetPRStatus(ctx context.Context, prNumber int) (*models.PRStatus, error) {
	// Get PR details
	pr, err := c.GetPR(ctx, prNumber)
	if err != nil {
		return nil, err
	}

	status := &models.PRStatus{
		Number:    pr.Number,
		State:     pr.State,
		Draft:     pr.Draft,
		Merged:    pr.Merged,
		Mergeable: pr.Mergeable,
		URL:       pr.URL,
	}

	// Get review status
	reviewStatus, err := c.getReviewStatus(ctx, prNumber)
	if err != nil {
		// Don't fail on review status error, just log it
		reviewStatus = &models.ReviewStatus{OverallStatus: "unknown"}
	}
	status.ReviewStatus = *reviewStatus

	// Get check status
	if pr.Head.SHA == "" {
		// No SHA available, mark as unknown
		status.CheckStatus = models.CheckStatus{OverallStatus: "unknown"}
	} else {
		checkStatus, err := c.getCheckStatus(ctx, pr.Head.SHA)
		if err != nil {
			// Don't fail on check status error, just log it
			checkStatus = &models.CheckStatus{OverallStatus: "unknown"}
		}
		status.CheckStatus = *checkStatus
	}

	return status, nil
}

// getReviewStatus retrieves review status for a PR
func (c *HTTPClient) getReviewStatus(ctx context.Context, prNumber int) (*models.ReviewStatus, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/pulls/%d/reviews", c.baseURL, c.owner, c.repo, prNumber)

	var reviews []ReviewResponse
	err := c.httpClient.DoJSONRequest(ctx, "GET", url, nil, &reviews, c.headers())
	if err != nil {
		return nil, fmt.Errorf("failed to get reviews: %w", err)
	}

	status := &models.ReviewStatus{}

	// Count reviews by state (only count the latest review per user)
	latestReviews := make(map[string]string)
	for _, review := range reviews {
		latestReviews[review.User.Login] = review.State
	}

	for _, state := range latestReviews {
		switch state {
		case "APPROVED":
			status.Approved++
		case "CHANGES_REQUESTED":
			status.ChangesRequested++
		case "COMMENTED":
			status.Commented++
		default:
			status.Pending++
		}
	}

	// Determine overall status
	if status.ChangesRequested > 0 {
		status.OverallStatus = "changes_requested"
	} else if status.Approved > 0 {
		status.OverallStatus = "approved"
	} else if status.Commented > 0 || status.Pending > 0 {
		status.OverallStatus = "pending"
	} else {
		status.OverallStatus = "none"
	}

	return status, nil
}

// getCheckStatus retrieves check status for a commit
func (c *HTTPClient) getCheckStatus(ctx context.Context, sha string) (*models.CheckStatus, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/commits/%s/check-runs", c.baseURL, c.owner, c.repo, sha)

	var resp struct {
		TotalCount int                `json:"total_count"`
		CheckRuns  []CheckRunResponse `json:"check_runs"`
	}

	err := c.httpClient.DoJSONRequest(ctx, "GET", url, nil, &resp, c.headers())
	if err != nil {
		return nil, fmt.Errorf("failed to get check runs: %w", err)
	}

	status := &models.CheckStatus{
		Total: resp.TotalCount,
	}

	for _, check := range resp.CheckRuns {
		switch check.Conclusion {
		case "success":
			status.Passed++
		case "failure", "timed_out", "action_required":
			status.Failed++
		default:
			if check.Status == "completed" {
				status.Passed++
			} else {
				status.Pending++
			}
		}
	}

	// Determine overall status
	if status.Failed > 0 {
		status.OverallStatus = "failure"
	} else if status.Pending > 0 {
		status.OverallStatus = "pending"
	} else if status.Passed > 0 {
		status.OverallStatus = "success"
	} else {
		status.OverallStatus = "none"
	}

	return status, nil
}

// ListPRs lists pull requests with optional state filter
//
//nolint:dupl // Similar to ListIssues but uses different types
func (c *HTTPClient) ListPRs(ctx context.Context, state string) ([]*models.PullRequest, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/pulls?state=%s", c.baseURL, c.owner, c.repo, state)

	var prs []PRResponse
	err := c.httpClient.DoJSONRequest(ctx, "GET", url, nil, &prs, c.headers())
	if err != nil {
		return nil, fmt.Errorf("failed to list PRs: %w", err)
	}

	result := make([]*models.PullRequest, len(prs))
	for i, pr := range prs {
		result[i] = pr.ToPullRequest()
	}

	return result, nil
}

// AddComment adds a comment to a pull request
func (c *HTTPClient) AddComment(ctx context.Context, prNumber int, body string) error {
	url := fmt.Sprintf("%s/repos/%s/%s/issues/%d/comments", c.baseURL, c.owner, c.repo, prNumber)

	payload := map[string]interface{}{
		"body": body,
	}

	var resp CommentResponse
	err := c.httpClient.DoJSONRequest(ctx, "POST", url, payload, &resp, c.headers())
	if err != nil {
		return fmt.Errorf("failed to add comment: %w", err)
	}

	return nil
}

// getTemplateFromPaths fetches a template from one of the given paths
func (c *HTTPClient) getTemplateFromPaths(ctx context.Context, paths []string) (string, error) {
	for _, path := range paths {
		url := fmt.Sprintf("%s/repos/%s/%s/contents/%s", c.baseURL, c.owner, c.repo, path)

		var resp struct {
			Content  string `json:"content"`
			Encoding string `json:"encoding"`
		}

		err := c.httpClient.DoJSONRequest(ctx, "GET", url, nil, &resp, c.headers())
		if err == nil && resp.Content != "" {
			// Decode base64 content
			decoded, err := base64.StdEncoding.DecodeString(resp.Content)
			if err == nil {
				return string(decoded), nil
			}
		}
	}

	return "", nil // No template found, return empty string
}

// GetPRTemplate retrieves the PR template from the repository
func (c *HTTPClient) GetPRTemplate(ctx context.Context) (string, error) {
	templatePaths := []string{
		".github/pull_request_template.md",
		".github/PULL_REQUEST_TEMPLATE.md",
		"docs/pull_request_template.md",
		"PULL_REQUEST_TEMPLATE.md",
	}
	return c.getTemplateFromPaths(ctx, templatePaths)
}

// IsGHCLIAvailable checks if gh CLI is available and configured
func IsGHCLIAvailable() bool {
	cmd := exec.Command("gh", "auth", "status")
	err := cmd.Run()
	return err == nil
}

// CreateIssue creates a new issue with support for Projects v2
//
// This uses a two-step process:
// 1. Create issue via REST API
// 2. Add to Projects v2 via GraphQL API (if ProjectIDs specified)
//
// If step 1 succeeds but step 2 fails, the issue is still created and a
// partial success error is returned with details about which projects failed.
func (c *HTTPClient) CreateIssue(ctx context.Context, req *models.IssueCreateRequest) (*models.Issue, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/issues", c.baseURL, c.owner, c.repo)

	payload := map[string]interface{}{
		"title": req.Title,
		"body":  req.Body,
	}

	if len(req.Assignees) > 0 {
		payload["assignees"] = req.Assignees
	}
	if len(req.Labels) > 0 {
		payload["labels"] = req.Labels
	}
	if req.Milestone != "" {
		payload["milestone"] = req.Milestone
	}

	var resp IssueResponse
	err := c.httpClient.DoJSONRequest(ctx, "POST", url, payload, &resp, c.headers())
	if err != nil {
		return nil, fmt.Errorf("failed to create issue: %w", err)
	}

	issue := resp.ToIssue()

	// Step 2: Add to projects if specified (GraphQL API for Projects v2)
	if len(req.ProjectIDs) > 0 {
		if resp.NodeID == "" {
			// This shouldn't happen - GitHub should always return node_id
			return issue, fmt.Errorf("issue created but cannot add to projects: GitHub did not return node_id")
		}
		err := c.addIssueToProjects(ctx, resp.Number, resp.NodeID, req.ProjectIDs)
		if err != nil {
			// Issue created successfully, but project add failed
			// Return partial success with error details
			return issue, fmt.Errorf("issue created but failed to add to projects: %w", err)
		}
	}

	return issue, nil
}

// GetIssue retrieves an issue by number
func (c *HTTPClient) GetIssue(ctx context.Context, issueNumber int, includeComments bool) (*models.Issue, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/issues/%d", c.baseURL, c.owner, c.repo, issueNumber)

	var resp IssueResponse
	err := c.httpClient.DoJSONRequest(ctx, "GET", url, nil, &resp, c.headers())
	if err != nil {
		return nil, fmt.Errorf("failed to get issue: %w", err)
	}

	issue := resp.ToIssue()

	// Fetch comments if requested
	if includeComments {
		comments, err := c.getIssueComments(ctx, issueNumber)
		if err != nil {
			// Don't fail on comments error, just log it
			issue.Comments = []models.IssueComment{}
		} else {
			issue.Comments = comments
		}
	}

	return issue, nil
}

// getIssueComments retrieves comments for an issue
func (c *HTTPClient) getIssueComments(ctx context.Context, issueNumber int) ([]models.IssueComment, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/issues/%d/comments", c.baseURL, c.owner, c.repo, issueNumber)

	var commentsResp []CommentRef
	err := c.httpClient.DoJSONRequest(ctx, "GET", url, nil, &commentsResp, c.headers())
	if err != nil {
		return nil, fmt.Errorf("failed to get comments: %w", err)
	}

	comments := make([]models.IssueComment, len(commentsResp))
	for i, c := range commentsResp {
		comments[i] = models.IssueComment{
			ID:   c.ID,
			Body: c.Body,
			User: models.GitHubUser{
				Login: c.User.Login,
				ID:    c.User.ID,
			},
			CreatedAt: c.CreatedAt,
		}
	}

	return comments, nil
}

// UpdateIssue updates an existing issue with support for Projects v2
//
// This uses a two-step process:
// 1. Update issue via REST API
// 2. Add to Projects v2 via GraphQL API (if ProjectIDs specified)
//
// If step 1 succeeds but step 2 fails, the issue is still updated and a
// partial success error is returned with details about which projects failed.
//
// Note: This only adds to projects, it does not remove from existing projects.
func (c *HTTPClient) UpdateIssue(ctx context.Context, issueNumber int, req *models.IssueUpdateRequest) (*models.Issue, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/issues/%d", c.baseURL, c.owner, c.repo, issueNumber)

	payload := make(map[string]interface{})
	if req.Title != nil {
		payload["title"] = *req.Title
	}
	if req.Body != nil {
		payload["body"] = *req.Body
	}
	if req.State != nil {
		payload["state"] = *req.State
	}
	if req.Assignees != nil {
		payload["assignees"] = *req.Assignees
	}
	if req.Labels != nil {
		payload["labels"] = *req.Labels
	}
	if req.Milestone != nil {
		payload["milestone"] = *req.Milestone
	}

	var resp IssueResponse
	err := c.httpClient.DoJSONRequest(ctx, "PATCH", url, payload, &resp, c.headers())
	if err != nil {
		return nil, fmt.Errorf("failed to update issue: %w", err)
	}

	issue := resp.ToIssue()

	// Step 2: Add to projects if specified (GraphQL API for Projects v2)
	if req.ProjectIDs != nil && len(*req.ProjectIDs) > 0 {
		if resp.NodeID == "" {
			// This shouldn't happen - GitHub should always return node_id
			return issue, fmt.Errorf("issue updated but cannot add to projects: GitHub did not return node_id")
		}
		err := c.addIssueToProjects(ctx, resp.Number, resp.NodeID, *req.ProjectIDs)
		if err != nil {
			// Issue updated successfully, but project add failed
			// Return partial success with error details
			return issue, fmt.Errorf("issue updated but failed to add to projects: %w", err)
		}
	}

	return issue, nil
}

// ListIssues lists issues with optional state filter
//
//nolint:dupl // Similar to ListPRs but uses different types
func (c *HTTPClient) ListIssues(ctx context.Context, state string) ([]*models.Issue, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/issues?state=%s&per_page=100", c.baseURL, c.owner, c.repo, state)

	var issues []IssueResponse
	err := c.httpClient.DoJSONRequest(ctx, "GET", url, nil, &issues, c.headers())
	if err != nil {
		return nil, fmt.Errorf("failed to list issues: %w", err)
	}

	result := make([]*models.Issue, len(issues))
	for i, issue := range issues {
		result[i] = issue.ToIssue()
	}

	return result, nil
}

// GetIssueTemplate retrieves the issue template from the repository
func (c *HTTPClient) GetIssueTemplate(ctx context.Context) (string, error) {
	templatePaths := []string{
		".github/ISSUE_TEMPLATE.md",
		".github/issue_template.md",
		"ISSUE_TEMPLATE.md",
		"docs/ISSUE_TEMPLATE.md",
	}
	return c.getTemplateFromPaths(ctx, templatePaths)
}
