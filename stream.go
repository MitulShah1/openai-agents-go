package agents

// StreamEventType represents the type of event in a stream
type StreamEventType string

const (
	// StreamEventToken represents a content token
	StreamEventToken StreamEventType = "token"

	// StreamEventToolCall represents a tool call event
	StreamEventToolCall StreamEventType = "tool_call"

	// StreamEventError represents an error event
	StreamEventError StreamEventType = "error"

	// StreamEventResult represents the final result event
	StreamEventResult StreamEventType = "result"
)

// StreamEvent represents a single event in the agent execution stream
type StreamEvent struct {
	Type StreamEventType

	// Content contains the text delta for token events
	Content string

	// ToolCall contains tool call information (delta or complete)
	// For now, we might just emit the full tool call when ready, or deltas.
	// Let's start with simplifying: for now, streaming usually means text content.
	// Tool execution steps might be better as status updates.
	ToolCall *ToolCall

	// Result contains the final execution result (only for StreamEventResult)
	Result *Result

	// Error contains any error that occurred
	Error error
}

// StreamHandler is a function that processes stream events
type StreamHandler func(event StreamEvent)
