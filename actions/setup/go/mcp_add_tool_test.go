//go:build !integration

package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMCPAddTool_ValidInput(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "mcp-add-test-*")
	require.NoError(t, err, "Failed to create temp directory")
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	oldWd, err := os.Getwd()
	require.NoError(t, err, "Failed to get working directory")
	defer os.Chdir(oldWd)

	err = os.Chdir(tempDir)
	require.NoError(t, err, "Failed to change to temp directory")

	// Create .github/workflows directory
	workflowsDir := filepath.Join(".github", "workflows")
	err = os.MkdirAll(workflowsDir, 0755)
	require.NoError(t, err, "Failed to create workflows directory")

	// Create a test workflow file
	workflowContent := `---
name: Test Workflow
on:
  schedule:
    - cron: "0 9 * * 1"
tools:
  github:
---

# Test Workflow

This is a test workflow.
`
	workflowPath := filepath.Join(workflowsDir, "test-workflow.md")
	err = os.WriteFile(workflowPath, []byte(workflowContent), 0644)
	require.NoError(t, err, "Failed to write test workflow")

	// Prepare input JSON
	input := MCPAddInput{
		WorkflowFile:  "test-workflow",
		MCPServerID:   "test-server",
		RegistryURL:   "https://example.com",
		TransportType: "stdio",
		CustomToolID:  "test-tool",
		Verbose:       false,
	}

	inputJSON, err := json.Marshal(input)
	require.NoError(t, err, "Failed to marshal input JSON")

	// Note: This test validates the struct definitions and JSON marshaling
	// The actual execution would require mocking the MCP registry
	assert.NotEmpty(t, inputJSON, "Input JSON should not be empty")

	var parsed MCPAddInput
	err = json.Unmarshal(inputJSON, &parsed)
	require.NoError(t, err, "Failed to unmarshal input JSON")

	assert.Equal(t, input.WorkflowFile, parsed.WorkflowFile)
	assert.Equal(t, input.MCPServerID, parsed.MCPServerID)
	assert.Equal(t, input.RegistryURL, parsed.RegistryURL)
}

func TestMCPAddTool_OutputFormat(t *testing.T) {
	// Test output struct
	output := MCPAddOutput{
		Success: true,
		Message: "Successfully added MCP tool 'test' to workflow 'test-workflow'",
	}

	outputJSON, err := json.Marshal(output)
	require.NoError(t, err, "Failed to marshal output JSON")

	var parsed MCPAddOutput
	err = json.Unmarshal(outputJSON, &parsed)
	require.NoError(t, err, "Failed to unmarshal output JSON")

	assert.True(t, parsed.Success)
	assert.Equal(t, output.Message, parsed.Message)
	assert.Empty(t, parsed.Error)
}

func TestMCPAddTool_ErrorOutput(t *testing.T) {
	// Test error output
	output := MCPAddOutput{
		Success: false,
		Message: "Failed to add MCP tool 'test' to workflow 'test-workflow'",
		Error:   "tool already exists",
	}

	outputJSON, err := json.Marshal(output)
	require.NoError(t, err, "Failed to marshal output JSON")

	var parsed MCPAddOutput
	err = json.Unmarshal(outputJSON, &parsed)
	require.NoError(t, err, "Failed to unmarshal output JSON")

	assert.False(t, parsed.Success)
	assert.Equal(t, output.Message, parsed.Message)
	assert.Equal(t, output.Error, parsed.Error)
}

func TestMCPAddTool_OutputError(t *testing.T) {
	// Test outputError function
	var buf bytes.Buffer

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	outputError("test error", nil)

	w.Close()
	os.Stdout = oldStdout

	buf.ReadFrom(r)
	outputJSON := buf.String()

	var output MCPAddOutput
	err := json.Unmarshal([]byte(outputJSON), &output)
	require.NoError(t, err, "Failed to unmarshal error output")

	assert.False(t, output.Success)
	assert.Equal(t, "test error", output.Message)
	assert.Empty(t, output.Error)
}
