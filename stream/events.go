package stream

// Event is the interface that all streaming events implement.
// This allows for type switching on events while maintaining type safety.
type Event interface {
	// EventType returns the type identifier for this event
	EventType() string
}

// RawResponseEvent wraps raw LLM streaming events from the OpenAI Responses API.
// These are 'raw' events, i.e. they are directly passed through from the LLM.
// The Data field contains the actual event data in OpenAI Responses API format.
type RawResponseEvent struct {
	// Type is the event type (e.g., "response.created", "response.output_text.delta")
	Type string

	// Data contains the raw response event data from the LLM
	// This can be various OpenAI Responses API event types
	Data any

	// SequenceNumber is used for ordering events
	SequenceNumber int
}

// EventType implements the Event interface
func (e *RawResponseEvent) EventType() string {
	return "raw_response_event"
}

// RunItemEvent represents streaming events that wrap a RunItem.
// As the agent processes the LLM response, it generates these events for
// new messages, tool calls, tool outputs, handoffs, etc.
type RunItemEvent struct {
	// Name is the specific event name (e.g., "message_output_created", "tool_called")
	Name string

	// Item contains the run item data (type is any to avoid import cycle)
	Item any

	// SequenceNumber is used for ordering events
	SequenceNumber int
}

// EventType implements the Event interface
func (e *RunItemEvent) EventType() string {
	return "run_item_stream_event"
}

// RunItemEventName represents the different types of RunItemEvents
type RunItemEventName string

const (
	// MessageOutputCreated indicates a new message output was created
	MessageOutputCreated RunItemEventName = "message_output_created"

	// HandoffRequested indicates a handoff was requested
	HandoffRequested RunItemEventName = "handoff_requested"

	// HandoffOccurred indicates a handoff occurred
	// Note: This is intentionally misspelled to match Python SDK for compatibility
	//nolint:misspell
	HandoffOccurred RunItemEventName = "handoff_occured"

	// ToolCalled indicates a tool was called
	ToolCalled RunItemEventName = "tool_called"

	// ToolOutput indicates a tool output was received
	ToolOutput RunItemEventName = "tool_output"

	// ReasoningItemCreated indicates a reasoning item was created
	ReasoningItemCreated RunItemEventName = "reasoning_item_created"

	// MCPApprovalRequested indicates MCP approval was requested
	MCPApprovalRequested RunItemEventName = "mcp_approval_requested"

	// MCPApprovalResponse indicates MCP approval response was received
	MCPApprovalResponse RunItemEventName = "mcp_approval_response"

	// MCPListTools indicates MCP tools were listed
	MCPListTools RunItemEventName = "mcp_list_tools"
)

// ApprovalRequiredEvent notifies that one or more tool calls require approval
// before execution can continue. The caller should use Runner.Resume to continue
// execution after making approval decisions.
type ApprovalRequiredEvent struct {
	// Requests contains the pending approval requests (type is any to avoid import cycle)
	Requests any

	// SequenceNumber is used for ordering events
	SequenceNumber int
}

// EventType implements the Event interface
func (e *ApprovalRequiredEvent) EventType() string {
	return "approval_required_event"
}

// AgentUpdatedEvent notifies that there is a new agent running.
// This typically occurs during handoffs.
type AgentUpdatedEvent struct {
	// NewAgent is the agent that is now running (type is any to avoid import cycle)
	NewAgent any

	// SequenceNumber is used for ordering events
	SequenceNumber int
}

// EventType implements the Event interface
func (e *AgentUpdatedEvent) EventType() string {
	return "agent_updated_stream_event"
}
