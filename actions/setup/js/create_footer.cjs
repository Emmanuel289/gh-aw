// @ts-check
/// <reference types="@actions/github-script" />

/**
 * Footer Creation Module
 *
 * This module provides enhanced footer generation that includes agent, model, and cost information
 * from aw_info.json and workflow logs.
 */

const fs = require("fs");
const path = require("path");

/**
 * @typedef {Object} AwInfoData
 * @property {string} engine_id - Engine identifier (copilot, claude, codex, custom)
 * @property {string} engine_name - Engine display name
 * @property {string} [model] - Model name if specified
 * @property {string} [version] - Engine version
 * @property {string} [agent_version] - Agent installation version
 * @property {string} [cli_version] - CLI version
 * @property {string} workflow_name - Workflow name
 * @property {boolean} [experimental] - Whether engine is experimental
 * @property {number} run_id - GitHub Actions run ID
 * @property {string} [repository] - Repository name (owner/repo)
 */

/**
 * @typedef {Object} CostInfo
 * @property {number} [total_cost_usd] - Total cost in USD
 * @property {string} [agent] - Agent name/engine
 * @property {string} [model] - Model used
 */

/**
 * @typedef {Object} FooterInfo
 * @property {string} agent - Agent/engine name
 * @property {string} [model] - Model name if available
 * @property {number} [cost] - Cost in USD if available
 * @property {Object} [detection] - Detection job info if available
 * @property {string} [detection.agent] - Detection agent/engine
 * @property {string} [detection.model] - Detection model
 * @property {number} [detection.cost] - Detection cost
 */

/**
 * Reads aw_info.json file if it exists
 * @returns {AwInfoData|null} Parsed aw_info data or null if not found
 */
function readAwInfo() {
  const awInfoPath = "/tmp/gh-aw/aw_info.json";

  try {
    if (!fs.existsSync(awInfoPath)) {
      return null;
    }

    const content = fs.readFileSync(awInfoPath, "utf8");
    return JSON.parse(content);
  } catch (error) {
    // Silently return null if file doesn't exist or can't be parsed
    // This is expected in workflows without aw_info generation
    return null;
  }
}

/**
 * Attempts to parse cost information from log files
 * @param {string} logDir - Directory containing log files
 * @returns {CostInfo|null} Parsed cost info or null
 */
function parseCostFromLogs(logDir) {
  // TODO: Implement log parsing for cost extraction
  // This would read from parsed log files in /tmp/gh-aw/
  // and extract total_cost_usd from the last entry
  return null;
}

/**
 * Formats cost value for display
 * @param {number} cost - Cost in USD
 * @returns {string} Formatted cost string
 */
function formatCost(cost) {
  if (cost < 0.01) {
    return `$${cost.toFixed(4)}`;
  }
  return `$${cost.toFixed(2)}`;
}

/**
 * Creates footer information from aw_info.json and logs
 * @returns {FooterInfo|null} Footer info or null if not available
 */
function createFooterInfo() {
  const awInfo = readAwInfo();

  if (!awInfo) {
    return null;
  }

  /** @type {FooterInfo} */
  const footerInfo = {
    agent: awInfo.engine_name || awInfo.engine_id || "unknown",
  };

  // Add model if specified
  if (awInfo.model) {
    footerInfo.model = awInfo.model;
  }

  // Attempt to parse cost from logs
  const costInfo = parseCostFromLogs("/tmp/gh-aw");
  if (costInfo && costInfo.total_cost_usd) {
    footerInfo.cost = costInfo.total_cost_usd;
  }

  // TODO: Add detection job information parsing
  // This would require reading detection job outputs or logs

  return footerInfo;
}

/**
 * Generates a single-line informational footer with agent, model, and cost info
 * @returns {string} Single-line footer string or empty if no info available
 */
function generateInfoFooter() {
  const info = createFooterInfo();

  if (!info) {
    return "";
  }

  const parts = [];

  // Agent name
  parts.push(info.agent);

  // Model (if specified)
  if (info.model) {
    parts.push(info.model);
  }

  // Cost (if known)
  if (info.cost !== undefined) {
    parts.push(formatCost(info.cost));
  }

  // Detection info (if available)
  if (info.detection) {
    const detectionParts = [];

    if (info.detection.agent) {
      detectionParts.push(info.detection.agent);
    }

    if (info.detection.model) {
      detectionParts.push(info.detection.model);
    }

    if (info.detection.cost !== undefined) {
      detectionParts.push(formatCost(info.detection.cost));
    }

    if (detectionParts.length > 0) {
      parts.push(`detection: ${detectionParts.join(", ")}`);
    }
  }

  // Return single-line format
  return parts.join(", ");
}

module.exports = {
  readAwInfo,
  createFooterInfo,
  generateInfoFooter,
  formatCost,
  parseCostFromLogs,
};
