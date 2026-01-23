package tracing

import (
	"github.com/MitulShah1/openai-agents-go/tracing/exporter"
	"github.com/MitulShah1/openai-agents-go/tracing/processor"
)

// AddTraceProcessor adds a processor to the global trace provider.
// The processor will receive all trace and span events.
//
// This is thread-safe and can be called while tracing is active.
//
// Example:
//
//	consoleExporter := exporter.NewConsole()
//	consoleProcessor := processor.NewBatch(consoleExporter)
//	tracing.AddTraceProcessor(consoleProcessor)
func AddTraceProcessor(p processor.Processor) {
	provider := GetProvider()
	if dp, ok := provider.(*defaultProvider); ok {
		dp.AddProcessor(p)
	}
}

// SetTraceProcessors replaces all processors with the given list.
// This is thread-safe and can be called while tracing is active.
//
// Example:
//
//	processors := []processor.Processor{
//	    processor.NewBatch(exporter.NewBackend()),
//	    processor.NewBatch(exporter.NewConsole()),
//	}
//	tracing.SetTraceProcessors(processors)
func SetTraceProcessors(processors []processor.Processor) {
	provider := GetProvider()
	if dp, ok := provider.(*defaultProvider); ok {
		dp.SetProcessors(processors)
	}
}

// SetTracingDisabled enables or disables tracing globally.
// When disabled, all trace and span operations become no-ops.
//
// This is useful for temporarily disabling tracing without changing code.
//
// Example:
//
//	// Disable tracing in tests
//	tracing.SetTracingDisabled(true)
//	defer tracing.SetTracingDisabled(false)
func SetTracingDisabled(disabled bool) {
	if disabled {
		SetProvider(NewNoopProvider())
	} else if !IsTracingDisabled() {
		// Re-initialize with default provider
		// This matches the init() behavior
		exp := exporter.NewBackendExporter()
		proc := processor.NewBatch(exp)
		SetProvider(NewProvider(proc))
	}
}
