package tracing

import "context"

type contextKey struct{}

var traceKey = contextKey{}

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
