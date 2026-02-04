// Package main provides the vibe CLI application entry point.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	vibe "github.com/rithyhuot/vibe"
	"github.com/rithyhuot/vibe/internal/commands"
	"github.com/rithyhuot/vibe/internal/config"
)

// commandContextKey is defined in the commands package to ensure type consistency
var commandContextKey = commands.CommandContextKey

var (
	// Version is set during build
	Version = "dev"
	// BuildTime is set during build
	BuildTime = "unknown"

	// Global flags
	configFile string
)

func main() {
	if err := run(); err != nil {
		red := color.New(color.FgRed, color.Bold)
		_, _ = red.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	rootCmd := &cobra.Command{
		Use:   "vibe",
		Short: "A CLI tool to streamline developer workflow with ClickUp, GitHub, and CircleCI",
		Long: `vibe is a production-quality CLI tool that integrates ClickUp (project management),
GitHub (code repository), and CircleCI (CI/CD) to streamline developer workflow
from ticket assignment to PR merge.`,
		Version:      fmt.Sprintf("%s (built: %s)", Version, BuildTime),
		SilenceUsage: true,
	}

	// Global flags
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file (default is ~/.config/vibe/config.yaml)")

	// Add commands that don't require config
	rootCmd.AddCommand(commands.NewInitCommand(vibe.SkillsFS))
	rootCmd.AddCommand(commands.NewSkillsCommand(vibe.SkillsFS))
	rootCmd.AddCommand(newCompletionCommand())

	// Add config-dependent commands
	addConfigDependentCommands(rootCmd)

	return rootCmd.Execute()
}

func newCompletionCommand() *cobra.Command {
	completionCmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion script",
		Long: `Generate shell completion script for vibe.

The completion script can be generated for bash, zsh, fish, and powershell shells.

To load completions:

Bash (requires bash-completion package):

  $ source <(vibe completion bash)

  # To load completions for each session, execute once:
  # Linux (requires sudo):
  $ sudo vibe completion bash > /etc/bash_completion.d/vibe

  # macOS with Homebrew (requires: brew install bash-completion):
  $ vibe completion bash > $(brew --prefix)/etc/bash_completion.d/vibe

  # Or install to user directory (no sudo required):
  $ mkdir -p ~/.bash_completion.d
  $ vibe completion bash > ~/.bash_completion.d/vibe
  $ echo 'source ~/.bash_completion.d/vibe' >> ~/.bashrc

Zsh:

  # Create a directory for custom completions and generate the completion file:
  $ mkdir -p ~/.zsh/completions
  $ vibe completion zsh > ~/.zsh/completions/_vibe

  # Add to ~/.zshrc (if not already present):
  $ echo 'fpath=(~/.zsh/completions $fpath)' >> ~/.zshrc
  $ echo 'autoload -U compinit && compinit' >> ~/.zshrc

  # Restart your shell:
  $ exec zsh

Fish:

  $ vibe completion fish | source

  # To load completions for each session, execute once:
  $ vibe completion fish > ~/.config/fish/completions/vibe.fish

PowerShell:

  PS> vibe completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> vibe completion powershell > vibe.ps1
  # and source this file from your PowerShell profile.
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				return cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				return cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			}
			return nil
		},
	}

	return completionCmd
}

//nolint:gocyclo // Complexity is acceptable for command registration
func addConfigDependentCommands(rootCmd *cobra.Command) {
	// Lazy load config and context
	var cmdCtx *commands.CommandContext

	getContext := func() (*commands.CommandContext, error) {
		if cmdCtx != nil {
			return cmdCtx, nil
		}

		// Load config
		cfg, err := config.Load(configFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load configuration: %w\n\nRun 'vibe init' to create a configuration file", err)
		}

		// Create command context
		cmdCtx, err = commands.NewCommandContext(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize: %w", err)
		}

		return cmdCtx, nil
	}

	// Ticket command
	ticketCmd := &cobra.Command{
		Use:   "ticket [ticket-id]",
		Short: "View ticket details",
		Long: `Displays detailed information about a ticket. If no ticket ID is provided, uses the current branch.

Examples:
  vibe ticket                    # View ticket for current branch
  vibe ticket abc123             # View specific ticket by ID
  vibe ticket 86b7x5453          # View ticket with full ClickUp ID`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			ctx, err := getContext()
			if err != nil {
				return err
			}
			ticketActual := commands.NewTicketCommand(ctx)
			return ticketActual.RunE(nil, args)
		},
	}

	// Workon command
	workonCmd := &cobra.Command{
		Use:   "workon <ticket-id>",
		Short: "Start working on a ticket",
		Long: `Fetches a task from ClickUp, creates a branch, and updates the task status.

Examples:
  vibe workon abc123             # Start working on ticket abc123
  vibe workon 86b7x5453          # Start working with full ClickUp ID
  vibe abc123                    # Shorthand: vibe command works the same`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			ctx, err := getContext()
			if err != nil {
				return err
			}
			vibeActual := commands.NewVibeCommand(ctx)
			return vibeActual.RunE(nil, args)
		},
	}

	// Comment command
	commentCmd := &cobra.Command{
		Use:   "comment <text>",
		Short: "Add a comment to the current ticket",
		Long: `Adds a comment to the ticket associated with the current branch. Can read from stdin if no text is provided.

Examples:
  vibe comment "Fixed the authentication bug"           # Add inline comment
  vibe comment "Updated dependencies" "Added tests"     # Multi-part comment
  echo "Deployment completed" | vibe comment            # Read from stdin
  cat notes.txt | vibe comment                          # Read from file via pipe`,
		RunE: func(_ *cobra.Command, args []string) error {
			ctx, err := getContext()
			if err != nil {
				return err
			}
			commentActual := commands.NewCommentCommand(ctx)
			return commentActual.RunE(nil, args)
		},
	}

	// Create a temporary context-less PR command to get the structure
	// The actual context will be loaded in PreRunE
	dummyCtx := &commands.CommandContext{}

	prCmd := commands.NewPRCommand(dummyCtx)
	prCmd.PreRunE = func(cmd *cobra.Command, _ []string) error {
		ctx, err := getContext()
		if err != nil {
			return err
		}
		// Store context in cobra's context so RunE can access it
		cmd.SetContext(context.WithValue(cmd.Context(), commandContextKey, ctx))
		return nil
	}

	prStatusCmd := commands.NewPRStatusCommand(dummyCtx)
	prStatusCmd.PreRunE = func(_ *cobra.Command, _ []string) error {
		ctx, err := getContext()
		if err != nil {
			return err
		}
		*prStatusCmd = *commands.NewPRStatusCommand(ctx)
		return nil
	}

	prUpdateCmd := commands.NewPRUpdateCommand(dummyCtx)
	prUpdateCmd.PreRunE = func(_ *cobra.Command, _ []string) error {
		ctx, err := getContext()
		if err != nil {
			return err
		}
		*prUpdateCmd = *commands.NewPRUpdateCommand(ctx)
		return nil
	}

	// Start command
	startCmd := commands.NewStartCommand(dummyCtx)
	startCmd.PreRunE = func(_ *cobra.Command, _ []string) error {
		ctx, err := getContext()
		if err != nil {
			return err
		}
		*startCmd = *commands.NewStartCommand(ctx)
		return nil
	}

	// Merge command
	mergeCmd := commands.NewMergeCommand(dummyCtx)
	mergeCmd.PreRunE = func(_ *cobra.Command, _ []string) error {
		ctx, err := getContext()
		if err != nil {
			return err
		}
		*mergeCmd = *commands.NewMergeCommand(ctx)
		return nil
	}

	// CI Status command
	ciStatusCmd := commands.NewCIStatusCommand(dummyCtx)
	ciStatusCmd.PreRunE = func(_ *cobra.Command, _ []string) error {
		ctx, err := getContext()
		if err != nil {
			return err
		}
		*ciStatusCmd = *commands.NewCIStatusCommand(ctx)
		return nil
	}

	// CI Failure command
	ciFailureCmd := commands.NewCIFailureCommand(dummyCtx)
	ciFailureCmd.PreRunE = func(_ *cobra.Command, _ []string) error {
		ctx, err := getContext()
		if err != nil {
			return err
		}
		*ciFailureCmd = *commands.NewCIFailureCommand(ctx)
		return nil
	}

	// Issue commands
	issuesCmd := commands.NewIssuesCommand(dummyCtx)
	issuesCmd.PreRunE = func(cmd *cobra.Command, _ []string) error {
		ctx, err := getContext()
		if err != nil {
			return err
		}
		// Store context in cobra's context so RunE can access it
		cmd.SetContext(context.WithValue(cmd.Context(), commandContextKey, ctx))
		return nil
	}

	issueCmd := commands.NewIssueCommand(dummyCtx)
	issueCmd.PreRunE = func(cmd *cobra.Command, _ []string) error {
		ctx, err := getContext()
		if err != nil {
			return err
		}
		// Store context in cobra's context so RunE can access it
		cmd.SetContext(context.WithValue(cmd.Context(), commandContextKey, ctx))
		return nil
	}

	issueCreateCmd := commands.NewIssueCreateCommand(dummyCtx)
	issueCreateCmd.PreRunE = func(cmd *cobra.Command, _ []string) error {
		ctx, err := getContext()
		if err != nil {
			return err
		}
		// Store context in cobra's context so RunE can access it
		cmd.SetContext(context.WithValue(cmd.Context(), commandContextKey, ctx))
		return nil
	}

	issueUpdateCmd := commands.NewIssueUpdateCommand(dummyCtx)
	issueUpdateCmd.PreRunE = func(cmd *cobra.Command, _ []string) error {
		ctx, err := getContext()
		if err != nil {
			return err
		}
		// Store context in cobra's context so RunE can access it
		cmd.SetContext(context.WithValue(cmd.Context(), commandContextKey, ctx))
		return nil
	}

	// Branch command
	branchCmd := commands.NewBranchCommand(dummyCtx)
	branchCmd.PreRunE = func(_ *cobra.Command, _ []string) error {
		ctx, err := getContext()
		if err != nil {
			return err
		}
		*branchCmd = *commands.NewBranchCommand(ctx)
		return nil
	}

	// Add vibeCmd handling to root so "vibe <ticket-id>" works directly
	rootCmd.Args = func(cmd *cobra.Command, args []string) error {
		// If there's exactly one arg and it's not a subcommand, treat it as a ticket ID
		if len(args) == 1 {
			// Check if it's a known subcommand
			for _, subcmd := range cmd.Commands() {
				if subcmd.Name() == args[0] {
					return nil // Let cobra handle the subcommand
				}
			}
			// It's a ticket ID, allow it
			return nil
		}
		// Otherwise, use default behavior (show help if no args)
		return nil
	}

	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		// If exactly one arg that's not a subcommand, run vibe command
		if len(args) == 1 {
			isSubcmd := false
			for _, subcmd := range cmd.Commands() {
				if subcmd.Name() == args[0] {
					isSubcmd = true
					break
				}
			}
			if !isSubcmd {
				// Run the vibe command logic
				ctx, err := getContext()
				if err != nil {
					return err
				}
				vibeActual := commands.NewVibeCommand(ctx)
				return vibeActual.RunE(nil, args)
			}
		}
		// Otherwise show help
		return cmd.Help()
	}

	rootCmd.AddCommand(workonCmd, ticketCmd, commentCmd, prCmd, prStatusCmd, prUpdateCmd, startCmd, mergeCmd, ciStatusCmd, ciFailureCmd, issuesCmd, issueCmd, issueCreateCmd, issueUpdateCmd, branchCmd)
}
