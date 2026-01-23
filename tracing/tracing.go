// Package tracing implements OpenAI Traces (traces=v1) compatible tracing for this SDK.
//
// This package provides a Go-idiomatic API for creating and managing traces and spans
// that are compatible with the OpenAI Traces dashboard.
//
// # Basic Usage
//
//	// Set up a trace provider
//	provider := tracing.NewProvider(
//	    processor.NewBatch(exporter.NewBackend()),
//	)
//	tracing.SetProvider(provider)
//	defer provider.Shutdown(context.Background())
//
//	// Start a trace
//	ctx, trace, err := tracing.GetProvider().StartTrace(ctx,
//	    tracing.WithWorkflowName("my-workflow"),
//	)
//	if err != nil {
//	    log.Warn("failed to start trace", "error", err)
//	}
//	defer trace.End(ctx)
//
//	// Start a span
//	ctx, span, err := trace.StartSpan(ctx, schema.SpanTypeAgent,
//	    tracing.WithName("my-agent"),
//	    tracing.WithModel("gpt-4"),
//	)
//	if err != nil {
//	    log.Warn("failed to start span", "error", err)
//	}
//	defer span.End(ctx)
//
// # Convenience Functions
//
// For common span types, use the convenience functions:
//
//	ctx, span, err := tracing.StartAgentSpan(ctx, "my-agent",
//	    tracing.WithModel("gpt-4"),
//	)
//	defer span.End(ctx)
//
// # Error Handling
//
// Tracing errors never break core SDK functionality. If trace or span creation fails,
// a no-op implementation is returned along with an error. The SDK continues working normally.
package tracing

import (
	"context"

	"github.com/MitulShah1/openai-agents-go/tracing/schema"
)

// Provider creates and manages traces.
//
// Implementations must be safe for concurrent use.
type Provider interface {
	// StartTrace creates a new trace with the given options.
	// Returns a context containing the trace, the trace itself, and any error.
	//
	// If an error occurs, a no-op trace is returned so the SDK continues working.
	StartTrace(ctx context.Context, opts ...TraceOption) (context.Context, Trace, error)

	// Shutdown gracefully shuts down the provider, flushing any pending data.
	Shutdown(ctx context.Context) error
}

// Trace represents a unit of work compatible with the OpenAI Traces dashboard.
//
// All methods are safe for concurrent use.
type Trace interface {
	// ID returns the unique trace identifier.
	ID() string

	// WorkflowName returns the workflow name for this trace.
	WorkflowName() string

	// StartSpan creates a new child span with the given type and options.
	// Returns a context containing the span, the span itself, and any error.
	//
	// If an error occurs, a no-op span is returned so the SDK continues working.
	StartSpan(ctx context.Context, spanType schema.SpanType, opts ...SpanOption) (context.Context, Span, error)

	// End marks the trace as complete.
	// This should be called with defer: defer trace.End(ctx)
	End(ctx context.Context)
}

// Span represents a trace span compatible with the OpenAI Traces dashboard.
//
// All methods are safe for concurrent use.
type Span interface {
	// ID returns the unique span identifier.
	ID() string

	// TraceID returns the parent trace identifier.
	TraceID() string

	// Type returns the span type.
	Type() schema.SpanType

	// RecordError records an error on this span.
	RecordError(err error)

	// SetAttributes sets span-specific attributes.
	// This can be called multiple times to add or update attributes.
	SetAttributes(attrs map[string]any)

	// End marks the span as complete.
	// This should be called with defer: defer span.End(ctx)
	End(ctx context.Context)
}
