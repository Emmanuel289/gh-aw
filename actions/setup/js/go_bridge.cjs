// @ts-check

/**
 * Go Bridge Module
 *
 * This module provides a bridge for calling Go functions from JavaScript
 * by invoking the gh-aw binary with appropriate commands.
 */

const { spawn } = require("child_process");
const path = require("path");

/**
 * Find the gh-aw binary path
 * @returns {string} Path to gh-aw binary
 */
function getGhAwBinaryPath() {
  // Check if GH_AW_BINARY environment variable is set
  if (process.env.GH_AW_BINARY) {
    return process.env.GH_AW_BINARY;
  }

  // Default to ./gh-aw in the workspace root
  const workspaceRoot = process.env.GITHUB_WORKSPACE || process.cwd();
  return path.join(workspaceRoot, "gh-aw");
}

/**
 * Call the Go AddMCPTool function
 * @param {Object} params - Parameters for AddMCPTool
 * @param {string} params.workflowFile - Workflow file path or ID
 * @param {string} params.mcpServerID - MCP server identifier
 * @param {string} [params.registryURL] - Optional registry URL
 * @param {string} [params.transportType] - Optional transport type
 * @param {string} [params.customToolID] - Optional custom tool ID
 * @param {boolean} [params.verbose] - Verbose output
 * @returns {Promise<Object>} Result of the operation
 */
async function AddMCPTool(params) {
  const {
    workflowFile,
    mcpServerID,
    registryURL = "",
    transportType = "",
    customToolID = "",
    verbose = false,
  } = params;

  const binaryPath = getGhAwBinaryPath();
  const args = ["mcp", "add", workflowFile, mcpServerID];

  if (registryURL) {
    args.push("--registry", registryURL);
  }
  if (transportType) {
    args.push("--transport", transportType);
  }
  if (customToolID) {
    args.push("--tool-id", customToolID);
  }
  if (verbose) {
    args.push("--verbose");
  }

  return new Promise((resolve, reject) => {
    const childProcess = spawn(binaryPath, args, {
      cwd: process.env.GITHUB_WORKSPACE || process.cwd(),
      env: process.env,
    });

    let stdout = "";
    let stderr = "";

    childProcess.stdout.on("data", data => {
      stdout += data.toString();
    });

    childProcess.stderr.on("data", data => {
      stderr += data.toString();
    });

    childProcess.on("error", error => {
      reject(new Error(`Failed to spawn gh-aw: ${error.message}`));
    });

    childProcess.on("close", code => {
      if (code === 0) {
        resolve({
          success: true,
          stdout: stdout.trim(),
          stderr: stderr.trim(),
        });
      } else {
        reject(new Error(`gh-aw mcp add failed with exit code ${code}\nStderr: ${stderr}\nStdout: ${stdout}`));
      }
    });
  });
}

module.exports = {
  AddMCPTool,
  getGhAwBinaryPath,
};
