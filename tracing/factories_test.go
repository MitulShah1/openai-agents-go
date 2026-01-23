package tracing

import (
	"context"
	"testing"
)

func TestAgentSpan(t *testing.T) {
	mock := &mockProcessor{}
	provider := NewProvider(mock)
	ctx := context.Background()

	ctx, trace, _ := provider.StartTrace(ctx, WithWorkflowName("test"))
	ctx = ContextWithTrace(ctx, trace)
	defer trace.End(ctx)

	ctx, span, err := AgentSpan(ctx, "test-agent",
		WithModel("gpt-4"),
		WithInstructions("test instructions"),
	)

	if err != nil {
		t.Fatalf("AgentSpan failed: %v", err)
	}

	if span == nil {
		t.Fatal("AgentSpan returned nil span")
	}

	span.End(ctx)
}

func TestGenerationSpan(t *testing.T) {
	mock := &mockProcessor{}
	provider := NewProvider(mock)
	ctx := context.Background()

	ctx, trace, _ := provider.StartTrace(ctx, WithWorkflowName("test"))
	ctx = ContextWithTrace(ctx, trace)
	defer trace.End(ctx)

	ctx, span, err := GenerationSpan(ctx,
		WithModel("gpt-4"),
	)

	if err != nil {
		t.Fatalf("GenerationSpan failed: %v", err)
	}

	span.End(ctx)
}

func TestFunctionSpan(t *testing.T) {
	mock := &mockProcessor{}
	provider := NewProvider(mock)
	ctx := context.Background()

	ctx, trace, _ := provider.StartTrace(ctx, WithWorkflowName("test"))
	ctx = ContextWithTrace(ctx, trace)
	defer trace.End(ctx)

	ctx, span, err := FunctionSpan(ctx, "get_weather",
		WithArguments(map[string]any{"city": "SF"}),
	)

	if err != nil {
		t.Fatalf("FunctionSpan failed: %v", err)
	}

	span.End(ctx)
}

func TestHandoffSpan(t *testing.T) {
	mock := &mockProcessor{}
	provider := NewProvider(mock)
	ctx := context.Background()

	ctx, trace, _ := provider.StartTrace(ctx, WithWorkflowName("test"))
	ctx = ContextWithTrace(ctx, trace)
	defer trace.End(ctx)

	ctx, span, err := HandoffSpan(ctx,
		WithFromAgent("triage"),
		WithToAgent("specialist"),
		WithReason("complex issue"),
	)

	if err != nil {
		t.Fatalf("HandoffSpan failed: %v", err)
	}

	span.End(ctx)
}

func TestGuardrailSpan(t *testing.T) {
	mock := &mockProcessor{}
	provider := NewProvider(mock)
	ctx := context.Background()

	ctx, trace, _ := provider.StartTrace(ctx, WithWorkflowName("test"))
	ctx = ContextWithTrace(ctx, trace)
	defer trace.End(ctx)

	ctx, span, err := GuardrailSpan(ctx, "content-filter",
		WithGuardrailType("moderation"),
	)

	if err != nil {
		t.Fatalf("GuardrailSpan failed: %v", err)
	}

	span.End(ctx)
}

func TestCustomSpan(t *testing.T) {
	mock := &mockProcessor{}
	provider := NewProvider(mock)
	ctx := context.Background()

	ctx, trace, _ := provider.StartTrace(ctx, WithWorkflowName("test"))
	ctx = ContextWithTrace(ctx, trace)
	defer trace.End(ctx)

	ctx, span, err := CustomSpan(ctx, "database-query",
		WithAttributes(map[string]any{
			"query": "SELECT * FROM users",
		}),
	)

	if err != nil {
		t.Fatalf("CustomSpan failed: %v", err)
	}

	span.End(ctx)
}

func TestFactoriesWithoutTrace(t *testing.T) {
	ctx := context.Background()

	// All factories should return ErrNoActiveTrace
	tests := []struct {
		name string
		fn   func() (context.Context, Span, error)
	}{
		{"AgentSpan", func() (context.Context, Span, error) {
			return AgentSpan(ctx, "test")
		}},
		{"GenerationSpan", func() (context.Context, Span, error) {
			return GenerationSpan(ctx)
		}},
		{"FunctionSpan", func() (context.Context, Span, error) {
			return FunctionSpan(ctx, "test")
		}},
		{"HandoffSpan", func() (context.Context, Span, error) {
			return HandoffSpan(ctx)
		}},
		{"GuardrailSpan", func() (context.Context, Span, error) {
			return GuardrailSpan(ctx, "test")
		}},
		{"CustomSpan", func() (context.Context, Span, error) {
			return CustomSpan(ctx, "test")
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(_ *testing.T) {
			_, span, err := tt.fn()

			if err != ErrNoActiveTrace {
				t.Errorf("Expected ErrNoActiveTrace, got %v", err)
			}

			// Should return no-op span
			if _, ok := span.(*noopSpan); !ok {
				t.Errorf("Expected noopSpan, got %T", span)
			}
		})
	}
}

func TestFactoriesWithOptions(t *testing.T) {
	mock := &mockProcessor{}
	provider := NewProvider(mock)
	ctx := context.Background()

	ctx, trace, _ := provider.StartTrace(ctx, WithWorkflowName("test"))
	ctx = ContextWithTrace(ctx, trace)
	defer trace.End(ctx)

	// Test that options are properly applied
	ctx, span, _ := AgentSpan(ctx, "agent",
		WithModel("gpt-4"),
		WithInstructions("test"),
		WithAttributes(map[string]any{"key": "value"}),
	)
	span.End(ctx)

	// Verify span was created and ended
	if len(mock.spanEnds) != 1 {
		t.Errorf("Expected 1 span end, got %d", len(mock.spanEnds))
	}
}
