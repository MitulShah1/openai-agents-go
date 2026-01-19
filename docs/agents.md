# Agents

Agents are the core building blocks of the SDK. An agent represents an AI model configured with specific behavior, tools, and capabilities.

## Creating an Agent

The simplest way to create an agent is with `NewAgent()`:

```go
agent := agents.NewAgent("My Agent")
agent.Instructions = "You are a helpful coding assistant"
```

## Agent Configuration

### Basic Properties

```go
agent := agents.NewAgent("CodeHelper")

// Required: Instructions define the agent's behavior
agent.Instructions = "You are an expert Go programmer who helps with code reviews"

// Optional: Model selection (defaults to gpt-4o)
agent.Model = "gpt-4"
agent.Model = "gpt-4-turbo"
agent.Model = "gpt-3.5-turbo"

// Optional: Temperature controls randomness (0.0 - 2.0)
agent.Temperature = 0.7  // More creative
agent.Temperature = 0.0  // More deterministic

// Optional: Max tokens for responses
agent.MaxTokens = 500
```

### Tools

Agents can use tools to perform actions:

```go
agent.Tools = []agents.Tool{
    myCustomTool,
    agents.HandoffTool(otherAgent, "Transfer to specialist"),
}
```

See the [Tools documentation](tools.md) for more details.

### Response Format

Control output format with structured outputs:

```go
import "github.com/MitulShah1/openai-agents-go/internal/jsonschema"

schema := jsonschema.Object().
    WithProperty("summary", jsonschema.String()).
    WithRequired("summary")

agent.ResponseFormat = jsonschema.JSONSchema("response", schema)
```

See [Structured Outputs](structured_outputs.md) for details.

## Lifecycle Hooks

Execute code before and after agent runs:

```go
agent.OnBeforeRun = func(ctx context.Context, agent *agents.Agent) error {
    fmt.Println("Starting agent:", agent.Name)
    return nil
}

agent.OnAfterRun = func(ctx context.Context, agent *agents.Agent, result *agents.Result) error {
    fmt.Printf("Agent completed. Tokens used: %d\n", result.Usage.TotalTokens)
    return nil
}
```

## Complete Example

```go
package main

import (
    "context"
    "fmt"

    agents "github.com/MitulShah1/openai-agents-go"
    "github.com/openai/openai-go"
)

func main() {
    // Create fully configured agent
    agent := agents.NewAgent("ProductivityCoach")
    agent.Instructions = `You are a productivity coach who helps users:
    - Break down large tasks
    - Create actionable plans
    - Provide motivation and accountability`
    
    agent.Model = "gpt-4"
    agent.Temperature = 0.8
    agent.MaxTokens = 1000
    
    agent.OnBeforeRun = func(ctx context.Context, a *agents.Agent) error {
        fmt.Println("🤖 Agent starting...")
        return nil
    }
    
    agent.OnAfterRun = func(ctx context.Context, a *agents.Agent, r *agents.Result) error {
        fmt.Printf("✅ Completed in %d steps\n", len(r.Steps))
        return nil
    }

    // Run the agent
    client := openai.NewClient(/* ... */)
    runner := agents.NewRunner(&client)
    
    result, err := runner.Run(
        context.Background(),
        agent,
        []openai.ChatCompletionMessageParamUnion{
            openai.UserMessage("Help me organize my week"),
        },
        nil,
        nil,
    )
    
    if err != nil {
        panic(err)
    }
    
    fmt.Println(result.FinalOutput)
}
```

## Best Practices

### Clear Instructions

✅ **Good**: Specific, actionable instructions
```go
agent.Instructions = `You are a code reviewer. For each code submission:
1. Check for bugs and security issues
2. Suggest improvements for readability
3. Verify tests are present
4. Provide specific line-by-line feedback`
```

❌ **Bad**: Vague instructions
```go
agent.Instructions = "You help with code"
```

### Model Selection

- **gpt-4o**: Best balance of speed and capability (default)
- **gpt-4**: Most capable, slower and more expensive
- **gpt-4-turbo**: Fast and capable
- **gpt-3.5-turbo**: Fastest and cheapest, less capable

### Temperature Guidelines

| Temperature | Use Case | Example |
|------------|----------|---------|
| 0.0 - 0.3 | Deterministic tasks | Code generation, data extraction |
| 0.4 - 0.7 | Balanced | General assistance, Q&A |
| 0.8 - 1.0 | Creative tasks | Writing, brainstorming |
| 1.0+ | Highly creative | Poetry, experimental |

## Related Topics

- [Running Agents](running_agents.md) - Execute agents and handle results
- [Tools](tools.md) - Give agents capabilities
- [Handoffs](handoffs.md) - Transfer between agents
- [Structured Outputs](structured_outputs.md) - Type-safe responses
