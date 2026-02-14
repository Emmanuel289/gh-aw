package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	copilot "github.com/github/copilot-sdk/go"
)

// Metrics collects runtime metrics from the SDK session events.
// It is safe for concurrent use.
type Metrics struct {
	mu sync.Mutex

	// Token usage
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`

	// Turn tracking
	Turns int `json:"turns"`

	// Tool call tracking
	ToolCalls      map[string]int `json:"tool_calls"`
	TotalToolCalls int            `json:"total_tool_calls"`

	// Timing
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Duration  string    `json:"duration,omitempty"`

	// Session info
	SessionID string `json:"session_id,omitempty"`
	Model     string `json:"model,omitempty"`

	// Errors
	Errors []string `json:"errors,omitempty"`
}

// NewMetrics creates a new Metrics instance.
func NewMetrics() *Metrics {
	return &Metrics{
		ToolCalls: make(map[string]int),
		StartTime: time.Now(),
	}
}

// HandleEvent processes a session event and updates metrics accordingly.
// This is designed to be used as a session event handler callback.
func (m *Metrics) HandleEvent(event copilot.SessionEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()

	switch event.Type {
	case copilot.AssistantUsage:
		// Extract token usage from usage events
		// The usage event has ModelMetrics with per-model token counts
		if event.Data.ModelMetrics != nil {
			for _, metric := range event.Data.ModelMetrics {
				m.InputTokens += int(metric.Usage.InputTokens)
				m.OutputTokens += int(metric.Usage.OutputTokens)
				m.TotalTokens = m.InputTokens + m.OutputTokens
			}
		}
		// Also check direct InputTokens/OutputTokens fields on the event data
		if event.Data.InputTokens != nil {
			m.InputTokens += int(*event.Data.InputTokens)
		}
		if event.Data.OutputTokens != nil {
			m.OutputTokens += int(*event.Data.OutputTokens)
			m.TotalTokens = m.InputTokens + m.OutputTokens
		}

	case copilot.ToolExecutionStart:
		// Track tool invocations
		if event.Data.ToolName != nil {
			toolName := *event.Data.ToolName
			m.ToolCalls[toolName]++
			m.TotalToolCalls++
		}

	case copilot.AssistantTurnStart:
		m.Turns++

	case copilot.SessionStart:
		if event.Data.SessionID != nil {
			m.SessionID = *event.Data.SessionID
		}
		if event.Data.SelectedModel != nil {
			m.Model = *event.Data.SelectedModel
		}

	case copilot.SessionError:
		if event.Data.Message != nil {
			m.Errors = append(m.Errors, *event.Data.Message)
		}
	}
}

// Finalize marks the end time and calculates duration.
func (m *Metrics) Finalize() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.EndTime = time.Now()
	m.Duration = m.EndTime.Sub(m.StartTime).String()
}

// WriteToFile writes the metrics as JSON to the specified file path.
func (m *Metrics) WriteToFile(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write metrics file %s: %w", path, err)
	}

	return nil
}

// PrintSummary outputs a human-readable summary of the metrics to stderr.
func (m *Metrics) PrintSummary() {
	m.mu.Lock()
	defer m.mu.Unlock()

	fmt.Fprintf(os.Stderr, "\n=== Copilot SDK Runner Metrics ===\n")
	if m.Model != "" {
		fmt.Fprintf(os.Stderr, "Model: %s\n", m.Model)
	}
	fmt.Fprintf(os.Stderr, "Turns: %d\n", m.Turns)
	fmt.Fprintf(os.Stderr, "Tokens: %d (input: %d, output: %d)\n", m.TotalTokens, m.InputTokens, m.OutputTokens)
	fmt.Fprintf(os.Stderr, "Tool calls: %d\n", m.TotalToolCalls)
	for tool, count := range m.ToolCalls {
		fmt.Fprintf(os.Stderr, "  %s: %d\n", tool, count)
	}
	if m.Duration != "" {
		fmt.Fprintf(os.Stderr, "Duration: %s\n", m.Duration)
	}
	if len(m.Errors) > 0 {
		fmt.Fprintf(os.Stderr, "Errors: %d\n", len(m.Errors))
		for _, e := range m.Errors {
			fmt.Fprintf(os.Stderr, "  - %s\n", e)
		}
	}
	fmt.Fprintf(os.Stderr, "=================================\n\n")
}
