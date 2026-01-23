package exporter

import (
	"context"
	"testing"

	"github.com/MitulShah1/openai-agents-go/tracing/schema"
)

func TestNewNoopExporter(t *testing.T) {
	exporter := NewNoopExporter()

	if exporter == nil {
		t.Fatal("NewNoopExporter returned nil")
	}
}

func TestNoopExporterExport(t *testing.T) {
	exporter := NewNoopExporter()
	ctx := context.Background()

	traces := []schema.TraceExport{
		{ID: "trace_123"},
	}
	spans := []schema.SpanExport{
		{ID: "span_123"},
	}

	// Should not error
	err := exporter.Export(ctx, "test-key", traces, spans)
	if err != nil {
		t.Errorf("Noop export failed: %v", err)
	}
}

func TestNoopExporterClose(t *testing.T) {
	exporter := NewNoopExporter()

	err := exporter.Close()
	if err != nil {
		t.Errorf("Noop close failed: %v", err)
	}
}
