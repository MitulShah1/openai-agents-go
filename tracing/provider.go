package tracing

import (
	"context"
	"errors"
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

// provider implements the Provider interface.
type provider struct {
	processors []processor.Processor
}

// NewProvider creates a new trace provider with the given processors.
func NewProvider(processors ...processor.Processor) Provider {
	// IF tracing is globally disabled via env var, return no-op
	if IsTracingDisabled() {
		return NewNoopProvider()
	}
	return &provider{
		processors: processors,
	}
}

func (p *provider) StartTrace(ctx context.Context, opts ...TraceOption) (context.Context, Trace, error) {
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
		return NewNoopProvider().StartTrace(ctx)
		// We return the no-op trace AND the error so the user sees the error
		// but the flow continues safely.
	}
	// Note: We need to return error explicitly if we want users to know.
	// But above call returns (ctx, trace, nil).
	// Let's refactor:
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

	t := &trace{
		id:                   cfg.traceID,
		workflowName:         cfg.workflowName,
		groupID:              cfg.groupID,
		metadata:             cfg.metadata,
		includeSensitiveData: cfg.includeSensitiveData,
		exportAPIKey:         cfg.exportAPIKey,
		startedAt:            time.Now().UTC(),
		processors:           p.processors,
	}

	export := schema.TraceExport{
		Object:       "trace",
		ID:           t.id,
		WorkflowName: t.workflowName,
		GroupID:      t.groupID,
		Metadata:     t.metadata,
	}

	// Notify processors
	for _, pr := range p.processors {
		pr.OnTraceStart(export, t.exportAPIKey)
	}

	return ContextWithTrace(ctx, t), t, nil
}

func (p *provider) Shutdown(ctx context.Context) error {
	var errs []error
	for _, pr := range p.processors {
		if err := pr.Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
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
