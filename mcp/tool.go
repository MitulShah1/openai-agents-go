package mcp

// HostedMCPTool represents a tool hosted on an MCP server.
type HostedMCPTool struct {
	Name        string
	Description string
	Schema      map[string]any
	Client      *Client
}

// ApprovalRequest represents a request for tool execution approval.
type ApprovalRequest struct {
	ToolName string
	Args     map[string]any
}

// ApprovalFunction is a function that approves or denies tool execution.
type ApprovalFunction func(ApprovalRequest) bool
