//go:build !integration

package workflow

import (
	"strings"
	"testing"

	"github.com/github/gh-aw/pkg/constants"
)

func TestCopilotSDKEngine(t *testing.T) {
	engine := NewCopilotSDKEngine()

	// Test basic properties
	if engine.GetID() != "copilot-sdk" {
		t.Errorf("Expected copilot-sdk engine ID, got '%s'", engine.GetID())
	}

	if engine.GetDisplayName() != "GitHub Copilot SDK" {
		t.Errorf("Expected 'GitHub Copilot SDK' display name, got '%s'", engine.GetDisplayName())
	}

	if !engine.IsExperimental() {
		t.Error("Expected copilot-sdk engine to be experimental")
	}

	if !engine.SupportsToolsAllowlist() {
		t.Error("Expected copilot-sdk engine to support tools allowlist")
	}

	if !engine.SupportsHTTPTransport() {
		t.Error("Expected copilot-sdk engine to support HTTP transport")
	}

	if engine.SupportsMaxTurns() {
		t.Error("Expected copilot-sdk engine to not support max-turns yet")
	}

	if !engine.SupportsFirewall() {
		t.Error("Expected copilot-sdk engine to support firewall")
	}

	if engine.SupportsPlugins() {
		t.Error("Expected copilot-sdk engine to not support plugins")
	}

	// Test declared output files
	outputFiles := engine.GetDeclaredOutputFiles()
	if len(outputFiles) != 1 {
		t.Errorf("Expected 1 declared output file, got %d", len(outputFiles))
	}

	if outputFiles[0] != "/tmp/gh-aw/sandbox/agent/logs/" {
		t.Errorf("Expected declared output file to be logs folder, got %s", outputFiles[0])
	}
}

func TestCopilotSDKEngineDefaultDetectionModel(t *testing.T) {
	engine := NewCopilotSDKEngine()

	defaultModel := engine.GetDefaultDetectionModel()
	if defaultModel != string(constants.DefaultCopilotDetectionModel) {
		t.Errorf("Expected default detection model '%s', got '%s'", string(constants.DefaultCopilotDetectionModel), defaultModel)
	}
}

func TestCopilotSDKEngineRequiredSecrets(t *testing.T) {
	engine := NewCopilotSDKEngine()

	// Basic workflow - should require COPILOT_GITHUB_TOKEN
	workflowData := &WorkflowData{}
	secrets := engine.GetRequiredSecretNames(workflowData)

	if len(secrets) != 1 {
		t.Errorf("Expected 1 required secret, got %d: %v", len(secrets), secrets)
	}

	if secrets[0] != "COPILOT_GITHUB_TOKEN" {
		t.Errorf("Expected 'COPILOT_GITHUB_TOKEN', got '%s'", secrets[0])
	}
}

func TestCopilotSDKEngineLogParserScript(t *testing.T) {
	engine := NewCopilotSDKEngine()
	script := engine.GetLogParserScriptId()

	if script != "parse_copilot_log" {
		t.Errorf("Expected 'parse_copilot_log', got '%s'", script)
	}
}

func TestCopilotSDKEngineLogFileForParsing(t *testing.T) {
	engine := NewCopilotSDKEngine()
	logFile := engine.GetLogFileForParsing()

	expected := "/tmp/gh-aw/sandbox/agent/logs/"
	if logFile != expected {
		t.Errorf("Expected '%s', got '%s'", expected, logFile)
	}
}

func TestCopilotSDKEngineRegistered(t *testing.T) {
	registry := NewEngineRegistry()

	engine, err := registry.GetEngine("copilot-sdk")
	if err != nil {
		t.Fatalf("Expected copilot-sdk engine to be registered, got error: %v", err)
	}

	if engine.GetID() != "copilot-sdk" {
		t.Errorf("Expected engine ID 'copilot-sdk', got '%s'", engine.GetID())
	}
}

func TestCopilotSDKEngineComputeAvailableTools(t *testing.T) {
	engine := NewCopilotSDKEngine()

	tests := []struct {
		name        string
		tools       map[string]any
		safeOutputs *SafeOutputsConfig
		expected    []string
		expectNil   bool // nil means "all tools"
	}{
		{
			name:     "empty tools",
			tools:    map[string]any{},
			expected: []string{},
		},
		{
			name: "bash with specific commands",
			tools: map[string]any{
				"bash": []any{"echo", "ls"},
			},
			expected: []string{"shell(echo)", "shell(ls)"},
		},
		{
			name: "bash with wildcard returns nil (all tools)",
			tools: map[string]any{
				"bash": []any{":*"},
			},
			expectNil: true,
		},
		{
			name: "bash with nil (all commands allowed)",
			tools: map[string]any{
				"bash": nil,
			},
			expected: []string{"shell"},
		},
		{
			name: "edit tool",
			tools: map[string]any{
				"edit": nil,
			},
			expected: []string{"write"},
		},
		{
			name:  "safe outputs",
			tools: map[string]any{},
			safeOutputs: &SafeOutputsConfig{
				CreateIssues: &CreateIssuesConfig{},
			},
			expected: []string{"safeoutputs"},
		},
		{
			name: "github tool with allowed tools",
			tools: map[string]any{
				"github": map[string]any{
					"allowed": []any{"get_file_contents", "list_commits"},
				},
			},
			expected: []string{"github(get_file_contents)", "github(list_commits)"},
		},
		{
			name: "github tool with wildcard",
			tools: map[string]any{
				"github": map[string]any{
					"allowed": []any{"*"},
				},
			},
			expected: []string{"github"},
		},
		{
			name: "github tool without allowed field",
			tools: map[string]any{
				"github": map[string]any{},
			},
			expected: []string{"github"},
		},
		{
			name: "mixed tools sorted",
			tools: map[string]any{
				"bash": []any{"git status", "npm test"},
				"edit": nil,
			},
			safeOutputs: &SafeOutputsConfig{
				CreateIssues: &CreateIssuesConfig{},
			},
			expected: []string{"safeoutputs", "shell(git status)", "shell(npm test)", "write"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.computeSDKAvailableTools(tt.tools, tt.safeOutputs, nil, nil)

			if tt.expectNil {
				if result != nil {
					t.Errorf("Expected nil (all tools), got %v", result)
				}
				return
			}

			if result == nil && len(tt.expected) == 0 {
				// Both empty, ok
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d tools, got %d: %v", len(tt.expected), len(result), result)
				return
			}

			for i, expected := range tt.expected {
				if i >= len(result) || result[i] != expected {
					t.Errorf("Expected tool %d to be '%s', got '%s'", i, expected, result[i])
				}
			}
		})
	}
}

func TestCopilotSDKEngineInstallationSteps(t *testing.T) {
	engine := NewCopilotSDKEngine()

	workflowData := &WorkflowData{}
	steps := engine.GetInstallationSteps(workflowData)

	// Should have: secret validation + copilot CLI install + copilot-runner install
	if len(steps) < 2 {
		t.Errorf("Expected at least 2 installation steps, got %d", len(steps))
	}

	// First step should be secret validation
	if len(steps) > 0 {
		stepContent := strings.Join([]string(steps[0]), "\n")
		if !strings.Contains(stepContent, "COPILOT_GITHUB_TOKEN") {
			t.Errorf("Expected first step to validate COPILOT_GITHUB_TOKEN, got:\n%s", stepContent)
		}
	}
}

func TestCopilotSDKEngineSkipInstallationWithCommand(t *testing.T) {
	engine := NewCopilotSDKEngine()

	workflowData := &WorkflowData{
		EngineConfig: &EngineConfig{Command: "/usr/local/bin/custom-runner"},
	}
	steps := engine.GetInstallationSteps(workflowData)

	if len(steps) != 0 {
		t.Errorf("Expected 0 installation steps when command is specified, got %d", len(steps))
	}
}

func TestCopilotSDKEngineExecutionSteps(t *testing.T) {
	engine := NewCopilotSDKEngine()
	workflowData := &WorkflowData{
		Name: "test-workflow",
	}
	steps := engine.GetExecutionSteps(workflowData, "/tmp/gh-aw/test.log")

	if len(steps) != 1 {
		t.Fatalf("Expected 1 execution step, got %d", len(steps))
	}

	stepContent := strings.Join([]string(steps[0]), "\n")

	// Should contain the step name
	if !strings.Contains(stepContent, "Execute Copilot SDK Runner") {
		t.Errorf("Expected step name 'Execute Copilot SDK Runner' in:\n%s", stepContent)
	}

	// Should contain copilot-runner command
	if !strings.Contains(stepContent, "copilot-runner") {
		t.Errorf("Expected 'copilot-runner' command in:\n%s", stepContent)
	}

	// Should contain config file reference
	if !strings.Contains(stepContent, "copilot-runner-config.json") {
		t.Errorf("Expected config file reference in:\n%s", stepContent)
	}

	// Should contain COPILOT_GITHUB_TOKEN env var
	if !strings.Contains(stepContent, "COPILOT_GITHUB_TOKEN") {
		t.Errorf("Expected COPILOT_GITHUB_TOKEN in:\n%s", stepContent)
	}

	// Should contain the log file
	if !strings.Contains(stepContent, "/tmp/gh-aw/test.log") {
		t.Errorf("Expected log file path in:\n%s", stepContent)
	}

	// Should contain GITHUB_WORKSPACE env var
	if !strings.Contains(stepContent, "GITHUB_WORKSPACE") {
		t.Errorf("Expected GITHUB_WORKSPACE in:\n%s", stepContent)
	}
}

func TestCopilotSDKEngineExecutionStepsWithTools(t *testing.T) {
	engine := NewCopilotSDKEngine()
	workflowData := &WorkflowData{
		Name: "test-workflow",
		Tools: map[string]any{
			"bash": []any{"echo", "git status"},
			"edit": nil,
		},
		ParsedTools: NewTools(map[string]any{
			"bash": []any{"echo", "git status"},
			"edit": nil,
		}),
	}
	steps := engine.GetExecutionSteps(workflowData, "/tmp/gh-aw/test.log")

	if len(steps) != 1 {
		t.Fatalf("Expected 1 execution step, got %d", len(steps))
	}

	stepContent := strings.Join([]string(steps[0]), "\n")

	// Should contain the SDK tool arguments comment
	if !strings.Contains(stepContent, "# Copilot SDK available tools") {
		t.Errorf("Expected SDK tool arguments comment in:\n%s", stepContent)
	}

	// Should contain available tools in JSON config
	if !strings.Contains(stepContent, "available_tools") {
		t.Errorf("Expected 'available_tools' in config:\n%s", stepContent)
	}
}

func TestCopilotSDKEngineExecutionStepsWithOutput(t *testing.T) {
	engine := NewCopilotSDKEngine()
	workflowData := &WorkflowData{
		Name:        "test-workflow",
		SafeOutputs: &SafeOutputsConfig{},
	}
	steps := engine.GetExecutionSteps(workflowData, "/tmp/gh-aw/test.log")

	if len(steps) != 1 {
		t.Fatalf("Expected 1 execution step, got %d", len(steps))
	}

	stepContent := strings.Join([]string(steps[0]), "\n")

	// Should contain GH_AW_SAFE_OUTPUTS
	if !strings.Contains(stepContent, "GH_AW_SAFE_OUTPUTS") {
		t.Errorf("Expected GH_AW_SAFE_OUTPUTS when SafeOutputs is not nil in:\n%s", stepContent)
	}
}

func TestCopilotSDKEngineBuildRunnerConfig(t *testing.T) {
	engine := NewCopilotSDKEngine()

	tests := []struct {
		name     string
		workflow *WorkflowData
		check    func(t *testing.T, config RunnerConfig)
	}{
		{
			name:     "basic config",
			workflow: &WorkflowData{},
			check: func(t *testing.T, config RunnerConfig) {
				if config.CLIPath != "/usr/local/bin/copilot" {
					t.Errorf("Expected CLI path '/usr/local/bin/copilot', got '%s'", config.CLIPath)
				}
				if config.PromptFile != "/tmp/gh-aw/aw-prompts/prompt.txt" {
					t.Errorf("Expected prompt file path, got '%s'", config.PromptFile)
				}
				if !config.Streaming {
					t.Error("Expected streaming to be enabled")
				}
			},
		},
		{
			name: "with custom model",
			workflow: &WorkflowData{
				EngineConfig: &EngineConfig{Model: "gpt-4o"},
			},
			check: func(t *testing.T, config RunnerConfig) {
				if config.Model != "gpt-4o" {
					t.Errorf("Expected model 'gpt-4o', got '%s'", config.Model)
				}
			},
		},
		{
			name: "with tools",
			workflow: &WorkflowData{
				Tools: map[string]any{
					"bash": []any{"echo"},
					"edit": nil,
				},
			},
			check: func(t *testing.T, config RunnerConfig) {
				if len(config.AvailableTools) != 2 {
					t.Errorf("Expected 2 available tools, got %d: %v", len(config.AvailableTools), config.AvailableTools)
				}
			},
		},
		{
			name: "with wildcard bash (all tools)",
			workflow: &WorkflowData{
				Tools: map[string]any{
					"bash": []any{"*"},
				},
			},
			check: func(t *testing.T, config RunnerConfig) {
				if config.AvailableTools != nil {
					t.Errorf("Expected nil available tools for wildcard, got %v", config.AvailableTools)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := engine.buildRunnerConfig(tt.workflow)
			tt.check(t, config)
		})
	}
}

func TestCopilotSDKEngineGenerateToolArgumentsComment(t *testing.T) {
	engine := NewCopilotSDKEngine()

	tests := []struct {
		name     string
		tools    map[string]any
		expected string
	}{
		{
			name:     "empty tools",
			tools:    map[string]any{},
			expected: "",
		},
		{
			name: "bash with commands",
			tools: map[string]any{
				"bash": []any{"echo", "ls"},
			},
			expected: "        # Copilot SDK available tools (sorted):\n        # - shell(echo)\n        # - shell(ls)\n",
		},
		{
			name: "wildcard bash",
			tools: map[string]any{
				"bash": []any{":*"},
			},
			expected: "        # SDK tools: all tools enabled (wildcard)\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.generateSDKToolArgumentsComment(tt.tools, nil, nil, nil, "        ")
			if result != tt.expected {
				t.Errorf("Expected:\n%q\nGot:\n%q", tt.expected, result)
			}
		})
	}
}
