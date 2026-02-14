// This file provides Copilot SDK engine installation logic.
//
// The Copilot SDK engine requires:
//   1. The Copilot CLI binary (since the SDK wraps it via JSON-RPC)
//   2. The copilot-runner binary (built from cmd/copilot-runner/)
//
// Installation order:
//   1. Secret validation (COPILOT_GITHUB_TOKEN)
//   2. Copilot CLI installation (using official installer)
//   3. copilot-runner binary installation (downloaded from gh-aw releases)
//   4. Sandbox installation (AWF, if needed)

package workflow

import (
	"fmt"

	"github.com/github/gh-aw/pkg/constants"
	"github.com/github/gh-aw/pkg/logger"
)

var copilotSDKInstallLog = logger.New("workflow:copilot_sdk_engine_installation")

// GetInstallationSteps generates the complete installation workflow for the Copilot SDK engine.
func (e *CopilotSDKEngine) GetInstallationSteps(workflowData *WorkflowData) []GitHubActionStep {
	copilotSDKInstallLog.Printf("Generating installation steps for Copilot SDK engine: workflow=%s", workflowData.Name)

	// Skip installation if custom command is specified
	if workflowData.EngineConfig != nil && workflowData.EngineConfig.Command != "" {
		copilotSDKInstallLog.Printf("Skipping installation steps: custom command specified (%s)", workflowData.EngineConfig.Command)
		return []GitHubActionStep{}
	}

	var steps []GitHubActionStep

	// Add secret validation step
	secretValidation := GenerateMultiSecretValidationStep(
		[]string{"COPILOT_GITHUB_TOKEN"},
		"GitHub Copilot SDK",
		"https://github.github.com/gh-aw/reference/engines/#github-copilot-sdk",
	)
	steps = append(steps, secretValidation)

	// Determine Copilot CLI version - SDK still needs the CLI binary
	copilotVersion := string(constants.DefaultCopilotVersion)
	if workflowData.EngineConfig != nil && workflowData.EngineConfig.Version != "" {
		copilotVersion = workflowData.EngineConfig.Version
	}

	// Install Copilot CLI using the official installer script
	copilotSDKInstallLog.Print("Adding Copilot CLI installation step (required by SDK)")
	copilotInstallSteps := GenerateCopilotInstallerSteps(copilotVersion, "Install GitHub Copilot CLI (for SDK)")
	steps = append(steps, copilotInstallSteps...)

	// Add copilot-runner binary installation step
	copilotSDKInstallLog.Print("Adding copilot-runner binary installation step")
	runnerInstallStep := e.generateRunnerInstallationStep()
	steps = append(steps, runnerInstallStep)

	// Add sandbox installation steps if firewall is enabled
	if isFirewallEnabled(workflowData) {
		firewallConfig := getFirewallConfig(workflowData)
		agentConfig := getAgentConfig(workflowData)
		var awfVersion string
		if firewallConfig != nil {
			awfVersion = firewallConfig.Version
		}

		awfInstall := generateAWFInstallationStep(awfVersion, agentConfig)
		if len(awfInstall) > 0 {
			steps = append(steps, awfInstall)
		}
	}

	return steps
}

// generateRunnerInstallationStep creates a GitHub Actions step to install the copilot-runner binary.
// The runner is built from cmd/copilot-runner/ and uses the Copilot SDK Go package.
func (e *CopilotSDKEngine) generateRunnerInstallationStep() GitHubActionStep {
	stepLines := []string{
		"      - name: Install copilot-runner binary",
		"        run: |",
		"          # Build copilot-runner from source using the gh-aw setup action",
		"          # The binary is pre-built and included in the setup action",
		fmt.Sprintf("          echo \"Installing copilot-runner for Copilot SDK engine\""),
		"          if [ -f /opt/gh-aw/bin/copilot-runner ]; then",
		"            cp /opt/gh-aw/bin/copilot-runner /usr/local/bin/copilot-runner",
		"            chmod +x /usr/local/bin/copilot-runner",
		"            echo \"copilot-runner installed successfully\"",
		"          else",
		"            echo \"Warning: copilot-runner binary not found in setup action\"",
		"            echo \"The copilot-sdk engine requires the copilot-runner binary\"",
		"            exit 1",
		"          fi",
	}

	return GitHubActionStep(stepLines)
}
