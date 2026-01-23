package agents

import (
	"time"

	"github.com/MitulShah1/openai-agents-go/jsonschema"
	"github.com/MitulShah1/openai-agents-go/tracing"
)

// RunConfig configures how an agent execution should behave
type RunConfig struct {
	// MaxTurns limits the number of agent loop iterations
	// 0 means unlimited (use with caution)
	MaxTurns int

	// Temperature controls randomness (0.0 to 2.0)
	// If nil, uses agent's default or model default
	Temperature *float64

	// MaxTokens limits response length
	// If nil, uses model default
	MaxTokens *int

	// ParallelToolCalls enables concurrent tool execution
	// Overrides agent's ParallelToolCalls setting if set
	ParallelToolCalls *bool

	// MaxToolConcurrency limits the number of tools running in parallel
	// 0 means unlimited (bounded by tool count)
	MaxToolConcurrency int

	// Debug enables verbose logging
	Debug bool

	// Timeout for the entire agent run
	// 0 means no timeout
	Timeout time.Duration

	// ResponseFormat can override agent's response format
	// If nil, uses agent's ResponseFormat
	ResponseFormat *jsonschema.ResponseFormat

	// Tracing configuration
	// TraceWorkflowName sets the workflow name for the trace (default: "Agent workflow")
	TraceWorkflowName string

	// TraceID sets a specific trace ID (auto-generated if not provided)
	TraceID string

	// TraceGroupID sets the group ID to link multiple traces
	TraceGroupID string

	// TraceMetadata sets custom metadata for the trace
	TraceMetadata map[string]any

	// TraceEnabled enables/disables tracing (nil uses global default)
	TraceEnabled *bool

	// TraceIncludeSensitiveData controls whether sensitive data is included in traces
	TraceIncludeSensitiveData *bool

	// TraceAPIKey sets the API key for trace export (optional)
	TraceAPIKey string
}

// DefaultRunConfig returns sensible defaults
func DefaultRunConfig() *RunConfig {
	return &RunConfig{
		MaxTurns:          10,
		Debug:             false,
		Timeout:           5 * time.Minute,
		TraceWorkflowName: "Agent workflow",
	}
}

// Merge creates a new config with overrides applied
func (c *RunConfig) Merge(overrides *RunConfig) *RunConfig {
	if overrides == nil {
		return c
	}

	result := *c // copy

	if overrides.MaxTurns > 0 {
		result.MaxTurns = overrides.MaxTurns
	}
	if overrides.Temperature != nil {
		result.Temperature = overrides.Temperature
	}
	if overrides.MaxTokens != nil {
		result.MaxTokens = overrides.MaxTokens
	}
	if overrides.ParallelToolCalls != nil {
		result.ParallelToolCalls = overrides.ParallelToolCalls
	}
	if overrides.MaxToolConcurrency > 0 {
		result.MaxToolConcurrency = overrides.MaxToolConcurrency
	}
	if overrides.Debug {
		result.Debug = true
	}
	if overrides.Timeout > 0 {
		result.Timeout = overrides.Timeout
	}
	if overrides.ResponseFormat != nil {
		result.ResponseFormat = overrides.ResponseFormat
	}
	if overrides.TraceWorkflowName != "" {
		result.TraceWorkflowName = overrides.TraceWorkflowName
	}
	if overrides.TraceID != "" {
		result.TraceID = overrides.TraceID
	}
	if overrides.TraceGroupID != "" {
		result.TraceGroupID = overrides.TraceGroupID
	}
	if overrides.TraceMetadata != nil {
		result.TraceMetadata = overrides.TraceMetadata
	}
	if overrides.TraceEnabled != nil {
		result.TraceEnabled = overrides.TraceEnabled
	}
	if overrides.TraceIncludeSensitiveData != nil {
		result.TraceIncludeSensitiveData = overrides.TraceIncludeSensitiveData
	}
	if overrides.TraceAPIKey != "" {
		result.TraceAPIKey = overrides.TraceAPIKey
	}

	return &result
}

// traceOptionsFromRunConfig converts RunConfig tracing options to trace options
func traceOptionsFromRunConfig(cfg *RunConfig) []tracing.TraceOption {
	if cfg == nil {
		cfg = DefaultRunConfig()
	}

	var opts []tracing.TraceOption

	if cfg.TraceWorkflowName != "" {
		opts = append(opts, tracing.WithWorkflowName(cfg.TraceWorkflowName))
	}
	if cfg.TraceID != "" {
		opts = append(opts, tracing.WithTraceID(cfg.TraceID))
	}
	if cfg.TraceGroupID != "" {
		opts = append(opts, tracing.WithGroupID(cfg.TraceGroupID))
	}
	if cfg.TraceMetadata != nil {
		opts = append(opts, tracing.WithMetadata(cfg.TraceMetadata))
	}
	if cfg.TraceEnabled != nil && !*cfg.TraceEnabled {
		opts = append(opts, tracing.Disabled())
	}

	includeSensitive := tracing.IncludeSensitiveDataDefault()
	if cfg.TraceIncludeSensitiveData != nil {
		includeSensitive = *cfg.TraceIncludeSensitiveData
	}
	opts = append(opts, tracing.WithSensitiveData(includeSensitive))

	if cfg.TraceAPIKey != "" {
		opts = append(opts, tracing.WithExportAPIKey(cfg.TraceAPIKey))
	}

	return opts
}
