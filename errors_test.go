package agents

import (
	"errors"
	"testing"
)

func TestSentinelErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		msg  string
	}{
		{
			name: "ErrMaxTurnsExceeded",
			err:  ErrMaxTurnsExceeded,
			msg:  "max turns exceeded",
		},
		{
			name: "ErrTimeout",
			err:  ErrTimeout,
			msg:  "agent execution timeout",
		},
		{
			name: "ErrNoMessages",
			err:  ErrNoMessages,
			msg:  "no messages provided",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.msg {
				t.Errorf("expected error message %q, got %q", tt.msg, tt.err.Error())
			}
		})
	}
}

func TestToolExecutionError(t *testing.T) {
	baseErr := errors.New("connection failed")
	toolErr := NewToolExecutionError("get_weather", baseErr)

	// Test Error() method
	expected := "tool get_weather failed: connection failed"
	if toolErr.Error() != expected {
		t.Errorf("expected %q, got %q", expected, toolErr.Error())
	}

	// Test Unwrap
	if !errors.Is(toolErr, baseErr) {
		t.Error("expected error to unwrap to base error")
	}

	// Test type assertion
	var te *ToolExecutionError
	if !errors.As(toolErr, &te) {
		t.Error("expected error to be ToolExecutionError type")
	}

	if te.ToolName != "get_weather" {
		t.Errorf("expected ToolName=get_weather, got %s", te.ToolName)
	}
}

func TestOutputValidationError(t *testing.T) {
	baseErr := errors.New("type mismatch")
	validationErr := &OutputValidationError{
		Expected: "string",
		Got:      "number",
		Err:      baseErr,
	}

	// Test Error() method
	expected := "output validation failed: expected string, got number: type mismatch"
	if validationErr.Error() != expected {
		t.Errorf("expected %q, got %q", expected, validationErr.Error())
	}

	// Test Unwrap
	if !errors.Is(validationErr, baseErr) {
		t.Error("expected error to unwrap to base error")
	}
}
