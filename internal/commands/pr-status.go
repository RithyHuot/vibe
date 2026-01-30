package commands

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/rithyhuot/vibe/internal/utils"
	"github.com/spf13/cobra"
)

// NewPRStatusCommand creates the pr-status command
func NewPRStatusCommand(ctx *CommandContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pr-status [pr-number]",
		Short: "Check PR approval and CI status",
		Long: `Shows the status of a pull request including reviews, checks, and CI status. If no PR number is provided, uses the current branch's PR.

Examples:
  vibe pr-status                 # Check status for current branch's PR
  vibe pr-status 123             # Check status for PR #123
  vibe pr-status --json          # Output as JSON (if implemented)`,
		RunE: func(_ *cobra.Command, args []string) error {
			prNumber := ""
			if len(args) > 0 {
				prNumber = args[0]
			}
			return runPRStatus(ctx, prNumber)
		},
	}

	return cmd
}

func runPRStatus(ctx *CommandContext, prNumberArg string) error {
	// Check if gh CLI is available
	if !isGHAvailable() {
		return fmt.Errorf("GitHub CLI (gh) is not available. Run: gh auth login")
	}

	// Get PR number
	prNumber, err := resolvePRNumberForStatus(prNumberArg)
	if err != nil {
		return err
	}

	// Fetch PR status
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Checking status..."
	s.Start()

	prInfo, err := fetchPRInfoWithFallback(ctx, prNumber, s)
	if err != nil {
		s.Stop()
		return err
	}

	// Use detected repo from prInfo for status query
	statusRepo := prInfo.Repo
	// Skip if repo has placeholder values
	if strings.Contains(statusRepo, "org-name") || strings.Contains(statusRepo, "repo-name") ||
		strings.Contains(statusRepo, "your-") {
		statusRepo = ""
	}

	status, err := getPRStatusInfo(prNumber, statusRepo)
	if err != nil {
		s.Stop()
		return fmt.Errorf("failed to fetch PR status: %w", err)
	}

	s.Stop()

	displayPRStatusAndReadiness(prNumber, prInfo, status)
	return nil
}

func resolvePRNumberForStatus(prNumberArg string) (string, error) {
	if prNumberArg != "" {
		return prNumberArg, nil
	}

	// Get PR for current branch
	pr, err := getPRForCurrentBranch()
	if err != nil || pr == nil {
		yellow := color.New(color.FgYellow)
		_, _ = yellow.Println("No PR found for this branch.")
		return "", fmt.Errorf("no PR found for current branch")
	}

	return fmt.Sprintf("%d", pr.Number), nil
}

func fetchPRInfoWithFallback(ctx *CommandContext, prNumber string, s *spinner.Spinner) (*PRInfo, error) {
	// Try with config repo first
	var configRepo string
	if ctx != nil && ctx.Config != nil {
		configRepo = fmt.Sprintf("%s/%s", ctx.Config.GitHub.Owner, ctx.Config.GitHub.Repo)
		// Skip if config has placeholder values
		if strings.Contains(configRepo, "org-name") || strings.Contains(configRepo, "repo-name") ||
			strings.Contains(configRepo, "your-") {
			configRepo = ""
		}
	}

	prInfo, err := getPRInfo(prNumber, configRepo)

	// If not found in config repo, try with git remote repo
	if err != nil && (strings.Contains(err.Error(), "Could not resolve to a PullRequest") ||
		strings.Contains(err.Error(), "Could not resolve to a Repository")) {
		owner, repo, repoErr := getRepoFromGitRemote()
		gitRemoteRepo := fmt.Sprintf("%s/%s", owner, repo)
		if repoErr == nil && gitRemoteRepo != "" && gitRemoteRepo != configRepo {
			dim := color.New(color.Faint)
			s.Stop()
			_, _ = dim.Printf("PR not found in configured repo, trying %s...\n", gitRemoteRepo)
			s.Start()

			prInfo, err = getPRInfo(prNumber, gitRemoteRepo)
		}
	}

	if err != nil {
		// Check if it's a "not found" error
		if strings.Contains(err.Error(), "Could not resolve to a PullRequest") {
			return nil, fmt.Errorf("PR #%s not found in this repository", prNumber)
		}
		if strings.Contains(err.Error(), "Could not resolve to a Repository") {
			return nil, fmt.Errorf("repository not found - check your config at ~/.config/vibe/config.yaml")
		}
		return nil, fmt.Errorf("failed to fetch PR: %w", err)
	}

	return prInfo, nil
}

func displayPRStatusAndReadiness(prNumber string, prInfo *PRInfo, status *PRStatusInfo) {
	// Display PR info
	bold := color.New(color.Bold)
	dim := color.New(color.Faint)

	fmt.Println()
	_, _ = bold.Printf("PR #%s: %s\n", prNumber, prInfo.Title)
	_, _ = dim.Println(prInfo.URL)
	fmt.Println()

	// Display status
	displayPRStatus(status)

	// Determine readiness
	isReady := status.CIPassed &&
		status.CIPending == 0 &&
		len(status.Approvals) > 0 &&
		len(status.ChangesRequested) == 0

	fmt.Println()
	if isReady {
		green := color.New(color.FgGreen, color.Bold)
		_, _ = green.Println("Status: READY TO MERGE ✓")
	} else {
		red := color.New(color.FgRed, color.Bold)
		_, _ = red.Println("Status: NOT READY")
	}
}

// PRInfo represents basic pull request information
type PRInfo struct {
	Number int
	Title  string
	URL    string
	State  string
	Repo   string // owner/repo format
}

// PRStatusInfo represents the status of a pull request
type PRStatusInfo struct {
	CIPassed         bool
	CIPending        int
	CIFailed         []string
	Approvals        []string
	ChangesRequested []string
}

func getPRInfo(prNumber, repo string) (*PRInfo, error) {
	args := []string{"pr", "view", prNumber, "--json", "number,title,url,state"}
	if repo != "" {
		args = append(args, "-R", repo)
	}

	cmd := exec.Command("gh", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, string(output))
	}

	var pr struct {
		Number int    `json:"number"`
		Title  string `json:"title"`
		URL    string `json:"url"`
		State  string `json:"state"`
	}

	if err := utils.ParseJSON(output, &pr); err != nil {
		return nil, err
	}

	return &PRInfo{
		Number: pr.Number,
		Title:  pr.Title,
		URL:    pr.URL,
		State:  pr.State,
		Repo:   repo,
	}, nil
}

func getPRStatusInfo(prNumber, repo string) (*PRStatusInfo, error) {
	args := []string{"pr", "view", prNumber, "--json", "reviewDecision,reviews,statusCheckRollup"}
	if repo != "" {
		args = append(args, "-R", repo)
	}

	cmd := exec.Command("gh", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var data struct {
		ReviewDecision string `json:"reviewDecision"`
		Reviews        []struct {
			Author struct {
				Login string `json:"login"`
			} `json:"author"`
			State string `json:"state"`
		} `json:"reviews"`
		StatusCheckRollup []struct {
			Name       string  `json:"name"`
			Status     string  `json:"status"`
			Conclusion *string `json:"conclusion"`
		} `json:"statusCheckRollup"`
	}

	if err := utils.ParseJSON(output, &data); err != nil {
		return nil, err
	}

	status := &PRStatusInfo{}

	// Parse reviews
	approvals := make(map[string]bool)
	changesRequested := make(map[string]bool)

	for _, review := range data.Reviews {
		switch review.State {
		case "APPROVED":
			approvals[review.Author.Login] = true
			delete(changesRequested, review.Author.Login)
		case "CHANGES_REQUESTED":
			changesRequested[review.Author.Login] = true
			delete(approvals, review.Author.Login)
		}
	}

	for login := range approvals {
		status.Approvals = append(status.Approvals, login)
	}
	for login := range changesRequested {
		status.ChangesRequested = append(status.ChangesRequested, login)
	}

	// Parse CI checks
	ciPassed := true
	for _, check := range data.StatusCheckRollup {
		if check.Conclusion != nil {
			switch *check.Conclusion {
			case "success":
				// passed
			case "failure", "timed_out", "action_required":
				ciPassed = false
				status.CIFailed = append(status.CIFailed, check.Name)
			}
		} else {
			// No conclusion yet means pending
			status.CIPending++
			ciPassed = false
		}
	}

	status.CIPassed = ciPassed && len(status.CIFailed) == 0 && status.CIPending == 0

	return status, nil
}

func displayPRStatus(status *PRStatusInfo) {
	green := color.New(color.FgGreen)
	red := color.New(color.FgRed)
	yellow := color.New(color.FgYellow)
	dim := color.New(color.Faint)

	// Display CI status
	if status.CIPassed {
		fmt.Printf("  %s CI checks passed\n", green.Sprint("✓"))
	} else if len(status.CIFailed) > 0 {
		fmt.Printf("  %s CI checks failed\n", red.Sprint("✗"))
		for _, check := range status.CIFailed {
			fmt.Printf("    %s %s\n", red.Sprint("•"), check)
		}
	} else if status.CIPending > 0 {
		fmt.Printf("  %s CI checks pending (%d)\n", yellow.Sprint("⋯"), status.CIPending)
	} else {
		fmt.Printf("  %s No CI checks\n", dim.Sprint("—"))
	}

	// Display review status
	if len(status.Approvals) > 0 {
		fmt.Printf("  %s Approved by: %s\n", green.Sprint("✓"), strings.Join(status.Approvals, ", "))
	} else {
		fmt.Printf("  %s No approvals\n", dim.Sprint("—"))
	}

	if len(status.ChangesRequested) > 0 {
		fmt.Printf("  %s Changes requested by: %s\n", red.Sprint("✗"), strings.Join(status.ChangesRequested, ", "))
	}
}
