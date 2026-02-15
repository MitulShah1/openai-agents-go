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

func TestApplySettings_StopSingle(t *testing.T) {
	params := openai.ChatCompletionNewParams{}
	settings := ModelSettings{Stop: []string{"END"}}

	applySettings(&params, settings)

	if params.Stop.OfString.Value != "END" {
		t.Errorf("expected Stop OfString END, got %v", params.Stop.OfString.Value)
	}
	if params.Stop.OfStringArray != nil {
		t.Errorf("expected OfStringArray to be nil for single stop, got %v", params.Stop.OfStringArray)
	}
}

func TestApplySettings_StopMultiple(t *testing.T) {
	params := openai.ChatCompletionNewParams{}
	settings := ModelSettings{Stop: []string{"END", "STOP", "DONE"}}

	applySettings(&params, settings)

	if params.Stop.OfStringArray == nil {
		t.Fatal("expected OfStringArray to be set for multiple stop sequences")
	}
	if len(params.Stop.OfStringArray) != 3 {
		t.Fatalf("expected 3 stop sequences, got %d", len(params.Stop.OfStringArray))
	}
	expected := []string{"END", "STOP", "DONE"}
	for i, v := range params.Stop.OfStringArray {
		if v != expected[i] {
			t.Errorf("stop[%d]: expected %q, got %q", i, expected[i], v)
		}
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

func TestApplySettings_ModelNameOverride(t *testing.T) {
	params := openai.ChatCompletionNewParams{
		Model: openai.ChatModel("openai/gpt-4o"),
	}
	settings := ModelSettings{}
	applySettings(&params, settings)

	// applySettings should NOT change the model
	if params.Model != openai.ChatModel("openai/gpt-4o") {
		t.Errorf("applySettings should not change model, got %v", params.Model)
	}
}

func TestOpenAIChatCompletionsModel_ModelName(t *testing.T) {
	client := openai.NewClient()
	model := NewOpenAIChatCompletionsModel(&client, "gpt-4o")
	if model.ModelName() != "gpt-4o" {
		t.Errorf("expected gpt-4o, got %s", model.ModelName())
	}
}
