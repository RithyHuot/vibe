package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/briandowns/spinner"
	"github.com/rithyhuot/vibe/internal/utils"
	"github.com/spf13/cobra"
)

// NewTicketCommand creates the ticket command
func NewTicketCommand(ctx *CommandContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ticket [ticket-id]",
		Short: "View ticket details",
		Long: `Displays detailed information about a ticket. If no ticket ID is provided, uses the current branch.

Examples:
  vibe ticket                    # View ticket for current branch
  vibe ticket abc123             # View specific ticket by ID
  vibe ticket 86b7x5453          # View ticket with full ClickUp ID`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			var ticketID string

			if len(args) > 0 {
				ticketID = args[0]
			} else {
				// Extract from current branch
				currentBranch, err := ctx.GitRepo.CurrentBranch()
				if err != nil {
					return fmt.Errorf("failed to get current branch: %w", err)
				}

				ticketID, err = utils.ExtractTicketID(currentBranch)
				if err != nil {
					return fmt.Errorf("could not extract ticket ID from branch '%s': %w", currentBranch, err)
				}
			}

			return runTicket(ctx, ticketID)
		},
	}

	return cmd
}

func runTicket(ctx *CommandContext, ticketID string) error {
	// Validate ticket ID
	if !utils.IsTicketID(ticketID) {
		return fmt.Errorf("invalid ticket ID format: %s", ticketID)
	}

	cmdCtx := context.Background()

	// Create spinner
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Fetching task details..."
	s.Start()

	// Fetch task
	task, err := ctx.ClickUpClient.GetTask(cmdCtx, ticketID)
	if err != nil {
		s.Stop()
		return fmt.Errorf("failed to fetch task: %w", err)
	}

	s.Stop()

	// Display task
	displayTask(task)

	return nil
}
