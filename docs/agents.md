# Agents

The `Agent` is the core entity in the SDK. It encapsulates an LLM model, instructions, tools, and configuration settings.

## Basic Structure

At its simplest, an agent only needs a name and instructions:

```go
import agents "github.com/MitulShah1/openai-agents-go"

agent := agents.NewAgent("Assistant")
agent.Instructions = "You are a helpful AI assistant."
```

## Agent Attributes

| Field | Type | But | Description |
|-------|------|-----|-------------|
| `Name` | `string` | Required | The name of the agent. Used for logging and tool identification. |
| `Model` | `string` | Optional | The OpenAI model to use. Defaults to `gpt-4o`. |
| `Instructions` | `string` \| `func` | Required | The system prompt or instructions for the agent. |
| `Tools` | `[]Tool` | Optional | A list of tools the agent can use. |
| `ResponseFormat`| `*jsonschema.ResponseFormat` | Optional | Schema for [Structured Outputs](structured_outputs.md). |
| `Temperature` | `*float64` | Optional | Sampling temperature (0.0 - 2.0). |
| `MaxTokens` | `*int` | Optional | Max tokens for generated response. |
| `ParallelToolCalls` | `bool` | Optional | Whether to allow parallel tool execution (default: `true`). |

## Instructions

Instructions define the behavior and persona of the agent.

### Static Instructions

Most agents use a simple string for instructions:

```go
agent.Instructions = `You are a math tutor. 
Always explain your reasoning step-by-step.`
```

### Dynamic Instructions

For more advanced use cases, you can provide a function that returns the instructions string. This is useful for injecting dynamic context, such as user details or current state.

```go
agent.Instructions = func(ctx context.Context) string {
    userName := ctx.Value("user_name")
    return fmt.Sprintf("You are assisting user %s. Be polite.", userName)
}
```

## Tools

Agents can be equipped with tools to interact with external systems.

```go
agent.Tools = []agents.Tool{
    weatherTool,
    databaseTool,
}
```

When an agent decides to call a tool, the `Runner` executes the corresponding Go function and feeds the result back to the agent. See [Tools](tools.md) for more comprehensive documentation.

## Handoffs

Agents can "hand off" the conversation to another agent. This is the basis for multi-agent orchestration. A handoff occurs when a **Tool returns an Agent object**.

```go
// Define a specialized agent
salesAgent := agents.NewAgent("Sales")
salesAgent.Instructions = "You process sales orders."

// Define a tool that performs the handoff
transferTool := agents.FunctionTool("transfer_to_sales", "Transfer to sales department", nil, 
    func(args map[string]any, ctx agents.ContextVariables) (any, error) {
        // Returning *Agent triggers the handoff
        return salesAgent, nil
    },
)

// Equip the main agent with the transfer tool
mainAgent.Tools = []agents.Tool{transferTool}
```

## Lifecycle Hooks

You can attach hooks to run code before or after an agent executes. This is useful for logging, setup, or cleanup.

```go
// Run before the agent starts processing
agent.OnBeforeRun = func(ctx context.Context, agent *agents.Agent) error {
    log.Printf("Starting agent: %s", agent.Name)
    return nil
}

// Run after the agent finishes
agent.OnAfterRun = func(ctx context.Context, agent *agents.Agent, result *agents.Result) error {
    log.Printf("Agent finished. Usage: %d tokens", result.Usage.TotalTokens)
    return nil
}
```

## Guardrails

Agents can be configured with input guardrails (to validate user messages) and output guardrails (to validate model responses) for safety and compliance.

```go
agent.InputGuardrails = []*guardrail.Guardrail{piiGuardrail}
// ...
```

See [Guardrails](guardrails.md) for implementation details.
