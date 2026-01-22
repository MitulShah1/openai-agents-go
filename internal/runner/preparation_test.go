package runner

import (
	"context"
	"testing"

	"github.com/openai/openai-go/v3"

	"github.com/MitulShah1/openai-agents-go/jsonschema"
)

func TestPrepareRequest(t *testing.T) {
	// Setup checks
	ctx := context.Background()

	t.Run("basic configuration", func(t *testing.T) {
		temp := 0.7
		maxTokens := 100
		parallel := true

		config := &RequestConfig{
			Model:              openai.ChatModelGPT4o,
			Temperature:        &temp,
			MaxTokens:          &maxTokens,
			ParallelToolCalls:  &parallel,
			SystemInstructions: "You are a bot.",
		}

		params, err := PrepareRequest(ctx, config, nil, nil, nil)
		if err != nil {
			t.Fatalf("PrepareRequest failed: %v", err)
		}

		if params.Model != openai.ChatModelGPT4o {
			t.Errorf("expected model gpt-4o, got %v", params.Model)
		}
		if params.Temperature.Value != temp {
			t.Errorf("expected temp %v, got %v", temp, params.Temperature.Value)
		}
		if params.MaxTokens.Value != int64(maxTokens) {
			t.Errorf("expected max tokens %v, got %v", maxTokens, params.MaxTokens.Value)
		}
	})

	t.Run("with tools and parallel disabled", func(t *testing.T) {
		parallel := false
		config := &RequestConfig{
			Model:             openai.ChatModelGPT4o,
			ParallelToolCalls: &parallel,
		}

		// Tool Name is plain string in v3
		toolParam := openai.ChatCompletionFunctionTool(openai.FunctionDefinitionParam{
			Name: "test_tool",
		})
		tools := []openai.ChatCompletionToolUnionParam{toolParam}

		params, err := PrepareRequest(ctx, config, tools, nil, nil)
		if err != nil {
			t.Fatalf("PrepareRequest failed: %v", err)
		}

		if len(params.Tools) != 1 {
			t.Errorf("expected 1 tool, got %d", len(params.Tools))
		}

		if params.ParallelToolCalls.Value != false {
			t.Errorf("expected parallel tool calls false, got %v", params.ParallelToolCalls.Value)
		}
	})

	t.Run("message history concatenation", func(t *testing.T) {
		config := &RequestConfig{
			Model:              openai.ChatModelGPT4o,
			SystemInstructions: "System prompt",
		}

		history := []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("Hello"),
			openai.AssistantMessage("Hi"),
		}

		params, err := PrepareRequest(ctx, config, nil, history, nil)
		if err != nil {
			t.Fatalf("PrepareRequest failed: %v", err)
		}

		if len(params.Messages) != 3 {
			t.Fatalf("expected 3 messages, got %d", len(params.Messages))
		}
	})

	t.Run("response format", func(t *testing.T) {
		config := &RequestConfig{
			Model:          openai.ChatModelGPT4o,
			ResponseFormat: &jsonschema.ResponseFormat{Type: "json_object"},
		}

		// converter mock returning OfText variant using string literal for type
		converter := func(_ any) (openai.ChatCompletionNewParamsResponseFormatUnion, error) {
			return openai.ChatCompletionNewParamsResponseFormatUnion{
				OfText: &openai.ResponseFormatTextParam{
					Type: "text",
				},
			}, nil
		}

		params, err := PrepareRequest(ctx, config, nil, nil, converter)
		if err != nil {
			t.Fatalf("PrepareRequest failed: %v", err)
		}

		// Verify it was set
		if params.ResponseFormat.OfText == nil {
			t.Error("expected response format to be set")
		}
	})
}

func TestConfigMerger(t *testing.T) {
	agentTemp := 0.5
	runTemp := 1.0

	t.Run("run config overrides agent", func(t *testing.T) {
		merger := &ConfigMerger{
			AgentTemperature: &agentTemp,
			RunTemperature:   &runTemp,
		}

		if got := merger.GetTemperature(); got == nil || *got != runTemp {
			t.Errorf("expected run temp %v, got %v", runTemp, got)
		}
	})

	t.Run("fallback to agent config", func(t *testing.T) {
		merger := &ConfigMerger{
			AgentTemperature: &agentTemp,
			RunTemperature:   nil,
		}

		if got := merger.GetTemperature(); got == nil || *got != agentTemp {
			t.Errorf("expected agent temp %v, got %v", agentTemp, got)
		}
	})

	t.Run("response format priority", func(t *testing.T) {
		agentFormat := &jsonschema.ResponseFormat{Type: "text"}
		runFormat := &jsonschema.ResponseFormat{Type: "json_object"}

		merger := &ConfigMerger{
			AgentResponseFormat: agentFormat,
			RunResponseFormat:   runFormat,
		}

		got := merger.GetResponseFormat()
		if got != runFormat {
			t.Errorf("expected run format, got %v", got)
		}
	})

	t.Run("response format fallback", func(t *testing.T) {
		agentFormat := &jsonschema.ResponseFormat{Type: "text"}

		merger := &ConfigMerger{
			AgentResponseFormat: agentFormat,
			RunResponseFormat:   nil,
		}

		got := merger.GetResponseFormat()
		if got != agentFormat {
			t.Errorf("expected agent format, got %v", got)
		}
	})
}
