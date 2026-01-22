package exporter

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/MitulShah1/openai-agents-go/tracing/schema"
)

// ConsoleExporter exports traces and spans to stdout.
type ConsoleExporter struct{}

// NewConsoleExporter returns a new ConsoleExporter.
func NewConsoleExporter() Exporter {
	return &ConsoleExporter{}
}

// Export exports the given traces and spans to the console.
func (e *ConsoleExporter) Export(_ context.Context, _ string, traces []schema.TraceExport, spans []schema.SpanExport) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")

	for _, t := range traces {
		fmt.Println("--- Trace Started ---")
		if err := enc.Encode(t); err != nil {
			return err
		}
	}

	for _, s := range spans {
		fmt.Println("--- Span Ended ---")
		if err := enc.Encode(s); err != nil {
			return err
		}
	}

	return nil
}

// Close closes the exporter.
func (e *ConsoleExporter) Close() error {
	return nil
}
