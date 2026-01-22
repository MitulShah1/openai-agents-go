# Content Guardrails

## Content Length Limits

Limit content size with multiple counting modes:

```go
import "github.com/MitulShah1/openai-agents-go/guardrail/content"

// Limit characters
charLimit := content.NewLength(content.Config{
    Max:          1000,
    Mode:         content.CountModeCharacters,
    Tripwire:     false,
})

// Limit words
wordLimit := content.NewLength(content.Config{
    Max:      200,
    Mode:     content.CountModeWords,
})

// Limit lines
lineLimit := content.NewLength(content.Config{
    Max:      50,
    Mode:     content.CountModeLines,
})
```

## Custom Regex Guardrails

Create pattern-based validation:

```go
import "github.com/MitulShah1/openai-agents-go/guardrail/content"

// Block specific patterns
regex := content.NewRegex(
    `\b\d{3}-\d{2}-\d{4}\b`,
    content.WithRegexTripwire(true),
)
```
