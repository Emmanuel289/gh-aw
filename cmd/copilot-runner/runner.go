package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	copilot "github.com/github/copilot-sdk/go"
)

// Runner executes a Copilot session using the SDK.
type Runner struct {
	config  *Config
	metrics *Metrics
}

// NewRunner creates a new Runner with the given configuration.
func NewRunner(config *Config) *Runner {
	return &Runner{
		config:  config,
		metrics: NewMetrics(),
	}
}

// Run executes the Copilot session:
// 1. Creates a Copilot SDK client
// 2. Starts the client (spawns CLI server process)
// 3. Creates a session with the configured tools and model
// 4. Registers event handlers for metrics collection
// 5. Sends the prompt and waits for completion
// 6. Writes metrics and returns
func (r *Runner) Run(ctx context.Context) error {
	// Read the prompt from file
	promptData, err := os.ReadFile(r.config.PromptFile)
	if err != nil {
		return fmt.Errorf("failed to read prompt file %s: %w", r.config.PromptFile, err)
	}
	prompt := string(promptData)

	if prompt == "" {
		return fmt.Errorf("prompt file %s is empty", r.config.PromptFile)
	}

	fmt.Fprintf(os.Stderr, "Starting Copilot SDK runner...\n")
	fmt.Fprintf(os.Stderr, "CLI path: %s\n", r.config.CLIPath)
	fmt.Fprintf(os.Stderr, "Model: %s\n", r.config.Model)
	fmt.Fprintf(os.Stderr, "Prompt length: %d chars\n", len(prompt))

	// Build environment for the CLI process
	env := os.Environ()
	for key, value := range r.config.Env {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}

	// Create the SDK client
	clientOpts := &copilot.ClientOptions{
		CLIPath:  r.config.CLIPath,
		LogLevel: r.config.LogLevel,
		Env:      env,
	}

	// Set working directory if specified
	if r.config.WorkingDirectory != "" {
		clientOpts.Cwd = r.config.WorkingDirectory
	}

	// Set GitHub token for authentication
	if r.config.GithubToken != "" {
		clientOpts.GithubToken = r.config.GithubToken
	}

	client := copilot.NewClient(clientOpts)

	// Start the client (spawns CLI server)
	fmt.Fprintf(os.Stderr, "Starting Copilot CLI server...\n")
	if err := client.Start(ctx); err != nil {
		return fmt.Errorf("failed to start Copilot CLI server: %w", err)
	}
	defer func() {
		fmt.Fprintf(os.Stderr, "Stopping Copilot CLI server...\n")
		if err := client.Stop(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: error stopping client: %v\n", err)
		}
	}()

	// Build session configuration
	sessionConfig := &copilot.SessionConfig{
		Streaming: r.config.Streaming,
	}

	if r.config.Model != "" {
		sessionConfig.Model = r.config.Model
	}

	if r.config.WorkingDirectory != "" {
		sessionConfig.WorkingDirectory = r.config.WorkingDirectory
	}

	// Set available/excluded tools
	if len(r.config.AvailableTools) > 0 {
		sessionConfig.AvailableTools = r.config.AvailableTools
	}
	if len(r.config.ExcludedTools) > 0 {
		sessionConfig.ExcludedTools = r.config.ExcludedTools
	}

	// Load MCP server configuration if specified
	if r.config.MCPConfigPath != "" {
		mcpServers, err := loadMCPServers(r.config.MCPConfigPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to load MCP config: %v\n", err)
		} else if mcpServers != nil {
			sessionConfig.MCPServers = mcpServers
		}
	}

	// Create the session
	fmt.Fprintf(os.Stderr, "Creating session...\n")
	session, err := client.CreateSession(ctx, sessionConfig)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Session created: %s\n", session.SessionID)

	// Register event handler for metrics collection
	unsubscribe := session.On(r.metrics.HandleEvent)
	defer unsubscribe()

	// Register event handler for streaming output
	unsubscribeStreaming := session.On(func(event copilot.SessionEvent) {
		switch event.Type {
		case copilot.AssistantMessage:
			if event.Data.Content != nil {
				fmt.Print(*event.Data.Content)
			}
		case copilot.AssistantMessageDelta:
			if event.Data.DeltaContent != nil {
				fmt.Print(*event.Data.DeltaContent)
			}
		case copilot.ToolExecutionStart:
			if event.Data.ToolName != nil {
				fmt.Fprintf(os.Stderr, "Executing tool: %s\n", *event.Data.ToolName)
			}
		case copilot.ToolExecutionComplete:
			if event.Data.ToolName != nil {
				fmt.Fprintf(os.Stderr, "Tool complete: %s\n", *event.Data.ToolName)
			}
		case copilot.SessionError:
			if event.Data.Message != nil {
				fmt.Fprintf(os.Stderr, "Session error: %s\n", *event.Data.Message)
			}
		}
	})
	defer unsubscribeStreaming()

	// Determine timeout
	timeout := 30 * time.Minute // Default timeout
	if r.config.Timeout > 0 {
		timeout = time.Duration(r.config.Timeout) * time.Second
	}

	// Create a context with timeout
	sendCtx, sendCancel := context.WithTimeout(ctx, timeout)
	defer sendCancel()

	// Send the prompt and wait for completion
	fmt.Fprintf(os.Stderr, "Sending prompt...\n")
	response, err := session.SendAndWait(sendCtx, copilot.MessageOptions{
		Prompt: prompt,
	})
	if err != nil {
		return fmt.Errorf("failed to send prompt: %w", err)
	}

	// Print final response if we got one and it wasn't already streamed
	if response != nil && !r.config.Streaming && response.Data.Content != nil {
		fmt.Println(*response.Data.Content)
	}

	// Finalize metrics
	r.metrics.Finalize()
	r.metrics.PrintSummary()

	// Write metrics file if configured
	if r.config.MetricsFile != "" {
		if err := r.metrics.WriteToFile(r.config.MetricsFile); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to write metrics: %v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "Metrics written to %s\n", r.config.MetricsFile)
		}
	}

	fmt.Fprintf(os.Stderr, "Copilot SDK runner completed successfully\n")
	return nil
}

// loadMCPServers loads MCP server configuration from a JSON file.
// The format matches the Copilot CLI's mcp-config.json structure.
func loadMCPServers(path string) (map[string]copilot.MCPServerConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read MCP config: %w", err)
	}

	// Expand environment variables in the config
	expanded := os.ExpandEnv(string(data))

	// The MCP config has a top-level "mcpServers" key
	var wrapper struct {
		MCPServers map[string]copilot.MCPServerConfig `json:"mcpServers"`
	}

	if err := json.Unmarshal([]byte(expanded), &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse MCP config: %w", err)
	}

	return wrapper.MCPServers, nil
}
