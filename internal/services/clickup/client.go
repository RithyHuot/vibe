// Package clickup provides a client for interacting with ClickUp API.
package clickup

import (
	"context"
	"fmt"
	"net/url"

	"github.com/rithyhuot/vibe/internal/models"
	"github.com/rithyhuot/vibe/internal/utils"
)

const (
	baseURL = "https://api.clickup.com/api/v2"
)

// Client interface defines ClickUp operations
type Client interface {
	GetTask(ctx context.Context, taskID string) (*models.Task, error)
	ListTasks(ctx context.Context, listID string, filters map[string]string) ([]*models.Task, error)
	CreateTask(ctx context.Context, listID string, req *models.TaskCreateRequest) (*models.Task, error)
	UpdateTask(ctx context.Context, taskID string, req *models.TaskUpdateRequest) (*models.Task, error)
	AddComment(ctx context.Context, taskID string, commentText string) (*models.Comment, error)
	GetFolders(ctx context.Context, spaceID string) ([]*models.Folder, error)
	SearchTeamTasks(ctx context.Context, teamID string, searchTerm string) ([]*models.Task, error)
}

// HTTPClient implements the Client interface using HTTP
type HTTPClient struct {
	httpClient *utils.HTTPClient
	apiToken   string
}

// NewClient creates a new ClickUp HTTP client
func NewClient(apiToken string) *HTTPClient {
	return &HTTPClient{
		httpClient: utils.NewHTTPClient(0). // Use default timeout from HTTPClient
							WithUserAgent("vibe"),
		apiToken: apiToken,
	}
}

// headers returns the common headers for ClickUp API requests
func (c *HTTPClient) headers() map[string]string {
	return map[string]string{
		"Authorization": c.apiToken,
		"Content-Type":  "application/json",
	}
}

// GetTask retrieves a single task by ID
func (c *HTTPClient) GetTask(ctx context.Context, taskID string) (*models.Task, error) {
	url := fmt.Sprintf("%s/task/%s", baseURL, taskID)

	var resp TaskResponse
	err := c.httpClient.DoJSONRequest(ctx, "GET", url, nil, &resp, c.headers())
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return resp.ToTask(), nil
}

// ListTasks retrieves tasks from a list with optional filters
func (c *HTTPClient) ListTasks(ctx context.Context, listID string, filters map[string]string) ([]*models.Task, error) {
	u := fmt.Sprintf("%s/list/%s/task", baseURL, listID)

	// Add query parameters
	if len(filters) > 0 {
		query := url.Values{}
		for key, value := range filters {
			query.Add(key, value)
		}
		u = fmt.Sprintf("%s?%s", u, query.Encode())
	}

	var resp TasksResponse
	err := c.httpClient.DoJSONRequest(ctx, "GET", u, nil, &resp, c.headers())
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	tasks := make([]*models.Task, len(resp.Tasks))
	for i, taskResp := range resp.Tasks {
		tasks[i] = taskResp.ToTask()
	}

	return tasks, nil
}

// CreateTask creates a new task in a list
func (c *HTTPClient) CreateTask(ctx context.Context, listID string, req *models.TaskCreateRequest) (*models.Task, error) {
	url := fmt.Sprintf("%s/list/%s/task", baseURL, listID)

	var resp TaskResponse
	err := c.httpClient.DoJSONRequest(ctx, "POST", url, req, &resp, c.headers())
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	return resp.ToTask(), nil
}

// UpdateTask updates an existing task
func (c *HTTPClient) UpdateTask(ctx context.Context, taskID string, req *models.TaskUpdateRequest) (*models.Task, error) {
	url := fmt.Sprintf("%s/task/%s", baseURL, taskID)

	var resp TaskResponse
	err := c.httpClient.DoJSONRequest(ctx, "PUT", url, req, &resp, c.headers())
	if err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	return resp.ToTask(), nil
}

// AddComment adds a comment to a task
func (c *HTTPClient) AddComment(ctx context.Context, taskID string, commentText string) (*models.Comment, error) {
	url := fmt.Sprintf("%s/task/%s/comment", baseURL, taskID)

	req := &models.CommentRequest{
		CommentText: commentText,
		Notify:      false,
	}

	var resp CommentResponse
	err := c.httpClient.DoJSONRequest(ctx, "POST", url, req, &resp, c.headers())
	if err != nil {
		return nil, fmt.Errorf("failed to add comment: %w", err)
	}

	comment := &models.Comment{
		ID:          resp.ID,
		CommentText: resp.CommentText,
		User: models.User{
			ID:       resp.User.ID,
			Username: resp.User.Username,
			Email:    resp.User.Email,
			Color:    resp.User.Color,
		},
		DateCreated: resp.Date,
	}

	for _, c := range resp.Comment {
		comment.Comment = append(comment.Comment, models.Content{
			Text: c.Text,
		})
	}

	return comment, nil
}

// GetFolders retrieves folders from a space
func (c *HTTPClient) GetFolders(ctx context.Context, spaceID string) ([]*models.Folder, error) {
	url := fmt.Sprintf("%s/space/%s/folder", baseURL, spaceID)

	var resp FoldersResponse
	err := c.httpClient.DoJSONRequest(ctx, "GET", url, nil, &resp, c.headers())
	if err != nil {
		return nil, fmt.Errorf("failed to get folders: %w", err)
	}

	folders := make([]*models.Folder, len(resp.Folders))
	for i, f := range resp.Folders {
		folders[i] = &models.Folder{
			ID:   f.ID,
			Name: f.Name,
		}
	}

	return folders, nil
}

// SearchTeamTasks searches for tasks across a team/workspace
func (c *HTTPClient) SearchTeamTasks(ctx context.Context, teamID string, searchTerm string) ([]*models.Task, error) {
	u := fmt.Sprintf("%s/team/%s/task", baseURL, teamID)

	// Add search query parameter
	query := url.Values{}
	if searchTerm != "" {
		query.Add("search", searchTerm)
	}
	// Limit to 50 results for interactive selection
	query.Add("page", "0")
	query.Add("order_by", "updated")
	query.Add("reverse", "true")

	u = fmt.Sprintf("%s?%s", u, query.Encode())

	var resp TasksResponse
	err := c.httpClient.DoJSONRequest(ctx, "GET", u, nil, &resp, c.headers())
	if err != nil {
		return nil, fmt.Errorf("failed to search tasks: %w", err)
	}

	tasks := make([]*models.Task, len(resp.Tasks))
	for i, taskResp := range resp.Tasks {
		tasks[i] = taskResp.ToTask()
	}

	return tasks, nil
}
