package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	configJSON := `{
		"cli_path": "/usr/local/bin/copilot",
		"model": "gpt-4o",
		"working_directory": "/workspace",
		"log_level": "debug",
		"log_dir": "/tmp/logs",
		"prompt_file": "/tmp/prompt.txt",
		"available_tools": ["shell", "write", "github"],
		"streaming": true,
		"timeout": 300,
		"metrics_file": "/tmp/metrics.json"
	}`

	if err := os.WriteFile(configPath, []byte(configJSON), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if config.CLIPath != "/usr/local/bin/copilot" {
		t.Errorf("Expected CLI path '/usr/local/bin/copilot', got '%s'", config.CLIPath)
	}

	if config.Model != "gpt-4o" {
		t.Errorf("Expected model 'gpt-4o', got '%s'", config.Model)
	}

	if config.WorkingDirectory != "/workspace" {
		t.Errorf("Expected working directory '/workspace', got '%s'", config.WorkingDirectory)
	}

	if config.LogLevel != "debug" {
		t.Errorf("Expected log level 'debug', got '%s'", config.LogLevel)
	}

	if config.PromptFile != "/tmp/prompt.txt" {
		t.Errorf("Expected prompt file '/tmp/prompt.txt', got '%s'", config.PromptFile)
	}

	if len(config.AvailableTools) != 3 {
		t.Errorf("Expected 3 available tools, got %d", len(config.AvailableTools))
	}

	if !config.Streaming {
		t.Error("Expected streaming to be true")
	}

	if config.Timeout != 300 {
		t.Errorf("Expected timeout 300, got %d", config.Timeout)
	}

	if config.MetricsFile != "/tmp/metrics.json" {
		t.Errorf("Expected metrics file '/tmp/metrics.json', got '%s'", config.MetricsFile)
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	// Minimal config
	configJSON := `{
		"prompt_file": "/tmp/prompt.txt"
	}`

	if err := os.WriteFile(configPath, []byte(configJSON), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Check defaults
	if config.CLIPath != "copilot" {
		t.Errorf("Expected default CLI path 'copilot', got '%s'", config.CLIPath)
	}

	if config.LogLevel != "info" {
		t.Errorf("Expected default log level 'info', got '%s'", config.LogLevel)
	}
}

func TestLoadConfigEnvExpansion(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	// Set an env var for testing
	t.Setenv("TEST_WORKSPACE", "/test/workspace")

	configJSON := `{
		"prompt_file": "/tmp/prompt.txt",
		"working_directory": "$TEST_WORKSPACE"
	}`

	if err := os.WriteFile(configPath, []byte(configJSON), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if config.WorkingDirectory != "/test/workspace" {
		t.Errorf("Expected working directory '/test/workspace', got '%s'", config.WorkingDirectory)
	}
}

func TestLoadConfigGithubTokenFromEnv(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	t.Setenv("COPILOT_GITHUB_TOKEN", "test-token-from-env")

	configJSON := `{
		"prompt_file": "/tmp/prompt.txt"
	}`

	if err := os.WriteFile(configPath, []byte(configJSON), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if config.GithubToken != "test-token-from-env" {
		t.Errorf("Expected GitHub token from env, got '%s'", config.GithubToken)
	}
}

func TestLoadConfigMissingFile(t *testing.T) {
	_, err := LoadConfig("/nonexistent/config.json")
	if err == nil {
		t.Error("Expected error when loading non-existent config file")
	}
}

func TestLoadConfigInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	if err := os.WriteFile(configPath, []byte("{invalid json}"), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	_, err := LoadConfig(configPath)
	if err == nil {
		t.Error("Expected error when loading invalid JSON config")
	}
}
