package commands

import (
	"context"
	"fmt"
	"strconv"
	"time"

	survey "github.com/AlecAivazis/survey/v2"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// NewMergeCommand creates the merge command
func NewMergeCommand(ctx *CommandContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "merge [pr-number]",
		Short: "Post /merge comment to trigger merge automation",
		Long: `Posts a /merge comment on the PR to trigger merge automation. Checks PR status first. If no PR number is provided, uses the current branch's PR.

Examples:
  vibe merge                     # Merge PR for current branch
  vibe merge 123                 # Merge PR #123
  vibe merge --force             # Force merge (if implemented)`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			prNumberArg := ""
			if len(args) > 0 {
				prNumberArg = args[0]
			}
			return runMerge(ctx, prNumberArg)
		},
	}

	return cmd
}

func runMerge(ctx *CommandContext, prNumberArg string) error {
	// Check if gh CLI is available
	if !isGHAvailable() {
		return fmt.Errorf("GitHub CLI (gh) is not available. Run: gh auth login")
	}

	// Get PR number
	prNumber, err := getPRNumber(prNumberArg)
	if err != nil {
		return err
	}

	// Get PR status first
	prInfo, statusInfo, err := fetchPRDetails(prNumber)
	if err != nil {
		return err
	}

	// Display PR info
	displayMergePRInfo(prNumber, prInfo, statusInfo)

	// Determine if ready
	isReady := isPRReadyToMerge(statusInfo)

	return handleMergeAction(ctx, prNumber, isReady)
}

func getPRNumber(prNumberArg string) (string, error) {
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

func fetchPRDetails(prNumber string) (*PRInfo, *PRStatusInfo, error) {
	// Use empty string for repo to query the current repository
	prInfo, err := getPRInfo(prNumber, "")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch PR: %w", err)
	}

	statusInfo, err := getPRStatusInfo(prNumber, "")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch PR status: %w", err)
	}

	return prInfo, statusInfo, nil
}

func displayMergePRInfo(prNumber string, prInfo *PRInfo, statusInfo *PRStatusInfo) {
	bold := color.New(color.Bold)
	dim := color.New(color.Faint)

	fmt.Println()
	_, _ = bold.Printf("PR #%s: %s\n", prNumber, prInfo.Title)
	_, _ = dim.Println(prInfo.URL)
	fmt.Println()

	displayPRStatus(statusInfo)
	fmt.Println()
}

func isPRReadyToMerge(statusInfo *PRStatusInfo) bool {
	return statusInfo.CIPassed &&
		statusInfo.CIPending == 0 &&
		len(statusInfo.Approvals) > 0 &&
		len(statusInfo.ChangesRequested) == 0
}

func handleMergeAction(ctx *CommandContext, prNumber string, isReady bool) error {
	if isReady {
		return handleReadyMerge(ctx, prNumber)
	}
	return handleForcedMerge(ctx, prNumber)
}

func handleReadyMerge(ctx *CommandContext, prNumber string) error {
	var shouldMerge bool
	prompt := &survey.Confirm{
		Message: "Post /merge comment?",
		Default: true,
	}
	if err := survey.AskOne(prompt, &shouldMerge); err != nil {
		return err
	}

	if !shouldMerge {
		return nil
	}

	if err := postMergeComment(ctx, prNumber); err != nil {
		return err
	}

	green := color.New(color.FgGreen)
	dim := color.New(color.Faint)
	_, _ = green.Println("✓ Posted /merge comment")
	_, _ = dim.Println("  Merge automation will process shortly.")
	return nil
}

func handleForcedMerge(ctx *CommandContext, prNumber string) error {
	yellow := color.New(color.FgYellow)
	dim := color.New(color.Faint)

	_, _ = yellow.Println("PR is not ready to merge.")
	fmt.Println()

	var forceAnyway bool
	prompt := &survey.Confirm{
		Message: yellow.Sprint("Post /merge anyway? (not recommended)"),
		Default: false,
	}
	if err := survey.AskOne(prompt, &forceAnyway); err != nil {
		return err
	}

	if !forceAnyway {
		return nil
	}

	if err := postMergeComment(ctx, prNumber); err != nil {
		return err
	}

	_, _ = yellow.Println("⚠ Posted /merge comment (forced)")
	_, _ = dim.Println("  The merge may fail if requirements are not met.")
	return nil
}

func postMergeComment(ctx *CommandContext, prNumber string) error {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Posting /merge comment..."
	s.Start()
	defer s.Stop()

	prNum, _ := strconv.Atoi(prNumber)
	cmdCtx := context.Background()
	err := ctx.GitHubClient.AddComment(cmdCtx, prNum, "/merge")
	if err != nil {
		return fmt.Errorf("failed to post comment: %w", err)
	}

	return nil
}
