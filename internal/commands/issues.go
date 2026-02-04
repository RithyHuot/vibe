package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	survey "github.com/AlecAivazis/survey/v2"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/rithyhuot/vibe/internal/models"
	"github.com/rithyhuot/vibe/internal/services/github"
)

const (
	// Display constants for issue list formatting
	maxTitleDisplayLength  = 50 // Fits standard 110-char terminal width with other columns
	maxLabelsToShow        = 2  // Balance between information density and readability
	maxAssigneesToShow     = 2  // Balance between information density and readability
	issueSelectionMaxTitle = 60 // Prevents selection prompt from wrapping on standard terminals
)

// IssuesCommandOptions holds flags for the issues command
type IssuesCommandOptions struct {
	State  string
	Limit  int
	Select bool
}

// NewIssuesCommand creates the issues command
func NewIssuesCommand(ctx *CommandContext) *cobra.Command {
	opts := &IssuesCommandOptions{}

	cmd := &cobra.Command{
		Use:   "issues",
		Short: "List GitHub issues",
		Long: `List GitHub issues with optional filtering by state.

Examples:
  vibe issues                    # List open issues
  vibe issues --state closed     # List closed issues
  vibe issues --state all        # List all issues
  vibe issues --select           # List issues and select one to view details`,
		RunE: func(cobraCmd *cobra.Command, _ []string) error {
			// Get context from the command's context value (set by PreRunE)
			ctx = getCommandContext(cobraCmd, ctx)
			return runIssues(ctx, opts)
		},
	}

	cmd.Flags().StringVar(&opts.State, "state", "open", "Filter by state (open, closed, all)")
	cmd.Flags().IntVar(&opts.Limit, "limit", 30, "Maximum number of issues to display")
	cmd.Flags().BoolVarP(&opts.Select, "select", "s", false, "Enable interactive selection to view details")

	return cmd
}

func runIssues(ctx *CommandContext, opts *IssuesCommandOptions) error {
	// Validate state
	validStates := map[string]bool{"open": true, "closed": true, "all": true}
	if !validStates[opts.State] {
		return fmt.Errorf("invalid state: %s (must be one of: open, closed, all)", opts.State)
	}

	// Validate limit
	if opts.Limit <= 0 {
		return fmt.Errorf("limit must be positive, got %d", opts.Limit)
	}

	// Fetch issues with fallback to git remote repo
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Fetching issues..."
	s.Start()

	issues, err := fetchIssuesWithFallback(ctx, opts.State, s)
	if err != nil {
		s.Stop()
		return fmt.Errorf("failed to fetch issues: %w", err)
	}

	s.Stop()

	// Limit issues
	if len(issues) > opts.Limit {
		issues = issues[:opts.Limit]
	}

	if len(issues) == 0 {
		yellow := color.New(color.FgYellow)
		_, _ = yellow.Printf("No %s issues found.\n", opts.State)
		return nil
	}

	// Display issues
	displayIssuesTable(issues, opts.State)

	// Interactive selection mode
	if opts.Select {
		return handleIssueSelection(ctx, issues)
	}

	return nil
}

// displayIssuesTable displays issues in a table format
func displayIssuesTable(issues []*models.Issue, state string) {
	bold := color.New(color.Bold)
	dim := color.New(color.Faint)
	green := color.New(color.FgGreen)
	red := color.New(color.FgRed)
	cyan := color.New(color.FgCyan)

	// Capitalize state for display
	var stateDisplay string
	if state != "all" {
		stateDisplay = strings.ToUpper(state[:1]) + state[1:]
	} else {
		stateDisplay = "All"
	}

	fmt.Println()
	_, _ = bold.Printf("%s Issues (%d)\n", stateDisplay, len(issues))
	fmt.Println()

	// Print header
	fmt.Printf("%-8s %-50s %-10s %-20s %-20s\n", "NUMBER", "TITLE", "STATE", "LABELS", "ASSIGNEES")
	_, _ = dim.Println(strings.Repeat("-", 110))

	// Print issues
	for _, issue := range issues {
		// Truncate title if too long (use runes to handle UTF-8 properly)
		title := issue.Title
		titleRunes := []rune(title)
		if len(titleRunes) > maxTitleDisplayLength-3 {
			title = string(titleRunes[:maxTitleDisplayLength-3]) + "..."
		}

		// Format state
		stateStr := ""
		if issue.State == "open" {
			stateStr = green.Sprint("OPEN")
		} else {
			stateStr = red.Sprint("CLOSED")
		}

		// Format labels
		labelStr := ""
		if len(issue.Labels) > 0 {
			labels := make([]string, 0, maxLabelsToShow)
			for i, label := range issue.Labels {
				if i >= maxLabelsToShow {
					labels = append(labels, "...")
					break
				}
				labels = append(labels, label.Name)
			}
			labelStr = cyan.Sprint(strings.Join(labels, ","))
		} else {
			labelStr = dim.Sprint("-")
		}

		// Format assignees
		assigneeStr := ""
		if len(issue.Assignees) > 0 {
			assignees := make([]string, 0, maxAssigneesToShow)
			for i, assignee := range issue.Assignees {
				if i >= maxAssigneesToShow {
					assignees = append(assignees, "...")
					break
				}
				assignees = append(assignees, "@"+assignee.Login)
			}
			assigneeStr = strings.Join(assignees, ",")
		} else {
			assigneeStr = dim.Sprint("-")
		}

		fmt.Printf("#%-7d %-50s %-10s %-20s %-20s\n", issue.Number, title, stateStr, labelStr, assigneeStr)
	}

	fmt.Println()
}

// handleIssueSelection handles interactive issue selection
func handleIssueSelection(ctx *CommandContext, issues []*models.Issue) error {
	// Create selection options
	options := make([]string, len(issues))
	issueMap := make(map[string]*models.Issue)

	for i, issue := range issues {
		// Format: "#123 - Title [labels] (@assignee)"
		labelsPart := ""
		if len(issue.Labels) > 0 {
			labelNames := make([]string, 0, maxLabelsToShow)
			for j, label := range issue.Labels {
				if j >= maxLabelsToShow {
					break
				}
				labelNames = append(labelNames, label.Name)
			}
			labelsPart = fmt.Sprintf(" [%s]", strings.Join(labelNames, ", "))
		}

		assigneePart := ""
		if len(issue.Assignees) > 0 {
			assigneePart = fmt.Sprintf(" (@%s)", issue.Assignees[0].Login)
			if len(issue.Assignees) > 1 {
				assigneePart = fmt.Sprintf(" (@%s +%d)", issue.Assignees[0].Login, len(issue.Assignees)-1)
			}
		}

		// Truncate title if too long (use runes to handle UTF-8 properly)
		title := issue.Title
		titleRunes := []rune(title)
		if len(titleRunes) > issueSelectionMaxTitle {
			title = string(titleRunes[:issueSelectionMaxTitle]) + "..."
		}

		option := fmt.Sprintf("#%d - %s%s%s", issue.Number, title, labelsPart, assigneePart)
		options[i] = option
		issueMap[option] = issue
	}

	// Prompt for selection
	var selected string
	prompt := &survey.Select{
		Message: "Select an issue to view:",
		Options: options,
	}

	if err := survey.AskOne(prompt, &selected); err != nil {
		return err
	}

	// Get selected issue
	selectedIssue := issueMap[selected]
	if selectedIssue == nil {
		return fmt.Errorf("failed to find selected issue")
	}

	// Ask if user wants to include comments
	var includeComments bool
	commentPrompt := &survey.Confirm{
		Message: "Include comments?",
		Default: false,
	}
	if err := survey.AskOne(commentPrompt, &includeComments); err != nil {
		return err
	}

	// Fetch full issue details with optional comments using fallback
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Fetching issue details..."
	s.Start()

	issue, err := fetchIssueWithFallback(ctx, selectedIssue.Number, includeComments, s)
	if err != nil {
		s.Stop()
		return fmt.Errorf("failed to fetch issue details: %w", err)
	}

	s.Stop()

	// Display full issue details
	displayIssue(issue)

	// Offer to create a branch
	return offerCreateBranchForIssue(ctx, issue)
}

// fetchIssuesWithFallback tries to fetch issues with the configured client,
// and falls back to git remote repo if the configured repo is invalid
func fetchIssuesWithFallback(ctx *CommandContext, state string, s *spinner.Spinner) ([]*models.Issue, error) {
	return withRepoFallback(ctx, s, func(client github.Client) ([]*models.Issue, error) {
		return client.ListIssues(context.Background(), state)
	})
}
