import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import fs from "fs";
import path from "path";
import os from "os";

describe("create_footer.cjs", () => {
  let readAwInfo;
  let createFooterInfo;
  let generateInfoFooter;
  let formatCost;
  let testTmpDir;
  let originalAwInfoPath;

  beforeEach(async () => {
    // Set up test tmp directory
    testTmpDir = path.join(os.tmpdir(), "gh-aw-test-create-footer");
    if (!fs.existsSync(testTmpDir)) {
      fs.mkdirSync(testTmpDir, { recursive: true });
    }

    // Dynamic import to get fresh module state
    const module = await import("./create_footer.cjs");
    readAwInfo = module.readAwInfo;
    createFooterInfo = module.createFooterInfo;
    generateInfoFooter = module.generateInfoFooter;
    formatCost = module.formatCost;
  });

  afterEach(() => {
    // Clean up test directory
    if (testTmpDir && fs.existsSync(testTmpDir)) {
      fs.rmSync(testTmpDir, { recursive: true, force: true });
    }
  });

  describe("formatCost", () => {
    it("should format small costs with 4 decimal places", () => {
      expect(formatCost(0.0012)).toBe("$0.0012");
      expect(formatCost(0.0099)).toBe("$0.0099");
      expect(formatCost(0.0001)).toBe("$0.0001");
    });

    it("should format larger costs with 2 decimal places", () => {
      expect(formatCost(0.01)).toBe("$0.01");
      expect(formatCost(0.15)).toBe("$0.15");
      expect(formatCost(1.5)).toBe("$1.50");
      expect(formatCost(10.25)).toBe("$10.25");
    });

    it("should handle zero cost", () => {
      expect(formatCost(0)).toBe("$0.0000");
    });

    it("should handle very small costs", () => {
      expect(formatCost(0.00001)).toBe("$0.0000");
    });
  });

  describe("readAwInfo", () => {
    it("should return null when aw_info.json does not exist", () => {
      const result = readAwInfo();
      // Since /tmp/gh-aw/aw_info.json likely doesn't exist in test env
      expect(result).toBeNull();
    });

    it("should parse valid aw_info.json if it exists", () => {
      // Create a test aw_info.json file
      const awInfoDir = "/tmp/gh-aw";
      const awInfoPath = path.join(awInfoDir, "aw_info.json");

      try {
        // Ensure directory exists
        if (!fs.existsSync(awInfoDir)) {
          fs.mkdirSync(awInfoDir, { recursive: true });
        }

        const testData = {
          engine_id: "copilot",
          engine_name: "GitHub Copilot",
          model: "gpt-5",
          version: "1.0.0",
          workflow_name: "Test Workflow",
        };

        fs.writeFileSync(awInfoPath, JSON.stringify(testData, null, 2), "utf8");

        const result = readAwInfo();
        expect(result).toEqual(testData);
      } finally {
        // Clean up
        if (fs.existsSync(awInfoPath)) {
          fs.unlinkSync(awInfoPath);
        }
      }
    });

    it("should return null for invalid JSON", () => {
      const awInfoDir = "/tmp/gh-aw";
      const awInfoPath = path.join(awInfoDir, "aw_info.json");

      try {
        if (!fs.existsSync(awInfoDir)) {
          fs.mkdirSync(awInfoDir, { recursive: true });
        }

        fs.writeFileSync(awInfoPath, "invalid json {", "utf8");

        const result = readAwInfo();
        expect(result).toBeNull();
      } finally {
        if (fs.existsSync(awInfoPath)) {
          fs.unlinkSync(awInfoPath);
        }
      }
    });
  });

  describe("createFooterInfo", () => {
    it("should return null when aw_info.json does not exist", () => {
      const result = createFooterInfo();
      expect(result).toBeNull();
    });

    it("should extract basic info from aw_info.json", () => {
      const awInfoDir = "/tmp/gh-aw";
      const awInfoPath = path.join(awInfoDir, "aw_info.json");

      try {
        if (!fs.existsSync(awInfoDir)) {
          fs.mkdirSync(awInfoDir, { recursive: true });
        }

        const testData = {
          engine_id: "copilot",
          engine_name: "GitHub Copilot",
          model: "gpt-5",
          version: "1.0.0",
          workflow_name: "Test Workflow",
        };

        fs.writeFileSync(awInfoPath, JSON.stringify(testData, null, 2), "utf8");

        const result = createFooterInfo();
        expect(result).toBeDefined();
        expect(result.agent).toBe("GitHub Copilot");
        expect(result.model).toBe("gpt-5");
      } finally {
        if (fs.existsSync(awInfoPath)) {
          fs.unlinkSync(awInfoPath);
        }
      }
    });

    it("should use engine_id as fallback for agent name", () => {
      const awInfoDir = "/tmp/gh-aw";
      const awInfoPath = path.join(awInfoDir, "aw_info.json");

      try {
        if (!fs.existsSync(awInfoDir)) {
          fs.mkdirSync(awInfoDir, { recursive: true });
        }

        const testData = {
          engine_id: "claude",
          workflow_name: "Test Workflow",
        };

        fs.writeFileSync(awInfoPath, JSON.stringify(testData, null, 2), "utf8");

        const result = createFooterInfo();
        expect(result).toBeDefined();
        expect(result.agent).toBe("claude");
      } finally {
        if (fs.existsSync(awInfoPath)) {
          fs.unlinkSync(awInfoPath);
        }
      }
    });

    it("should not include model if not specified", () => {
      const awInfoDir = "/tmp/gh-aw";
      const awInfoPath = path.join(awInfoDir, "aw_info.json");

      try {
        if (!fs.existsSync(awInfoDir)) {
          fs.mkdirSync(awInfoDir, { recursive: true });
        }

        const testData = {
          engine_id: "copilot",
          engine_name: "GitHub Copilot",
          workflow_name: "Test Workflow",
        };

        fs.writeFileSync(awInfoPath, JSON.stringify(testData, null, 2), "utf8");

        const result = createFooterInfo();
        expect(result).toBeDefined();
        expect(result.agent).toBe("GitHub Copilot");
        expect(result.model).toBeUndefined();
      } finally {
        if (fs.existsSync(awInfoPath)) {
          fs.unlinkSync(awInfoPath);
        }
      }
    });
  });

  describe("generateInfoFooter", () => {
    it("should return empty string when no aw_info.json exists", () => {
      const result = generateInfoFooter();
      expect(result).toBe("");
    });

    it("should generate footer with agent name only", () => {
      const awInfoDir = "/tmp/gh-aw";
      const awInfoPath = path.join(awInfoDir, "aw_info.json");

      try {
        if (!fs.existsSync(awInfoDir)) {
          fs.mkdirSync(awInfoDir, { recursive: true });
        }

        const testData = {
          engine_id: "copilot",
          engine_name: "GitHub Copilot",
          workflow_name: "Test Workflow",
        };

        fs.writeFileSync(awInfoPath, JSON.stringify(testData, null, 2), "utf8");

        const result = generateInfoFooter();
        expect(result).toBe("GitHub Copilot");
      } finally {
        if (fs.existsSync(awInfoPath)) {
          fs.unlinkSync(awInfoPath);
        }
      }
    });

    it("should generate footer with agent and model", () => {
      const awInfoDir = "/tmp/gh-aw";
      const awInfoPath = path.join(awInfoDir, "aw_info.json");

      try {
        if (!fs.existsSync(awInfoDir)) {
          fs.mkdirSync(awInfoDir, { recursive: true });
        }

        const testData = {
          engine_id: "claude",
          engine_name: "Claude",
          model: "claude-sonnet-4.5",
          workflow_name: "Test Workflow",
        };

        fs.writeFileSync(awInfoPath, JSON.stringify(testData, null, 2), "utf8");

        const result = generateInfoFooter();
        expect(result).toBe("Claude, claude-sonnet-4.5");
      } finally {
        if (fs.existsSync(awInfoPath)) {
          fs.unlinkSync(awInfoPath);
        }
      }
    });

    it("should generate single-line format", () => {
      const awInfoDir = "/tmp/gh-aw";
      const awInfoPath = path.join(awInfoDir, "aw_info.json");

      try {
        if (!fs.existsSync(awInfoDir)) {
          fs.mkdirSync(awInfoDir, { recursive: true });
        }

        const testData = {
          engine_id: "copilot",
          engine_name: "GitHub Copilot",
          model: "gpt-5",
          workflow_name: "Test Workflow",
        };

        fs.writeFileSync(awInfoPath, JSON.stringify(testData, null, 2), "utf8");

        const result = generateInfoFooter();
        // Should be single line with comma separation
        expect(result).not.toContain("\n");
        expect(result).toContain(",");
      } finally {
        if (fs.existsSync(awInfoPath)) {
          fs.unlinkSync(awInfoPath);
        }
      }
    });
  });
});
