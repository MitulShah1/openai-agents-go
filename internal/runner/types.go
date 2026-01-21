// Package runner provides internal implementation details for agent execution.
package runner

import (
	"time"

	"github.com/openai/openai-go/v3"
)

// Constants for truncation and defaults
const (
	// MaxToolCallIDLength is the maximum length for tool call IDs
	MaxToolCallIDLength = 40
)

// ExecutionContext holds the state during agent execution
type ExecutionContext struct {
	// CurrentAgent is the agent currently executing.
	// It is typed as *interface{} to avoid a circular dependency with the agents package.
	// In practice, this always holds a *agents.Agent.
	CurrentAgent *any
	History      []openai.ChatCompletionMessageParamUnion
	Usage        any   // Will be agents.Usage
	Steps        []any // Will be []agents.Step
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
