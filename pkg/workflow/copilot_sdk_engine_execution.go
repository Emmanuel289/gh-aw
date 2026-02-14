// This file provides Copilot SDK engine execution logic.
//
// Instead of constructing copilot CLI command-line flags directly, this engine
// generates a JSON configuration that is passed to the copilot-runner binary.
// The runner uses the Copilot SDK Go package (github.com/github/copilot-sdk/go)
// to create a client, establish a session, and send the prompt programmatically.
//
// The execution strategy:
//   1. Generate a JSON config with model, tools, MCP servers, etc.
//   2. Write the config to a temporary file
//   3. Invoke copilot-runner with the config file path
//   4. The runner uses SDK's Client/Session/SendAndWait APIs
//   5. Optionally wrap with AWF for sandboxed execution

package workflow

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/github/gh-aw/pkg/constants"
	"github.com/github/gh-aw/pkg/logger"
)

var copilotSDKExecLog = logger.New("workflow:copilot_sdk_engine_execution")

// RunnerConfig represents the JSON configuration passed to the copilot-runner binary.
// This structure mirrors the SDK's SessionConfig but is serialized as JSON for
// cross-process communication.
type RunnerConfig struct {
	CLIPath          string            `json:"cli_path"`
	GithubToken      string            `json:"github_token,omitempty"`
	Model            string            `json:"model,omitempty"`
	WorkingDirectory string            `json:"working_directory,omitempty"`
	LogLevel         string            `json:"log_level,omitempty"`
	LogDir           string            `json:"log_dir,omitempty"`
	PromptFile       string            `json:"prompt_file"`
	AvailableTools   []string          `json:"available_tools,omitempty"`
	ExcludedTools    []string          `json:"excluded_tools,omitempty"`
	MCPConfigPath    string            `json:"mcp_config_path,omitempty"`
	Streaming        bool              `json:"streaming"`
	Timeout          int               `json:"timeout,omitempty"`
	MetricsFile      string            `json:"metrics_file,omitempty"`
	Env              map[string]string `json:"env,omitempty"`
}

// GetExecutionSteps returns the GitHub Actions steps for executing the Copilot SDK engine.
func (e *CopilotSDKEngine) GetExecutionSteps(workflowData *WorkflowData, logFile string) []GitHubActionStep {
	copilotSDKExecLog.Printf("Generating execution steps for Copilot SDK: workflow=%s, firewall=%v", workflowData.Name, isFirewallEnabled(workflowData))

	// Handle custom steps if they exist in engine config
	steps := InjectCustomEngineSteps(workflowData, e.convertStepToYAML)

	// Build the runner configuration
	config := e.buildRunnerConfig(workflowData)

	// Serialize config to JSON for embedding in the shell script
	configJSON, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		copilotSDKExecLog.Printf("Error marshaling runner config: %v", err)
		configJSON = []byte("{}")
	}

	// Determine which copilot-runner command to use
	sandboxEnabled := isFirewallEnabled(workflowData) || isSRTEnabled(workflowData)
	var runnerCommandName string
	if workflowData.EngineConfig != nil && workflowData.EngineConfig.Command != "" {
		runnerCommandName = workflowData.EngineConfig.Command
	} else if sandboxEnabled {
		runnerCommandName = "/usr/local/bin/copilot-runner"
	} else {
		runnerCommandName = "copilot-runner"
	}

	// Conditionally add model from environment variable
	modelConfigured := workflowData.EngineConfig != nil && workflowData.EngineConfig.Model != ""
	isDetectionJob := workflowData.SafeOutputs == nil
	var modelEnvVar string
	if isDetectionJob {
		modelEnvVar = constants.EnvVarModelDetectionCopilot
	} else {
		modelEnvVar = constants.EnvVarModelAgentCopilot
	}

	// Build the command that writes config and runs copilot-runner
	var commandParts []string
	commandParts = append(commandParts, "set -o pipefail")
	commandParts = append(commandParts, fmt.Sprintf("mkdir -p %s", copilotSDKLogsFolder))

	// Write runner config to a temporary file
	commandParts = append(commandParts, "# Write Copilot SDK runner configuration")
	commandParts = append(commandParts, fmt.Sprintf("cat > /tmp/gh-aw/copilot-runner-config.json << 'RUNNER_CONFIG_EOF'\n%s\nRUNNER_CONFIG_EOF", string(configJSON)))

	// Optionally inject model from environment variable into the config
	if !modelConfigured {
		commandParts = append(commandParts, fmt.Sprintf(`# Inject model from environment variable if set
if [ -n "$%s" ]; then
  jq --arg model "$%s" '.model = $model' /tmp/gh-aw/copilot-runner-config.json > /tmp/gh-aw/copilot-runner-config.json.tmp && mv /tmp/gh-aw/copilot-runner-config.json.tmp /tmp/gh-aw/copilot-runner-config.json
fi`, modelEnvVar, modelEnvVar))
	}

	// Build the runner command
	runnerCommand := fmt.Sprintf("%s --config /tmp/gh-aw/copilot-runner-config.json", runnerCommandName)

	// Conditionally wrap with sandbox (AWF)
	if isFirewallEnabled(workflowData) {
		firewallConfig := getFirewallConfig(workflowData)
		agentConfig := getAgentConfig(workflowData)
		var awfLogLevel = "info"
		if firewallConfig != nil && firewallConfig.LogLevel != "" {
			awfLogLevel = firewallConfig.LogLevel
		}

		allowedDomains := GetCopilotAllowedDomainsWithToolsAndRuntimes(workflowData.NetworkPermissions, workflowData.Tools, workflowData.Runtimes)

		var awfArgs []string
		awfArgs = append(awfArgs, "--env-all")
		awfArgs = append(awfArgs, "--container-workdir", "\"${GITHUB_WORKSPACE}\"")

		if agentConfig != nil && len(agentConfig.Mounts) > 0 {
			sortedMounts := make([]string, len(agentConfig.Mounts))
			copy(sortedMounts, agentConfig.Mounts)
			sort.Strings(sortedMounts)
			for _, mount := range sortedMounts {
				awfArgs = append(awfArgs, "--mount", mount)
			}
		}

		awfArgs = append(awfArgs, "--allow-domains", allowedDomains)

		blockedDomains := formatBlockedDomains(workflowData.NetworkPermissions)
		if blockedDomains != "" {
			awfArgs = append(awfArgs, "--block-domains", blockedDomains)
		}

		awfArgs = append(awfArgs, "--log-level", awfLogLevel)
		awfArgs = append(awfArgs, "--proxy-logs-dir", "/tmp/gh-aw/sandbox/firewall/logs")

		if HasMCPServers(workflowData) {
			awfArgs = append(awfArgs, "--enable-host-access")
		}

		awfImageTag := getAWFImageTag(firewallConfig)
		awfArgs = append(awfArgs, "--image-tag", awfImageTag)
		awfArgs = append(awfArgs, "--skip-pull")

		sslBumpArgs := getSSLBumpArgs(firewallConfig)
		awfArgs = append(awfArgs, sslBumpArgs...)

		if firewallConfig != nil && len(firewallConfig.Args) > 0 {
			awfArgs = append(awfArgs, firewallConfig.Args...)
		}

		if agentConfig != nil && len(agentConfig.Args) > 0 {
			awfArgs = append(awfArgs, agentConfig.Args...)
		}

		var awfCommand string
		if agentConfig != nil && agentConfig.Command != "" {
			awfCommand = agentConfig.Command
		} else {
			awfCommand = "sudo -E awf"
		}

		escapedRunnerCommand := shellEscapeArg(runnerCommand)
		commandParts = append(commandParts, fmt.Sprintf(`%s %s \
  -- %s \
  2>&1 | tee %s`, awfCommand, shellJoinArgs(awfArgs), escapedRunnerCommand, shellEscapeArg(logFile)))
	} else {
		// Non-sandbox mode: run copilot-runner directly
		commandParts = append(commandParts, fmt.Sprintf(`%s 2>&1 | tee %s`, runnerCommand, logFile))
	}

	command := strings.Join(commandParts, "\n")

	// Build environment variables
	copilotGitHubToken := "${{ secrets.COPILOT_GITHUB_TOKEN }}"
	if workflowData.GitHubToken != "" {
		copilotGitHubToken = workflowData.GitHubToken
	}

	env := map[string]string{
		"XDG_CONFIG_HOME":           "/home/runner",
		"COPILOT_AGENT_RUNNER_TYPE": "SDK",
		"COPILOT_GITHUB_TOKEN":      copilotGitHubToken,
		"GITHUB_STEP_SUMMARY":       "${{ env.GITHUB_STEP_SUMMARY }}",
		"GITHUB_HEAD_REF":           "${{ github.head_ref }}",
		"GITHUB_REF_NAME":           "${{ github.ref_name }}",
		"GITHUB_WORKSPACE":          "${{ github.workspace }}",
		"GH_AW_PROMPT":              "/tmp/gh-aw/aw-prompts/prompt.txt",
	}

	// Add MCP config path if MCP servers are present
	if HasMCPServers(workflowData) {
		env["GH_AW_MCP_CONFIG"] = "/home/runner/.copilot/mcp-config.json"
	}

	if hasGitHubTool(workflowData.ParsedTools) {
		customGitHubToken := getGitHubToken(workflowData.Tools["github"])
		effectiveToken := getEffectiveGitHubToken(customGitHubToken, workflowData.GitHubToken)
		env["GITHUB_MCP_SERVER_TOKEN"] = effectiveToken
	}

	applySafeOutputEnvToMap(env, workflowData)

	if workflowData.ToolsStartupTimeout > 0 {
		env["GH_AW_STARTUP_TIMEOUT"] = fmt.Sprintf("%d", workflowData.ToolsStartupTimeout)
	}
	if workflowData.ToolsTimeout > 0 {
		env["GH_AW_TOOL_TIMEOUT"] = fmt.Sprintf("%d", workflowData.ToolsTimeout)
	}
	if workflowData.EngineConfig != nil && workflowData.EngineConfig.MaxTurns != "" {
		env["GH_AW_MAX_TURNS"] = workflowData.EngineConfig.MaxTurns
	}

	// Add model environment variable if not explicitly configured
	if workflowData.EngineConfig == nil || workflowData.EngineConfig.Model == "" {
		if isDetectionJob {
			env[constants.EnvVarModelDetectionCopilot] = fmt.Sprintf("${{ vars.%s || '' }}", constants.EnvVarModelDetectionCopilot)
		} else {
			env[constants.EnvVarModelAgentCopilot] = fmt.Sprintf("${{ vars.%s || '' }}", constants.EnvVarModelAgentCopilot)
		}
	}

	// Add custom environment variables from engine config
	if workflowData.EngineConfig != nil && len(workflowData.EngineConfig.Env) > 0 {
		for key, value := range workflowData.EngineConfig.Env {
			env[key] = value
		}
	}

	// Add custom environment variables from agent config
	agentConfig := getAgentConfig(workflowData)
	if agentConfig != nil && len(agentConfig.Env) > 0 {
		for key, value := range agentConfig.Env {
			env[key] = value
		}
	}

	// Add HTTP MCP header secrets
	headerSecrets := collectHTTPMCPHeaderSecrets(workflowData.Tools)
	for varName, secretExpr := range headerSecrets {
		if _, exists := env[varName]; !exists {
			env[varName] = secretExpr
		}
	}

	// Add safe-inputs secrets
	if IsSafeInputsEnabled(workflowData.SafeInputs, workflowData) {
		safeInputsSecrets := collectSafeInputsSecrets(workflowData.SafeInputs)
		for varName, secretExpr := range safeInputsSecrets {
			if _, exists := env[varName]; !exists {
				env[varName] = secretExpr
			}
		}
	}

	// Generate the step
	stepName := "Execute Copilot SDK Runner"
	var stepLines []string

	stepLines = append(stepLines, fmt.Sprintf("      - name: %s", stepName))
	stepLines = append(stepLines, "        id: agentic_execution")

	// Add tool arguments comment
	toolArgsComment := e.generateSDKToolArgumentsComment(workflowData.Tools, workflowData.SafeOutputs, workflowData.SafeInputs, workflowData, "        ")
	if toolArgsComment != "" {
		commentLines := strings.Split(strings.TrimSuffix(toolArgsComment, "\n"), "\n")
		stepLines = append(stepLines, commentLines...)
	}

	// Add timeout
	if workflowData.TimeoutMinutes != "" {
		timeoutValue := strings.TrimPrefix(workflowData.TimeoutMinutes, "timeout-minutes: ")
		stepLines = append(stepLines, fmt.Sprintf("        timeout-minutes: %s", timeoutValue))
	} else {
		stepLines = append(stepLines, fmt.Sprintf("        timeout-minutes: %d", int(constants.DefaultAgenticWorkflowTimeout/time.Minute)))
	}

	// Filter environment variables
	allowedSecrets := e.GetRequiredSecretNames(workflowData)
	filteredEnv := FilterEnvForSecrets(env, allowedSecrets)

	// Format step with command and env
	stepLines = FormatStepWithCommandAndEnv(stepLines, command, filteredEnv)

	steps = append(steps, GitHubActionStep(stepLines))

	return steps
}

// buildRunnerConfig constructs the RunnerConfig from workflow data.
func (e *CopilotSDKEngine) buildRunnerConfig(workflowData *WorkflowData) RunnerConfig {
	config := RunnerConfig{
		CLIPath:          "/usr/local/bin/copilot",
		LogLevel:         "info",
		LogDir:           copilotSDKLogsFolder,
		PromptFile:       "/tmp/gh-aw/aw-prompts/prompt.txt",
		Streaming:        true,
		MetricsFile:      copilotSDKLogsFolder + "sdk-metrics.json",
		WorkingDirectory: "${GITHUB_WORKSPACE}",
	}

	// Set model if explicitly configured
	if workflowData.EngineConfig != nil && workflowData.EngineConfig.Model != "" {
		config.Model = workflowData.EngineConfig.Model
	}

	// Set MCP config path if MCP servers are present
	if HasMCPServers(workflowData) {
		config.MCPConfigPath = "/home/runner/.copilot/mcp-config.json"
	}

	// Compute available tools
	availableTools := e.computeSDKAvailableTools(workflowData.Tools, workflowData.SafeOutputs, workflowData.SafeInputs, workflowData)
	if availableTools != nil {
		config.AvailableTools = availableTools
	}

	return config
}
