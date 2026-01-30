package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rithyhuot/vibe/internal/utils"
)

// mustEncode encodes a response to JSON, ignoring errors (for test code)
func mustEncodeGraphQL(w http.ResponseWriter, v interface{}) {
	_ = json.NewEncoder(w).Encode(v)
}

// mustDecode decodes a JSON request, ignoring errors (for test code)
func mustDecodeGraphQL(r *http.Request, v interface{}) {
	_ = json.NewDecoder(r.Body).Decode(v)
}

func TestExecuteGraphQL_Success(t *testing.T) {
	// Create a test server that returns a successful GraphQL response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "token test-token" {
			t.Errorf("Expected Authorization header with token")
		}

		// Return successful response
		resp := graphQLResponse{
			Data: json.RawMessage(`{"test": "data"}`),
		}
		w.Header().Set("Content-Type", "application/json")
		mustEncodeGraphQL(w, resp)
	}))
	defer server.Close()

	// Create client with test server URL
	client := &HTTPClient{
		httpClient: utils.NewHTTPClient(0),
		token:      "test-token",
		owner:      "test-owner",
		repo:       "test-repo",
		baseURL:    server.URL,
		graphqlURL: server.URL + "/graphql",
	}

	// Execute GraphQL query
	var result map[string]string
	err := client.executeGraphQL(context.Background(), "query { test }", nil, &result)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result["test"] != "data" {
		t.Errorf("Expected result.test = 'data', got %v", result["test"])
	}
}

func TestExecuteGraphQL_GraphQLError(t *testing.T) {
	// Create a test server that returns GraphQL errors
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := graphQLResponse{
			Errors: []graphQLError{
				{Message: "Field not found", Type: "NOT_FOUND"},
				{Message: "Permission denied", Type: "FORBIDDEN"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		mustEncodeGraphQL(w, resp)
	}))
	defer server.Close()

	client := &HTTPClient{
		httpClient: utils.NewHTTPClient(0),
		token:      "test-token",
		baseURL:    server.URL,
		graphqlURL: server.URL + "/graphql",
		owner:      "test-owner",
		repo:       "test-repo",
	}

	var result map[string]string
	err := client.executeGraphQL(context.Background(), "query { test }", nil, &result)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "Field not found") || !strings.Contains(errMsg, "Permission denied") {
		t.Errorf("Expected error to contain both error messages, got: %v", errMsg)
	}
}

func TestResolveProjectID_NodeID(t *testing.T) {
	// Node ID should be returned as-is without any API calls
	client := &HTTPClient{
		httpClient: utils.NewHTTPClient(0),
		token:      "test-token",
		owner:      "test-owner",
		repo:       "test-repo",
		baseURL:    "https://api.github.com",
		graphqlURL: "https://api.github.com/graphql",
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"PVT prefix", "PVT_kwDOABC123", "PVT_kwDOABC123"},
		{"PVTSSF prefix", "PVTSSF_kwDOABC123", "PVTSSF_kwDOABC123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := client.resolveProjectID(context.Background(), tt.input)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestResolveProjectID_Number(t *testing.T) {
	// Mock server that returns a project by number
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse request to determine if it's org or user query
		var req graphQLRequest
		mustDecodeGraphQL(r, &req)

		var resp graphQLResponse
		if strings.Contains(req.Query, "organization") {
			// Return organization project
			resp.Data = json.RawMessage(`{
				"organization": {
					"projectV2": {
						"id": "PVT_orgProject",
						"title": "Test Project",
						"number": 5
					}
				}
			}`)
		} else {
			// Return user project
			resp.Data = json.RawMessage(`{
				"user": {
					"projectV2": {
						"id": "PVT_userProject",
						"title": "Test Project",
						"number": 5
					}
				}
			}`)
		}
		mustEncodeGraphQL(w, resp)
	}))
	defer server.Close()

	client := &HTTPClient{
		httpClient: utils.NewHTTPClient(0),
		token:      "test-token",
		baseURL:    server.URL,
		graphqlURL: server.URL + "/graphql",
		owner:      "test-owner",
		repo:       "test-repo",
	}

	tests := []struct {
		name  string
		input string
	}{
		{"Plain number", "5"},
		{"Number with hash", "#5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := client.resolveProjectID(context.Background(), tt.input)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}
			if !strings.HasPrefix(result, "PVT_") {
				t.Errorf("Expected PVT_ prefix in result, got %s", result)
			}
		})
	}
}

func TestResolveProjectID_Name(t *testing.T) {
	// Mock server that returns projects list
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req graphQLRequest
		mustDecodeGraphQL(r, &req)

		var resp graphQLResponse
		if strings.Contains(req.Query, "organization") {
			resp.Data = json.RawMessage(`{
				"organization": {
					"projectsV2": {
						"nodes": [
							{"id": "PVT_project1", "title": "Sprint 2024", "number": 1},
							{"id": "PVT_project2", "title": "Backlog", "number": 2}
						]
					}
				}
			}`)
		} else {
			resp.Data = json.RawMessage(`{
				"user": {
					"projectsV2": {
						"nodes": [
							{"id": "PVT_project1", "title": "Sprint 2024", "number": 1}
						]
					}
				}
			}`)
		}
		mustEncodeGraphQL(w, resp)
	}))
	defer server.Close()

	client := &HTTPClient{
		httpClient: utils.NewHTTPClient(0),
		token:      "test-token",
		baseURL:    server.URL,
		graphqlURL: server.URL + "/graphql",
		owner:      "test-owner",
		repo:       "test-repo",
	}

	result, err := client.resolveProjectID(context.Background(), "Sprint 2024")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result != "PVT_project1" {
		t.Errorf("Expected PVT_project1, got %s", result)
	}
}

func TestResolveProjectID_NotFound(t *testing.T) {
	// Mock server that returns empty results
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req graphQLRequest
		mustDecodeGraphQL(r, &req)

		var resp graphQLResponse
		if strings.Contains(req.Query, "organization") {
			resp.Data = json.RawMessage(`{"organization": {"projectsV2": {"nodes": []}}}`)
		} else {
			resp.Data = json.RawMessage(`{"user": {"projectsV2": {"nodes": []}}}`)
		}
		mustEncodeGraphQL(w, resp)
	}))
	defer server.Close()

	client := &HTTPClient{
		httpClient: utils.NewHTTPClient(0),
		token:      "test-token",
		baseURL:    server.URL,
		graphqlURL: server.URL + "/graphql",
		owner:      "test-owner",
		repo:       "test-repo",
	}

	_, err := client.resolveProjectID(context.Background(), "NonExistent")
	if err == nil {
		t.Fatal("Expected error for non-existent project, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' in error message, got: %v", err)
	}
}

func TestAddIssueToProjects_Success(t *testing.T) {
	// Mock server that handles project resolution and mutation
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req graphQLRequest
		mustDecodeGraphQL(r, &req)

		var resp graphQLResponse
		if strings.Contains(req.Query, "mutation") {
			// Add to project mutation
			resp.Data = json.RawMessage(`{
				"addProjectV2ItemById": {
					"item": {
						"id": "PVTI_item123"
					}
				}
			}`)
		} else if strings.Contains(req.Query, "organization") {
			// Organization project lookup
			resp.Data = json.RawMessage(`{
				"organization": {
					"projectV2": {
						"id": "PVT_project1",
						"title": "Test Project",
						"number": 1
					}
				}
			}`)
		}
		mustEncodeGraphQL(w, resp)
	}))
	defer server.Close()

	client := &HTTPClient{
		httpClient: utils.NewHTTPClient(0),
		token:      "test-token",
		baseURL:    server.URL,
		graphqlURL: server.URL + "/graphql",
		owner:      "test-owner",
		repo:       "test-repo",
	}

	err := client.addIssueToProjects(context.Background(), 123, "MDU6SXNzdWUx", []string{"PVT_project1"})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestAddIssueToProjects_Multiple(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req graphQLRequest
		mustDecodeGraphQL(r, &req)

		var resp graphQLResponse
		if strings.Contains(req.Query, "mutation") {
			callCount++
			resp.Data = json.RawMessage(`{
				"addProjectV2ItemById": {
					"item": {
						"id": "PVTI_item123"
					}
				}
			}`)
		}
		mustEncodeGraphQL(w, resp)
	}))
	defer server.Close()

	client := &HTTPClient{
		httpClient: utils.NewHTTPClient(0),
		token:      "test-token",
		baseURL:    server.URL,
		graphqlURL: server.URL + "/graphql",
		owner:      "test-owner",
		repo:       "test-repo",
	}

	projects := []string{"PVT_project1", "PVT_project2", "PVT_project3"}
	err := client.addIssueToProjects(context.Background(), 126, "MDU6SXNzdWUx", projects)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if callCount != 3 {
		t.Errorf("Expected 3 mutation calls, got %d", callCount)
	}
}

func TestAddIssueToProjects_PartialFailure(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req graphQLRequest
		mustDecodeGraphQL(r, &req)
		callCount++

		var resp graphQLResponse
		if strings.Contains(req.Query, "mutation") {
			// First call succeeds, second fails, third succeeds
			if callCount == 2 {
				resp.Errors = []graphQLError{
					{Message: "Project not found", Type: "NOT_FOUND"},
				}
			} else {
				resp.Data = json.RawMessage(`{
					"addProjectV2ItemById": {
						"item": {
							"id": "PVTI_item123"
						}
					}
				}`)
			}
		}
		mustEncodeGraphQL(w, resp)
	}))
	defer server.Close()

	client := &HTTPClient{
		httpClient: utils.NewHTTPClient(0),
		token:      "test-token",
		baseURL:    server.URL,
		graphqlURL: server.URL + "/graphql",
		owner:      "test-owner",
		repo:       "test-repo",
	}

	projects := []string{"PVT_project1", "PVT_project2", "PVT_project3"}
	err := client.addIssueToProjects(context.Background(), 127, "MDU6SXNzdWUx", projects)

	// Should return PartialProjectError
	if err == nil {
		t.Fatal("Expected PartialProjectError, got nil")
	}

	partialErr, ok := err.(*PartialProjectError)
	if !ok {
		t.Fatalf("Expected *PartialProjectError, got %T", err)
	}

	if len(partialErr.SuccessProjects) != 2 {
		t.Errorf("Expected 2 successful projects, got %d", len(partialErr.SuccessProjects))
	}

	if len(partialErr.FailedProjects) != 1 {
		t.Errorf("Expected 1 failed project, got %d", len(partialErr.FailedProjects))
	}

	if partialErr.IssueNumber != 127 {
		t.Errorf("Expected issue number 127, got %d", partialErr.IssueNumber)
	}
}

func TestAddIssueToProjects_Empty(t *testing.T) {
	client := &HTTPClient{
		httpClient: utils.NewHTTPClient(0),
		token:      "test-token",
		owner:      "test-owner",
		repo:       "test-repo",
		baseURL:    "https://api.github.com",
		graphqlURL: "https://api.github.com/graphql",
	}

	// Should not error with empty project list
	err := client.addIssueToProjects(context.Background(), 130, "MDU6SXNzdWUx", []string{})
	if err != nil {
		t.Fatalf("Expected no error with empty projects, got %v", err)
	}
}

func TestPartialProjectError_Error(t *testing.T) {
	tests := []struct {
		name            string
		err             *PartialProjectError
		expectedContain string
	}{
		{
			name: "All failed",
			err: &PartialProjectError{
				IssueNumber:     123,
				SuccessProjects: []string{},
				FailedProjects: []ProjectFailure{
					{ProjectID: "P1", Reason: "Not found"},
				},
			},
			expectedContain: "failed to add issue #123 to all projects: P1",
		},
		{
			name: "Partial success",
			err: &PartialProjectError{
				IssueNumber:     123,
				SuccessProjects: []string{"P1", "P2"},
				FailedProjects: []ProjectFailure{
					{ProjectID: "P3", Reason: "Permission denied"},
				},
			},
			expectedContain: "added to 2 project(s), but failed for: P3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tt.err.Error()
			if !strings.Contains(msg, tt.expectedContain) {
				t.Errorf("Expected error message to contain '%s', got: %s", tt.expectedContain, msg)
			}
		})
	}
}
