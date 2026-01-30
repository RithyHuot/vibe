package commands

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/rithyhuot/vibe/internal/config"
	"github.com/rithyhuot/vibe/internal/skills"
	"github.com/spf13/cobra"
)

// NewInitCommand creates the init command
func NewInitCommand(skillsFS embed.FS) *cobra.Command {
	var installSkillsFlag bool
	var forceFlag bool
	var localFlag bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize vibe configuration",
		Long: `Creates a configuration file at ~/.config/vibe/config.yaml with example values.

Use --local to create a .vibe.yaml file in the current directory to override
settings for this project only.

Examples:
  vibe init                      # Create config file (prompts if exists)
  vibe init --force              # Overwrite existing config
  vibe init --local              # Create local .vibe.yaml override file
  vibe init --install-skills     # Create config and install Claude Code skills`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(cmd, args, installSkillsFlag, forceFlag, localFlag, skillsFS)
		},
	}

	cmd.Flags().BoolVar(&installSkillsFlag, "install-skills", false, "Install Claude Code skills to ~/.claude/skills/ for global availability")
	cmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Overwrite existing configuration file")
	cmd.Flags().BoolVarP(&localFlag, "local", "l", false, "Create local .vibe.yaml override file in current directory")

	return cmd
}

func runInit(_ *cobra.Command, _ []string, installSkillsFlag bool, forceFlag bool, localFlag bool, skillsFS embed.FS) error {
	// Handle local config file creation
	if localFlag {
		return runInitLocal(forceFlag)
	}

	configPath, err := config.GetConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	// Check if config file already exists
	configExists := false
	if _, err := os.Stat(configPath); err == nil {
		configExists = true
	}

	// Determine if we should create/overwrite the config file
	shouldCreateConfig := !configExists || forceFlag

	// Handle existing config file
	if configExists && !forceFlag {
		cyan := color.New(color.FgCyan)
		yellow := color.New(color.FgYellow)

		fmt.Println()
		_, _ = cyan.Println("✓ Configuration file already exists")
		_, _ = yellow.Printf("Location: %s\n", configPath)

		// Ask if user wants to overwrite (only in interactive mode)
		if isTerminal() {
			fmt.Println()
			var overwrite bool
			prompt := &survey.Confirm{
				Message: "Do you want to overwrite it?",
				Default: false,
			}

			if err := survey.AskOne(prompt, &overwrite); err != nil {
				return fmt.Errorf("prompt cancelled")
			}

			shouldCreateConfig = overwrite

			if !overwrite {
				fmt.Println()
				fmt.Println("Skipping configuration file creation.")
				fmt.Println("Use 'vibe init --force' to overwrite without prompting.")
			}
		} else {
			// Non-interactive mode without --force flag - just skip config creation
			fmt.Println()
			fmt.Println("Skipping configuration file creation.")
			fmt.Println("Use 'vibe init --force' to overwrite the existing configuration.")
		}
	}

	// Create config file if needed
	if shouldCreateConfig {
		err = config.CreateConfigFile(configPath, true) // Always force since we've already handled the checks
		if err != nil {
			return err
		}

		// Print success message
		green := color.New(color.FgGreen, color.Bold)
		cyan := color.New(color.FgCyan)

		_, _ = green.Println("✓ Configuration file created successfully!")
		fmt.Println()
		_, _ = cyan.Printf("Location: %s\n", configPath)
		fmt.Println()
		fmt.Println("Next steps:")
		fmt.Println("1. Edit the config file with your API tokens and settings")
		fmt.Println("2. Get your ClickUp API token from: https://app.clickup.com/settings/apps")
		fmt.Println("3. Get your GitHub token from: https://github.com/settings/tokens")
		fmt.Println("4. Run 'vibe <ticket-id>' to start working on a ticket")
		fmt.Println()
		_, _ = cyan.Println("Tip: Create a .vibe.yaml file in your project to override settings")
		fmt.Println("     Run 'vibe init --local' to generate an example override file")
	}

	// Install Claude Code skills if requested
	if installSkillsFlag {
		fmt.Println()
		fmt.Println("Installing Claude Code skills...")

		if err := skills.Install(skillsFS); err != nil {
			return fmt.Errorf("failed to install skills: %w", err)
		}

		skills.PrintInstallSuccess()
	}

	// Offer to set up shell completion
	fmt.Println()
	if err := offerShellCompletion(); err != nil {
		// Don't fail init if completion setup has an issue
		yellow := color.New(color.FgYellow)
		_, _ = yellow.Printf("⚠ Shell completion setup: %v\n", err)
	}

	return nil
}

func runInitLocal(forceFlag bool) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	localConfigPath := filepath.Join(cwd, ".vibe.yaml")

	// Check if local config file already exists
	localConfigExists := false
	if _, err := os.Stat(localConfigPath); err == nil {
		localConfigExists = true
	}

	// Determine if we should create/overwrite the local config file
	shouldCreateConfig := !localConfigExists || forceFlag

	// Handle existing local config file
	if localConfigExists && !forceFlag {
		cyan := color.New(color.FgCyan)
		yellow := color.New(color.FgYellow)

		fmt.Println()
		_, _ = cyan.Println("✓ Local configuration file already exists")
		_, _ = yellow.Printf("Location: %s\n", localConfigPath)

		// Ask if user wants to overwrite (only in interactive mode)
		if isTerminal() {
			fmt.Println()
			var overwrite bool
			prompt := &survey.Confirm{
				Message: "Do you want to overwrite it?",
				Default: false,
			}

			if err := survey.AskOne(prompt, &overwrite); err != nil {
				return fmt.Errorf("prompt cancelled")
			}

			shouldCreateConfig = overwrite

			if !overwrite {
				fmt.Println()
				fmt.Println("Skipping local configuration file creation.")
				fmt.Println("Use 'vibe init --local --force' to overwrite without prompting.")
			}
		} else {
			// Non-interactive mode without --force flag - just skip config creation
			fmt.Println()
			fmt.Println("Skipping local configuration file creation.")
			fmt.Println("Use 'vibe init --local --force' to overwrite the existing configuration.")
		}
	}

	// Create local config file if needed
	if shouldCreateConfig {
		err = config.CreateLocalConfigFile(localConfigPath, true)
		if err != nil {
			return err
		}

		// Print success message
		green := color.New(color.FgGreen, color.Bold)
		cyan := color.New(color.FgCyan)

		_, _ = green.Println("✓ Local configuration override file created successfully!")
		fmt.Println()
		_, _ = cyan.Printf("Location: %s\n", localConfigPath)
		fmt.Println()
		fmt.Println("This file will override settings from ~/.config/vibe/config.yaml")
		fmt.Println("You only need to include the fields you want to override.")
		fmt.Println()
		fmt.Println("Next steps:")
		fmt.Println("1. Edit .vibe.yaml with the settings you want to override")
		fmt.Println("2. The changes will apply automatically when running vibe commands in this directory")
	}

	return nil
}

func offerShellCompletion() error {
	// Skip if not interactive (stdout is not a terminal)
	if !isTerminal() {
		return nil
	}

	var setupCompletion bool
	prompt := &survey.Confirm{
		Message: "Set up shell autocomplete?",
		Default: true,
	}

	if err := survey.AskOne(prompt, &setupCompletion); err != nil {
		return nil // User cancelled, not an error
	}

	if !setupCompletion {
		fmt.Println()
		fmt.Println("You can set up autocomplete later with:")
		fmt.Println("  vibe completion --help")
		return nil
	}

	// Detect shell and provide instructions
	shell := detectShell()
	printCompletionInstructions(shell)

	return nil
}

func isTerminal() bool {
	fileInfo, _ := os.Stdout.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

func detectShell() string {
	// Try SHELL environment variable first
	if shell := os.Getenv("SHELL"); shell != "" {
		// Extract just the shell name (e.g., "/bin/zsh" -> "zsh")
		return filepath.Base(shell)
	}

	// Default based on OS
	if runtime.GOOS == "windows" {
		return "powershell"
	}
	if runtime.GOOS == "darwin" {
		return "zsh" // macOS default since Catalina
	}
	return "bash"
}

func printCompletionInstructions(shell string) {
	cyan := color.New(color.FgCyan, color.Bold)
	yellow := color.New(color.FgYellow)

	fmt.Println()
	_, _ = cyan.Printf("Shell Completion Setup (%s)\n", shell)
	fmt.Println()

	switch shell {
	case "zsh":
		fmt.Println("Run these commands:")
		_, _ = yellow.Println("  mkdir -p ~/.zsh/completions")
		_, _ = yellow.Println("  vibe completion zsh > ~/.zsh/completions/_vibe")
		fmt.Println()
		fmt.Println("Then add to your ~/.zshrc (if not already present):")
		_, _ = yellow.Println("  fpath=(~/.zsh/completions $fpath)")
		_, _ = yellow.Println("  autoload -U compinit && compinit")
		fmt.Println()
		fmt.Println("Finally, restart your shell:")
		_, _ = yellow.Println("  exec zsh")
	case "bash":
		fmt.Println("Run this command:")
		if runtime.GOOS == "darwin" {
			_, _ = yellow.Println("  vibe completion bash > $(brew --prefix)/etc/bash_completion.d/vibe")
		} else {
			_, _ = yellow.Println("  vibe completion bash > /etc/bash_completion.d/vibe")
		}
		fmt.Println()
		fmt.Println("Then restart your terminal")
	case "fish":
		fmt.Println("Run this command:")
		_, _ = yellow.Println("  vibe completion fish > ~/.config/fish/completions/vibe.fish")
	case "powershell":
		fmt.Println("Run this command:")
		_, _ = yellow.Println("  vibe completion powershell > vibe.ps1")
		fmt.Println()
		fmt.Println("Then source this file from your PowerShell profile")
	default:
		fmt.Println("Run:")
		_, _ = yellow.Printf("  vibe completion %s --help\n", shell)
	}

	fmt.Println()
	if shell == "zsh" {
		fmt.Println("Or for a quick test without installation:")
		_, _ = yellow.Println("  autoload -U compinit && compinit")
		_, _ = yellow.Println("  source <(vibe completion zsh)")
	} else {
		fmt.Println("Or for a quick test without installation:")
		_, _ = yellow.Printf("  source <(vibe completion %s)\n", shell)
	}
}
