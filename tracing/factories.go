package tracing

import (
	"context"

	"github.com/MitulShah1/openai-agents-go/tracing/schema"
)

// AgentSpan creates a new agent span with type-safe options.
// This is a convenience function that automatically sets the span type to SpanTypeAgent.
//
// Example:
//
//	ctx, span, err := tracing.AgentSpan(ctx, "customer-support",
//	    tracing.WithModel("gpt-4"),
//	    tracing.WithInstructions("You are a helpful assistant"),
//	)
//	defer span.End(ctx)
func AgentSpan(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span, error) {
	trace := FromContext(ctx)
	if trace == nil {
		return ctx, &noopSpan{}, ErrNoActiveTrace
	}

	allOpts := append([]SpanOption{WithSpanName(name)}, opts...)
	return trace.StartSpan(ctx, schema.SpanTypeAgent, allOpts...)
}

// GenerationSpan creates a new generation span for LLM API calls.
// This is a convenience function that automatically sets the span type to SpanTypeGeneration.
//
// Example:
//
//	ctx, span, err := tracing.GenerationSpan(ctx,
//	    tracing.WithModel("gpt-4"),
//	    tracing.WithRequest(messages),
//	)
//	defer span.End(ctx)
func GenerationSpan(ctx context.Context, opts ...SpanOption) (context.Context, Span, error) {
	trace := FromContext(ctx)
	if trace == nil {
		return ctx, &noopSpan{}, ErrNoActiveTrace
	}

	return trace.StartSpan(ctx, schema.SpanTypeGeneration, opts...)
}

// FunctionSpan creates a new function/tool span.
// This is a convenience function that automatically sets the span type to SpanTypeFunction.
//
// Example:
//
//	ctx, span, err := tracing.FunctionSpan(ctx, "get_weather",
//	    tracing.WithArguments(args),
//	)
//	defer span.End(ctx)
//	span.SetAttributes(map[string]any{"output": result})
func FunctionSpan(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span, error) {
	trace := FromContext(ctx)
	if trace == nil {
		return ctx, &noopSpan{}, ErrNoActiveTrace
	}

	allOpts := append([]SpanOption{WithSpanName(name)}, opts...)
	return trace.StartSpan(ctx, schema.SpanTypeFunction, allOpts...)
}

// HandoffSpan creates a new handoff span for agent-to-agent transitions.
// This is a convenience function that automatically sets the span type to SpanTypeHandoff.
//
// Example:
//
//	ctx, span, err := tracing.HandoffSpan(ctx,
//	    tracing.WithFromAgent("triage"),
//	    tracing.WithToAgent("specialist"),
//	)
//	defer span.End(ctx)
func HandoffSpan(ctx context.Context, opts ...SpanOption) (context.Context, Span, error) {
	trace := FromContext(ctx)
	if trace == nil {
		return ctx, &noopSpan{}, ErrNoActiveTrace
	}

	return trace.StartSpan(ctx, schema.SpanTypeHandoff, opts...)
}

// GuardrailSpan creates a new guardrail span for safety checks.
// This is a convenience function that automatically sets the span type to SpanTypeGuardrail.
//
// Example:
//
//	ctx, span, err := tracing.GuardrailSpan(ctx, "content-filter",
//	    tracing.WithGuardrailType("content_moderation"),
//	)
//	defer span.End(ctx)
func GuardrailSpan(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span, error) {
	trace := FromContext(ctx)
	if trace == nil {
		return ctx, &noopSpan{}, ErrNoActiveTrace
	}

	allOpts := append([]SpanOption{WithSpanName(name)}, opts...)
	return trace.StartSpan(ctx, schema.SpanTypeGuardrail, allOpts...)
}

// CustomSpan creates a new custom span for user-defined operations.
// This is a convenience function that automatically sets the span type to SpanTypeCustom.
//
// Example:
//
//	ctx, span, err := tracing.CustomSpan(ctx, "database-query",
//	    tracing.WithAttributes(map[string]any{
//	        "query": "SELECT * FROM users",
//	        "duration_ms": 42,
//	    }),
//	)
//	defer span.End(ctx)
func CustomSpan(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span, error) {
	trace := FromContext(ctx)
	if trace == nil {
		return ctx, &noopSpan{}, ErrNoActiveTrace
	}

	allOpts := append([]SpanOption{WithSpanName(name)}, opts...)
	return trace.StartSpan(ctx, schema.SpanTypeCustom, allOpts...)
}
