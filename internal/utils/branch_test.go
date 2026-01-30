package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateBranchName(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		ticketID string
		title    string
		expected string
	}{
		{
			name:     "simple title",
			prefix:   "john",
			ticketID: "abc123xyz",
			title:    "Add user login",
			expected: "john/abc123xyz/add-user-login",
		},
		{
			name:     "title with special characters",
			prefix:   "jane",
			ticketID: "def456uvw",
			title:    "Fix bug: API timeout!!!",
			expected: "jane/def456uvw/fix-bug-api-timeout",
		},
		{
			name:     "very long title",
			prefix:   "bob",
			ticketID: "ghi789rst",
			title:    "Implement a very long feature name that should be truncated to fit within reasonable limits",
			expected: "bob/ghi789rst/implement-a-very-long-feature-name-that-should-be",
		},
		{
			name:     "title with multiple spaces",
			prefix:   "alice",
			ticketID: "jkl012mno",
			title:    "Update    database    schema",
			expected: "alice/jkl012mno/update-database-schema",
		},
		{
			name:     "title with hyphens",
			prefix:   "tom",
			ticketID: "pqr345stu",
			title:    "Add CI-CD pipeline",
			expected: "tom/pqr345stu/add-ci-cd-pipeline",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateBranchName(tt.prefix, tt.ticketID, tt.title)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateBranchName(t *testing.T) {
	tests := []struct {
		name        string
		branchName  string
		expectError bool
	}{
		{
			name:        "valid branch name",
			branchName:  "john/abc123xyz/add-feature",
			expectError: false,
		},
		{
			name:        "empty name",
			branchName:  "",
			expectError: true,
		},
		{
			name:        "contains shell metacharacters",
			branchName:  "john/abc123xyz/add-feature;rm -rf /",
			expectError: true,
		},
		{
			name:        "contains double dots",
			branchName:  "john/../etc/passwd",
			expectError: true,
		},
		{
			name:        "starts with slash",
			branchName:  "/john/abc123xyz/feature",
			expectError: true,
		},
		{
			name:        "ends with slash",
			branchName:  "john/abc123xyz/feature/",
			expectError: true,
		},
		{
			name:        "contains pipe",
			branchName:  "john/abc123xyz/feature|command",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBranchName(tt.branchName)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExtractTicketID(t *testing.T) {
	tests := []struct {
		name        string
		branchName  string
		expected    string
		expectError bool
	}{
		{
			name:        "valid branch with ticket ID",
			branchName:  "john/abc123xyz/add-feature",
			expected:    "abc123xyz",
			expectError: false,
		},
		{
			name:        "valid branch with different ticket ID",
			branchName:  "jane/xyz789abc/fix-bug",
			expected:    "xyz789abc",
			expectError: false,
		},
		{
			name:        "branch without ticket ID",
			branchName:  "feature/add-something",
			expected:    "",
			expectError: true,
		},
		{
			name:        "main branch",
			branchName:  "main",
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExtractTicketID(tt.branchName)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestIsTicketID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid ticket ID",
			input:    "abc123xyz",
			expected: true,
		},
		{
			name:     "valid ticket ID with numbers only",
			input:    "123456789",
			expected: true,
		},
		{
			name:     "too short",
			input:    "abc123",
			expected: false,
		},
		{
			name:     "too long",
			input:    "abc123xyz01",
			expected: false,
		},
		{
			name:     "contains uppercase",
			input:    "ABC123xyz",
			expected: false,
		},
		{
			name:     "contains special characters",
			input:    "abc-123xy",
			expected: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsTicketID(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeInput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "clean input",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "input with semicolon",
			input:    "hello; rm -rf /",
			expected: "hello rm -rf /",
		},
		{
			name:     "input with pipe",
			input:    "cat file | grep pattern",
			expected: "cat file  grep pattern",
		},
		{
			name:     "input with backticks",
			input:    "echo `whoami`",
			expected: "echo whoami",
		},
		{
			name:     "input with dollar sign",
			input:    "echo $HOME",
			expected: "echo HOME",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeInput(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
