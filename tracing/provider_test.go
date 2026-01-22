package tracing

import (
	"context"
	"sync"
	"testing"

	"github.com/MitulShah1/openai-agents-go/tracing/processor"
	"github.com/MitulShah1/openai-agents-go/tracing/schema"
)

// Mock processor for testing
type mockProcessor struct {
	mu          sync.Mutex
	traceStarts []schema.TraceExport
	traceEnds   []schema.TraceExport
	spanEnds    []schema.SpanExport
	shutdowns   int
}

func (m *mockProcessor) OnTraceStart(trace schema.TraceExport, _ string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.traceStarts = append(m.traceStarts, trace)
}

func (m *mockProcessor) OnTraceEnd(trace schema.TraceExport, _ string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.traceEnds = append(m.traceEnds, trace)
}

func (m *mockProcessor) OnSpanEnd(span schema.SpanExport, _ string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.spanEnds = append(m.spanEnds, span)
}

func (m *mockProcessor) Shutdown(_ context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shutdowns++
	return nil
}

func TestNewProvider(t *testing.T) {
	mock := &mockProcessor{}
	provider := NewProvider(mock)

	if provider == nil {
		t.Fatal("NewProvider returned nil")
	}

	// Verify it's a defaultProvider
	if _, ok := provider.(*defaultProvider); !ok {
		t.Errorf("Expected *defaultProvider, got %T", provider)
	}
}

func TestProviderStartTrace(t *testing.T) {
	mock := &mockProcessor{}
	provider := NewProvider(mock)
	ctx := context.Background()

	ctx, trace, err := provider.StartTrace(ctx,
		WithWorkflowName("test-workflow"),
	)

	if err != nil {
		t.Fatalf("StartTrace failed: %v", err)
	}

	if trace == nil {
		t.Fatal("StartTrace returned nil trace")
	}

	if trace.ID() == "" {
		t.Error("Trace ID is empty")
	}

	if trace.WorkflowName() != "test-workflow" {
		t.Errorf("Expected workflow name 'test-workflow', got %q", trace.WorkflowName())
	}

	// Verify processor was called
	if len(mock.traceStarts) != 1 {
		t.Errorf("Expected 1 trace start, got %d", len(mock.traceStarts))
	}

	// Verify trace is in context
	if FromContext(ctx) != trace {
		t.Error("Trace not found in context")
	}
}

func TestProviderStartTraceWithOptions(t *testing.T) {
	mock := &mockProcessor{}
	provider := NewProvider(mock)
	ctx := context.Background()

	metadata := map[string]any{"key": "value"}
	_, trace, err := provider.StartTrace(ctx,
		WithWorkflowName("test"),
		WithGroupID("group_123456789012345678901234"),
		WithMetadata(metadata),
		WithSensitiveData(false),
	)

	if err != nil {
		t.Fatalf("StartTrace failed: %v", err)
	}

	if trace == nil {
		t.Fatal("StartTrace returned nil trace")
	}
}

func TestProviderStartTraceWithInvalidTraceID(t *testing.T) {
	mock := &mockProcessor{}
	provider := NewProvider(mock)
	ctx := context.Background()

	_, trace, err := provider.StartTrace(ctx,
		WithTraceID("invalid-id"),
	)

	// Should return error for invalid trace ID
	if err == nil {
		t.Error("Expected error for invalid trace ID, got nil")
	}

	// Should return no-op trace
	if _, ok := trace.(*noopTrace); !ok {
		t.Errorf("Expected noopTrace, got %T", trace)
	}
}

func TestProviderStartTraceWithInvalidGroupID(t *testing.T) {
	mock := &mockProcessor{}
	provider := NewProvider(mock)
	ctx := context.Background()

	_, trace, err := provider.StartTrace(ctx,
		WithGroupID("invalid-group"),
	)

	if err == nil {
		t.Error("Expected error for invalid group ID")
	}

	// Should return no-op trace
	if _, ok := trace.(*noopTrace); !ok {
		t.Errorf("Expected noopTrace, got %T", trace)
	}
}

func TestProviderShutdown(t *testing.T) {
	mock := &mockProcessor{}
	provider := NewProvider(mock)
	ctx := context.Background()

	err := provider.Shutdown(ctx)
	if err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}

	if mock.shutdowns != 1 {
		t.Errorf("Expected 1 shutdown call, got %d", mock.shutdowns)
	}
}

func TestProviderAddProcessor(t *testing.T) {
	mock1 := &mockProcessor{}
	mock2 := &mockProcessor{}

	provider := NewProvider(mock1).(*defaultProvider)
	provider.AddProcessor(mock2)

	ctx := context.Background()
	ctx, trace, _ := provider.StartTrace(ctx, WithWorkflowName("test"))
	trace.End(ctx)

	// Both processors should receive events
	if len(mock1.traceStarts) != 1 {
		t.Errorf("mock1: expected 1 trace start, got %d", len(mock1.traceStarts))
	}
	if len(mock2.traceStarts) != 1 {
		t.Errorf("mock2: expected 1 trace start, got %d", len(mock2.traceStarts))
	}
}

func TestProviderSetProcessors(t *testing.T) {
	mock1 := &mockProcessor{}
	mock2 := &mockProcessor{}
	mock3 := &mockProcessor{}

	provider := NewProvider(mock1).(*defaultProvider)
	provider.SetProcessors([]processor.Processor{mock2, mock3})

	ctx := context.Background()
	ctx, trace, _ := provider.StartTrace(ctx, WithWorkflowName("test"))
	trace.End(ctx)

	// Only mock2 and mock3 should receive events
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

func TestProviderTracePooling(t *testing.T) {
	mock := &mockProcessor{}
	provider := NewProvider(mock)
	ctx := context.Background()

	// Create and end multiple traces
	for i := 0; i < 10; i++ {
		ctx, trace, _ := provider.StartTrace(ctx, WithWorkflowName("test"))
		trace.End(ctx)
	}

	// Verify all traces were created and ended
	if len(mock.traceStarts) != 10 {
		t.Errorf("Expected 10 trace starts, got %d", len(mock.traceStarts))
	}
	if len(mock.traceEnds) != 10 {
		t.Errorf("Expected 10 trace ends, got %d", len(mock.traceEnds))
	}
}

func TestProviderConcurrentTraceCreation(t *testing.T) {
	mock := &mockProcessor{}
	provider := NewProvider(mock)

	const numGoroutines = 100
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			ctx := context.Background()
			ctx, trace, err := provider.StartTrace(ctx, WithWorkflowName("concurrent"))
			if err != nil {
				t.Errorf("StartTrace failed: %v", err)
			}
			trace.End(ctx)
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
