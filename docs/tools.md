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

## Tool Interface

All tools implement the `Tool` interface:

```go
type Tool interface {
    ToParam() openai.ChatCompletionToolParam
    Execute(args string, ctx ContextVariables) (any, error)
}
```

## Complete Example

See [Quickstart Guide](quickstart.md#adding-tools) and [API Reference](ref/tool.md) for more details.

## Related Topics

- [Agents](agents.md)
- [Handoffs](handoffs.md)
- [Context Variables](context.md)
