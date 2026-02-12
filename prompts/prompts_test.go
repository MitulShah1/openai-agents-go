package prompts

import (
	"fmt"
	"testing"
)

func TestPrompt_Fields(t *testing.T) {
	p := &Prompt{
		ID:      "prompt_abc",
		Version: "v2",
		Variables: map[string]any{
			"tone": "friendly",
			"lang": "en",
		},
	}

	if p.ID != "prompt_abc" {
		t.Errorf("expected ID prompt_abc, got %s", p.ID)
	}
	if p.Version != "v2" {
		t.Errorf("expected Version v2, got %s", p.Version)
	}
	if p.Variables["tone"] != "friendly" {
		t.Errorf("expected tone=friendly, got %v", p.Variables["tone"])
	}
}

func TestPrompt_EmptyVersion(t *testing.T) {
	p := &Prompt{ID: "prompt_latest"}

	if p.Version != "" {
		t.Errorf("expected empty Version, got %s", p.Version)
	}
}

func TestDynamicPromptFunc_Basic(t *testing.T) {
	fn := DynamicPromptFunc(func(data DynamicPromptData) (*Prompt, error) {
		return &Prompt{
			ID:      "prompt_" + data.Agent.Name,
			Version: "v1",
		}, nil
	})

	result, err := fn(DynamicPromptData{
		Agent: AgentInfo{Name: "helper", Model: "gpt-4o"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "prompt_helper" {
		t.Errorf("expected prompt_helper, got %s", result.ID)
	}
}

func TestDynamicPromptFunc_UsesContextVars(t *testing.T) {
	fn := DynamicPromptFunc(func(data DynamicPromptData) (*Prompt, error) {
		tier, _ := data.ContextVariables["tier"].(string)
		if tier == "premium" {
			return &Prompt{ID: "prompt_premium"}, nil
		}
		return &Prompt{ID: "prompt_default"}, nil
	})

	// Premium tier
	result, err := fn(DynamicPromptData{
		Agent:            AgentInfo{Name: "bot", Model: "gpt-4o"},
		ContextVariables: map[string]any{"tier": "premium"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "prompt_premium" {
		t.Errorf("expected prompt_premium, got %s", result.ID)
	}

	// Default tier
	result, err = fn(DynamicPromptData{
		Agent:            AgentInfo{Name: "bot", Model: "gpt-4o"},
		ContextVariables: map[string]any{"tier": "free"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "prompt_default" {
		t.Errorf("expected prompt_default, got %s", result.ID)
	}
}

func TestDynamicPromptFunc_ReturnsError(t *testing.T) {
	fn := DynamicPromptFunc(func(_ DynamicPromptData) (*Prompt, error) {
		return nil, fmt.Errorf("prompt service unavailable")
	})

	_, err := fn(DynamicPromptData{Agent: AgentInfo{Name: "test"}})
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "prompt service unavailable" {
		t.Errorf("expected 'prompt service unavailable', got %q", err.Error())
	}
}

func TestDynamicPromptData_AgentInfo(t *testing.T) {
	data := DynamicPromptData{
		Agent: AgentInfo{
			Name:  "Sales Bot",
			Model: "gpt-4o-mini",
		},
		ContextVariables: map[string]any{"user_id": "123"},
	}

	if data.Agent.Name != "Sales Bot" {
		t.Errorf("expected Sales Bot, got %s", data.Agent.Name)
	}
	if data.Agent.Model != "gpt-4o-mini" {
		t.Errorf("expected gpt-4o-mini, got %s", data.Agent.Model)
	}
	if data.ContextVariables["user_id"] != "123" {
		t.Errorf("expected user_id=123, got %v", data.ContextVariables["user_id"])
	}
}
