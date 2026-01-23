package schema

// IngestPayload represents the payload sent to the OpenAI Traces ingest API.
type IngestPayload struct {
	Traces []TraceExport `json:"traces,omitempty"`
	Spans  []SpanExport  `json:"spans,omitempty"`
}

// TraceExport represents a trace in the OpenAI Traces export format.
type TraceExport struct {
	Object       string         `json:"object"` // "trace"
	ID           string         `json:"id"`
	WorkflowName string         `json:"workflow_name,omitempty"`
	GroupID      string         `json:"group_id,omitempty"`
	Metadata     map[string]any `json:"metadata,omitempty"`
}
