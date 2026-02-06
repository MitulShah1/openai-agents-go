package prompts

// Prompt represents a configuration for OpenAI's Prompts API.
// This allows dynamic prompt configuration at runtime.
type Prompt struct {
	// ID is the unique identifier of the prompt
	ID string `json:"id"`

	// Version is an optional version identifier
	Version string `json:"version"`

	// Variables contains optional substitution variables for the prompt
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// DynamicPromptFunc is a function that generates a prompt dynamically.
// It can be used to customize prompts based on runtime context.
type DynamicPromptFunc func(contextVars map[string]interface{}) (*Prompt, error)
