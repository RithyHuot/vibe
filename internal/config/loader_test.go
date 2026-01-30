package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadWithLocalOverride(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "vibe-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a global config file
	globalConfigPath := filepath.Join(tmpDir, "global-config.yaml")
	globalConfig := `clickup:
  api_token: "global_token"
  user_id: "12345"
  workspace_id: "67890"
  team_id: "67890"

github:
  token: "global_github_token"
  username: "global-user"
  owner: "global-org"
  repo: "global-repo"
  mode: "api"

git:
  branch_prefix: "global-prefix"
  base_branch: "main"

workspaces:
  - name: "Global Workspace"
    folder_id: "111111"
    sprint_patterns:
      - "Sprint \\d+ \\("

defaults:
  status: "In Progress"

ai:
  enabled: true
  generate_descriptions: true

ui:
  color_enabled: true
`
	if err := os.WriteFile(globalConfigPath, []byte(globalConfig), 0600); err != nil {
		t.Fatalf("Failed to write global config: %v", err)
	}

	// Create a local config directory and file
	localDir := filepath.Join(tmpDir, "project")
	if err := os.MkdirAll(localDir, 0755); err != nil {
		t.Fatalf("Failed to create local dir: %v", err)
	}

	localConfigPath := filepath.Join(localDir, ".vibe.yaml")
	localConfig := `github:
  owner: "local-org"
  repo: "local-repo"

git:
  branch_prefix: "local-prefix"

ai:
  enabled: false
`
	if err := os.WriteFile(localConfigPath, []byte(localConfig), 0600); err != nil {
		t.Fatalf("Failed to write local config: %v", err)
	}

	// Change to the local directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current dir: %v", err)
	}
	defer func() { _ = os.Chdir(originalDir) }()

	if err := os.Chdir(localDir); err != nil {
		t.Fatalf("Failed to change to local dir: %v", err)
	}

	// Load the config (should merge global + local)
	cfg, err := Load(globalConfigPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify that local overrides took effect
	if cfg.GitHub.Owner != "local-org" {
		t.Errorf("Expected GitHub owner to be 'local-org' (from local), got '%s'", cfg.GitHub.Owner)
	}

	if cfg.GitHub.Repo != "local-repo" {
		t.Errorf("Expected GitHub repo to be 'local-repo' (from local), got '%s'", cfg.GitHub.Repo)
	}

	if cfg.Git.BranchPrefix != "local-prefix" {
		t.Errorf("Expected git branch prefix to be 'local-prefix' (from local), got '%s'", cfg.Git.BranchPrefix)
	}

	if cfg.AI.Enabled != false {
		t.Errorf("Expected AI enabled to be false (from local), got %v", cfg.AI.Enabled)
	}

	// Verify that non-overridden values remain from global config
	if cfg.GitHub.Username != "global-user" {
		t.Errorf("Expected GitHub username to be 'global-user' (from global), got '%s'", cfg.GitHub.Username)
	}

	if cfg.ClickUp.APIToken != "global_token" {
		t.Errorf("Expected ClickUp token to be 'global_token' (from global), got '%s'", cfg.ClickUp.APIToken)
	}

	if cfg.Git.BaseBranch != "main" {
		t.Errorf("Expected git base branch to be 'main' (from global), got '%s'", cfg.Git.BaseBranch)
	}

	if cfg.UI.ColorEnabled != true {
		t.Errorf("Expected UI color enabled to be true (from global), got %v", cfg.UI.ColorEnabled)
	}
}

func TestLoadWithoutLocalOverride(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "vibe-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a global config file
	globalConfigPath := filepath.Join(tmpDir, "global-config.yaml")
	globalConfig := `clickup:
  api_token: "test_token"
  user_id: "12345"
  workspace_id: "67890"
  team_id: "67890"

github:
  token: "test_github_token"
  username: "test-user"
  owner: "test-org"
  repo: "test-repo"
  mode: "auto"

git:
  branch_prefix: "test-prefix"
  base_branch: "main"

workspaces:
  - name: "Test Workspace"
    folder_id: "111111"
    sprint_patterns:
      - "Sprint \\d+ \\("

defaults:
  status: "In Progress"

ai:
  enabled: true
  generate_descriptions: true

ui:
  color_enabled: true
`
	if err := os.WriteFile(globalConfigPath, []byte(globalConfig), 0600); err != nil {
		t.Fatalf("Failed to write global config: %v", err)
	}

	// Change to a directory without .vibe.yaml
	emptyDir := filepath.Join(tmpDir, "empty-project")
	if err := os.MkdirAll(emptyDir, 0755); err != nil {
		t.Fatalf("Failed to create empty dir: %v", err)
	}

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current dir: %v", err)
	}
	defer func() { _ = os.Chdir(originalDir) }()

	if err := os.Chdir(emptyDir); err != nil {
		t.Fatalf("Failed to change to empty dir: %v", err)
	}

	// Load the config (should only use global)
	cfg, err := Load(globalConfigPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify global config values
	if cfg.GitHub.Owner != "test-org" {
		t.Errorf("Expected GitHub owner to be 'test-org', got '%s'", cfg.GitHub.Owner)
	}

	if cfg.GitHub.Repo != "test-repo" {
		t.Errorf("Expected GitHub repo to be 'test-repo', got '%s'", cfg.GitHub.Repo)
	}

	if cfg.Git.BranchPrefix != "test-prefix" {
		t.Errorf("Expected git branch prefix to be 'test-prefix', got '%s'", cfg.Git.BranchPrefix)
	}

	if cfg.AI.Enabled != true {
		t.Errorf("Expected AI enabled to be true, got %v", cfg.AI.Enabled)
	}
}

func TestFindLocalConfig(t *testing.T) {
	// Test when .vibe.yaml exists
	tmpDir, err := os.MkdirTemp("", "vibe-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	localConfigPath := filepath.Join(tmpDir, ".vibe.yaml")
	if err := os.WriteFile(localConfigPath, []byte("test: true"), 0600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current dir: %v", err)
	}
	defer func() { _ = os.Chdir(originalDir) }()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change dir: %v", err)
	}

	result := findLocalConfig()
	// Normalize paths to handle symlinks (e.g., /tmp -> /private/tmp on macOS)
	expectedPath, _ := filepath.EvalSymlinks(localConfigPath)
	actualPath, _ := filepath.EvalSymlinks(result)
	if actualPath != expectedPath {
		t.Errorf("Expected to find local config at '%s', got '%s'", expectedPath, actualPath)
	}

	// Test when .vibe.yaml doesn't exist
	emptyDir := filepath.Join(tmpDir, "empty")
	if err := os.MkdirAll(emptyDir, 0755); err != nil {
		t.Fatalf("Failed to create empty dir: %v", err)
	}

	if err := os.Chdir(emptyDir); err != nil {
		t.Fatalf("Failed to change to empty dir: %v", err)
	}

	result = findLocalConfig()
	if result != "" {
		t.Errorf("Expected empty string when no local config exists, got '%s'", result)
	}
}
