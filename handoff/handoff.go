// Package handoff provides high-level abstractions for agent-to-agent transfers.
//
// This package enables declarative handoff configuration with features like:
//   - Automatic tool generation from target agents
//   - Input filtering to transform data before transfer
//   - History nesting to summarize conversations
//   - Dynamic enablement predicates for conditional handoffs
//
// Example usage:
//
//	supportAgent := agents.NewAgent("Support")
//
//	// Create a handoff with history nesting
//	toSupport := handoff.New(
//	    supportAgent,
//	    handoff.WithDescription("Transfer to technical support"),
//	    handoff.WithHistoryNesting(true),
//	).ToTool()
//
//	// Add the handoff tool to your agent
//	triageAgent.Tools = []agents.Tool{toSupport}
package handoff

import (
	"context"

	"github.com/openai/openai-go/v3"

	agents "github.com/MitulShah1/openai-agents-go"
)

// Handoff represents a transfer from one agent to another.
// It encapsulates the target agent and configuration options for how
// the handoff should behave (filtering, history nesting, enablement).
type Handoff struct {
	agent       *agents.Agent
	toolName    string
	description string

	// Configuration options
	inputFilter InputFilterFunc
	nestHistory bool
	isEnabled   EnabledPredicateFunc
}

// InputData contains the context and state passed to input filters.
// Filters can inspect and modify this data before the handoff executes.
type InputData struct {
	// Agent is the target agent for the handoff
	Agent *agents.Agent

	// NewItems are the new conversation messages being passed to the agent
	NewItems []openai.ChatCompletionMessageParamUnion

	// ContextVars are the context variables available during execution
	ContextVars agents.ContextVariables
}

// InputFilterFunc is a function that filters or transforms input data before a handoff.
// It can modify the messages, context variables, or even change the target agent.
// Returning an error will prevent the handoff from executing.
type InputFilterFunc func(ctx context.Context, data InputData) (InputData, error)

// EnabledPredicateFunc determines if a handoff is enabled at runtime.
// This allows dynamic control over when handoffs are available based on
// context variables, agent state, or external conditions (e.g., business hours).
// Returning false will make the tool unavailable without generating an error.
type EnabledPredicateFunc func(ctx context.Context, agent *agents.Agent, vars agents.ContextVariables) (bool, error)

// New creates a new Handoff to the specified target agent.
// The handoff can be configured using functional options.
//
// By default:
//   - Tool name is generated from agent name (e.g., "transfer_to_support")
//   - Description is generated from agent name
//   - No input filtering
//   - History is flattened (not nested)
//   - Handoff is always enabled
//
// Example:
//
//	h := handoff.New(
//	    supportAgent,
//	    handoff.WithToolName("escalate_to_support"),
//	    handoff.WithHistoryNesting(true),
//	)
func New(targetAgent *agents.Agent, opts ...Option) *Handoff {
	h := &Handoff{
		agent:       targetAgent,
		nestHistory: false,
		isEnabled:   nil, // nil means always enabled
	}

	// Apply options
	for _, opt := range opts {
		opt(h)
	}

	return h
}

// TargetAgent returns the agent this handoff transfers to.
func (h *Handoff) TargetAgent() *agents.Agent {
	return h.agent
}

// NestHistory returns whether history nesting is enabled for this handoff.
func (h *Handoff) NestHistory() bool {
	return h.nestHistory
}

// InputFilter returns the input filter function, or nil if not set.
func (h *Handoff) InputFilter() InputFilterFunc {
	return h.inputFilter
}

// EnabledPredicate returns the enabled predicate function, or nil if always enabled.
func (h *Handoff) EnabledPredicate() EnabledPredicateFunc {
	return h.isEnabled
}
