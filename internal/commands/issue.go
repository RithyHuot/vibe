package commands

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	survey "github.com/AlecAivazis/survey/v2"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/rithyhuot/vibe/internal/models"
	"github.com/rithyhuot/vibe/internal/services/github"
	"github.com/rithyhuot/vibe/internal/utils"
)

// IssueCommandOptions holds flags for the issue command
type IssueCommandOptions struct {
	Comments bool
}

// NewIssueCommand creates the issue command
func NewIssueCommand(ctx *CommandContext) *cobra.Command {
	opts := &IssueCommandOptions{}

	cmd := &cobra.Command{
		Use:   "issue [issue-number]",
		Short: "View issue details",
		Long: `View GitHub issue details including title, description, assignees, labels, milestone, state, and optionally comments.

If no issue number is provided, attempts to extract it from the current branch name.
Supports branch patterns like: issue-123, 123-description, username/issue-123/..., fix-issue-456

Examples:
  vibe issue 123                 # View issue #123
  vibe issue                     # View issue from current branch name
  vibe issue 456 --comments      # View issue #456 with comments
  vibe issue -c                  # View current issue with comments`,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			// Get context from the command's context value (set by PreRunE)
			ctx = getCommandContext(cobraCmd, ctx)
			issueNumber := ""
			if len(args) > 0 {
				issueNumber = args[0]
			}
			return runIssue(ctx, issueNumber, opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.Comments, "comments", "c", false, "Include comments")

	return cmd
}

func runIssue(ctx *CommandContext, issueNumberArg string, opts *IssueCommandOptions) error {
	// Resolve issue number
	issueNumber, err := resolveIssueNumber(ctx, issueNumberArg)
	if err != nil {
		return err
	}

	// Ask if user wants to include comments (if not explicitly set via flag)
	includeComments := opts.Comments
	if !opts.Comments {
		var wantComments bool
		commentPrompt := &survey.Confirm{
			Message: "Include comments?",
			Default: false,
		}
		if err := survey.AskOne(commentPrompt, &wantComments); err != nil {
			return err
		}
		includeComments = wantComments
	}

	// Fetch issue with fallback to git remote repo
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Fetching issue..."
	s.Start()

	issue, err := fetchIssueWithFallback(ctx, issueNumber, includeComments, s)
	if err != nil {
		s.Stop()
		return fmt.Errorf("failed to fetch issue: %w", err)
	}

	s.Stop()

	// Display issue
	displayIssue(issue)

	// Offer to create a branch
	return offerCreateBranchForIssue(ctx, issue)
}

// resolveIssueNumber resolves the issue number from argument or branch name
func resolveIssueNumber(ctx *CommandContext, issueNumberArg string) (int, error) {
	// If provided explicitly, validate it
	if issueNumberArg != "" {
		return parseAndValidateIssueNumber(issueNumberArg)
	}

	// Try to extract from branch name
	currentBranch, err := ctx.GitRepo.CurrentBranch()
	if err != nil {
		return 0, fmt.Errorf("failed to get current branch: %w", err)
	}

	issueNum := extractIssueNumberFromBranch(currentBranch)
	if issueNum == 0 {
		yellow := color.New(color.FgYellow)
		_, _ = yellow.Println("Could not determine issue number from branch name.")
		return 0, fmt.Errorf("please provide an issue number: vibe issue <number>")
	}

	return issueNum, nil
}

// extractIssueNumberFromBranch extracts issue number from branch name
// Supports patterns: issue-123, 123-description, username/issue-123/..., fix-issue-456
func extractIssueNumberFromBranch(branch string) int {
	// Try various patterns
	patterns := []string{
		`^issue-(\d+)`,   // issue-123
		`^(\d+)-`,        // 123-description
		`/issue-(\d+)`,   // username/issue-123/...
		`-issue-(\d+)`,   // fix-issue-456
		`issue[_-](\d+)`, // issue_123 or issue-123 anywhere
		`^.*?/(\d+)-`,    // username/123-description
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(branch); len(matches) > 1 {
			if num, err := strconv.Atoi(matches[1]); err == nil {
				return num
			}
		}
	}

	return 0
}

// displayIssue displays issue details with formatting
//
//nolint:gocyclo // Complexity is acceptable for comprehensive issue display
func displayIssue(issue *models.Issue) {
	bold := color.New(color.Bold)
	dim := color.New(color.Faint)
	green := color.New(color.FgGreen)
	red := color.New(color.FgRed)
	yellow := color.New(color.FgYellow)
	cyan := color.New(color.FgCyan)

	fmt.Println()
	_, _ = bold.Printf("Issue #%d: %s\n", issue.Number, issue.Title)
	_, _ = dim.Println(issue.URL)
	fmt.Println()

	// Display state
	fmt.Print("State: ")
	if issue.State == "open" {
		_, _ = green.Println("OPEN")
	} else {
		_, _ = red.Println("CLOSED")
	}

	// Display author
	fmt.Printf("Author: @%s\n", issue.User.Login)

	// Display assignees
	if len(issue.Assignees) > 0 {
		assignees := make([]string, len(issue.Assignees))
		for i, a := range issue.Assignees {
			assignees[i] = "@" + a.Login
		}
		fmt.Printf("Assignees: %s\n", strings.Join(assignees, ", "))
	} else {
		_, _ = dim.Println("Assignees: None")
	}

	// Display labels
	if len(issue.Labels) > 0 {
		fmt.Print("Labels: ")
		for i, label := range issue.Labels {
			if i > 0 {
				fmt.Print(", ")
			}
			_, _ = cyan.Printf("[%s]", label.Name)
		}
		fmt.Println()
	} else {
		_, _ = dim.Println("Labels: None")
	}

	// Display milestone
	if issue.Milestone != nil {
		fmt.Printf("Milestone: %s", issue.Milestone.Title)
		if issue.Milestone.State == "open" {
			_, _ = green.Printf(" [OPEN]\n")
		} else {
			_, _ = dim.Printf(" [CLOSED]\n")
		}
	} else {
		_, _ = dim.Println("Milestone: None")
	}

	// Display projects
	if len(issue.Projects) > 0 {
		projects := make([]string, len(issue.Projects))
		for i, p := range issue.Projects {
			projects[i] = p.Title
		}
		fmt.Printf("Projects: %s\n", strings.Join(projects, ", "))
	} else {
		_, _ = dim.Println("Projects: None")
	}

	// Display timestamps
	fmt.Println()
	_, _ = dim.Printf("Created: %s\n", issue.CreatedAt.Format("2006-01-02 15:04:05"))
	_, _ = dim.Printf("Updated: %s\n", issue.UpdatedAt.Format("2006-01-02 15:04:05"))
	if issue.ClosedAt != nil {
		_, _ = dim.Printf("Closed: %s\n", issue.ClosedAt.Format("2006-01-02 15:04:05"))
	}

	// Display body
	if issue.Body != "" {
		fmt.Println()
		_, _ = bold.Println("Description:")
		fmt.Println(issue.Body)
	}

	// Display comments
	if len(issue.Comments) > 0 {
		fmt.Println()
		_, _ = bold.Printf("Comments (%d):\n", len(issue.Comments))
		fmt.Println()

		for i, comment := range issue.Comments {
			_, _ = yellow.Printf("Comment #%d by @%s\n", i+1, comment.User.Login)
			_, _ = dim.Printf("%s\n", comment.CreatedAt.Format("2006-01-02 15:04:05"))
			fmt.Println(comment.Body)
			if i < len(issue.Comments)-1 {
				fmt.Println()
				_, _ = dim.Println("---")
				fmt.Println()
			}
		}
	}

	fmt.Println()
}

// offerCreateBranchForIssue prompts the user to create a branch for the issue
func offerCreateBranchForIssue(ctx *CommandContext, issue *models.Issue) error {
	// Ask if user wants to create a branch
	var shouldCreateBranch bool
	prompt := &survey.Confirm{
		Message: "Would you like to create a branch for this issue?",
		Default: false,
	}
	if err := survey.AskOne(prompt, &shouldCreateBranch); err != nil {
		return err
	}

	if !shouldCreateBranch {
		return nil
	}

	// Resolve branch prefix with username fallback
	branchPrefix, gitUsername, err := utils.ResolveBranchPrefix(ctx.Config.Git.BranchPrefix)
	if err != nil {
		return err
	}

	// Generate branch name from issue
	issueID := fmt.Sprintf("issue-%d", issue.Number)
	var branchName string
	if gitUsername != "" {
		branchName = utils.GenerateBranchName(branchPrefix, issueID, issue.Title, gitUsername)
	} else {
		branchName = utils.GenerateBranchName(branchPrefix, issueID, issue.Title)
	}

	// Validate branch name
	if err := utils.ValidateBranchName(branchName); err != nil {
		return fmt.Errorf("invalid branch name '%s': %w", branchName, err)
	}

	// Check if branch already exists
	exists, err := ctx.GitRepo.BranchExists(branchName)
	if err != nil {
		return fmt.Errorf("failed to check if branch exists: %w", err)
	}

	// Check for uncommitted changes before checkout
	if err := handleUncommittedChanges(ctx); err != nil {
		return err
	}

	if exists {
		yellow := color.New(color.FgYellow)
		_, _ = yellow.Printf("\n⚠ Branch '%s' already exists\n", branchName)

		// Ask if they want to checkout the existing branch
		var shouldCheckout bool
		checkoutPrompt := &survey.Confirm{
			Message: "Would you like to check out the existing branch?",
			Default: true,
		}
		if err := survey.AskOne(checkoutPrompt, &shouldCheckout); err != nil {
			return err
		}

		if !shouldCheckout {
			return nil
		}

		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Suffix = " Checking out existing branch..."
		s.Start()

		err = ctx.GitRepo.Checkout(branchName)
		s.Stop()
		if err != nil {
			return fmt.Errorf("failed to checkout branch: %w", err)
		}

		green := color.New(color.FgGreen)
		_, _ = green.Printf("✓ Checked out branch: %s\n", branchName)
	} else {
		// Create new branch
		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Suffix = " Creating branch..."
		s.Start()

		err = ctx.GitRepo.CreateBranch(branchName)
		if err != nil {
			s.Stop()
			return fmt.Errorf("failed to create branch: %w", err)
		}

		// Checkout the new branch
		err = ctx.GitRepo.Checkout(branchName)
		s.Stop()
		if err != nil {
			return fmt.Errorf("failed to checkout branch: %w", err)
		}

		green := color.New(color.FgGreen, color.Bold)
		cyan := color.New(color.FgCyan)
		fmt.Println()
		_, _ = green.Printf("✓ Created and checked out branch\n")
		_, _ = cyan.Printf("  %s\n", branchName)
		fmt.Println()
		_, _ = cyan.Println("Ready to start working! 🚀")
	}

	return nil
}

// fetchIssueWithFallback tries to fetch an issue with the configured client,
// and falls back to git remote repo if the configured repo is invalid
func fetchIssueWithFallback(ctx *CommandContext, issueNumber int, includeComments bool, s *spinner.Spinner) (*models.Issue, error) {
	return withRepoFallback(ctx, s, func(client github.Client) (*models.Issue, error) {
		return client.GetIssue(context.Background(), issueNumber, includeComments)
	})
}
