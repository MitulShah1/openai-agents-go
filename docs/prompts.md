# Prompts

The Prompts API allows agents to use centrally managed prompt configurations from OpenAI's Prompts API, enabling prompt versioning, A/B testing, and runtime customization without code changes.

## Overview

The `prompts` package provides:

- **`Prompt`** — A static prompt configuration with ID, version, and variables
- **`DynamicPromptFunc`** — A function that generates prompts at runtime based on context

## Static Prompts

Set a fixed prompt on an agent using a `*prompts.Prompt`:

```go
import "github.com/MitulShah1/openai-agents-go/prompts"

agent := agents.NewAgent("Assistant")
agent.Prompt = &prompts.Prompt{
    ID:      "prompt_helpful_assistant",
    Version: "v2",
    Variables: map[string]any{
        "tone": "friendly",
    },
}
```

The prompt ID references a prompt in OpenAI's Prompts API. The version is optional (defaults to latest). Variables are substituted into the prompt template.

## Dynamic Prompts

Use `DynamicPromptFunc` for runtime prompt selection:

```go
agent.Prompt = prompts.DynamicPromptFunc(func(data prompts.DynamicPromptData) (*prompts.Prompt, error) {
    tier, _ := data.ContextVariables["tier"].(string)
    if tier == "premium" {
        return &prompts.Prompt{ID: "prompt_premium", Version: "v3"}, nil
    }
    return &prompts.Prompt{ID: "prompt_free"}, nil
})
```

The `DynamicPromptData` struct provides:

| Field | Type | Description |
|-------|------|-------------|
| `Agent` | `AgentInfo` | Snapshot of the current agent (Name, Model) |
| `ContextVariables` | `map[string]any` | Context variables passed to the current run |

## How Prompts Flow

When the runner executes an agent with a prompt configured:

1. The runner calls `agent.GetPrompt(contextVars)` to resolve the prompt
2. The resolved `*prompts.Prompt` is placed in `ModelSettings.Prompt`
3. The model implementation receives the prompt via `ModelSettings`
4. For OpenAI models, the prompt is forwarded to the API

```
Agent.Prompt → GetPrompt(contextVars) → ModelSettings.Prompt → Model.GetResponse()
```

## Using with Context Variables

Pass context variables through `WithContextVariables` — they are available to dynamic prompts:

```go
result, err := runner.Run(ctx, agent, messages,
    agents.WithContextVariables(agents.ContextVariables{
        "tier":      "premium",
        "user_name": "Alice",
    }),
)
```

## Error Handling

If a dynamic prompt function returns an error, the runner aborts the run:

```go
agent.Prompt = prompts.DynamicPromptFunc(func(data prompts.DynamicPromptData) (*prompts.Prompt, error) {
    if data.ContextVariables["tier"] == nil {
        return nil, fmt.Errorf("tier is required for prompt resolution")
    }
    return &prompts.Prompt{ID: "prompt_default"}, nil
})
```

Setting an unsupported type on `agent.Prompt` (anything other than `*prompts.Prompt` or `DynamicPromptFunc`) also returns an error.

## Prompts vs Instructions

Prompts and instructions serve different purposes:

| Aspect | Instructions | Prompts |
|--------|-------------|---------|
| **Location** | In code | Managed externally via Prompts API |
| **Updates** | Requires code change | Update without redeployment |
| **Versioning** | Via git | Built-in version management |
| **Variables** | Not supported | Template variable substitution |
| **Use case** | Core agent behavior | A/B testing, personalization |

Both can be used together — the agent's instructions provide the base system prompt, while the prompt adds API-managed configuration on top.
