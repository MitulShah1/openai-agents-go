# Tools

Tools enable agents to perform actions by calling Go functions.

## Overview

Tools are functions that agents can discover and call during execution. The SDK provides a simple interface for creating and using tools.

## Creating a Tool

### Function Tool

The most common type of tool wraps a Go function:

```go
weatherTool := agents.FunctionTool(
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
    func(args map[string]any, ctx agents.ContextVariables) (any, error) {
        city := args["city"].(string)
        // Call weather API
        return fmt.Sprintf("Weather in %s is sunny", city), nil
    },
)
```

### Handoff Tool

Transfer control to another agent:

```go
handoff := agents.HandoffTool(specialistAgent, "Transfer to specialist")
```

### Multimodal Tool (New in v0.3.0)

Return rich content like images and files that multimodal models (e.g., GPT-4o) can process:

```go
cameraTool := agents.FunctionTool(
    "get_camera_feed",
    "Get snapshot from security camera",
    /* ... params ... */,
    func(args map[string]any, ctx agents.ContextVariables) (any, error) {
        // Return structured content
        return []agents.Content{
            agents.TextContent("Here is the latest snapshot:"),
            agents.ImageContent("https://example.com/snap.jpg", "high"),
        }, nil
    },
)
```

## Tool Interface

All tools implement the `Tool` interface:

```go
type Tool interface {
    ToParam() openai.ChatCompletionToolParam
    Execute(args string, ctx ContextVariables) (any, error)
}
```

## Complete Example

See [Quickstart Guide](quickstart.md#adding-tools) for more details.

## Related Topics

- [Agents](agents.md)
- [Quickstart Guide](quickstart.md#adding-tools)
