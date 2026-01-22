// Package exporter provides trace exporters.
//
// Exporters send traces and spans to external backends.
package exporter

import (
	"context"

	"github.com/MitulShah1/openai-agents-go/tracing/schema"
)

// Exporter exports traces and spans to a backend.
//
// Implementations must be safe for concurrent use.
type Exporter interface {
	// Export exports the given traces and spans.
	// The apiKey is used for authentication with the backend.
	Export(ctx context.Context, apiKey string, traces []schema.TraceExport, spans []schema.SpanExport) error

	// Close closes the exporter and releases resources.
	Close() error
}
