// Package prompts provides types and utilities for dynamic prompt configuration.
//
// The Prompts API allows agents to use centrally managed prompt configurations
// from OpenAI's Prompts API, enabling prompt versioning, A/B testing, and
// runtime customization without code changes.
//
// Prompts can be:
//   - Static: A fixed [Prompt] struct with an ID, version, and variables
//   - Dynamic: A [DynamicPromptFunc] that generates prompts at runtime
//     based on context variables, agent state, or other runtime data
//
// Example (static prompt):
//
//	agent.Prompt = &prompts.Prompt{
//	    ID:      "prompt_helpful_assistant",
//	    Version: "v2",
//	    Variables: map[string]any{
//	        "tone": "friendly",
//	    },
//	}
//
// Example (dynamic prompt):
//
//	agent.Prompt = prompts.DynamicPromptFunc(func(data prompts.DynamicPromptData) (*prompts.Prompt, error) {
//	    promptID := "prompt_default"
//	    if data.Agent.Name == "Premium" {
//	        promptID = "prompt_premium"
//	    }
//	    return &prompts.Prompt{ID: promptID}, nil
//	})
package prompts

// Prompt represents a configuration for OpenAI's Prompts API.
// When set on an agent, it is passed to the model alongside (or instead of)
// the agent's system instructions.
//
// The Prompts API allows managing prompt content, model, and tool configuration
// externally, so changes can be made without redeploying code.
type Prompt struct {
	// ID is the unique identifier of the prompt in the Prompts API.
	ID string `json:"id"`

	// Version is an optional version identifier for the prompt.
	// If empty, the latest version is used.
	Version string `json:"version,omitempty"`

	// Variables contains optional substitution variables for the prompt template.
	// These are replaced in the prompt content at runtime.
	Variables map[string]any `json:"variables,omitempty"`
}

// DynamicPromptData provides context to dynamic prompt functions.
// It carries information about the current agent and runtime state
// so the function can decide which prompt to use.
type DynamicPromptData struct {
	// Agent is a lightweight snapshot of the current agent.
	// Use Agent.Name to distinguish between agents in a multi-agent workflow.
	Agent AgentInfo

	// ContextVariables are the context variables passed to the current run.
	ContextVariables map[string]any
}

// AgentInfo is a lightweight snapshot of agent metadata passed to dynamic prompts.
// It avoids importing the top-level agents package to prevent circular dependencies.
type AgentInfo struct {
	// Name is the agent's name.
	Name string

	// Model is the agent's configured model name (e.g., "gpt-4o").
	Model string
}

// DynamicPromptFunc is a function that generates a prompt dynamically based on
// runtime context. It receives a [DynamicPromptData] with agent info and context
// variables, and returns the resolved [Prompt] to use.
//
// Return an error to abort the run if prompt resolution fails.
//
// Example:
//
//	prompts.DynamicPromptFunc(func(data prompts.DynamicPromptData) (*prompts.Prompt, error) {
//	    if data.ContextVariables["tier"] == "premium" {
//	        return &prompts.Prompt{ID: "prompt_premium", Version: "v3"}, nil
//	    }
//	    return &prompts.Prompt{ID: "prompt_default"}, nil
//	})
type DynamicPromptFunc func(data DynamicPromptData) (*Prompt, error)
