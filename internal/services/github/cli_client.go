// Package github provides a client for interacting with GitHub CLI.
package github

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/rithyhuot/vibe/internal/models"
)

// CLIClient implements the Client interface using gh CLI
type CLIClient struct {
	owner string
	repo  string
}

// NewCLIClient creates a new GitHub CLI client
func NewCLIClient(owner, repo string) *CLIClient {
	return &CLIClient{
		owner: owner,
		repo:  repo,
	}
}

// runGH executes a gh CLI command and returns the output
func (c *CLIClient) runGH(ctx context.Context, args ...string) (string, error) {
	// Add repo context
	fullArgs := append([]string{"--repo", fmt.Sprintf("%s/%s", c.owner, c.repo)}, args...)

	cmd := exec.CommandContext(ctx, "gh", fullArgs...)
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if context was cancelled
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		return "", fmt.Errorf("gh CLI command failed (gh %s): %w\nOutput: %s",
			strings.Join(fullArgs, " "), err, string(output))
	}

	return strings.TrimSpace(string(output)), nil
}

// runGHWithStdin executes a gh CLI command with stdin input
func (c *CLIClient) runGHWithStdin(ctx context.Context, stdin string, args ...string) (string, error) {
	// Add repo context
	fullArgs := append([]string{"--repo", fmt.Sprintf("%s/%s", c.owner, c.repo)}, args...)

	cmd := exec.CommandContext(ctx, "gh", fullArgs...)
	cmd.Env = os.Environ()

	// Set stdin to a reader containing the content
	cmd.Stdin = strings.NewReader(stdin)

	// Run command and get output
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if context was cancelled
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		return "", fmt.Errorf("gh CLI command failed (gh %s): %w\nOutput: %s",
			strings.Join(fullArgs, " "), err, string(output))
	}

	return strings.TrimSpace(string(output)), nil
}

// CreatePR creates a new pull request using gh CLI
func (c *CLIClient) CreatePR(ctx context.Context, req *models.PRCreateRequest) (*models.PullRequest, error) {
	args := []string{"pr", "create",
		"--title", req.Title,
		"--body", req.Body,
		"--base", req.Base,
		"--head", req.Head,
	}

	if req.Draft {
		args = append(args, "--draft")
	}

	// Get PR URL from output
	output, err := c.runGH(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to create PR: %w", err)
	}

	// Validate and extract PR number from URL (format: https://github.com/owner/repo/pull/123)
	prURL := strings.TrimSpace(output)
	if !strings.HasPrefix(prURL, "https://github.com/") {
		return nil, fmt.Errorf("unexpected PR URL format (expected GitHub URL): %s", prURL)
	}

	parts := strings.Split(prURL, "/")
	if len(parts) < 7 || parts[len(parts)-2] != "pull" {
		return nil, fmt.Errorf("unexpected PR URL format (expected .../pull/123): %s", prURL)
	}

	prNumber, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		return nil, fmt.Errorf("failed to parse PR number from URL %s: %w", prURL, err)
	}

	// Get full PR details
	return c.GetPR(ctx, prNumber)
}

// GetPR retrieves a pull request by number
func (c *CLIClient) GetPR(ctx context.Context, prNumber int) (*models.PullRequest, error) {
	var prData struct {
		Number              int    `json:"number"`
		Title               string `json:"title"`
		Body                string `json:"body"`
		State               string `json:"state"`
		IsDraft             bool   `json:"isDraft"`
		Merged              bool   `json:"merged"`
		Mergeable           string `json:"mergeable"`
		URL                 string `json:"url"`
		HeadRefName         string `json:"headRefName"`
		BaseRefName         string `json:"baseRefName"`
		HeadRepositoryOwner struct {
			Login string `json:"login"`
		} `json:"headRepositoryOwner"`
	}

	args := []string{"pr", "view", strconv.Itoa(prNumber), "--json",
		"number,title,body,state,isDraft,merged,mergeable,url,headRefName,baseRefName,headRepositoryOwner",
	}

	output, err := c.runGH(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get PR: %w", err)
	}

	if err := json.Unmarshal([]byte(output), &prData); err != nil {
		return nil, fmt.Errorf("failed to parse PR data: %w", err)
	}

	mergeable := prData.Mergeable == "MERGEABLE"

	return &models.PullRequest{
		Number:    prData.Number,
		Title:     prData.Title,
		Body:      prData.Body,
		State:     strings.ToLower(prData.State),
		Draft:     prData.IsDraft,
		Merged:    prData.Merged,
		Mergeable: mergeable,
		URL:       prData.URL,
		Head: models.Branch{
			Ref: prData.HeadRefName,
		},
		Base: models.Branch{
			Ref: prData.BaseRefName,
		},
	}, nil
}

// UpdatePR updates a pull request
func (c *CLIClient) UpdatePR(ctx context.Context, prNumber int, title, body *string) (*models.PullRequest, error) {
	args := []string{"pr", "edit", strconv.Itoa(prNumber)}

	if title != nil {
		args = append(args, "--title", *title)
	}

	if body != nil {
		args = append(args, "--body", *body)
	}

	_, err := c.runGH(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update PR: %w", err)
	}

	// Get updated PR details
	return c.GetPR(ctx, prNumber)
}

// GetPRStatus retrieves the status of a pull request including reviews and checks
func (c *CLIClient) GetPRStatus(ctx context.Context, prNumber int) (*models.PRStatus, error) {
	prData, err := c.fetchPRData(ctx, prNumber)
	if err != nil {
		return nil, err
	}

	status := buildBasicPRStatus(prData)
	status.ReviewStatus = parseReviewStatus(prData)
	status.CheckStatus = parseCheckStatus(prData)

	return status, nil
}

type prStatusData struct {
	Number            int    `json:"number"`
	State             string `json:"state"`
	IsDraft           bool   `json:"isDraft"`
	Merged            bool   `json:"merged"`
	Mergeable         string `json:"mergeable"`
	URL               string `json:"url"`
	StatusCheckRollup []struct {
		Context    string `json:"context"`
		State      string `json:"state"`
		Status     string `json:"status"`
		Conclusion string `json:"conclusion"`
	} `json:"statusCheckRollup"`
	Reviews struct {
		Nodes []struct {
			State  string `json:"state"`
			Author struct {
				Login string `json:"login"`
			} `json:"author"`
		} `json:"nodes"`
	} `json:"reviews"`
}

func (c *CLIClient) fetchPRData(ctx context.Context, prNumber int) (*prStatusData, error) {
	args := []string{"pr", "view", strconv.Itoa(prNumber), "--json",
		"number,state,isDraft,merged,mergeable,url,statusCheckRollup,reviews",
	}

	output, err := c.runGH(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get PR status: %w", err)
	}

	var prData prStatusData
	if err := json.Unmarshal([]byte(output), &prData); err != nil {
		return nil, fmt.Errorf("failed to parse PR status: %w", err)
	}

	return &prData, nil
}

func buildBasicPRStatus(prData *prStatusData) *models.PRStatus {
	return &models.PRStatus{
		Number:    prData.Number,
		State:     strings.ToLower(prData.State),
		Draft:     prData.IsDraft,
		Merged:    prData.Merged,
		Mergeable: prData.Mergeable == "MERGEABLE",
		URL:       prData.URL,
	}
}

func parseReviewStatus(prData *prStatusData) models.ReviewStatus {
	reviewStatus := models.ReviewStatus{}

	// Use a map to track only the latest review from each reviewer
	latestReviews := make(map[string]string)
	for _, review := range prData.Reviews.Nodes {
		latestReviews[review.Author.Login] = review.State
	}

	// Count review states
	for _, state := range latestReviews {
		switch state {
		case "APPROVED":
			reviewStatus.Approved++
		case "CHANGES_REQUESTED":
			reviewStatus.ChangesRequested++
		case "COMMENTED":
			reviewStatus.Commented++
		default:
			reviewStatus.Pending++
		}
	}

	// Determine overall status
	reviewStatus.OverallStatus = determineOverallReviewStatus(reviewStatus)
	return reviewStatus
}

func determineOverallReviewStatus(rs models.ReviewStatus) string {
	if rs.ChangesRequested > 0 {
		return "changes_requested"
	} else if rs.Approved > 0 {
		return "approved"
	} else if rs.Commented > 0 || rs.Pending > 0 {
		return "pending"
	}
	return "none"
}

func parseCheckStatus(prData *prStatusData) models.CheckStatus {
	checkStatus := models.CheckStatus{
		Total: len(prData.StatusCheckRollup),
	}

	for _, check := range prData.StatusCheckRollup {
		checkState := check.Conclusion
		if checkState == "" {
			checkState = check.State
		}

		switch strings.ToUpper(checkState) {
		case "SUCCESS", "NEUTRAL", "SKIPPED":
			checkStatus.Passed++
		case "FAILURE", "ERROR", "TIMED_OUT", "ACTION_REQUIRED":
			checkStatus.Failed++
		case "PENDING", "QUEUED", "IN_PROGRESS":
			checkStatus.Pending++
		default:
			checkStatus.Pending++
		}
	}

	checkStatus.OverallStatus = determineOverallCheckStatus(checkStatus)
	return checkStatus
}

func determineOverallCheckStatus(cs models.CheckStatus) string {
	if cs.Failed > 0 {
		return "failure"
	} else if cs.Pending > 0 {
		return "pending"
	} else if cs.Passed > 0 {
		return "success"
	}
	return "none"
}

// ListPRs lists pull requests with optional state filter
func (c *CLIClient) ListPRs(ctx context.Context, state string) ([]*models.PullRequest, error) {
	args := []string{"pr", "list", "--state", state, "--json",
		"number,title,body,state,isDraft,merged,url,headRefName,baseRefName",
	}

	output, err := c.runGH(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list PRs: %w", err)
	}

	var prDataList []struct {
		Number      int    `json:"number"`
		Title       string `json:"title"`
		Body        string `json:"body"`
		State       string `json:"state"`
		IsDraft     bool   `json:"isDraft"`
		Merged      bool   `json:"merged"`
		URL         string `json:"url"`
		HeadRefName string `json:"headRefName"`
		BaseRefName string `json:"baseRefName"`
	}

	if err := json.Unmarshal([]byte(output), &prDataList); err != nil {
		return nil, fmt.Errorf("failed to parse PR list: %w", err)
	}

	result := make([]*models.PullRequest, len(prDataList))
	for i, pr := range prDataList {
		result[i] = &models.PullRequest{
			Number:    pr.Number,
			Title:     pr.Title,
			Body:      pr.Body,
			State:     strings.ToLower(pr.State),
			Draft:     pr.IsDraft,
			Merged:    pr.Merged,
			Mergeable: false, // Not available in list view
			URL:       pr.URL,
			Head: models.Branch{
				Ref: pr.HeadRefName,
			},
			Base: models.Branch{
				Ref: pr.BaseRefName,
			},
		}
	}

	return result, nil
}

// AddComment adds a comment to a pull request
func (c *CLIClient) AddComment(ctx context.Context, prNumber int, body string) error {
	args := []string{"pr", "comment", strconv.Itoa(prNumber), "--body", body}

	_, err := c.runGH(ctx, args...)
	if err != nil {
		return fmt.Errorf("failed to add comment: %w", err)
	}

	return nil
}

// GetPRTemplate retrieves the PR template from the repository
func (c *CLIClient) GetPRTemplate(ctx context.Context) (string, error) {
	// Try common PR template locations
	templatePaths := []string{
		".github/pull_request_template.md",
		".github/PULL_REQUEST_TEMPLATE.md",
		"docs/pull_request_template.md",
		"PULL_REQUEST_TEMPLATE.md",
	}

	for _, path := range templatePaths {
		// Use gh api to fetch file contents
		args := []string{"api", fmt.Sprintf("repos/%s/%s/contents/%s", c.owner, c.repo, path),
			"--jq", ".content"}

		output, err := c.runGH(ctx, args...)
		if err != nil {
			continue // Try next path
		}

		if output != "" {
			// The GitHub API returns base64-encoded content
			// Decode it before returning
			decoded, err := decodeBase64Content(output)
			if err != nil {
				continue // Try next path if decoding fails
			}
			return decoded, nil
		}
	}

	return "", nil // No template found
}

// decodeBase64Content decodes base64-encoded content from GitHub API
func decodeBase64Content(encoded string) (string, error) {
	// Remove any whitespace/newlines that might be in the base64 string
	encoded = strings.TrimSpace(encoded)
	encoded = strings.ReplaceAll(encoded, "\n", "")

	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 content: %w", err)
	}

	return string(decoded), nil
}

// CreateIssue creates a new issue using gh CLI
func (c *CLIClient) CreateIssue(ctx context.Context, req *models.IssueCreateRequest) (*models.Issue, error) {
	args := []string{"issue", "create", "--title", req.Title}

	// For multiline or complex bodies, use stdin instead of --body flag
	useStdin := req.Body != "" && (strings.Contains(req.Body, "\n") || len(req.Body) > 200)

	if useStdin {
		args = append(args, "--body-file", "-") // Read from stdin
	} else {
		// Always include --body flag, even if empty (gh CLI requires it for non-interactive mode)
		args = append(args, "--body", req.Body)
	}

	// Add assignees
	for _, assignee := range req.Assignees {
		args = append(args, "--assignee", assignee)
	}

	// Add labels
	for _, label := range req.Labels {
		args = append(args, "--label", label)
	}

	// Add milestone
	if req.Milestone != "" {
		args = append(args, "--milestone", req.Milestone)
	}

	// Add projects (gh CLI v2.20+ supports --project flag)
	for _, projectID := range req.ProjectIDs {
		args = append(args, "--project", projectID)
	}

	// Execute command - use custom execution for stdin support
	var output string
	var err error

	if useStdin {
		output, err = c.runGHWithStdin(ctx, req.Body, args...)
	} else {
		output, err = c.runGH(ctx, args...)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create issue: %w", err)
	}

	// Extract issue number from URL (format: https://github.com/owner/repo/issues/123)
	issueURL := strings.TrimSpace(output)
	if !strings.HasPrefix(issueURL, "https://github.com/") {
		return nil, fmt.Errorf("unexpected issue URL format (expected GitHub URL): %s", issueURL)
	}

	parts := strings.Split(issueURL, "/")
	if len(parts) < 7 || parts[len(parts)-2] != "issues" {
		return nil, fmt.Errorf("unexpected issue URL format (expected .../issues/123): %s", issueURL)
	}

	issueNumber, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		return nil, fmt.Errorf("failed to parse issue number from URL %s: %w", issueURL, err)
	}

	// Get full issue details
	return c.GetIssue(ctx, issueNumber, false)
}

// GetIssue retrieves an issue by number
// Note: The Projects field is not populated because gh CLI does not provide
// project information in the issue view JSON output. To support projects,
// consider implementing via GitHub GraphQL API.
func (c *CLIClient) GetIssue(ctx context.Context, issueNumber int, includeComments bool) (*models.Issue, error) {
	fields := "number,title,body,state,url,author,assignees,labels,milestone,createdAt,updatedAt,closedAt"
	if includeComments {
		fields += ",comments"
	}

	args := []string{"issue", "view", strconv.Itoa(issueNumber), "--json", fields}

	output, err := c.runGH(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue: %w", err)
	}

	var issueData struct {
		Number int    `json:"number"`
		Title  string `json:"title"`
		Body   string `json:"body"`
		State  string `json:"state"`
		URL    string `json:"url"`
		Author struct {
			Login string `json:"login"`
		} `json:"author"`
		Assignees []struct {
			Login string      `json:"login"`
			ID    interface{} `json:"id"` // Can be string or int
		} `json:"assignees"`
		Labels []struct {
			Name        string `json:"name"`
			Color       string `json:"color"`
			Description string `json:"description"`
		} `json:"labels"`
		Milestone *struct {
			Number      int    `json:"number"`
			Title       string `json:"title"`
			Description string `json:"description"`
			State       string `json:"state"`
		} `json:"milestone"`
		Comments []struct {
			ID     interface{} `json:"id"` // Can be string or int
			Body   string      `json:"body"`
			Author struct {
				Login string `json:"login"`
			} `json:"author"`
			CreatedAt string `json:"createdAt"`
		} `json:"comments"`
		CreatedAt string `json:"createdAt"`
		UpdatedAt string `json:"updatedAt"`
		ClosedAt  string `json:"closedAt"`
	}

	if err := json.Unmarshal([]byte(output), &issueData); err != nil {
		return nil, fmt.Errorf("failed to parse issue data: %w", err)
	}

	// Parse timestamps
	createdAt, err := time.Parse(time.RFC3339, issueData.CreatedAt)
	if err != nil {
		createdAt = time.Now() // Fallback to current time if parsing fails
	}

	updatedAt, err := time.Parse(time.RFC3339, issueData.UpdatedAt)
	if err != nil {
		updatedAt = createdAt // Fallback to created time
	}

	var closedAt *time.Time
	if issueData.ClosedAt != "" {
		t, err := time.Parse(time.RFC3339, issueData.ClosedAt)
		if err == nil {
			closedAt = &t
		}
	}

	issue := &models.Issue{
		Number:    issueData.Number,
		Title:     issueData.Title,
		Body:      issueData.Body,
		State:     strings.ToLower(issueData.State),
		URL:       issueData.URL,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		ClosedAt:  closedAt,
		User: models.GitHubUser{
			Login: issueData.Author.Login,
		},
	}

	// Parse assignees
	issue.Assignees = make([]models.GitHubUser, len(issueData.Assignees))
	for i, a := range issueData.Assignees {
		issue.Assignees[i] = models.GitHubUser{
			Login: a.Login,
			ID:    parseIDToInt(a.ID),
		}
	}

	// Parse labels
	issue.Labels = make([]models.Label, len(issueData.Labels))
	for i, l := range issueData.Labels {
		issue.Labels[i] = models.Label{
			Name:        l.Name,
			Color:       l.Color,
			Description: l.Description,
		}
	}

	// Parse milestone
	if issueData.Milestone != nil {
		issue.Milestone = &models.Milestone{
			Number:      issueData.Milestone.Number,
			Title:       issueData.Milestone.Title,
			Description: issueData.Milestone.Description,
			State:       strings.ToLower(issueData.Milestone.State),
		}
	}

	// Parse comments if included
	if includeComments {
		issue.Comments = make([]models.IssueComment, len(issueData.Comments))
		for i, c := range issueData.Comments {
			// Parse timestamp (gh CLI returns ISO 8601 format)
			commentCreatedAt, err := time.Parse(time.RFC3339, c.CreatedAt)
			if err != nil {
				commentCreatedAt = time.Now() // Fallback to current time
			}

			issue.Comments[i] = models.IssueComment{
				ID:   parseIDToInt(c.ID),
				Body: c.Body,
				User: models.GitHubUser{
					Login: c.Author.Login,
				},
				CreatedAt: commentCreatedAt,
			}
		}
	}

	return issue, nil
}

// UpdateIssue updates an existing issue
//
// Important: In CLI mode, assignees and labels are additive:
//   - Assignees: Uses --add-assignee (adds to existing, does not replace)
//   - Labels: Uses --add-label (adds to existing, does not replace)
//
// For replacement behavior, use API mode instead.
//
//nolint:gocyclo // Complexity is acceptable for handling all update fields
func (c *CLIClient) UpdateIssue(ctx context.Context, issueNumber int, req *models.IssueUpdateRequest) (*models.Issue, error) {
	args := []string{"issue", "edit", strconv.Itoa(issueNumber)}

	if req.Title != nil {
		args = append(args, "--title", *req.Title)
	}

	if req.Body != nil {
		args = append(args, "--body", *req.Body)
	}

	// Handle assignees
	// NOTE: gh CLI uses --add-assignee which is additive, not replace
	// For true replacement behavior, consider using API mode
	if req.Assignees != nil {
		for _, assignee := range *req.Assignees {
			args = append(args, "--add-assignee", assignee)
		}
	}

	// Handle labels
	// NOTE: gh CLI uses --add-label which is additive, not replace
	// For true replacement behavior, consider using API mode
	if req.Labels != nil {
		for _, label := range *req.Labels {
			args = append(args, "--add-label", label)
		}
	}

	// Handle milestone
	if req.Milestone != nil {
		args = append(args, "--milestone", *req.Milestone)
	}

	// Handle projects
	if req.ProjectIDs != nil {
		for _, projectID := range *req.ProjectIDs {
			args = append(args, "--add-project", projectID)
		}
	}

	_, err := c.runGH(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update issue: %w", err)
	}

	// Handle state changes separately (close/reopen)
	if req.State != nil {
		stateArgs := []string{"issue"}
		if *req.State == "closed" {
			stateArgs = append(stateArgs, "close", strconv.Itoa(issueNumber))
		} else if *req.State == "open" {
			stateArgs = append(stateArgs, "reopen", strconv.Itoa(issueNumber))
		}

		if len(stateArgs) > 1 {
			_, err := c.runGH(ctx, stateArgs...)
			if err != nil {
				return nil, fmt.Errorf("failed to update issue state: %w", err)
			}
		}
	}

	// Get updated issue details
	return c.GetIssue(ctx, issueNumber, false)
}

// ListIssues lists issues with optional state filter
func (c *CLIClient) ListIssues(ctx context.Context, state string) ([]*models.Issue, error) {
	args := []string{"issue", "list", "--state", state, "--json",
		"number,title,body,state,url,assignees,labels",
	}

	output, err := c.runGH(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list issues: %w", err)
	}

	var issueDataList []struct {
		Number    int    `json:"number"`
		Title     string `json:"title"`
		Body      string `json:"body"`
		State     string `json:"state"`
		URL       string `json:"url"`
		Assignees []struct {
			Login string      `json:"login"`
			ID    interface{} `json:"id"` // Can be string or int
		} `json:"assignees"`
		Labels []struct {
			Name        string `json:"name"`
			Color       string `json:"color"`
			Description string `json:"description"`
		} `json:"labels"`
	}

	if err := json.Unmarshal([]byte(output), &issueDataList); err != nil {
		return nil, fmt.Errorf("failed to parse issue list: %w", err)
	}

	result := make([]*models.Issue, len(issueDataList))
	for i, issue := range issueDataList {
		result[i] = &models.Issue{
			Number: issue.Number,
			Title:  issue.Title,
			Body:   issue.Body,
			State:  strings.ToLower(issue.State),
			URL:    issue.URL,
		}

		// Parse assignees
		result[i].Assignees = make([]models.GitHubUser, len(issue.Assignees))
		for j, a := range issue.Assignees {
			result[i].Assignees[j] = models.GitHubUser{
				Login: a.Login,
				ID:    parseIDToInt(a.ID),
			}
		}

		// Parse labels
		result[i].Labels = make([]models.Label, len(issue.Labels))
		for j, l := range issue.Labels {
			result[i].Labels[j] = models.Label{
				Name:        l.Name,
				Color:       l.Color,
				Description: l.Description,
			}
		}
	}

	return result, nil
}

// GetIssueTemplate retrieves the issue template from the repository
func (c *CLIClient) GetIssueTemplate(ctx context.Context) (string, error) {
	// Try common issue template locations
	templatePaths := []string{
		".github/ISSUE_TEMPLATE.md",
		".github/issue_template.md",
		"ISSUE_TEMPLATE.md",
		"docs/ISSUE_TEMPLATE.md",
	}

	for _, path := range templatePaths {
		// Use gh api to fetch file contents
		args := []string{"api", fmt.Sprintf("repos/%s/%s/contents/%s", c.owner, c.repo, path),
			"--jq", ".content"}

		output, err := c.runGH(ctx, args...)
		if err != nil {
			continue // Try next path
		}

		if output != "" {
			// Decode base64 content
			decoded, err := decodeBase64Content(output)
			if err != nil {
				continue // Try next path if decoding fails
			}
			return decoded, nil
		}
	}

	return "", nil // No template found
}

// parseIDToInt converts an interface{} ID (which can be string or int) to int
// GitHub CLI sometimes returns IDs as strings, sometimes as ints
func parseIDToInt(id interface{}) int {
	if id == nil {
		return 0
	}

	switch v := id.(type) {
	case int:
		return v
	case float64:
		return int(v)
	case string:
		// Try to parse string to int
		if parsed, err := strconv.Atoi(v); err == nil {
			return parsed
		}
		return 0
	default:
		return 0
	}
}
