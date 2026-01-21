package agents

import (
	"errors"
	"testing"
	"time"
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

func TestRateLimitError(t *testing.T) {
	t.Run("with retry after", func(t *testing.T) {
		err := &RateLimitError{
			Message:    "too many requests",
			RetryAfter: 5 * time.Second,
			Limit:      100,
			Remaining:  0,
		}

		expected := "rate limit exceeded: too many requests (retry after 5s)"
		if err.Error() != expected {
			t.Errorf("expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("without retry after", func(t *testing.T) {
		err := &RateLimitError{
			Message:   "quota exceeded",
			Limit:     1000,
			Remaining: 0,
		}

		expected := "rate limit exceeded: quota exceeded"
		if err.Error() != expected {
			t.Errorf("expected %q, got %q", expected, err.Error())
		}
	})
}

func TestTimeoutError(t *testing.T) {
	err := &TimeoutError{
		Message:  "API call timed out",
		Duration: 35 * time.Second,
		Timeout:  30 * time.Second,
	}

	if !errors.Is(err, err) {
		t.Error("TimeoutError should match itself")
	}

	expected := "request timeout after 35s (limit: 30s): API call timed out"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestNetworkError(t *testing.T) {
	t.Run("with cause", func(t *testing.T) {
		cause := errors.New("connection refused")
		err := &NetworkError{
			Message:   "failed to connect",
			Cause:     cause,
			Retryable: true,
		}

		if !errors.Is(err, cause) {
			t.Error("NetworkError should unwrap to cause")
		}
	})

	t.Run("without cause", func(t *testing.T) {
		err := &NetworkError{
			Message:   "dns resolution failed",
			Retryable: false,
		}

		expected := "network error: dns resolution failed"
		if err.Error() != expected {
			t.Errorf("expected %q, got %q", expected, err.Error())
		}
	})
}

func TestErrorContext(t *testing.T) {
	t.Run("with session ID", func(t *testing.T) {
		baseErr := errors.New("tool execution failed")
		err := &ErrorContext{
			AgentName:  "assistant",
			StepNumber: 5,
			SessionID:  "sess_123",
			Err:        baseErr,
		}

		expected := "[agent=assistant, step=5, session=sess_123] tool execution failed"
		if err.Error() != expected {
			t.Errorf("expected %q, got %q", expected, err.Error())
		}

		if !errors.Is(err, baseErr) {
			t.Error("ErrorContext should unwrap to base error")
		}
	})

	t.Run("without session ID", func(t *testing.T) {
		baseErr := errors.New("validation failed")
		err := WrapError(baseErr, "validator", 1, "")

		expected := "[agent=validator, step=1] validation failed"
		if err.Error() != expected {
			t.Errorf("expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("wrap nil error", func(t *testing.T) {
		err := WrapError(nil, "test", 1, "")
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})
}
