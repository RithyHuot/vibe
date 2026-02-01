// Package utils provides utility functions for the vibe CLI.
package utils

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

var (
	// ticketIDPattern matches ticket IDs from branch names
	ticketIDPattern = regexp.MustCompile(`/([a-z0-9]{9})/`)

	// invalidCharsPattern matches characters that shouldn't be in branch names
	invalidCharsPattern = regexp.MustCompile(`[^a-zA-Z0-9\-_/]`)

	// shellMetacharPattern matches potentially dangerous shell metacharacters
	shellMetacharPattern = regexp.MustCompile(`[;&|><$` + "`" + `(){}[\]\\]`)
)

// GenerateBranchName creates a branch name from components
// Format: prefix/ticketID/sanitized-title (when title is provided)
// Format: prefix/ticketID (when title is empty)
// If prefix is empty and username is provided, uses sanitized username as prefix
func GenerateBranchName(prefix, ticketID, title string, username ...string) string {
	// Use username as fallback if prefix is empty
	branchPrefix := prefix
	if branchPrefix == "" && len(username) > 0 && username[0] != "" {
		branchPrefix = SanitizeUsername(username[0])
	}

	// If title is empty, return simple format: prefix/ticketID
	if title == "" {
		return fmt.Sprintf("%s/%s", branchPrefix, ticketID)
	}

	// Sanitize title: lowercase, replace spaces with hyphens, remove invalid chars
	sanitized := strings.ToLower(title)
	sanitized = strings.TrimSpace(sanitized)
	sanitized = regexp.MustCompile(`\s+`).ReplaceAllString(sanitized, "-")
	sanitized = invalidCharsPattern.ReplaceAllString(sanitized, "")

	// Remove multiple consecutive hyphens
	sanitized = regexp.MustCompile(`-+`).ReplaceAllString(sanitized, "-")

	// Trim hyphens from start and end
	sanitized = strings.Trim(sanitized, "-")

	// Limit length
	if len(sanitized) > 50 {
		sanitized = sanitized[:50]
		sanitized = strings.TrimRight(sanitized, "-")
	}

	return fmt.Sprintf("%s/%s/%s", branchPrefix, ticketID, sanitized)
}

// SanitizeUsername sanitizes a username for use in branch names
// Converts to lowercase, replaces spaces with hyphens, removes special characters
func SanitizeUsername(username string) string {
	// Convert to lowercase
	s := strings.ToLower(username)

	// Replace spaces with hyphens
	s = strings.ReplaceAll(s, " ", "-")

	// Remove special characters (keep only alphanumeric and hyphens)
	s = regexp.MustCompile(`[^a-z0-9\-]`).ReplaceAllString(s, "")

	// Remove multiple consecutive hyphens
	s = regexp.MustCompile(`-+`).ReplaceAllString(s, "-")

	// Trim hyphens from start and end
	s = strings.Trim(s, "-")

	return s
}

// SanitizeBranchPart sanitizes a branch part (description) for use in branch names
func SanitizeBranchPart(part string) string {
	// Convert to lowercase
	s := strings.ToLower(part)

	// Trim spaces
	s = strings.TrimSpace(s)

	// Replace spaces with hyphens
	s = regexp.MustCompile(`\s+`).ReplaceAllString(s, "-")

	// Remove special characters (keep only alphanumeric, hyphens, and underscores)
	s = regexp.MustCompile(`[^a-z0-9\-_]`).ReplaceAllString(s, "")

	// Remove multiple consecutive hyphens
	s = regexp.MustCompile(`-+`).ReplaceAllString(s, "-")

	// Trim hyphens from start and end
	s = strings.Trim(s, "-")

	// Limit length
	if len(s) > 50 {
		s = s[:50]
		s = strings.TrimRight(s, "-")
	}

	return s
}

// GetGitUsername retrieves the git user.name from git config
func GetGitUsername() (string, error) {
	cmd := exec.Command("git", "config", "user.name")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git config user.name not set")
	}

	username := strings.TrimSpace(string(output))
	if username == "" {
		return "", fmt.Errorf("git config user.name is empty")
	}

	return username, nil
}

// ResolveBranchPrefix resolves the branch prefix, using git username as fallback if prefix is "your-username"
// Returns the resolved prefix and username (empty string if not using username fallback)
// Returns error if prefix is "your-username" but git username is not configured
func ResolveBranchPrefix(configPrefix string) (prefix string, username string, err error) {
	if configPrefix == "your-username" {
		gitUsername, err := GetGitUsername()
		if err != nil {
			// Return error instead of silently falling back to placeholder
			return "", "", fmt.Errorf("branch prefix is set to 'your-username' but git config user.name is not set: %w", err)
		}
		// Use empty prefix with username fallback
		return "", gitUsername, nil
	}
	return configPrefix, "", nil
}

// ValidateBranchName checks if a branch name is safe
func ValidateBranchName(name string) error {
	if name == "" {
		return fmt.Errorf("branch name cannot be empty")
	}

	if shellMetacharPattern.MatchString(name) {
		return fmt.Errorf("branch name contains unsafe characters")
	}

	if strings.Contains(name, "..") {
		return fmt.Errorf("branch name contains invalid '..' sequence")
	}

	if strings.HasPrefix(name, "/") || strings.HasSuffix(name, "/") {
		return fmt.Errorf("branch name cannot start or end with '/'")
	}

	return nil
}

// ExtractTicketID extracts a ticket ID from a branch name
func ExtractTicketID(branchName string) (string, error) {
	matches := ticketIDPattern.FindStringSubmatch(branchName)
	if len(matches) < 2 {
		return "", fmt.Errorf("no ticket ID found in branch name: %s", branchName)
	}
	return matches[1], nil
}

// IsTicketID checks if a string looks like a valid ticket ID
func IsTicketID(s string) bool {
	if len(s) != 9 {
		return false
	}
	return regexp.MustCompile(`^[a-z0-9]{9}$`).MatchString(s)
}

// SanitizeInput removes potentially dangerous characters from user input
func SanitizeInput(input string) string {
	return shellMetacharPattern.ReplaceAllString(input, "")
}

// GetRepoFromGitRemote extracts owner/repo from git remote URL
// Supports both SSH and HTTPS formats:
// - git@github.com:owner/repo.git
// - https://github.com/owner/repo.git
func GetRepoFromGitRemote(remoteURL string) (owner, repo string, err error) {
	if remoteURL == "" {
		return "", "", fmt.Errorf("empty remote URL")
	}

	// Handle SSH format: git@github.com:owner/repo.git
	if strings.HasPrefix(remoteURL, "git@") {
		parts := strings.Split(remoteURL, ":")
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid SSH remote format")
		}
		path := strings.TrimSuffix(parts[1], ".git")
		pathParts := strings.Split(path, "/")
		if len(pathParts) != 2 {
			return "", "", fmt.Errorf("invalid repository path in SSH remote")
		}
		return pathParts[0], pathParts[1], nil
	}

	// Handle HTTPS format: https://github.com/owner/repo.git
	if path, found := strings.CutPrefix(remoteURL, "https://"); found {
		path, _ = strings.CutPrefix(path, "github.com/")
		path = strings.TrimSuffix(path, ".git")
		parts := strings.Split(path, "/")
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid repository path in HTTPS remote")
		}
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unsupported remote URL format")
}
