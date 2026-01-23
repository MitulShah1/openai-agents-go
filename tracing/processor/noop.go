package processor

import (
	"context"

	"github.com/MitulShah1/openai-agents-go/tracing/schema"
)

// noopProcessor is a no-op processor that discards all events.
type noopProcessor struct{}

// NewNoopProcessor returns a processor that discards all events.
func NewNoopProcessor() Processor {
	return &noopProcessor{}
}

func (p *noopProcessor) OnTraceStart(_ schema.TraceExport, _ string) {}
func (p *noopProcessor) OnSpanEnd(_ schema.SpanExport, _ string)     {}
func (p *noopProcessor) OnTraceEnd(_ schema.TraceExport, _ string)   {}
func (p *noopProcessor) Shutdown(_ context.Context) error            { return nil }
