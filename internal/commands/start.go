package commands

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/rithyhuot/vibe/internal/ui"
	"github.com/rithyhuot/vibe/internal/utils"
	"github.com/spf13/cobra"
)

// NewStartCommand creates the start command
func NewStartCommand(ctx *CommandContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start [ticket-id]",
		Short: "Interactive mode to search and start work on a ticket",
		Long: `Interactively search for tickets, select from a list, and start working on them. If ticket ID is provided, starts work directly.

Examples:
  vibe start                     # Interactive ticket selection
  vibe start abc123              # Start directly with ticket ID
  vibe start --search "bug fix"  # Search for tickets (if implemented)`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			ticketIDArg := ""
			if len(args) > 0 {
				ticketIDArg = args[0]
			}
			return runStart(ctx, ticketIDArg)
		},
	}

	return cmd
}

func runStart(ctx *CommandContext, ticketIDArg string) error {
	// If ticket ID provided, go straight to existing flow
	if ticketIDArg != "" && utils.IsTicketID(ticketIDArg) {
		return startFromExisting(ctx, ticketIDArg)
	}

	// Safety check: warn if not on base branch
	currentBranch, err := ctx.GitRepo.CurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	baseBranch := ctx.Config.Git.BaseBranch
	if currentBranch != baseBranch {
		_, _ = ui.Warning.Printf("\n⚠️  Warning: You are on branch '%s', not '%s'.\n", currentBranch, baseBranch)
		_, _ = ui.Warning.Println("   Creating a new branch from here may not be what you intended.")
		fmt.Println()

		var proceed bool
		prompt := &survey.Confirm{
			Message: fmt.Sprintf("Continue creating a new branch from '%s'?", currentBranch),
			Default: false,
		}
		if err := survey.AskOne(prompt, &proceed); err != nil {
			return err
		}

		if !proceed {
			_, _ = ui.Dim.Printf("Tip: Run 'git checkout %s' first, then try again.\n", baseBranch)
			return nil
		}
	}

	// Prompt for ticket ID or search term
	var userInput string
	inputPrompt := &survey.Input{
		Message: "Ticket ID or search:",
	}
	if err := survey.AskOne(inputPrompt, &userInput); err != nil {
		return err
	}

	userInput = strings.TrimSpace(userInput)

	if userInput == "" {
		return fmt.Errorf("ticket ID or search term required")
	}

	if utils.IsTicketID(userInput) {
		// Direct ticket ID
		return startFromExisting(ctx, strings.ToLower(userInput))
	}

	// Search term
	return handleSearchAndSelect(ctx, userInput)
}

func handleSearchAndSelect(ctx *CommandContext, searchTerm string) error {
	s := ui.CreateSpinner("Searching for tickets...")
	s.Start()

	cmdCtx := context.Background()
	// Use team ID from config for searching
	teamID := ctx.Config.ClickUp.TeamID
	tasks, err := ctx.ClickUpClient.SearchTeamTasks(cmdCtx, teamID, searchTerm)
	if err != nil {
		s.Stop()
		return fmt.Errorf("failed to search tasks: %w", err)
	}

	s.Stop()

	if len(tasks) == 0 {
		_, _ = ui.Warning.Printf("No tasks found matching '%s'\n", searchTerm)
		return nil
	}

	_, _ = ui.Success.Printf("Found %d task(s) matching '%s'\n\n", len(tasks), searchTerm)

	// Create selection options
	options := make([]string, len(tasks))
	taskMap := make(map[string]string) // map option string to task ID

	for i, task := range tasks {
		// Format: "ticket-id - Title [Status]"
		// Truncate title if too long
		title := task.Name
		titleRunes := []rune(title)
		maxTitleLength := 60
		if len(titleRunes) > maxTitleLength {
			title = string(titleRunes[:maxTitleLength-3]) + "..."
		}

		// Extract short ticket ID from full ID
		ticketID := task.ID
		if len(ticketID) > 9 {
			ticketID = ticketID[len(ticketID)-9:]
		}

		option := fmt.Sprintf("%s - %s [%s]", ticketID, title, task.Status.Status)
		options[i] = option
		taskMap[option] = task.ID
	}

	// Prompt for selection
	var selected string
	prompt := &survey.Select{
		Message: "Select a ticket to start working on:",
		Options: options,
	}

	if err := survey.AskOne(prompt, &selected); err != nil {
		return err
	}

	// Get selected task ID
	selectedTaskID := taskMap[selected]
	if selectedTaskID == "" {
		return fmt.Errorf("failed to find selected task")
	}

	// Extract short ticket ID for startFromExisting
	shortTicketID := selectedTaskID
	if len(selectedTaskID) > 9 {
		shortTicketID = selectedTaskID[len(selectedTaskID)-9:]
	}

	fmt.Println()
	return startFromExisting(ctx, strings.ToLower(shortTicketID))
}

func startFromExisting(ctx *CommandContext, ticketID string) error {
	s := ui.CreateSpinner("Fetching ticket...")
	s.Start()

	cmdCtx := context.Background()
	task, err := ctx.ClickUpClient.GetTask(cmdCtx, ticketID)
	if err != nil {
		s.Stop()
		return fmt.Errorf("failed to fetch ticket: %w", err)
	}

	s.Stop()
	_, _ = ui.Success.Printf("✓ Found: %s\n", task.Name)

	// Generate branch name
	branchName := utils.GenerateBranchName(ctx.Config.Git.BranchPrefix, ticketID, task.Name)

	// Get current branch
	currentBranch, err := ctx.GitRepo.CurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	// Check if we're already on this branch
	if currentBranch == branchName {
		_, _ = ui.Success.Printf("Already on branch: %s\n", branchName)
		fmt.Printf("  Ticket: %s\n", ui.Info.Sprint(task.URL))
		return nil
	}

	// Check if branch exists
	exists, err := ctx.GitRepo.BranchExists(branchName)
	if err != nil {
		return fmt.Errorf("failed to check if branch exists: %w", err)
	}

	if exists {
		var action string
		prompt := &survey.Select{
			Message: fmt.Sprintf("Branch %s already exists.", ui.Cyan.Sprint(branchName)),
			Options: []string{"Check it out", "Delete and recreate", "Cancel"},
		}
		if err := survey.AskOne(prompt, &action); err != nil {
			return err
		}

		switch action {
		case "Cancel":
			return nil
		case "Check it out":
			if err := ctx.GitRepo.Checkout(branchName); err != nil {
				return fmt.Errorf("failed to checkout branch: %w", err)
			}
			_, _ = ui.Success.Printf("✓ Checked out branch: %s\n", branchName)
			return nil
		case "Delete and recreate":
			// Delete branch using git command (go-git doesn't support delete easily)
			if err := deleteBranch(branchName); err != nil {
				return fmt.Errorf("failed to delete branch: %w", err)
			}
		}
	}

	// Create and checkout new branch
	if err := ctx.GitRepo.CreateBranch(branchName); err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}

	if err := ctx.GitRepo.Checkout(branchName); err != nil {
		return fmt.Errorf("failed to checkout branch: %w", err)
	}

	fmt.Println()
	_, _ = ui.Success.Printf("✓ Created branch: %s\n", ui.Cyan.Sprint(branchName))
	fmt.Printf("  Ticket: %s\n", ui.Info.Sprint(task.URL))

	return nil
}

func deleteBranch(branchName string) error {
	// Use git command to delete branch (go-git doesn't have easy branch delete)
	cmd := exec.Command("git", "branch", "-D", branchName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to delete branch: %w\nOutput: %s", err, string(output))
	}
	return nil
}
