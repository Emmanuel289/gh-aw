// @ts-check
import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { neutralizeWorkflowCommands, safeInfo, safeDebug, safeWarning, safeError } from "./sanitized_logging.cjs";

// Mock the global core object
global.core = {
  info: vi.fn(),
  debug: vi.fn(),
  warning: vi.fn(),
  error: vi.fn(),
};

describe("neutralizeWorkflowCommands", () => {
  it("should neutralize double colons at start of line", () => {
    const input = "::set-output name=test::value";
    const output = neutralizeWorkflowCommands(input);
    // Should replace :: at start with : (zero-width space) :
    expect(output).toBe(":\u200B:set-output name=test::value");
    expect(output).not.toBe(input);
  });

  it("should neutralize ::warning:: command at start of line", () => {
    const input = "::warning file=app.js,line=1::This is a warning";
    const output = neutralizeWorkflowCommands(input);
    expect(output).toBe(":\u200B:warning file=app.js,line=1::This is a warning");
  });

  it("should neutralize ::error:: command at start of line", () => {
    const input = "::error file=app.js,line=1::This is an error";
    const output = neutralizeWorkflowCommands(input);
    expect(output).toBe(":\u200B:error file=app.js,line=1::This is an error");
  });

  it("should neutralize ::debug:: command at start of line", () => {
    const input = "::debug::Debug message";
    const output = neutralizeWorkflowCommands(input);
    expect(output).toBe(":\u200B:debug::Debug message");
  });

  it("should neutralize ::group:: and ::endgroup:: commands at line starts", () => {
    const input = "::group::My Group\nContent\n::endgroup::";
    const output = neutralizeWorkflowCommands(input);
    expect(output).toBe(":\u200B:group::My Group\nContent\n:\u200B:endgroup::");
  });

  it("should neutralize ::add-mask:: command at start of line", () => {
    const input = "::add-mask::secret123";
    const output = neutralizeWorkflowCommands(input);
    expect(output).toBe(":\u200B:add-mask::secret123");
  });

  it("should handle text without workflow commands", () => {
    const input = "This is a normal message with no commands";
    const output = neutralizeWorkflowCommands(input);
    expect(output).toBe(input);
  });

  it("should handle text with single colons (not workflow commands)", () => {
    const input = "Time is 12:30 PM, ratio is 3:1";
    const output = neutralizeWorkflowCommands(input);
    expect(output).toBe(input);
  });

  it("should preserve :: in middle of text (IPv6, C++, etc)", () => {
    const input = "IPv6 address ::1, C++ namespace std::vector";
    const output = neutralizeWorkflowCommands(input);
    // :: in middle of text should NOT be neutralized
    expect(output).toBe(input);
  });

  it("should preserve :: after text on same line", () => {
    const input = "Time 12:30 or ratio 3::1 is fine";
    const output = neutralizeWorkflowCommands(input);
    expect(output).toBe(input);
  });

  it("should neutralize workflow command at start but preserve :: in middle", () => {
    const input = "::warning::Message about std::vector";
    const output = neutralizeWorkflowCommands(input);
    expect(output).toBe(":\u200B:warning::Message about std::vector");
  });

  it("should handle multiple workflow commands on separate lines", () => {
    const input = "::warning::First\n::error::Second\n::debug::Third";
    const output = neutralizeWorkflowCommands(input);
    expect(output).toBe(":\u200B:warning::First\n:\u200B:error::Second\n:\u200B:debug::Third");
  });

  it("should not neutralize indented :: patterns", () => {
    const input = "  ::warning::This is indented";
    const output = neutralizeWorkflowCommands(input);
    // Indented :: is not at line start, should be preserved
    expect(output).toBe(input);
  });

  it("should neutralize after newline but not in middle of line", () => {
    const input = "Some text ::not-command::\n::real-command::value";
    const output = neutralizeWorkflowCommands(input);
    expect(output).toBe("Some text ::not-command::\n:\u200B:real-command::value");
  });

  it("should handle empty string", () => {
    const input = "";
    const output = neutralizeWorkflowCommands(input);
    expect(output).toBe("");
  });

  it("should handle non-string input by converting to string", () => {
    const input = 12345;
    const output = neutralizeWorkflowCommands(input);
    expect(output).toBe("12345");
  });

  it("should handle null by converting to string", () => {
    const input = null;
    const output = neutralizeWorkflowCommands(input);
    expect(output).toBe("null");
  });

  it("should handle undefined by converting to string", () => {
    const input = undefined;
    const output = neutralizeWorkflowCommands(input);
    expect(output).toBe("undefined");
  });

  it("should preserve readability with zero-width space at line start only", () => {
    const input = "User message: ::set-output name=token::abc123";
    const output = neutralizeWorkflowCommands(input);
    // The zero-width space should be invisible but prevent command execution
    // Only the :: in middle of line is preserved
    expect(output).toBe("User message: ::set-output name=token::abc123");
  });

  it("should neutralize workflow command at actual start of string", () => {
    const input = "::set-output name=token::abc123";
    const output = neutralizeWorkflowCommands(input);
    expect(output).toBe(":\u200B:set-output name=token::abc123");
  });
});

describe("safeInfo", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("should call core.info with neutralized message at line start", () => {
    const message = "::set-output name=test::value";
    safeInfo(message);
    expect(core.info).toHaveBeenCalledWith(":\u200B:set-output name=test::value");
  });

  it("should handle normal messages without modification", () => {
    const message = "This is a normal message";
    safeInfo(message);
    expect(core.info).toHaveBeenCalledWith(message);
  });

  it("should handle messages with single colons unchanged", () => {
    const message = "Time: 12:30 PM";
    safeInfo(message);
    expect(core.info).toHaveBeenCalledWith(message);
  });

  it("should preserve :: in middle of message", () => {
    const message = "C++ std::vector is fine";
    safeInfo(message);
    expect(core.info).toHaveBeenCalledWith(message);
  });
});

describe("safeDebug", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("should call core.debug with neutralized message at line start", () => {
    const message = "::debug::User input";
    safeDebug(message);
    expect(core.debug).toHaveBeenCalledWith(":\u200B:debug::User input");
  });
});

describe("safeWarning", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("should call core.warning with neutralized message at line start", () => {
    const message = "::warning::Malicious warning";
    safeWarning(message);
    expect(core.warning).toHaveBeenCalledWith(":\u200B:warning::Malicious warning");
  });
});

describe("safeError", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("should call core.error with neutralized message at line start", () => {
    const message = "::error::Malicious error";
    safeError(message);
    expect(core.error).toHaveBeenCalledWith(":\u200B:error::Malicious error");
  });
});

describe("Integration tests", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("should prevent workflow command injection at start of message", () => {
    const userMessage = "::set-output name=hack::compromised";
    safeInfo(userMessage);
    const callArg = core.info.mock.calls[0][0];

    // Verify :: at start is neutralized
    expect(callArg).toBe(":\u200B:set-output name=hack::compromised");
  });

  it("should preserve :: in middle of text", () => {
    const userMessage = "No changes needed for std::vector";
    safeInfo(`No-op message: ${userMessage}`);
    const callArg = core.info.mock.calls[0][0];

    // :: in middle should be preserved
    expect(callArg).toBe("No-op message: No changes needed for std::vector");
  });

  it("should prevent workflow command injection in multiline text", () => {
    const title = "Bug report\n::add-mask::secret123";
    safeInfo(`Created issue: ${title}`);
    const callArg = core.info.mock.calls[0][0];

    // :: after newline should be neutralized
    expect(callArg).toBe("Created issue: Bug report\n:\u200B:add-mask::secret123");
  });

  it("should handle real-world case with command at line start", () => {
    const body = "::warning file=x.js::injected";
    safeInfo(body);
    const callArg = core.info.mock.calls[0][0];

    expect(callArg).toBe(":\u200B:warning file=x.js::injected");
  });

  it("should preserve legitimate :: usage in logged content", () => {
    const content = "IPv6 ::1 and C++::function are preserved";
    safeInfo(content);
    const callArg = core.info.mock.calls[0][0];

    expect(callArg).toBe(content);
  });
});
