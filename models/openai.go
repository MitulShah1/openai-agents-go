package models

import (
	"context"
	"fmt"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/packages/ssestream"
)

// OpenAIChatCompletionsModel implements the Model interface using
// OpenAI's Chat Completions API.
type OpenAIChatCompletionsModel struct {
	client    *openai.Client
	modelName string
}

// NewOpenAIChatCompletionsModel creates a new model backed by the given client.
func NewOpenAIChatCompletionsModel(client *openai.Client, modelName string) *OpenAIChatCompletionsModel {
	return &OpenAIChatCompletionsModel{
		client:    client,
		modelName: modelName,
	}
}

// ModelName returns the model identifier.
func (m *OpenAIChatCompletionsModel) ModelName() string {
	return m.modelName
}

// GetResponse sends a chat completion request and returns the normalized response.
func (m *OpenAIChatCompletionsModel) GetResponse(
	ctx context.Context,
	params openai.ChatCompletionNewParams,
	settings ModelSettings,
) (*ModelResponse, error) {
	// Override model name to strip any provider prefix (e.g., "openai/gpt-4o" → "gpt-4o")
	params.Model = openai.ChatModel(m.modelName)

	// Apply model settings to the request
	applySettings(&params, settings)

	completion, err := m.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("openai chat completion failed: %w", err)
	}

	return &ModelResponse{
		Completion: completion,
		Usage: ModelUsage{
			PromptTokens:     int(completion.Usage.PromptTokens),
			CompletionTokens: int(completion.Usage.CompletionTokens),
			TotalTokens:      int(completion.Usage.TotalTokens),
		},
		ResponseID: completion.ID,
	}, nil
}

// StreamResponse sends a streaming chat completion request.
func (m *OpenAIChatCompletionsModel) StreamResponse(
	ctx context.Context,
	params openai.ChatCompletionNewParams,
	settings ModelSettings,
) (*ssestream.Stream[openai.ChatCompletionChunk], error) {
	// Override model name to strip any provider prefix (e.g., "openai/gpt-4o" → "gpt-4o")
	params.Model = openai.ChatModel(m.modelName)

	// Apply model settings to the request
	applySettings(&params, settings)

	stream := m.client.Chat.Completions.NewStreaming(ctx, params)
	return stream, nil
}

// applySettings applies ModelSettings to the request parameters.
// Settings here serve as additional overrides; the params already contain
// the base configuration from the runner's request preparation.
func applySettings(params *openai.ChatCompletionNewParams, settings ModelSettings) {
	if settings.Temperature != nil {
		params.Temperature = openai.Float(*settings.Temperature)
	}
	if settings.MaxTokens != nil {
		params.MaxTokens = openai.Int(int64(*settings.MaxTokens))
	}
	if settings.TopP != nil {
		params.TopP = openai.Float(*settings.TopP)
	}
	if settings.FrequencyPenalty != nil {
		params.FrequencyPenalty = openai.Float(*settings.FrequencyPenalty)
	}
	if settings.PresencePenalty != nil {
		params.PresencePenalty = openai.Float(*settings.PresencePenalty)
	}
	if len(settings.Stop) == 1 {
		params.Stop = openai.ChatCompletionNewParamsStopUnion{
			OfString: openai.String(settings.Stop[0]),
		}
	} else if len(settings.Stop) > 1 {
		params.Stop = openai.ChatCompletionNewParamsStopUnion{
			OfStringArray: settings.Stop,
		}
	}
}

// OpenAIProvider implements ModelProvider for OpenAI's API.
// It creates OpenAIChatCompletionsModel instances for any model name.
type OpenAIProvider struct {
	client *openai.Client
}

// NewOpenAIProvider creates a provider backed by the given OpenAI client.
//
// Example:
//
//	client := openai.NewClient(option.WithAPIKey("sk-..."))
//	provider := models.NewOpenAIProvider(&client)
func NewOpenAIProvider(client *openai.Client) *OpenAIProvider {
	return &OpenAIProvider{client: client}
}

// GetModel returns an OpenAIChatCompletionsModel for the given model name.
func (p *OpenAIProvider) GetModel(modelName string) (Model, error) {
	if modelName == "" {
		return nil, fmt.Errorf("models: model name cannot be empty")
	}
	return NewOpenAIChatCompletionsModel(p.client, modelName), nil
}

// Client returns the underlying OpenAI client.
// This is useful for backward compatibility when code needs direct API access.
func (p *OpenAIProvider) Client() *openai.Client {
	return p.client
}
