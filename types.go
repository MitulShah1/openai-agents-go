package agents

import (
	"time"

	"github.com/openai/openai-go"
)

// Result is the output of running an agent.
type Result struct {
	// Messages is the conversation history.
	Messages []openai.ChatCompletionMessageParamUnion

	// Agent is the final agent that handled the request.
	Agent *Agent

	// Usage contains token usage statistics
	Usage Usage

	// Steps records the execution trace
	Steps []Step

	// FinalOutput is the last assistant message content
	FinalOutput string
}

// Usage tracks token consumption and costs
type Usage struct {
	// PromptTokens used across all LLM calls
	PromptTokens int

	// CompletionTokens generated across all LLM calls
	CompletionTokens int

	// TotalTokens = PromptTokens + CompletionTokens
	TotalTokens int
}

// Add combines usage from multiple calls
func (u *Usage) Add(other Usage) {
	u.PromptTokens += other.PromptTokens
	u.CompletionTokens += other.CompletionTokens
	u.TotalTokens += other.TotalTokens
}

// Step represents one iteration of the agent loop
type Step struct {
	// Agent that executed this step
	AgentName string

	// ToolCalls made during this step
	ToolCalls []ToolCall

	// Duration of this step
	Duration time.Duration

	// StepNumber in the execution sequence
	StepNumber int
}

// ToolCall represents a tool execution
type ToolCall struct {
	// ToolName that was called
	ToolName string

	// Arguments passed to the tool (JSON string)
	Arguments string

	// Result returned from the tool
	Result any

	// Error if tool execution failed
	Error error

	// Duration of tool execution
	Duration time.Duration
}

// ContextVariables is a map of variables that can be passed to functions.
type ContextVariables map[string]any
