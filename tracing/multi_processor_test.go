package tracing

import (
	"context"
	"testing"

	"github.com/MitulShah1/openai-agents-go/tracing/processor"
	"github.com/MitulShah1/openai-agents-go/tracing/schema"
)

// Panicking processor for testing panic recovery
type panicProcessor struct {
	shouldPanic bool
}

func (p *panicProcessor) OnTraceStart(_ schema.TraceExport, _ string) {
	if p.shouldPanic {
		panic("test panic in OnTraceStart")
	}
}

func (p *panicProcessor) OnTraceEnd(_ schema.TraceExport, _ string) {
	if p.shouldPanic {
		panic("test panic in OnTraceEnd")
	}
}

func (p *panicProcessor) OnSpanEnd(_ schema.SpanExport, _ string) {
	if p.shouldPanic {
		panic("test panic in OnSpanEnd")
	}
}

func (p *panicProcessor) Shutdown(_ context.Context) error {
	return nil
}

func TestMultiProcessorOnTraceStart(t *testing.T) {
	mock1 := &mockProcessor{}
	mock2 := &mockProcessor{}

	multi := newMultiProcessor(mock1, mock2)

	export := schema.TraceExport{
		ID:           "trace_123",
		WorkflowName: "test",
	}

	multi.OnTraceStart(export, "test-key")

	// Both processors should receive the event
	if len(mock1.traceStarts) != 1 {
		t.Errorf("mock1: expected 1 trace start, got %d", len(mock1.traceStarts))
	}
	if len(mock2.traceStarts) != 1 {
		t.Errorf("mock2: expected 1 trace start, got %d", len(mock2.traceStarts))
	}
}

func TestMultiProcessorOnTraceEnd(t *testing.T) {
	mock1 := &mockProcessor{}
	mock2 := &mockProcessor{}

	multi := newMultiProcessor(mock1, mock2)

	export := schema.TraceExport{
		ID:           "trace_123",
		WorkflowName: "test",
	}

	multi.OnTraceEnd(export, "test-key")

	// Both processors should receive the event
	if len(mock1.traceEnds) != 1 {
		t.Errorf("mock1: expected 1 trace end, got %d", len(mock1.traceEnds))
	}
	if len(mock2.traceEnds) != 1 {
		t.Errorf("mock2: expected 1 trace end, got %d", len(mock2.traceEnds))
	}
}

func TestMultiProcessorOnSpanEnd(t *testing.T) {
	mock1 := &mockProcessor{}
	mock2 := &mockProcessor{}

	multi := newMultiProcessor(mock1, mock2)

	export := schema.SpanExport{
		ID:      "span_123",
		TraceID: "trace_123",
	}

	multi.OnSpanEnd(export, "test-key")

	// Both processors should receive the event
	if len(mock1.spanEnds) != 1 {
		t.Errorf("mock1: expected 1 span end, got %d", len(mock1.spanEnds))
	}
	if len(mock2.spanEnds) != 1 {
		t.Errorf("mock2: expected 1 span end, got %d", len(mock2.spanEnds))
	}
}

func TestMultiProcessorAddProcessor(t *testing.T) {
	mock1 := &mockProcessor{}
	mock2 := &mockProcessor{}

	multi := newMultiProcessor(mock1)
	multi.AddProcessor(mock2)

	export := schema.TraceExport{
		ID:           "trace_123",
		WorkflowName: "test",
	}

	multi.OnTraceStart(export, "test-key")

	// Both processors should receive the event
	if len(mock1.traceStarts) != 1 {
		t.Errorf("mock1: expected 1 trace start, got %d", len(mock1.traceStarts))
	}
	if len(mock2.traceStarts) != 1 {
		t.Errorf("mock2: expected 1 trace start, got %d", len(mock2.traceStarts))
	}
}

func TestMultiProcessorSetProcessors(t *testing.T) {
	mock1 := &mockProcessor{}
	mock2 := &mockProcessor{}
	mock3 := &mockProcessor{}

	multi := newMultiProcessor(mock1)
	multi.SetProcessors([]processor.Processor{mock2, mock3})

	export := schema.TraceExport{
		ID:           "trace_123",
		WorkflowName: "test",
	}

	multi.OnTraceStart(export, "test-key")

	// Only mock2 and mock3 should receive the event
	if len(mock1.traceStarts) != 0 {
		t.Errorf("mock1: expected 0 trace starts, got %d", len(mock1.traceStarts))
	}
	if len(mock2.traceStarts) != 1 {
		t.Errorf("mock2: expected 1 trace start, got %d", len(mock2.traceStarts))
	}
	if len(mock3.traceStarts) != 1 {
		t.Errorf("mock3: expected 1 trace start, got %d", len(mock3.traceStarts))
	}
}

func TestMultiProcessorShutdown(t *testing.T) {
	mock1 := &mockProcessor{}
	mock2 := &mockProcessor{}

	multi := newMultiProcessor(mock1, mock2)

	ctx := context.Background()
	err := multi.Shutdown(ctx)

	if err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}

	// Both processors should be shut down
	if mock1.shutdowns != 1 {
		t.Errorf("mock1: expected 1 shutdown, got %d", mock1.shutdowns)
	}
	if mock2.shutdowns != 1 {
		t.Errorf("mock2: expected 1 shutdown, got %d", mock2.shutdowns)
	}
}

func TestMultiProcessorShutdownWithError(t *testing.T) {
	mock1 := &mockProcessor{}
	multi := newMultiProcessor(mock1)

	ctx := context.Background()
	err := multi.Shutdown(ctx)

	// Should not return error (continues shutting down other processors)
	if err != nil {
		t.Logf("Shutdown returned error (expected): %v", err)
	}
}

func TestMultiProcessorPanicRecovery(t *testing.T) {
	mock := &mockProcessor{}
	panicProc := &panicProcessor{shouldPanic: true}

	multi := newMultiProcessor(panicProc, mock)

	export := schema.TraceExport{
		ID:           "trace_123",
		WorkflowName: "test",
	}

	// Should not panic - panic should be recovered
	multi.OnTraceStart(export, "test-key")

	// Mock processor should still receive the event
	if len(mock.traceStarts) != 1 {
		t.Errorf("Expected 1 trace start in mock, got %d", len(mock.traceStarts))
	}
}

func TestMultiProcessorConcurrent(t *testing.T) {
	mock := &mockProcessor{}
	multi := newMultiProcessor(mock)

	const numGoroutines = 100
	done := make(chan bool, numGoroutines)

	export := schema.TraceExport{
		ID:           "trace_123",
		WorkflowName: "test",
	}

	for i := 0; i < numGoroutines; i++ {
		go func() {
			multi.OnTraceStart(export, "test-key")
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	if len(mock.traceStarts) != numGoroutines {
		t.Errorf("Expected %d trace starts, got %d", numGoroutines, len(mock.traceStarts))
	}
}
