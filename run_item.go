package agents

// RunItem represents a single item in an agent run.
// This can be a message, tool call, tool output, handoff, reasoning item, etc.
type RunItem struct {
	// Type indicates the kind of item (e.g., "message", "tool_call", "tool_output")
	Type string

	// Content contains the item's content (for messages)
	Content string

	// ToolCall contains tool call information (if Type is "tool_call")
	ToolCall *ToolCall

	// ToolOutput contains tool output information (if Type is "tool_output")
	ToolOutput *ToolOutput

	// Agent is the agent that created this item
	Agent *Agent

	// RawItem contains the raw underlying data
	RawItem any
}

// ToolOutput represents the output from a tool execution
type ToolOutput struct {
	// ToolCallID is the ID of the tool call this output corresponds to
	ToolCallID string

	// Output is the actual output from the tool
	Output string

	// Error contains any error that occurred during tool execution
	Error error
}
