package agents

import (
	"errors"
	"fmt"
	"time"
)

var (
	// ErrMaxTurnsExceeded is returned when the agent loop exceeds MaxTurns
	ErrMaxTurnsExceeded = errors.New("max turns exceeded")

	// ErrTimeout is returned when agent execution exceeds timeout
	ErrTimeout = errors.New("agent execution timeout")

	// ErrNoMessages is returned when Run is called with empty messages
	ErrNoMessages = errors.New("no messages provided")

	// ErrEmptyModelResponse is returned when a model provider returns a response with no choices
	ErrEmptyModelResponse = errors.New("model returned empty response with no choices")
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

// RateLimitError represents an API rate limit error with retry information.
type RateLimitError struct {
	// Message describes the rate limit error
	Message string
	// RetryAfter indicates when the request can be retried
	RetryAfter time.Duration
	// Limit is the rate limit threshold that was exceeded
	Limit int
	// Remaining is the number of requests remaining (usually 0)
	Remaining int
}

func (e *RateLimitError) Error() string {
	if e.RetryAfter > 0 {
		return fmt.Sprintf("rate limit exceeded: %s (retry after %v)", e.Message, e.RetryAfter)
	}
	return fmt.Sprintf("rate limit exceeded: %s", e.Message)
}

// TimeoutError represents a request timeout error.
type TimeoutError struct {
	// Message describes the timeout
	Message string
	// Duration is how long the request took before timing out
	Duration time.Duration
	// Timeout is the configured timeout value
	Timeout time.Duration
}

func (e *TimeoutError) Error() string {
	return fmt.Sprintf("request timeout after %v (limit: %v): %s", e.Duration, e.Timeout, e.Message)
}

// NetworkError represents a network-level failure.
type NetworkError struct {
	// Message describes the network error
	Message string
	// Cause is the underlying error
	Cause error
	// Retryable indicates if the error is transient and can be retried
	Retryable bool
}

func (e *NetworkError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("network error: %s: %v", e.Message, e.Cause)
	}
	return fmt.Sprintf("network error: %s", e.Message)
}

func (e *NetworkError) Unwrap() error {
	return e.Cause
}

// ErrorContext wraps an error with contextual information about the agent execution.
type ErrorContext struct {
	// AgentName is the name of the agent that encountered the error
	AgentName string
	// StepNumber is the execution step where the error occurred
	StepNumber int
	// SessionID is the session identifier (if applicable)
	SessionID string
	// Err is the underlying error
	Err error
}

func (e *ErrorContext) Error() string {
	context := fmt.Sprintf("[agent=%s, step=%d", e.AgentName, e.StepNumber)
	if e.SessionID != "" {
		context += fmt.Sprintf(", session=%s", e.SessionID)
	}
	context += "]"
	return fmt.Sprintf("%s %s", context, e.Err.Error())
}

func (e *ErrorContext) Unwrap() error {
	return e.Err
}

// WrapError wraps an error with execution context.
func WrapError(err error, agentName string, stepNumber int, sessionID string) error {
	if err == nil {
		return nil
	}
	return &ErrorContext{
		AgentName:  agentName,
		StepNumber: stepNumber,
		SessionID:  sessionID,
		Err:        err,
	}
}
