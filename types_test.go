package agents

import (
	"testing"
	"time"
)

func TestUsageAdd(t *testing.T) {
	usage1 := Usage{
		PromptTokens:     100,
		CompletionTokens: 50,
		TotalTokens:      150,
	}

	usage2 := Usage{
		PromptTokens:     200,
		CompletionTokens: 75,
		TotalTokens:      275,
	}

	usage1.Add(usage2)

	if usage1.PromptTokens != 300 {
		t.Errorf("expected PromptTokens=300, got %d", usage1.PromptTokens)
	}

	if usage1.CompletionTokens != 125 {
		t.Errorf("expected CompletionTokens=125, got %d", usage1.CompletionTokens)
	}

	if usage1.TotalTokens != 425 {
		t.Errorf("expected TotalTokens=425, got %d", usage1.TotalTokens)
	}
}

const testAgentName = "TestAgent"

func TestStepCreation(t *testing.T) {
	step := Step{
		AgentName:  testAgentName,
		StepNumber: 1,
		Duration:   100 * time.Millisecond,
		ToolCalls:  []ToolCall{},
	}

	if step.AgentName != testAgentName {
		t.Errorf("expected AgentName=%s, got %s", testAgentName, step.AgentName)
	}

	if step.StepNumber != 1 {
		t.Errorf("expected StepNumber=1, got %d", step.StepNumber)
	}

	if step.Duration != 100*time.Millisecond {
		t.Errorf("expected Duration=100ms, got %v", step.Duration)
	}
}

func TestToolCallCreation(t *testing.T) {
	duration := 50 * time.Millisecond
	toolCall := ToolCall{
		ToolName:  "get_weather",
		Arguments: `{"location": "Paris"}`,
		Result:    "Sunny, 25°C",
		Error:     nil,
		Duration:  duration,
	}

	if toolCall.ToolName != "get_weather" {
		t.Errorf("expected ToolName=get_weather, got %s", toolCall.ToolName)
	}

	if toolCall.Arguments != `{"location": "Paris"}` {
		t.Errorf("unexpected Arguments: %s", toolCall.Arguments)
	}

	if toolCall.Result != "Sunny, 25°C" {
		t.Errorf("unexpected Result: %v", toolCall.Result)
	}

	if toolCall.Error != nil {
		t.Error("expected Error=nil")
	}

	if toolCall.Duration != duration {
		t.Errorf("expected Duration=%v, got %v", duration, toolCall.Duration)
	}
}

func TestContextVariables(t *testing.T) {
	ctx := ContextVariables{
		"user_id": "123",
		"session": "abc",
	}

	if ctx["user_id"] != "123" {
		t.Errorf("expected user_id=123, got %v", ctx["user_id"])
	}

	if ctx["session"] != "abc" {
		t.Errorf("expected session=abc, got %v", ctx["session"])
	}

	// Test modification
	ctx["new_key"] = "new_value"
	if ctx["new_key"] != "new_value" {
		t.Error("failed to add new key")
	}
}
