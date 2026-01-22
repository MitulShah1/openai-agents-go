package handoff

import (
	"testing"

	agents "github.com/MitulShah1/openai-agents-go"
)

const simpleTransferDesc = "Transfer the conversation to Support"

func TestGenerateToolName_Simple(t *testing.T) {
	tests := []struct {
		agentName string
		expected  string
	}{
		{"Support", "transfer_to_support"},
		{"Billing", "transfer_to_billing"},
		{"Sales", "transfer_to_sales"},
	}

	for _, tt := range tests {
		t.Run(tt.agentName, func(t *testing.T) {
			agent := agents.NewAgent(tt.agentName)
			h := New(agent)

			result := h.generateToolName()

			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestGenerateToolName_WithSpaces(t *testing.T) {
	agent := agents.NewAgent("Technical Support")
	h := New(agent)

	result := h.generateToolName()
	expected := "transfer_to_technical_support"

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestGenerateToolName_MixedCase(t *testing.T) {
	agent := agents.NewAgent("BillingAgent")
	h := New(agent)

	result := h.generateToolName()
	expected := "transfer_to_billingagent"

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestGenerateToolName_MultipleSpaces(t *testing.T) {
	agent := agents.NewAgent("Customer   Support   Team")
	h := New(agent)

	result := h.generateToolName()

	// Should replace all spaces with underscores
	if result != "transfer_to_customer___support___team" {
		t.Errorf("Expected all spaces to be replaced, got %q", result)
	}
}

func TestGenerateDescription_WithShortInstructions(t *testing.T) {
	agent := agents.NewAgent("Support")
	agent.Instructions = "Help customers"

	h := New(agent)
	result := h.generateDescription()

	expected := "Transfer the conversation to Support. Help customers"

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestGenerateDescription_WithLongInstructions(t *testing.T) {
	agent := agents.NewAgent("Support")
	// 150 character instruction
	agent.Instructions = "This is a very long instruction that exceeds one hundred characters and should be truncated when generating the tool description for handoff"

	h := New(agent)
	result := h.generateDescription()

	// Should be truncated to 100 chars + "..."
	if len(result) > len("Transfer the conversation to Support. ")+103 {
		t.Errorf("Expected description to be truncated, got length %d: %q", len(result), result)
	}

	// Should end with "..."
	if result[len(result)-3:] != "..." {
		t.Errorf("Expected truncated description to end with '...', got %q", result)
	}
}

func TestGenerateDescription_WithoutInstructions(t *testing.T) {
	agent := agents.NewAgent("Support")
	// NewAgent sets default instructions, explicitly set to empty
	agent.Instructions = ""

	h := New(agent)
	result := h.generateDescription()

	expected := "Transfer the conversation to Support"

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestGenerateDescription_WithFunctionInstructions(t *testing.T) {
	agent := agents.NewAgent("Support")
	// Set instructions as a function (not a string)
	agent.Instructions = func() string {
		return "Dynamic instructions"
	}

	h := New(agent)
	result := h.generateDescription()

	// Should fall back to simple description without instructions

	if result != simpleTransferDesc {
		t.Errorf("Expected %q for non-string instructions, got %q", simpleTransferDesc, result)
	}
}

func TestGetToolName_UsesCustomName(t *testing.T) {
	agent := agents.NewAgent("Support")
	h := New(agent, WithToolName("custom_tool"))

	result := h.getToolName()

	if result != "custom_tool" {
		t.Errorf("Expected custom name to be used, got %q", result)
	}
}

func TestGetToolName_GeneratesWhenNotSet(t *testing.T) {
	agent := agents.NewAgent("Support")
	h := New(agent)

	result := h.getToolName()

	if result != "transfer_to_support" {
		t.Errorf("Expected generated name, got %q", result)
	}
}

func TestGetDescription_UsesCustomDescription(t *testing.T) {
	agent := agents.NewAgent("Support")
	h := New(agent, WithDescription("Custom description"))

	result := h.getDescription()

	if result != "Custom description" {
		t.Errorf("Expected custom description to be used, got %q", result)
	}
}

func TestGetDescription_GeneratesWhenNotSet(t *testing.T) {
	agent := agents.NewAgent("Support")
	// Explicitly clear default instructions
	agent.Instructions = ""

	h := New(agent)
	result := h.getDescription()

	if result != simpleTransferDesc {
		t.Errorf("Expected generated description, got %q", result)
	}
}

func TestToTool_Parameters(t *testing.T) {
	agent := agents.NewAgent("TestAgent")
	h := New(agent)
	tool := h.ToTool()

	// Check parameters structure
	if tool.Parameters == nil {
		t.Fatal("Expected parameters to be set")
	}

	params, ok := tool.Parameters["type"]
	if !ok || params != "object" {
		t.Error("Expected parameters to have type 'object'")
	}

	properties, ok := tool.Parameters["properties"].(map[string]any)
	if !ok {
		t.Fatal("Expected properties to be a map")
	}

	// Should have 'summary' parameter
	summary, ok := properties["summary"]
	if !ok {
		t.Error("Expected 'summary' parameter to exist")
	}

	summaryMap, ok := summary.(map[string]any)
	if !ok {
		t.Fatal("Expected summary to be a map")
	}

	if summaryMap["type"] != "string" {
		t.Errorf("Expected summary type to be 'string', got %v", summaryMap["type"])
	}
}

func TestCreateCallback_ReturnsFunction(t *testing.T) {
	agent := agents.NewAgent("TestAgent")
	h := New(agent)

	callback := h.createCallback()

	if callback == nil {
		t.Error("Expected callback to be created")
	}
}

func TestCreateCallback_ExecutesSuccessfully(t *testing.T) {
	agent := agents.NewAgent("TestAgent")
	h := New(agent)

	callback := h.createCallback()
	result, err := callback(map[string]any{}, agents.ContextVariables{})

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	resultAgent, ok := result.(*agents.Agent)
	if !ok {
		t.Fatal("Expected result to be *agents.Agent")
	}

	if resultAgent != agent {
		t.Error("Expected callback to return target agent")
	}
}
