package tracing

import "errors"

var (
	// ErrNoActiveTrace is returned when attempting to create a span without an active trace.
	ErrNoActiveTrace = errors.New("no active trace in context")
)
