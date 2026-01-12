package agents

import "github.com/MitulShah1/openai-agents-go/session"

// RunOption configures a Run call.
type RunOption func(*runOptions)

// runOptions holds the configuration for a Run call.
type runOptions struct {
	config        *RunConfig
	contextParams ContextVariables
	sess          session.Session
	sessionID     string
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
