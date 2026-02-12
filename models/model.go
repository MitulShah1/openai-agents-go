package models

import (
	"context"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/packages/ssestream"

	"github.com/MitulShah1/openai-agents-go/tools"
)

// Model is the core interface for interacting with any LLM.
//
// Implementations must support both synchronous and streaming responses.
// The runner calls GetResponse for non-streaming execution and
// StreamResponse for streaming execution.
//
// Built-in implementations:
//   - [OpenAIChatCompletionsModel]: Uses OpenAI's Chat Completions API
//
// Custom implementations can wrap any LLM provider (Anthropic, Google, etc.)
// by converting requests/responses to and from the normalized types.
type Model interface {
	// GetResponse sends a chat completion request and returns the full response.
	//
	// Parameters:
	//   - ctx: Context for cancellation and deadline propagation
	//   - params: The fully-constructed chat completion request parameters
	//   - settings: Model-level settings (temperature, max tokens, etc.)
	//
	// The params argument already contains messages, tools, and system instructions
	// as prepared by the runner. Implementations should apply settings on top and
	// forward the request to the underlying LLM.
	GetResponse(
		ctx context.Context,
		params openai.ChatCompletionNewParams,
		settings ModelSettings,
	) (*ModelResponse, error)

	// StreamResponse sends a chat completion request and returns a streaming response.
	//
	// The returned stream follows the same contract as openai-go's streaming API:
	// callers iterate with Next()/Current()/Err().
	//
	// Parameters:
	//   - ctx: Context for cancellation and deadline propagation
	//   - params: The fully-constructed chat completion request parameters
	//   - settings: Model-level settings (temperature, max tokens, etc.)
	StreamResponse(
		ctx context.Context,
		params openai.ChatCompletionNewParams,
		settings ModelSettings,
	) (*ssestream.Stream[openai.ChatCompletionChunk], error)

	// ModelName returns the identifier of the model (e.g., "gpt-4o").
	ModelName() string
}

// ModelRequest contains all parameters needed to make a model call.
// This is used internally by the runner to pass context to the model.
type ModelRequest struct {
	// SystemInstructions is the system prompt for the model.
	SystemInstructions string

	// Messages is the conversation history.
	Messages []openai.ChatCompletionMessageParamUnion

	// Tools is the list of tools available for the model to call.
	Tools []tools.Tool

	// Settings contains model configuration (temperature, max tokens, etc.).
	Settings ModelSettings

	// ModelName is the name/ID of the model to use.
	ModelName string
}
