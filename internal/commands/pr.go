package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/rithyhuot/vibe/internal/models"
	"github.com/rithyhuot/vibe/internal/utils"
	"github.com/spf13/cobra"
)

// PRCommandOptions holds flags for the PR command
type PRCommandOptions struct {
	Draft       bool
	Title       string
	Summary     string
	Description string
	Testing     string
	Base        string
	BodyFile    string
	Yes         bool
	AI          bool
}

// NewPRCommand creates the pr command
func NewPRCommand(ctx *CommandContext) *cobra.Command {
	opts := &PRCommandOptions{}

	cmd := &cobra.Command{
		Use:   "pr",
		Short: "Create a pull request",
		Long: `Creates a pull request for the current branch with optional AI-generated description.

For long PR descriptions, use --body-file to read from a file:
  vibe pr --yes --title "My PR" --body-file pr_body.md

Or pass individual sections:
  vibe pr --yes --title "My PR" --summary "..." --description "..." --testing "..."`,
		RunE: func(cobraCmd *cobra.Command, _ []string) error {
			// Get context from the command's context value (set by PreRunE)
			ctx = getCommandContext(cobraCmd, ctx)
			return runPR(ctx, opts)
		},
	}

	cmd.Flags().BoolVar(&opts.Draft, "draft", false, "Create as draft PR")
	cmd.Flags().StringVar(&opts.Title, "title", "", "PR title")
	cmd.Flags().StringVar(&opts.Summary, "summary", "", "PR summary")
	cmd.Flags().StringVar(&opts.Description, "description", "", "PR description")
	cmd.Flags().StringVar(&opts.Testing, "testing", "", "How to test")
	cmd.Flags().StringVar(&opts.Base, "base", "", "Base branch (default: from config)")
	cmd.Flags().StringVar(&opts.BodyFile, "body-file", "", "Read PR body from file")
	cmd.Flags().BoolVarP(&opts.Yes, "yes", "y", false, "Skip confirmation prompts")
	cmd.Flags().BoolVar(&opts.AI, "ai", false, "Use AI to generate PR description from git diff")

	return cmd
}

func runPR(ctx *CommandContext, opts *PRCommandOptions) error {
	// Safety check: verify we're in a git repository
	currentBranch, err := ctx.GitRepo.CurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	// Safety check: never create PR from main or master
	if currentBranch == "main" || currentBranch == "master" {
		return fmt.Errorf("safety check: cannot create PR from %s branch. Create a feature branch first", currentBranch)
	}

	// Check if gh CLI is available
	if !isGHAvailable() {
		return fmt.Errorf("GitHub CLI (gh) is not available or not authenticated. Run: gh auth login")
	}

	// Check for existing PR
	existingPR, _ := getPRForCurrentBranch()
	if existingPR != nil {
		yellow := color.New(color.FgYellow)
		_, _ = yellow.Printf("PR already exists: %s\n", existingPR.URL)

		if hasNonInteractiveOptions(opts) {
			// In non-interactive mode, suggest using pr-update
			return fmt.Errorf("PR already exists. Use 'vibe pr-update' to modify it")
		}

		// Interactive: ask what to do
		var action string
		prompt := &survey.Select{
			Message: "What would you like to do?",
			Options: []string{"View status", "Open in browser", "Cancel"},
		}
		if err := survey.AskOne(prompt, &action); err != nil {
			return err
		}

		switch action {
		case "View status":
			return runPRStatus(ctx, fmt.Sprintf("%d", existingPR.Number))
		case "Open in browser":
			cmd := exec.Command("open", existingPR.URL)
			return cmd.Run()
		}
		return nil
	}

	// Determine if this is interactive or non-interactive mode
	// If --yes flag is set, always use non-interactive mode
	if opts.Yes || hasNonInteractiveOptions(opts) {
		return createPRNonInteractive(ctx, opts, currentBranch)
	}

	return createPRInteractive(ctx, opts, currentBranch)
}

func hasNonInteractiveOptions(opts *PRCommandOptions) bool {
	return opts.Yes || opts.BodyFile != "" || opts.Title != "" || opts.Summary != "" || opts.Description != "" || opts.Testing != ""
}

//nolint:gocyclo // Complex user interaction flow
func createPRInteractive(ctx *CommandContext, opts *PRCommandOptions, branch string) error {
	// Determine base branch
	baseBranch := opts.Base
	if baseBranch == "" {
		exists, _ := ctx.GitRepo.BranchExists(ctx.Config.Git.BaseBranch)
		if exists {
			baseBranch = ctx.Config.Git.BaseBranch
		} else {
			baseBranch = "main" // fallback
		}
	}

	// Show push summary
	fmt.Println()
	bold := color.New(color.Bold)
	cyan := color.New(color.FgCyan)
	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)
	dim := color.New(color.Faint)

	_, _ = bold.Println("📦 Push Summary")
	fmt.Println()
	fmt.Printf("  Branch: %s\n", cyan.Sprint(branch))
	fmt.Printf("  Base:   %s\n", dim.Sprint(baseBranch))

	// Get commit and file counts
	commits, _ := getCommitCount(baseBranch, branch)
	files, _ := getChangedFilesCount(baseBranch, branch)

	commitText := "commit"
	if commits != 1 {
		commitText = "commits"
	}
	fileText := "file"
	if files != 1 {
		fileText = "files"
	}

	fmt.Printf("  %s %s, %s %s changed\n", green.Sprint(commits), commitText, yellow.Sprint(files), fileText)
	fmt.Println()

	// Confirm push
	if !opts.Yes {
		var shouldPush bool
		prompt := &survey.Confirm{
			Message: "Push to GitHub and create PR?",
			Default: true,
		}
		if err := survey.AskOne(prompt, &shouldPush); err != nil {
			return err
		}

		if !shouldPush {
			yellow := color.New(color.FgYellow)
			_, _ = yellow.Println("Cancelled.")
			return nil
		}
	}

	// Push branch
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Pushing branch..."
	s.Start()

	if err := pushBranch(); err != nil {
		s.Stop()
		return fmt.Errorf("failed to push branch: %w", err)
	}
	s.Stop()
	_, _ = green.Println("✓ Branch pushed")

	// Extract ticket ID and fetch details
	ticketID, _ := utils.ExtractTicketID(branch)
	var ticketName string

	if ticketID != "" {
		s2 := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s2.Suffix = " Fetching ticket details..."
		s2.Start()

		cmdCtx := context.Background()
		task, err := ctx.ClickUpClient.GetTask(cmdCtx, ticketID)
		if err == nil && task != nil {
			ticketName = task.Name
			s2.Stop()
			_, _ = green.Printf("✓ Ticket: %s\n", ticketName)
		} else {
			s2.Stop()
			_, _ = dim.Println("⚠ Could not fetch ticket details")
		}
	}

	// Read PR template
	repoRoot, _ := ctx.GitRepo.GetRootPath()
	templatePath := filepath.Join(repoRoot, ".github", "PULL_REQUEST_TEMPLATE.md")
	var template string
	if data, err := os.ReadFile(templatePath); err == nil {
		template = string(data)
		_, _ = dim.Println("Found PR template")
	}

	// AI-generated description if requested
	var aiGeneratedDesc string
	useAI := opts.AI
	if !useAI && ctx.Config.AI.Enabled && ctx.ClaudeClient != nil {
		// Ask if user wants to use AI
		var wantAI bool
		aiPrompt := &survey.Confirm{
			Message: "Use AI to generate PR description?",
			Default: false,
		}
		if err := survey.AskOne(aiPrompt, &wantAI); err == nil {
			useAI = wantAI
		}
	}

	if useAI && ctx.ClaudeClient != nil {
		// Get git diff
		cmdCtx := context.Background()
		diffCmd := exec.Command("git", "diff", fmt.Sprintf("%s...%s", baseBranch, branch))
		diffOutput, err := diffCmd.Output()
		if err == nil && len(diffOutput) > 0 {
			s3 := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
			s3.Suffix = " Generating AI description..."
			s3.Start()

			// Generate description
			prTitle := ticketName
			if prTitle == "" {
				prTitle = branch
			}
			aiGeneratedDesc, err = ctx.ClaudeClient.EnhancePRDescription(cmdCtx, prTitle, string(diffOutput))
			s3.Stop()

			if err == nil && aiGeneratedDesc != "" {
				_, _ = green.Println("✓ AI description generated")
				fmt.Println()
				_, _ = bold.Println("Generated Description:")
				fmt.Println(aiGeneratedDesc)
				fmt.Println()
			} else {
				_, _ = yellow.Printf("⚠ AI generation failed: %v\n", err)
			}
		}
	}

	// Prompt for PR content
	var prTitle string
	titlePrompt := &survey.Input{
		Message: "PR title:",
		Default: ticketName,
	}
	if err := survey.AskOne(titlePrompt, &prTitle); err != nil {
		return err
	}

	var summary string
	summaryDefault := ""
	if aiGeneratedDesc != "" {
		// Extract first paragraph as summary
		lines := strings.Split(aiGeneratedDesc, "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" && !strings.HasPrefix(line, "#") {
				summaryDefault = strings.TrimSpace(line)
				break
			}
		}
	}
	summaryPrompt := &survey.Input{
		Message: "Summary (what and why):",
		Default: summaryDefault,
	}
	if err := survey.AskOne(summaryPrompt, &summary); err != nil {
		return err
	}

	var description string
	descDefault := ""
	if aiGeneratedDesc != "" {
		descDefault = aiGeneratedDesc
	}
	descPrompt := &survey.Input{
		Message: "Description (details):",
		Default: descDefault,
	}
	if err := survey.AskOne(descPrompt, &description); err != nil {
		return err
	}

	var testing string
	testingPrompt := &survey.Input{
		Message: "How to test:",
	}
	if err := survey.AskOne(testingPrompt, &testing); err != nil {
		return err
	}

	// Build PR body
	prBody := buildPRBody(template, ticketID, summary, description, testing)

	// Preview
	fmt.Println()
	_, _ = bold.Println("━━━ PR Preview ━━━")
	fmt.Println()
	_, _ = yellow.Printf("Title: %s\n\n", prTitle)
	fmt.Println(prBody)
	fmt.Println()
	_, _ = bold.Println("━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// Confirm action
	var action string
	actionPrompt := &survey.Select{
		Message: "Action:",
		Options: []string{"Create PR", "Create as draft", "Cancel"},
	}
	if err := survey.AskOne(actionPrompt, &action); err != nil {
		return err
	}

	if action == "Cancel" {
		return nil
	}

	draft := action == "Create as draft" || opts.Draft

	// Create PR
	s3 := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s3.Suffix = " Creating PR..."
	s3.Start()

	pr, err := createPRWithGH(ctx, prTitle, prBody, baseBranch, branch, draft)
	s3.Stop()

	if err != nil {
		return fmt.Errorf("failed to create PR: %w", err)
	}

	_, _ = green.Printf("✓ Created PR #%d\n", pr.Number)
	fmt.Println()
	blue := color.New(color.FgBlue)
	fmt.Printf("  %s\n", blue.Sprint(pr.URL))

	return nil
}

func createPRNonInteractive(ctx *CommandContext, opts *PRCommandOptions, branch string) error {
	// Determine base branch
	baseBranch := determineBaseBranch(ctx, opts.Base)

	// Show push summary
	showPushSummary(branch, baseBranch)

	// Push branch
	if err := pushBranch(); err != nil {
		return fmt.Errorf("failed to push branch: %w", err)
	}

	// Extract ticket ID
	ticketID := extractTicketIDIfNeeded(branch, opts)

	// Build PR body
	prBody, err := buildPRBodyFromOptions(ctx, opts, branch, ticketID)
	if err != nil {
		return err
	}

	// Determine title
	prTitle := determinePRTitle(ctx, opts, branch, ticketID)

	// Create PR
	pr, err := createPRWithGH(ctx, prTitle, prBody, baseBranch, branch, opts.Draft)
	if err != nil {
		return fmt.Errorf("failed to create PR: %w", err)
	}

	showPRCreated(pr)
	return nil
}

func buildPRBody(template, ticketID, summary, description, testing string) string {
	body := template

	// If template exists, try to fill in sections
	if template != "" {
		if summary != "" {
			body = strings.Replace(body, "## Summary", fmt.Sprintf("## Summary\n\n%s", summary), 1)
		}
		if ticketID != "" {
			body = strings.Replace(body, "CU-", fmt.Sprintf("CU-%s", ticketID), 1)
		}
		if description != "" {
			body = strings.Replace(body, "### Description", fmt.Sprintf("### Description\n\n%s", description), 1)
		}
		if testing != "" {
			body = strings.Replace(body, "### How to Test", fmt.Sprintf("### How to Test\n\n%s", testing), 1)
		}
	} else {
		// Build from scratch
		body = fmt.Sprintf(`## Summary

%s

### Description

%s

### How to Test

%s
`, summary, description, testing)

		if ticketID != "" {
			body = fmt.Sprintf(`## Summary

%s

Ticket: CU-%s

### Description

%s

### How to Test

%s
`, summary, ticketID, description, testing)
		}
	}

	return body
}

// Helper functions

func isGHAvailable() bool {
	cmd := exec.Command("gh", "auth", "status")
	return cmd.Run() == nil
}

func getPRForCurrentBranch() (*models.PullRequest, error) {
	cmd := exec.Command("gh", "pr", "view", "--json", "number,title,url,state")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// Parse JSON output
	var pr struct {
		Number int    `json:"number"`
		Title  string `json:"title"`
		URL    string `json:"url"`
		State  string `json:"state"`
	}

	if err := utils.ParseJSON(output, &pr); err != nil {
		return nil, err
	}

	return &models.PullRequest{
		Number: pr.Number,
		Title:  pr.Title,
		URL:    pr.URL,
		State:  pr.State,
	}, nil
}

func pushBranch() error {
	cmd := exec.Command("git", "push", "-u", "origin", "HEAD")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git push failed: %w\nOutput: %s", err, string(output))
	}
	return nil
}

func createPRWithGH(ctx *CommandContext, title, body, base, head string, draft bool) (*models.PullRequest, error) {
	req := &models.PRCreateRequest{
		Title: title,
		Body:  body,
		Head:  head,
		Base:  base,
		Draft: draft,
	}

	cmdCtx := context.Background()
	pr, err := ctx.GitHubClient.CreatePR(cmdCtx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create PR: %w", err)
	}

	return pr, nil
}

func getCommitCount(base, head string) (int, error) {
	cmd := exec.Command("git", "rev-list", "--count", fmt.Sprintf("%s..%s", base, head))
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	count := 0
	_, _ = fmt.Sscanf(strings.TrimSpace(string(output)), "%d", &count)
	return count, nil
}

func getChangedFilesCount(base, head string) (int, error) {
	cmd := exec.Command("git", "diff", "--name-only", fmt.Sprintf("%s...%s", base, head))
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return 0, nil
	}
	return len(lines), nil
}

// Helper functions to reduce cyclomatic complexity

func determineBaseBranch(ctx *CommandContext, optBase string) string {
	if optBase != "" {
		return optBase
	}
	exists, _ := ctx.GitRepo.BranchExists(ctx.Config.Git.BaseBranch)
	if exists {
		return ctx.Config.Git.BaseBranch
	}
	return "main"
}

func showPushSummary(branch, baseBranch string) {
	cyan := color.New(color.FgCyan)
	dim := color.New(color.Faint)
	fmt.Println()
	fmt.Printf("Branch: %s → %s\n", cyan.Sprint(branch), dim.Sprint(baseBranch))
}

func extractTicketIDIfNeeded(branch string, opts *PRCommandOptions) string {
	if opts.Summary != "" || opts.Description != "" || opts.BodyFile == "" {
		ticketID, _ := utils.ExtractTicketID(branch)
		return ticketID
	}
	return ""
}

func buildPRBodyFromOptions(ctx *CommandContext, opts *PRCommandOptions, _ string, ticketID string) (string, error) {
	// If body file provided, use it directly
	if opts.BodyFile != "" {
		data, err := os.ReadFile(opts.BodyFile)
		if err != nil {
			return "", fmt.Errorf("failed to read body file: %w", err)
		}
		return string(data), nil
	}

	// Read PR template
	repoRoot, _ := ctx.GitRepo.GetRootPath()
	templatePath := filepath.Join(repoRoot, ".github", "PULL_REQUEST_TEMPLATE.md")
	template := ""
	if data, err := os.ReadFile(templatePath); err == nil {
		template = string(data)
	} else {
		// Default template
		template = `## Summary

#### Ticket: CU-

### Description

### How to Test

## Best Practices
`
	}

	// Build PR body from template
	return buildPRBody(template, ticketID, opts.Summary, opts.Description, opts.Testing), nil
}

func determinePRTitle(ctx *CommandContext, opts *PRCommandOptions, branch, ticketID string) string {
	if opts.Title != "" {
		return opts.Title
	}

	// Try to get from ticket
	if ticketID != "" {
		cmdCtx := context.Background()
		task, err := ctx.ClickUpClient.GetTask(cmdCtx, ticketID)
		if err == nil && task != nil {
			return task.Name
		}
	}

	// Fallback to branch name
	return strings.ReplaceAll(branch, "-", " ")
}

func showPRCreated(pr *models.PullRequest) {
	green := color.New(color.FgGreen)
	blue := color.New(color.FgBlue)
	_, _ = green.Printf("✓ Created PR #%d\n", pr.Number)
	fmt.Printf("  %s\n", blue.Sprint(pr.URL))
}
