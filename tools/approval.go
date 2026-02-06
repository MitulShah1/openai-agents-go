package tools

// ApprovalRequest represents a request for human approval of a tool execution.
type ApprovalRequest struct {
	// ToolName is the name of the tool requesting approval
	ToolName string

	// CallID is the unique identifier for this tool call
	CallID string

	// Args are the arguments that will be passed to the tool
	Args map[string]any

	// Context contains the context variables for this execution
	Context ContextVariables
}

// ApprovalResponse represents the response to an approval request.
type ApprovalResponse struct {
	// Approved indicates whether the tool execution is approved
	Approved bool

	// Reason optionally provides a reason for rejection (if Approved is false)
	Reason string
}

// ApprovalHandler is a function that handles approval requests.
// It receives an ApprovalRequest and returns an ApprovalResponse.
// This allows for custom approval logic (e.g., UI prompts, logging, external validation).
type ApprovalHandler func(req ApprovalRequest) (*ApprovalResponse, error)
