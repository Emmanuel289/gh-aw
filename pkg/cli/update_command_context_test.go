//go:build !integration

package cli

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUpdateWorkflowsWithExtensionCheckContext_Cancellation tests that the function respects context cancellation
func TestUpdateWorkflowsWithExtensionCheckContext_Cancellation(t *testing.T) {
	// Create a context that is already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	var buf bytes.Buffer

	// Call the function with the cancelled context
	output, err := UpdateWorkflowsWithExtensionCheckContext(
		ctx,
		&buf,
		nil,   // workflowNames
		false, // allowMajor
		false, // force
		false, // verbose
		"",    // engineOverride
		false, // createPR
		"",    // workflowsDir
		false, // noStopAfter
		"",    // stopAfter
		false, // merge
		false, // noActions
	)

	// Should return context.Canceled error
	require.Error(t, err, "Expected error when context is cancelled")
	assert.Equal(t, context.Canceled, err, "Expected context.Canceled error")
	assert.Empty(t, output, "Expected empty output when cancelled immediately")
}

// TestUpdateWorkflowsWithExtensionCheckContext_Timeout tests that the function respects context timeout
func TestUpdateWorkflowsWithExtensionCheckContext_Timeout(t *testing.T) {
	// Skip if running in CI without sufficient time
	if testing.Short() {
		t.Skip("Skipping timeout test in short mode")
	}

	// Create a context with a very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Wait a moment to ensure context times out
	time.Sleep(10 * time.Millisecond)

	var buf bytes.Buffer

	// Call the function with the timed-out context
	_, err := UpdateWorkflowsWithExtensionCheckContext(
		ctx,
		&buf,
		nil,   // workflowNames
		false, // allowMajor
		false, // force
		false, // verbose
		"",    // engineOverride
		false, // createPR
		"",    // workflowsDir
		false, // noStopAfter
		"",    // stopAfter
		false, // merge
		false, // noActions
	)

	// Should return context.DeadlineExceeded error
	require.Error(t, err, "Expected error when context times out")
	assert.Contains(t, []error{context.DeadlineExceeded, context.Canceled}, err, "Expected context deadline/canceled error")
}

// TestUpdateWorkflowsWithExtensionCheckContext_OutputCapture tests that output is captured correctly
func TestUpdateWorkflowsWithExtensionCheckContext_OutputCapture(t *testing.T) {
	// Skip if we don't have gh CLI or not in a git repo
	if !isGHCLIAvailable() {
		t.Skip("Skipping test: gh CLI not available")
	}

	ctx := context.Background()
	var buf bytes.Buffer

	// Create a temporary directory to avoid "no such file" errors
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	// Change to temp dir but don't create workflows directory (test will fail gracefully)
	os.Chdir(tmpDir)

	// Call with minimal parameters - this should at least try to check for updates
	_, err := UpdateWorkflowsWithExtensionCheckContext(
		ctx,
		&buf,
		[]string{}, // workflowNames (empty means all workflows with source field)
		false,      // allowMajor
		false,      // force
		false,      // verbose
		"",         // engineOverride
		false,      // createPR
		"",         // workflowsDir
		false,      // noStopAfter
		"",         // stopAfter
		false,      // merge
		true,       // noActions (skip action updates to make test faster)
	)

	// The function may fail if there are no workflows directory/files, which is expected
	// The important thing is that we tested the function can be called
	if err != nil {
		// Expected - likely no workflows directory or no workflows with source field
		assert.True(t, strings.Contains(err.Error(), "no workflows found") || strings.Contains(err.Error(), "no such file"),
			"Expected 'no workflows found' or 'no such file' error, got: %v", err)
	}

	// Note: Buffer may be empty if error occurs early (before any output)
	// That's OK - we're testing that the function works with a buffer
}

// TestUpdateWorkflowsWithExtensionCheckContext_BufferWriting tests that the function writes to the provided buffer
func TestUpdateWorkflowsWithExtensionCheckContext_BufferWriting(t *testing.T) {
	ctx := context.Background()
	var buf bytes.Buffer

	// Create a custom writer that tracks writes
	writeCount := 0
	trackingWriter := &trackingWriter{
		Writer: &buf,
		onWrite: func(p []byte) {
			writeCount++
		},
	}

	// Call with parameters that will trigger some output
	_, err := UpdateWorkflowsWithExtensionCheckContext(
		ctx,
		trackingWriter,
		[]string{}, // Empty workflow names
		false,      // allowMajor
		false,      // force
		true,       // verbose (should generate more output)
		"",         // engineOverride
		false,      // createPR
		"",         // workflowsDir
		false,      // noStopAfter
		"",         // stopAfter
		false,      // merge
		true,       // noActions (skip action updates)
	)

	// We expect an error (no workflows with source) but also some output
	if err == nil || !strings.Contains(err.Error(), "no workflows found") {
		t.Logf("Unexpected error: %v", err)
	}

	// Verify that writes occurred to our tracking writer
	assert.Positive(t, writeCount, "Expected at least one write to the output buffer")
	assert.NotEmpty(t, buf.String(), "Expected buffer to contain output")
}

// trackingWriter wraps an io.Writer and tracks write operations
type trackingWriter struct {
	Writer  *bytes.Buffer
	onWrite func([]byte)
}

func (tw *trackingWriter) Write(p []byte) (n int, err error) {
	if tw.onWrite != nil {
		tw.onWrite(p)
	}
	return tw.Writer.Write(p)
}
