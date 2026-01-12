package runner

import (
	"context"
	"fmt"

	"github.com/openai/openai-go"
)

// RequestConfig holds configuration for preparing requests
type RequestConfig struct {
	Model              string
	Temperature        *float64
	MaxTokens          *int
	ParallelToolCalls  *bool
	ResponseFormat     any // Will be *jsonschema.ResponseFormat
	SystemInstructions string
}

// PrepareRequest builds an OpenAI chat completion request with the given configuration
func PrepareRequest(
	_ context.Context,
	config *RequestConfig,
	tools []openai.ChatCompletionToolParam,
	history []openai.ChatCompletionMessageParamUnion,
	responseFormatConverter func(any) (openai.ChatCompletionNewParamsResponseFormatUnion, error),
) (openai.ChatCompletionNewParams, error) {
	req := openai.ChatCompletionNewParams{
		Model: openai.ChatModel(config.Model),
	}

	// Apply model settings
	if config.Temperature != nil {
		req.Temperature = openai.Float(*config.Temperature)
	}

	if config.MaxTokens != nil {
		req.MaxTokens = openai.Int(int64(*config.MaxTokens))
	}

	// Apply tools configuration
	if len(tools) > 0 {
		req.Tools = tools
		if config.ParallelToolCalls != nil && !*config.ParallelToolCalls {
			req.ParallelToolCalls = openai.Bool(false)
		}
	}

	// Apply response format if provided
	if config.ResponseFormat != nil && responseFormatConverter != nil {
		format, err := responseFormatConverter(config.ResponseFormat)
		if err != nil {
			return req, fmt.Errorf("invalid response format: %w", err)
		}
		req.ResponseFormat = format
	}

	// Inject system instructions and build messages
	messagesForTurn := make([]openai.ChatCompletionMessageParamUnion, 0, len(history)+1)
	messagesForTurn = append(messagesForTurn, openai.SystemMessage(config.SystemInstructions))
	messagesForTurn = append(messagesForTurn, history...)
	req.Messages = messagesForTurn

	return req, nil
}

// ConfigMerger handles merging agent config with run config, with run config taking precedence
type ConfigMerger struct {
	AgentTemperature       *float64
	AgentMaxTokens         *int
	AgentParallelToolCalls bool
	AgentResponseFormat    any

	RunTemperature       *float64
	RunMaxTokens         *int
	RunParallelToolCalls *bool
	RunResponseFormat    any
}

// GetTemperature returns the effective temperature setting
func (cm *ConfigMerger) GetTemperature() *float64 {
	if cm.RunTemperature != nil {
		return cm.RunTemperature
	}
	return cm.AgentTemperature
}

// GetMaxTokens returns the effective max tokens setting
func (cm *ConfigMerger) GetMaxTokens() *int {
	if cm.RunMaxTokens != nil {
		return cm.RunMaxTokens
	}
	return cm.AgentMaxTokens
}

// GetParallelToolCalls returns the effective parallel tool calls setting
func (cm *ConfigMerger) GetParallelToolCalls() bool {
	if cm.RunParallelToolCalls != nil {
		return *cm.RunParallelToolCalls
	}
	return cm.AgentParallelToolCalls
}

// GetResponseFormat returns the effective response format setting
func (cm *ConfigMerger) GetResponseFormat() any {
	if cm.RunResponseFormat != nil {
		return cm.RunResponseFormat
	}
	return cm.AgentResponseFormat
}
