// @ts-check

/**
 * Neutralize GitHub Actions workflow commands in text by escaping double colons.
 * This prevents injection of commands like ::set-output::, ::warning::, ::error::, etc.
 *
 * GitHub Actions workflow commands have the format:
 * ::command parameter1={data},parameter2={data}::{command value}
 *
 * IMPORTANT: Workflow commands are only recognized when :: appears at the start of a line.
 * This function only neutralizes :: at line boundaries, preserving :: in normal text
 * (e.g., "12:30", "C++::function", "IPv6 ::1").
 *
 * By replacing :: with :\u200B: (zero-width space) only at line starts, we prevent
 * command injection while maintaining readability and not affecting legitimate uses.
 *
 * @param {string} text - The text to neutralize
 * @returns {string} The neutralized text
 */
function neutralizeWorkflowCommands(text) {
  if (typeof text !== "string") {
    return String(text);
  }
  // Replace :: only at the start of a line (after start or newline)
  // This matches GitHub Actions' actual command parsing behavior
  // Preserves :: in normal text like "12:30", "C++::function", etc.
  return text.replace(/^::/gm, ":\u200B:").replace(/\n::/g, "\n:\u200B:");
}

/**
 * Sanitized wrapper for core.info that neutralizes workflow commands in user-generated content.
 * Use this instead of core.info() when logging user-generated text that might contain
 * malicious workflow commands.
 *
 * @param {string} message - The message to log (will be sanitized)
 */
function safeInfo(message) {
  core.info(neutralizeWorkflowCommands(message));
}

/**
 * Sanitized wrapper for core.debug that neutralizes workflow commands in user-generated content.
 * Use this instead of core.debug() when logging user-generated text that might contain
 * malicious workflow commands.
 *
 * @param {string} message - The message to log (will be sanitized)
 */
function safeDebug(message) {
  core.debug(neutralizeWorkflowCommands(message));
}

/**
 * Sanitized wrapper for core.warning that neutralizes workflow commands in user-generated content.
 * Use this instead of core.warning() when logging user-generated text that might contain
 * malicious workflow commands.
 *
 * @param {string} message - The message to log (will be sanitized)
 */
function safeWarning(message) {
  core.warning(neutralizeWorkflowCommands(message));
}

/**
 * Sanitized wrapper for core.error that neutralizes workflow commands in user-generated content.
 * Use this instead of core.error() when logging user-generated text that might contain
 * malicious workflow commands.
 *
 * @param {string} message - The message to log (will be sanitized)
 */
function safeError(message) {
  core.error(neutralizeWorkflowCommands(message));
}

module.exports = {
  neutralizeWorkflowCommands,
  safeInfo,
  safeDebug,
  safeWarning,
  safeError,
};
