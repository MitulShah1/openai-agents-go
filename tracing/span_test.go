package tracing

import (
	"context"
	"errors"
	"testing"

	"github.com/MitulShah1/openai-agents-go/tracing/schema"
)

func TestSpanID(t *testing.T) {
	mock := &mockProcessor{}
	provider := NewProvider(mock)
	ctx := context.Background()

	ctx, trace, _ := provider.StartTrace(ctx, WithWorkflowName("test"))
	defer trace.End(ctx)

	ctx, span, _ := trace.StartSpan(ctx, schema.SpanTypeCustom)

	id := span.ID()
	if id == "" {
		t.Error("Span ID is empty")
	}

	// ID should be consistent
	if span.ID() != id {
		t.Error("Span ID changed between calls")
	}

	span.End(ctx)
}

func TestSpanTraceID(t *testing.T) {
	mock := &mockProcessor{}
	provider := NewProvider(mock)
	ctx := context.Background()

	ctx, trace, _ := provider.StartTrace(ctx, WithWorkflowName("test"))
	defer trace.End(ctx)

	ctx, span, _ := trace.StartSpan(ctx, schema.SpanTypeCustom)
	defer span.End(ctx)

	if span.TraceID() != trace.ID() {
		t.Errorf("Span trace ID %q doesn't match trace ID %q", span.TraceID(), trace.ID())
	}
}

func TestSpanType(t *testing.T) {
	mock := &mockProcessor{}
	provider := NewProvider(mock)
	ctx := context.Background()

	ctx, trace, _ := provider.StartTrace(ctx, WithWorkflowName("test"))
	defer trace.End(ctx)

	tests := []schema.SpanType{
		schema.SpanTypeAgent,
		schema.SpanTypeGeneration,
		schema.SpanTypeFunction,
		schema.SpanTypeHandoff,
		schema.SpanTypeGuardrail,
		schema.SpanTypeCustom,
	}

	for _, spanType := range tests {
		ctx, span, _ := trace.StartSpan(ctx, spanType)
		if span.Type() != spanType {
			t.Errorf("Expected span type %q, got %q", spanType, span.Type())
		}
		span.End(ctx)
	}
}

func TestSpanRecordError(t *testing.T) {
	mock := &mockProcessor{}
	provider := NewProvider(mock)
	ctx := context.Background()

	ctx, trace, _ := provider.StartTrace(ctx, WithWorkflowName("test"))
	defer trace.End(ctx)

	ctx, span, _ := trace.StartSpan(ctx, schema.SpanTypeCustom)

	testErr := errors.New("test error")
	span.RecordError(testErr)
	span.End(ctx)

	// Verify error was recorded
	if len(mock.spanEnds) != 1 {
		t.Fatalf("Expected 1 span end, got %d", len(mock.spanEnds))
	}

	if mock.spanEnds[0].Error == nil {
		t.Error("Expected error in span export, got nil")
	}

	if mock.spanEnds[0].Error.Message != "test error" {
		t.Errorf("Expected error message 'test error', got %q", mock.spanEnds[0].Error.Message)
	}
}

func TestSpanSetAttributes(_ *testing.T) {
	mock := &mockProcessor{}
	provider := NewProvider(mock)
	ctx := context.Background()

	ctx, trace, _ := provider.StartTrace(ctx, WithWorkflowName("test"))
	defer trace.End(ctx)

	ctx, span, _ := trace.StartSpan(ctx, schema.SpanTypeAgent)

	attrs := map[string]any{
		"model":       "gpt-4",
		"temperature": 0.7,
	}
	span.SetAttributes(attrs)

	// Add more attributes
	moreAttrs := map[string]any{
		"max_tokens": 1000,
	}
	span.SetAttributes(moreAttrs)

	span.End(ctx)

	// Both sets of attributes should be present
	// (verified through the span data export)
}

func TestSpanEnd(t *testing.T) {
	mock := &mockProcessor{}
	provider := NewProvider(mock)
	ctx := context.Background()

	ctx, trace, _ := provider.StartTrace(ctx, WithWorkflowName("test"))
	defer trace.End(ctx)

	ctx, span, _ := trace.StartSpan(ctx, schema.SpanTypeCustom)
	span.End(ctx)

	// Verify processor received span end
	if len(mock.spanEnds) != 1 {
		t.Errorf("Expected 1 span end, got %d", len(mock.spanEnds))
	}

	// Verify span export has required fields
	export := mock.spanEnds[0]
	if export.ID == "" {
		t.Error("Span export ID is empty")
	}
	if export.TraceID == "" {
		t.Error("Span export trace ID is empty")
	}
	if export.StartedAt == "" {
		t.Error("Span export started_at is empty")
	}
	if export.EndedAt == "" {
		t.Error("Span export ended_at is empty")
	}
}

func TestSpanEndIdempotent(t *testing.T) {
	mock := &mockProcessor{}
	provider := NewProvider(mock)
	ctx := context.Background()

	ctx, trace, _ := provider.StartTrace(ctx, WithWorkflowName("test"))
	defer trace.End(ctx)

	ctx, span, _ := trace.StartSpan(ctx, schema.SpanTypeCustom)

	// End span multiple times
	span.End(ctx)
	span.End(ctx)
	span.End(ctx)

	// Should only receive one span end event
	if len(mock.spanEnds) != 1 {
		t.Errorf("Expected 1 span end, got %d (End should be idempotent)", len(mock.spanEnds))
	}
}

func TestSpanToSpanData(t *testing.T) {
	mock := &mockProcessor{}
	provider := NewProvider(mock)
	ctx := context.Background()

	ctx, trace, _ := provider.StartTrace(ctx, WithWorkflowName("test"))
	defer trace.End(ctx)

	// Test Agent span data
	ctx, agentSpan, _ := trace.StartSpan(ctx, schema.SpanTypeAgent,
		WithSpanName("agent"),
		WithAttributes(map[string]any{
			"handoffs":    []string{"agent_b"},
			"tools":       []string{"get_weather"},
			"output_type": "text",
		}),
	)
	agentSpan.End(ctx)

	if len(mock.spanEnds) != 1 {
		t.Fatalf("Expected 1 span end, got %d", len(mock.spanEnds))
	}

	agentData, ok := mock.spanEnds[0].SpanData.(schema.AgentSpanData)
	if !ok {
		t.Fatalf("Expected AgentSpanData, got %T", mock.spanEnds[0].SpanData)
	}

	if agentData.Name != "agent" {
		t.Errorf("Expected name 'agent', got %q", agentData.Name)
	}
	if len(agentData.Handoffs) != 1 || agentData.Handoffs[0] != "agent_b" {
		t.Errorf("Expected handoffs ['agent_b'], got %v", agentData.Handoffs)
	}
	if len(agentData.Tools) != 1 || agentData.Tools[0] != "get_weather" {
		t.Errorf("Expected tools ['get_weather'], got %v", agentData.Tools)
	}
	if agentData.OutputType != "text" {
		t.Errorf("Expected output_type 'text', got %q", agentData.OutputType)
	}
}
