package tools

import (
	"encoding/json"
	"fmt"

	"github.com/openai/openai-go/v3"
)

// ContextVariables is a map of variables that can be passed to functions.
type ContextVariables map[string]any

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

	// IsHandoffTool indicates this tool performs an agent handoff.
	// This field is set automatically by the handoff package.
	IsHandoffTool bool

	// NeedsApproval indicates whether this tool requires human approval before execution.
	// When true, the runner will pause execution and wait for explicit approval.
	NeedsApproval bool

	// ApprovalFunc is an optional callback that determines if approval is needed for a specific call.
	// If set, this function is called with the tool arguments and call ID.
	// It should return true if approval is required for this specific invocation.
	ApprovalFunc func(args map[string]any, callID string, ctx ContextVariables) (bool, error)
}

// ToParam converts the Tool to an openai.ChatCompletionToolUnionParam.
func (t Tool) ToParam() openai.ChatCompletionToolUnionParam {
	// If parameters are empty, default to empty object
	params := t.Parameters
	if params == nil {
		params = map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		}
	}

	// Use v3 API helper function
	return openai.ChatCompletionFunctionTool(openai.FunctionDefinitionParam{
		Name:        t.Name,
		Description: openai.String(t.Description),
		Parameters:  openai.FunctionParameters(params),
	})
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

// RequiresApproval checks if a tool requires approval for execution.
// It evaluates both the NeedsApproval flag and the ApprovalFunc if present.
func (t Tool) RequiresApproval(args map[string]any, callID string, ctx ContextVariables) (bool, error) {
	// If the tool always needs approval, return true
	if t.NeedsApproval {
		return true, nil
	}

	// If there's a dynamic approval function, call it
	if t.ApprovalFunc != nil {
		return t.ApprovalFunc(args, callID, ctx)
	}

	// No approval required
	return false, nil
}

// New creates a Tool from a simpler definition.
// For now, it accepts manual schema. In the future, we could use reflection.
func New(name, description string, params map[string]any, callback func(map[string]any, ContextVariables) (any, error)) Tool {
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
