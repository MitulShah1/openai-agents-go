package schema

// SpanType represents the type of a span.
type SpanType string

const (
	// SpanTypeAgent represents an agent span.
	SpanTypeAgent SpanType = "agent"
	// SpanTypeGeneration represents a generation span.
	SpanTypeGeneration SpanType = "generation"
	// SpanTypeFunction represents a function/tool span.
	SpanTypeFunction SpanType = "function"
	// SpanTypeGuardrail represents a guardrail span.
	SpanTypeGuardrail SpanType = "guardrail"
	// SpanTypeHandoff represents a handoff span.
	SpanTypeHandoff SpanType = "handoff"
	// SpanTypeCustom represents a custom span type.
	SpanTypeCustom SpanType = "custom"
)

// Usage represents token usage for generation spans.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens,omitempty"`
	CompletionTokens int `json:"completion_tokens,omitempty"`
	TotalTokens      int `json:"total_tokens,omitempty"`
}

// SpanError represents an error that occurred during span execution.
type SpanError struct {
	Message string `json:"message,omitempty"`
	Type    string `json:"type,omitempty"`
	Code    string `json:"code,omitempty"`
}
