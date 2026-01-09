package agents

import (
	"errors"
	"fmt"
)

var (
	// ErrMaxTurnsExceeded is returned when the agent loop exceeds MaxTurns
	ErrMaxTurnsExceeded = errors.New("max turns exceeded")

	// ErrTimeout is returned when agent execution exceeds timeout
	ErrTimeout = errors.New("agent execution timeout")

	// ErrNoMessages is returned when Run is called with empty messages
	ErrNoMessages = errors.New("no messages provided")
)

// ToolExecutionError wraps errors from tool execution
type ToolExecutionError struct {
	ToolName string
	Err      error
}

func (e *ToolExecutionError) Error() string {
	return fmt.Sprintf("tool %s failed: %v", e.ToolName, e.Err)
}

func (e *ToolExecutionError) Unwrap() error {
	return e.Err
}

// NewToolExecutionError creates a ToolExecutionError
func NewToolExecutionError(toolName string, err error) error {
	return &ToolExecutionError{
		ToolName: toolName,
		Err:      err,
	}
}

// OutputValidationError is returned when output doesn't match expected schema
type OutputValidationError struct {
	Expected string
	Got      string
	Err      error
}

func (e *OutputValidationError) Error() string {
	return fmt.Sprintf("output validation failed: expected %s, got %s: %v",
		e.Expected, e.Got, e.Err)
}

func (e *OutputValidationError) Unwrap() error {
	return e.Err
}
