package handoff

// Option configures a Handoff using the functional options pattern.
type Option func(*Handoff)

// WithToolName overrides the automatically generated tool name.
// By default, the tool name is derived from the agent name.
//
// Example:
//
//	handoff.New(agent, handoff.WithToolName("escalate_to_billing"))
func WithToolName(name string) Option {
	return func(h *Handoff) {
		h.toolName = name
	}
}

// WithDescription overrides the automatically generated tool description.
// A good description helps the LLM decide when to use this handoff.
//
// Example:
//
//	handoff.New(agent, handoff.WithDescription("Transfer to billing for payment issues"))
func WithDescription(desc string) Option {
	return func(h *Handoff) {
		h.description = desc
	}
}

// WithInputFilter sets a function to filter or transform input data before the handoff.
// The filter receives the current conversation state and can modify it.
//
// Common use cases:
//   - Redacting sensitive information (PII, credentials)
//   - Enriching context variables
//   - Validating preconditions
//
// Example:
//
//	handoff.New(agent, handoff.WithInputFilter(func(ctx context.Context, data handoff.InputData) (handoff.InputData, error) {
//	    // Redact credit card numbers
//	    filtered := redactSensitiveData(data.NewItems)
//	    return handoff.InputData{
//	        Agent:       data.Agent,
//	        NewItems:    filtered,
//	        ContextVars: data.ContextVars,
//	    }, nil
//	}))
func WithInputFilter(filter InputFilterFunc) Option {
	return func(h *Handoff) {
		h.inputFilter = filter
	}
}

// WithHistoryNesting enables or disables conversation history nesting.
// When enabled, the conversation history before the handoff is summarized
// into a single assistant message, reducing token usage for the next agent.
//
// When disabled (default), the full conversation history is passed through.
//
// Example:
//
//	handoff.New(agent, handoff.WithHistoryNesting(true))
func WithHistoryNesting(enabled bool) Option {
	return func(h *Handoff) {
		h.nestHistory = enabled
	}
}

// WithEnabledPredicate sets a function that determines if the handoff is available.
// The predicate is evaluated at runtime, allowing dynamic handoff availability.
//
// Use cases:
//   - Business hours restrictions
//   - Feature flags
//   - User permissions
//   - Resource availability
//
// If the predicate returns false, the tool will not be available for that request.
// If it returns an error, the error will be logged but the tool will be disabled.
//
// Example:
//
//	handoff.New(agent, handoff.WithEnabledPredicate(func(ctx context.Context, agent *agents.Agent, vars agents.ContextVariables) (bool, error) {
//	    // Only enable during business hours
//	    hour := time.Now().Hour()
//	    return hour >= 9 && hour < 17, nil
//	}))
func WithEnabledPredicate(predicate EnabledPredicateFunc) Option {
	return func(h *Handoff) {
		h.isEnabled = predicate
	}
}
