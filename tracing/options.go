package tracing

import "github.com/MitulShah1/openai-agents-go/tracing/schema"

// TraceOption configures a Trace.
type TraceOption func(*traceConfig)

type traceConfig struct {
	traceID              string
	workflowName         string
	groupID              string
	metadata             map[string]any
	includeSensitiveData bool
	exportAPIKey         string
	disabled             bool
}

// WithTraceID sets a custom trace ID.
// The ID must follow the format: trace_<32 hex characters>
func WithTraceID(id string) TraceOption {
	return func(c *traceConfig) {
		c.traceID = id
	}
}

// WithWorkflowName sets the workflow name.
func WithWorkflowName(name string) TraceOption {
	return func(c *traceConfig) {
		c.workflowName = name
	}
}

// WithGroupID sets a group ID for correlating related traces.
// The ID must follow the format: group_<24 hex characters>
func WithGroupID(id string) TraceOption {
	return func(c *traceConfig) {
		c.groupID = id
	}
}

// WithMetadata adds metadata to the trace.
func WithMetadata(metadata map[string]any) TraceOption {
	return func(c *traceConfig) {
		c.metadata = metadata
	}
}

// WithSensitiveData controls whether sensitive data is included in exports.
func WithSensitiveData(include bool) TraceOption {
	return func(c *traceConfig) {
		c.includeSensitiveData = include
	}
}

// WithExportAPIKey sets the API key for exporting traces.
func WithExportAPIKey(apiKey string) TraceOption {
	return func(c *traceConfig) {
		c.exportAPIKey = apiKey
	}
}

// Disabled creates a no-op trace that doesn't record or export anything.
func Disabled() TraceOption {
	return func(c *traceConfig) {
		c.disabled = true
	}
}

// SpanOption configures a Span.
type SpanOption func(*spanConfig)

type spanConfig struct {
	name       string
	attributes map[string]any
}

// WithName sets the span name (for agent, function, guardrail spans).
func WithName(name string) SpanOption {
	return func(c *spanConfig) {
		c.name = name
	}
}

// WithSpanName is an alias for WithName for consistency with factory functions.
func WithSpanName(name string) SpanOption {
	return WithName(name)
}

// WithModel sets the model (for agent, generation spans).
func WithModel(model string) SpanOption {
	return func(c *spanConfig) {
		if c.attributes == nil {
			c.attributes = make(map[string]any)
		}
		c.attributes["model"] = model
	}
}

// WithInstructions sets instructions (for agent spans).
func WithInstructions(instructions string) SpanOption {
	return func(c *spanConfig) {
		if c.attributes == nil {
			c.attributes = make(map[string]any)
		}
		c.attributes["instructions"] = instructions
	}
}

// WithRequest sets the request (for generation spans).
func WithRequest(request any) SpanOption {
	return func(c *spanConfig) {
		if c.attributes == nil {
			c.attributes = make(map[string]any)
		}
		c.attributes["request"] = request
	}
}

// WithResponse sets the response (for generation spans).
func WithResponse(response any) SpanOption {
	return func(c *spanConfig) {
		if c.attributes == nil {
			c.attributes = make(map[string]any)
		}
		c.attributes["response"] = response
	}
}

// WithUsage sets the usage (for generation spans).
func WithUsage(usage *schema.Usage) SpanOption {
	return func(c *spanConfig) {
		if c.attributes == nil {
			c.attributes = make(map[string]any)
		}
		c.attributes["usage"] = usage
	}
}

// WithArguments sets arguments (for function spans).
func WithArguments(args any) SpanOption {
	return func(c *spanConfig) {
		if c.attributes == nil {
			c.attributes = make(map[string]any)
		}
		c.attributes["arguments"] = args
	}
}

// WithOutput sets output (for function spans).
func WithOutput(output any) SpanOption {
	return func(c *spanConfig) {
		if c.attributes == nil {
			c.attributes = make(map[string]any)
		}
		c.attributes["output"] = output
	}
}

// WithGuardrailType sets the guardrail type.
func WithGuardrailType(guardType string) SpanOption {
	return func(c *spanConfig) {
		if c.attributes == nil {
			c.attributes = make(map[string]any)
		}
		c.attributes["guardrail_type"] = guardType
	}
}

// WithTripwireTriggered sets whether the tripwire was triggered.
func WithTripwireTriggered(triggered bool) SpanOption {
	return func(c *spanConfig) {
		if c.attributes == nil {
			c.attributes = make(map[string]any)
		}
		c.attributes["tripwire_triggered"] = triggered
	}
}

// WithMessage sets a message (for guardrail spans).
func WithMessage(message string) SpanOption {
	return func(c *spanConfig) {
		if c.attributes == nil {
			c.attributes = make(map[string]any)
		}
		c.attributes["message"] = message
	}
}

// WithHandoff sets handoff details (for handoff spans).
func WithHandoff(fromAgent, toAgent, reason string) SpanOption {
	return func(c *spanConfig) {
		if c.attributes == nil {
			c.attributes = make(map[string]any)
		}
		c.attributes["from_agent"] = fromAgent
		c.attributes["to_agent"] = toAgent
		c.attributes["reason"] = reason
	}
}

// WithFromAgent sets the from_agent attribute for handoff spans.
func WithFromAgent(agent string) SpanOption {
	return func(c *spanConfig) {
		if c.attributes == nil {
			c.attributes = make(map[string]any)
		}
		c.attributes["from_agent"] = agent
	}
}

// WithToAgent sets the to_agent attribute for handoff spans.
func WithToAgent(agent string) SpanOption {
	return func(c *spanConfig) {
		if c.attributes == nil {
			c.attributes = make(map[string]any)
		}
		c.attributes["to_agent"] = agent
	}
}

// WithReason sets the reason attribute for handoff spans.
func WithReason(reason string) SpanOption {
	return func(c *spanConfig) {
		if c.attributes == nil {
			c.attributes = make(map[string]any)
		}
		c.attributes["reason"] = reason
	}
}

// WithAttributes sets custom attributes on the span.
func WithAttributes(attrs map[string]any) SpanOption {
	return func(c *spanConfig) {
		if c.attributes == nil {
			c.attributes = make(map[string]any)
		}
		for k, v := range attrs {
			c.attributes[k] = v
		}
	}
}
