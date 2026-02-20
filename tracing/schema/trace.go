package schema

// IngestPayload represents the payload sent to the OpenAI Traces ingest API.
// The API expects a top-level "data" array containing both traces and spans,
// matching the Python SDK's payload = {"data": data} format.
type IngestPayload struct {
	Data []any `json:"data"`
}

// TraceExport represents a trace in the OpenAI Traces export format.
type TraceExport struct {
	Object       string         `json:"object"` // "trace"
	ID           string         `json:"id"`
	WorkflowName string         `json:"workflow_name,omitempty"`
	GroupID      string         `json:"group_id,omitempty"`
	Metadata     map[string]any `json:"metadata,omitempty"`
}
