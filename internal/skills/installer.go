// Package skills manages Claude AI skill installation and management.
package skills

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/fatih/color"
)

// SkillsFS holds the embedded skills directory
var SkillsFS embed.FS

// GetClaudeSkillsPath returns the path to the Claude Code skills directory
func GetClaudeSkillsPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".claude", "skills"), nil
}

// Install copies the embedded skills to the Claude Code skills directory
func Install(skillsFS embed.FS) error {
	destPath, err := GetClaudeSkillsPath()
	if err != nil {
		return fmt.Errorf("failed to get Claude skills path: %w", err)
	}

	// Create the destination directory
	if err := os.MkdirAll(destPath, 0755); err != nil {
		return fmt.Errorf("failed to create Claude skills directory: %w", err)
	}

	// Walk through the embedded skills directory
	// The embed path starts with "skills" since we're embedding "skills"
	err = fs.WalkDir(skillsFS, "skills", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip the root "skills" directory entry
		if path == "skills" {
			return nil
		}

		// Remove the "skills/" prefix from the path
		relPath := path[7:] // len("skills/") = 7

		// Build the destination path
		destFilePath := filepath.Join(destPath, relPath)

		if d.IsDir() {
			// Create directory
			return os.MkdirAll(destFilePath, 0755)
		}

		// Read file content from embedded FS
		content, err := fs.ReadFile(skillsFS, path)
		if err != nil {
			return fmt.Errorf("failed to read embedded file %s: %w", path, err)
		}

		// Write file to destination
		if err := os.WriteFile(destFilePath, content, 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", destFilePath, err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to install skills: %w", err)
	}

	return nil
}

// Uninstall removes the skills from the Claude Code skills directory
func Uninstall() error {
	skillsPath, err := GetClaudeSkillsPath()
	if err != nil {
		return fmt.Errorf("failed to get Claude skills path: %w", err)
	}

	// Check if skills directory exists
	if _, err := os.Stat(skillsPath); os.IsNotExist(err) {
		return fmt.Errorf("skills are not installed at %s", skillsPath)
	}

	// List of vibe skill directories to remove
	skillDirs := []string{
		"vibe",
		"vibe-ticket",
		"vibe-comment",
		"vibe-pr",
		"vibe-pr-status",
		"vibe-pr-update",
		"vibe-merge",
		"vibe-ci-status",
		"vibe-issues",
		"vibe-issue",
		"vibe-issue-create",
		"vibe-issue-update",
		"vibe-code-review",
	}

	// Remove each skill directory
	for _, skillDir := range skillDirs {
		skillPath := filepath.Join(skillsPath, skillDir)
		if err := os.RemoveAll(skillPath); err != nil {
			return fmt.Errorf("failed to remove skill %s: %w", skillDir, err)
		}
	}

	return nil
}

// PrintInstallSuccess prints a success message after installing skills
func PrintInstallSuccess() {
	green := color.New(color.FgGreen, color.Bold)
	cyan := color.New(color.FgCyan)
	yellow := color.New(color.FgYellow)

	skillsPath, _ := GetClaudeSkillsPath()

	fmt.Println()
	_, _ = green.Println("✓ Claude Code skills installed successfully!")
	fmt.Println()
	_, _ = cyan.Printf("Location: %s\n", skillsPath)
	fmt.Println()
	_, _ = yellow.Println("Available skills:")
	fmt.Println("  - vibe                    Start work on a ClickUp ticket")
	fmt.Println("  - vibe-ticket             Get context on current ticket")
	fmt.Println("  - vibe-comment            Add comment to ticket")
	fmt.Println("  - vibe-pr                 Create a pull request")
	fmt.Println("  - vibe-pr-status          Check PR status")
	fmt.Println("  - vibe-pr-update          Update PR description")
	fmt.Println("  - vibe-merge              Merge a pull request")
	fmt.Println("  - vibe-ci-status          Check CircleCI status")
	fmt.Println("  - vibe-issues             List GitHub issues")
	fmt.Println("  - vibe-issue              View issue details")
	fmt.Println("  - vibe-issue-create       Create a new issue")
	fmt.Println("  - vibe-issue-update       Update existing issue")
	fmt.Println("  - vibe-code-review        Perform comprehensive code review")
	fmt.Println()
	fmt.Println("These skills are now available in ALL your projects when using Claude Code!")
}

// PrintUninstallSuccess prints a success message after uninstalling skills
func PrintUninstallSuccess() {
	green := color.New(color.FgGreen, color.Bold)
	cyan := color.New(color.FgCyan)

	skillsPath, _ := GetClaudeSkillsPath()

	fmt.Println()
	_, _ = green.Println("✓ Claude Code skills uninstalled successfully!")
	fmt.Println()
	_, _ = cyan.Printf("Removed from: %s\n", skillsPath)
	fmt.Println()
	fmt.Println("You can reinstall anytime with: vibe skills")
}
