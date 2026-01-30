// Package models defines data models for the vibe CLI.
package models

import "time"

// PullRequest represents a GitHub pull request
type PullRequest struct {
	Number    int        `json:"number"`
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	State     string     `json:"state"`
	Draft     bool       `json:"draft"`
	Merged    bool       `json:"merged"`
	Mergeable bool       `json:"mergeable"`
	URL       string     `json:"html_url"`
	Head      Branch     `json:"head"`
	Base      Branch     `json:"base"`
	User      GitHubUser `json:"user"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	MergedAt  *time.Time `json:"merged_at"`
}

// Branch represents a git branch reference
type Branch struct {
	Ref  string `json:"ref"`
	SHA  string `json:"sha"`
	Repo Repo   `json:"repo"`
}

// Repo represents a GitHub repository
type Repo struct {
	Name     string `json:"name"`
	FullName string `json:"full_name"`
}

// GitHubUser represents a GitHub user
type GitHubUser struct {
	Login string `json:"login"`
	ID    int    `json:"id"`
}

// PRStatus represents the status of a pull request
type PRStatus struct {
	Number       int
	State        string
	Draft        bool
	Merged       bool
	Mergeable    bool
	ReviewStatus ReviewStatus
	CheckStatus  CheckStatus
	URL          string
}

// ReviewStatus represents PR review status
type ReviewStatus struct {
	Approved         int
	ChangesRequested int
	Commented        int
	Pending          int
	OverallStatus    string // "approved", "changes_requested", "pending"
}

// CheckStatus represents CI check status
type CheckStatus struct {
	Total         int
	Passed        int
	Failed        int
	Pending       int
	OverallStatus string // "success", "failure", "pending"
}

// PRCreateRequest represents a PR creation request
type PRCreateRequest struct {
	Title string
	Body  string
	Head  string
	Base  string
	Draft bool
}
