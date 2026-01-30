package models

import "time"

// Task represents a ClickUp task
type Task struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Description  string        `json:"description"`
	Status       Status        `json:"status"`
	Priority     *Priority     `json:"priority"`
	DueDate      *time.Time    `json:"due_date"`
	StartDate    *time.Time    `json:"start_date"`
	TimeSpent    int64         `json:"time_spent"`
	Assignees    []User        `json:"assignees"`
	Tags         []Tag         `json:"tags"`
	CustomFields []CustomField `json:"custom_fields"`
	URL          string        `json:"url"`
	ListID       string        `json:"list_id"`
	FolderID     string        `json:"folder_id"`
	SpaceID      string        `json:"space_id"`
}

// Status represents a task status
type Status struct {
	Status string `json:"status"`
	Color  string `json:"color"`
	Type   string `json:"type"`
}

// Priority represents task priority
type Priority struct {
	ID       string `json:"id"`
	Priority string `json:"priority"`
	Color    string `json:"color"`
}

// User represents a ClickUp user
type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Color    string `json:"color"`
}

// Tag represents a task tag
type Tag struct {
	Name string `json:"name"`
	FG   string `json:"tag_fg"`
	BG   string `json:"tag_bg"`
}

// CustomField represents a custom field on a task
type CustomField struct {
	ID    string      `json:"id"`
	Name  string      `json:"name"`
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

// GetCustomField retrieves a custom field value by name
func (t *Task) GetCustomField(name string) *CustomField {
	for _, field := range t.CustomFields {
		if field.Name == name {
			return &field
		}
	}
	return nil
}

// GetCustomFieldString retrieves a custom field value as string
func (t *Task) GetCustomFieldString(name string) string {
	field := t.GetCustomField(name)
	if field == nil || field.Value == nil {
		return ""
	}
	if str, ok := field.Value.(string); ok {
		return str
	}
	return ""
}

// TaskCreateRequest represents a request to create a task
type TaskCreateRequest struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description,omitempty"`
	Status       string                 `json:"status,omitempty"`
	Priority     int                    `json:"priority,omitempty"`
	DueDate      *time.Time             `json:"due_date,omitempty"`
	Assignees    []int                  `json:"assignees,omitempty"`
	Tags         []string               `json:"tags,omitempty"`
	CustomFields map[string]interface{} `json:"custom_fields,omitempty"`
}

// TaskUpdateRequest represents a request to update a task
type TaskUpdateRequest struct {
	Name         *string                `json:"name,omitempty"`
	Description  *string                `json:"description,omitempty"`
	Status       *string                `json:"status,omitempty"`
	Priority     *int                   `json:"priority,omitempty"`
	Assignees    *TaskUpdateAssignees   `json:"assignees,omitempty"`
	CustomFields map[string]interface{} `json:"custom_fields,omitempty"`
}

// TaskUpdateAssignees represents assignee updates
type TaskUpdateAssignees struct {
	Add []int `json:"add,omitempty"`
	Rem []int `json:"rem,omitempty"`
}

// Comment represents a task comment
type Comment struct {
	ID          string    `json:"id"`
	Comment     []Content `json:"comment"`
	CommentText string    `json:"comment_text"`
	User        User      `json:"user"`
	DateCreated string    `json:"date"`
}

// Content represents comment content
type Content struct {
	Text string `json:"text"`
}

// CommentRequest represents a request to add a comment
type CommentRequest struct {
	CommentText string `json:"comment_text"`
	Notify      bool   `json:"notify_all,omitempty"`
}
