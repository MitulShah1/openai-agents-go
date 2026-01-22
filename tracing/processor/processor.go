// Package processor provides trace and span processors.
//
// Processors receive trace and span events and can batch, filter, or export them.
package processor

import (
	"context"

	"github.com/MitulShah1/openai-agents-go/tracing/schema"
)

// Processor processes traces and spans.
//
// Implementations must be safe for concurrent use.
type Processor interface {
	// OnTraceStart is called when a trace starts.
	OnTraceStart(trace schema.TraceExport, apiKey string)

	// OnSpanEnd is called when a span ends.
	OnSpanEnd(span schema.SpanExport, apiKey string)

	// OnTraceEnd is called when a trace ends.
	OnTraceEnd(trace schema.TraceExport, apiKey string)

	// Shutdown gracefully shuts down the processor, flushing any pending data.
	Shutdown(ctx context.Context) error
}
