package handoff

import (
	"context"
	"strings"
	"testing"

	agents "github.com/MitulShah1/openai-agents-go"
	"github.com/MitulShah1/openai-agents-go/tools"
)

func TestWithToolName(t *testing.T) {
	agent := agents.NewAgent("TestAgent")
	customName := "my_custom_handoff"

	h := New(agent, WithToolName(customName))
	tool := h.ToTool()

	if tool.Name != customName {
		t.Errorf("Expected tool name %q, got %q", customName, tool.Name)
	}
}

func TestWithDescription(t *testing.T) {
	agent := agents.NewAgent("TestAgent")
	customDesc := "Custom description for this handoff"

	h := New(agent, WithDescription(customDesc))
	tool := h.ToTool()

	if tool.Description != customDesc {
		t.Errorf("Expected description %q, got %q", customDesc, tool.Description)
	}
}

func TestWithInputFilter_Success(t *testing.T) {
	agent := agents.NewAgent("TestAgent")

	filterCalled := false
	h := New(agent, WithInputFilter(func(_ context.Context, data InputData) (InputData, error) {
		filterCalled = true
		// Modify context vars
		data.ContextVars["filtered"] = true
		return data, nil
	}))

	tool := h.ToTool()
	contextVars := agents.ContextVariables{}

	_, err := tool.Callback(map[string]any{}, contextVars)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !filterCalled {
		t.Error("Expected filter to be called")
	}
}

func TestWithInputFilter_Error(t *testing.T) {
	agent := agents.NewAgent("TestAgent")

	h := New(agent, WithInputFilter(func(_ context.Context, data InputData) (InputData, error) {
		return data, context.Canceled
	}))

	tool := h.ToTool()
	contextVars := agents.ContextVariables{}

	_, err := tool.Callback(map[string]any{}, contextVars)

	if err == nil {
		t.Error("Expected error from filter, got nil")
	}

	if !strings.Contains(err.Error(), "input filter failed") {
		t.Errorf("Expected 'input filter failed' error, got %v", err)
	}
}

func TestWithHistoryNesting(t *testing.T) {
	agent := agents.NewAgent("TestAgent")

	tests := []struct {
		name     string
		enabled  bool
		expected bool
	}{
		{"enabled", true, true},
		{"disabled", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := New(agent, WithHistoryNesting(tt.enabled))

			if h.NestHistory() != tt.expected {
				t.Errorf("Expected NestHistory() = %v, got %v", tt.expected, h.NestHistory())
			}
		})
	}
}

func TestWithEnabledPredicate_Enabled(t *testing.T) {
	agent := agents.NewAgent("TestAgent")

	h := New(agent, WithEnabledPredicate(func(_ context.Context, _ *agents.Agent, _ agents.ContextVariables) (bool, error) {
		return true, nil
	}))

	tool := h.ToTool()
	contextVars := agents.ContextVariables{}

	result, err := tool.Callback(map[string]any{}, contextVars)

	if err != nil {
		t.Errorf("Expected no error when predicate returns true, got %v", err)
	}

	if result == nil {
		t.Error("Expected result when predicate returns true")
	}
}

func TestWithEnabledPredicate_Disabled(t *testing.T) {
	agent := agents.NewAgent("TestAgent")

	h := New(agent, WithEnabledPredicate(func(_ context.Context, _ *agents.Agent, _ agents.ContextVariables) (bool, error) {
		return false, nil
	}))

	tool := h.ToTool()
	contextVars := agents.ContextVariables{}

	_, err := tool.Callback(map[string]any{}, contextVars)

	if err == nil {
		t.Error("Expected error when predicate returns false")
	}

	if !strings.Contains(err.Error(), "disabled") {
		t.Errorf("Expected 'disabled' in error message, got %v", err)
	}
}

func TestWithEnabledPredicate_Error(t *testing.T) {
	agent := agents.NewAgent("TestAgent")

	h := New(agent, WithEnabledPredicate(func(_ context.Context, _ *agents.Agent, _ agents.ContextVariables) (bool, error) {
		return false, context.Canceled
	}))

	tool := h.ToTool()
	contextVars := agents.ContextVariables{}

	_, err := tool.Callback(map[string]any{}, contextVars)

	if err == nil {
		t.Error("Expected error when predicate returns error")
	}

	if !strings.Contains(err.Error(), "error checking handoff enablement") {
		t.Errorf("Expected enablement error message, got %v", err)
	}
}

func TestMultipleOptions(t *testing.T) {
	agent := agents.NewAgent("TestAgent")

	h := New(
		agent,
		WithToolName("custom_name"),
		WithDescription("custom description"),
		WithHistoryNesting(true),
		WithInputFilter(func(_ context.Context, data InputData) (InputData, error) {
			return data, nil
		}),
		WithEnabledPredicate(func(_ context.Context, _ *agents.Agent, _ agents.ContextVariables) (bool, error) {
			return true, nil
		}),
	)

	tool := h.ToTool()

	if tool.Name != "custom_name" {
		t.Errorf("Expected custom name, got %q", tool.Name)
	}

	if tool.Description != "custom description" {
		t.Errorf("Expected custom description, got %q", tool.Description)
	}

	if !h.NestHistory() {
		t.Error("Expected history nesting to be enabled")
	}

	if h.InputFilter() == nil {
		t.Error("Expected input filter to be set")
	}

	if h.EnabledPredicate() == nil {
		t.Error("Expected enabled predicate to be set")
	}
}

func TestGenerateToolName(t *testing.T) {
	tests := []struct {
		name         string
		agentName    string
		expectedName string
	}{
		{"simple name", "Support", "transfer_to_support"},
		{"name with spaces", "Technical Support", "transfer_to_technical_support"},
		{"mixed case", "BillingAgent", "transfer_to_billingagent"},
		{"single word", "AI", "transfer_to_ai"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := agents.NewAgent(tt.agentName)
			h := New(agent)
			tool := h.ToTool()

			if tool.Name != tt.expectedName {
				t.Errorf("Expected tool name %q, got %q", tt.expectedName, tool.Name)
			}
		})
	}
}

func TestGenerateDescription(t *testing.T) {
	tests := []struct {
		name          string
		agentName     string
		instructions  string
		checkContains []string
	}{
		{
			"with short instructions",
			"Support",
			"Help customers",
			[]string{"Transfer", "Support", "Help customers"},
		},
		{
			"with long instructions",
			"Billing",
			strings.Repeat("A very long instruction that should be truncated ", 10),
			[]string{"Transfer", "Billing", "..."},
		},
		{
			"without instructions",
			"Sales",
			"",
			[]string{"Transfer", "Sales"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := agents.NewAgent(tt.agentName)
			if tt.instructions != "" {
				agent.Instructions = tt.instructions
			}

			h := New(agent)
			tool := h.ToTool()

			for _, substr := range tt.checkContains {
				if !strings.Contains(tool.Description, substr) {
					t.Errorf("Expected description to contain %q, got %q", substr, tool.Description)
				}
			}
		})
	}
}

func TestCallback_WithSummary(t *testing.T) {
	agent := agents.NewAgent("TestAgent")
	h := New(agent)
	tool := h.ToTool()

	contextVars := agents.ContextVariables{}
	args := map[string]any{
		"summary": "This is a conversation summary",
	}

	_, err := tool.Callback(args, contextVars)

	if err != nil {
		t.Errorf("Callback should not error: %v", err)
	}

	// Summary should be stored in context vars
	if summary, ok := contextVars["_handoff_summary"]; !ok {
		t.Error("Expected summary to be stored in context vars")
	} else if summary != "This is a conversation summary" {
		t.Errorf("Expected summary in context vars, got %v", summary)
	}
}

func TestCallback_WithoutSummary(t *testing.T) {
	agent := agents.NewAgent("TestAgent")
	h := New(agent)
	tool := h.ToTool()

	contextVars := agents.ContextVariables{}
	args := map[string]any{}

	result, err := tool.Callback(args, contextVars)

	if err != nil {
		t.Errorf("Callback should not error: %v", err)
	}

	if result.(*agents.Agent) != agent {
		t.Error("Expected callback to return the target agent")
	}
}

func TestIsHandoffToolMarker(t *testing.T) {
	agent := agents.NewAgent("TestAgent")
	h := New(agent)
	tool := h.ToTool()

	if !tool.IsHandoffTool {
		t.Error("Expected IsHandoffTool to be true for handoff-generated tools")
	}

	// Regular tool should	// Test enabling with a real tool
	target := agents.NewAgent("TargetAgent")
	handoffTool := New(target).ToTool()
	regularTool := tools.New("test", "desc", nil, func(_ map[string]any, _ agents.ContextVariables) (any, error) { return nil, nil })
	agent.Tools = []tools.Tool{handoffTool, regularTool}

	if regularTool.IsHandoffTool {
		t.Error("Expected IsHandoffTool to be false for regular tools")
	}
}
