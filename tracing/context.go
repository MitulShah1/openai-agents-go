package tracing

import "context"

type contextKey struct{}
type spanContextKey struct{}

var (
	traceKey = contextKey{}
	spanKey  = spanContextKey{}
)

// FromContext returns the current trace from the context.
// Returns nil if no trace is present.
func FromContext(ctx context.Context) Trace {
	if trace, ok := ctx.Value(traceKey).(Trace); ok {
		return trace
	}
	return nil
}

// ContextWithTrace returns a new context with the given trace.
func ContextWithTrace(ctx context.Context, trace Trace) context.Context {
	return context.WithValue(ctx, traceKey, trace)
}

// SpanFromContext returns the current span from the context.
// Returns nil if no span is present.
func SpanFromContext(ctx context.Context) Span {
	if span, ok := ctx.Value(spanKey).(Span); ok {
		return span
	}
	return nil
}

// ContextWithSpan returns a new context with the given span.
func ContextWithSpan(ctx context.Context, span Span) context.Context {
	return context.WithValue(ctx, spanKey, span)
}
