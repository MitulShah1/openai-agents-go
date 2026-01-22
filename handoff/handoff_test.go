package handoff

import (
	"context"
	"testing"

	"github.com/openai/openai-go/v3"

	agents "github.com/MitulShah1/openai-agents-go"
)

func TestNew(t *testing.T) {
	agent := agents.NewAgent("TestAgent")

	h := New(agent)

	if h.TargetAgent() != agent {
		t.Error("TargetAgent() should return the agent passed to New()")
	}

	if h.NestHistory() {
		t.Error("NestHistory() should be false by default")
	}

	if h.InputFilter() != nil {
		t.Error("InputFilter() should be nil by default")
	}

	if h.EnabledPredicate() != nil {
		t.Error("EnabledPredicate() should be nil by default")
	}
}

func TestNew_WithOptions(t *testing.T) {
	agent := agents.NewAgent("TestAgent")

	// Test with all options
	inputFilter := func(_ context.Context, data InputData) (InputData, error) {
		return data, nil
	}

	predicate := func(_ context.Context, _ *agents.Agent, _ agents.ContextVariables) (bool, error) {
		return true, nil
	}

	h := New(
		agent,
		WithToolName("custom_transfer"),
		WithDescription("Custom description"),
		WithInputFilter(inputFilter),
		WithHistoryNesting(true),
		WithEnabledPredicate(predicate),
	)

	if h.TargetAgent() != agent {
		t.Error("TargetAgent() should return the agent")
	}

	if !h.NestHistory() {
		t.Error("NestHistory() should be true")
	}

	if h.InputFilter() == nil {
		t.Error("InputFilter() should not be nil")
	}

	if h.EnabledPredicate() == nil {
		t.Error("EnabledPredicate() should not be nil")
	}
}

func TestToTool(t *testing.T) {
	agent := agents.NewAgent("SupportAgent")
	agent.Instructions = "You are a support agent"

	h := New(agent)
	tool := h.ToTool()

	// Check tool properties
	if tool.Name == "" {
		t.Error("Tool name should not be empty")
	}

	if tool.Description == "" {
		t.Error("Tool description should not be empty")
	}

	if tool.Callback == nil {
		t.Error("Tool callback should not be nil")
	}

	if !tool.IsHandoffTool {
		t.Error("IsHandoffTool should be true")
	}

	// Check generated tool name format
	expectedName := "transfer_to_supportagent"
	if tool.Name != expectedName {
		t.Errorf("Expected tool name %q, got %q", expectedName, tool.Name)
	}
}

func TestToTool_WithCustomName(t *testing.T) {
	agent := agents.NewAgent("SupportAgent")

	h := New(agent, WithToolName("escalate_to_support"))
	tool := h.ToTool()

	if tool.Name != "escalate_to_support" {
		t.Errorf("Expected custom tool name, got %q", tool.Name)
	}
}

func TestToTool_Callback(t *testing.T) {
	agent := agents.NewAgent("SupportAgent")

	h := New(agent)
	tool := h.ToTool()

	// Execute the callback
	contextVars := agents.ContextVariables{}
	result, err := tool.Callback(map[string]any{}, contextVars)

	if err != nil {
		t.Errorf("Callback should not return error: %v", err)
	}

	// Result should be the agent (handoff)
	resultAgent, ok := result.(*agents.Agent)
	if !ok {
		t.Error("Callback result should be *agents.Agent")
	}

	if resultAgent != agent {
		t.Error("Callback should return the target agent")
	}
}

func TestToTool_CallbackWithPredicate(t *testing.T) {
	agent := agents.NewAgent("SupportAgent")

	// Test with enabled predicate returning false
	h := New(agent, WithEnabledPredicate(func(_ context.Context, _ *agents.Agent, _ agents.ContextVariables) (bool, error) {
		return false, nil
	}))
	tool := h.ToTool()

	contextVars := agents.ContextVariables{}
	_, err := tool.Callback(map[string]any{}, contextVars)

	if err == nil {
		t.Error("Callback should return error when predicate returns false")
	}
}

func TestDefaultHistoryMapper(t *testing.T) {
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Hello"),
		openai.AssistantMessage("Hi there!"),
		openai.UserMessage("How are you?"),
	}

	result, err := DefaultHistoryMapper(context.Background(), messages)

	if err != nil {
		t.Errorf("DefaultHistoryMapper should not return error: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("Expected 1 summary message, got %d", len(result))
	}

	// Summary should be an assistant message
	// We can't easily inspect the result without marshaling, but we check it exists
}

func TestDefaultHistoryMapper_Empty(t *testing.T) {
	messages := []openai.ChatCompletionMessageParamUnion{}

	result, err := DefaultHistoryMapper(context.Background(), messages)

	if err != nil {
		t.Errorf("DefaultHistoryMapper should not return error: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Expected 0 messages, got %d", len(result))
	}
}

func TestFlattenHistoryMapper(t *testing.T) {
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Hello"),
		openai.AssistantMessage("Hi there!"),
	}

	result, err := FlattenHistoryMapper(context.Background(), messages)

	if err != nil {
		t.Errorf("FlattenHistoryMapper should not return error: %v", err)
	}

	if len(result) != len(messages) {
		t.Errorf("Expected %d messages, got %d", len(messages), len(result))
	}
}
