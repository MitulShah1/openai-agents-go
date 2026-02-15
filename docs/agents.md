# Agents

The `Agent` is the core entity in the SDK. It encapsulates an LLM model, instructions, tools, and configuration settings.

## Basic Structure

At its simplest, an agent only needs a name and instructions:

```go
import (
    agents "github.com/MitulShah1/openai-agents-go"
    "github.com/MitulShah1/openai-agents-go/tools"
)

agent := agents.NewAgent("Assistant")
agent.Instructions = "You are a helpful AI assistant."
```

## Agent Attributes

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `Name` | `string` | Required | The name of the agent. Used for logging and tool identification. |
| `Model` | `string` | Optional | The OpenAI model to use. Defaults to `gpt-4o`. |
| `ModelProvider` | `models.ModelProvider` | Optional | Custom model provider for this agent. Overrides runner's default. See [Models](models.md). |
| `Instructions` | `string` \| `func` | Required | The system prompt or instructions for the agent. |
| `Prompt` | `*prompts.Prompt` \| `DynamicPromptFunc` | Optional | OpenAI Prompts API configuration. See [Prompts](prompts.md). |
| `Tools` | `[]tools.Tool` | Optional | A list of tools the agent can use. |
| `Skills` | `[]agents.Skill` | Optional | Runtime capability bundles that append instructions, tools, and guardrails. |
| `ResponseFormat`| `*jsonschema.ResponseFormat` | Optional | Schema for [Structured Outputs](structured_outputs.md). |
| `Temperature` | `*float64` | Optional | Sampling temperature (0.0 - 2.0). |
| `MaxTokens` | `*int` | Optional | Max tokens for generated response. |
| `ParallelToolCalls` | `bool` | Optional | Whether to allow parallel tool execution (default: `true`). |
| `ModelProvider` | `models.ModelProvider` | Optional | Per-agent model provider. Overrides runner's default. See [Models](models.md). |
| `Prompt` | `*prompts.Prompt` \| `DynamicPromptFunc` | Optional | Prompts API configuration. See [Prompts](prompts.md). |

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
agent.Tools = []tools.Tool{
    weatherTool,
    databaseTool,
}
```

When an agent decides to call a tool, the `Runner` executes the corresponding Go function and feeds the result back to the agent. See [Tools](tools.md) for more comprehensive documentation.


## Runtime Skills

Use runtime skills to compose reusable capabilities onto an agent.

A runtime skill can contribute:
- Additional instructions
- Tools
- Input guardrails
- Output guardrails

```go
supportSkill := agents.Skill{
    Name:         "customer_support",
    Description:  "Handle support triage and escalation",
    Instructions: "Ask clarifying questions first, then classify urgency.",
    Tools:        []tools.Tool{lookupTicketTool, escalateTool},
}

agent := agents.NewAgent("Support Assistant")
agent.Instructions = "You are a customer support assistant."
agent.AddSkill(supportSkill)
```

`Agent.GetInstructions` appends runtime skill instructions after the agent's base instructions.

> Note: This runtime `agents.Skill` API is separate from project-local Codex skill files stored in `.agents/skills`.

## Model Provider

By default, agents use the runner's model provider. Set `ModelProvider` on an agent for per-agent provider overrides:

```go
import "github.com/MitulShah1/openai-agents-go/models"

agent := agents.NewAgent("Premium")
agent.ModelProvider = models.NewOpenAIProvider(&premiumClient)
```

See [Models](models.md) for provider patterns and custom implementations.

## Prompts

Agents can use OpenAI's Prompts API for externally managed prompt configurations:

```go
import "github.com/MitulShah1/openai-agents-go/prompts"

// Static prompt
agent.Prompt = &prompts.Prompt{
    ID:      "prompt_helpful",
    Version: "v2",
}

// Dynamic prompt
agent.Prompt = prompts.DynamicPromptFunc(func(data prompts.DynamicPromptData) (*prompts.Prompt, error) {
    return &prompts.Prompt{ID: "prompt_" + data.Agent.Name}, nil
})
```

See [Prompts](prompts.md) for full documentation.

## Handoffs

Agents can "hand off" the conversation to another agent. This is the basis for multi-agent orchestration. A handoff occurs when a **Tool returns an Agent object**.

```go
import (
    "github.com/MitulShah1/openai-agents-go/handoff"
    "github.com/MitulShah1/openai-agents-go/tools"
)

// Define a specialized agent
salesAgent := agents.NewAgent("Sales")
salesAgent.Instructions = "You process sales orders."

// Define a tool that performs the handoff
transferTool := handoff.New(salesAgent).ToTool()

// Equip the main agent with the transfer tool
mainAgent.Tools = []tools.Tool{transferTool}
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

## Tool Approvals

Tools can require human approval before execution. Set `NeedsApproval` for static approval or `ApprovalFunc` for conditional logic.

```go
dangerousTool := tools.New("delete_db", "Delete database", params, callback)
dangerousTool.NeedsApproval = true

agent.Tools = []tools.Tool{dangerousTool}
```

When the runner encounters a tool requiring approval, it either calls the configured `ApprovalHandler` or returns a `ToolApprovalRequiredError` for pause/resume workflows. See [Tools - Tool Approvals](tools.md#tool-approvals) for details.

## Guardrails

Agents can be configured with input guardrails (to validate user messages) and output guardrails (to validate model responses) for safety and compliance.

```go
agent.InputGuardrails = []*guardrail.Guardrail{piiGuardrail}
// ...
```

See [Guardrails](guardrails/index.md) for implementation details.
