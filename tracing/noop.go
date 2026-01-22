package tracing

import (
	"context"

	"github.com/MitulShah1/openai-agents-go/tracing/schema"
)

// noopProvider is a no-op provider that creates no-op traces.
type noopProvider struct{}

// NewNoopProvider returns a provider that creates no-op traces.
func NewNoopProvider() Provider {
	return &noopProvider{}
}

func (p *noopProvider) StartTrace(ctx context.Context, _ ...TraceOption) (context.Context, Trace, error) {
	t := &noopTrace{}
	return ContextWithTrace(ctx, t), t, nil
}

func (p *noopProvider) Shutdown(_ context.Context) error {
	return nil
}

// noopTrace is a no-op trace that creates no-op spans.
type noopTrace struct{}

func (t *noopTrace) ID() string           { return "" }
func (t *noopTrace) WorkflowName() string { return "" }

func (t *noopTrace) StartSpan(ctx context.Context, _ schema.SpanType, _ ...SpanOption) (context.Context, Span, error) {
	s := &noopSpan{}
	return ctx, s, nil
}

func (t *noopTrace) End(_ context.Context) {}

// noopSpan is a no-op span that does nothing.
type noopSpan struct{}

func (s *noopSpan) ID() string                     { return "" }
func (s *noopSpan) TraceID() string                { return "" }
func (s *noopSpan) Type() schema.SpanType          { return "" }
func (s *noopSpan) RecordError(_ error)            {}
func (s *noopSpan) SetAttributes(_ map[string]any) {}
func (s *noopSpan) End(_ context.Context)          {}
