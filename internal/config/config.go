// Package config provides configuration management for the vibe CLI.
package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// ExampleConfig returns an example configuration as YAML
func ExampleConfig() string {
	return `# Vibe CLI Configuration
#
# This is the global config file (~/.config/vibe/config.yaml)
# You can override settings per-project by creating a .vibe.yaml file
# in your project directory with only the fields you want to override.
#
# Priority: CLI flags > local .vibe.yaml > global config > env vars

# ClickUp configuration
clickup:
  api_token: "pk_your_clickup_api_token"
  user_id: "12345678"
  workspace_id: "1234567"
  team_id: "1234567"

# GitHub configuration
github:
  token: "ghp_your_github_token"  # Optional if using CLI mode
  username: "your-username"
  owner: "org-name"
  repo: "repo-name"
  mode: "auto"  # Options: "api", "cli", or "auto" (default: auto-detect)

# Git configuration
git:
  branch_prefix: "your-username"
  base_branch: "main"

# CircleCI configuration (optional)
circleci:
  api_token: "circle_your_circleci_token"

# Claude AI configuration (optional)
claude:
  api_key: "sk-ant-your_claude_api_key"

# Workspace configuration
workspaces:
  - name: "Engineering"
    folder_id: "123456789"
    sprint_patterns:
      - "Sprint \\d+ \\("

# Default values
defaults:
  # Status to automatically set when starting work on a ticket
  # Must be a valid status name that exists in your ClickUp space
  # Common values: "doing", "on deck", "backlog", "prioritized", "in code review"
  # To find: Check any ticket in ClickUp to see available status names
  # Comment out this line to disable automatic status updates
  status: "doing"

# AI features
ai:
  enabled: true
  generate_descriptions: true

# UI preferences
ui:
  color_enabled: true
`
}

// ExampleLocalConfig returns an example local override configuration as YAML
func ExampleLocalConfig() string {
	return `# Local Vibe Configuration Override (.vibe.yaml)
#
# This file overrides settings from ~/.config/vibe/config.yaml for this project only.
# You only need to include the fields you want to override - not the entire config.
#
# Example: Override GitHub repo for this project
github:
  owner: "different-org"
  repo: "different-repo"

# Example: Override git branch prefix
git:
  branch_prefix: "feature"

# Example: Use CLI mode for GitHub in this project
# github:
#   mode: "cli"

# Example: Override workspace for this project
# workspaces:
#   - name: "Engineering"
#     folder_id: "123456789"
#     sprint_patterns:
#       - "Sprint \\d+ \\("
`
}

// CreateConfigFile creates a new config file with example content
func CreateConfigFile(path string, force bool) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Check if file already exists (only if not forcing)
	if !force {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("config file already exists at %s", path)
		}
	}

	// Write example config
	if err := os.WriteFile(path, []byte(ExampleConfig()), 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// CreateLocalConfigFile creates a local .vibe.yaml override file
func CreateLocalConfigFile(path string, force bool) error {
	// Check if file already exists (only if not forcing)
	if !force {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("local config file already exists at %s", path)
		}
	}

	// Write example local config
	if err := os.WriteFile(path, []byte(ExampleLocalConfig()), 0600); err != nil {
		return fmt.Errorf("failed to write local config file: %w", err)
	}

	return nil
}
