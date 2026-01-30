package github

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

const (
	graphqlURL = "https://api.github.com/graphql"
)

// graphQLRequest represents a GraphQL request
type graphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

// graphQLResponse represents a GraphQL response
type graphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []graphQLError  `json:"errors"`
}

// graphQLError represents a GraphQL error
type graphQLError struct {
	Message string        `json:"message"`
	Type    string        `json:"type"`
	Path    []interface{} `json:"path"`
}

// PartialProjectError represents a partial success when adding to projects
type PartialProjectError struct {
	IssueNumber     int
	SuccessProjects []string
	FailedProjects  []ProjectFailure
}

// ProjectFailure represents a failed project operation
type ProjectFailure struct {
	ProjectID string
	Reason    string
}

// Error implements the error interface
func (e *PartialProjectError) Error() string {
	if len(e.SuccessProjects) == 0 {
		failedIDs := make([]string, len(e.FailedProjects))
		for i, f := range e.FailedProjects {
			failedIDs[i] = f.ProjectID
		}
		return fmt.Sprintf("failed to add issue #%d to all projects: %s", e.IssueNumber, strings.Join(failedIDs, ", "))
	}
	failedIDs := make([]string, len(e.FailedProjects))
	for i, f := range e.FailedProjects {
		failedIDs[i] = f.ProjectID
	}
	return fmt.Sprintf("issue #%d added to %d project(s), but failed for: %s",
		e.IssueNumber, len(e.SuccessProjects), strings.Join(failedIDs, ", "))
}

// executeGraphQL executes a GraphQL query against the GitHub GraphQL API
// This method reuses the existing HTTPClient.DoJSONRequest() for consistency
func (c *HTTPClient) executeGraphQL(ctx context.Context, query string, variables map[string]interface{}, result interface{}) error {
	req := graphQLRequest{
		Query:     query,
		Variables: variables,
	}

	var resp graphQLResponse
	headers := map[string]string{
		"Authorization": fmt.Sprintf("token %s", c.token),
	}

	err := c.httpClient.DoJSONRequest(ctx, "POST", c.graphqlURL, req, &resp, headers)
	if err != nil {
		return fmt.Errorf("GraphQL request failed: %w", err)
	}

	// GraphQL returns 200 OK even with errors in the response body
	if len(resp.Errors) > 0 {
		// Build a comprehensive error message
		var errMsgs []string
		for _, e := range resp.Errors {
			errMsgs = append(errMsgs, e.Message)
		}
		return fmt.Errorf("GraphQL errors: %s", strings.Join(errMsgs, "; "))
	}

	// Unmarshal the data into the result
	if result != nil && len(resp.Data) > 0 {
		if err := json.Unmarshal(resp.Data, result); err != nil {
			return fmt.Errorf("failed to unmarshal GraphQL response: %w", err)
		}
	}

	return nil
}

// projectV2 represents a GitHub Project v2
type projectV2 struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Number int    `json:"number"`
}

// resolveProjectID resolves a project identifier to a node ID
// Supports three formats:
// 1. Node ID (PVT_xxx) - returned as-is
// 2. Project number (5 or #5) - requires GraphQL lookup
// 3. Project name ("Sprint 2024") - requires GraphQL search
func (c *HTTPClient) resolveProjectID(ctx context.Context, projectID string) (string, error) {
	// Fast path: Already a node ID
	if strings.HasPrefix(projectID, "PVT_") || strings.HasPrefix(projectID, "PVTSSF_") {
		return projectID, nil
	}

	// Remove leading # if present
	projectID = strings.TrimPrefix(projectID, "#")

	// Try to parse as number
	if num, err := strconv.Atoi(projectID); err == nil {
		return c.getProjectByNumber(ctx, num)
	}

	// Search by name
	return c.getProjectByName(ctx, projectID)
}

// getProjectByNumber retrieves a project by its number
func (c *HTTPClient) getProjectByNumber(ctx context.Context, number int) (string, error) {
	// Try organization first
	query := `
		query GetOrgProject($owner: String!, $number: Int!) {
			organization(login: $owner) {
				projectV2(number: $number) {
					id
					title
					number
				}
			}
		}
	`

	variables := map[string]interface{}{
		"owner":  c.owner,
		"number": number,
	}

	var result struct {
		Organization struct {
			ProjectV2 *projectV2 `json:"projectV2"`
		} `json:"organization"`
	}

	err := c.executeGraphQL(ctx, query, variables, &result)
	if err == nil && result.Organization.ProjectV2 != nil {
		return result.Organization.ProjectV2.ID, nil
	}

	// Fallback to user projects
	query = `
		query GetUserProject($owner: String!, $number: Int!) {
			user(login: $owner) {
				projectV2(number: $number) {
					id
					title
					number
				}
			}
		}
	`

	var userResult struct {
		User struct {
			ProjectV2 *projectV2 `json:"projectV2"`
		} `json:"user"`
	}

	err = c.executeGraphQL(ctx, query, variables, &userResult)
	if err != nil {
		return "", fmt.Errorf("project #%d not found: %w", number, err)
	}

	if userResult.User.ProjectV2 == nil {
		return "", fmt.Errorf("project #%d not found", number)
	}

	return userResult.User.ProjectV2.ID, nil
}

// getProjectByName retrieves a project by its name
func (c *HTTPClient) getProjectByName(ctx context.Context, name string) (string, error) {
	// Try organization first
	query := `
		query ListOrgProjects($owner: String!, $first: Int!) {
			organization(login: $owner) {
				projectsV2(first: $first) {
					nodes {
						id
						title
						number
					}
				}
			}
		}
	`

	variables := map[string]interface{}{
		"owner": c.owner,
		"first": 100, // Get first 100 projects
	}

	var result struct {
		Organization struct {
			ProjectsV2 struct {
				Nodes []projectV2 `json:"nodes"`
			} `json:"projectsV2"`
		} `json:"organization"`
	}

	err := c.executeGraphQL(ctx, query, variables, &result)
	if err == nil {
		for _, project := range result.Organization.ProjectsV2.Nodes {
			if project.Title == name {
				return project.ID, nil
			}
		}
	}

	// Fallback to user projects
	query = `
		query ListUserProjects($owner: String!, $first: Int!) {
			user(login: $owner) {
				projectsV2(first: $first) {
					nodes {
						id
						title
						number
					}
				}
			}
		}
	`

	var userResult struct {
		User struct {
			ProjectsV2 struct {
				Nodes []projectV2 `json:"nodes"`
			} `json:"projectsV2"`
		} `json:"user"`
	}

	err = c.executeGraphQL(ctx, query, variables, &userResult)
	if err != nil {
		return "", fmt.Errorf("project '%s' not found: %w", name, err)
	}

	for _, project := range userResult.User.ProjectsV2.Nodes {
		if project.Title == name {
			return project.ID, nil
		}
	}

	return "", fmt.Errorf("project '%s' not found", name)
}

// addIssueToProjects adds an issue to one or more Projects v2
// This is a two-step process:
// 1. Resolve each project identifier to a node ID
// 2. Execute the addProjectV2ItemById mutation for each project
//
// Note: Projects are currently processed sequentially for simplicity and reliability.
// For performance optimization with many projects, consider parallelizing with goroutines
// in a future iteration. Sequential processing is sufficient for typical use cases (1-3 projects).
func (c *HTTPClient) addIssueToProjects(ctx context.Context, issueNumber int, issueNodeID string, projectIDs []string) error {
	if len(projectIDs) == 0 {
		return nil
	}

	var successProjects []string
	var failedProjects []ProjectFailure

	for _, projectID := range projectIDs {
		// Resolve project ID to node ID
		nodeID, err := c.resolveProjectID(ctx, projectID)
		if err != nil {
			failedProjects = append(failedProjects, ProjectFailure{
				ProjectID: projectID,
				Reason:    fmt.Sprintf("failed to resolve project: %v", err),
			})
			continue
		}

		// Add issue to project
		mutation := `
			mutation AddToProject($projectId: ID!, $contentId: ID!) {
				addProjectV2ItemById(input: {
					projectId: $projectId
					contentId: $contentId
				}) {
					item {
						id
					}
				}
			}
		`

		variables := map[string]interface{}{
			"projectId": nodeID,
			"contentId": issueNodeID,
		}

		var result struct {
			AddProjectV2ItemByID struct {
				Item struct {
					ID string `json:"id"`
				} `json:"item"`
			} `json:"addProjectV2ItemById"`
		}

		err = c.executeGraphQL(ctx, mutation, variables, &result)
		if err != nil {
			failedProjects = append(failedProjects, ProjectFailure{
				ProjectID: projectID,
				Reason:    fmt.Sprintf("failed to add to project: %v", err),
			})
			continue
		}

		successProjects = append(successProjects, projectID)
	}

	// Return error if any projects failed
	if len(failedProjects) > 0 {
		return &PartialProjectError{
			IssueNumber:     issueNumber,
			SuccessProjects: successProjects,
			FailedProjects:  failedProjects,
		}
	}

	return nil
}
