package tracing

import (
	"context"

	"github.com/MitulShah1/openai-agents-go/tracing/schema"
)

// StartAgentSpan creates an agent span under the current trace.
func StartAgentSpan(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span, error) {
	t := FromContext(ctx)
	if t == nil {
		return (&noopTrace{}).StartSpan(ctx, schema.SpanTypeAgent)
	}

	defaultOpts := []SpanOption{WithName(name)}
	return t.StartSpan(ctx, schema.SpanTypeAgent, append(defaultOpts, opts...)...)
}

// StartGenerationSpan creates a generation span under the current trace.
func StartGenerationSpan(ctx context.Context, opts ...SpanOption) (context.Context, Span, error) {
	t := FromContext(ctx)
	if t == nil {
		return (&noopTrace{}).StartSpan(ctx, schema.SpanTypeGeneration)
	}
	return t.StartSpan(ctx, schema.SpanTypeGeneration, opts...)
}

// StartFunctionSpan creates a function/tool span under the current trace.
func StartFunctionSpan(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span, error) {
	t := FromContext(ctx)
	if t == nil {
		return (&noopTrace{}).StartSpan(ctx, schema.SpanTypeFunction)
	}
	defaultOpts := []SpanOption{WithName(name)}
	return t.StartSpan(ctx, schema.SpanTypeFunction, append(defaultOpts, opts...)...)
}

// StartGuardrailSpan creates a guardrail span under the current trace.
func StartGuardrailSpan(ctx context.Context, name string, guardrailType string, opts ...SpanOption) (context.Context, Span, error) {
	t := FromContext(ctx)
	if t == nil {
		return (&noopTrace{}).StartSpan(ctx, schema.SpanTypeGuardrail)
	}
	defaultOpts := []SpanOption{WithName(name), WithGuardrailType(guardrailType)}
	return t.StartSpan(ctx, schema.SpanTypeGuardrail, append(defaultOpts, opts...)...)
}

// StartHandoffSpan creates a handoff span under the current trace.
func StartHandoffSpan(ctx context.Context, fromAgent, toAgent, reason string, opts ...SpanOption) (context.Context, Span, error) {
	t := FromContext(ctx)
	if t == nil {
		return (&noopTrace{}).StartSpan(ctx, schema.SpanTypeHandoff)
	}
	defaultOpts := []SpanOption{WithHandoff(fromAgent, toAgent, reason)}
	return t.StartSpan(ctx, schema.SpanTypeHandoff, append(defaultOpts, opts...)...)
}
