package github

import (
	"time"

	"github.com/rithyhuot/vibe/internal/models"
)

// PRResponse represents the GitHub API response for a pull request
type PRResponse struct {
	Number    int        `json:"number"`
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	State     string     `json:"state"`
	Draft     bool       `json:"draft"`
	Merged    bool       `json:"merged"`
	Mergeable *bool      `json:"mergeable"`
	HTMLURL   string     `json:"html_url"`
	Head      BranchRef  `json:"head"`
	Base      BranchRef  `json:"base"`
	User      UserRef    `json:"user"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	MergedAt  *time.Time `json:"merged_at"`
}

// BranchRef represents a branch reference in GitHub API
type BranchRef struct {
	Ref  string  `json:"ref"`
	SHA  string  `json:"sha"`
	Repo RepoRef `json:"repo"`
}

// RepoRef represents a repository reference in GitHub API
type RepoRef struct {
	Name     string `json:"name"`
	FullName string `json:"full_name"`
}

// UserRef represents a user reference in GitHub API
type UserRef struct {
	Login string `json:"login"`
	ID    int    `json:"id"`
}

// ReviewResponse represents a PR review in GitHub API
type ReviewResponse struct {
	ID          int       `json:"id"`
	User        UserRef   `json:"user"`
	State       string    `json:"state"`
	Body        string    `json:"body"`
	SubmittedAt time.Time `json:"submitted_at"`
}

// CheckRunResponse represents a check run in GitHub API
type CheckRunResponse struct {
	ID          int        `json:"id"`
	Name        string     `json:"name"`
	Status      string     `json:"status"`
	Conclusion  string     `json:"conclusion"`
	StartedAt   time.Time  `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
	HTMLURL     string     `json:"html_url"`
}

// CheckSuiteResponse represents a check suite in GitHub API
type CheckSuiteResponse struct {
	ID         int    `json:"id"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	HeadBranch string `json:"head_branch"`
	HeadSHA    string `json:"head_sha"`
}

// CommentResponse represents a PR comment in GitHub API
type CommentResponse struct {
	ID        int       `json:"id"`
	Body      string    `json:"body"`
	User      UserRef   `json:"user"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// IssueResponse represents the GitHub API response for an issue
type IssueResponse struct {
	Number    int           `json:"number"`
	Title     string        `json:"title"`
	Body      string        `json:"body"`
	State     string        `json:"state"`
	HTMLURL   string        `json:"html_url"`
	NodeID    string        `json:"node_id"` // GitHub GraphQL node ID for Projects v2
	User      UserRef       `json:"user"`
	Assignees []UserRef     `json:"assignees"`
	Labels    []LabelRef    `json:"labels"`
	Milestone *MilestoneRef `json:"milestone"`
	Comments  []CommentRef  `json:"comments"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
	ClosedAt  *time.Time    `json:"closed_at"`
}

// LabelRef represents a label reference in GitHub API
type LabelRef struct {
	Name        string `json:"name"`
	Color       string `json:"color"`
	Description string `json:"description"`
}

// MilestoneRef represents a milestone reference in GitHub API
type MilestoneRef struct {
	Number      int    `json:"number"`
	Title       string `json:"title"`
	Description string `json:"description"`
	State       string `json:"state"`
}

// CommentRef represents a comment reference in GitHub API
type CommentRef struct {
	ID        int       `json:"id"`
	Body      string    `json:"body"`
	User      UserRef   `json:"user"`
	CreatedAt time.Time `json:"created_at"`
}

// ProjectV2Ref represents a GitHub Project v2 reference in GraphQL API
type ProjectV2Ref struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Number int    `json:"number"`
}

// ToPullRequest converts PRResponse to models.PullRequest
func (pr *PRResponse) ToPullRequest() *models.PullRequest {
	mergeable := false
	if pr.Mergeable != nil {
		mergeable = *pr.Mergeable
	}

	return &models.PullRequest{
		Number:    pr.Number,
		Title:     pr.Title,
		Body:      pr.Body,
		State:     pr.State,
		Draft:     pr.Draft,
		Merged:    pr.Merged,
		Mergeable: mergeable,
		URL:       pr.HTMLURL,
		Head: models.Branch{
			Ref: pr.Head.Ref,
			SHA: pr.Head.SHA,
			Repo: models.Repo{
				Name:     pr.Head.Repo.Name,
				FullName: pr.Head.Repo.FullName,
			},
		},
		Base: models.Branch{
			Ref: pr.Base.Ref,
			SHA: pr.Base.SHA,
			Repo: models.Repo{
				Name:     pr.Base.Repo.Name,
				FullName: pr.Base.Repo.FullName,
			},
		},
		User: models.GitHubUser{
			Login: pr.User.Login,
			ID:    pr.User.ID,
		},
		CreatedAt: pr.CreatedAt,
		UpdatedAt: pr.UpdatedAt,
		MergedAt:  pr.MergedAt,
	}
}

// ToIssue converts IssueResponse to models.Issue
func (ir *IssueResponse) ToIssue() *models.Issue {
	issue := &models.Issue{
		Number: ir.Number,
		Title:  ir.Title,
		Body:   ir.Body,
		State:  ir.State,
		URL:    ir.HTMLURL,
		User: models.GitHubUser{
			Login: ir.User.Login,
			ID:    ir.User.ID,
		},
		CreatedAt: ir.CreatedAt,
		UpdatedAt: ir.UpdatedAt,
		ClosedAt:  ir.ClosedAt,
	}

	// Convert assignees
	issue.Assignees = make([]models.GitHubUser, len(ir.Assignees))
	for i, a := range ir.Assignees {
		issue.Assignees[i] = models.GitHubUser{
			Login: a.Login,
			ID:    a.ID,
		}
	}

	// Convert labels
	issue.Labels = make([]models.Label, len(ir.Labels))
	for i, l := range ir.Labels {
		issue.Labels[i] = models.Label{
			Name:        l.Name,
			Color:       l.Color,
			Description: l.Description,
		}
	}

	// Convert milestone
	if ir.Milestone != nil {
		issue.Milestone = &models.Milestone{
			Number:      ir.Milestone.Number,
			Title:       ir.Milestone.Title,
			Description: ir.Milestone.Description,
			State:       ir.Milestone.State,
		}
	}

	// Convert comments
	issue.Comments = make([]models.IssueComment, len(ir.Comments))
	for i, c := range ir.Comments {
		issue.Comments[i] = models.IssueComment{
			ID:   c.ID,
			Body: c.Body,
			User: models.GitHubUser{
				Login: c.User.Login,
				ID:    c.User.ID,
			},
			CreatedAt: c.CreatedAt,
		}
	}

	return issue
}
