---
description: Developer Instructions for GitHub Agentic Workflows
applyTo: "**/*"
---

# Developer Instructions

This document consolidates specifications and guidelines from the scratchpad/ directory into unified developer instructions for GitHub Agentic Workflows.

## Table of Contents

- [Core Architecture](#core-architecture)
- [Code Organization](#code-organization)
- [Validation Architecture](#validation-architecture)
- [Safe Outputs System](#safe-outputs-system)
- [Testing Guidelines](#testing-guidelines)
- [CLI Command Patterns](#cli-command-patterns)
- [Error Handling](#error-handling)
- [Security Best Practices](#security-best-practices)
- [Workflow Patterns](#workflow-patterns)
- [MCP Integration](#mcp-integration)
- [Go Type Patterns](#go-type-patterns)
- [Quick Reference](#quick-reference)

---

## Core Architecture

### Four-Layer Security Model

The system implements defense-in-depth through four security layers:

``````mermaid
graph TD
    A[Layer 1: Frontmatter Configuration] --> B[Layer 2: MCP Server]
    B --> C[Layer 3: Validation Guardrails]
    C --> D[Layer 4: Execution Handlers]

    A1[Workflow author declares<br/>safe-outputs configuration] -.-> A
    B1[AI agent requests operations<br/>via MCP protocol] -.-> B
    C1[Schema validation, limits,<br/>sanitization] -.-> C
    D1[GitHub Actions jobs with<br/>write permissions] -.-> D
``````

**Layer Responsibilities**:
- **Layer 1** (Frontmatter): Workflow author declares what operations are allowed and their constraints
- **Layer 2** (MCP): AI agent interface - exposes tools to agents, accepts structured JSON requests
- **Layer 3** (Guardrails): Validates all requests, enforces limits, sanitizes content
- **Layer 4** (Execution): Separate jobs with write permissions execute validated operations

**Security Principle**: AI agents operate with read-only access. Write operations execute in isolated jobs after validation.

### Compilation and Runtime Flow

``````mermaid
graph LR
    MD[Markdown Workflow<br/>.md file] --> Parser
    Parser --> Frontmatter[Frontmatter<br/>YAML metadata]
    Parser --> Body[Workflow Body<br/>Markdown content]

    Frontmatter --> Compiler
    Body --> Compiler

    Compiler --> Lock[Lock File<br/>.compiled.yaml]
    Lock --> GHA[GitHub Actions<br/>Workflow Runner]

    GHA --> Agent[Agent Job<br/>Read-only]
    GHA --> Safe[Safe Output Jobs<br/>Write permissions]
``````

**Compilation Process**:
1. Parse markdown workflow file into frontmatter and body
2. Validate configuration (frontmatter) and content (body)
3. Generate GitHub Actions YAML "lock file"
4. Lock file defines agent job (read-only) and safe output handler jobs (write)

**Runtime Process**:
1. GitHub Actions runs the compiled workflow
2. Agent job executes with MCP tools for reading and structured output
3. Safe output handler jobs execute validated write operations
4. Results propagate through job dependencies

---

## Code Organization

### File Organization Principles

#### 1. Prefer Multiple Small Files Over Large Files

**Recommended approach** (100-500 lines each):
```
create_issue.go (160 lines)
create_pull_request.go (238 lines)
create_discussion.go (118 lines)
```

**Avoid** (single large file):
```
github_operations.go (2000+ lines)  // All operations in one file
```

**Rationale**: Smaller files improve maintainability, reduce merge conflicts, enable parallel development, and simplify test organization.

#### 2. Group by Functionality, Not by Type

**Good** (feature-based):
```
create_issue.go              # Issue creation logic
create_issue_test.go         # Issue creation tests
add_comment.go               # Comment addition logic
add_comment_test.go          # Comment tests
```

**Avoid** (type-based):
```
models.go                    # All structs
logic.go                     # All business logic
tests.go                     # All tests
```

#### 3. Use Descriptive File Names

**Good**:
- `create_pull_request_reviewers_test.go` - Clear what's being tested
- `engine_error_patterns_infinite_loop_test.go` - Specific scenario
- `copilot_mcp_http_integration_test.go` - Clear scope and type

**Avoid**:
- `utils.go` - Too vague
- `helpers.go` - Too generic (unless truly shared like `engine_helpers.go`)
- `misc.go` - Indicates poor organization

### When to Create New Files

#### Decision Tree

``````mermaid
graph TD
    Start[New Feature or Code] --> Q1{Is this a new<br/>safe output type?}
    Q1 -->|Yes| A[Create create_entity.go]
    Q1 -->|No| Q2{Is this a new<br/>AI engine?}
    Q2 -->|Yes| B[Create engine_name_engine.go]
    Q2 -->|No| Q3{Is current file<br/>> 800 lines?}
    Q3 -->|Yes| C[Split by logical boundaries]
    Q3 -->|No| Q4{Is functionality<br/>independent?}
    Q4 -->|Yes| D[Create new focused file]
    Q4 -->|No| E[Add to existing file]
``````

#### Create a New File When:

1. **Implementing a new safe output type**
   - Pattern: `create_<entity>.go`
   - Example: `create_gist.go` for gist creation

2. **Adding a new engine**
   - Pattern: `<engine-name>_engine.go`
   - Example: `gemini_engine.go` for Google Gemini support

3. **Building a new domain feature**
   - Pattern: `<feature-name>.go`
   - Example: `webhooks.go` for webhook handling

4. **Current file exceeds 800 lines**
   - Extract related functionality to new file
   - Split by logical boundaries (not arbitrary line counts)

5. **Adding significant test coverage**
   - Pattern: `feature_<scenario>_test.go`
   - Example: `create_issue_assignees_test.go`

#### Extend Existing Files When:

1. Adding to existing functionality (e.g., new field to struct)
2. Fixing bugs in existing code
3. File is under 500 lines and change is related
4. Adding related helper functions to established utility files

### File Size Guidelines

- **Small files**: 50-200 lines (utilities, single-purpose functions)
- **Medium files**: 200-500 lines (most feature implementations)
- **Large files**: 500-800 lines (complex features with many aspects)
- **Very large files**: 800+ lines (core infrastructure only, consider refactoring)

**Function Count Threshold**: Consider splitting files when they exceed **50 functions**.

**Monitoring**: Run `make check-file-sizes` to identify files approaching thresholds.

### Recommended Patterns

#### 1. Create Functions Pattern

One file per GitHub entity creation operation:
- `create_issue.go` - GitHub issue creation logic
- `create_pull_request.go` - Pull request creation logic
- `create_discussion.go` - Discussion creation logic
- `create_code_scanning_alert.go` - Code scanning alert creation
- `create_agent_task.go` - Agent session creation logic

**Benefits**: Clear separation of concerns, quick location of functionality, prevents large files, facilitates parallel development.

#### 2. Engine Separation Pattern

Each AI engine has its own file with shared helpers centralized:
- `copilot_engine.go` (971 lines) - GitHub Copilot engine
- `claude_engine.go` (340 lines) - Claude engine
- `codex_engine.go` (639 lines) - Codex engine
- `custom_engine.go` (300 lines) - Custom engine support
- `agentic_engine.go` (450 lines) - Base agentic engine interface
- `engine_helpers.go` (424 lines) - Shared engine utilities

**Benefits**: Engine-specific logic is isolated, shared code is centralized, allows adding engines without affecting others, clear boundaries reduce conflicts.

#### 3. Test Organization Pattern

Tests live alongside implementation files with descriptive names:
- Feature tests: `feature.go` + `feature_test.go`
- Integration tests: `feature_integration_test.go`
- Security tests: `feature_security_regression_test.go`

**Benefits**: Tests co-located with implementation, clear test purpose from filename, supports coverage requirements.

---

## Validation Architecture

### Organization

Validation is organized into centralized and domain-specific layers:

``````mermaid
graph TD
    A[Validation Request] --> B{Validation Type}
    B -->|Cross-cutting| C[validation.go]
    B -->|Security| D[strict_mode_validation.go]
    B -->|Python packages| E[pip.go]
    B -->|NPM packages| F[npm.go]
    B -->|Docker images| G[docker.go]
    B -->|Expressions| H[expression_safety.go]
    B -->|Engines| I[engine.go]
    B -->|MCP config| J[mcp-config.go]
    B -->|JS bundling| K[bundler_*_validation.go]
``````

### Centralized Validation (validation.go)

**Purpose**: General-purpose validation applying across the entire workflow system

**Key Functions**:
- `validateExpressionSizes()` - Ensures GitHub Actions expression size limits
- `validateContainerImages()` - Verifies Docker images exist and are accessible
- `validateRuntimePackages()` - Validates runtime package dependencies
- `validateGitHubActionsSchema()` - Validates against GitHub Actions YAML schema
- `validateNoDuplicateCacheIDs()` - Ensures unique cache identifiers
- `validateSecretReferences()` - Validates secret reference syntax
- `validateRepositoryFeatures()` - Checks repository capabilities (issues, discussions)

**When to add here**:
- Cross-cutting concerns spanning multiple domains
- Core workflow integrity checks
- GitHub Actions compatibility validation
- General schema and configuration validation

### Domain-Specific Validation Files

#### Strict Mode Validation (strict_mode_validation.go)

**Purpose**: Enforces security and safety constraints in strict mode

**Functions**:
- `validateStrictMode()` - Main strict mode orchestrator
- `validateStrictPermissions()` - Refuses write permissions
- `validateStrictNetwork()` - Requires explicit network configuration
- `validateStrictMCPNetwork()` - Requires network config on custom MCP servers
- `validateStrictBashTools()` - Refuses bash wildcard tools

**Pattern**: Progressive validation (multiple checks in sequence)

#### External Resource Validation

**Python Packages** (pip.go):
- `validatePipPackages()` - Validates pip packages exist on PyPI
- `validateUvPackages()` - Validates uv packages
- **Pattern**: External registry validation with warnings (not errors)

**NPM Packages** (npm.go):
- `validateNpxPackages()` - Validates npm packages for npx
- **Pattern**: External registry validation with error reporting

**Docker Images** (docker.go):
- `validateDockerImage()` - Validates Docker image exists and is pullable
- **Pattern**: External resource validation with caching

#### JavaScript Bundle Validation

**Safety** (bundler_safety_validation.go):
- `validateNoLocalRequires()` - Ensures local require() statements are bundled
- `validateNoModuleReferences()` - Ensures no module.exports remain in GitHub Script mode
- `ValidateEmbeddedResourceRequires()` - Validates embedded dependencies exist

**Script Content** (bundler_script_validation.go):
- `validateNoExecSync()` - GitHub Script mode should use async exec
- `validateNoGitHubScriptGlobals()` - Node.js scripts shouldn't use GitHub Actions globals
- **Enforcement**: Registration-time validation with panics on violation

**Runtime Mode** (bundler_runtime_validation.go):
- `validateNoRuntimeMixing()` - Prevents mixing nodejs-only with github-script
- `detectRuntimeMode()` - Detects intended runtime mode

### Validation Patterns

#### 1. Allowlist Validation

Used for security-sensitive validation (expression_safety.go):

```go
func validateExpressionSafety(content string) error {
    matches := expressionRegex.FindAllStringSubmatch(content, -1)
    var unauthorized []string

    for _, match := range matches {
        expression := strings.TrimSpace(match[1])
        if !isAllowed(expression) {
            unauthorized = append(unauthorized, expression)
        }
    }

    if len(unauthorized) > 0 {
        return fmt.Errorf("unauthorized expressions: %v", unauthorized)
    }
    return nil
}
```

**When to use**: Security-sensitive validation, limited set of valid options, preventing injection attacks.

#### 2. External Resource Validation

Used for verifying external dependencies exist before runtime:

```go
func validateDockerImage(image string, verbose bool) error {
    cmd := exec.Command("docker", "inspect", image)
    _, err := cmd.CombinedOutput()

    if err != nil {
        pullCmd := exec.Command("docker", "pull", image)
        if pullErr := pullCmd.Run(); pullErr != nil {
            return fmt.Errorf("docker image not found: %s", image)
        }
    }
    return nil
}
```

**When to use**: Validating external dependencies, package registry checks, container image availability.

#### 3. Progressive Validation

Used for multiple related validation steps (strict_mode_validation.go):

```go
func (c *Compiler) validateStrictMode(frontmatter map[string]any, networkPermissions *NetworkPermissions) error {
    if !c.strictMode {
        return nil
    }

    // 1. Refuse write permissions
    if err := c.validateStrictPermissions(frontmatter); err != nil {
        return err
    }

    // 2. Require network configuration
    if err := c.validateStrictNetwork(networkPermissions); err != nil {
        return err
    }

    // 3. Validate MCP network
    if err := c.validateStrictMCPNetwork(frontmatter); err != nil {
        return err
    }

    return nil
}
```

**When to use**: Multiple related validation steps, security policy enforcement, layered validation requirements.

#### 4. Warning vs Error Validation

Used to distinguish between hard failures and soft warnings:

```go
func (c *Compiler) validatePythonPackagesWithPip(packages []string) {
    for _, pkg := range packages {
        cmd := exec.Command("pip", "index", "versions", pkg)
        _, err := cmd.CombinedOutput()

        if err != nil {
            // Warning: Don't fail compilation
            fmt.Fprintln(os.Stderr, console.FormatWarningMessage(
                fmt.Sprintf("pip package '%s' validation failed - skipping", pkg)))
        } else {
            if c.verbose {
                fmt.Fprintln(os.Stderr, console.FormatInfoMessage(
                    fmt.Sprintf("✓ pip package validated: %s", pkg)))
            }
        }
    }
}
```

**When to use**: Optional dependency validation, best-effort external checks, non-critical validations.

### Decision Tree: Where to Add New Validation

```
New Validation Requirement
  │
  ├─ Is it about security or strict mode?
  │  └─ YES → strict_mode_validation.go
  │
  ├─ Does it only apply to one specific domain?
  │  ├─ YES → Is there a domain-specific file?
  │  │  ├─ YES → Add to domain file
  │  │  └─ NO → Create new domain file
  │  └─ NO → Continue
  │
  ├─ Is it a cross-cutting concern?
  │  └─ YES → validation.go
  │
  ├─ Does it validate external resources?
  │  └─ YES → Domain-specific file (pip.go, npm.go, docker.go)
  │
  └─ DEFAULT → validation.go
```

---

## Safe Outputs System

### Overview

The Safe Outputs System enables AI agents to request write operations to GitHub resources without possessing write permissions. It implements a security-oriented architecture through four layers.

### Architecture Diagram

``````mermaid
graph LR
    FM[Frontmatter<br/>Configuration] --> MCP[MCP Server<br/>Tool Interface]
    MCP --> Guard[Validation<br/>Guardrails]
    Guard --> Exec[Execution<br/>Handlers]

    FM -.->|Declares allowed ops| MCP
    MCP -.->|Structured JSON| Guard
    Guard -.->|Validated requests| Exec
    Exec -.->|GitHub API calls| GH[GitHub]
``````

### Configuration Schema

Workflow authors declare safe-outputs in frontmatter:

```yaml
safe-outputs:
  # Builtin system tools
  missing-tool: {}
  missing-data: {}
  noop: {}

  # GitHub operations
  create-issue:
    max: 5
    title-prefix: "[bot] "
    labels:
      - automation

  add-comment:
    max: 10

  create-pull-request:
    max: 1
    title-prefix: "[docs] "
    labels:
      - documentation
      - automation
```

**Configuration Controls**:
- `max`: Maximum number of operations allowed (per type)
- `title-prefix`: Automatic prefix for titles
- `labels`: Automatic labels applied
- `target`: Where operation can run (`triggering`, `*`, or repository name)
- Per-operation specific fields (see individual operation specs)

### Data Flow

``````mermaid
sequenceDiagram
    participant Agent
    participant MCP
    participant Guardrails
    participant Handler
    participant GitHub

    Agent->>MCP: call_tool("create-issue", {...})
    MCP->>MCP: Collect structured output (NDJSON)
    MCP-->>Agent: Success response

    Note over Agent,MCP: Agent job completes (read-only)

    Handler->>Handler: Parse NDJSON output
    Handler->>Guardrails: Validate request
    Guardrails->>Guardrails: Check schema, max count, sanitize
    Guardrails-->>Handler: Validation result

    alt Validation Passed
        Handler->>GitHub: Create issue via API
        GitHub-->>Handler: Issue created
        Handler->>Handler: Record operation
    else Validation Failed
        Handler->>Handler: Log error, skip operation
    end
``````

### Builtin System Tools

Three builtin tools are always available:

1. **missing-tool**: Report unavailable tools
   - Use when required tool is missing
   - Provides alternatives or manual steps

2. **missing-data**: Report missing information
   - Use when data needed to complete task is unavailable
   - Explains what data is needed and why

3. **noop**: No-operation status logging
   - Use when analysis is complete but no actions needed
   - Ensures workflow produces visible output

### Security Controls

**Layer 1 (Frontmatter)**:
- Workflow author explicitly declares allowed operations
- Sets operation limits (max counts)
- Configures title prefixes, labels, auto-expiration

**Layer 2 (MCP)**:
- Tools registered per configuration
- No direct GitHub API access for agent
- Structured JSON output only (prevents injection)

**Layer 3 (Guardrails)**:
- Schema validation against operation type
- Max count enforcement
- Label sanitization (character restrictions, length limits)
- Content sanitization (HTML escaping, markdown limits)
- Target validation (repo restrictions)

**Layer 4 (Execution)**:
- Separate jobs with minimal write permissions
- Each operation type gets only required permissions
- Operations logged and traceable
- Error handling with fallback strategies

---

## Testing Guidelines

### Test Organization

Tests are co-located with implementation:
- Unit tests: `feature.go` + `feature_test.go`
- Integration tests: `feature_integration_test.go` (marked with `//go:build integration`)
- Security tests: `feature_security_regression_test.go`
- Fuzz tests: `feature_fuzz_test.go`

### Assert vs Require

Use **testify** assertions appropriately:

**`require.*`** - For critical setup steps (stops test immediately on failure):
- Creating test files
- Parsing input
- Setting up test data

**`assert.*`** - For actual test validations (allows test to continue):
- Verifying behavior
- Checking output values
- Testing multiple conditions

**Example**:

```go
func TestWorkflowCompilation(t *testing.T) {
    compiler := NewCompilerWithVersion("1.0.0")

    // Setup - use require
    tmpDir := t.TempDir()
    testFile := filepath.Join(tmpDir, "test.md")
    err := os.WriteFile(testFile, []byte(markdown), 0644)
    require.NoError(t, err, "Failed to write test file")

    // Parse - use require (critical for test to continue)
    workflowData, err := compiler.ParseWorkflowFile(testFile)
    require.NoError(t, err, "Failed to parse workflow")
    require.NotNil(t, workflowData, "Workflow data should not be nil")

    // Verify behavior - use assert (actual validations)
    assert.Equal(t, "expected-value", workflowData.Field)
    assert.Contains(t, workflowData.Features, "feature-name")
    assert.NoError(t, workflowData.Validate())
}
```

### Table-Driven Tests

Use table-driven tests with `t.Run()` for multiple scenarios:

```go
func TestValidation(t *testing.T) {
    tests := []struct {
        name        string
        input       string
        expectError bool
        errorMsg    string
    }{
        {
            name:        "valid input",
            input:       "valid-data",
            expectError: false,
        },
        {
            name:        "empty input",
            input:       "",
            expectError: true,
            errorMsg:    "input cannot be empty",
        },
        {
            name:        "invalid format",
            input:       "invalid@format",
            expectError: true,
            errorMsg:    "invalid format",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := Validate(tt.input)

            if tt.expectError {
                assert.Error(t, err, "Expected error for %s", tt.input)
                if tt.errorMsg != "" {
                    assert.Contains(t, err.Error(), tt.errorMsg)
                }
            } else {
                assert.NoError(t, err, "Unexpected error for %s", tt.input)
            }
        })
    }
}
```

**Key principles**:
- Use descriptive test case names
- Structure: Define test cases → Loop with `t.Run()` → Test logic
- Each sub-test runs independently

### Why No Mocks or Test Suites?

This project intentionally avoids mocking frameworks and test suites:

**No mocks because**:
- Tests use real component interactions
- Tests verify actual behavior, not mock behavior
- No mock setup/teardown boilerplate
- Tests catch real integration issues

**No test suites (testify/suite) because**:
- Standard Go tests run in parallel by default
- No suite lifecycle methods to understand
- Setup is visible in each test
- Compatible with standard `go test` tooling

### Running Tests

```bash
# Fast unit tests (recommended during development)
make test-unit       # ~25s - Unit tests only

# Full test suite
make test            # ~30s - All tests including integration

# Specific tests
go test -v ./pkg/workflow/...                    # Test specific package
go test -run TestSafeOutputs ./pkg/workflow/...  # Run specific test

# Security regression tests
make test-security

# With coverage
make test-coverage

# Benchmarks
make bench

# Fuzz testing
make fuzz

# Linting (includes test quality checks)
make lint            # Runs golangci-lint with testifylint rules

# Complete validation (before committing)
make agent-finish    # Runs build, test, recompile, fmt, lint
```

**Note**: The project uses testifylint (via golangci-lint) to enforce consistent test assertion usage.

---

## CLI Command Patterns

### Standard Command Structure

Every command follows this structure:

```go
package cli

import (
    "fmt"
    "os"

    "github.com/github/gh-aw/pkg/console"
    "github.com/github/gh-aw/pkg/logger"
    "github.com/spf13/cobra"
)

// Logger instance with namespace following cli:command_name pattern
var commandLog = logger.New("cli:command_name")

// NewCommandNameCommand creates the command-name command
func NewCommandNameCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "command-name [args]",
        Short: "Brief one-line description under 80 chars",
        Long:  `Detailed description with examples...`,
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            // Parse flags
            flagValue, _ := cmd.Flags().GetString("flag-name")
            verbose, _ := cmd.Flags().GetBool("verbose")

            // Call main function
            return RunCommandName(args[0], flagValue, verbose)
        },
    }

    // Add flags
    cmd.Flags().StringP("flag-name", "f", "default", "Flag description")
    addVerboseFlag(cmd)

    // Register completions
    RegisterDirFlagCompletion(cmd, "output")

    return cmd
}

// RunCommandName executes the command logic
func RunCommandName(arg string, flagValue string, verbose bool) error {
    commandLog.Printf("Starting command: arg=%s, flagValue=%s", arg, flagValue)

    // Validate inputs early
    if err := validateInputs(arg, flagValue); err != nil {
        return err
    }

    // Execute command logic
    result, err := executeCommand(arg, flagValue)
    if err != nil {
        return fmt.Errorf("failed to execute command: %w", err)
    }

    // Output results
    fmt.Fprintln(os.Stderr, console.FormatSuccessMessage(result))

    commandLog.Print("Command completed successfully")
    return nil
}
```

### Key Structure Elements

1. **Logger Namespace**: Always `cli:command_name`
2. **Public API**: Two exported functions:
   - `NewXCommand() *cobra.Command` - Creates the command
   - `RunX(...) error` - Executes command logic (testable)
3. **Internal Functions**: Private helper functions for implementation
4. **Flag Parsing**: Parse flags in `RunE` function
5. **Early Validation**: Validate all inputs before processing
6. **Error Wrapping**: Use `fmt.Errorf` with `%w` for context

### File Organization

#### Single File Pattern (< 500 lines)
- `command_name.go` - All command logic
- `command_name_test.go` - All tests

#### Multi-File Pattern (> 500 lines)
- `command_name_command.go` - Command definition
- `command_name_config.go` - Configuration types
- `command_name_helpers.go` - Utility functions
- `command_name_validation.go` - Validation logic
- `command_name_orchestrator.go` - Main orchestration

### Naming Conventions

**Logger Namespaces**: Format is `cli:command_name`

```go
// ✅ CORRECT
var auditLog = logger.New("cli:audit")
var compileLog = logger.New("cli:compile_command")

// ❌ INCORRECT
var log = logger.New("audit")  // Missing cli: prefix
```

**Public Functions**:
- Command creator: `NewXCommand()`
- Command runner: `RunX(...)`

**Configuration Structs**: End with `Config`

```go
type CompileConfig struct {
    WorkflowFile string
    OutputDir    string
    Verbose      bool
}
```

---

## Error Handling

### Error Patterns

#### 1. Error Wrapping

Always wrap errors with context using `fmt.Errorf` with `%w`:

```go
func ProcessWorkflow(file string) error {
    data, err := os.ReadFile(file)
    if err != nil {
        return fmt.Errorf("failed to read workflow file %s: %w", file, err)
    }

    workflow, err := ParseWorkflow(data)
    if err != nil {
        return fmt.Errorf("failed to parse workflow: %w", err)
    }

    return nil
}
```

**Benefits**: Preserves error chain for `errors.Is()` and `errors.As()`, provides context for debugging.

#### 2. Error Collection

Collect multiple errors before returning:

```go
func ValidateWorkflow(wf *Workflow) error {
    var errs []string

    if wf.Name == "" {
        errs = append(errs, "name is required")
    }

    if wf.Trigger == "" {
        errs = append(errs, "trigger is required")
    }

    if len(errs) > 0 {
        return fmt.Errorf("validation failed:\n  - %s", strings.Join(errs, "\n  - "))
    }

    return nil
}
```

**Benefits**: User sees all validation errors at once, not just the first.

#### 3. Console Output for Errors

Use console formatting for user-facing errors:

```go
import "github.com/github/gh-aw/pkg/console"

// Success
fmt.Fprintln(os.Stderr, console.FormatSuccessMessage("✓ Compilation successful"))

// Warning
fmt.Fprintln(os.Stderr, console.FormatWarningMessage("Package validation skipped"))

// Error
fmt.Fprintln(os.Stderr, console.FormatErrorMessage(err.Error()))

// Info (verbose mode)
if verbose {
    fmt.Fprintln(os.Stderr, console.FormatInfoMessage("Processing workflow..."))
}
```

#### 4. Logging for Debugging

Use logger package for debug output:

```go
var myLog = logger.New("pkg:feature")

func ProcessData() error {
    myLog.Print("Starting data processing")
    myLog.Printf("Processing %d items", count)

    // ...

    myLog.Print("Processing complete")
    return nil
}
```

Enable with: `DEBUG=pkg:feature gh aw <command>`

---

## Security Best Practices

### Template Injection Prevention

**Never** use user-provided content in GitHub Actions expressions without proper escaping:

```yaml
# ❌ DANGEROUS - User input directly in expression
run: echo "${{ github.event.issue.title }}"

# ✅ SAFE - User input via environment variable
env:
  ISSUE_TITLE: ${{ github.event.issue.title }}
run: echo "$ISSUE_TITLE"
```

**Rationale**: GitHub Actions expressions are evaluated before execution. Malicious input like `title": "x", "run": "malicious code"` could inject commands.

### Expression Allowlist

Only allow specific GitHub Actions expressions in workflow content. Maintain an allowlist of safe expressions:

**Allowed**:
- `github.event.issue.number`
- `github.event.pull_request.number`
- `github.event.discussion.number`
- `github.repository`
- `github.run_id`
- `github.run_attempt`

**Blocked**: Any expression not in allowlist, especially:
- `secrets.*` (secret access)
- `steps.*` (step outputs)
- Arbitrary code execution expressions

### String Sanitization

Sanitize user-provided content before using in workflows:

```go
func sanitizeLabel(label string) (string, error) {
    // Remove non-alphanumeric, hyphen, underscore
    sanitized := regexp.MustCompile(`[^a-zA-Z0-9\-_]`).ReplaceAllString(label, "")

    // Enforce length limits
    if len(sanitized) > 50 {
        return "", fmt.Errorf("label too long (max 50 chars)")
    }

    if sanitized == "" {
        return "", fmt.Errorf("label cannot be empty after sanitization")
    }

    return sanitized, nil
}
```

### Strict Mode

Enable strict mode for security-sensitive workflows:

```yaml
---
strict: true
---
```

**Strict mode enforces**:
- No write permissions allowed
- Explicit network configuration required
- No bash wildcard tools
- MCP servers must declare network access

---

## Workflow Patterns

### Refactoring for Size Reduction

When workflows exceed size limits, use these strategies:

``````mermaid
graph TD
    Large[Large Workflow<br/>> 1MB] --> Strategy{Reduction Strategy}

    Strategy -->|1| Extract[Extract reusable<br/>content to files]
    Strategy -->|2| Inline[Use runtime-import<br/>for inlining]
    Strategy -->|3| Split[Split into multiple<br/>workflows]

    Extract --> Result[Smaller workflow,<br/>reusable components]
    Inline --> Result
    Split --> Result
``````

#### 1. Extract to Files

Move reusable content to separate files:

**Before**:
```markdown
---
title: Large Workflow
---

# Prompt

Long prompt content...

## Examples

Many examples...
```

**After**:
```markdown
---
title: Refactored Workflow
---

# Prompt

<<< ./prompts/main-prompt.md

## Examples

<<< ./examples/example-set.md
```

#### 2. Runtime Import (File Inlining)

Use `runtime-import` for content inlined at runtime:

```yaml
steps:
  - runtime-import: ./scripts/setup.js
```

**Processing Flow**:

``````mermaid
sequenceDiagram
    participant Compiler
    participant Runtime
    participant Script

    Compiler->>Compiler: Parse runtime-import directive
    Compiler->>Compiler: Generate placeholder in YAML
    Compiler->>Runtime: Pass file path metadata

    Runtime->>Script: Read file at runtime
    Runtime->>Runtime: Inline content into step
    Runtime->>Runtime: Execute inlined script
``````

**Benefits**: Reduces compiled YAML size, enables dynamic content, allows code reuse.

#### 3. Split into Multiple Workflows

For very large workflows, split into specialized workflows:

```
workflows/
├── main.md              # Orchestrator workflow
├── validation.md        # Validation workflow
├── processing.md        # Processing workflow
└── reporting.md         # Reporting workflow
```

**Orchestration**:
```yaml
on:
  workflow_dispatch:

jobs:
  trigger-validation:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/github-script@v7
        with:
          script: |
            await github.rest.actions.createWorkflowDispatch({
              owner: context.repo.owner,
              repo: context.repo.repo,
              workflow_id: 'validation.yml',
              ref: 'main'
            })
```

---

## MCP Integration

### MCP Access Control

Three-layer access control architecture:

``````mermaid
graph TD
    A[Layer 1:<br/>Frontmatter Allowlist] --> B[Layer 2:<br/>Tool Filtering]
    B --> C[Layer 3:<br/>Operation Validation]

    A1[Workflow author declares<br/>allowed MCP servers] -.-> A
    B1[Only allowed tools<br/>registered with agent] -.-> B
    C1[Validate each operation<br/>against configuration] -.-> C
``````

**Layer 1 (Frontmatter Allowlist)**:
- Workflow author explicitly lists allowed MCP servers
- Declares server configuration (args, env)
- Sets network access permissions

**Layer 2 (Tool Filtering)**:
- Only tools from allowed servers are registered
- Tools inherit server's network permissions
- Agent cannot access tools from non-allowed servers

**Layer 3 (Operation Validation)**:
- Each tool call validated against configuration
- Enforce operation-specific limits (max count)
- Validate request parameters against schema

### MCP Server Configuration

```yaml
---
mcp:
  servers:
    github:
      command: "npx"
      args:
        - "-y"
        - "@modelcontextprotocol/server-github"
      network:
        allowed: true
        domains:
          - "api.github.com"

    custom-server:
      command: "python"
      args:
        - "-m"
        - "custom_server"
      network:
        allowed: false
---
```

**Configuration Fields**:
- `command`: Executable to run
- `args`: Command arguments
- `env`: Environment variables (optional)
- `network.allowed`: Whether server can access network
- `network.domains`: Allowed domains (when network.allowed: true)

### Safe Outputs via MCP

MCP servers can register safe-output tools:

```typescript
server.setRequestHandler(ListToolsRequestSchema, async () => ({
  tools: [
    {
      name: "create-issue",
      description: "Create a GitHub issue",
      inputSchema: {
        type: "object",
        properties: {
          title: { type: "string" },
          body: { type: "string" },
          labels: { type: "array", items: { type: "string" } }
        },
        required: ["title", "body"]
      }
    }
  ]
}));
```

**Tool Call Flow**:
1. Agent calls MCP tool with structured parameters
2. MCP server validates request against inputSchema
3. Server writes NDJSON line to stdout
4. Workflow parser collects NDJSON output
5. Safe output handler validates and executes operation

---

## Go Type Patterns

### Type Safety Patterns

#### 1. String-Based Enums

Use typed strings for enum-like values:

```go
// EngineID represents an AI engine identifier
type EngineID string

const (
    EngineCopilot EngineID = "copilot"
    EngineClaude  EngineID = "claude"
    EngineCodex   EngineID = "codex"
    EngineCustom  EngineID = "custom"
)

// Valid returns whether the engine ID is valid
func (e EngineID) Valid() bool {
    switch e {
    case EngineCopilot, EngineClaude, EngineCodex, EngineCustom:
        return true
    default:
        return false
    }
}
```

**Benefits**: Type safety, autocomplete support, validation methods.

#### 2. Configuration Structs

Use structs for configuration with YAML tags:

```go
type WorkflowConfig struct {
    Title       string            `yaml:"title"`
    Description string            `yaml:"description"`
    Engine      EngineID          `yaml:"engine"`
    Permissions map[string]string `yaml:"permissions"`
    SafeOutputs *SafeOutputsConfig `yaml:"safe-outputs,omitempty"`
}

type SafeOutputsConfig struct {
    MaxIssues       int      `yaml:"max-issues"`
    IssueTitlePrefix string   `yaml:"issue-title-prefix"`
    IssueLabels     []string `yaml:"issue-labels"`
}
```

**Benefits**: Strong typing, validation, YAML marshaling/unmarshaling.

#### 3. Validation Methods

Add validation methods to config structs:

```go
func (c *WorkflowConfig) Validate() error {
    if c.Title == "" {
        return fmt.Errorf("title is required")
    }

    if !c.Engine.Valid() {
        return fmt.Errorf("invalid engine: %s", c.Engine)
    }

    if c.SafeOutputs != nil {
        if err := c.SafeOutputs.Validate(); err != nil {
            return fmt.Errorf("safe-outputs validation failed: %w", err)
        }
    }

    return nil
}
```

**Benefits**: Centralized validation, composable validation chains.

#### 4. Builder Pattern for Complex Types

Use builder pattern for complex object construction:

```go
type CompilerBuilder struct {
    version    string
    verbose    bool
    strictMode bool
    outputDir  string
}

func NewCompilerBuilder() *CompilerBuilder {
    return &CompilerBuilder{}
}

func (b *CompilerBuilder) WithVersion(version string) *CompilerBuilder {
    b.version = version
    return b
}

func (b *CompilerBuilder) WithVerbose(verbose bool) *CompilerBuilder {
    b.verbose = verbose
    return b
}

func (b *CompilerBuilder) WithStrictMode(strict bool) *CompilerBuilder {
    b.strictMode = strict
    return b
}

func (b *CompilerBuilder) Build() *Compiler {
    return &Compiler{
        version:    b.version,
        verbose:    b.verbose,
        strictMode: b.strictMode,
        outputDir:  b.outputDir,
    }
}
```

**Usage**:

```go
compiler := NewCompilerBuilder().
    WithVersion("1.0.0").
    WithVerbose(true).
    WithStrictMode(true).
    Build()
```

**Benefits**: Fluent API, optional parameters, immutability.

---

## Quick Reference

### Common Commands

```bash
# Build
make build

# Test
make test-unit       # Fast unit tests (~25s)
make test            # Full test suite (~30s)

# Validation
make lint            # Linting
make fmt             # Format code
make agent-finish    # Complete validation

# Compilation
gh aw compile workflow.md
gh aw compile workflow.md --verbose
gh aw compile workflow.md --strict

# Audit
gh aw audit <run-id>
gh aw audit <run-id> --verbose
```

### Key File Locations

- Compiler: `pkg/workflow/compiler.go`
- Validation: `pkg/workflow/validation.go`, `pkg/workflow/*_validation.go`
- Safe Outputs: `pkg/workflow/safe_outputs.go`, `pkg/workflow/create_*.go`
- CLI Commands: `pkg/cli/*_command.go`
- Tests: `*_test.go` (co-located with implementation)

### Logger Namespaces

- CLI commands: `cli:command_name`
- Workflow package: `workflow:feature`
- Parser: `parser:feature`

### Decision Trees

**Create New File?**
1. Is it a new safe output type? → `create_<entity>.go`
2. Is it a new AI engine? → `<engine>_engine.go`
3. Is current file > 800 lines? → Split by logical boundaries
4. Is functionality independent? → Create new file
5. Otherwise → Add to existing file

**Add Validation?**
1. Is it security/strict mode? → `strict_mode_validation.go`
2. Is it domain-specific? → Domain-specific file or create new
3. Is it cross-cutting? → `validation.go`
4. Is it external resources? → `pip.go`, `npm.go`, `docker.go`
5. Otherwise → `validation.go`

---

## Contributing

When contributing to this codebase:

1. **Follow established patterns**: Use existing code as examples
2. **Write tests**: Add tests for all new functionality
3. **Document code**: Add comments explaining complex logic
4. **Use type safety**: Leverage Go's type system
5. **Validate early**: Check inputs before processing
6. **Handle errors**: Wrap errors with context
7. **Run validation**: Use `make agent-finish` before committing

For questions or clarifications, refer to the scratchpad/ directory for detailed specifications on specific topics.

---

**Last Updated**: 2026-02-11
**Consolidated From**: 53 specification files in scratchpad/ directory
**Previous Version**: 2026-02-10 consolidation (1257 lines)
**Current Version**: Enhanced with additional Mermaid diagrams and improved technical tone
