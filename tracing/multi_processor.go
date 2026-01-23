package tracing

import (
	"context"
	"sync"

	"github.com/MitulShah1/openai-agents-go/tracing/processor"
	"github.com/MitulShah1/openai-agents-go/tracing/schema"
)

// multiProcessor forwards trace and span events to multiple processors.
// This enables users to send traces to multiple backends simultaneously.
type multiProcessor struct {
	mu         sync.RWMutex
	processors []processor.Processor
}

// newMultiProcessor creates a new multi-processor.
func newMultiProcessor(processors ...processor.Processor) *multiProcessor {
	return &multiProcessor{
		processors: processors,
	}
}

// OnTraceStart forwards the event to all processors.
func (m *multiProcessor) OnTraceStart(trace schema.TraceExport, apiKey string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, p := range m.processors {
		// Catch panics from individual processors to prevent cascading failures
		func() {
			defer func() {
				if r := recover(); r != nil {
					// TODO: Log panic but continue processing
					// In production, this should use a proper logger
					_ = r // Suppress linter warning
				}
			}()
			p.OnTraceStart(trace, apiKey)
		}()
	}
}

// OnSpanEnd forwards the event to all processors.
func (m *multiProcessor) OnSpanEnd(span schema.SpanExport, apiKey string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, p := range m.processors {
		func() {
			defer func() {
				if r := recover(); r != nil {
					// TODO: Log panic but continue processing
					_ = r // Suppress linter warning
				}
			}()
			p.OnSpanEnd(span, apiKey)
		}()
	}
}

// OnTraceEnd forwards the event to all processors.
func (m *multiProcessor) OnTraceEnd(trace schema.TraceExport, apiKey string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, p := range m.processors {
		func() {
			defer func() {
				if r := recover(); r != nil {
					// TODO: Log panic but continue processing
					_ = r // Suppress linter warning
				}
			}()
			p.OnTraceEnd(trace, apiKey)
		}()
	}
}

// Shutdown shuts down all processors.
func (m *multiProcessor) Shutdown(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var errs []error
	for _, p := range m.processors {
		if err := p.Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		// Return first error for simplicity
		// In production, consider using errors.Join (Go 1.20+)
		return errs[0]
	}
	return nil
}

// AddProcessor adds a processor to the list.
// This is thread-safe and can be called while tracing is active.
func (m *multiProcessor) AddProcessor(p processor.Processor) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.processors = append(m.processors, p)
}

// SetProcessors replaces all processors with the given list.
// This is thread-safe and can be called while tracing is active.
func (m *multiProcessor) SetProcessors(processors []processor.Processor) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.processors = processors
}

// GetProcessors returns a copy of the current processor list.
func (m *multiProcessor) GetProcessors() []processor.Processor {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make([]processor.Processor, len(m.processors))
	copy(result, m.processors)
	return result
}
