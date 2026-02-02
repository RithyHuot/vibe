package commands

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/rithyhuot/vibe/internal/config"
	"github.com/rithyhuot/vibe/internal/services/claude"
	"github.com/rithyhuot/vibe/internal/services/clickup"
	"github.com/rithyhuot/vibe/internal/services/git"
	"github.com/rithyhuot/vibe/internal/services/github"
	"github.com/rithyhuot/vibe/internal/ui"
	"github.com/spf13/cobra"
)

const (
	// User action constants for prompts
	actionStashChanges    = "Stash changes"
	actionCheckItOut      = "Check it out"
	actionDeleteRecreate  = "Delete and recreate"
	actionCancel          = "Cancel"

	// gitStashTimeout is the maximum time to wait for git stash operation
	// 30s should be sufficient for most repos; very large repos may need adjustment
	gitStashTimeout = 30 * time.Second
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

// CommandContextKey is the key used to store CommandContext in cobra's context
const CommandContextKey contextKey = "commandContext"

// For internal backwards compatibility
const commandContextKey = CommandContextKey

// getCommandContext retrieves the CommandContext from a cobra command's context
func getCommandContext(cmd *cobra.Command, currentCtx *CommandContext) *CommandContext {
	if ctxVal := cmd.Context().Value(commandContextKey); ctxVal != nil {
		return ctxVal.(*CommandContext)
	}
	return currentCtx
}

// CommandContext holds shared dependencies for all commands
type CommandContext struct {
	Config        *config.Config
	ClickUpClient clickup.Client
	GitHubClient  github.Client
	GitRepo       git.Repository
	ClaudeClient  claude.Client
}

// NewCommandContext creates a new command context
func NewCommandContext(cfg *config.Config) (*CommandContext, error) {
	// Initialize ClickUp client
	clickUpClient := clickup.NewClient(cfg.ClickUp.APIToken)

	// Initialize GitHub client with mode support
	githubClient, err := github.NewClientWithMode(
		cfg.GitHub.Mode,
		cfg.GitHub.Token,
		cfg.GitHub.Owner,
		cfg.GitHub.Repo,
	)
	if err != nil {
		return nil, err
	}

	// Initialize Git repository
	gitRepo, err := git.OpenRepository(".")
	if err != nil {
		return nil, err
	}

	// Initialize Claude client (auto-detect CLI or API)
	var claudeClient claude.Client
	if cfg.AI.Enabled {
		claudeClient = claude.NewClientAuto(cfg.Claude.APIKey)
	}

	return &CommandContext{
		Config:        cfg,
		ClickUpClient: clickUpClient,
		GitHubClient:  githubClient,
		GitRepo:       gitRepo,
		ClaudeClient:  claudeClient,
	}, nil
}

// handleUncommittedChanges checks for uncommitted changes and prompts user to stash if needed
// Returns nil if it's safe to proceed with checkout, error otherwise
func handleUncommittedChanges(ctx *CommandContext) error {
	// Check for uncommitted changes
	status, err := ctx.GitRepo.Status()
	if err != nil {
		return fmt.Errorf("failed to check git status: %w", err)
	}

	// If no changes, proceed
	if len(status) == 0 {
		return nil
	}

	// Count modified/added/deleted files (exclude untracked)
	// Note: map iteration order is undefined, but order doesn't matter for counting
	changedFiles := 0
	untrackedFiles := 0
	for _, fileStatus := range status {
		if fileStatus == "untracked" {
			untrackedFiles++
		} else {
			changedFiles++
		}
	}

	// If only untracked files, proceed (but notify user)
	if changedFiles == 0 {
		if untrackedFiles > 0 {
			_, _ = ui.Info.Printf("Note: %d untracked file(s) will remain in your working directory\n", untrackedFiles)
		}
		return nil
	}

	// Prompt user to stash changes
	_, _ = ui.Warning.Printf("\n⚠️  You have %d uncommitted change(s)\n", changedFiles)
	if untrackedFiles > 0 {
		_, _ = ui.Dim.Printf("   (%d untracked file(s) will not be stashed)\n", untrackedFiles)
	}
	fmt.Println()

	var action string
	prompt := &survey.Select{
		Message: "What would you like to do?",
		Options: []string{actionStashChanges, actionCancel},
		Default: actionStashChanges,
	}
	if err := survey.AskOne(prompt, &action); err != nil {
		return err
	}

	if action == actionCancel {
		return fmt.Errorf("checkout cancelled: uncommitted changes present")
	}

	// Stash the changes with timeout
	_, _ = ui.Info.Println("Stashing uncommitted changes...")
	cmdCtx, cancel := context.WithTimeout(context.Background(), gitStashTimeout)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, "git", "stash", "push", "-m", "Auto-stash by vibe CLI")
	output, err := cmd.CombinedOutput()
	if err != nil {
		if cmdCtx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("stash operation timed out after %v", gitStashTimeout)
		}
		outputStr := strings.TrimSpace(string(output))
		if len(outputStr) > 500 {
			outputStr = outputStr[:500] + "... (truncated)"
		}
		return fmt.Errorf("failed to stash changes: %w\nOutput:\n%s", err, outputStr)
	}

	_, _ = ui.Success.Printf("✓ Changes stashed successfully\n")
	_, _ = ui.Dim.Println("  (Use 'git stash pop' to restore them later)")

	return nil
}
