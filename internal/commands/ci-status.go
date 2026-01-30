package commands

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/rithyhuot/vibe/internal/services/circleci"
	"github.com/spf13/cobra"
)

// NewCIStatusCommand creates the ci-status command
func NewCIStatusCommand(ctx *CommandContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ci-status [branch]",
		Short: "Show CI status for a branch",
		Long: `Shows CircleCI pipeline status including workflows, jobs, and failed tests. If no branch is provided, uses the current branch.

Examples:
  vibe ci-status                 # Check CI for current branch
  vibe ci-status main            # Check CI for main branch
  vibe ci-status feature-branch  # Check CI for specific branch`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			branch := ""
			if len(args) > 0 {
				branch = args[0]
			}
			return runCIStatus(ctx, branch)
		},
	}

	return cmd
}

func runCIStatus(ctx *CommandContext, branchArg string) error {
	// Get CircleCI token
	token := ctx.Config.CircleCI.APIToken
	if token == "" {
		token = os.Getenv("CIRCLECI_TOKEN")
		if token == "" {
			token = os.Getenv("CIRCLE_TOKEN")
		}
	}

	if token == "" {
		return fmt.Errorf("CircleCI API token not found.\nSet one of the following:\n  - Add circleci.apiToken to your vibe config\n  - Set CIRCLECI_TOKEN environment variable\n  - Set CIRCLE_TOKEN environment variable")
	}

	// Get branch
	branch := branchArg
	if branch == "" {
		var err error
		branch, err = ctx.GitRepo.CurrentBranch()
		if err != nil {
			return fmt.Errorf("failed to get current branch: %w", err)
		}
	}

	// Get project slug
	projectSlug, err := circleci.GetProjectSlug()
	if err != nil {
		return fmt.Errorf("could not determine project from git remote: %w\nMake sure you have a GitHub remote configured", err)
	}

	cyan := color.New(color.FgCyan)

	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = fmt.Sprintf(" Checking CI status for %s...", cyan.Sprint(branch))
	s.Start()

	// Create CircleCI client
	client := circleci.NewClient(token)
	cmdCtx := context.Background()

	status, err := client.GetCIStatusForBranch(cmdCtx, branch, projectSlug)
	s.Stop()

	if err != nil {
		return fmt.Errorf("failed to fetch CI status: %w", err)
	}

	if status == nil {
		yellow := color.New(color.FgYellow)
		_, _ = yellow.Printf("\nNo CI pipelines found for branch: %s\n", branch)
		dim := color.New(color.Faint)
		_, _ = dim.Println("This branch may not have been pushed or CircleCI may not be configured.")
		return nil
	}

	// Display CI status
	displayCIStatus(status)

	return nil
}

func displayCIStatus(status *circleci.CIStatus) {
	displayCIHeader(status)

	// Aggregate and display workflows
	jobCounts := aggregateAndDisplayWorkflows(status.Workflows)

	// Display failed job details
	displayFailedJobs(status.FailedJobs)

	// Display overall status
	displayOverallStatus(jobCounts, status.Workflows)
}

type jobCounts struct {
	running   int
	pending   int
	failed    int
	success   int
	isRunning bool
}

func displayCIHeader(status *circleci.CIStatus) {
	bold := color.New(color.Bold)
	dim := color.New(color.Faint)

	fmt.Println()
	_, _ = bold.Printf("CI Status: %s\n", status.Branch)
	_, _ = dim.Printf("Project: %s\n", status.ProjectSlug)
	_, _ = dim.Printf("Pipeline #%d\n", status.PipelineNumber)
	fmt.Println()
}

func aggregateAndDisplayWorkflows(workflows []circleci.WorkflowStatus) jobCounts {
	bold := color.New(color.Bold)
	counts := jobCounts{}

	for _, workflow := range workflows {
		wfCounts := countWorkflowJobs(workflow)

		counts.running += wfCounts.running
		counts.pending += wfCounts.pending
		counts.failed += wfCounts.failed
		counts.success += wfCounts.success

		if workflow.Status == "running" || wfCounts.running > 0 || wfCounts.pending > 0 {
			counts.isRunning = true
		}

		// Display workflow status
		_, _ = bold.Printf("%s: ", workflow.Name)
		fmt.Println(formatWorkflowStatus(workflow.Status))

		// Display jobs
		for _, job := range workflow.Jobs {
			fmt.Printf("  %s %s\n", formatJobStatus(job.Status), job.Name)
		}
		fmt.Println()
	}

	return counts
}

func countWorkflowJobs(workflow circleci.WorkflowStatus) jobCounts {
	counts := jobCounts{}
	for _, job := range workflow.Jobs {
		switch job.Status {
		case "running":
			counts.running++
		case "queued", "not_run":
			counts.pending++
		case "failed":
			counts.failed++
		case "success":
			counts.success++
		}
	}
	return counts
}

func displayFailedJobs(failedJobs []circleci.FailedJob) {
	if len(failedJobs) == 0 {
		return
	}

	red := color.New(color.FgRed)
	_, _ = red.Println("─── Failed Jobs ───")
	for _, job := range failedJobs {
		displayFailedJob(&job)
	}
	fmt.Println()
}

func displayOverallStatus(counts jobCounts, workflows []circleci.WorkflowStatus) {
	green := color.New(color.FgGreen)
	red := color.New(color.FgRed)
	blue := color.New(color.FgBlue)
	yellow := color.New(color.FgYellow)

	running := counts.isRunning || counts.running > 0 || counts.pending > 0
	passed := !running && counts.failed == 0 && allWorkflowsSuccess(workflows)

	if running {
		inProgressCount := counts.running + counts.pending
		jobText := "job"
		if inProgressCount != 1 {
			jobText = "jobs"
		}
		_, _ = blue.Printf("CI is still running... (%d %s pending)\n", inProgressCount, jobText)
	} else if passed {
		_, _ = green.Println("All CI checks passed!")
	} else if counts.failed > 0 {
		_, _ = red.Printf("%d job(s) failed.\n", counts.failed)
	} else {
		statuses := getUniqueStatuses(workflows)
		_, _ = yellow.Printf("CI status: %s\n", statuses)
	}
}

func formatJobStatus(status string) string {
	green := color.New(color.FgGreen)
	red := color.New(color.FgRed)
	blue := color.New(color.FgBlue)
	gray := color.New(color.Faint)
	yellow := color.New(color.FgYellow)

	switch status {
	case "success":
		return green.Sprint("✓ passed")
	case "failed":
		return red.Sprint("✗ failed")
	case "running":
		return blue.Sprint("◔ running")
	case "queued", "not_run":
		return gray.Sprint("○ pending")
	case "canceled":
		return yellow.Sprint("⊘ canceled")
	default:
		return gray.Sprint(status)
	}
}

func formatWorkflowStatus(status string) string {
	green := color.New(color.FgGreen, color.Bold)
	red := color.New(color.FgRed, color.Bold)
	blue := color.New(color.FgBlue, color.Bold)
	yellow := color.New(color.FgYellow, color.Bold)
	gray := color.New(color.Faint, color.Bold)

	switch status {
	case "success":
		return green.Sprint("PASSED")
	case "failed":
		return red.Sprint("FAILED")
	case "running":
		return blue.Sprint("RUNNING")
	case "canceled":
		return yellow.Sprint("CANCELED")
	default:
		return gray.Sprint(strings.ToUpper(status))
	}
}

func displayFailedJob(job *circleci.FailedJob) {
	red := color.New(color.FgRed)
	dim := color.New(color.Faint)
	yellow := color.New(color.FgYellow)
	cyan := color.New(color.FgCyan)

	fmt.Println()
	_, _ = red.Printf("  %s > %s\n", job.WorkflowName, job.Name)
	_, _ = dim.Printf("  %s\n", job.WebURL)

	if len(job.FailedTests) > 0 {
		_, _ = red.Printf("\n  Failed tests (%d):\n", len(job.FailedTests))
		maxTests := 5
		if len(job.FailedTests) < maxTests {
			maxTests = len(job.FailedTests)
		}

		for i := 0; i < maxTests; i++ {
			test := job.FailedTests[i]
			testName := test.Name
			if test.Classname != "" {
				testName = fmt.Sprintf("%s > %s", test.Classname, test.Name)
			}

			fmt.Println()
			_, _ = red.Printf("    %s\n", testName)
			if test.File != "" {
				_, _ = dim.Printf("    File: %s\n", test.File)
			}
			if test.Message != "" {
				_, _ = yellow.Println("    Error:")
				lines := strings.Split(test.Message, "\n")
				maxLines := 20
				if len(lines) < maxLines {
					maxLines = len(lines)
				}
				for j := 0; j < maxLines; j++ {
					_, _ = dim.Printf("      %s\n", lines[j])
				}
				if len(lines) > maxLines {
					_, _ = dim.Println("      ... (truncated)")
				}
			}
		}

		if len(job.FailedTests) > maxTests {
			fmt.Println()
			_, _ = dim.Printf("    ... and %d more failed tests\n", len(job.FailedTests)-maxTests)
		}
	}

	fmt.Println()
	_, _ = dim.Printf("  Run %s for full output\n", cyan.Sprintf("vibe ci-failure %d", job.JobNumber))
}

func allWorkflowsSuccess(workflows []circleci.WorkflowStatus) bool {
	for _, w := range workflows {
		if w.Status != "success" {
			return false
		}
	}
	return true
}

func getUniqueStatuses(workflows []circleci.WorkflowStatus) string {
	statusMap := make(map[string]bool)
	for _, w := range workflows {
		statusMap[w.Status] = true
	}

	var statuses []string
	for status := range statusMap {
		statuses = append(statuses, status)
	}

	return strings.Join(statuses, ", ")
}
