// Package models provides abstractions for LLM model providers.
//
// The core interfaces ([Model] and [ModelProvider]) allow the agent runner
// to work with any LLM backend, not just OpenAI. Built-in implementations
// are provided for OpenAI's Chat Completions API, and users can implement
// their own providers for Anthropic, Google, or any other LLM service.
package models

import "github.com/MitulShah1/openai-agents-go/jsonschema"

// ModelSettings holds configuration for a model request.
// These settings are resolved from agent-level and run-level configuration,
// with run-level taking precedence.
type ModelSettings struct {
	// Temperature controls randomness in the model's output (0.0 to 2.0).
	// If nil, the provider default is used.
	Temperature *float64

	// MaxTokens limits the maximum number of tokens in the response.
	// If nil, the provider default is used.
	MaxTokens *int

	// TopP controls nucleus sampling. If nil, the provider default is used.
	TopP *float64

	// FrequencyPenalty penalizes tokens based on frequency in the text so far.
	// If nil, the provider default is used.
	FrequencyPenalty *float64

	// PresencePenalty penalizes tokens based on their presence in the text so far.
	// If nil, the provider default is used.
	PresencePenalty *float64

	// Stop sequences that signal the model to stop generating further tokens.
	Stop []string

	// ParallelToolCalls determines if tools can be called in parallel.
	// If nil, the provider default is used.
	ParallelToolCalls *bool

	// ResponseFormat defines the structure of the response (for structured outputs).
	// If nil, responses will be unstructured text.
	ResponseFormat *jsonschema.ResponseFormat
}

// Resolve merges two ModelSettings, with the override taking precedence
// for any non-nil fields. This is used to merge run-level overrides
// on top of agent-level settings.
func (s ModelSettings) Resolve(override ModelSettings) ModelSettings {
	result := s

	if override.Temperature != nil {
		result.Temperature = override.Temperature
	}
	if override.MaxTokens != nil {
		result.MaxTokens = override.MaxTokens
	}
	if override.TopP != nil {
		result.TopP = override.TopP
	}
	if override.FrequencyPenalty != nil {
		result.FrequencyPenalty = override.FrequencyPenalty
	}
	if override.PresencePenalty != nil {
		result.PresencePenalty = override.PresencePenalty
	}
	if len(override.Stop) > 0 {
		result.Stop = override.Stop
	}
	if override.ParallelToolCalls != nil {
		result.ParallelToolCalls = override.ParallelToolCalls
	}
	if override.ResponseFormat != nil {
		result.ResponseFormat = override.ResponseFormat
	}

	return result
}
