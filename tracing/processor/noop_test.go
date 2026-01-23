package processor

import (
	"context"
	"testing"

	"github.com/MitulShah1/openai-agents-go/tracing/schema"
)

func TestNewNoopProcessor(t *testing.T) {
	processor := NewNoopProcessor()

	if processor == nil {
		t.Fatal("NewNoopProcessor returned nil")
	}
}

func TestNoopProcessorOnTraceStart(_ *testing.T) {
	processor := NewNoopProcessor()

	trace := schema.TraceExport{
		ID: "trace_123",
	}

	// Should not panic
	processor.OnTraceStart(trace, "test-key")
}

func TestNoopProcessorOnTraceEnd(_ *testing.T) {
	processor := NewNoopProcessor()

	trace := schema.TraceExport{
		ID: "trace_123",
	}

	// Should not panic
	processor.OnTraceEnd(trace, "test-key")
}

func TestNoopProcessorOnSpanEnd(_ *testing.T) {
	processor := NewNoopProcessor()

	span := schema.SpanExport{
		ID: "span_123",
	}

	// Should not panic
	processor.OnSpanEnd(span, "test-key")
}

func TestNoopProcessorShutdown(t *testing.T) {
	processor := NewNoopProcessor()

	ctx := context.Background()
	err := processor.Shutdown(ctx)

	if err != nil {
		t.Errorf("Noop shutdown failed: %v", err)
	}
}
