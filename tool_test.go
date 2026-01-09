package agents

import (
	"errors"
	"testing"
)

func TestFunctionToolCreation(t *testing.T) {
	callback := func(_ map[string]any, _ ContextVariables) (any, error) {
		return "result", nil
	}

	tool := FunctionTool(
		"test_tool",
		"A test tool",
		map[string]any{"type": "object"},
		callback,
	)

	if tool.Name != "test_tool" {
		t.Errorf("expected Name=test_tool, got %s", tool.Name)
	}

	if tool.Description != "A test tool" {
		t.Errorf("expected Description='A test tool', got %s", tool.Description)
	}

	if tool.Callback == nil {
		t.Error("expected Callback to be set")
	}
}

func TestFunctionToolPanic(t *testing.T) {
	tests := []struct {
		name        string
		toolName    string
		callback    func(map[string]any, ContextVariables) (any, error)
		shouldPanic bool
	}{
		{
			name:     "empty name panics",
			toolName: "",
			callback: func(_ map[string]any, _ ContextVariables) (any, error) {
				return nil, nil
			},
			shouldPanic: true,
		},
		{
			name:        "nil callback panics",
			toolName:    "valid_name",
			callback:    nil,
			shouldPanic: true,
		},
		{
			name:     "valid tool does not panic",
			toolName: "valid_tool",
			callback: func(_ map[string]any, _ ContextVariables) (any, error) {
				return nil, nil
			},
			shouldPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if tt.shouldPanic && r == nil {
					t.Error("expected panic but didn't get one")
				}
				if !tt.shouldPanic && r != nil {
					t.Errorf("unexpected panic: %v", r)
				}
			}()

			FunctionTool(tt.toolName, "desc", nil, tt.callback)
		})
	}
}

func TestToolExecute(t *testing.T) {
	tests := []struct {
		name         string
		argsJSON     string
		callback     func(map[string]any, ContextVariables) (any, error)
		expectError  bool
		expectResult any
	}{
		{
			name:     "valid JSON arguments",
			argsJSON: `{"location": "Paris"}`,
			callback: func(args map[string]any, _ ContextVariables) (any, error) {
				return args["location"], nil
			},
			expectError:  false,
			expectResult: "Paris",
		},
		{
			name:     "empty JSON defaults to empty object",
			argsJSON: "",
			callback: func(args map[string]any, _ ContextVariables) (any, error) {
				return len(args), nil
			},
			expectError:  false,
			expectResult: 0,
		},
		{
			name:     "invalid JSON returns error",
			argsJSON: `{invalid json}`,
			callback: func(_ map[string]any, _ ContextVariables) (any, error) {
				return nil, nil
			},
			expectError: true,
		},
		{
			name:     "callback error is returned",
			argsJSON: `{}`,
			callback: func(_ map[string]any, _ ContextVariables) (any, error) {
				return nil, errors.New("callback error")
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := Tool{
				Name:     "test_tool",
				Callback: tt.callback,
			}

			result, err := tool.Execute(tt.argsJSON, nil)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.expectError && result != tt.expectResult {
				t.Errorf("expected result=%v, got %v", tt.expectResult, result)
			}
		})
	}
}

func TestToolExecuteNilCallback(t *testing.T) {
	tool := Tool{
		Name:     "test_tool",
		Callback: nil,
	}

	_, err := tool.Execute(`{}`, nil)
	if err == nil {
		t.Error("expected error for nil callback")
	}

	expectedMsg := "tool test_tool has no callback function"
	if err.Error() != expectedMsg {
		t.Errorf("expected error %q, got %q", expectedMsg, err.Error())
	}
}

func TestIsHandoff(t *testing.T) {
	agent := NewAgent("SupportAgent")

	// Test with agent pointer
	result, ok := IsHandoff(agent)
	if !ok {
		t.Error("expected IsHandoff to return true for agent pointer")
	}
	if result != agent {
		t.Error("expected IsHandoff to return the same agent")
	}

	// Test with non-agent value
	result, ok = IsHandoff("not an agent")
	if ok {
		t.Error("expected IsHandoff to return false for non-agent value")
	}
	if result != nil {
		t.Error("expected IsHandoff to return nil for non-agent value")
	}
}

func TestToParam(t *testing.T) {
	tool := Tool{
		Name:        "test_tool",
		Description: "A test tool",
		Parameters: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}

	param := tool.ToParam()

	if param.Function.Name != "test_tool" {
		t.Errorf("expected Name=test_tool, got %s", param.Function.Name)
	}

	// Test with nil parameters
	toolNilParams := Tool{
		Name:        "test_tool2",
		Description: "Another test",
		Parameters:  nil,
	}

	param2 := toolNilParams.ToParam()
	// Should default to empty object schema
	if param2.Function.Name != "test_tool2" {
		t.Error("failed to create param with nil parameters")
	}
}
