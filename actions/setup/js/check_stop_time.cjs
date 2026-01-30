// @ts-check
/// <reference types="@actions/github-script" />

/**
 * Validates date format and returns true if valid
 * @param {Date} date - Date to validate
 * @returns {boolean} True if date is valid
 */
function isValidDate(date) {
  return !isNaN(date.getTime());
}

/**
 * Checks if workflow execution should stop based on configured stop time
 * @returns {Promise<void>}
 */
async function main() {
  const { GH_AW_STOP_TIME: stopTime, GH_AW_WORKFLOW_NAME: workflowName } = process.env;

  if (!stopTime) {
    core.setFailed("Configuration error: GH_AW_STOP_TIME not specified.");
    return;
  }

  if (!workflowName) {
    core.setFailed("Configuration error: GH_AW_WORKFLOW_NAME not specified.");
    return;
  }

  core.info(`Checking stop-time limit: ${stopTime}`);

  const stopTimeDate = new Date(stopTime);

  if (!isValidDate(stopTimeDate)) {
    core.setFailed(`Invalid stop-time format: ${stopTime}. Expected format: YYYY-MM-DD HH:MM:SS`);
    return;
  }

  const currentTime = new Date();
  core.info(`Current time: ${currentTime.toISOString()}`);
  core.info(`Stop time: ${stopTimeDate.toISOString()}`);

  if (currentTime >= stopTimeDate) {
    core.warning(`⏰ Stop time reached. Workflow execution will be prevented by activation job.`);
    core.setOutput("stop_time_ok", "false");
    return;
  }

  core.setOutput("stop_time_ok", "true");
}

module.exports = { main };
