// This file provides Copilot SDK engine tool permission mapping.
//
// Instead of generating --allow-tool CLI flags (like the standard Copilot engine),
// the SDK engine maps workflow tool configurations to AvailableTools and ExcludedTools
// arrays that are passed via JSON config to the copilot-runner binary.
//
// The copilot-runner then passes these arrays directly to the SDK's SessionConfig.

package workflow

import (
	"fmt"
	"sort"

	"github.com/github/gh-aw/pkg/constants"
	"github.com/github/gh-aw/pkg/logger"
)

var copilotSDKToolsLog = logger.New("workflow:copilot_sdk_engine_tools")

// computeSDKAvailableTools maps workflow tool configurations to a sorted list
// of tool names for the SDK's AvailableTools field.
//
// This replaces the --allow-tool flag generation used by the standard Copilot engine.
// The returned tool names match the Copilot CLI's internal tool identifiers.
func (e *CopilotSDKEngine) computeSDKAvailableTools(tools map[string]any, safeOutputs *SafeOutputsConfig, safeInputs *SafeInputsConfig, workflowData *WorkflowData) []string {
	if tools == nil {
		tools = make(map[string]any)
	}

	// Initialize as empty slice (not nil) to distinguish from "all tools" (nil)
	available := make([]string, 0)

	// Check if bash has wildcard - if so, all tools are available (empty list = all)
	if bashConfig, hasBash := tools["bash"]; hasBash {
		if bashCommands, ok := bashConfig.([]any); ok {
			for _, cmd := range bashCommands {
				if cmdStr, ok := cmd.(string); ok {
					if cmdStr == ":*" || cmdStr == "*" {
						// Return empty list to signal all tools are available
						// The runner will not set AvailableTools (allowing all)
						return nil
					}
				}
			}
		}
	}

	// Handle bash/shell tools
	if bashConfig, hasBash := tools["bash"]; hasBash {
		if bashCommands, ok := bashConfig.([]any); ok {
			for _, cmd := range bashCommands {
				if cmdStr, ok := cmd.(string); ok {
					available = append(available, fmt.Sprintf("shell(%s)", cmdStr))
				}
			}
		} else {
			// Bash with no specific commands - allow all shell
			available = append(available, "shell")
		}
	}

	// Handle edit tools
	if _, hasEdit := tools["edit"]; hasEdit {
		available = append(available, "write")
	}

	// Handle safe_outputs MCP server
	if HasSafeOutputsEnabled(safeOutputs) {
		available = append(available, constants.SafeOutputsMCPServerID)
	}

	// Handle safe_inputs MCP server
	if IsSafeInputsEnabled(safeInputs, workflowData) {
		available = append(available, constants.SafeInputsMCPServerID)
	}

	// Handle web-fetch builtin tool
	if _, hasWebFetch := tools["web-fetch"]; hasWebFetch {
		available = append(available, "web_fetch")
	}

	// Built-in tool names that should be skipped when processing MCP servers
	builtInTools := map[string]bool{
		"bash":       true,
		"edit":       true,
		"web-search": true,
		"playwright": true,
	}

	// Handle MCP server tools
	for toolName, toolConfig := range tools {
		if builtInTools[toolName] {
			continue
		}

		// GitHub is a special case
		if toolName == "github" {
			if toolConfigMap, ok := toolConfig.(map[string]any); ok {
				if allowed, hasAllowed := toolConfigMap["allowed"]; hasAllowed {
					if allowedList, ok := allowed.([]any); ok {
						hasWildcard := false
						for _, allowedTool := range allowedList {
							if toolStr, ok := allowedTool.(string); ok {
								if toolStr == "*" {
									hasWildcard = true
								} else {
									available = append(available, fmt.Sprintf("github(%s)", toolStr))
								}
							}
						}
						if hasWildcard {
							available = append(available, "github")
						}
					}
				} else {
					available = append(available, "github")
				}
			} else {
				available = append(available, "github")
			}
			continue
		}

		// Check if this is an MCP server configuration
		if toolConfigMap, ok := toolConfig.(map[string]any); ok {
			if hasMcp, _ := hasMCPConfig(toolConfigMap); hasMcp {
				available = append(available, toolName)

				if allowed, hasAllowed := toolConfigMap["allowed"]; hasAllowed {
					if allowedList, ok := allowed.([]any); ok {
						for _, allowedTool := range allowedList {
							if toolStr, ok := allowedTool.(string); ok {
								available = append(available, fmt.Sprintf("%s(%s)", toolName, toolStr))
							}
						}
					}
				}
			}
		}
	}

	sort.Strings(available)

	copilotSDKToolsLog.Printf("Computed %d available tools for SDK config", len(available))
	return available
}

// generateSDKToolArgumentsComment generates a multi-line comment showing SDK tool configuration.
func (e *CopilotSDKEngine) generateSDKToolArgumentsComment(tools map[string]any, safeOutputs *SafeOutputsConfig, safeInputs *SafeInputsConfig, workflowData *WorkflowData, indent string) string {
	available := e.computeSDKAvailableTools(tools, safeOutputs, safeInputs, workflowData)
	if available == nil {
		return indent + "# SDK tools: all tools enabled (wildcard)\n"
	}
	if len(available) == 0 {
		return ""
	}

	var comment string
	comment += indent + "# Copilot SDK available tools (sorted):\n"
	for _, tool := range available {
		comment += indent + "# - " + tool + "\n"
	}

	return comment
}
