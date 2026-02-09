package agents

import "github.com/MitulShah1/openai-agents-go/tools"

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

	// StreamEventApprovalRequired represents a tool approval request event.
	// When this event is emitted, the stream will terminate with a
	// ToolApprovalRequiredError. The caller should use Runner.Resume
	// to continue execution after making approval decisions.
	StreamEventApprovalRequired StreamEventType = "approval_required"
)

// StreamEvent represents a single event in the agent execution stream
type StreamEvent struct {
	Type StreamEventType

	// Content contains the text delta for token events
	Content string

	// ToolCall contains tool call information (delta or complete)
	ToolCall *ToolCall

	// ApprovalRequests contains pending approval requests (only for StreamEventApprovalRequired)
	ApprovalRequests []tools.ApprovalRequest

	// Result contains the final execution result (only for StreamEventResult)
	Result *Result

	// Error contains any error that occurred
	Error error
}

// StreamHandler is a function that processes stream events
type StreamHandler func(event StreamEvent)
