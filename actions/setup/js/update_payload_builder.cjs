// @ts-check

/**
 * Shared helper for building update payload data across different entity types
 * (issues, pull requests, discussions).
 *
 * This module extracts common payload-building logic to reduce duplication
 * and ensure consistent behavior across update handlers.
 *
 * @module update_payload_builder
 */

/**
 * Configuration for building update payload data
 * @typedef {Object} BuildPayloadConfig
 * @property {string} [defaultOperation] - Default operation for body updates (e.g., "append", "replace")
 * @property {boolean} [allowTitle] - Whether title updates are allowed (default: true)
 * @property {boolean} [allowBody] - Whether body updates are allowed (default: true)
 * @property {boolean} [acceptStateAndStatus] - Whether to accept both 'state' and 'status' fields (default: false)
 * @property {string[]} [additionalFields] - Additional fields to copy from item to updateData (e.g., ["labels", "assignees", "milestone"])
 * @property {boolean} [requireUpdates] - Whether to return a skip result if no updates are provided (default: false)
 */

/**
 * Build update payload data from a message item
 *
 * This function handles common patterns across all update handlers:
 * - Title updates (with optional allow_title config)
 * - Body updates with operation semantics (with optional allow_body config)
 * - State/status updates (with optional backwards compatibility)
 * - Additional entity-specific fields (labels, assignees, milestone, base, etc.)
 *
 * @param {Object} item - The message item containing update fields
 * @param {Object} config - Configuration object from safe-outputs
 * @param {BuildPayloadConfig} payloadConfig - Payload building configuration
 * @returns {{success: true, data: Object} | {success: true, skipped: true, reason: string}} Update data result
 */
function buildUpdatePayloadData(item, config, payloadConfig) {
  const { defaultOperation = "append", allowTitle = true, allowBody = true, acceptStateAndStatus = false, additionalFields = [], requireUpdates = false } = payloadConfig;

  // Check if title and body updates are allowed per config
  const canUpdateTitle = allowTitle && config.allow_title !== false;
  const canUpdateBody = allowBody && config.allow_body !== false;

  const updateData = {};
  let hasUpdates = false;

  // Handle title updates
  if (canUpdateTitle && item.title !== undefined) {
    updateData.title = item.title;
    hasUpdates = true;
  }

  // Handle body updates with operation semantics
  if (canUpdateBody && item.body !== undefined) {
    // Store operation information for consistent footer/append behavior
    // Use operation from item, or fall back to config default, or use the defaultOperation
    const operation = item.operation || config.default_operation || defaultOperation;
    updateData._operation = operation;
    updateData._rawBody = item.body;
    // Note: Some handlers set updateData.body = item.body here (e.g., PR handler)
    // but this is not universal. Callers can add it if needed.
    hasUpdates = true;
  }

  // Handle state/status updates
  // The safe-outputs schema uses "status" (open/closed), while the GitHub API uses "state"
  // Accept both for backwards compatibility if acceptStateAndStatus is true
  if (item.state !== undefined) {
    updateData.state = item.state;
    hasUpdates = true;
  } else if (acceptStateAndStatus && item.status !== undefined) {
    updateData.state = item.status;
    hasUpdates = true;
  }

  // Handle additional entity-specific fields
  for (const fieldName of additionalFields) {
    if (item[fieldName] !== undefined) {
      updateData[fieldName] = item[fieldName];
      hasUpdates = true;
    }
  }

  // If requireUpdates is true and no updates were provided, return a skip result
  if (requireUpdates && !hasUpdates) {
    return {
      success: true,
      skipped: true,
      reason: "No update fields provided or all fields are disabled",
    };
  }

  return { success: true, data: updateData };
}

module.exports = {
  buildUpdatePayloadData,
};
