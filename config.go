package agents

import (
	"time"

	"github.com/MitulShah1/openai-agents-go/internal/jsonschema"
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

	// Debug enables verbose logging
	Debug bool

	// Timeout for the entire agent run
	// 0 means no timeout
	Timeout time.Duration

	// ResponseFormat can override agent's response format
	// If nil, uses agent's ResponseFormat
	ResponseFormat *jsonschema.ResponseFormat
}

// DefaultRunConfig returns sensible defaults
func DefaultRunConfig() *RunConfig {
	return &RunConfig{
		MaxTurns: 10,
		Debug:    false,
		Timeout:  5 * time.Minute,
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
	if overrides.Debug {
		result.Debug = true
	}
	if overrides.Timeout > 0 {
		result.Timeout = overrides.Timeout
	}
	if overrides.ResponseFormat != nil {
		result.ResponseFormat = overrides.ResponseFormat
	}

	return &result
}
