// Package circleci provides a client for interacting with CircleCI API.
package circleci

import (
	"context"
	"fmt"
	"net/url"
	"os/exec"
	"regexp"
	"strings"

	"github.com/rithyhuot/vibe/internal/utils"
)

const (
	baseURL   = "https://circleci.com/api/v2"
	baseURLV1 = "https://circleci.com/api/v1.1"
)

// Client interface defines CircleCI operations
type Client interface {
	GetPipelinesByBranch(ctx context.Context, projectSlug, branch string) ([]Pipeline, error)
	GetWorkflows(ctx context.Context, pipelineID string) ([]Workflow, error)
	GetJobs(ctx context.Context, workflowID string) ([]Job, error)
	GetJobDetail(ctx context.Context, projectSlug string, jobNumber int) (*JobDetail, error)
	GetTestMetadata(ctx context.Context, projectSlug string, jobNumber int) ([]TestMetadata, error)
	GetBuildDetails(ctx context.Context, projectSlug string, buildNumber int) ([]FailedStep, error)
	GetCIStatusForBranch(ctx context.Context, branch, projectSlug string) (*CIStatus, error)
}

// HTTPClient implements the Client interface using HTTP
type HTTPClient struct {
	httpClient *utils.HTTPClient
	apiToken   string
}

// NewClient creates a new CircleCI HTTP client
func NewClient(apiToken string) *HTTPClient {
	return &HTTPClient{
		httpClient: utils.NewHTTPClient(0).WithUserAgent("vibe"),
		apiToken:   apiToken,
	}
}

// headers returns the common headers for CircleCI API requests
func (c *HTTPClient) headers() map[string]string {
	return map[string]string{
		"Circle-Token": c.apiToken,
		"Content-Type": "application/json",
	}
}

// GetProjectSlug extracts the project slug from git remote
func GetProjectSlug() (string, error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git remote: %w", err)
	}

	urlStr := strings.TrimSpace(string(output))

	// Handle SSH format: git@github.com:org/repo.git
	sshRegex := regexp.MustCompile(`git@github\.com:([^/]+)/([^.]+)`)
	if matches := sshRegex.FindStringSubmatch(urlStr); len(matches) == 3 {
		return fmt.Sprintf("gh/%s/%s", matches[1], matches[2]), nil
	}

	// Handle HTTPS format: https://github.com/org/repo.git
	httpsRegex := regexp.MustCompile(`github\.com/([^/]+)/([^.]+)`)
	if matches := httpsRegex.FindStringSubmatch(urlStr); len(matches) == 3 {
		return fmt.Sprintf("gh/%s/%s", matches[1], matches[2]), nil
	}

	return "", fmt.Errorf("could not parse git remote URL: %s", urlStr)
}

// GetPipelinesByBranch retrieves pipelines for a specific branch
func (c *HTTPClient) GetPipelinesByBranch(ctx context.Context, projectSlug, branch string) ([]Pipeline, error) {
	u := fmt.Sprintf("%s/project/%s/pipeline?branch=%s", baseURL, projectSlug, url.QueryEscape(branch))

	var resp PipelineResponse
	err := c.httpClient.DoJSONRequest(ctx, "GET", u, nil, &resp, c.headers())
	if err != nil {
		return nil, fmt.Errorf("failed to get pipelines: %w", err)
	}

	return resp.Items, nil
}

// GetWorkflows retrieves workflows for a pipeline
func (c *HTTPClient) GetWorkflows(ctx context.Context, pipelineID string) ([]Workflow, error) {
	u := fmt.Sprintf("%s/pipeline/%s/workflow", baseURL, pipelineID)

	var resp WorkflowResponse
	err := c.httpClient.DoJSONRequest(ctx, "GET", u, nil, &resp, c.headers())
	if err != nil {
		return nil, fmt.Errorf("failed to get workflows: %w", err)
	}

	return resp.Items, nil
}

// GetJobs retrieves jobs for a workflow
func (c *HTTPClient) GetJobs(ctx context.Context, workflowID string) ([]Job, error) {
	u := fmt.Sprintf("%s/workflow/%s/job", baseURL, workflowID)

	var resp JobResponse
	err := c.httpClient.DoJSONRequest(ctx, "GET", u, nil, &resp, c.headers())
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs: %w", err)
	}

	return resp.Items, nil
}

// GetJobDetail retrieves detailed information about a job
func (c *HTTPClient) GetJobDetail(ctx context.Context, projectSlug string, jobNumber int) (*JobDetail, error) {
	u := fmt.Sprintf("%s/project/%s/job/%d", baseURL, projectSlug, jobNumber)

	var detail JobDetail
	err := c.httpClient.DoJSONRequest(ctx, "GET", u, nil, &detail, c.headers())
	if err != nil {
		return nil, fmt.Errorf("failed to get job detail: %w", err)
	}

	return &detail, nil
}

// GetTestMetadata retrieves test results for a job
func (c *HTTPClient) GetTestMetadata(ctx context.Context, projectSlug string, jobNumber int) ([]TestMetadata, error) {
	u := fmt.Sprintf("%s/project/%s/%d/tests", baseURL, projectSlug, jobNumber)

	var resp TestMetadataResponse
	err := c.httpClient.DoJSONRequest(ctx, "GET", u, nil, &resp, c.headers())
	if err != nil {
		return nil, fmt.Errorf("failed to get test metadata: %w", err)
	}

	return resp.Items, nil
}

// GetBuildDetails retrieves build details including step output (v1.1 API)
func (c *HTTPClient) GetBuildDetails(ctx context.Context, projectSlug string, buildNumber int) ([]FailedStep, error) {
	// Convert project slug from "gh/org/repo" to "github/org/repo" for v1.1 API
	parts := strings.Split(projectSlug, "/")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid project slug: %s", projectSlug)
	}

	vcsShort := parts[0]
	vcsType := vcsShort
	if vcsShort == "gh" {
		vcsType = "github"
	}

	org := parts[1]
	repo := parts[2]

	u := fmt.Sprintf("%s/project/%s/%s/%s/%d", baseURLV1, vcsType, org, repo, buildNumber)

	var details BuildDetails
	err := c.httpClient.DoJSONRequest(ctx, "GET", u, nil, &details, c.headers())
	if err != nil {
		return nil, fmt.Errorf("failed to get build details: %w", err)
	}

	// Extract failed steps
	var failedSteps []FailedStep

	for _, step := range details.Steps {
		var failedActions []FailedAction

		for _, action := range step.Actions {
			if action.Status == "failed" {
				var output string

				// Fetch output if available
				if action.OutputURL != "" {
					outputData, err := c.fetchStepOutput(ctx, action.OutputURL)
					if err == nil {
						output = outputData
					}
				}

				failedActions = append(failedActions, FailedAction{
					Name:     action.Name,
					Status:   action.Status,
					ExitCode: action.ExitCode,
					Output:   output,
				})
			}
		}

		if len(failedActions) > 0 {
			failedSteps = append(failedSteps, FailedStep{
				Name:    step.Name,
				Actions: failedActions,
			})
		}
	}

	return failedSteps, nil
}

// fetchStepOutput fetches the output for a step action
func (c *HTTPClient) fetchStepOutput(ctx context.Context, outputURL string) (string, error) {
	var messages []OutputMessage
	err := c.httpClient.DoJSONRequest(ctx, "GET", outputURL, nil, &messages, c.headers())
	if err != nil {
		return "", err
	}

	var output strings.Builder
	for _, msg := range messages {
		output.WriteString(msg.Message)
	}

	return output.String(), nil
}

// GetCIStatusForBranch retrieves comprehensive CI status for a branch
func (c *HTTPClient) GetCIStatusForBranch(ctx context.Context, branch, projectSlug string) (*CIStatus, error) {
	// Get most recent pipeline for the branch
	pipelines, err := c.GetPipelinesByBranch(ctx, projectSlug, branch)
	if err != nil {
		return nil, err
	}

	if len(pipelines) == 0 {
		return nil, nil
	}

	pipeline := pipelines[0]

	// Get all workflows for this pipeline
	workflows, err := c.GetWorkflows(ctx, pipeline.ID)
	if err != nil {
		return nil, err
	}

	if len(workflows) == 0 {
		return nil, nil
	}

	// Fetch jobs for all workflows
	var workflowStatuses []WorkflowStatus
	var failedJobs []FailedJob

	for _, workflow := range workflows {
		jobs, err := c.GetJobs(ctx, workflow.ID)
		if err != nil {
			continue // Skip workflows we can't fetch jobs for
		}

		workflowStatuses = append(workflowStatuses, WorkflowStatus{
			ID:     workflow.ID,
			Name:   workflow.Name,
			Status: workflow.Status,
			Jobs:   jobs,
		})

		// Find failed jobs
		for _, job := range jobs {
			if job.Status == "failed" {
				detail, err := c.GetJobDetail(ctx, projectSlug, job.JobNumber)
				if err != nil {
					// If we can't get details, include basic info
					failedJobs = append(failedJobs, FailedJob{
						Name:         job.Name,
						JobNumber:    job.JobNumber,
						WebURL:       fmt.Sprintf("https://app.circleci.com/pipelines/%s/%d/workflows/%s/jobs/%d", projectSlug, pipeline.Number, workflow.ID, job.JobNumber),
						WorkflowName: workflow.Name,
						FailedTests:  []TestMetadata{},
					})
					continue
				}

				// Get test metadata
				var failedTests []TestMetadata
				tests, err := c.GetTestMetadata(ctx, projectSlug, job.JobNumber)
				if err == nil {
					for _, test := range tests {
						if test.Result == "failure" {
							failedTests = append(failedTests, test)
						}
					}
				}

				failedJobs = append(failedJobs, FailedJob{
					Name:         job.Name,
					JobNumber:    job.JobNumber,
					WebURL:       detail.WebURL,
					WorkflowName: workflow.Name,
					FailedTests:  failedTests,
				})
			}
		}
	}

	return &CIStatus{
		Branch:         branch,
		ProjectSlug:    projectSlug,
		PipelineNumber: pipeline.Number,
		PipelineID:     pipeline.ID,
		Workflows:      workflowStatuses,
		FailedJobs:     failedJobs,
	}, nil
}
