package agents

import (
	"context"
	"testing"

	"github.com/MitulShah1/openai-agents-go/guardrail"
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
