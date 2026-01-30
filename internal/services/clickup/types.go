package clickup

import (
	"github.com/rithyhuot/vibe/internal/models"
)

// API response structures

// TaskResponse wraps a single task response
type TaskResponse struct {
	ID           string                `json:"id"`
	Name         string                `json:"name"`
	Description  string                `json:"description"`
	Status       StatusResponse        `json:"status"`
	Priority     *PriorityResponse     `json:"priority"`
	DueDate      *string               `json:"due_date"`
	StartDate    *string               `json:"start_date"`
	TimeSpent    int64                 `json:"time_spent"`
	Assignees    []UserResponse        `json:"assignees"`
	Tags         []TagResponse         `json:"tags"`
	CustomFields []CustomFieldResponse `json:"custom_fields"`
	URL          string                `json:"url"`
	List         ListResponse          `json:"list"`
	Folder       FolderResponse        `json:"folder"`
	Space        SpaceResponse         `json:"space"`
}

// TasksResponse wraps a list of tasks
type TasksResponse struct {
	Tasks []TaskResponse `json:"tasks"`
}

// StatusResponse represents a status in API responses
type StatusResponse struct {
	Status string `json:"status"`
	Color  string `json:"color"`
	Type   string `json:"type"`
}

// PriorityResponse represents priority in API responses
type PriorityResponse struct {
	ID       string `json:"id"`
	Priority string `json:"priority"`
	Color    string `json:"color"`
}

// UserResponse represents a user in API responses
type UserResponse struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Color    string `json:"color"`
}

// TagResponse represents a tag in API responses
type TagResponse struct {
	Name string `json:"name"`
	FG   string `json:"tag_fg"`
	BG   string `json:"tag_bg"`
}

// CustomFieldResponse represents a custom field in API responses
type CustomFieldResponse struct {
	ID    string      `json:"id"`
	Name  string      `json:"name"`
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

// ListResponse represents a list in API responses
type ListResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// FolderResponse represents a folder in API responses
type FolderResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// SpaceResponse represents a space in API responses
type SpaceResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// CommentResponse wraps a comment response
type CommentResponse struct {
	ID          string            `json:"id"`
	Comment     []ContentResponse `json:"comment"`
	CommentText string            `json:"comment_text"`
	User        UserResponse      `json:"user"`
	Date        string            `json:"date"`
}

// ContentResponse represents comment content
type ContentResponse struct {
	Text string `json:"text"`
}

// FoldersResponse wraps a list of folders
type FoldersResponse struct {
	Folders []FolderResponse `json:"folders"`
}

// ToTask converts TaskResponse to models.Task
func (tr *TaskResponse) ToTask() *models.Task {
	task := &models.Task{
		ID:          tr.ID,
		Name:        tr.Name,
		Description: tr.Description,
		Status: models.Status{
			Status: tr.Status.Status,
			Color:  tr.Status.Color,
			Type:   tr.Status.Type,
		},
		TimeSpent: tr.TimeSpent,
		URL:       tr.URL,
		ListID:    tr.List.ID,
		FolderID:  tr.Folder.ID,
		SpaceID:   tr.Space.ID,
	}

	if tr.Priority != nil {
		task.Priority = &models.Priority{
			ID:       tr.Priority.ID,
			Priority: tr.Priority.Priority,
			Color:    tr.Priority.Color,
		}
	}

	for _, a := range tr.Assignees {
		task.Assignees = append(task.Assignees, models.User{
			ID:       a.ID,
			Username: a.Username,
			Email:    a.Email,
			Color:    a.Color,
		})
	}

	for _, t := range tr.Tags {
		task.Tags = append(task.Tags, models.Tag{
			Name: t.Name,
			FG:   t.FG,
			BG:   t.BG,
		})
	}

	for _, cf := range tr.CustomFields {
		task.CustomFields = append(task.CustomFields, models.CustomField{
			ID:    cf.ID,
			Name:  cf.Name,
			Type:  cf.Type,
			Value: cf.Value,
		})
	}

	return task
}
