package circleci

import "time"

// Pipeline represents a CircleCI pipeline
type Pipeline struct {
	ID        string    `json:"id"`
	Number    int       `json:"number"`
	State     string    `json:"state"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Workflow represents a CircleCI workflow
type Workflow struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Status     string     `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
	StoppedAt  *time.Time `json:"stopped_at"`
	PipelineID string     `json:"pipeline_id"`
}

// Job represents a CircleCI job
type Job struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Status     string     `json:"status"`
	JobNumber  int        `json:"job_number"`
	Type       string     `json:"type"`
	StartedAt  *time.Time `json:"started_at"`
	StoppedAt  *time.Time `json:"stopped_at"`
	ApprovedBy string     `json:"approved_by,omitempty"`
}

// JobDetail represents detailed information about a job
type JobDetail struct {
	ID        string    `json:"id"`
	JobNumber int       `json:"job_number"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	StartedAt time.Time `json:"started_at"`
	StoppedAt time.Time `json:"stopped_at"`
	WebURL    string    `json:"web_url"`
}

// TestMetadata represents test results for a job
type TestMetadata struct {
	Name      string  `json:"name"`
	Classname string  `json:"classname"`
	File      string  `json:"file"`
	Result    string  `json:"result"`
	Message   string  `json:"message"`
	RunTime   float64 `json:"run_time"`
}

// BuildStep represents a build step in v1.1 API
type BuildStep struct {
	Name    string       `json:"name"`
	Actions []StepAction `json:"actions"`
}

// StepAction represents an action within a build step
type StepAction struct {
	Name      string `json:"name"`
	Status    string `json:"status"`
	ExitCode  *int   `json:"exit_code"`
	OutputURL string `json:"output_url,omitempty"`
}

// BuildDetails represents build details from v1.1 API
type BuildDetails struct {
	Steps []BuildStep `json:"steps"`
}

// OutputMessage represents a message in step output
type OutputMessage struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// FailedStep represents a failed step with its output
type FailedStep struct {
	Name    string
	Actions []FailedAction
}

// FailedAction represents a failed action with its output
type FailedAction struct {
	Name     string
	Status   string
	ExitCode *int
	Output   string
}

// FailedJob represents a failed job with test details
type FailedJob struct {
	Name         string
	JobNumber    int
	WebURL       string
	WorkflowName string
	FailedTests  []TestMetadata
}

// WorkflowStatus represents the status of a workflow with its jobs
type WorkflowStatus struct {
	ID     string
	Name   string
	Status string
	Jobs   []Job
}

// CIStatus represents comprehensive CI status for a branch
type CIStatus struct {
	Branch         string
	ProjectSlug    string
	PipelineNumber int
	PipelineID     string
	Workflows      []WorkflowStatus
	FailedJobs     []FailedJob
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Items         interface{} `json:"items"`
	NextPageToken *string     `json:"next_page_token"`
}

// PipelineResponse represents the pipeline list response
type PipelineResponse struct {
	Items         []Pipeline `json:"items"`
	NextPageToken *string    `json:"next_page_token"`
}

// WorkflowResponse represents the workflow list response
type WorkflowResponse struct {
	Items         []Workflow `json:"items"`
	NextPageToken *string    `json:"next_page_token"`
}

// JobResponse represents the job list response
type JobResponse struct {
	Items         []Job   `json:"items"`
	NextPageToken *string `json:"next_page_token"`
}

// TestMetadataResponse represents the test metadata response
type TestMetadataResponse struct {
	Items         []TestMetadata `json:"items"`
	NextPageToken *string        `json:"next_page_token"`
}
