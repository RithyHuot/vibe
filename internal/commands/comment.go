package commands

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/rithyhuot/vibe/internal/ui"
	"github.com/rithyhuot/vibe/internal/utils"
)

// NewCommentCommand creates the comment command
func NewCommentCommand(ctx *CommandContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "comment <text>",
		Short: "Add a comment to the current ticket",
		Long: `Adds a comment to the ticket associated with the current branch. Can read from stdin if no text is provided.

Examples:
  vibe comment "Fixed the authentication bug"           # Add inline comment
  vibe comment "Updated dependencies" "Added tests"     # Multi-part comment
  echo "Deployment completed" | vibe comment            # Read from stdin
  cat notes.txt | vibe comment                          # Read from file via pipe`,
		RunE: func(_ *cobra.Command, args []string) error {
			var commentText string

			if len(args) > 0 {
				// Use args as comment text
				commentText = strings.Join(args, " ")
			} else {
				// Check if stdin has data
				stat, err := os.Stdin.Stat()
				if err != nil {
					return fmt.Errorf("failed to check stdin: %w", err)
				}

				if (stat.Mode() & os.ModeCharDevice) == 0 {
					// Reading from pipe/redirect
					reader := bufio.NewReader(os.Stdin)
					var lines []string
					for {
						line, err := reader.ReadString('\n')
						if err != nil {
							if err == io.EOF {
								if line != "" {
									lines = append(lines, line)
								}
								break
							}
							return fmt.Errorf("failed to read stdin: %w", err)
						}
						lines = append(lines, strings.TrimRight(line, "\n"))
					}
					commentText = strings.Join(lines, "\n")
				} else {
					return fmt.Errorf("no comment text provided. Usage: vibe comment <text> or echo <text> | vibe comment")
				}
			}

			if strings.TrimSpace(commentText) == "" {
				return fmt.Errorf("comment text cannot be empty")
			}

			return runComment(ctx, commentText)
		},
	}

	return cmd
}

func runComment(ctx *CommandContext, commentText string) error {
	// Get current branch
	currentBranch, err := ctx.GitRepo.CurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	// Extract ticket ID from branch
	ticketID, err := utils.ExtractTicketID(currentBranch)
	if err != nil {
		return fmt.Errorf("could not extract ticket ID from branch '%s': %w", currentBranch, err)
	}

	cmdCtx := context.Background()

	// Create spinner
	s := ui.CreateSpinner("Adding comment...")
	s.Start()

	// Add comment
	_, err = ctx.ClickUpClient.AddComment(cmdCtx, ticketID, commentText)
	s.Stop()

	if err != nil {
		return fmt.Errorf("failed to add comment: %w", err)
	}

	ui.ShowSuccess("Comment added successfully!")

	return nil
}
