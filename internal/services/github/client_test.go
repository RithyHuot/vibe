package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rithyhuot/vibe/internal/models"
	"github.com/rithyhuot/vibe/internal/utils"
)

// setupTestServer creates a test HTTP server for integration testing
func setupTestServer(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

// createTestClient creates a test HTTPClient configured to use the test server
func createTestClient(serverURL string) *HTTPClient {
	return &HTTPClient{
		httpClient: utils.NewHTTPClient(0),
		token:      "test-token",
		owner:      "test-owner",
		repo:       "test-repo",
		baseURL:    serverURL,
		graphqlURL: serverURL + "/graphql",
	}
}

// mustEncode encodes a response to JSON, ignoring errors (for test code)
func mustEncode(w http.ResponseWriter, v interface{}) {
	_ = json.NewEncoder(w).Encode(v)
}

// mustDecode decodes a JSON request, ignoring errors (for test code)
func mustDecode(r *http.Request, v interface{}) {
	_ = json.NewDecoder(r.Body).Decode(v)
}

func TestCreateIssue_WithProjectNodeID(t *testing.T) {
	issueCreated := false
	projectAdded := false

	server := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/issues") && r.Method == "POST" {
			// REST API: Create issue
			issueCreated = true
			resp := IssueResponse{
				Number:  123,
				Title:   "Test Issue",
				Body:    "Test body",
				State:   "open",
				HTMLURL: "https://github.com/test/test/issues/123",
				NodeID:  "MDU6SXNzdWUx",
			}
			mustEncode(w, resp)
		} else if strings.Contains(r.URL.Path, "/graphql") {
			// GraphQL API: Add to project
			var req graphQLRequest
			mustDecode(r, &req)

			if strings.Contains(req.Query, "addProjectV2ItemById") {
				projectAdded = true
				resp := graphQLResponse{
					Data: json.RawMessage(`{
						"addProjectV2ItemById": {
							"item": {"id": "PVTI_item123"}
						}
					}`),
				}
				mustEncode(w, resp)
			}
		}
	})
	defer server.Close()

	client := createTestClient(server.URL)

	req := &models.IssueCreateRequest{
		Title:      "Test Issue",
		Body:       "Test body",
		ProjectIDs: []string{"PVT_kwDOABC123"},
	}

	issue, err := client.CreateIssue(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !issueCreated {
		t.Error("Expected issue to be created")
	}

	if !projectAdded {
		t.Error("Expected issue to be added to project")
	}

	if issue.Number != 123 {
		t.Errorf("Expected issue number 123, got %d", issue.Number)
	}
}

func TestCreateIssue_WithProjectNumber(t *testing.T) {
	projectResolved := false
	projectAdded := false

	server := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/issues") && r.Method == "POST" {
			resp := IssueResponse{
				Number:  124,
				Title:   "Test Issue",
				Body:    "Test body",
				State:   "open",
				HTMLURL: "https://github.com/test/test/issues/124",
				NodeID:  "MDU6SXNzdWUy",
			}
			mustEncode(w, resp)
		} else if strings.Contains(r.URL.Path, "/graphql") {
			var req graphQLRequest
			mustDecode(r, &req)

			if strings.Contains(req.Query, "organization") && strings.Contains(req.Query, "projectV2(number:") {
				// Resolve project by number
				projectResolved = true
				resp := graphQLResponse{
					Data: json.RawMessage(`{
						"organization": {
							"projectV2": {
								"id": "PVT_resolved123",
								"title": "Test Project",
								"number": 5
							}
						}
					}`),
				}
				mustEncode(w, resp)
			} else if strings.Contains(req.Query, "addProjectV2ItemById") {
				projectAdded = true
				resp := graphQLResponse{
					Data: json.RawMessage(`{
						"addProjectV2ItemById": {
							"item": {"id": "PVTI_item124"}
						}
					}`),
				}
				mustEncode(w, resp)
			}
		}
	})
	defer server.Close()

	client := createTestClient(server.URL)

	req := &models.IssueCreateRequest{
		Title:      "Test Issue",
		Body:       "Test body",
		ProjectIDs: []string{"5"},
	}

	issue, err := client.CreateIssue(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !projectResolved {
		t.Error("Expected project to be resolved by number")
	}

	if !projectAdded {
		t.Error("Expected issue to be added to project")
	}

	if issue.Number != 124 {
		t.Errorf("Expected issue number 124, got %d", issue.Number)
	}
}

func TestCreateIssue_WithProjectName(t *testing.T) {
	projectSearched := false
	projectAdded := false

	server := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/issues") && r.Method == "POST" {
			resp := IssueResponse{
				Number:  125,
				Title:   "Test Issue",
				Body:    "Test body",
				State:   "open",
				HTMLURL: "https://github.com/test/test/issues/125",
				NodeID:  "MDU6SXNzdWUz",
			}
			mustEncode(w, resp)
		} else if strings.Contains(r.URL.Path, "/graphql") {
			var req graphQLRequest
			mustDecode(r, &req)

			if strings.Contains(req.Query, "projectsV2(first:") {
				// Search projects by name
				projectSearched = true
				resp := graphQLResponse{
					Data: json.RawMessage(`{
						"organization": {
							"projectsV2": {
								"nodes": [
									{"id": "PVT_sprint", "title": "Sprint 2024", "number": 1},
									{"id": "PVT_backlog", "title": "Backlog", "number": 2}
								]
							}
						}
					}`),
				}
				mustEncode(w, resp)
			} else if strings.Contains(req.Query, "addProjectV2ItemById") {
				projectAdded = true
				resp := graphQLResponse{
					Data: json.RawMessage(`{
						"addProjectV2ItemById": {
							"item": {"id": "PVTI_item125"}
						}
					}`),
				}
				mustEncode(w, resp)
			}
		}
	})
	defer server.Close()

	client := createTestClient(server.URL)

	req := &models.IssueCreateRequest{
		Title:      "Test Issue",
		Body:       "Test body",
		ProjectIDs: []string{"Sprint 2024"},
	}

	issue, err := client.CreateIssue(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !projectSearched {
		t.Error("Expected projects to be searched by name")
	}

	if !projectAdded {
		t.Error("Expected issue to be added to project")
	}

	if issue.Number != 125 {
		t.Errorf("Expected issue number 125, got %d", issue.Number)
	}
}

func TestCreateIssue_MultipleProjects(t *testing.T) {
	addCount := 0

	server := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/issues") && r.Method == "POST" {
			resp := IssueResponse{
				Number:  126,
				Title:   "Test Issue",
				Body:    "Test body",
				State:   "open",
				HTMLURL: "https://github.com/test/test/issues/126",
				NodeID:  "MDU6SXNzdWU0",
			}
			mustEncode(w, resp)
		} else if strings.Contains(r.URL.Path, "/graphql") {
			var req graphQLRequest
			mustDecode(r, &req)

			if strings.Contains(req.Query, "addProjectV2ItemById") {
				addCount++
				resp := graphQLResponse{
					Data: json.RawMessage(`{
						"addProjectV2ItemById": {
							"item": {"id": "PVTI_item126"}
						}
					}`),
				}
				mustEncode(w, resp)
			}
		}
	})
	defer server.Close()

	client := createTestClient(server.URL)

	req := &models.IssueCreateRequest{
		Title:      "Test Issue",
		Body:       "Test body",
		ProjectIDs: []string{"PVT_project1", "PVT_project2", "PVT_project3"},
	}

	issue, err := client.CreateIssue(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if addCount != 3 {
		t.Errorf("Expected 3 project additions, got %d", addCount)
	}

	if issue.Number != 126 {
		t.Errorf("Expected issue number 126, got %d", issue.Number)
	}
}

func TestCreateIssue_InvalidProject(t *testing.T) {
	issueCreated := false

	server := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/issues") && r.Method == "POST" {
			issueCreated = true
			resp := IssueResponse{
				Number:  127,
				Title:   "Test Issue",
				Body:    "Test body",
				State:   "open",
				HTMLURL: "https://github.com/test/test/issues/127",
				NodeID:  "MDU6SXNzdWU1",
			}
			mustEncode(w, resp)
		} else if strings.Contains(r.URL.Path, "/graphql") {
			// Return error for invalid project
			resp := graphQLResponse{
				Errors: []graphQLError{
					{Message: "Project not found", Type: "NOT_FOUND"},
				},
			}
			mustEncode(w, resp)
		}
	})
	defer server.Close()

	client := createTestClient(server.URL)

	req := &models.IssueCreateRequest{
		Title:      "Test Issue",
		Body:       "Test body",
		ProjectIDs: []string{"PVT_invalid"},
	}

	issue, err := client.CreateIssue(context.Background(), req)

	// Issue should be created
	if !issueCreated {
		t.Error("Expected issue to be created")
	}

	// But error should be returned about project failure
	if err == nil {
		t.Fatal("Expected error about project failure, got nil")
	}

	if !strings.Contains(err.Error(), "failed to add to projects") {
		t.Errorf("Expected error about project failure, got: %v", err)
	}

	// Issue should still be returned
	if issue == nil {
		t.Fatal("Expected issue to be returned even with project error")
	}

	if issue.Number != 127 {
		t.Errorf("Expected issue number 127, got %d", issue.Number)
	}
}

func TestCreateIssue_ProjectsFailIssueSucceeds(t *testing.T) {
	issueCreated := false

	server := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/issues") && r.Method == "POST" {
			issueCreated = true
			resp := IssueResponse{
				Number:  128,
				Title:   "Test Issue",
				Body:    "Test body",
				State:   "open",
				HTMLURL: "https://github.com/test/test/issues/128",
				NodeID:  "MDU6SXNzdWU2",
			}
			mustEncode(w, resp)
		} else if strings.Contains(r.URL.Path, "/graphql") {
			// GraphQL fails
			resp := graphQLResponse{
				Errors: []graphQLError{
					{Message: "Permission denied", Type: "FORBIDDEN"},
				},
			}
			mustEncode(w, resp)
		}
	})
	defer server.Close()

	client := createTestClient(server.URL)

	req := &models.IssueCreateRequest{
		Title:      "Test Issue",
		Body:       "Test body",
		ProjectIDs: []string{"PVT_project1"},
	}

	issue, err := client.CreateIssue(context.Background(), req)

	// Issue should be created
	if !issueCreated {
		t.Error("Expected issue to be created")
	}

	// Error should indicate partial success
	if err == nil {
		t.Fatal("Expected error about project failure, got nil")
	}

	// Issue should be returned with valid data
	if issue == nil {
		t.Fatal("Expected issue to be returned")
	}

	if issue.Number != 128 {
		t.Errorf("Expected issue number 128, got %d", issue.Number)
	}
}

func TestUpdateIssue_WithProjects(t *testing.T) {
	issueUpdated := false
	projectAdded := false

	server := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/issues/") && r.Method == "PATCH" {
			issueUpdated = true
			resp := IssueResponse{
				Number:  129,
				Title:   "Updated Issue",
				Body:    "Updated body",
				State:   "open",
				HTMLURL: "https://github.com/test/test/issues/129",
				NodeID:  "MDU6SXNzdWU3",
			}
			mustEncode(w, resp)
		} else if strings.Contains(r.URL.Path, "/graphql") {
			var req graphQLRequest
			mustDecode(r, &req)

			if strings.Contains(req.Query, "addProjectV2ItemById") {
				projectAdded = true
				resp := graphQLResponse{
					Data: json.RawMessage(`{
						"addProjectV2ItemById": {
							"item": {"id": "PVTI_item129"}
						}
					}`),
				}
				mustEncode(w, resp)
			}
		}
	})
	defer server.Close()

	client := createTestClient(server.URL)

	title := "Updated Issue"
	body := "Updated body"
	projectIDs := []string{"PVT_project1"}

	req := &models.IssueUpdateRequest{
		Title:      &title,
		Body:       &body,
		ProjectIDs: &projectIDs,
	}

	issue, err := client.UpdateIssue(context.Background(), 129, req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !issueUpdated {
		t.Error("Expected issue to be updated")
	}

	if !projectAdded {
		t.Error("Expected issue to be added to project")
	}

	if issue.Number != 129 {
		t.Errorf("Expected issue number 129, got %d", issue.Number)
	}

	if issue.Title != "Updated Issue" {
		t.Errorf("Expected title 'Updated Issue', got %s", issue.Title)
	}
}

func TestCreateIssue_NoProjects(t *testing.T) {
	graphQLCalled := false

	server := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/issues") && r.Method == "POST" {
			resp := IssueResponse{
				Number:  130,
				Title:   "Test Issue",
				Body:    "Test body",
				State:   "open",
				HTMLURL: "https://github.com/test/test/issues/130",
				NodeID:  "MDU6SXNzdWU4",
			}
			mustEncode(w, resp)
		} else if strings.Contains(r.URL.Path, "/graphql") {
			graphQLCalled = true
		}
	})
	defer server.Close()

	client := createTestClient(server.URL)

	req := &models.IssueCreateRequest{
		Title:      "Test Issue",
		Body:       "Test body",
		ProjectIDs: []string{}, // Empty projects
	}

	issue, err := client.CreateIssue(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if graphQLCalled {
		t.Error("Expected GraphQL not to be called with empty projects")
	}

	if issue.Number != 130 {
		t.Errorf("Expected issue number 130, got %d", issue.Number)
	}
}

func TestCreateIssue_MissingNodeID(t *testing.T) {
	server := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/issues") && r.Method == "POST" {
			// Return issue without NodeID
			resp := IssueResponse{
				Number:  131,
				Title:   "Test Issue",
				Body:    "Test body",
				State:   "open",
				HTMLURL: "https://github.com/test/test/issues/131",
				NodeID:  "", // Missing NodeID
			}
			mustEncode(w, resp)
		}
	})
	defer server.Close()

	client := createTestClient(server.URL)

	req := &models.IssueCreateRequest{
		Title:      "Test Issue",
		Body:       "Test body",
		ProjectIDs: []string{"PVT_project1"}, // Projects specified
	}

	issue, err := client.CreateIssue(context.Background(), req)

	// Issue should be created
	if issue == nil {
		t.Fatal("Expected issue to be returned")
	}

	if issue.Number != 131 {
		t.Errorf("Expected issue number 131, got %d", issue.Number)
	}

	// But error should indicate missing node_id
	if err == nil {
		t.Fatal("Expected error about missing node_id, got nil")
	}

	if !strings.Contains(err.Error(), "did not return node_id") {
		t.Errorf("Expected error about missing node_id, got: %v", err)
	}
}

func TestCreateIssue_PartialProjectSuccess(t *testing.T) {
	callCount := 0
	issueCreated := false

	server := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/issues") && r.Method == "POST" {
			issueCreated = true
			resp := IssueResponse{
				Number:  132,
				Title:   "Test Issue",
				Body:    "Test body",
				State:   "open",
				HTMLURL: "https://github.com/test/test/issues/132",
				NodeID:  "MDU6SXNzdWU5",
			}
			mustEncode(w, resp)
		} else if strings.Contains(r.URL.Path, "/graphql") {
			var req graphQLRequest
			mustDecode(r, &req)
			callCount++

			var resp graphQLResponse
			if strings.Contains(req.Query, "addProjectV2ItemById") {
				// First and third projects succeed, second fails
				if callCount == 2 {
					resp.Errors = []graphQLError{
						{Message: "Project not found", Type: "NOT_FOUND"},
					}
				} else {
					resp.Data = json.RawMessage(`{
						"addProjectV2ItemById": {
							"item": {"id": "PVTI_item132"}
						}
					}`)
				}
			}
			mustEncode(w, resp)
		}
	})
	defer server.Close()

	client := createTestClient(server.URL)

	req := &models.IssueCreateRequest{
		Title:      "Test Issue",
		Body:       "Test body",
		ProjectIDs: []string{"PVT_project1", "PVT_project2", "PVT_project3"},
	}

	issue, err := client.CreateIssue(context.Background(), req)

	// Issue should be created
	if !issueCreated {
		t.Error("Expected issue to be created")
	}

	if issue == nil {
		t.Fatal("Expected issue to be returned")
	}

	if issue.Number != 132 {
		t.Errorf("Expected issue number 132, got %d", issue.Number)
	}

	// Error should indicate partial success
	if err == nil {
		t.Fatal("Expected error about partial project failure, got nil")
	}

	if !strings.Contains(err.Error(), "added to 2 project(s)") {
		t.Errorf("Expected error about partial success, got: %v", err)
	}

	if !strings.Contains(err.Error(), "failed for: PVT_project2") {
		t.Errorf("Expected error to mention failed project, got: %v", err)
	}
}
