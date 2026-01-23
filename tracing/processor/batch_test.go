package processor

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/MitulShah1/openai-agents-go/tracing/schema"
)

// Mock exporter for testing
type mockExporter struct {
	mu      sync.Mutex
	exports int
	closes  int
	err     error
}

func (m *mockExporter) Export(_ context.Context, _ string, _ []schema.TraceExport, _ []schema.SpanExport) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.exports++
	return m.err
}

func (m *mockExporter) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closes++
	return nil
}

func (m *mockExporter) getExports() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.exports
}

func TestNewBatch(t *testing.T) {
	exporter := &mockExporter{}
	processor := NewBatch(exporter)

	if processor == nil {
		t.Fatal("NewBatch returned nil")
	}
}

func TestNewBatchWithOptions(t *testing.T) {
	exporter := &mockExporter{}
	processor := NewBatch(exporter,
		WithMaxBatchSize(50),
		WithFlushInterval(1*time.Second),
		WithMaxQueueSize(1000),
	)

	if processor == nil {
		t.Fatal("NewBatch with options returned nil")
	}
}

func TestBatchProcessorOnTraceStart(t *testing.T) {
	exporter := &mockExporter{}
	processor := NewBatch(exporter,
		WithMaxBatchSize(10),
		WithFlushInterval(100*time.Millisecond),
	)
	defer func() { _ = processor.Shutdown(context.Background()) }()

	trace := schema.TraceExport{
		Object:       "trace",
		ID:           "trace_123",
		WorkflowName: "test",
	}

	processor.OnTraceStart(trace, "test-key")

	// Wait for batch to be processed
	time.Sleep(200 * time.Millisecond)

	if exporter.getExports() == 0 {
		t.Error("Expected at least one export")
	}
}

func TestBatchProcessorOnSpanEnd(t *testing.T) {
	exporter := &mockExporter{}
	processor := NewBatch(exporter,
		WithMaxBatchSize(10),
		WithFlushInterval(100*time.Millisecond),
	)
	defer func() { _ = processor.Shutdown(context.Background()) }()

	span := schema.SpanExport{
		Object:  "trace.span",
		ID:      "span_123",
		TraceID: "trace_123",
	}

	processor.OnSpanEnd(span, "test-key")

	// Wait for batch to be processed
	time.Sleep(200 * time.Millisecond)

	if exporter.getExports() == 0 {
		t.Error("Expected at least one export")
	}
}

func TestBatchProcessorBatchSizeTrigger(t *testing.T) {
	exporter := &mockExporter{}
	batchSize := 5
	processor := NewBatch(exporter,
		WithMaxBatchSize(batchSize),
		WithFlushInterval(10*time.Second), // Long timeout so only size triggers
	)
	defer func() { _ = processor.Shutdown(context.Background()) }()

	// Send exactly batchSize events
	for i := 0; i < batchSize; i++ {
		span := schema.SpanExport{
			ID:      "span_" + string(rune(i)),
			TraceID: "trace_123",
		}
		processor.OnSpanEnd(span, "test-key")
	}

	// Give it a moment to process
	time.Sleep(100 * time.Millisecond)

	exports := exporter.getExports()
	if exports == 0 {
		t.Error("Expected batch to be exported when size threshold reached")
	}
}

func TestBatchProcessorShutdown(t *testing.T) {
	exporter := &mockExporter{}
	processor := NewBatch(exporter)

	// Add some events
	for i := 0; i < 3; i++ {
		span := schema.SpanExport{
			ID:      "span_" + string(rune(i)),
			TraceID: "trace_123",
		}
		processor.OnSpanEnd(span, "test-key")
	}

	ctx := context.Background()
	err := processor.Shutdown(ctx)

	if err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}

	// Should have flushed pending events
	if exporter.getExports() == 0 {
		t.Error("Expected pending events to be flushed on shutdown")
	}

	// Exporter should be closed
	if exporter.closes == 0 {
		t.Error("Expected exporter to be closed")
	}
}

func TestBatchProcessorShutdownIdempotent(t *testing.T) {
	exporter := &mockExporter{}
	processor := NewBatch(exporter)

	ctx := context.Background()

	// First shutdown
	err := processor.Shutdown(ctx)
	if err != nil {
		t.Errorf("First shutdown failed: %v", err)
	}

	// Second shutdown should not error
	err = processor.Shutdown(ctx)
	if err != nil {
		t.Errorf("Second shutdown failed: %v", err)
	}
}

func TestBatchProcessorConcurrent(t *testing.T) {
	exporter := &mockExporter{}
	processor := NewBatch(exporter,
		WithMaxBatchSize(100),
		WithFlushInterval(100*time.Millisecond),
	)
	defer func() { _ = processor.Shutdown(context.Background()) }()

	const numGoroutines = 50
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			span := schema.SpanExport{
				ID:      "span_" + string(rune(id)),
				TraceID: "trace_123",
			}
			processor.OnSpanEnd(span, "test-key")
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Wait for processing
	time.Sleep(200 * time.Millisecond)

	if exporter.getExports() == 0 {
		t.Error("Expected exports from concurrent operations")
	}
}

func TestBatchProcessorOnTraceEnd(t *testing.T) {
	exporter := &mockExporter{}
	processor := NewBatch(exporter,
		WithFlushInterval(100*time.Millisecond),
	)
	defer func() { _ = processor.Shutdown(context.Background()) }()

	trace := schema.TraceExport{
		Object: "trace",
		ID:     "trace_123",
	}

	processor.OnTraceEnd(trace, "test-key")

	// Wait for batch
	time.Sleep(200 * time.Millisecond)

	if exporter.getExports() == 0 {
		t.Error("Expected export for trace end")
	}
}
