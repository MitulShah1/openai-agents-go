package exporter

import (
	"context"
	"testing"

	"github.com/MitulShah1/openai-agents-go/tracing/schema"
)

func TestNewConsole(t *testing.T) {
	exporter := NewConsoleExporter()

	if exporter == nil {
		t.Fatal("NewConsoleExporter returned nil")
	}
}

func TestConsoleExport(t *testing.T) {
	exporter := NewConsoleExporter()
	ctx := context.Background()

	traces := []schema.TraceExport{
		{
			Object:       "trace",
			ID:           "trace_123",
			WorkflowName: "test-workflow",
		},
	}

	spans := []schema.SpanExport{
		{
			Object:    "trace.span",
			ID:        "span_123",
			TraceID:   "trace_123",
			StartedAt: "2024-01-01T00:00:00Z",
			EndedAt:   "2024-01-01T00:00:01Z",
		},
	}

	// Should not error
	err := exporter.Export(ctx, "test-key", traces, spans)
	if err != nil {
		t.Errorf("Console export failed: %v", err)
	}
}

func TestConsoleExportEmpty(t *testing.T) {
	exporter := NewConsoleExporter()
	ctx := context.Background()

	// Should handle empty slices
	err := exporter.Export(ctx, "", []schema.TraceExport{}, []schema.SpanExport{})
	if err != nil {
		t.Errorf("Console export with empty data failed: %v", err)
	}
}

func TestConsoleClose(t *testing.T) {
	exporter := NewConsoleExporter()

	err := exporter.Close()
	if err != nil {
		t.Errorf("Console close failed: %v", err)
	}
}
