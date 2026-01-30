package commands

import (
	"context"
	"fmt"

	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/rithyhuot/vibe/internal/models"
	"github.com/rithyhuot/vibe/internal/utils"
	"github.com/spf13/cobra"
)

// NewVibeCommand creates the vibe command
func NewVibeCommand(ctx *CommandContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vibe <ticket-id>",
		Short: "Start working on a ticket",
		Long: `Fetches a task from ClickUp, creates a branch, and updates the task status.

Examples:
  vibe workon abc123             # Start working on ticket abc123
  vibe workon 86b7x5453          # Start working with full ClickUp ID
  vibe abc123                    # Shorthand: vibe command works the same`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return runVibe(ctx, args[0])
		},
	}

	return cmd
}

func runVibe(ctx *CommandContext, ticketID string) error {
	// Validate ticket ID
	if !utils.IsTicketID(ticketID) {
		return fmt.Errorf("invalid ticket ID format: %s (expected 9 alphanumeric characters)", ticketID)
	}

	cmdCtx := context.Background()

	// Create spinner
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Fetching task from ClickUp..."
	s.Start()

	// Fetch task from ClickUp
	task, err := ctx.ClickUpClient.GetTask(cmdCtx, ticketID)
	if err != nil {
		s.Stop()
		return fmt.Errorf("failed to fetch task: %w", err)
	}

	s.Stop()

	// Display task details
	displayTask(task)

	// Generate branch name
	branchName := utils.GenerateBranchName(ctx.Config.Git.BranchPrefix, ticketID, task.Name)

	// Validate branch name
	if err := utils.ValidateBranchName(branchName); err != nil {
		return fmt.Errorf("invalid branch name: %w", err)
	}

	// Check if branch already exists
	exists, err := ctx.GitRepo.BranchExists(branchName)
	if err != nil {
		return fmt.Errorf("failed to check if branch exists: %w", err)
	}

	if exists {
		yellow := color.New(color.FgYellow)
		_, _ = yellow.Printf("\n⚠ Branch '%s' already exists\n", branchName)

		s = spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Suffix = " Checking out existing branch..."
		s.Start()

		err = ctx.GitRepo.Checkout(branchName)
		s.Stop()
		if err != nil {
			return fmt.Errorf("failed to checkout branch: %w", err)
		}
	} else {
		// Create new branch
		s = spinner.New(spinner.CharSets[14], 100*time.Millisecond)
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

		green := color.New(color.FgGreen)
		_, _ = green.Printf("✓ Created and checked out branch: %s\n", branchName)
	}

	// Update task status to "In Progress" if not already
	if task.Status.Status != ctx.Config.Defaults.Status {
		s = spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Suffix = " Updating task status..."
		s.Start()

		status := ctx.Config.Defaults.Status
		updateReq := &models.TaskUpdateRequest{
			Status: &status,
		}

		_, err = ctx.ClickUpClient.UpdateTask(cmdCtx, ticketID, updateReq)
		s.Stop()
		if err != nil {
			yellow := color.New(color.FgYellow)
			_, _ = yellow.Printf("⚠ Failed to update task status: %v\n", err)
		} else {
			green := color.New(color.FgGreen)
			_, _ = green.Printf("✓ Updated task status to: %s\n", ctx.Config.Defaults.Status)
		}
	}

	fmt.Println()
	cyan := color.New(color.FgCyan, color.Bold)
	_, _ = cyan.Println("Ready to start working! 🚀")

	return nil
}

func displayTask(task *models.Task) {
	fmt.Println()
	titleColor := color.New(color.FgCyan, color.Bold)
	_, _ = titleColor.Printf("📋 %s\n", task.Name)
	fmt.Println()

	// Task ID
	fmt.Printf("ID:     %s\n", task.ID)

	// Status
	statusColor := getStatusColor(task.Status.Status)
	fmt.Printf("Status: ")
	_, _ = statusColor.Printf("%s\n", task.Status.Status)

	// Assignees
	if len(task.Assignees) > 0 {
		fmt.Printf("Assigned: ")
		for i, assignee := range task.Assignees {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Print(assignee.Username)
		}
		fmt.Println()
	}

	// Custom fields
	if taskType := task.GetCustomFieldString("Type"); taskType != "" {
		fmt.Printf("Type:   %s\n", taskType)
	}
	if domain := task.GetCustomFieldString("Domain"); domain != "" {
		fmt.Printf("Domain: %s\n", domain)
	}

	// Description
	if task.Description != "" {
		fmt.Println()
		gray := color.New(color.FgHiBlack)
		_, _ = gray.Println("Description:")
		fmt.Println(task.Description)
	}

	// URL
	fmt.Println()
	fmt.Printf("URL: %s\n", task.URL)
	fmt.Println()
}

func getStatusColor(status string) *color.Color {
	switch status {
	case "In Progress", "in progress":
		return color.New(color.FgYellow)
	case "Done", "done", "Complete", "complete":
		return color.New(color.FgGreen)
	case "To Do", "to do", "Open", "open":
		return color.New(color.FgBlue)
	default:
		return color.New(color.FgWhite)
	}
}
