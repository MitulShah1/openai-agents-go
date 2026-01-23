package tracing

import (
	"context"
	"testing"
)

func TestNoopProvider(t *testing.T) {
	provider := NewNoopProvider()

	if provider == nil {
		t.Fatal("NewNoopProvider returned nil")
	}

	ctx := context.Background()

	// Should be able to start trace
	_, trace, err := provider.StartTrace(ctx)
	if err != nil {
		t.Errorf("NoopProvider.StartTrace failed: %v", err)
	}

	if trace == nil {
		t.Fatal("NoopProvider.StartTrace returned nil trace")
	}

	// Should be no-op trace
	if _, ok := trace.(*noopTrace); !ok {
		t.Errorf("Expected noopTrace, got %T", trace)
	}

	// Shutdown should not error
	if err := provider.Shutdown(ctx); err != nil {
		t.Errorf("NoopProvider.Shutdown failed: %v", err)
	}
}

func TestNoopTrace(t *testing.T) {
	trace := &noopTrace{}

	// ID should return empty string
	if trace.ID() != "" {
		t.Errorf("Expected empty ID, got %q", trace.ID())
	}

	// WorkflowName should return empty string
	if trace.WorkflowName() != "" {
		t.Errorf("Expected empty workflow name, got %q", trace.WorkflowName())
	}

	// StartSpan should return no-op span
	_, span, err := trace.StartSpan(context.TODO(), "custom")
	if err != nil {
		t.Errorf("noopTrace.StartSpan failed: %v", err)
	}

	if _, ok := span.(*noopSpan); !ok {
		t.Errorf("Expected noopSpan, got %T", span)
	}

	// End should not panic
	trace.End(context.TODO())
}

func TestNoopSpan(t *testing.T) {
	span := &noopSpan{}

	// ID should return empty string
	if span.ID() != "" {
		t.Errorf("Expected empty ID, got %q", span.ID())
	}

	// TraceID should return empty string
	if span.TraceID() != "" {
		t.Errorf("Expected empty trace ID, got %q", span.TraceID())
	}

	// Type should return empty string
	if span.Type() != "" {
		t.Errorf("Expected empty type, got %q", span.Type())
	}

	// RecordError should not panic
	span.RecordError(nil)

	// SetAttributes should not panic
	span.SetAttributes(nil)
	span.SetAttributes(map[string]any{"key": "value"})

	// End should not panic
	span.End(context.TODO())
}
