package agents

import (
	"encoding/json"
	"fmt"

	"github.com/openai/openai-go"
)

// Tool represents a function that can be called by an agent.
type Tool struct {
	// Name is the name of the tool.
	Name string
	// Description is the description of the tool.
	Description string
	// Parameters is the JSON schema for the tool parameters.
	Parameters map[string]any
	// Callback is the function to execute when the tool is called.
	// It receives the arguments as a map and context variables.
	Callback func(args map[string]any, ctx ContextVariables) (any, error)
}

// ToParam converts the Tool to an openai.ChatCompletionToolParam.
func (t Tool) ToParam() openai.ChatCompletionToolParam {
	// If parameters are empty, default to empty object
	params := t.Parameters
	if params == nil {
		params = map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		}
	}

	return openai.ChatCompletionToolParam{
		Function: openai.FunctionDefinitionParam{
			Name:        t.Name,
			Description: openai.String(t.Description),
			Parameters:  openai.FunctionParameters(params),
		},
	}
}

// Execute runs the tool's callback with the provided arguments.
func (t Tool) Execute(argsJSON string, ctx ContextVariables) (any, error) {
	// Handle empty args - default to empty JSON object
	if argsJSON == "" {
		argsJSON = "{}"
	}

	var args map[string]any
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return nil, fmt.Errorf("failed to unmarshal arguments: %w", err)
	}

	// Validate callback exists
	if t.Callback == nil {
		return nil, fmt.Errorf("tool %s has no callback function", t.Name)
	}

	return t.Callback(args, ctx)
}

// FunctionTool is a helper to create a Tool from a simpler definition.
// For now, it accepts manual schema. In the future, we could use reflection.
func FunctionTool(name, description string, params map[string]any, callback func(map[string]any, ContextVariables) (any, error)) Tool {
	if name == "" {
		panic("tool name cannot be empty")
	}
	if callback == nil {
		panic("tool callback cannot be nil")
	}

	return Tool{
		Name:        name,
		Description: description,
		Parameters:  params,
		Callback:    callback,
	}
}

// IsHandoff checks if the result is an Agent, indicating a handoff.
func IsHandoff(result any) (*Agent, bool) {
	a, ok := result.(*Agent)
	return a, ok
}
