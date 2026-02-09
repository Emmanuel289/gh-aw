import { describe, it, expect } from "vitest";
import { buildUpdatePayloadData } from "./update_payload_builder.cjs";

describe("update_payload_builder.cjs", () => {
  describe("buildUpdatePayloadData - basic fields", () => {
    it("should build payload with title only", () => {
      const item = { title: "New Title" };
      const config = {};
      const payloadConfig = {};

      const result = buildUpdatePayloadData(item, config, payloadConfig);

      expect(result.success).toBe(true);
      expect(result.data).toEqual({ title: "New Title" });
    });

    it("should build payload with body and default operation", () => {
      const item = { body: "New body content" };
      const config = {};
      const payloadConfig = { defaultOperation: "append" };

      const result = buildUpdatePayloadData(item, config, payloadConfig);

      expect(result.success).toBe(true);
      expect(result.data._operation).toBe("append");
      expect(result.data._rawBody).toBe("New body content");
    });

    it("should build payload with title and body", () => {
      const item = {
        title: "New Title",
        body: "New body content",
      };
      const config = {};
      const payloadConfig = { defaultOperation: "replace" };

      const result = buildUpdatePayloadData(item, config, payloadConfig);

      expect(result.success).toBe(true);
      expect(result.data.title).toBe("New Title");
      expect(result.data._operation).toBe("replace");
      expect(result.data._rawBody).toBe("New body content");
    });

    it("should build payload with state field", () => {
      const item = { state: "closed" };
      const config = {};
      const payloadConfig = {};

      const result = buildUpdatePayloadData(item, config, payloadConfig);

      expect(result.success).toBe(true);
      expect(result.data.state).toBe("closed");
    });
  });

  describe("buildUpdatePayloadData - operation handling", () => {
    it("should use operation from item if provided", () => {
      const item = {
        body: "New content",
        operation: "prepend",
      };
      const config = {};
      const payloadConfig = { defaultOperation: "append" };

      const result = buildUpdatePayloadData(item, config, payloadConfig);

      expect(result.success).toBe(true);
      expect(result.data._operation).toBe("prepend");
    });

    it("should use config.default_operation if item.operation is not provided", () => {
      const item = { body: "New content" };
      const config = { default_operation: "replace-island" };
      const payloadConfig = { defaultOperation: "append" };

      const result = buildUpdatePayloadData(item, config, payloadConfig);

      expect(result.success).toBe(true);
      expect(result.data._operation).toBe("replace-island");
    });

    it("should fall back to payloadConfig.defaultOperation", () => {
      const item = { body: "New content" };
      const config = {};
      const payloadConfig = { defaultOperation: "replace" };

      const result = buildUpdatePayloadData(item, config, payloadConfig);

      expect(result.success).toBe(true);
      expect(result.data._operation).toBe("replace");
    });
  });

  describe("buildUpdatePayloadData - state and status compatibility", () => {
    it("should accept state field regardless of acceptStateAndStatus", () => {
      const item = { state: "open" };
      const config = {};
      const payloadConfig = { acceptStateAndStatus: false };

      const result = buildUpdatePayloadData(item, config, payloadConfig);

      expect(result.success).toBe(true);
      expect(result.data.state).toBe("open");
    });

    it("should accept status field when acceptStateAndStatus is true", () => {
      const item = { status: "closed" };
      const config = {};
      const payloadConfig = { acceptStateAndStatus: true };

      const result = buildUpdatePayloadData(item, config, payloadConfig);

      expect(result.success).toBe(true);
      expect(result.data.state).toBe("closed");
    });

    it("should ignore status field when acceptStateAndStatus is false", () => {
      const item = { status: "closed" };
      const config = {};
      const payloadConfig = { acceptStateAndStatus: false };

      const result = buildUpdatePayloadData(item, config, payloadConfig);

      expect(result.success).toBe(true);
      expect(result.data).toEqual({});
    });

    it("should prefer state over status when both are provided", () => {
      const item = { state: "open", status: "closed" };
      const config = {};
      const payloadConfig = { acceptStateAndStatus: true };

      const result = buildUpdatePayloadData(item, config, payloadConfig);

      expect(result.success).toBe(true);
      expect(result.data.state).toBe("open");
    });
  });

  describe("buildUpdatePayloadData - additional fields", () => {
    it("should copy additional fields from item to updateData", () => {
      const item = {
        title: "Title",
        labels: ["bug", "enhancement"],
        assignees: ["user1", "user2"],
        milestone: 5,
      };
      const config = {};
      const payloadConfig = {
        additionalFields: ["labels", "assignees", "milestone"],
      };

      const result = buildUpdatePayloadData(item, config, payloadConfig);

      expect(result.success).toBe(true);
      expect(result.data.title).toBe("Title");
      expect(result.data.labels).toEqual(["bug", "enhancement"]);
      expect(result.data.assignees).toEqual(["user1", "user2"]);
      expect(result.data.milestone).toBe(5);
    });

    it("should only copy additional fields that are present in item", () => {
      const item = {
        title: "Title",
        labels: ["bug"],
        // assignees and milestone not provided
      };
      const config = {};
      const payloadConfig = {
        additionalFields: ["labels", "assignees", "milestone"],
      };

      const result = buildUpdatePayloadData(item, config, payloadConfig);

      expect(result.success).toBe(true);
      expect(result.data.title).toBe("Title");
      expect(result.data.labels).toEqual(["bug"]);
      expect(result.data).not.toHaveProperty("assignees");
      expect(result.data).not.toHaveProperty("milestone");
    });

    it("should handle base field for pull requests", () => {
      const item = {
        title: "PR Title",
        base: "develop",
      };
      const config = {};
      const payloadConfig = {
        additionalFields: ["base"],
      };

      const result = buildUpdatePayloadData(item, config, payloadConfig);

      expect(result.success).toBe(true);
      expect(result.data.title).toBe("PR Title");
      expect(result.data.base).toBe("develop");
    });
  });

  describe("buildUpdatePayloadData - allow_title and allow_body config", () => {
    it("should respect config.allow_title = false", () => {
      const item = { title: "New Title", body: "New body" };
      const config = { allow_title: false };
      const payloadConfig = {};

      const result = buildUpdatePayloadData(item, config, payloadConfig);

      expect(result.success).toBe(true);
      expect(result.data).not.toHaveProperty("title");
      expect(result.data._rawBody).toBe("New body");
    });

    it("should respect config.allow_body = false", () => {
      const item = { title: "New Title", body: "New body" };
      const config = { allow_body: false };
      const payloadConfig = {};

      const result = buildUpdatePayloadData(item, config, payloadConfig);

      expect(result.success).toBe(true);
      expect(result.data.title).toBe("New Title");
      expect(result.data).not.toHaveProperty("_rawBody");
      expect(result.data).not.toHaveProperty("_operation");
    });

    it("should respect payloadConfig.allowTitle = false", () => {
      const item = { title: "New Title", body: "New body" };
      const config = {};
      const payloadConfig = { allowTitle: false };

      const result = buildUpdatePayloadData(item, config, payloadConfig);

      expect(result.success).toBe(true);
      expect(result.data).not.toHaveProperty("title");
      expect(result.data._rawBody).toBe("New body");
    });

    it("should respect payloadConfig.allowBody = false", () => {
      const item = { title: "New Title", body: "New body" };
      const config = {};
      const payloadConfig = { allowBody: false };

      const result = buildUpdatePayloadData(item, config, payloadConfig);

      expect(result.success).toBe(true);
      expect(result.data.title).toBe("New Title");
      expect(result.data).not.toHaveProperty("_rawBody");
      expect(result.data).not.toHaveProperty("_operation");
    });

    it("should combine payloadConfig.allowTitle and config.allow_title", () => {
      const item = { title: "New Title" };
      const config = { allow_title: false };
      const payloadConfig = { allowTitle: true }; // This should still be respected

      const result = buildUpdatePayloadData(item, config, payloadConfig);

      expect(result.success).toBe(true);
      expect(result.data).not.toHaveProperty("title");
    });
  });

  describe("buildUpdatePayloadData - requireUpdates option", () => {
    it("should return skipped result when requireUpdates is true and no updates provided", () => {
      const item = {};
      const config = {};
      const payloadConfig = { requireUpdates: true };

      const result = buildUpdatePayloadData(item, config, payloadConfig);

      expect(result.success).toBe(true);
      expect(result.skipped).toBe(true);
      expect(result.reason).toBe("No update fields provided or all fields are disabled");
    });

    it("should return skipped result when all updates are disabled", () => {
      const item = { title: "New Title", body: "New body" };
      const config = { allow_title: false, allow_body: false };
      const payloadConfig = { requireUpdates: true };

      const result = buildUpdatePayloadData(item, config, payloadConfig);

      expect(result.success).toBe(true);
      expect(result.skipped).toBe(true);
    });

    it("should not return skipped result when requireUpdates is false", () => {
      const item = {};
      const config = {};
      const payloadConfig = { requireUpdates: false };

      const result = buildUpdatePayloadData(item, config, payloadConfig);

      expect(result.success).toBe(true);
      expect(result.skipped).toBeUndefined();
      expect(result.data).toEqual({});
    });

    it("should not return skipped result when at least one update is provided", () => {
      const item = { title: "New Title" };
      const config = {};
      const payloadConfig = { requireUpdates: true };

      const result = buildUpdatePayloadData(item, config, payloadConfig);

      expect(result.success).toBe(true);
      expect(result.skipped).toBeUndefined();
      expect(result.data.title).toBe("New Title");
    });
  });

  describe("buildUpdatePayloadData - comprehensive scenarios", () => {
    it("should build issue-like payload with all fields", () => {
      const item = {
        title: "Issue Title",
        body: "Issue body",
        state: "closed",
        labels: ["bug", "wontfix"],
        assignees: ["octocat"],
        milestone: 3,
      };
      const config = {};
      const payloadConfig = {
        defaultOperation: "append",
        acceptStateAndStatus: true,
        additionalFields: ["labels", "assignees", "milestone"],
      };

      const result = buildUpdatePayloadData(item, config, payloadConfig);

      expect(result.success).toBe(true);
      expect(result.data.title).toBe("Issue Title");
      expect(result.data._operation).toBe("append");
      expect(result.data._rawBody).toBe("Issue body");
      expect(result.data.state).toBe("closed");
      expect(result.data.labels).toEqual(["bug", "wontfix"]);
      expect(result.data.assignees).toEqual(["octocat"]);
      expect(result.data.milestone).toBe(3);
    });

    it("should build PR-like payload with allow_title/allow_body config", () => {
      const item = {
        title: "PR Title",
        body: "PR description",
        state: "open",
        base: "main",
      };
      const config = {
        allow_title: true,
        allow_body: true,
        default_operation: "replace",
      };
      const payloadConfig = {
        defaultOperation: "append", // Should be overridden by config.default_operation
        additionalFields: ["base"],
        requireUpdates: true,
      };

      const result = buildUpdatePayloadData(item, config, payloadConfig);

      expect(result.success).toBe(true);
      expect(result.skipped).toBeUndefined();
      expect(result.data.title).toBe("PR Title");
      expect(result.data._operation).toBe("replace");
      expect(result.data._rawBody).toBe("PR description");
      expect(result.data.state).toBe("open");
      expect(result.data.base).toBe("main");
    });

    it("should build discussion-like payload with minimal fields", () => {
      const item = {
        title: "Discussion Title",
        body: "Discussion content",
      };
      const config = {};
      const payloadConfig = {
        // Discussions don't use operations, but we can still pass a default
        defaultOperation: "replace",
        acceptStateAndStatus: false,
        additionalFields: [],
      };

      const result = buildUpdatePayloadData(item, config, payloadConfig);

      expect(result.success).toBe(true);
      expect(result.data.title).toBe("Discussion Title");
      expect(result.data._operation).toBe("replace");
      expect(result.data._rawBody).toBe("Discussion content");
      // No state, labels, assignees, etc.
    });
  });
});
