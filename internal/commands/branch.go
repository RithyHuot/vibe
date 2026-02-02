package commands

import (
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/rithyhuot/vibe/internal/ui"
	"github.com/rithyhuot/vibe/internal/utils"
	"github.com/spf13/cobra"
)

// NewBranchCommand creates the branch command
func NewBranchCommand(ctx *CommandContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "branch [ticket-id]",
		Short: "Create and checkout a new branch",
		Long: `Create and checkout a new branch with or without a ticket ID.

Without a ticket ID, you'll be prompted for a branch description.
This command creates simple branches without ClickUp integration.

Branch formats:
  - With ticket ID:   username/ticketid
  - With description: username/description

Examples:
  vibe branch abc123xyz          # Create branch with ticket ID
  vibe branch                    # Interactive: prompts for description`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			ticketID := ""
			if len(args) > 0 {
				ticketID = args[0]
			}
			return runBranch(ctx, ticketID)
		},
	}

	return cmd
}

func runBranch(ctx *CommandContext, ticketID string) error {
	// Get username from git config
	username, err := utils.GetGitUsername()
	if err != nil {
		_, _ = ui.Error.Println("\nGit username not configured.")
		_, _ = ui.Dim.Println("Run: git config user.name \"Your Name\"")
		return fmt.Errorf("failed to get git username: %w", err)
	}

	var branchName string

	if ticketID != "" {
		// Validate ticket ID format
		if !utils.IsTicketID(ticketID) {
			return fmt.Errorf("invalid ticket ID format: '%s'\n\nTicket ID must be exactly 9 alphanumeric characters (e.g., abc123xyz)", ticketID)
		}

		// Create branch name with ticket ID using GenerateBranchName
		// Pass empty string for prefix and title to get format: username/ticketid
		branchName = utils.GenerateBranchName("", ticketID, "", username)
	} else {
		// Interactive mode: prompt for branch description
		var description string
		prompt := &survey.Input{
			Message: "Branch description:",
		}
		if err := survey.AskOne(prompt, &description); err != nil {
			return err
		}

		description = strings.TrimSpace(description)
		if description == "" {
			return fmt.Errorf("branch description cannot be empty")
		}

		// Sanitize description
		sanitizedDesc := utils.SanitizeBranchPart(description)
		if sanitizedDesc == "" {
			return fmt.Errorf("branch description contains only invalid characters")
		}

		// Create branch name with description using GenerateBranchName
		// Pass sanitized description as ticketID to get format: username/description
		branchName = utils.GenerateBranchName("", sanitizedDesc, "", username)
	}

	// Validate the final branch name
	if err := utils.ValidateBranchName(branchName); err != nil {
		return fmt.Errorf("invalid branch name: %w", err)
	}

	return createOrCheckoutBranch(ctx, branchName)
}

// createOrCheckoutBranch handles branch creation/checkout with uncommitted changes check
func createOrCheckoutBranch(ctx *CommandContext, branchName string) error {
	// Check if branch already exists
	exists, err := ctx.GitRepo.BranchExists(branchName)
	if err != nil {
		return fmt.Errorf("failed to check if branch exists: %w", err)
	}

	if exists {
		var action string
		prompt := &survey.Select{
			Message: fmt.Sprintf("Branch %s already exists.", ui.Cyan.Sprint(branchName)),
			Options: []string{actionCheckItOut, actionCancel},
		}
		if err := survey.AskOne(prompt, &action); err != nil {
			return err
		}

		if action == actionCancel {
			return nil
		}

		// Check for uncommitted changes before checkout
		if err := handleUncommittedChanges(ctx); err != nil {
			return err
		}

		// Check it out
		if err := ctx.GitRepo.Checkout(branchName); err != nil {
			return fmt.Errorf("failed to checkout branch: %w", err)
		}

		_, _ = ui.Success.Printf("✓ Checked out existing branch: %s\n", ui.Cyan.Sprint(branchName))
		return nil
	}

	// Check for uncommitted changes before any branch operations
	if err := handleUncommittedChanges(ctx); err != nil {
		return err
	}

	// Create branch reference and checkout
	if err := ctx.GitRepo.CreateBranch(branchName); err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}

	if err := ctx.GitRepo.Checkout(branchName); err != nil {
		// Branch was created but checkout failed - offer cleanup
		_, _ = ui.Warning.Printf("Branch %s was created but checkout failed\n", branchName)
		return fmt.Errorf("checkout failed: %w\nManual cleanup: git branch -d %s", err, branchName)
	}

	_, _ = ui.Success.Printf("✓ Created and checked out branch: %s\n", ui.Cyan.Sprint(branchName))

	return nil
}
