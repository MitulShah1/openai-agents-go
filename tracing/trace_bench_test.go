package tracing

import (
	"context"
	"testing"

	"github.com/MitulShah1/openai-agents-go/tracing/schema"
)

// BenchmarkTraceCreation measures the overhead of creating traces.
func BenchmarkTraceCreation(b *testing.B) {
	provider := NewProvider()
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ctx, trace, _ := provider.StartTrace(ctx,
			WithWorkflowName("benchmark"),
		)
		trace.End(ctx)
	}
}

// BenchmarkSpanCreation measures the overhead of creating spans.
func BenchmarkSpanCreation(b *testing.B) {
	provider := NewProvider()
	ctx := context.Background()
	ctx, trace, _ := provider.StartTrace(ctx,
		WithWorkflowName("benchmark"),
	)
	defer trace.End(ctx)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ctx, span, _ := trace.StartSpan(ctx, schema.SpanTypeCustom,
			WithSpanName("test"),
		)
		span.End(ctx)
	}
}

// BenchmarkNestedSpans measures the overhead of creating nested spans.
func BenchmarkNestedSpans(b *testing.B) {
	provider := NewProvider()
	ctx := context.Background()
	ctx, trace, _ := provider.StartTrace(ctx,
		WithWorkflowName("benchmark"),
	)
	defer trace.End(ctx)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ctx, span1, _ := trace.StartSpan(ctx, schema.SpanTypeAgent,
			WithSpanName("parent"),
		)
		ctx, span2, _ := trace.StartSpan(ctx, schema.SpanTypeFunction,
			WithSpanName("child"),
		)
		span2.End(ctx)
		span1.End(ctx)
	}
}

// BenchmarkSpanWithAttributes measures the overhead of setting attributes.
func BenchmarkSpanWithAttributes(b *testing.B) {
	provider := NewProvider()
	ctx := context.Background()
	ctx, trace, _ := provider.StartTrace(ctx,
		WithWorkflowName("benchmark"),
	)
	defer trace.End(ctx)

	attrs := map[string]any{
		"model":        "gpt-4",
		"temperature":  0.7,
		"max_tokens":   1000,
		"instructions": "You are a helpful assistant",
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ctx, span, _ := trace.StartSpan(ctx, schema.SpanTypeAgent,
			WithSpanName("test"),
			WithAttributes(attrs),
		)
		span.End(ctx)
	}
}

// BenchmarkConcurrentTraces measures concurrent trace creation.
func BenchmarkConcurrentTraces(b *testing.B) {
	provider := NewProvider()

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		for pb.Next() {
			ctx, trace, _ := provider.StartTrace(ctx,
				WithWorkflowName("benchmark"),
			)
			ctx, span, _ := trace.StartSpan(ctx, schema.SpanTypeCustom,
				WithSpanName("test"),
			)
			span.End(ctx)
			trace.End(ctx)
		}
	})
}

// BenchmarkFactoryFunctions measures the overhead of type-safe factory functions.
func BenchmarkFactoryFunctions(b *testing.B) {
	provider := NewProvider()
	ctx := context.Background()
	ctx, trace, _ := provider.StartTrace(ctx,
		WithWorkflowName("benchmark"),
	)
	ctx = ContextWithTrace(ctx, trace)
	defer trace.End(ctx)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ctx, span, _ := AgentSpan(ctx, "test-agent",
			WithModel("gpt-4"),
		)
		span.End(ctx)
	}
}
