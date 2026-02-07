//go:build !integration

package workflow

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFormatCompilerError tests the formatCompilerError helper function
func TestFormatCompilerError(t *testing.T) {
	tests := []struct {
		name        string
		filePath    string
		errType     string
		message     string
		cause       error
		wantContain []string
	}{
		{
			name:     "error type with simple message and no cause",
			filePath: "/path/to/workflow.md",
			errType:  "error",
			message:  "validation failed",
			cause:    nil,
			wantContain: []string{
				"/path/to/workflow.md",
				"1:1",
				"error",
				"validation failed",
			},
		},
		{
			name:     "error type with cause wrapping",
			filePath: "/path/to/workflow.md",
			errType:  "error",
			message:  "validation failed",
			cause:    errors.New("underlying error"),
			wantContain: []string{
				"/path/to/workflow.md",
				"1:1",
				"error",
				"validation failed",
				"underlying error",
			},
		},
		{
			name:     "warning type with detailed message",
			filePath: "/path/to/workflow.md",
			errType:  "warning",
			message:  "missing required permission",
			cause:    nil,
			wantContain: []string{
				"/path/to/workflow.md",
				"1:1",
				"warning",
				"missing required permission",
			},
		},
		{
			name:     "lock file path",
			filePath: "/path/to/workflow.lock.yml",
			errType:  "error",
			message:  "failed to write lock file",
			cause:    nil,
			wantContain: []string{
				"/path/to/workflow.lock.yml",
				"1:1",
				"error",
				"failed to write lock file",
			},
		},
		{
			name:     "formatted message with error details",
			filePath: "test.md",
			errType:  "error",
			message:  "failed to generate YAML",
			cause:    errors.New("syntax error"),
			wantContain: []string{
				"test.md",
				"1:1",
				"error",
				"failed to generate YAML",
				"syntax error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := formatCompilerError(tt.filePath, tt.errType, tt.message, tt.cause)
			require.Error(t, err, "formatCompilerError should return an error")

			errStr := err.Error()
			for _, want := range tt.wantContain {
				assert.Contains(t, errStr, want, "Error message should contain: %s", want)
			}
		})
	}
}

// TestFormatCompilerError_OutputFormat verifies the output format remains consistent
func TestFormatCompilerError_OutputFormat(t *testing.T) {
	err := formatCompilerError("/test/workflow.md", "error", "test message", nil)
	require.Error(t, err)

	errStr := err.Error()

	// Verify the error format contains the standard compiler error structure
	assert.Contains(t, errStr, "/test/workflow.md", "Should contain file path")
	assert.Contains(t, errStr, "1:1", "Should contain line:column")
	assert.Contains(t, errStr, "error", "Should contain error type")
	assert.Contains(t, errStr, "test message", "Should contain message")
}

// TestFormatCompilerError_ErrorVsWarning tests differentiation between error and warning types
func TestFormatCompilerError_ErrorVsWarning(t *testing.T) {
	errorErr := formatCompilerError("test.md", "error", "error message", nil)
	warningErr := formatCompilerError("test.md", "warning", "warning message", nil)

	require.Error(t, errorErr)
	require.Error(t, warningErr)

	assert.Contains(t, errorErr.Error(), "error", "Error type should be present")
	assert.Contains(t, warningErr.Error(), "warning", "Warning type should be present")

	// Ensure they produce different outputs
	assert.NotEqual(t, errorErr.Error(), warningErr.Error(), "Error and warning should have different outputs")
}

// TestFormatCompilerError_ErrorChaining tests that error chains are preserved
func TestFormatCompilerError_ErrorChaining(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := formatCompilerError("test.md", "error", "validation failed", originalErr)

	require.Error(t, wrappedErr)

	// Verify the error message contains both the formatted message and the original error
	assert.Contains(t, wrappedErr.Error(), "validation failed", "Should contain formatted message")
	assert.Contains(t, wrappedErr.Error(), "original error", "Should contain original error")

	// Verify error chain is preserved using errors.Is
	assert.True(t, errors.Is(wrappedErr, originalErr), "Should preserve error chain")

	// Verify error chain can be unwrapped
	unwrapped := errors.Unwrap(wrappedErr)
	assert.Equal(t, originalErr, unwrapped, "Should unwrap to original error")
}

// TestFormatCompilerMessage tests the formatCompilerMessage helper function
func TestFormatCompilerMessage(t *testing.T) {
	tests := []struct {
		name        string
		filePath    string
		msgType     string
		message     string
		wantContain []string
	}{
		{
			name:     "warning message",
			filePath: "/path/to/workflow.md",
			msgType:  "warning",
			message:  "container image validation failed",
			wantContain: []string{
				"/path/to/workflow.md",
				"1:1",
				"warning",
				"container image validation failed",
			},
		},
		{
			name:     "error message as string",
			filePath: "test.md",
			msgType:  "error",
			message:  "validation error",
			wantContain: []string{
				"test.md",
				"1:1",
				"error",
				"validation error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := formatCompilerMessage(tt.filePath, tt.msgType, tt.message)

			for _, want := range tt.wantContain {
				assert.Contains(t, msg, want, "Message should contain: %s", want)
			}
		})
	}
}
