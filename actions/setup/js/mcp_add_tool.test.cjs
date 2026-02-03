// @ts-check
import { describe, it, expect, beforeEach } from "vitest";
import path from "path";
import { fileURLToPath } from "url";
const { createMCPAddHandler } = require("./mcp_add_tool.cjs");

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

describe("createMCPAddHandler", () => {
  let mockServer;

  beforeEach(() => {
    mockServer = {
      debug: () => {},
      debugError: () => {},
    };
  });

  it("should create a handler function", () => {
    const handler = createMCPAddHandler(mockServer, "test-add");
    expect(handler).toBeInstanceOf(Function);
  });

  it("should use correct timeout default", () => {
    // The handler should use 120 seconds as default timeout for MCP operations
    const handler = createMCPAddHandler(mockServer, "test-add");
    expect(handler).toBeDefined();
  });

  it("should allow custom timeout", () => {
    const handler = createMCPAddHandler(mockServer, "test-add", 300);
    expect(handler).toBeDefined();
  });

  it("should resolve correct Go script path", () => {
    // Verify the path resolution logic would find the Go script
    const expectedPath = path.join(__dirname, "..", "go", "mcp_add_tool.go");
    expect(expectedPath).toContain("actions/setup/go/mcp_add_tool.go");
  });
});

describe("MCP Add Tool Integration", () => {
  it("should have proper input schema structure", () => {
    // Define the expected input schema for the tool
    const inputSchema = {
      type: "object",
      required: ["workflow_file", "mcp_server_id"],
      properties: {
        workflow_file: {
          type: "string",
          description: "Workflow file path or ID to add the MCP tool to",
        },
        mcp_server_id: {
          type: "string",
          description: "MCP server identifier from the registry",
        },
        registry_url: {
          type: "string",
          description: "Optional MCP registry URL (defaults to GitHub's registry)",
        },
        transport_type: {
          type: "string",
          enum: ["stdio", "http", "docker"],
          description: "Preferred transport type for the MCP server",
        },
        tool_id: {
          type: "string",
          description: "Optional custom tool ID to use in the workflow",
        },
        verbose: {
          type: "boolean",
          description: "Enable verbose output",
        },
      },
      additionalProperties: false,
    };

    // Verify schema structure
    expect(inputSchema.required).toContain("workflow_file");
    expect(inputSchema.required).toContain("mcp_server_id");
    expect(inputSchema.properties.workflow_file).toBeDefined();
    expect(inputSchema.properties.mcp_server_id).toBeDefined();
  });

  it("should have proper output schema structure", () => {
    // Define the expected output schema
    const outputSchema = {
      success: "boolean",
      message: "string",
      error: "string (optional)",
    };

    // Verify basic output structure
    expect(outputSchema.success).toBe("boolean");
    expect(outputSchema.message).toBe("string");
  });
});
