// This file implements the GitHub Copilot SDK agentic engine.
//
// The Copilot SDK engine uses the github.com/github/copilot-sdk/go package
// for programmatic agent control instead of constructing CLI flag arguments.
//
// The SDK engine is organized into focused modules:
//   - copilot_sdk_engine.go: Core engine interface and constructor
//   - copilot_sdk_engine_installation.go: Installation workflow generation
//   - copilot_sdk_engine_execution.go: Execution workflow and runtime configuration
//   - copilot_sdk_engine_tools.go: Tool permissions and available/excluded tools mapping
//
// This engine generates steps that invoke a copilot-runner binary with JSON config,
// rather than directly constructing copilot CLI command-line flags.

package workflow

import (
	"strings"

	"github.com/github/gh-aw/pkg/constants"
	"github.com/github/gh-aw/pkg/logger"
)

var copilotSDKLog = logger.New("workflow:copilot_sdk_engine")

const copilotSDKLogsFolder = "/tmp/gh-aw/sandbox/agent/logs/"

// CopilotSDKEngine represents the GitHub Copilot SDK agentic engine.
// It provides integration with the Copilot SDK Go package for programmatic
// agent control, using structured JSON configuration instead of CLI flags.
type CopilotSDKEngine struct {
	BaseEngine
}

func NewCopilotSDKEngine() *CopilotSDKEngine {
	copilotSDKLog.Print("Creating new Copilot SDK engine instance")
	return &CopilotSDKEngine{
		BaseEngine: BaseEngine{
			id:                     "copilot-sdk",
			displayName:            "GitHub Copilot SDK",
			description:            "Uses GitHub Copilot SDK (Go) for programmatic agent control",
			experimental:           true,
			supportsToolsAllowlist: true,
			supportsHTTPTransport:  true,  // Copilot SDK supports HTTP transport via MCP
			supportsMaxTurns:       false, // Max turns not directly supported yet
			supportsWebFetch:       true,  // Copilot SDK has web-fetch support via CLI
			supportsWebSearch:      false, // Copilot SDK does not have web-search
			supportsFirewall:       true,  // Copilot SDK supports network firewalling via AWF
			supportsPlugins:        false, // Plugins not yet supported via SDK
			supportsLLMGateway:     false, // LLM gateway not supported
		},
	}
}

// GetDefaultDetectionModel returns the default model for threat detection
// Uses the same model as the standard Copilot engine
func (e *CopilotSDKEngine) GetDefaultDetectionModel() string {
	return string(constants.DefaultCopilotDetectionModel)
}

// GetRequiredSecretNames returns the list of secrets required by the Copilot SDK engine
func (e *CopilotSDKEngine) GetRequiredSecretNames(workflowData *WorkflowData) []string {
	copilotSDKLog.Print("Collecting required secrets for Copilot SDK engine")
	secrets := []string{"COPILOT_GITHUB_TOKEN"}

	// Add MCP gateway API key if MCP servers are present
	if HasMCPServers(workflowData) {
		copilotSDKLog.Print("Adding MCP_GATEWAY_API_KEY secret")
		secrets = append(secrets, "MCP_GATEWAY_API_KEY")
	}

	// Add GitHub token for GitHub MCP server if present
	if hasGitHubTool(workflowData.ParsedTools) {
		copilotSDKLog.Print("Adding GITHUB_MCP_SERVER_TOKEN secret")
		secrets = append(secrets, "GITHUB_MCP_SERVER_TOKEN")
	}

	// Add HTTP MCP header secret names
	headerSecrets := collectHTTPMCPHeaderSecrets(workflowData.Tools)
	for varName := range headerSecrets {
		secrets = append(secrets, varName)
	}

	// Add safe-inputs secret names
	if IsSafeInputsEnabled(workflowData.SafeInputs, workflowData) {
		safeInputsSecrets := collectSafeInputsSecrets(workflowData.SafeInputs)
		for varName := range safeInputsSecrets {
			secrets = append(secrets, varName)
		}
	}

	copilotSDKLog.Printf("Total required secrets: %d", len(secrets))
	return secrets
}

func (e *CopilotSDKEngine) GetDeclaredOutputFiles() []string {
	return []string{copilotSDKLogsFolder}
}

// GetLogParserScriptId returns the JavaScript script name for parsing Copilot SDK logs
// Uses the same parser as the standard Copilot engine since log format is compatible
func (e *CopilotSDKEngine) GetLogParserScriptId() string {
	return "parse_copilot_log"
}

// GetLogFileForParsing returns the log directory for Copilot SDK logs
func (e *CopilotSDKEngine) GetLogFileForParsing() string {
	return copilotSDKLogsFolder
}

// ParseLogMetrics implements engine-specific log parsing for Copilot SDK.
// The SDK runner outputs structured JSON metrics, which makes parsing simpler
// than the standard Copilot engine's debug log format.
func (e *CopilotSDKEngine) ParseLogMetrics(logContent string, verbose bool) LogMetrics {
	// Reuse the Copilot engine's JSONL session parser since the SDK runner
	// outputs compatible JSONL format to the session state directory
	copilotEngine := &CopilotEngine{}
	return copilotEngine.ParseLogMetrics(logContent, verbose)
}

// RenderMCPConfig generates MCP server configuration for the Copilot SDK engine.
// Uses the same format as the standard Copilot engine since the SDK wraps the CLI.
func (e *CopilotSDKEngine) RenderMCPConfig(yamlBuilder *strings.Builder, tools map[string]any, mcpTools []string, workflowData *WorkflowData) {
	// Delegate to the standard Copilot engine's MCP config renderer
	// The SDK uses the same MCP config format since it wraps the CLI
	copilotEngine := &CopilotEngine{}
	copilotEngine.RenderMCPConfig(yamlBuilder, tools, mcpTools, workflowData)
}

// GetFirewallLogsCollectionStep returns steps for collecting firewall logs
func (e *CopilotSDKEngine) GetFirewallLogsCollectionStep(workflowData *WorkflowData) []GitHubActionStep {
	// Reuse the standard Copilot engine's session file copy step
	copilotEngine := &CopilotEngine{}
	return copilotEngine.GetFirewallLogsCollectionStep(workflowData)
}

// GetSquidLogsSteps returns the steps for uploading and parsing Squid logs
func (e *CopilotSDKEngine) GetSquidLogsSteps(workflowData *WorkflowData) []GitHubActionStep {
	copilotEngine := &CopilotEngine{}
	return copilotEngine.GetSquidLogsSteps(workflowData)
}

// GetCleanupStep returns the post-execution cleanup step
func (e *CopilotSDKEngine) GetCleanupStep(workflowData *WorkflowData) GitHubActionStep {
	return GitHubActionStep([]string{})
}

// GetErrorPatterns returns regex patterns for detecting errors in SDK runner output
func (e *CopilotSDKEngine) GetErrorPatterns() []string {
	return []string{
		`(?i)error:?\s+(.+)`,
		`(?i)fatal:?\s+(.+)`,
		`(?i)panic:?\s+(.+)`,
		`(?i)failed to\s+(.+)`,
		`(?i)SDK error:?\s+(.+)`,
		`(?i)session error:?\s+(.+)`,
	}
}
