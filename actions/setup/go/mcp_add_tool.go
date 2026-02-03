package main

import (
"encoding/json"
"fmt"
"os"

"github.com/github/gh-aw/pkg/cli"
)

// MCPAddInput represents the input parameters for the MCP add tool
type MCPAddInput struct {
WorkflowFile   string `json:"workflow_file"`
MCPServerID    string `json:"mcp_server_id"`
RegistryURL    string `json:"registry_url,omitempty"`
TransportType  string `json:"transport_type,omitempty"`
CustomToolID   string `json:"tool_id,omitempty"`
Verbose        bool   `json:"verbose,omitempty"`
}

// MCPAddOutput represents the output of the MCP add tool
type MCPAddOutput struct {
Success bool   `json:"success"`
Message string `json:"message"`
Error   string `json:"error,omitempty"`
}

func main() {
// Read input JSON from stdin
var input MCPAddInput
decoder := json.NewDecoder(os.Stdin)
if err := decoder.Decode(&input); err != nil {
outputError("Failed to parse input JSON", err)
os.Exit(1)
}

// Validate required parameters
if input.WorkflowFile == "" {
outputError("workflow_file is required", nil)
os.Exit(1)
}
if input.MCPServerID == "" {
outputError("mcp_server_id is required", nil)
os.Exit(1)
}

// Call the Go AddMCPTool function directly
err := cli.AddMCPTool(
input.WorkflowFile,
input.MCPServerID,
input.RegistryURL,
input.TransportType,
input.CustomToolID,
input.Verbose,
)

// Prepare output
output := MCPAddOutput{
Success: err == nil,
}

if err != nil {
output.Error = err.Error()
output.Message = fmt.Sprintf("Failed to add MCP tool '%s' to workflow '%s'", input.MCPServerID, input.WorkflowFile)
} else {
output.Message = fmt.Sprintf("Successfully added MCP tool '%s' to workflow '%s'", input.MCPServerID, input.WorkflowFile)
}

// Output result as JSON to stdout
encoder := json.NewEncoder(os.Stdout)
encoder.SetIndent("", "  ")
if err := encoder.Encode(output); err != nil {
fmt.Fprintf(os.Stderr, "Failed to encode output JSON: %v\n", err)
os.Exit(1)
}

if output.Success {
os.Exit(0)
} else {
os.Exit(1)
}
}

func outputError(message string, err error) {
output := MCPAddOutput{
Success: false,
Message: message,
}
if err != nil {
output.Error = err.Error()
}

encoder := json.NewEncoder(os.Stdout)
encoder.SetIndent("", "  ")
_ = encoder.Encode(output)
}
