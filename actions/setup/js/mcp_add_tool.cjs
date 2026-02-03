// @ts-check

/**
 * MCP Add Tool Handler for Safe-Inputs
 *
 * This handler provides a safe-inputs tool for adding MCP servers to workflows
 * by calling the Go AddMCPTool function directly via a Go script handler.
 *
 * This avoids the overhead of spawning the gh-aw binary and instead calls the
 * Go function directly through go run, which is the same pattern used by other
 * safe-inputs Go handlers.
 */

const path = require("path");
const { createGoHandler } = require("./mcp_handler_go.cjs");

/**
 * Create an MCP add tool handler that calls the Go AddMCPTool function directly
 * @param {Object} server - The MCP server instance for logging
 * @param {string} toolName - Name of the tool for logging purposes
 * @param {number} [timeoutSeconds=120] - Timeout in seconds (default 120 for MCP operations)
 * @returns {Function} Async handler function
 */
function createMCPAddHandler(server, toolName, timeoutSeconds = 120) {
  // Get the path to the Go script that wraps AddMCPTool
  const goScriptPath = path.join(__dirname, "..", "go", "mcp_add_tool.go");

  // Use the standard Go handler to execute the script
  return createGoHandler(server, toolName, goScriptPath, timeoutSeconds);
}

module.exports = {
  createMCPAddHandler,
};
