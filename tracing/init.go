package tracing

import (
	"github.com/MitulShah1/openai-agents-go/tracing/exporter"
	"github.com/MitulShah1/openai-agents-go/tracing/processor"
)

// Default global wiring:
// - Tracing enabled by default unless env disables it.
// - Default provider exports to OpenAI traces ingest via a batch processor.
func init() {
	if IsTracingDisabled() {
		SetProvider(NewNoopProvider())
		return
	}

	exp := exporter.NewBackendExporter()
	proc := processor.NewBatch(exp)
	SetProvider(NewProvider(proc))

	// Initialize signal handler for graceful shutdown
	initSignalHandler()
}
