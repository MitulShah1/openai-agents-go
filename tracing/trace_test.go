package tracing

import (
	"context"
	"testing"

	"github.com/MitulShah1/openai-agents-go/tracing/schema"
)

func TestTraceStartSpan(t *testing.T) {
	mock := &mockProcessor{}
	provider := NewProvider(mock)
	ctx := context.Background()

	ctx, trace, _ := provider.StartTrace(ctx, WithWorkflowName("test"))
	defer trace.End(ctx)

	ctx, span, err := trace.StartSpan(ctx, schema.SpanTypeAgent,
		WithSpanName("test-span"),
	)

	if err != nil {
		t.Fatalf("StartSpan failed: %v", err)
	}

	if span == nil {
		t.Fatal("StartSpan returned nil span")
	}

	if span.ID() == "" {
		t.Error("Span ID is empty")
	}

	if span.TraceID() != trace.ID() {
		t.Errorf("Span trace ID %q doesn't match trace ID %q", span.TraceID(), trace.ID())
	}

	if span.Type() != schema.SpanTypeAgent {
		t.Errorf("Expected span type %q, got %q", schema.SpanTypeAgent, span.Type())
	}

	span.End(ctx)

	// Verify processor received span end
	if len(mock.spanEnds) != 1 {
		t.Errorf("Expected 1 span end, got %d", len(mock.spanEnds))
	}
}

func TestTraceStartSpanWithOptions(t *testing.T) {
	mock := &mockProcessor{}
	provider := NewProvider(mock)
	ctx := context.Background()

	ctx, trace, _ := provider.StartTrace(ctx, WithWorkflowName("test"))
	defer trace.End(ctx)

	attrs := map[string]any{
		"model":        "gpt-4",
		"temperature":  0.7,
		"instructions": "test",
	}

	ctx, span, err := trace.StartSpan(ctx, schema.SpanTypeAgent,
		WithSpanName("agent"),
		WithAttributes(attrs),
	)

	if err != nil {
		t.Fatalf("StartSpan failed: %v", err)
	}

	span.End(ctx)
}

func TestTraceEnd(t *testing.T) {
	mock := &mockProcessor{}
	provider := NewProvider(mock)
	ctx := context.Background()

	ctx, trace, _ := provider.StartTrace(ctx, WithWorkflowName("test"))
	trace.End(ctx)

	// Verify processor received trace end
	if len(mock.traceEnds) != 1 {
		t.Errorf("Expected 1 trace end, got %d", len(mock.traceEnds))
	}

	// Verify trace data
	if mock.traceEnds[0].WorkflowName != "test" {
		t.Errorf("Expected workflow name 'test', got %q", mock.traceEnds[0].WorkflowName)
	}
}

func TestTraceSpanPooling(t *testing.T) {
	mock := &mockProcessor{}
	provider := NewProvider(mock)
	ctx := context.Background()

	ctx, trace, _ := provider.StartTrace(ctx, WithWorkflowName("test"))
	defer trace.End(ctx)

	// Create and end multiple spans
	for i := 0; i < 10; i++ {
		ctx, span, _ := trace.StartSpan(ctx, schema.SpanTypeCustom,
			WithSpanName("span"),
		)
		span.End(ctx)
	}

	// Verify all spans were ended
	if len(mock.spanEnds) != 10 {
		t.Errorf("Expected 10 span ends, got %d", len(mock.spanEnds))
	}
}

func TestTraceNestedSpans(t *testing.T) {
	mock := &mockProcessor{}
	provider := NewProvider(mock)
	ctx := context.Background()

	ctx, trace, _ := provider.StartTrace(ctx, WithWorkflowName("test"))
	defer trace.End(ctx)

	// Create parent span
	ctx, parent, _ := trace.StartSpan(ctx, schema.SpanTypeAgent,
		WithSpanName("parent"),
	)

	// Create child span (should auto-detect parent from context)
	ctx, child, _ := trace.StartSpan(ctx, schema.SpanTypeFunction,
		WithSpanName("child"),
	)

	child.End(ctx)
	parent.End(ctx)

	// Verify parent-child relationship
	if len(mock.spanEnds) != 2 {
		t.Fatalf("Expected 2 span ends, got %d", len(mock.spanEnds))
	}

	childSpan := mock.spanEnds[0]
	parentSpan := mock.spanEnds[1]

	if childSpan.ParentID != parentSpan.ID {
		t.Errorf("Child parent ID %q doesn't match parent ID %q", childSpan.ParentID, parentSpan.ID)
	}
}

func TestTraceConcurrentSpans(t *testing.T) {
	mock := &mockProcessor{}
	provider := NewProvider(mock)
	ctx := context.Background()

	ctx, trace, _ := provider.StartTrace(ctx, WithWorkflowName("test"))
	defer trace.End(ctx)

	const numGoroutines = 50
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			localCtx, span, err := trace.StartSpan(ctx, schema.SpanTypeCustom,
				WithSpanName("concurrent"),
			)
			if err != nil {
				t.Errorf("StartSpan failed: %v", err)
			}
			span.End(localCtx)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	if len(mock.spanEnds) != numGoroutines {
		t.Errorf("Expected %d span ends, got %d", numGoroutines, len(mock.spanEnds))
	}
}

func TestTraceID(t *testing.T) {
	mock := &mockProcessor{}
	provider := NewProvider(mock)
	ctx := context.Background()

	ctx, trace, _ := provider.StartTrace(ctx, WithWorkflowName("test"))
	defer trace.End(ctx)

	id := trace.ID()
	if id == "" {
		t.Error("Trace ID is empty")
	}

	// ID should be consistent
	if trace.ID() != id {
		t.Error("Trace ID changed between calls")
	}
}

func TestTraceWorkflowName(t *testing.T) {
	mock := &mockProcessor{}
	provider := NewProvider(mock)
	ctx := context.Background()

	ctx, trace, _ := provider.StartTrace(ctx,
		WithWorkflowName("my-workflow"),
	)
	defer trace.End(ctx)

	if trace.WorkflowName() != "my-workflow" {
		t.Errorf("Expected workflow name 'my-workflow', got %q", trace.WorkflowName())
	}
}
