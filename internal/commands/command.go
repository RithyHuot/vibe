package commands

import (
	"github.com/rithyhuot/vibe/internal/config"
	"github.com/rithyhuot/vibe/internal/services/claude"
	"github.com/rithyhuot/vibe/internal/services/clickup"
	"github.com/rithyhuot/vibe/internal/services/git"
	"github.com/rithyhuot/vibe/internal/services/github"
	"github.com/spf13/cobra"
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
