package exporter

import (
	"context"

	"github.com/MitulShah1/openai-agents-go/tracing/schema"
)

// noopExporter is a no-op exporter that discards all data.
type noopExporter struct{}

// NewNoopExporter returns an exporter that discards all data.
func NewNoopExporter() Exporter {
	return &noopExporter{}
}

func (e *noopExporter) Export(_ context.Context, _ string, _ []schema.TraceExport, _ []schema.SpanExport) error {
	return nil
}

func (e *noopExporter) Close() error {
	return nil
}
