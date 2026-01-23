package tracing

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/MitulShah1/openai-agents-go/tracing/internal"
	"github.com/MitulShah1/openai-agents-go/tracing/processor"
	"github.com/MitulShah1/openai-agents-go/tracing/schema"
)

var (
	globalProviderMu sync.RWMutex
	globalProvider   = NewNoopProvider()
)

// SetProvider sets the global trace provider.
func SetProvider(p Provider) {
	if p == nil {
		p = NewNoopProvider()
	}
	globalProviderMu.Lock()
	globalProvider = p
	globalProviderMu.Unlock()
}

// GetProvider returns the global trace provider.
func GetProvider() Provider {
	globalProviderMu.RLock()
	p := globalProvider
	globalProviderMu.RUnlock()
	return p
}

// defaultProvider implements the Provider interface with multi-processor support.
type defaultProvider struct {
	multi     *multiProcessor
	tracePool sync.Pool
}

// NewProvider creates a new trace provider with the given processors.
func NewProvider(processors ...processor.Processor) Provider {
	// IF tracing is globally disabled via env var, return no-op
	if IsTracingDisabled() {
		return NewNoopProvider()
	}
	dp := &defaultProvider{
		multi: newMultiProcessor(processors...),
	}
	dp.tracePool.New = func() any {
		return &trace{}
	}
	return dp
}

// AddProcessor adds a processor to the provider.
func (p *defaultProvider) AddProcessor(proc processor.Processor) {
	p.multi.AddProcessor(proc)
}

// SetProcessors replaces all processors.
func (p *defaultProvider) SetProcessors(processors []processor.Processor) {
	p.multi.SetProcessors(processors)
}

func (p *defaultProvider) StartTrace(ctx context.Context, opts ...TraceOption) (context.Context, Trace, error) {
	cfg := &traceConfig{
		workflowName:         "Agent Workflow",
		includeSensitiveData: IncludeSensitiveDataDefault(),
	}

	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.disabled || IsTracingDisabled() {
		return NewNoopProvider().StartTrace(ctx)
	}

	// Validation
	if cfg.traceID != "" && !internal.IsValidTraceID(cfg.traceID) {
		startCtx, startTrace, _ := NewNoopProvider().StartTrace(ctx)
		return startCtx, startTrace, fmt.Errorf("invalid trace ID: %s", cfg.traceID)
	}
	if cfg.groupID != "" && !internal.IsValidGroupID(cfg.groupID) {
		startCtx, startTrace, _ := NewNoopProvider().StartTrace(ctx)
		return startCtx, startTrace, fmt.Errorf("invalid group ID: %s", cfg.groupID)
	}

	// Generate IDs if missing
	if cfg.traceID == "" {
		cfg.traceID = internal.GenTraceID()
	}

	// Get trace from pool and initialize
	t := p.tracePool.Get().(*trace)
	t.id = cfg.traceID
	t.workflowName = cfg.workflowName
	t.groupID = cfg.groupID
	t.metadata = cfg.metadata
	t.includeSensitiveData = cfg.includeSensitiveData
	t.exportAPIKey = cfg.exportAPIKey
	t.startedAt = time.Now().UTC()
	t.processors = p.multi.GetProcessors()
	t.pool = &p.tracePool

	export := schema.TraceExport{
		Object:       "trace",
		ID:           t.id,
		WorkflowName: t.workflowName,
		GroupID:      cfg.groupID,
		Metadata:     t.metadata,
	}

	// Notify processors via multi-processor
	p.multi.OnTraceStart(export, t.exportAPIKey)

	return ContextWithTrace(ctx, t), t, nil
}

func (p *defaultProvider) Shutdown(ctx context.Context) error {
	return p.multi.Shutdown(ctx)
}

// Helper functions for env vars

// IsTracingDisabled returns true when tracing should be globally disabled.
func IsTracingDisabled() bool {
	if envTruthy(os.Getenv("OPENAI_AGENTS_DISABLE_TRACING")) {
		return true
	}
	legacy := os.Getenv("OPENAI_AGENTS_TRACE_ENABLED")
	if legacy == "" {
		return false
	}
	return !envTruthy(legacy)
}

// IncludeSensitiveDataDefault returns the default sensitive-data behavior.
func IncludeSensitiveDataDefault() bool {
	v := os.Getenv("OPENAI_AGENTS_TRACE_INCLUDE_SENSITIVE_DATA")
	if v == "" {
		return true
	}
	return envTruthy(v)
}

func envTruthy(v string) bool {
	switch v {
	case "1", "true", "TRUE", "yes", "YES", "y", "Y", "on", "ON":
		return true
	default:
		return false
	}
}
