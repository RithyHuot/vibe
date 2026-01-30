package commands

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/rithyhuot/vibe/internal/models"
	"github.com/rithyhuot/vibe/internal/services/github"
	"github.com/spf13/cobra"
)

// IssueCreateCommandOptions holds flags for the issue-create command
type IssueCreateCommandOptions struct {
	Title     string
	Body      string
	BodyFile  string
	Assignees []string
	Labels    []string
	Milestone string
	Projects  []string
	Yes       bool
}

// NewIssueCreateCommand creates the issue-create command
func NewIssueCreateCommand(ctx *CommandContext) *cobra.Command {
	opts := &IssueCreateCommandOptions{}

	cmd := &cobra.Command{
		Use:   "issue-create",
		Short: "Create a new GitHub issue",
		Long: `Create a new GitHub issue with optional metadata.

Examples:
  vibe issue-create                                    # Interactive mode
  vibe issue-create --yes --title "Bug" --body "..."  # Non-interactive
  vibe issue-create --yes --title "Bug" --body-file bug.md --labels bug,urgent`,
		RunE: func(cobraCmd *cobra.Command, _ []string) error {
			// Get context from the command's context value (set by PreRunE)
			ctx = getCommandContext(cobraCmd, ctx)
			return runIssueCreate(ctx, opts)
		},
	}

	cmd.Flags().StringVar(&opts.Title, "title", "", "Issue title")
	cmd.Flags().StringVar(&opts.Body, "body", "", "Issue body/description")
	cmd.Flags().StringVar(&opts.BodyFile, "body-file", "", "Read issue body from file")
	cmd.Flags().StringSliceVar(&opts.Assignees, "assignees", []string{}, "Assignees (comma-separated)")
	cmd.Flags().StringSliceVar(&opts.Labels, "labels", []string{}, "Labels (comma-separated)")
	cmd.Flags().StringVar(&opts.Milestone, "milestone", "", "Milestone")
	cmd.Flags().StringSliceVar(&opts.Projects, "projects", []string{}, "Projects (comma-separated)")
	cmd.Flags().BoolVarP(&opts.Yes, "yes", "y", false, "Skip confirmation prompts")

	return cmd
}

func runIssueCreate(ctx *CommandContext, opts *IssueCreateCommandOptions) error {
	// Determine if this is interactive or non-interactive mode
	if opts.Yes || hasIssueNonInteractiveOptions(opts) {
		return createIssueNonInteractive(ctx, opts)
	}

	return createIssueInteractive(ctx, opts)
}

func hasIssueNonInteractiveOptions(opts *IssueCreateCommandOptions) bool {
	return opts.Title != "" || opts.Body != "" || opts.BodyFile != ""
}

func createIssueInteractive(ctx *CommandContext, _ *IssueCreateCommandOptions) error {
	bold := color.New(color.Bold)

	fmt.Println()
	_, _ = bold.Println("Create New Issue")
	fmt.Println()

	// Try to get template
	template, err := ctx.GitHubClient.GetIssueTemplate(context.Background())
	if err != nil {
		dim := color.New(color.Faint)
		_, _ = dim.Println("Note: Could not load issue template")
	}

	// Prompt for title
	var title string
	titlePrompt := &survey.Input{
		Message: "Issue title:",
	}
	if err := survey.AskOne(titlePrompt, &title, survey.WithValidator(survey.Required)); err != nil {
		return err
	}

	// Prompt for body
	var body string
	bodyPrompt := &survey.Multiline{
		Message: "Issue description (press Ctrl+D or Ctrl+Z when done):",
		Default: template,
	}
	if err := survey.AskOne(bodyPrompt, &body); err != nil {
		return err
	}

	// Prompt for assignees
	var assigneesInput string
	assigneePrompt := &survey.Input{
		Message: "Assignees (comma-separated, e.g. user1,user2):",
	}
	if err := survey.AskOne(assigneePrompt, &assigneesInput); err != nil {
		return err
	}
	assignees := parseCommaSeparated(assigneesInput)

	// Prompt for labels
	var labelsInput string
	labelPrompt := &survey.Input{
		Message: "Labels (comma-separated, e.g. bug,urgent):",
	}
	if err := survey.AskOne(labelPrompt, &labelsInput); err != nil {
		return err
	}
	labels := parseCommaSeparated(labelsInput)

	// Prompt for milestone
	var milestone string
	milestonePrompt := &survey.Input{
		Message: "Milestone (optional):",
	}
	if err := survey.AskOne(milestonePrompt, &milestone); err != nil {
		return err
	}

	// Prompt for projects
	var projectsInput string
	projectPrompt := &survey.Input{
		Message: "Projects (comma-separated, optional):",
	}
	if err := survey.AskOne(projectPrompt, &projectsInput); err != nil {
		return err
	}
	projects := parseCommaSeparated(projectsInput)

	// Show preview
	displayIssueCreatePreview(title, body, assignees, labels, milestone, projects)

	// Confirm
	var shouldCreate bool
	confirmPrompt := &survey.Confirm{
		Message: "Create this issue?",
		Default: true,
	}
	if err := survey.AskOne(confirmPrompt, &shouldCreate); err != nil {
		return err
	}

	if !shouldCreate {
		yellow := color.New(color.FgYellow)
		_, _ = yellow.Println("Cancelled.")
		return nil
	}

	// Create issue
	req := &models.IssueCreateRequest{
		Title:      title,
		Body:       body,
		Assignees:  assignees,
		Labels:     labels,
		Milestone:  milestone,
		ProjectIDs: projects,
	}

	return createIssue(ctx, req)
}

func createIssueNonInteractive(ctx *CommandContext, opts *IssueCreateCommandOptions) error {
	// Validate required fields
	if opts.Title == "" {
		return fmt.Errorf("--title is required in non-interactive mode")
	}

	// Get body from file or flag
	body := opts.Body
	if opts.BodyFile != "" {
		content, err := os.ReadFile(opts.BodyFile)
		if err != nil {
			return fmt.Errorf("failed to read body file: %w", err)
		}
		body = string(content)
	}

	// If no body provided, try to get template
	if body == "" {
		template, err := ctx.GitHubClient.GetIssueTemplate(context.Background())
		if err != nil {
			dim := color.New(color.Faint)
			_, _ = dim.Println("Note: Could not load issue template")
		} else {
			body = template
		}
	}

	// Show preview if not --yes
	if !opts.Yes {
		displayIssueCreatePreview(opts.Title, body, opts.Assignees, opts.Labels, opts.Milestone, opts.Projects)

		// Confirm
		var shouldCreate bool
		prompt := &survey.Confirm{
			Message: "Create this issue?",
			Default: true,
		}
		if err := survey.AskOne(prompt, &shouldCreate); err != nil {
			return err
		}

		if !shouldCreate {
			yellow := color.New(color.FgYellow)
			_, _ = yellow.Println("Cancelled.")
			return nil
		}
	}

	// Create issue
	req := &models.IssueCreateRequest{
		Title:      opts.Title,
		Body:       body,
		Assignees:  opts.Assignees,
		Labels:     opts.Labels,
		Milestone:  opts.Milestone,
		ProjectIDs: opts.Projects,
	}

	return createIssue(ctx, req)
}

func createIssue(ctx *CommandContext, req *models.IssueCreateRequest) error {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Creating issue..."
	s.Start()

	issue, err := createIssueWithFallback(ctx, req, s)
	if err != nil {
		s.Stop()
		return fmt.Errorf("failed to create issue: %w", err)
	}

	s.Stop()

	// Show success message
	green := color.New(color.FgGreen, color.Bold)
	dim := color.New(color.Faint)

	fmt.Println()
	_, _ = green.Println("✓ Issue created successfully!")
	fmt.Println()
	fmt.Printf("  Issue #%d: %s\n", issue.Number, issue.Title)
	_, _ = dim.Println("  " + issue.URL)
	fmt.Println()

	return nil
}

// parseCommaSeparated splits a comma-separated string into a slice
func parseCommaSeparated(input string) []string {
	if input == "" {
		return []string{}
	}

	parts := strings.Split(input, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// indentText indents each line of text with the given prefix
func indentText(text, prefix string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = prefix + line
	}
	return strings.Join(lines, "\n")
}

// displayIssueCreatePreview shows a formatted preview of the issue to be created
func displayIssueCreatePreview(title, body string, assignees, labels []string, milestone string, projects []string) {
	bold := color.New(color.Bold)
	cyan := color.New(color.FgCyan)

	fmt.Println()
	_, _ = bold.Println("Preview:")
	fmt.Println()
	fmt.Printf("  Title: %s\n", cyan.Sprint(title))
	if len(assignees) > 0 {
		fmt.Printf("  Assignees: %s\n", strings.Join(assignees, ", "))
	}
	if len(labels) > 0 {
		fmt.Printf("  Labels: %s\n", strings.Join(labels, ", "))
	}
	if milestone != "" {
		fmt.Printf("  Milestone: %s\n", milestone)
	}
	if len(projects) > 0 {
		fmt.Printf("  Projects: %s\n", strings.Join(projects, ", "))
	}
	fmt.Println()

	if body != "" {
		fmt.Println("  Body:")
		fmt.Println(indentText(body, "    "))
		fmt.Println()
	}
}

// createIssueWithFallback tries to create an issue with the configured client,
// and falls back to git remote repo if the configured repo is invalid
func createIssueWithFallback(ctx *CommandContext, req *models.IssueCreateRequest, s *spinner.Spinner) (*models.Issue, error) {
	return withRepoFallback(ctx, s, func(client github.Client) (*models.Issue, error) {
		return client.CreateIssue(context.Background(), req)
	})
}
