package agents

import (
	"github.com/MitulShah1/openai-agents-go/session"
	"github.com/MitulShah1/openai-agents-go/tools"
)

// RunOption configures a Run call.
type RunOption func(*runOptions)

// runOptions holds the configuration for a Run call.
type runOptions struct {
	config          *RunConfig
	contextParams   ContextVariables
	sess            session.Session
	sessionID       string
	approvalHandler tools.ApprovalHandler
}

// WithConfig sets the runtime configuration for the agent execution.
//
// Example:
//
//	runner.Run(ctx, agent, messages,
//	    agents.WithConfig(&agents.RunConfig{MaxTurns: 5}),
//	)
func WithConfig(config *RunConfig) RunOption {
	return func(o *runOptions) {
		o.config = config
	}
}

// WithContextVariables sets variables that are accessible to tools during execution.
//
// Example:
//
//	vars := agents.ContextVariables{"user_id": "123", "tier": "premium"}
//	runner.Run(ctx, agent, messages,
//	    agents.WithContextVariables(vars),
//	)
func WithContextVariables(vars ContextVariables) RunOption {
	return func(o *runOptions) {
		o.contextParams = vars
	}
}

// WithSession enables conversation persistence using the session package.
// The session stores conversation history across multiple Run calls.
//
// Example:
//
//	memSession := session.NewMemorySession()
//	runner.Run(ctx, agent, messages,
//	    agents.WithSession(memSession, "user_123"),
//	)
func WithSession(sess session.Session, sessionID string) RunOption {
	return func(o *runOptions) {
		o.sess = sess
		o.sessionID = sessionID
	}
}

// WithApprovalHandler sets a handler for tool approval requests.
// When a tool with NeedsApproval=true is called, the handler is invoked
// synchronously to approve or reject the execution.
//
// If no handler is set and a tool requires approval, Run returns a
// ToolApprovalRequiredError with a RunState that can be passed to Resume.
//
// Example:
//
//	runner.Run(ctx, agent, messages,
//	    agents.WithApprovalHandler(func(req tools.ApprovalRequest) (*tools.ApprovalResponse, error) {
//	        fmt.Printf("Approve %s with args %v? ", req.ToolName, req.Args)
//	        // ... get user input ...
//	        return &tools.ApprovalResponse{Approved: true}, nil
//	    }),
//	)
func WithApprovalHandler(handler tools.ApprovalHandler) RunOption {
	return func(o *runOptions) {
		o.approvalHandler = handler
	}
}
