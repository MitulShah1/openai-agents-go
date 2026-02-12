package models

import (
	"testing"

	"github.com/openai/openai-go/v3"
)

func TestApplySettings_Temperature(t *testing.T) {
	params := openai.ChatCompletionNewParams{}
	settings := ModelSettings{Temperature: ptrFloat64(0.5)}

	applySettings(&params, settings)

	if params.Temperature.Value != 0.5 {
		t.Errorf("expected Temperature 0.5, got %v", params.Temperature)
	}
}

func TestApplySettings_MaxTokens(t *testing.T) {
	params := openai.ChatCompletionNewParams{}
	settings := ModelSettings{MaxTokens: ptrInt(200)}

	applySettings(&params, settings)

	if params.MaxTokens.Value != 200 {
		t.Errorf("expected MaxTokens 200, got %v", params.MaxTokens)
	}
}

func TestApplySettings_TopP(t *testing.T) {
	params := openai.ChatCompletionNewParams{}
	settings := ModelSettings{TopP: ptrFloat64(0.95)}

	applySettings(&params, settings)

	if params.TopP.Value != 0.95 {
		t.Errorf("expected TopP 0.95, got %v", params.TopP)
	}
}

func TestApplySettings_Penalties(t *testing.T) {
	params := openai.ChatCompletionNewParams{}
	settings := ModelSettings{
		FrequencyPenalty: ptrFloat64(0.3),
		PresencePenalty:  ptrFloat64(0.4),
	}

	applySettings(&params, settings)

	if params.FrequencyPenalty.Value != 0.3 {
		t.Errorf("expected FrequencyPenalty 0.3, got %v", params.FrequencyPenalty)
	}
	if params.PresencePenalty.Value != 0.4 {
		t.Errorf("expected PresencePenalty 0.4, got %v", params.PresencePenalty)
	}
}

func TestApplySettings_Stop(t *testing.T) {
	params := openai.ChatCompletionNewParams{}
	settings := ModelSettings{Stop: []string{"END"}}

	applySettings(&params, settings)

	if params.Stop.OfString.Value != "END" {
		t.Errorf("expected Stop END, got %v", params.Stop)
	}
}

func TestApplySettings_Empty(t *testing.T) {
	params := openai.ChatCompletionNewParams{
		Model: openai.ChatModelGPT4o,
	}
	settings := ModelSettings{} // all nil

	applySettings(&params, settings)

	// Model should be unchanged
	if params.Model != openai.ChatModelGPT4o {
		t.Errorf("expected model to remain gpt-4o, got %v", params.Model)
	}
}

func TestNewOpenAIChatCompletionsModel(t *testing.T) {
	client := openai.NewClient()
	model := NewOpenAIChatCompletionsModel(&client, "gpt-4o-mini")

	if model.ModelName() != "gpt-4o-mini" {
		t.Errorf("expected gpt-4o-mini, got %s", model.ModelName())
	}
}
