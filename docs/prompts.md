# Prompts

The `prompts` package allows you to manage and resolve agent instructions dynamically. It supports both static string prompts and dynamic functions that can adapt based on context.

## Overview

The `Prompt` struct is the central container for prompt logic. It can hold either a static string or a dynamic function.

```go
type Prompt struct {
    Static  string
    Dynamic DynamicPromptFunc
}
```

### Static Prompts

A static prompt is a fixed string. This is the simplest form of instruction.

```go
// Implicitly created when setting agent.Instructions = "..."
p := prompts.NewStaticPrompt("You are a helpful assistant.")
```

### Dynamic Prompts

Dynamic prompts allow you to generate instructions at runtime based on context variables.

```go
func dynamicInstructions(ctx context.Context, data prompts.DynamicPromptData) (string, error) {
    user := data.ContextVars["user_name"]
    return fmt.Sprintf("Hello %s, how can I help you?", user), nil
}

p := prompts.NewDynamicPrompt(dynamicInstructions)
```

## Integration with Agents

The `Agent` struct now has a `Prompt` field.

```go
agent := agents.NewAgent("Greeter")

// Option 1: Set Instructions (Helper for Static Prompt)
agent.Instructions = "Hello there!" 

// Option 2: Set Prompt directly
agent.Prompt = prompts.NewDynamicPrompt(myFunc)
```

When the agent runs, the `Runner` calls `agent.GetPrompt(ctx, contextVars)` to resolve the final instruction string before sending it to the model.

## Context Variables

`ContextVars` is a map passed to the dynamic prompt function. You can inject these variables when running the agent.

```go
vars := map[string]any{
    "user_name": "Alice",
    "role":      "admin",
}

runner.Run(ctx, agent, agents.RunnerOptions{
    ContextVars: vars,
})
```

Inside your dynamic prompt function, you access these via `data.ContextVars`.

## Advanced Usage

### Accessing Agent Info

The `DynamicPromptData` struct also contains information about the agent itself.

```go
type DynamicPromptData struct {
    ContextVars map[string]any
    Agent       AgentInfo
}

type AgentInfo struct {
    Name  string
    Model string
}
```

This ensures your prompt logic can stay decoupled from the specific agent instance while still knowing who it is "speaking" as.
