// Package models defines data models for the vibe CLI.
package models

import "time"

// Issue represents a GitHub issue
type Issue struct {
	Number    int            `json:"number"`
	Title     string         `json:"title"`
	Body      string         `json:"body"`
	State     string         `json:"state"`
	URL       string         `json:"html_url"`
	User      GitHubUser     `json:"user"`
	Assignees []GitHubUser   `json:"assignees"`
	Labels    []Label        `json:"labels"`
	Milestone *Milestone     `json:"milestone"`
	Comments  []IssueComment `json:"comments"`
	Projects  []ProjectV2    `json:"projects"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	ClosedAt  *time.Time     `json:"closed_at"`
}

// Label represents a GitHub label
type Label struct {
	Name        string `json:"name"`
	Color       string `json:"color"`
	Description string `json:"description"`
}

// Milestone represents a GitHub milestone
type Milestone struct {
	Number      int    `json:"number"`
	Title       string `json:"title"`
	Description string `json:"description"`
	State       string `json:"state"`
}

// IssueComment represents a GitHub issue comment
type IssueComment struct {
	ID        int        `json:"id"`
	Body      string     `json:"body"`
	User      GitHubUser `json:"user"`
	CreatedAt time.Time  `json:"created_at"`
}

// ProjectV2 represents a GitHub Project (Projects v2)
type ProjectV2 struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Number int    `json:"number"`
}

// IssueCreateRequest represents an issue creation request
type IssueCreateRequest struct {
	Title      string
	Body       string
	Assignees  []string
	Labels     []string
	Milestone  string
	ProjectIDs []string
}

// IssueUpdateRequest represents an issue update request
type IssueUpdateRequest struct {
	Title      *string
	Body       *string
	State      *string
	Assignees  *[]string
	Labels     *[]string
	Milestone  *string
	ProjectIDs *[]string
}
