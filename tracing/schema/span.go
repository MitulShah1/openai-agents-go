package schema

// SpanExport represents a span in the OpenAI Traces export format.
type SpanExport struct {
	Object    string     `json:"object"` // "trace.span"
	ID        string     `json:"id"`
	TraceID   string     `json:"trace_id"`
	ParentID  string     `json:"parent_id,omitempty"`
	StartedAt string     `json:"started_at"` // RFC3339Nano
	EndedAt   string     `json:"ended_at"`   // RFC3339Nano
	SpanData  SpanData   `json:"span_data"`
	Error     *SpanError `json:"error,omitempty"`
}

// SpanData is an interface over concrete span data structs.
// Each implementation must JSON-marshal to an object containing a "type" field.
type SpanData interface {
	isSpanData()
}

// AgentSpanData represents data for an agent span.
type AgentSpanData struct {
	Type       string   `json:"type"` // "agent"
	Name       string   `json:"name,omitempty"`
	Handoffs   []string `json:"handoffs,omitempty"`
	Tools      []string `json:"tools,omitempty"`
	OutputType string   `json:"output_type,omitempty"`
}

func (AgentSpanData) isSpanData() {}

// GenerationSpanData represents data for a generation span.
type GenerationSpanData struct {
	Type   string `json:"type"` // "generation"
	Model  string `json:"model,omitempty"`
	Input  any    `json:"input,omitempty"`
	Output any    `json:"output,omitempty"`
	Usage  *Usage `json:"usage,omitempty"`
}

func (GenerationSpanData) isSpanData() {}

// FunctionSpanData represents data for a function/tool span.
type FunctionSpanData struct {
	Type   string `json:"type"` // "function"
	Name   string `json:"name,omitempty"`
	Input  any    `json:"input,omitempty"`
	Output any    `json:"output,omitempty"`
}

func (FunctionSpanData) isSpanData() {}

// GuardrailSpanData represents data for a guardrail span.
type GuardrailSpanData struct {
	Type              string `json:"type"` // "guardrail"
	Name              string `json:"name,omitempty"`
	GuardrailType     string `json:"guardrail_type,omitempty"`
	TripwireTriggered *bool  `json:"tripwire_triggered,omitempty"`
	Message           string `json:"message,omitempty"`
}

func (GuardrailSpanData) isSpanData() {}

// HandoffSpanData represents data for a handoff span.
type HandoffSpanData struct {
	Type      string `json:"type"` // "handoff"
	FromAgent string `json:"from_agent,omitempty"`
	ToAgent   string `json:"to_agent,omitempty"`
	Reason    string `json:"reason,omitempty"`
}

func (HandoffSpanData) isSpanData() {}
