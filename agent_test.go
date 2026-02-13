package agents

import (
	"context"
	"fmt"
	"testing"

	"github.com/openai/openai-go/v3"

	"github.com/MitulShah1/openai-agents-go/guardrail"
	"github.com/MitulShah1/openai-agents-go/models"
	"github.com/MitulShah1/openai-agents-go/prompts"
	"github.com/MitulShah1/openai-agents-go/tools"
)

func TestNewAgent(t *testing.T) {
	agent := NewAgent("TestAgent")

	if agent.Name != "TestAgent" {
		t.Errorf("expected Name=TestAgent, got %s", agent.Name)
	}

	if agent.Model != DefaultModel {
		t.Errorf("expected Model=%s, got %s", DefaultModel, agent.Model)
	}

	if agent.Instructions != DefaultInstructions {
		t.Errorf("expected Instructions=%s, got %v", DefaultInstructions, agent.Instructions)
	}

	if !agent.ParallelToolCalls {
		t.Error("expected ParallelToolCalls=true")
	}

	if agent.Temperature != nil {
		t.Error("expected Temperature=nil")
	}

	if agent.MaxTokens != nil {
		t.Error("expected MaxTokens=nil")
	}
}

func TestGetInstructions(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name         string
		instructions any
		expected     string
	}{
		{
			name:         "string instructions",
			instructions: "You are a helpful bot",
			expected:     "You are a helpful bot",
		},
		{
			name:         "function without context",
			instructions: func() string { return "Dynamic instructions" },
			expected:     "Dynamic instructions",
		},
		{
			name: "function with context",
			instructions: func(_ context.Context) string {
				return "Context-aware instructions"
			},
			expected: "Context-aware instructions",
		},
		{
			name:         "invalid type defaults to default",
			instructions: 123,
			expected:     DefaultInstructions,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := &Agent{Instructions: tt.instructions}
			result := agent.GetInstructions(ctx)

			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestLifecycleHooks(t *testing.T) {
	ctx := context.Background()
	agent := NewAgent("TestAgent")

	// Test OnBeforeRun
	beforeCalled := false
	agent.OnBeforeRun = func(_ context.Context, a *Agent) error {
		beforeCalled = true
		if a.Name != "TestAgent" {
			t.Error("wrong agent passed to OnBeforeRun")
		}
		return nil
	}

	// Test OnAfterRun
	afterCalled := false
	agent.OnAfterRun = func(_ context.Context, _ *Agent) error {
		afterCalled = true
		return nil
	}

	// Manually invoke hooks to test
	if err := agent.OnBeforeRun(ctx, agent); err != nil {
		t.Errorf("OnBeforeRun failed: %v", err)
	}

	if !beforeCalled {
		t.Error("OnBeforeRun was not called")
	}

	if err := agent.OnAfterRun(ctx, agent); err != nil {
		t.Errorf("OnAfterRun failed: %v", err)
	}

	if !afterCalled {
		t.Error("OnAfterRun was not called")
	}
}

func TestGetPrompt_Nil(t *testing.T) {
	agent := NewAgent("test")
	// agent.Prompt is nil by default

	result, err := agent.GetPrompt(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Error("expected nil prompt when not set")
	}
}

func TestGetPrompt_StaticPrompt(t *testing.T) {
	agent := NewAgent("test")
	agent.Prompt = &prompts.Prompt{
		ID:      "prompt_helpful",
		Version: "v2",
		Variables: map[string]any{
			"tone": "friendly",
		},
	}

	result, err := agent.GetPrompt(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil prompt")
	}
	if result.ID != "prompt_helpful" {
		t.Errorf("expected ID prompt_helpful, got %s", result.ID)
	}
	if result.Version != "v2" {
		t.Errorf("expected Version v2, got %s", result.Version)
	}
	if result.Variables["tone"] != "friendly" {
		t.Errorf("expected tone=friendly, got %v", result.Variables["tone"])
	}
}

func TestGetPrompt_DynamicPrompt(t *testing.T) {
	agent := NewAgent("PremiumBot")
	agent.Model = "gpt-4o"
	agent.Prompt = prompts.DynamicPromptFunc(func(data prompts.DynamicPromptData) (*prompts.Prompt, error) {
		return &prompts.Prompt{
			ID: "prompt_" + data.Agent.Name,
			Variables: map[string]any{
				"model": data.Agent.Model,
			},
		}, nil
	})

	result, err := agent.GetPrompt(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "prompt_PremiumBot" {
		t.Errorf("expected prompt_PremiumBot, got %s", result.ID)
	}
	if result.Variables["model"] != "gpt-4o" {
		t.Errorf("expected model=gpt-4o, got %v", result.Variables["model"])
	}
}

func TestGetPrompt_DynamicWithContextVars(t *testing.T) {
	agent := NewAgent("test")
	agent.Prompt = prompts.DynamicPromptFunc(func(data prompts.DynamicPromptData) (*prompts.Prompt, error) {
		tier, _ := data.ContextVariables["tier"].(string)
		if tier == "premium" {
			return &prompts.Prompt{ID: "prompt_premium"}, nil
		}
		return &prompts.Prompt{ID: "prompt_free"}, nil
	})

	// Premium
	result, err := agent.GetPrompt(map[string]any{"tier": "premium"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "prompt_premium" {
		t.Errorf("expected prompt_premium, got %s", result.ID)
	}

	// Free
	result, err = agent.GetPrompt(map[string]any{"tier": "free"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "prompt_free" {
		t.Errorf("expected prompt_free, got %s", result.ID)
	}
}

func TestGetPrompt_DynamicError(t *testing.T) {
	agent := NewAgent("test")
	agent.Prompt = prompts.DynamicPromptFunc(func(_ prompts.DynamicPromptData) (*prompts.Prompt, error) {
		return nil, fmt.Errorf("prompt service down")
	})

	_, err := agent.GetPrompt(nil)
	if err == nil {
		t.Fatal("expected error from failing dynamic prompt")
	}
}

func TestGetPrompt_InvalidType(t *testing.T) {
	agent := NewAgent("test")
	agent.Prompt = "not a prompt" // wrong type

	_, err := agent.GetPrompt(nil)
	if err == nil {
		t.Fatal("expected error for unsupported Prompt type")
	}
}

func TestGetPrompt_IntegrationWithRunner_Static(t *testing.T) {
	// Verify the runner resolves the prompt and passes it to the model
	mockResp := &models.ModelResponse{
		Completion: &openai.ChatCompletion{
			Choices: []openai.ChatCompletionChoice{
				{Message: openai.ChatCompletionMessage{Role: "assistant", Content: "prompted!"}},
			},
		},
		Usage: models.ModelUsage{TotalTokens: 5},
	}

	mock := &testModel{name: "test", response: mockResp}
	r := NewRunnerWithProvider(&testProvider{model: mock})

	agent := NewAgent("test")
	agent.Prompt = &prompts.Prompt{ID: "prompt_test", Version: "v1"}

	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Hello"),
	}

	result, err := r.Run(context.Background(), agent, messages)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.FinalOutput != "prompted!" {
		t.Errorf("expected 'prompted!', got %q", result.FinalOutput)
	}

	// Verify the prompt was actually passed to the model via ModelSettings
	if mock.lastSettings.Prompt == nil {
		t.Fatal("expected prompt to be passed to model, got nil")
	}
	if mock.lastSettings.Prompt.ID != "prompt_test" {
		t.Errorf("expected prompt ID 'prompt_test', got %q", mock.lastSettings.Prompt.ID)
	}
	if mock.lastSettings.Prompt.Version != "v1" {
		t.Errorf("expected prompt Version 'v1', got %q", mock.lastSettings.Prompt.Version)
	}
}

func TestGetPrompt_IntegrationWithRunner_Dynamic(t *testing.T) {
	// Verify dynamic prompts receive context variables from the runner
	mockResp := &models.ModelResponse{
		Completion: &openai.ChatCompletion{
			Choices: []openai.ChatCompletionChoice{
				{Message: openai.ChatCompletionMessage{Role: "assistant", Content: "ok"}},
			},
		},
		Usage: models.ModelUsage{},
	}

	mock := &testModel{name: "test", response: mockResp}
	r := NewRunnerWithProvider(&testProvider{model: mock})

	agent := NewAgent("test")
	agent.Prompt = prompts.DynamicPromptFunc(func(data prompts.DynamicPromptData) (*prompts.Prompt, error) {
		tier, _ := data.ContextVariables["tier"].(string)
		return &prompts.Prompt{ID: "prompt_" + tier}, nil
	})

	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Hello"),
	}

	_, err := r.Run(context.Background(), agent, messages,
		WithContextVariables(ContextVariables{"tier": "premium"}),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the dynamic prompt resolved with context variables
	if mock.lastSettings.Prompt == nil {
		t.Fatal("expected prompt to be passed to model, got nil")
	}
	if mock.lastSettings.Prompt.ID != "prompt_premium" {
		t.Errorf("expected prompt ID 'prompt_premium', got %q", mock.lastSettings.Prompt.ID)
	}
}

func TestGetPrompt_IntegrationWithRunner_NoPrompt(t *testing.T) {
	// Verify no prompt is passed when agent has no prompt configured
	mockResp := &models.ModelResponse{
		Completion: &openai.ChatCompletion{
			Choices: []openai.ChatCompletionChoice{
				{Message: openai.ChatCompletionMessage{Role: "assistant", Content: "ok"}},
			},
		},
		Usage: models.ModelUsage{},
	}

	mock := &testModel{name: "test", response: mockResp}
	r := NewRunnerWithProvider(&testProvider{model: mock})

	agent := NewAgent("test")
	// agent.Prompt is nil

	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Hello"),
	}

	_, err := r.Run(context.Background(), agent, messages)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mock.lastSettings.Prompt != nil {
		t.Errorf("expected nil prompt when not configured, got %+v", mock.lastSettings.Prompt)
	}
}

func TestGetPrompt_StreamingPromptErrorPropagates(t *testing.T) {
	// Verify that if GetPrompt fails, the streaming runner propagates the error
	mock := &testModel{name: "test"}
	r := NewRunnerWithProvider(&testProvider{model: mock})

	agent := NewAgent("test")
	agent.Prompt = prompts.DynamicPromptFunc(func(_ prompts.DynamicPromptData) (*prompts.Prompt, error) {
		return nil, fmt.Errorf("prompt service unavailable")
	})

	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Hello"),
	}

	// Test Stream() path
	ch, err := r.Stream(context.Background(), agent, messages)
	if err != nil {
		t.Fatalf("Stream() should not return error immediately: %v", err)
	}

	// Drain channel and look for the error event
	var gotPromptError bool
	for event := range ch {
		if event.Type == StreamEventError && event.Error != nil {
			if contains := fmt.Sprintf("%v", event.Error); len(contains) > 0 {
				gotPromptError = true
			}
		}
	}
	if !gotPromptError {
		t.Error("expected prompt resolution error to propagate through Stream()")
	}
}

func TestGetInstructions_WithSkills(t *testing.T) {
	ctx := context.Background()
	agent := NewAgent("SkillAgent")
	agent.Instructions = "Base instructions"
	agent.AddSkills(
		Skill{Instructions: "Skill one instructions"},
		Skill{Instructions: "Skill two instructions"},
	)

	got := agent.GetInstructions(ctx)
	want := "Base instructions\n\nSkill one instructions\n\nSkill two instructions"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestAddSkill_AppendsToolsAndGuardrails(t *testing.T) {
	agent := NewAgent("SkillAgent")

	tool := tools.New(
		"mock_tool",
		"mock",
		map[string]any{"type": "object"},
		func(_ map[string]any, _ tools.ContextVariables) (any, error) { return "ok", nil },
	)
	inGR := guardrail.NewGuardrail("in", func(_ context.Context, _ string) (*guardrail.Result, error) {
		return &guardrail.Result{Passed: true}, nil
	})
	outGR := guardrail.NewGuardrail("out", func(_ context.Context, _ string) (*guardrail.Result, error) {
		return &guardrail.Result{Passed: true}, nil
	})

	agent.AddSkill(Skill{
		Name:             "support",
		Description:      "Support workflow",
		Instructions:     "Use support rubric",
		Tools:            []tools.Tool{tool},
		InputGuardrails:  []*guardrail.Guardrail{inGR},
		OutputGuardrails: []*guardrail.Guardrail{outGR},
	})

	if len(agent.Skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(agent.Skills))
	}
	if len(agent.Tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(agent.Tools))
	}
	if len(agent.InputGuardrails) != 1 {
		t.Fatalf("expected 1 input guardrail, got %d", len(agent.InputGuardrails))
	}
	if len(agent.OutputGuardrails) != 1 {
		t.Fatalf("expected 1 output guardrail, got %d", len(agent.OutputGuardrails))
	}
}
