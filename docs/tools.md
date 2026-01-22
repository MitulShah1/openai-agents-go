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

## Related Topics

- [Agents](agents.md)
- [Quickstart Guide](quickstart.md#adding-tools)
