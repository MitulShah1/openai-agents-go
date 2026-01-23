package tracing

import (
	"context"
	"sync"
	"time"

	"github.com/MitulShah1/openai-agents-go/tracing/internal"
	"github.com/MitulShah1/openai-agents-go/tracing/processor"
	"github.com/MitulShah1/openai-agents-go/tracing/schema"
)

// trace implements the Trace interface.
type trace struct {
	id                   string
	workflowName         string
	groupID              string
	metadata             map[string]any
	includeSensitiveData bool
	exportAPIKey         string
	startedAt            time.Time

	processors []processor.Processor
	pool       *sync.Pool // Pool to return to when done
	spanPool   sync.Pool  // Pool for span objects

	mu sync.RWMutex
}

func (t *trace) ID() string {
	return t.id
}

func (t *trace) WorkflowName() string {
	return t.workflowName
}

func (t *trace) StartSpan(ctx context.Context, spanType schema.SpanType, opts ...SpanOption) (context.Context, Span, error) {
	// Configure span
	cfg := &spanConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	if spanType == "" {
		spanType = schema.SpanTypeCustom
	}

	// Initialize span pool if needed (with mutex protection to prevent race)
	t.mu.Lock()
	if t.spanPool.New == nil {
		t.spanPool.New = func() any {
			return &span{}
		}
	}
	t.mu.Unlock()

	// Get span from pool and initialize
	s := t.spanPool.Get().(*span)
	s.id = internal.GenSpanID()
	s.traceID = t.id
	s.spanType = spanType
	s.startedAt = time.Now().UTC()
	s.name = cfg.name
	s.trace = t
	s.pool = &t.spanPool
	s.endedAt = time.Time{} // Reset
	s.err = nil             // Reset

	// Initialize or reset attributes
	if cfg.attributes != nil {
		if s.attrs == nil {
			s.attrs = make(map[string]any, len(cfg.attributes))
		} else {
			// Clear existing attributes
			for k := range s.attrs {
				delete(s.attrs, k)
			}
		}
		for k, v := range cfg.attributes {
			s.attrs[k] = v
		}
	} else if s.attrs != nil {
		// Clear attributes if none provided
		for k := range s.attrs {
			delete(s.attrs, k)
		}
	}

	// Automatically detect parent span from context
	if parentSpan := SpanFromContext(ctx); parentSpan != nil {
		s.parentID = parentSpan.ID()
	} else {
		s.parentID = "" // Reset
	}

	// Store the new span in context for automatic parent detection
	return ContextWithSpan(ctx, s), s, nil
}

func (t *trace) End(_ context.Context) {
	t.mu.RLock()
	processors := t.processors
	exportAPIKey := t.exportAPIKey
	t.mu.RUnlock()

	export := schema.TraceExport{
		Object:       "trace",
		ID:           t.id,
		WorkflowName: t.workflowName,
		GroupID:      t.groupID,
		Metadata:     t.metadata,
	}

	for _, p := range processors {
		p.OnTraceEnd(export, exportAPIKey)
	}

	// Return trace to pool
	if t.pool != nil {
		// Clear fields to prevent memory leaks
		t.processors = nil
		t.metadata = nil
		t.pool.Put(t)
	}
}

// span implements the Span interface.
type span struct {
	id        string
	traceID   string
	parentID  string
	spanType  schema.SpanType
	startedAt time.Time
	name      string

	mu      sync.Mutex
	endedAt time.Time
	attrs   map[string]any
	err     error

	trace *trace
	pool  *sync.Pool // Pool to return to when done
}

func (s *span) ID() string {
	return s.id
}

func (s *span) TraceID() string {
	return s.traceID
}

func (s *span) Type() schema.SpanType {
	return s.spanType
}

func (s *span) RecordError(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.err = err
}

func (s *span) SetAttributes(attrs map[string]any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.attrs == nil {
		s.attrs = make(map[string]any)
	}
	for k, v := range attrs {
		s.attrs[k] = v
	}
}

func (s *span) End(_ context.Context) {
	s.mu.Lock()
	if !s.endedAt.IsZero() {
		s.mu.Unlock()
		return // Already ended
	}
	s.endedAt = time.Now().UTC()

	// Convert attributes to specific SpanData type based on span type
	spanData := s.toSpanData()

	var spanErr *schema.SpanError
	if s.err != nil {
		spanErr = &schema.SpanError{
			Message: s.err.Error(),
		}
	}

	export := schema.SpanExport{
		Object:    "trace.span",
		ID:        s.id,
		TraceID:   s.traceID,
		ParentID:  s.parentID,
		StartedAt: s.startedAt.Format(time.RFC3339Nano),
		EndedAt:   s.endedAt.Format(time.RFC3339Nano),
		SpanData:  spanData,
		Error:     spanErr,
	}

	processors := s.trace.processors
	exportAPIKey := s.trace.exportAPIKey
	pool := s.pool
	s.mu.Unlock()

	// Notify processors
	for _, p := range processors {
		p.OnSpanEnd(export, exportAPIKey)
	}

	// Return span to pool
	if pool != nil {
		// Clear fields to prevent memory leaks
		s.trace = nil
		s.err = nil
		// Keep attrs map but it will be cleared on next use
		pool.Put(s)
	}
}

func (s *span) toSpanData() schema.SpanData {
	attrs := s.attrs
	if attrs == nil {
		attrs = make(map[string]any)
	}

	includeSensitive := s.trace.includeSensitiveData

	switch s.spanType {
	case schema.SpanTypeAgent:
		return schema.AgentSpanData{
			Type:         string(schema.SpanTypeAgent),
			Name:         s.name,
			Model:        getString(attrs, "model"),
			Instructions: getString(attrs, "instructions"),
		}
	case schema.SpanTypeGeneration:
		return schema.GenerationSpanData{
			Type:     string(schema.SpanTypeGeneration),
			Model:    getString(attrs, "model"),
			Request:  internal.RedactSensitive(attrs["request"], includeSensitive),
			Response: internal.RedactSensitive(attrs["response"], includeSensitive),
			Usage:    getUsage(attrs, "usage"),
		}
	case schema.SpanTypeFunction:
		return schema.FunctionSpanData{
			Type:      string(schema.SpanTypeFunction),
			Name:      s.name,
			Arguments: internal.RedactSensitive(attrs["arguments"], includeSensitive),
			Output:    internal.RedactSensitive(attrs["output"], includeSensitive),
		}
	case schema.SpanTypeGuardrail:
		var tripwire *bool
		if v, ok := attrs["tripwire_triggered"].(bool); ok {
			tripwire = &v
		}
		return schema.GuardrailSpanData{
			Type:              string(schema.SpanTypeGuardrail),
			Name:              s.name,
			GuardrailType:     getString(attrs, "guardrail_type"),
			TripwireTriggered: tripwire,
			Message:           getString(attrs, "message"),
		}
	case schema.SpanTypeHandoff:
		return schema.HandoffSpanData{
			Type:      string(schema.SpanTypeHandoff),
			FromAgent: getString(attrs, "from_agent"),
			ToAgent:   getString(attrs, "to_agent"),
			Reason:    getString(attrs, "reason"),
		}
	default:
		// Fallback for custom or unknown types
		// In a real implementation this might need a flexible map type,
		// but the schema implies specific structures.
		// For now we map to a minimal structure or invalid one.
		// Let's assume custom usage maps to FunctionSpanData loosely or similar,
		// but since strict schema compliance is required, we'll return a zero struct of a safe type
		// or maybe GenerationSpanData is the most generic?
		// Actually, let's just return a basic implementation.
		return schema.FunctionSpanData{
			Type: string(s.spanType),
			Name: s.name,
		}
	}
}

func getString(m map[string]any, k string) string {
	if v, ok := m[k].(string); ok {
		return v
	}
	return ""
}

func getUsage(m map[string]any, k string) *schema.Usage {
	if v, ok := m[k].(*schema.Usage); ok {
		return v
	}
	return nil
}
