package tracing

import (
	"context"
	"testing"
)

func TestContextWithTrace(t *testing.T) {
	mock := &mockProcessor{}
	provider := NewProvider(mock)
	ctx := context.Background()

	ctx, trace, _ := provider.StartTrace(ctx, WithWorkflowName("test"))
	defer trace.End(ctx)

	// Trace should be in context
	retrieved := FromContext(ctx)
	if retrieved != trace {
		t.Error("Trace not found in context or different trace returned")
	}
}

func TestFromContext(t *testing.T) {
	ctx := context.Background()

	// No trace in context
	if FromContext(ctx) != nil {
		t.Error("Expected nil trace from empty context")
	}

	// Add trace to context
	mock := &mockProcessor{}
	provider := NewProvider(mock)
	ctx, trace, _ := provider.StartTrace(ctx, WithWorkflowName("test"))
	defer trace.End(ctx)

	// Should retrieve trace
	if FromContext(ctx) == nil {
		t.Error("Expected trace from context, got nil")
	}
}

func TestContextWithSpan(t *testing.T) {
	mock := &mockProcessor{}
	provider := NewProvider(mock)
	ctx := context.Background()

	ctx, trace, _ := provider.StartTrace(ctx, WithWorkflowName("test"))
	defer trace.End(ctx)

	ctx, span, _ := trace.StartSpan(ctx, "custom")
	defer span.End(ctx)

	// Span should be in context
	retrieved := SpanFromContext(ctx)
	if retrieved != span {
		t.Error("Span not found in context or different span returned")
	}
}

func TestSpanFromContext(t *testing.T) {
	ctx := context.Background()

	// No span in context
	if SpanFromContext(ctx) != nil {
		t.Error("Expected nil span from empty context")
	}

	// Add span to context
	mock := &mockProcessor{}
	provider := NewProvider(mock)
	ctx, trace, _ := provider.StartTrace(ctx, WithWorkflowName("test"))
	defer trace.End(ctx)

	ctx, span, _ := trace.StartSpan(ctx, "custom")
	defer span.End(ctx)

	// Should retrieve span
	if SpanFromContext(ctx) == nil {
		t.Error("Expected span from context, got nil")
	}
}

func TestContextPropagation(t *testing.T) {
	mock := &mockProcessor{}
	provider := NewProvider(mock)
	ctx := context.Background()

	// Create trace
	ctx, trace, _ := provider.StartTrace(ctx, WithWorkflowName("test"))
	defer trace.End(ctx)

	// Create parent span
	ctx, parent, _ := trace.StartSpan(ctx, "agent")

	// Create child span - should auto-detect parent from context
	ctx, child, _ := trace.StartSpan(ctx, "function")

	child.End(ctx)
	parent.End(ctx)

	// Verify parent-child relationship through exported data
	if len(mock.spanEnds) != 2 {
		t.Fatalf("Expected 2 span ends, got %d", len(mock.spanEnds))
	}

	childExport := mock.spanEnds[0]
	parentExport := mock.spanEnds[1]

	if childExport.ParentID != parentExport.ID {
		t.Errorf("Child parent ID %q doesn't match parent ID %q", childExport.ParentID, parentExport.ID)
	}

	// Parent should have no parent
	if parentExport.ParentID != "" {
		t.Errorf("Parent should have no parent, got %q", parentExport.ParentID)
	}
}

func TestContextPropagationMultipleLevels(t *testing.T) {
	mock := &mockProcessor{}
	provider := NewProvider(mock)
	ctx := context.Background()

	ctx, trace, _ := provider.StartTrace(ctx, WithWorkflowName("test"))
	defer trace.End(ctx)

	// Level 1
	ctx, span1, _ := trace.StartSpan(ctx, "agent")

	// Level 2
	ctx, span2, _ := trace.StartSpan(ctx, "function")

	// Level 3
	ctx, span3, _ := trace.StartSpan(ctx, "custom")

	span3.End(ctx)
	span2.End(ctx)
	span1.End(ctx)

	// Verify hierarchy
	if len(mock.spanEnds) != 3 {
		t.Fatalf("Expected 3 span ends, got %d", len(mock.spanEnds))
	}

	// span3 -> span2 -> span1 -> (no parent)
	if mock.spanEnds[0].ParentID != mock.spanEnds[1].ID {
		t.Error("Level 3 parent doesn't match level 2")
	}
	if mock.spanEnds[1].ParentID != mock.spanEnds[2].ID {
		t.Error("Level 2 parent doesn't match level 1")
	}
	if mock.spanEnds[2].ParentID != "" {
		t.Error("Level 1 should have no parent")
	}
}
