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

**See the complete example:** [`examples/21_parallel_tools`](https://github.com/MitulShah1/openai-agents-go/blob/main/examples/21_parallel_tools/main.go)

### Key Features

- ✅ **Goroutine-based execution** - Lightweight concurrent tool execution
- ✅ **Order preservation** - Results maintain tool call order despite async execution
- ✅ **Error isolation** - One tool's error doesn't block others in parallel mode
- ✅ **Semaphore pattern** - `MaxToolConcurrency` prevents resource exhaustion

## Tool Approvals

Tool approvals enable human-in-the-loop safety for dangerous or sensitive operations. When a tool requires approval, the runner pauses execution and lets the caller decide whether to proceed.

### Static Approval

Mark a tool as always requiring approval:

```go
deleteTool := tools.New("delete_db", "Delete a database", params,
    func(args map[string]any, ctx tools.ContextVariables) (any, error) {
        return deleteDatabase(args["name"].(string))
    },
)
deleteTool.NeedsApproval = true
```

### Dynamic Approval

Use `ApprovalFunc` for conditional approval based on arguments:

```go
transferTool := tools.New("transfer", "Transfer funds", params,
    func(args map[string]any, ctx tools.ContextVariables) (any, error) {
        return transfer(args)
    },
)
transferTool.ApprovalFunc = func(args map[string]any, callID string, ctx tools.ContextVariables) (bool, error) {
    amount, _ := args["amount"].(float64)
    return amount > 1000, nil // Only require approval for large transfers
}
```

### Inline Approval Handler

For synchronous approval decisions (e.g., CLI prompts), use `WithApprovalHandler`:

```go
result, err := runner.Run(ctx, agent, messages,
    agents.WithApprovalHandler(func(req tools.ApprovalRequest) (*tools.ApprovalResponse, error) {
        fmt.Printf("Tool %s wants to run with args %v. Approve? ", req.ToolName, req.Args)
        // ... get user input ...
        return &tools.ApprovalResponse{Approved: true}, nil
    }),
)
```

When the handler approves, execution continues normally. When it rejects, the runner returns a `ToolApprovalRequiredError`.

### Pause/Resume Workflow

When no handler is set and a tool requires approval, `Run()` returns a `ToolApprovalRequiredError` with a `RunState` snapshot. The caller can then approve or reject and resume:

```go
result, err := runner.Run(ctx, agent, messages)

var approvalErr *agents.ToolApprovalRequiredError
if errors.As(err, &approvalErr) {
    // Inspect what needs approval
    for _, req := range approvalErr.Requests {
        fmt.Printf("Tool %s (call %s) needs approval\n", req.ToolName, req.CallID)
    }

    // Make decisions
    approvals := map[string]*tools.ApprovalResponse{
        approvalErr.Requests[0].CallID: {Approved: true},
    }

    // Resume execution
    result, err = runner.Resume(ctx, approvalErr.State, approvals)
}
```

### Streaming

In streaming mode, a `StreamEventApprovalRequired` event is emitted before the stream terminates with `ToolApprovalRequiredError`. This lets stream consumers display approval UI before the error arrives.

### Parallel Tool Calls

If any tool call in a parallel batch requires approval, the **entire batch** is interrupted. No tools are executed partially. This prevents side effects from tools that ran before the interrupted one.

**See the complete example:** [`examples/23_tool_approvals`](https://github.com/MitulShah1/openai-agents-go/blob/main/examples/23_tool_approvals/main.go)

## Related Topics

- [Agents](agents.md)
- [Quickstart Guide](quickstart.md#adding-tools)
- [Parallel Tools Example](https://github.com/MitulShah1/openai-agents-go/blob/main/examples/21_parallel_tools/main.go)
