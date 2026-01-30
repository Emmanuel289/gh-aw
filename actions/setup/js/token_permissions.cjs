// @ts-check
/// <reference types="@actions/github-script" />

/**
 * Token Permissions Validation Module
 *
 * This module provides utilities to check GitHub token scopes and permissions
 * before executing safe output operations. It validates that tokens have the
 * required permissions to perform specific GitHub API operations.
 *
 * Supports both classic personal access tokens (OAuth scopes) and fine-grained
 * personal access tokens (repository permissions).
 */

const { getErrorMessage } = require("./error_helpers.cjs");

/**
 * Required permissions for each safe output operation type
 * Maps operation types to their permission requirements
 */
const OPERATION_PERMISSIONS = {
  create_issue: {
    required: ["issues:write"],
    description: "Create issues",
  },
  update_issue: {
    required: ["issues:write"],
    description: "Update issues",
  },
  close_issue: {
    required: ["issues:write"],
    description: "Close issues",
  },
  add_comment: {
    required: ["issues:write", "pull_requests:write"],
    requiresAny: true, // Only need one of these
    description: "Add comments to issues or pull requests",
  },
  hide_comment: {
    required: ["issues:write", "pull_requests:write"],
    requiresAny: true,
    description: "Hide comments on issues or pull requests",
  },
  add_labels: {
    required: ["issues:write"],
    description: "Add labels to issues or pull requests",
  },
  remove_labels: {
    required: ["issues:write"],
    description: "Remove labels from issues or pull requests",
  },
  assign_milestone: {
    required: ["issues:write"],
    description: "Assign milestones to issues",
  },
  assign_to_user: {
    required: ["issues:write"],
    description: "Assign users to issues",
  },
  assign_to_agent: {
    required: ["issues:write"],
    description: "Assign Copilot agents to issues",
  },
  add_reviewer: {
    required: ["pull_requests:write"],
    description: "Request pull request reviews",
  },
  link_sub_issue: {
    required: ["issues:write"],
    description: "Link sub-issues to parent issues",
  },
  create_pull_request: {
    required: ["pull_requests:write", "contents:write"],
    description: "Create pull requests",
  },
  update_pull_request: {
    required: ["pull_requests:write"],
    description: "Update pull requests",
  },
  close_pull_request: {
    required: ["pull_requests:write"],
    description: "Close pull requests",
  },
  mark_pull_request_as_ready_for_review: {
    required: ["pull_requests:write"],
    description: "Mark pull requests as ready for review",
  },
  create_discussion: {
    required: ["discussions:write"],
    description: "Create discussions",
  },
  update_discussion: {
    required: ["discussions:write"],
    description: "Update discussions",
  },
  close_discussion: {
    required: ["discussions:write"],
    description: "Close discussions",
  },
  create_project: {
    required: ["projects:write"],
    description: "Create GitHub Projects",
  },
  update_project: {
    required: ["projects:write"],
    optional: ["issues:write"], // Optional for label operations
    description: "Update GitHub Projects",
  },
  copy_project: {
    required: ["projects:write"],
    description: "Copy GitHub Projects",
  },
  create_project_status_update: {
    required: ["projects:write"],
    description: "Create project status updates",
  },
  update_release: {
    required: ["contents:write"],
    description: "Update releases",
  },
  upload_assets: {
    required: ["contents:write"],
    description: "Upload assets to orphaned branches",
  },
};

/**
 * Check if a token has the required scopes for classic PATs
 * Returns information about available scopes from x-oauth-scopes header
 * @param {string} token - GitHub token to check
 * @param {string} owner - Repository owner
 * @param {string} repo - Repository name
 * @returns {Promise<{hasScopes: boolean, scopes: string[], error?: string}>}
 */
async function checkClassicTokenScopes(token, owner, repo) {
  try {
    // Make a lightweight API call to get token scopes from headers
    const response = await github.rest.repos.get({
      owner,
      repo,
    });

    // Extract scopes from response headers
    // The x-oauth-scopes header contains comma-separated list of scopes
    const scopesHeader = response.headers["x-oauth-scopes"] || "";
    const scopes = scopesHeader
      .split(",")
      .map(s => s.trim())
      .filter(s => s.length > 0);

    core.debug(`Classic token scopes: ${scopes.join(", ")}`);

    return {
      hasScopes: scopes.length > 0,
      scopes,
    };
  } catch (error) {
    return {
      hasScopes: false,
      scopes: [],
      error: getErrorMessage(error),
    };
  }
}

/**
 * Check if a token has required permissions for fine-grained PATs
 * Fine-grained tokens use repository permissions instead of scopes
 * @param {string} token - GitHub token to check
 * @param {string} owner - Repository owner
 * @param {string} repo - Repository name
 * @returns {Promise<{permissions: Object, error?: string}>}
 */
async function checkFineGrainedTokenPermissions(token, owner, repo) {
  try {
    // Try to access repository metadata which will reveal if we have access
    const repoResponse = await github.rest.repos.get({
      owner,
      repo,
    });

    // For fine-grained tokens, we can infer permissions by checking the x-accepted-github-permissions header
    const permissionsHeader = repoResponse.headers["x-accepted-github-permissions"] || "";

    core.debug(`Fine-grained token permissions header: ${permissionsHeader}`);

    // We'll also try to check specific endpoints to validate permissions
    const permissions = {
      metadata: "read", // We successfully read metadata
      contents: "unknown",
      issues: "unknown",
      pull_requests: "unknown",
      discussions: "unknown",
      projects: "unknown",
    };

    return { permissions };
  } catch (error) {
    return {
      permissions: {},
      error: getErrorMessage(error),
    };
  }
}

/**
 * Map OAuth scopes to permission categories
 * Classic tokens use broad scopes that map to multiple permissions
 * @param {string[]} scopes - Array of OAuth scopes
 * @returns {Object} Map of permission categories to access levels
 */
function mapScopesToPermissions(scopes) {
  const permissions = {
    contents: "none",
    issues: "none",
    pull_requests: "none",
    discussions: "none",
    projects: "none",
  };

  for (const scope of scopes) {
    switch (scope) {
      case "repo":
      case "public_repo":
        // Full repository access includes contents, issues, and PRs
        permissions.contents = "write";
        permissions.issues = "write";
        permissions.pull_requests = "write";
        break;
      case "repo:status":
        // Commit status access
        permissions.contents = "read";
        break;
      case "repo_deployment":
        // Deployment access
        permissions.contents = "read";
        break;
      case "public_repo":
        // Public repository access
        permissions.contents = "write";
        permissions.issues = "write";
        permissions.pull_requests = "write";
        break;
      case "repo:invite":
        // Repository invitations
        break;
      case "security_events":
        // Security events
        break;
      case "write:discussion":
      case "read:discussion":
        permissions.discussions = scope.startsWith("write") ? "write" : "read";
        break;
      case "project":
        permissions.projects = "write";
        break;
      case "read:project":
        permissions.projects = "read";
        break;
    }
  }

  return permissions;
}

/**
 * Validate if permissions meet requirements for a specific operation
 * @param {Object} permissions - Current token permissions
 * @param {string} operationType - Safe output operation type
 * @returns {{valid: boolean, missing: string[], optional: string[], description: string}}
 */
function validateOperationPermissions(permissions, operationType) {
  const operationReq = OPERATION_PERMISSIONS[operationType];

  if (!operationReq) {
    // Unknown operation type - allow it through
    return {
      valid: true,
      missing: [],
      optional: [],
      description: `Unknown operation type: ${operationType}`,
    };
  }

  const missing = [];
  const optional = [];

  // Check required permissions
  if (operationReq.requiresAny) {
    // Need at least one of the required permissions
    const hasAny = operationReq.required.some(req => {
      const [category, level] = req.split(":");
      return permissions[category] === level || permissions[category] === "write";
    });

    if (!hasAny) {
      missing.push(...operationReq.required);
    }
  } else {
    // Need all required permissions
    for (const req of operationReq.required) {
      const [category, level] = req.split(":");
      const currentLevel = permissions[category] || "none";

      if (currentLevel === "none" || (level === "write" && currentLevel !== "write")) {
        missing.push(req);
      }
    }
  }

  // Check optional permissions (permissions that enable extra features)
  if (operationReq.optional && Array.isArray(operationReq.optional)) {
    const optionalPerms = operationReq.optional;

    for (const opt of optionalPerms) {
      const [category, level] = opt.split(":");
      const currentLevel = permissions[category] || "none";

      if (currentLevel === "none" || (level === "write" && currentLevel !== "write")) {
        optional.push(opt);
      }
    }
  }

  return {
    valid: missing.length === 0,
    missing,
    optional,
    description: operationReq.description,
  };
}

/**
 * Validate token permissions for a list of safe output operations
 * @param {string} token - GitHub token to validate
 * @param {string} owner - Repository owner
 * @param {string} repo - Repository name
 * @param {string[]} operationTypes - Array of operation types to validate
 * @returns {Promise<{valid: boolean, results: Array<Object>, permissions: Object, tokenType: string}>}
 */
async function validateTokenPermissions(token, owner, repo, operationTypes) {
  core.info(`Validating token permissions for ${operationTypes.length} operation type(s)...`);

  // First, check what type of token we have and what scopes/permissions it has
  const scopeCheck = await checkClassicTokenScopes(token, owner, repo);

  let permissions;
  let tokenType;

  if (scopeCheck.hasScopes && scopeCheck.scopes.length > 0) {
    // Classic token with OAuth scopes
    tokenType = "classic";
    permissions = mapScopesToPermissions(scopeCheck.scopes);
    core.info(`Detected classic token with scopes: ${scopeCheck.scopes.join(", ")}`);
  } else {
    // Likely a fine-grained token or GITHUB_TOKEN
    tokenType = "fine-grained";
    const permCheck = await checkFineGrainedTokenPermissions(token, owner, repo);
    permissions = permCheck.permissions;
    core.info(`Detected fine-grained or GITHUB_TOKEN token`);
  }

  // Validate each operation type
  const results = [];
  let allValid = true;

  for (const operationType of operationTypes) {
    const validation = validateOperationPermissions(permissions, operationType);

    if (!validation.valid) {
      allValid = false;
    }

    results.push({
      operationType,
      ...validation,
    });
  }

  return {
    valid: allValid,
    results,
    permissions,
    tokenType,
  };
}

/**
 * Generate a user-friendly error message for missing permissions
 * @param {Array<Object>} validationResults - Results from validateTokenPermissions
 * @param {string} tokenType - Type of token (classic or fine-grained)
 * @returns {string} Formatted error message with remediation steps
 */
function generatePermissionErrorMessage(validationResults, tokenType) {
  const failedOperations = validationResults.filter(r => !r.valid);

  if (failedOperations.length === 0) {
    return "";
  }

  const lines = ["❌ Token Missing Required Permissions", "", "The GitHub token lacks permissions required for the following operations:", ""];

  for (const op of failedOperations) {
    lines.push(`• ${op.description} (${op.operationType})`);
    lines.push(`  Missing: ${op.missing.join(", ")}`);
    if (op.optional && op.optional.length > 0) {
      lines.push(`  Optional (degraded functionality): ${op.optional.join(", ")}`);
    }
  }

  lines.push("");
  lines.push("📋 Remediation Steps:");
  lines.push("");

  if (tokenType === "classic") {
    lines.push("For Classic Personal Access Tokens:");
    lines.push("1. Go to https://github.com/settings/tokens");
    lines.push("2. Edit your token or create a new one");
    lines.push("3. Enable the following scopes:");

    const requiredScopes = new Set();
    for (const op of failedOperations) {
      for (const perm of op.missing) {
        if (perm.includes("issues") || perm.includes("pull_requests")) {
          requiredScopes.add("repo (full repository access)");
        } else if (perm.includes("discussions")) {
          requiredScopes.add("write:discussion");
        } else if (perm.includes("projects")) {
          requiredScopes.add("project");
        } else if (perm.includes("contents")) {
          requiredScopes.add("repo");
        }
      }
    }

    for (const scope of requiredScopes) {
      lines.push(`   - ${scope}`);
    }
  } else {
    lines.push("For Fine-Grained Personal Access Tokens:");
    lines.push("1. Go to https://github.com/settings/personal-access-tokens/new");
    lines.push("2. Select the target repository");
    lines.push("3. Grant the following permissions:");

    const requiredPerms = new Set();
    for (const op of failedOperations) {
      for (const perm of op.missing) {
        requiredPerms.add(perm);
      }
    }

    for (const perm of requiredPerms) {
      const [category, level] = perm.split(":");
      lines.push(`   - ${category.charAt(0).toUpperCase() + category.slice(1)}: Read and Write`);
    }
  }

  lines.push("");
  lines.push("4. Update your workflow secret with the new token");
  lines.push("");
  lines.push("📚 Documentation:");
  lines.push("- Classic tokens: https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens");
  lines.push("- Fine-grained tokens: https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token");

  return lines.join("\n");
}

/**
 * Generate a user-friendly warning message for missing optional permissions
 * @param {Array<Object>} validationResults - Results from validateTokenPermissions
 * @returns {string} Formatted warning message
 */
function generatePermissionWarningMessage(validationResults) {
  const operationsWithOptional = validationResults.filter(r => r.optional && r.optional.length > 0);

  if (operationsWithOptional.length === 0) {
    return "";
  }

  const lines = ["⚠️  Optional Permissions Missing (Degraded Functionality)", "", "The following operations may have reduced functionality:", ""];

  for (const op of operationsWithOptional) {
    lines.push(`• ${op.description} (${op.operationType})`);
    lines.push(`  Optional: ${op.optional.join(", ")}`);
  }

  return lines.join("\n");
}

module.exports = {
  validateTokenPermissions,
  validateOperationPermissions,
  generatePermissionErrorMessage,
  generatePermissionWarningMessage,
  OPERATION_PERMISSIONS,
};
