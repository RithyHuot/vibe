package commands

import (
	"embed"
	"fmt"

	"github.com/fatih/color"
	"github.com/rithyhuot/vibe/internal/skills"
	"github.com/spf13/cobra"
)

// NewSkillsCommand creates the skills command
func NewSkillsCommand(skillsFS embed.FS) *cobra.Command {
	var uninstall bool

	cmd := &cobra.Command{
		Use:   "skills",
		Short: "Manage Claude Code skills",
		Long: `Install, update, or uninstall vibe skills for Claude Code.

Claude Code skills enable seamless integration with vibe commands directly from the Claude Code CLI.
Skills are installed to ~/.claude/skills/ for global availability.

Examples:
  vibe skills                    # Install or update skills
  vibe skills --uninstall        # Remove skills from Claude Code
  vibe init --install-skills     # Install during initialization`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if uninstall {
				return runSkillsUninstall(cmd, args)
			}
			return runSkillsInstall(cmd, args, skillsFS)
		},
	}

	cmd.Flags().BoolVar(&uninstall, "uninstall", false, "Remove skills from ~/.claude/skills/")

	return cmd
}

func runSkillsInstall(_ *cobra.Command, _ []string, skillsFS embed.FS) error {
	yellow := color.New(color.FgYellow, color.Bold)

	_, _ = yellow.Println("Installing Claude Code skills...")
	fmt.Println()

	if err := skills.Install(skillsFS); err != nil {
		return fmt.Errorf("failed to install skills: %w", err)
	}

	skills.PrintInstallSuccess()

	return nil
}

func runSkillsUninstall(_ *cobra.Command, _ []string) error {
	yellow := color.New(color.FgYellow, color.Bold)

	_, _ = yellow.Println("Uninstalling Claude Code skills...")
	fmt.Println()

	if err := skills.Uninstall(); err != nil {
		return fmt.Errorf("failed to uninstall skills: %w", err)
	}

	skills.PrintUninstallSuccess()

	return nil
}
