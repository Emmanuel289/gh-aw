// @ts-check
import { describe, it, expect, vi, beforeEach } from "vitest";

describe("token_permissions", () => {
  let validateOperationPermissions;
  let generatePermissionErrorMessage;
  let generatePermissionWarningMessage;
  let OPERATION_PERMISSIONS;

  beforeEach(async () => {
    // Dynamically import the module
    const module = await import("./token_permissions.cjs");
    validateOperationPermissions = module.validateOperationPermissions;
    generatePermissionErrorMessage = module.generatePermissionErrorMessage;
    generatePermissionWarningMessage = module.generatePermissionWarningMessage;
    OPERATION_PERMISSIONS = module.OPERATION_PERMISSIONS;
  });

  describe("validateOperationPermissions", () => {
    it("should validate create_issue with sufficient permissions", () => {
      const permissions = {
        issues: "write",
        contents: "read",
      };

      const result = validateOperationPermissions(permissions, "create_issue");

      expect(result.valid).toBe(true);
      expect(result.missing).toEqual([]);
      expect(result.description).toContain("Create issues");
    });

    it("should fail create_issue with insufficient permissions", () => {
      const permissions = {
        issues: "read", // Need write
        contents: "read",
      };

      const result = validateOperationPermissions(permissions, "create_issue");

      expect(result.valid).toBe(false);
      expect(result.missing).toContain("issues:write");
    });

    it("should validate add_comment with issues:write permission", () => {
      const permissions = {
        issues: "write",
        pull_requests: "none",
      };

      const result = validateOperationPermissions(permissions, "add_comment");

      expect(result.valid).toBe(true);
      expect(result.missing).toEqual([]);
    });

    it("should validate add_comment with pull_requests:write permission", () => {
      const permissions = {
        issues: "none",
        pull_requests: "write",
      };

      const result = validateOperationPermissions(permissions, "add_comment");

      expect(result.valid).toBe(true);
      expect(result.missing).toEqual([]);
    });

    it("should fail add_comment without either permission", () => {
      const permissions = {
        issues: "none",
        pull_requests: "none",
      };

      const result = validateOperationPermissions(permissions, "add_comment");

      expect(result.valid).toBe(false);
      expect(result.missing).toContain("issues:write");
      expect(result.missing).toContain("pull_requests:write");
    });

    it("should validate create_pull_request with both required permissions", () => {
      const permissions = {
        pull_requests: "write",
        contents: "write",
      };

      const result = validateOperationPermissions(permissions, "create_pull_request");

      expect(result.valid).toBe(true);
      expect(result.missing).toEqual([]);
    });

    it("should fail create_pull_request with only one permission", () => {
      const permissions = {
        pull_requests: "write",
        contents: "read", // Need write
      };

      const result = validateOperationPermissions(permissions, "create_pull_request");

      expect(result.valid).toBe(false);
      expect(result.missing).toContain("contents:write");
    });

    it("should handle update_project with optional permissions", () => {
      const permissions = {
        projects: "write",
        issues: "none", // Optional for labels
      };

      const result = validateOperationPermissions(permissions, "update_project");

      expect(result.valid).toBe(true);
      expect(result.missing).toEqual([]);
      expect(result.optional).toContain("issues:write");
    });

    it("should validate create_project with projects:write", () => {
      const permissions = {
        projects: "write",
      };

      const result = validateOperationPermissions(permissions, "create_project");

      expect(result.valid).toBe(true);
      expect(result.missing).toEqual([]);
    });

    it("should fail create_project without projects:write", () => {
      const permissions = {
        projects: "read", // Need write
        contents: "write",
      };

      const result = validateOperationPermissions(permissions, "create_project");

      expect(result.valid).toBe(false);
      expect(result.missing).toContain("projects:write");
    });

    it("should handle unknown operation types gracefully", () => {
      const permissions = {
        issues: "write",
      };

      const result = validateOperationPermissions(permissions, "unknown_operation");

      expect(result.valid).toBe(true);
      expect(result.missing).toEqual([]);
      expect(result.description).toContain("Unknown operation type");
    });

    it("should validate close_issue with issues:write", () => {
      const permissions = {
        issues: "write",
      };

      const result = validateOperationPermissions(permissions, "close_issue");

      expect(result.valid).toBe(true);
      expect(result.missing).toEqual([]);
    });

    it("should validate discussions operations", () => {
      const permissions = {
        discussions: "write",
      };

      expect(validateOperationPermissions(permissions, "create_discussion").valid).toBe(true);
      expect(validateOperationPermissions(permissions, "update_discussion").valid).toBe(true);
      expect(validateOperationPermissions(permissions, "close_discussion").valid).toBe(true);
    });

    it("should validate add_labels with issues:write", () => {
      const permissions = {
        issues: "write",
      };

      const result = validateOperationPermissions(permissions, "add_labels");

      expect(result.valid).toBe(true);
      expect(result.missing).toEqual([]);
    });
  });

  describe("generatePermissionErrorMessage", () => {
    it("should generate error message for failed operations", () => {
      const validationResults = [
        {
          operationType: "create_issue",
          valid: false,
          missing: ["issues:write"],
          optional: [],
          description: "Create issues",
        },
        {
          operationType: "create_pull_request",
          valid: false,
          missing: ["pull_requests:write", "contents:write"],
          optional: [],
          description: "Create pull requests",
        },
      ];

      const message = generatePermissionErrorMessage(validationResults, "classic");

      expect(message).toContain("Token Missing Required Permissions");
      expect(message).toContain("Create issues");
      expect(message).toContain("Create pull requests");
      expect(message).toContain("Remediation Steps");
      expect(message).toContain("Classic Personal Access Tokens");
      expect(message).toContain("repo");
    });

    it("should generate error message for fine-grained tokens", () => {
      const validationResults = [
        {
          operationType: "create_issue",
          valid: false,
          missing: ["issues:write"],
          optional: [],
          description: "Create issues",
        },
      ];

      const message = generatePermissionErrorMessage(validationResults, "fine-grained");

      expect(message).toContain("Fine-Grained Personal Access Tokens");
      expect(message).toContain("Issues: Read and Write");
    });

    it("should include optional permissions in error message", () => {
      const validationResults = [
        {
          operationType: "update_project",
          valid: false,
          missing: ["projects:write"],
          optional: ["issues:write"],
          description: "Update GitHub Projects",
        },
      ];

      const message = generatePermissionErrorMessage(validationResults, "classic");

      expect(message).toContain("Missing: projects:write");
      expect(message).toContain("Optional (degraded functionality): issues:write");
    });

    it("should return empty string when all operations are valid", () => {
      const validationResults = [
        {
          operationType: "create_issue",
          valid: true,
          missing: [],
          optional: [],
          description: "Create issues",
        },
      ];

      const message = generatePermissionErrorMessage(validationResults, "classic");

      expect(message).toBe("");
    });

    it("should include project scope for project operations", () => {
      const validationResults = [
        {
          operationType: "create_project",
          valid: false,
          missing: ["projects:write"],
          optional: [],
          description: "Create GitHub Projects",
        },
      ];

      const message = generatePermissionErrorMessage(validationResults, "classic");

      expect(message).toContain("project");
    });

    it("should include discussion scope for discussion operations", () => {
      const validationResults = [
        {
          operationType: "create_discussion",
          valid: false,
          missing: ["discussions:write"],
          optional: [],
          description: "Create discussions",
        },
      ];

      const message = generatePermissionErrorMessage(validationResults, "classic");

      expect(message).toContain("write:discussion");
    });
  });

  describe("generatePermissionWarningMessage", () => {
    it("should generate warning for operations with optional permissions", () => {
      const validationResults = [
        {
          operationType: "update_project",
          valid: true,
          missing: [],
          optional: ["issues:write"],
          description: "Update GitHub Projects",
        },
      ];

      const message = generatePermissionWarningMessage(validationResults);

      expect(message).toContain("Optional Permissions Missing");
      expect(message).toContain("Update GitHub Projects");
      expect(message).toContain("issues:write");
    });

    it("should return empty string when no optional permissions are missing", () => {
      const validationResults = [
        {
          operationType: "create_issue",
          valid: true,
          missing: [],
          optional: [],
          description: "Create issues",
        },
      ];

      const message = generatePermissionWarningMessage(validationResults);

      expect(message).toBe("");
    });

    it("should handle multiple operations with optional permissions", () => {
      const validationResults = [
        {
          operationType: "update_project",
          valid: true,
          missing: [],
          optional: ["issues:write"],
          description: "Update GitHub Projects",
        },
        {
          operationType: "add_labels",
          valid: true,
          missing: [],
          optional: ["pull_requests:write"],
          description: "Add labels",
        },
      ];

      const message = generatePermissionWarningMessage(validationResults);

      expect(message).toContain("Update GitHub Projects");
      expect(message).toContain("Add labels");
    });
  });

  describe("OPERATION_PERMISSIONS constant", () => {
    it("should define permissions for all safe output operations", () => {
      const expectedOperations = ["create_issue", "update_issue", "close_issue", "add_comment", "create_pull_request", "update_pull_request", "create_project", "update_project"];

      for (const op of expectedOperations) {
        expect(OPERATION_PERMISSIONS[op]).toBeDefined();
        expect(OPERATION_PERMISSIONS[op].required).toBeDefined();
        expect(OPERATION_PERMISSIONS[op].description).toBeDefined();
      }
    });

    it("should have requiresAny flag for add_comment", () => {
      expect(OPERATION_PERMISSIONS.add_comment.requiresAny).toBe(true);
    });

    it("should have optional permissions for update_project", () => {
      expect(OPERATION_PERMISSIONS.update_project.optional).toBeDefined();
      expect(OPERATION_PERMISSIONS.update_project.optional).toContain("issues:write");
    });
  });
});
