package commands

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/rithyhuot/vibe/internal/services/github"
	"github.com/rithyhuot/vibe/internal/utils"
)

// withRepoFallback executes an operation with the configured GitHub client,
// and falls back to git remote repo if the configured repo is invalid.
// This eliminates duplicate fallback logic across issue and PR commands.
func withRepoFallback[T any](
	ctx *CommandContext,
	s *spinner.Spinner,
	operation func(client github.Client) (T, error),
) (T, error) {
	// Try with configured client first
	result, err := operation(ctx.GitHubClient)

	// If successful, return immediately
	if err == nil {
		return result, nil
	}

	// Check if it's a "Could not resolve to a Repository" error
	if !strings.Contains(err.Error(), "Could not resolve to a Repository") {
		var zero T
		return zero, err
	}

	// Try to get repo from git remote
	owner, repo, repoErr := getRepoFromGitRemote()
	if repoErr != nil {
		// Return original error if we can't get git remote
		var zero T
		return zero, fmt.Errorf("%w (also failed to detect repo from git remote: %v)", err, repoErr)
	}

	// Inform user we're trying the git remote repo
	dim := color.New(color.Faint)
	s.Stop()
	_, _ = dim.Printf("Repository not found in config, trying detected repo: %s/%s\n", owner, repo)
	s.Start()

	// Create a new client with the detected repo
	newClient, err := github.NewClientWithMode(
		ctx.Config.GitHub.Mode,
		ctx.Config.GitHub.Token,
		owner,
		repo,
	)
	if err != nil {
		var zero T
		return zero, fmt.Errorf("failed to create client with detected repo: %w", err)
	}

	// Try again with the new client
	result, err = operation(newClient)
	if err != nil {
		var zero T
		return zero, fmt.Errorf("failed with detected repo %s/%s: %w", owner, repo, err)
	}

	return result, nil
}

// getRepoFromGitRemote extracts owner and repo from git remote URL
func getRepoFromGitRemote() (owner, repo string, err error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	output, cmdErr := cmd.Output()
	if cmdErr != nil {
		return "", "", fmt.Errorf("failed to get git remote: %w", cmdErr)
	}

	remoteURL := strings.TrimSpace(string(output))
	owner, repo, err = utils.GetRepoFromGitRemote(remoteURL)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse git remote '%s': %w", remoteURL, err)
	}

	return owner, repo, nil
}

// parseAndValidateIssueNumber parses and validates an issue number string
func parseAndValidateIssueNumber(issueNumberStr string) (int, error) {
	issueNumber, err := strconv.Atoi(issueNumberStr)
	if err != nil {
		return 0, fmt.Errorf("invalid issue number '%s': must be a positive integer", issueNumberStr)
	}

	if issueNumber <= 0 {
		return 0, fmt.Errorf("invalid issue number %d: must be positive", issueNumber)
	}

	return issueNumber, nil
}
