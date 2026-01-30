package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/rithyhuot/vibe/internal/models"
	"github.com/rithyhuot/vibe/internal/services/github"
	"github.com/spf13/cobra"
)

// IssueUpdateCommandOptions holds flags for the issue-update command
type IssueUpdateCommandOptions struct {
	Title     string
	Body      string
	State     string
	Assignees []string
	Labels    []string
	Milestone string
	Projects  []string
}

// NewIssueUpdateCommand creates the issue-update command
func NewIssueUpdateCommand(ctx *CommandContext) *cobra.Command {
	opts := &IssueUpdateCommandOptions{}

	cmd := &cobra.Command{
		Use:   "issue-update <issue-number>",
		Short: "Update an existing GitHub issue",
		Long: `Update an existing GitHub issue's title, body, state, or metadata.

Note: When using CLI mode (gh CLI), assignees and labels are additive, not replaced.
For true replacement behavior, use API mode by setting github.mode: "api" in config.

Examples:
  vibe issue-update 123 --state closed                      # Close issue #123
  vibe issue-update 123 --state open                        # Reopen issue #123
  vibe issue-update 123 --title "New title"                 # Update title
  vibe issue-update 123 --body "Updated description"        # Update body
  vibe issue-update 123 --assignees user1,user2             # Assign users
  vibe issue-update 123 --labels bug,urgent                 # Add labels
  vibe issue-update 123 --milestone "v1.0"                  # Set milestone`,
		Args: cobra.ExactArgs(1),
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			// Get context from the command's context value (set by PreRunE)
			ctx = getCommandContext(cobraCmd, ctx)
			return runIssueUpdate(ctx, opts, args[0])
		},
	}

	cmd.Flags().StringVar(&opts.Title, "title", "", "Update issue title")
	cmd.Flags().StringVar(&opts.Body, "body", "", "Update issue body/description")
	cmd.Flags().StringVar(&opts.State, "state", "", "Update issue state (open or closed)")
	cmd.Flags().StringSliceVar(&opts.Assignees, "assignees", []string{}, "Update assignees (replaces existing)")
	cmd.Flags().StringSliceVar(&opts.Labels, "labels", []string{}, "Update labels (replaces existing)")
	cmd.Flags().StringVar(&opts.Milestone, "milestone", "", "Update milestone")
	cmd.Flags().StringSliceVar(&opts.Projects, "projects", []string{}, "Update projects (replaces existing)")

	return cmd
}

func runIssueUpdate(ctx *CommandContext, opts *IssueUpdateCommandOptions, issueNumberArg string) error {
	// Parse and validate issue number
	issueNumber, err := parseAndValidateIssueNumber(issueNumberArg)
	if err != nil {
		return err
	}

	// Validate state if provided
	if opts.State != "" {
		if opts.State != "open" && opts.State != "closed" {
			return fmt.Errorf("invalid state: %s (must be 'open' or 'closed')", opts.State)
		}
	}

	// Warn about CLI mode limitations
	if ctx.Config.GitHub.Mode == "cli" && (len(opts.Assignees) > 0 || len(opts.Labels) > 0) {
		yellow := color.New(color.FgYellow)
		_, _ = yellow.Println("⚠ Note: In CLI mode, assignees and labels are additive (not replaced).")
		_, _ = yellow.Println("  For true replacement, set github.mode to 'api' in config.")
		fmt.Println()
	}

	// Check if any updates were provided
	if !hasIssueUpdates(opts) {
		return fmt.Errorf("no updates provided. Use --title, --body, --state, --assignees, --labels, --milestone, or --projects flags")
	}

	// Build update request
	req := buildIssueUpdateRequest(opts)

	// Update issue with fallback to git remote repo
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Updating issue..."
	s.Start()

	issue, err := updateIssueWithFallback(ctx, issueNumber, req, s)
	if err != nil {
		s.Stop()
		return fmt.Errorf("failed to update issue: %w", err)
	}

	s.Stop()

	// Show success message
	displayIssueUpdateSuccess(issue, opts)

	return nil
}

func hasIssueUpdates(opts *IssueUpdateCommandOptions) bool {
	return opts.Title != "" ||
		opts.Body != "" ||
		opts.State != "" ||
		len(opts.Assignees) > 0 ||
		len(opts.Labels) > 0 ||
		opts.Milestone != "" ||
		len(opts.Projects) > 0
}

func buildIssueUpdateRequest(opts *IssueUpdateCommandOptions) *models.IssueUpdateRequest {
	req := &models.IssueUpdateRequest{}

	if opts.Title != "" {
		req.Title = &opts.Title
	}
	if opts.Body != "" {
		req.Body = &opts.Body
	}
	if opts.State != "" {
		req.State = &opts.State
	}
	if len(opts.Assignees) > 0 {
		req.Assignees = &opts.Assignees
	}
	if len(opts.Labels) > 0 {
		req.Labels = &opts.Labels
	}
	if opts.Milestone != "" {
		req.Milestone = &opts.Milestone
	}
	if len(opts.Projects) > 0 {
		req.ProjectIDs = &opts.Projects
	}

	return req
}

func displayIssueUpdateSuccess(issue *models.Issue, opts *IssueUpdateCommandOptions) {
	green := color.New(color.FgGreen, color.Bold)
	dim := color.New(color.Faint)

	fmt.Println()
	_, _ = green.Printf("✓ Updated Issue #%d\n", issue.Number)
	fmt.Println()
	_, _ = dim.Println("  " + issue.URL)
	fmt.Println()

	// Show what was updated
	if opts.Title != "" {
		fmt.Println("  • Title updated")
	}
	if opts.Body != "" {
		fmt.Println("  • Description updated")
	}
	if opts.State != "" {
		stateColor := green
		if opts.State == "closed" {
			stateColor = color.New(color.FgRed)
		}
		fmt.Printf("  • State: %s\n", stateColor.Sprint(strings.ToUpper(opts.State)))
	}
	if len(opts.Assignees) > 0 {
		fmt.Printf("  • Assignees: %s\n", strings.Join(opts.Assignees, ", "))
	}
	if len(opts.Labels) > 0 {
		fmt.Printf("  • Labels: %s\n", strings.Join(opts.Labels, ", "))
	}
	if opts.Milestone != "" {
		fmt.Printf("  • Milestone: %s\n", opts.Milestone)
	}
	if len(opts.Projects) > 0 {
		fmt.Printf("  • Projects: %s\n", strings.Join(opts.Projects, ", "))
	}

	fmt.Println()
}

// updateIssueWithFallback tries to update an issue with the configured client,
// and falls back to git remote repo if the configured repo is invalid
func updateIssueWithFallback(ctx *CommandContext, issueNumber int, req *models.IssueUpdateRequest, s *spinner.Spinner) (*models.Issue, error) {
	return withRepoFallback(ctx, s, func(client github.Client) (*models.Issue, error) {
		return client.UpdateIssue(context.Background(), issueNumber, req)
	})
}
