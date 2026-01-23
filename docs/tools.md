# Tools

Tools enable agents to perform actions by calling Go functions.

## Overview

Tools are functions that agents can discover and call during execution. The SDK provides a simple interface for creating and using tools.

## Creating a Tool

### Function Tool

The most common type of tool wraps a Go function:

```go
import (
    "fmt"
    "github.com/MitulShah1/openai-agents-go/tools"
)

weatherTool := tools.New(
    "get_weather",
    "Get the current weather for a city",
    map[string]any{
        "type": "object",
        "properties": map[string]any{
            "city": map[string]any{
                "type":        "string",
                "description": "The city name",
            },
        },
        "required": []string{"city"},
    },
    func(args map[string]any, ctx tools.ContextVariables) (any, error) {
        city := args["city"].(string)
        // Call weather API
        return fmt.Sprintf("Weather in %s is sunny", city), nil
    },
)
```

### Handoff Tool

Transfer control to another agent:

```go
import "github.com/MitulShah1/openai-agents-go/handoff"

transferTool := handoff.New(specialistAgent).ToTool()
```

### Multimodal Tool (New in v0.3.0)

Return rich content like images and files that multimodal models (e.g., GPT-4o) can process:

```go
cameraTool := tools.New(
    "get_camera_feed",
    "Get snapshot from security camera",
    /* ... params ... */,
    func(args map[string]any, ctx tools.ContextVariables) (any, error) {
        // Return structured content
        return []tools.Content{
            tools.TextContent("Here is the latest snapshot:"),
            tools.ImageContent("https://example.com/snap.jpg", "high"),
        }, nil
    },
)
```

## Tool Interface

All tools implement the `Tool` interface:

```go
type Tool struct {
	Name        string
	Description string
	Parameters  map[string]any
	Callback    func(args map[string]any, ctx ContextVariables) (any, error)
}
```

## Complete Example

See [Quickstart Guide](quickstart.md#adding-tools) for more details.

## Parallel Tool Execution

By default, agents execute multiple tool calls in parallel using goroutines. This significantly improves performance when the model requests multiple independent tools.

### Enabling/Disabling Parallel Execution

**Agent-level configuration:**
```go
agent := agents.NewAgent("my-agent")
agent.ParallelToolCalls = true  // Default: enabled
```

**Runtime override:**
```go
runner.Run(ctx, agent, messages,
    agents.WithConfig(&agents.RunConfig{
        ParallelToolCalls: boolPtr(false), // Sequential execution
    }),
)
```

### Concurrency Limiting

Limit the number of tools running simultaneously to prevent resource exhaustion:

```go
runner.Run(ctx, agent, messages,
    agents.WithConfig(&agents.RunConfig{
        ParallelToolCalls:  boolPtr(true),
        MaxToolConcurrency: 3, // Max 3 tools at once
    }),
)
```

### Performance Considerations

| Execution Mode | Best For | Performance |
|----------------|----------|-------------|
| **Parallel** | I/O-bound tools (API calls, database queries) | ~2-3x faster for independent tools |
| **Sequential** | Stateful tools with dependencies | Predictable, deterministic order |
| **Limited Concurrency** | Resource-constrained environments | Balanced performance and resource usage |

### OpenAI API Integration

The `ParallelToolCalls` setting is transmitted to the OpenAI API:
- `true` - Explicitly enables parallel tool calls (model can request multiple tools)
- `false` - Restricts to one tool call per turn
- `nil` - Uses provider default (typically parallel)

### Example: Parallel vs Sequential

```go
// Parallel execution (default)
start := time.Now()
result, _ := runner.Run(ctx, agent, messages)
parallelDuration := time.Since(start)
fmt.Printf("Parallel: %v\n", parallelDuration)

// Sequential execution
start = time.Now()
result, _ = runner.Run(ctx, agent, messages,
    agents.WithConfig(&agents.RunConfig{
        ParallelToolCalls: boolPtr(false),
    }),
)
sequentialDuration := time.Since(start)
fmt.Printf("Sequential: %v\n", sequentialDuration)
```

**See the complete example:** [`examples/21_parallel_tools`](../examples/21_parallel_tools/main.go)

### Key Features

- ✅ **Goroutine-based execution** - Lightweight concurrent tool execution
- ✅ **Order preservation** - Results maintain tool call order despite async execution
- ✅ **Error isolation** - One tool's error doesn't block others in parallel mode
- ✅ **Semaphore pattern** - `MaxToolConcurrency` prevents resource exhaustion

## Related Topics

- [Agents](agents.md)
- [Quickstart Guide](quickstart.md#adding-tools)
- [Parallel Tools Example](../examples/21_parallel_tools/main.go)
