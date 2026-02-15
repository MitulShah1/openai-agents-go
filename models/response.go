package models

import "github.com/openai/openai-go/v3"

// ModelResponse is the normalized response from any model provider.
// It wraps the provider-specific completion data in a uniform type
// that the runner can consume regardless of backend.
type ModelResponse struct {
	// Completion is the underlying OpenAI chat completion response.
	// For non-OpenAI providers, this field is populated by converting
	// the provider's response into the OpenAI completion format.
	Completion *openai.ChatCompletion

	// Usage tracks token consumption for this response.
	Usage ModelUsage

	// ResponseID is an optional provider-specific response identifier.
	// For OpenAI's Responses API this is the response ID; for Chat
	// Completions it is the completion ID. May be empty for other providers.
	ResponseID string
}

// ModelUsage tracks token consumption for a single model call.
type ModelUsage struct {
	// PromptTokens is the number of tokens in the prompt.
	PromptTokens int

	// CompletionTokens is the number of tokens in the completion.
	CompletionTokens int

	// TotalTokens is the sum of prompt and completion tokens.
	TotalTokens int
}
