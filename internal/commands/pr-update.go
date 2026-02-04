package commands

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/rithyhuot/vibe/internal/utils"
)

// PRUpdateOptions holds flags for the pr-update command
type PRUpdateOptions struct {
	Title       string
	Summary     string
	Description string
	Testing     string
}

// NewPRUpdateCommand creates the pr-update command
func NewPRUpdateCommand(ctx *CommandContext) *cobra.Command {
	opts := &PRUpdateOptions{}

	cmd := &cobra.Command{
		Use:   "pr-update [pr-number]",
		Short: "Update a pull request",
		Long: `Updates a pull request's title, body, or specific sections. If no PR number is provided, uses the current branch's PR.

Examples:
  vibe pr-update --title "New title"                    # Update PR title
  vibe pr-update --summary "Updated implementation"     # Update summary section
  vibe pr-update 123 --description "New description"    # Update PR #123 description
  vibe pr-update --testing "Run tests with 'make test'" # Update testing section`,
		RunE: func(_ *cobra.Command, args []string) error {
			prNumber := ""
			if len(args) > 0 {
				prNumber = args[0]
			}
			return runPRUpdate(ctx, opts, prNumber)
		},
	}

	cmd.Flags().StringVar(&opts.Title, "title", "", "Update PR title")
	cmd.Flags().StringVar(&opts.Summary, "summary", "", "Update summary section")
	cmd.Flags().StringVar(&opts.Description, "description", "", "Update description section")
	cmd.Flags().StringVar(&opts.Testing, "testing", "", "Update testing section")

	return cmd
}

func runPRUpdate(_ *CommandContext, opts *PRUpdateOptions, prNumberArg string) error {
	// Check if gh CLI is available
	if !isGHAvailable() {
		return fmt.Errorf("GitHub CLI (gh) is not available. Run: gh auth login")
	}

	// Get PR number
	prNumber, err := resolvePRNumber(prNumberArg)
	if err != nil {
		return err
	}

	// Check if any updates were provided
	if !hasUpdates(opts) {
		return fmt.Errorf("no updates provided. Use --title, --summary, --description, or --testing flags")
	}

	// Update PR
	if err := updatePR(prNumber, opts); err != nil {
		return err
	}

	// Show success message
	displayUpdateSuccess(prNumber, opts)
	return nil
}

func resolvePRNumber(prNumberArg string) (string, error) {
	if prNumberArg != "" {
		return prNumberArg, nil
	}

	// Get PR for current branch
	pr, err := getPRForCurrentBranch()
	if err != nil || pr == nil {
		return "", fmt.Errorf("no PR found for this branch")
	}
	return fmt.Sprintf("%d", pr.Number), nil
}

func hasUpdates(opts *PRUpdateOptions) bool {
	return opts.Title != "" || opts.Summary != "" || opts.Description != "" || opts.Testing != ""
}

func updatePR(prNumber string, opts *PRUpdateOptions) error {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Updating PR..."
	s.Start()
	defer s.Stop()

	// Get updated body if needed
	body, err := buildUpdatedBody(prNumber, opts)
	if err != nil {
		return err
	}

	// Build and execute update command
	args := buildUpdateArgs(prNumber, opts.Title, body)
	cmd := exec.Command("gh", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to update PR: %w\nOutput: %s", err, string(output))
	}

	return nil
}

func buildUpdatedBody(prNumber string, opts *PRUpdateOptions) (string, error) {
	if opts.Summary == "" && opts.Description == "" && opts.Testing == "" {
		return "", nil
	}

	body, err := getPRBody(prNumber)
	if err != nil {
		return "", fmt.Errorf("failed to get PR body: %w", err)
	}

	if opts.Summary != "" {
		body = updatePRSection(body, "Summary", opts.Summary)
	}
	if opts.Description != "" {
		body = updatePRSection(body, "Description", opts.Description)
	}
	if opts.Testing != "" {
		body = updatePRSection(body, "How to Test", opts.Testing)
	}

	return body, nil
}

func buildUpdateArgs(prNumber, title, body string) []string {
	args := []string{"pr", "edit", prNumber}
	if title != "" {
		args = append(args, "--title", title)
	}
	if body != "" {
		args = append(args, "--body", body)
	}
	return args
}

func displayUpdateSuccess(prNumber string, opts *PRUpdateOptions) {
	green := color.New(color.FgGreen)
	_, _ = green.Printf("✓ Updated PR #%s\n", prNumber)

	if opts.Title != "" {
		fmt.Printf("  • Title updated\n")
	}
	if opts.Summary != "" {
		fmt.Printf("  • Summary updated\n")
	}
	if opts.Description != "" {
		fmt.Printf("  • Description updated\n")
	}
	if opts.Testing != "" {
		fmt.Printf("  • Testing section updated\n")
	}
}

func getPRBody(prNumber string) (string, error) {
	cmd := exec.Command("gh", "pr", "view", prNumber, "--json", "body")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	var data struct {
		Body string `json:"body"`
	}

	if err := utils.ParseJSON(output, &data); err != nil {
		return "", err
	}

	return data.Body, nil
}

func updatePRSection(body, section, content string) string {
	// Simple section replacement
	// This is a basic implementation - could be improved with better parsing
	var sectionHeader string
	switch section {
	case "Description":
		sectionHeader = "### Description"
	case "How to Test":
		sectionHeader = "### How to Test"
	default:
		sectionHeader = fmt.Sprintf("## %s", section)
	}

	// Find the section and replace its content
	lines := strings.Split(body, "\n")
	inSection := false
	var result []string
	sectionFound := false

	for _, line := range lines {
		if strings.HasPrefix(line, sectionHeader) {
			result = append(result, line)
			result = append(result, "")
			result = append(result, content)
			inSection = true
			sectionFound = true
			continue
		}

		if inSection && (strings.HasPrefix(line, "##") || strings.HasPrefix(line, "###")) {
			inSection = false
		}

		if !inSection {
			result = append(result, line)
		}
	}

	if !sectionFound {
		// Section not found, append it
		result = append(result, "")
		result = append(result, sectionHeader)
		result = append(result, "")
		result = append(result, content)
	}

	return strings.Join(result, "\n")
}
