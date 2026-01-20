# Multimodal Tools Example

This example demonstrates the new **multimodal tool output support** introduced in v0.3.0.

## Features Demonstrated

- **ImageContent**: Tools returning image URLs with detail levels
- **FileContent**: Tools generating files dynamically
- **TextContent**: Explicit text content (default behavior)
- **Backward Compatibility**: Existing string-returning tools still work

## What's New in v0.3.0

Tools can now return rich `Content` objects instead of just strings:

```go
// Image response
return agents.ImageContent("https://example.com/image.png", "high"), nil

// File response
return agents.FileContent(
    []byte("file content"),
    "report.md",
    "text/markdown",
), nil

// Traditional string (still works!)
return "Simple text response", nil
```

## Tool Examples

###  1. `generate_placeholder_image`
Returns an `ImageContent` object with a placeholder image URL.

### 2. `create_report`
Returns a `FileContent` object with a dynamically generated markdown report.

### 3. `get_weather`
Traditional string response demonstrating backward compatibility.

## Running the Example

```bash
export OPENAI_API_KEY="your-api-key"
go run main.go
```

## Output

The example shows three scenarios:
1. Agent uses image tool → receives `ImageContent`
2. Agent uses file tool → receives `FileContent`  
3. Agent uses text tool → receives plain string (backward compatible)

## Key APIs

```go
// Create multimodal content
content := agents.ImageContent(url, detail)
content := agents.FileContent(data, filename, mimeType)
content := agents.TextContent(text)

// Check if result is Content
if content, ok := agents.IsContent(result); ok {
    fmt.Println(content.String())
}
```

## Learn More

- See `tool.go` for full Content API
- Check `tool_multimodal_test.go` for comprehensive examples
