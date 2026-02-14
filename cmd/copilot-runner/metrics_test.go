package main

import (
	"os"
	"path/filepath"
	"testing"

	copilot "github.com/github/copilot-sdk/go"
)

func TestNewMetrics(t *testing.T) {
	m := NewMetrics()

	if m.InputTokens != 0 {
		t.Errorf("Expected 0 input tokens, got %d", m.InputTokens)
	}

	if m.OutputTokens != 0 {
		t.Errorf("Expected 0 output tokens, got %d", m.OutputTokens)
	}

	if m.Turns != 0 {
		t.Errorf("Expected 0 turns, got %d", m.Turns)
	}

	if m.TotalToolCalls != 0 {
		t.Errorf("Expected 0 total tool calls, got %d", m.TotalToolCalls)
	}

	if m.ToolCalls == nil {
		t.Error("Expected non-nil ToolCalls map")
	}

	if m.StartTime.IsZero() {
		t.Error("Expected non-zero start time")
	}
}

func TestMetricsHandleToolExecutionStart(t *testing.T) {
	m := NewMetrics()

	toolName := "shell"
	event := copilot.SessionEvent{
		Type: copilot.ToolExecutionStart,
		Data: copilot.Data{
			ToolName: &toolName,
		},
	}

	m.HandleEvent(event)
	m.HandleEvent(event) // Call twice

	if m.TotalToolCalls != 2 {
		t.Errorf("Expected 2 total tool calls, got %d", m.TotalToolCalls)
	}

	if m.ToolCalls["shell"] != 2 {
		t.Errorf("Expected 2 shell calls, got %d", m.ToolCalls["shell"])
	}
}

func TestMetricsHandleTurnStart(t *testing.T) {
	m := NewMetrics()

	event := copilot.SessionEvent{
		Type: copilot.AssistantTurnStart,
	}

	m.HandleEvent(event)
	m.HandleEvent(event)
	m.HandleEvent(event)

	if m.Turns != 3 {
		t.Errorf("Expected 3 turns, got %d", m.Turns)
	}
}

func TestMetricsHandleSessionStart(t *testing.T) {
	m := NewMetrics()

	sessionID := "test-session-123"
	model := "gpt-4o"
	event := copilot.SessionEvent{
		Type: copilot.SessionStart,
		Data: copilot.Data{
			SessionID:     &sessionID,
			SelectedModel: &model,
		},
	}

	m.HandleEvent(event)

	if m.SessionID != "test-session-123" {
		t.Errorf("Expected session ID 'test-session-123', got '%s'", m.SessionID)
	}

	if m.Model != "gpt-4o" {
		t.Errorf("Expected model 'gpt-4o', got '%s'", m.Model)
	}
}

func TestMetricsHandleSessionError(t *testing.T) {
	m := NewMetrics()

	errMsg := "something went wrong"
	event := copilot.SessionEvent{
		Type: copilot.SessionError,
		Data: copilot.Data{
			Message: &errMsg,
		},
	}

	m.HandleEvent(event)

	if len(m.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(m.Errors))
	}

	if m.Errors[0] != "something went wrong" {
		t.Errorf("Expected error message 'something went wrong', got '%s'", m.Errors[0])
	}
}

func TestMetricsFinalize(t *testing.T) {
	m := NewMetrics()
	m.Finalize()

	if m.EndTime.IsZero() {
		t.Error("Expected non-zero end time after finalize")
	}

	if m.Duration == "" {
		t.Error("Expected non-empty duration after finalize")
	}
}

func TestMetricsWriteToFile(t *testing.T) {
	m := NewMetrics()
	m.InputTokens = 100
	m.OutputTokens = 50
	m.TotalTokens = 150
	m.Turns = 3
	m.ToolCalls["shell"] = 5
	m.TotalToolCalls = 5
	m.Finalize()

	tmpDir := t.TempDir()
	metricsPath := filepath.Join(tmpDir, "metrics.json")

	if err := m.WriteToFile(metricsPath); err != nil {
		t.Fatalf("Failed to write metrics: %v", err)
	}

	// Verify file was created
	data, err := os.ReadFile(metricsPath)
	if err != nil {
		t.Fatalf("Failed to read metrics file: %v", err)
	}

	// Basic content checks
	content := string(data)
	if len(content) == 0 {
		t.Error("Expected non-empty metrics file")
	}

	// Check that key metrics are present in JSON
	if !contains(content, `"input_tokens": 100`) {
		t.Error("Expected input_tokens in metrics JSON")
	}

	if !contains(content, `"output_tokens": 50`) {
		t.Error("Expected output_tokens in metrics JSON")
	}

	if !contains(content, `"turns": 3`) {
		t.Error("Expected turns in metrics JSON")
	}
}

func TestMetricsConcurrency(t *testing.T) {
	m := NewMetrics()

	// Simulate concurrent event handling
	done := make(chan struct{})
	for i := 0; i < 10; i++ {
		go func() {
			toolName := "shell"
			event := copilot.SessionEvent{
				Type: copilot.ToolExecutionStart,
				Data: copilot.Data{
					ToolName: &toolName,
				},
			}
			m.HandleEvent(event)
			done <- struct{}{}
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	if m.TotalToolCalls != 10 {
		t.Errorf("Expected 10 total tool calls, got %d", m.TotalToolCalls)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
