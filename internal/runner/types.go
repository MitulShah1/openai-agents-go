// Package runner provides internal implementation details for agent execution.
package runner

import (
	"time"

	"github.com/openai/openai-go"
)

// Constants for truncation and defaults
const (
	// MaxToolCallIDLength is the maximum length for tool call IDs
	MaxToolCallIDLength = 40
)

// ExecutionContext holds the state during agent execution
type ExecutionContext struct {
	CurrentAgent *interface{} // Will be *agents.Agent but avoiding circular import
	History      []openai.ChatCompletionMessageParamUnion
	Usage        interface{}   // Will be agents.Usage
	Steps        []interface{} // Will be []agents.Step
	TurnCount    int
	LastMessage  openai.ChatCompletionMessage
}

// StepMetrics tracks metrics for a single execution step
type StepMetrics struct {
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
}

// ToolCallResult represents the result of a tool execution
type ToolCallResult struct {
	ToolName  string
	Arguments string
	Result    any
	Error     error
	Duration  time.Duration
}
