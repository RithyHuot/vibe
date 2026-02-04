// Package commands contains the CLI command implementations.
package commands

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/rithyhuot/vibe/internal/services/circleci"
)

// CIFailureOptions holds flags for the ci-failure command
type CIFailureOptions struct {
	Branch string
}

// NewCIFailureCommand creates the ci-failure command
func NewCIFailureCommand(ctx *CommandContext) *cobra.Command {
	opts := &CIFailureOptions{}

	cmd := &cobra.Command{
		Use:   "ci-failure [job-number]",
		Short: "Show detailed failure output from a CI job",
		Long: `Shows the full, untruncated error output from a failed CircleCI job. If no job number is provided, uses the first failed job from the current branch.

Examples:
  vibe ci-failure                # Show failure from current branch's first failed job
  vibe ci-failure 12345          # Show failure details for job #12345
  vibe ci-failure --branch main  # Show failure from main branch`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			jobNumberArg := ""
			if len(args) > 0 {
				jobNumberArg = args[0]
			}
			return runCIFailure(ctx, opts, jobNumberArg)
		},
	}

	cmd.Flags().StringVar(&opts.Branch, "branch", "", "Branch to check (default: current branch)")

	return cmd
}

func runCIFailure(ctx *CommandContext, opts *CIFailureOptions, jobNumberArg string) error {
	// Get CircleCI token
	token := ctx.Config.CircleCI.APIToken
	if token == "" {
		token = os.Getenv("CIRCLECI_TOKEN")
		if token == "" {
			token = os.Getenv("CIRCLE_TOKEN")
		}
	}

	if token == "" {
		return fmt.Errorf("CircleCI API token not found")
	}

	// Get project slug
	projectSlug, err := circleci.GetProjectSlug()
	if err != nil {
		return fmt.Errorf("could not determine project from git remote: %w", err)
	}

	// Create CircleCI client
	client := circleci.NewClient(token)
	cmdCtx := context.Background()

	var jobNumber int

	// If job number provided, use it directly
	if jobNumberArg != "" {
		jobNumber, err = strconv.Atoi(jobNumberArg)
		if err != nil {
			return fmt.Errorf("invalid job number: %s", jobNumberArg)
		}
	} else {
		// Otherwise, find the first failed job for the branch
		branch := opts.Branch
		if branch == "" {
			branch, err = ctx.GitRepo.CurrentBranch()
			if err != nil {
				return fmt.Errorf("failed to get current branch: %w", err)
			}
		}

		cyan := color.New(color.FgCyan)
		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Suffix = fmt.Sprintf(" Finding failed jobs for %s...", cyan.Sprint(branch))
		s.Start()

		status, err := client.GetCIStatusForBranch(cmdCtx, branch, projectSlug)
		s.Stop()

		if err != nil {
			return fmt.Errorf("failed to fetch CI status: %w", err)
		}

		if status == nil {
			yellow := color.New(color.FgYellow)
			_, _ = yellow.Printf("No CI pipelines found for branch: %s\n", branch)
			return nil
		}

		if len(status.FailedJobs) == 0 {
			green := color.New(color.FgGreen)
			_, _ = green.Println("No failed jobs found.")
			return nil
		}

		// Use the first failed job
		failedJob := status.FailedJobs[0]
		jobNumber = failedJob.JobNumber

		dim := color.New(color.Faint)
		_, _ = dim.Printf("Found failed job: %s > %s\n", failedJob.WorkflowName, failedJob.Name)
	}

	// Fetch failure details
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = fmt.Sprintf(" Fetching failure details for job #%d...", jobNumber)
	s.Start()

	failedSteps, err := client.GetBuildDetails(cmdCtx, projectSlug, jobNumber)
	s.Stop()

	if err != nil {
		return fmt.Errorf("failed to fetch failure details: %w", err)
	}

	if len(failedSteps) == 0 {
		yellow := color.New(color.FgYellow)
		_, _ = yellow.Println("No failed steps found for this job.")
		return nil
	}

	// Display failed steps
	displayFailedSteps(failedSteps)

	return nil
}

func displayFailedSteps(steps []circleci.FailedStep) {
	red := color.New(color.FgRed, color.Bold)
	dim := color.New(color.Faint)

	for _, step := range steps {
		fmt.Println()
		_, _ = red.Printf("═══ Failed Step: %s ═══\n\n", step.Name)

		for _, action := range step.Actions {
			if action.Output != "" {
				fmt.Println(action.Output)
			} else {
				if action.ExitCode != nil {
					_, _ = dim.Printf("Exit code: %d\n", *action.ExitCode)
				}
			}
		}
	}
}
