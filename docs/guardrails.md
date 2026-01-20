# Guardrails

Guardrails provide input and output validation to ensure your agents handle content safely and comply with your requirements.

## Overview

The guardrail framework allows you to:
- **Validate inputs** before sending to the LLM
- **Validate outputs** before returning to users
- **Chain multiple guardrails** for comprehensive protection
- **Use tripwires** to halt execution on critical failures

## Built-in Guardrails

The SDK includes 9+ production-ready guardrails:

| Guardrail | Purpose | Version |
|-----------|---------|---------|
| PII Detection | Detect emails, phones, SSNs, credit cards | v0.2.0 |
| URL Filtering | Block/allow specific URLs | v0.2.0 |
| Custom Regex | Pattern-based validation | v0.2.0 |
| Moderation API | OpenAI's content moderation (13 categories) | v0.2.1 |
| Content Length | Limit characters, words, or lines | v0.2.3 |
| Rate Limiting | Prevent abuse with token bucket algorithm | v0.2.3 |
| Profanity Detection | Filter toxic content with severity levels | v0.2.3 |
| Secrets Detection | Prevent credential leakage (12 types) | v0.2.3 |
| Prompt Injection | Detect LLM security attacks (13 patterns) | v0.2.3 |

## Basic Usage

```go
import "github.com/MitulShah1/openai-agents-go/guardrail/builtin"

func main() {
    // Create guardrails
    piiGuardrail := builtin.NewPIIGuardrail("PII Protection", true)
    
    // Run agent with guardrails
    result, err := runner.Run(
        ctx,
        agent,
        messages,
        nil,
        agents.WithGuardrails([]agents.Guardrail{piiGuardrail}),
    )
}
```

## PII Detection

Detects and blocks personally identifiable information:

```go
pii := builtin.NewPIIGuardrail("No PII", true)

// Detects:
// - Email addresses: user@example.com
// - Phone numbers: (555) 123-4567, +1-555-123-4567
// - SSNs: 123-45-6789
// - Credit cards: 4111-1111-1111-1111
```

## URL Filtering

Control which URLs are allowed:

```go
// Block specific domains
blocklist := []string{"malicious.com", "spam.net"}
urlGuard := builtin.NewURLGuardrail("URL Filter", blocklist, []string{}, true)

// Allow only specific domains
allowlist := []string{"example.com", "trusted.org"}
urlGuard := builtin.NewURLGuardrail("URL Filter", []string{}, allowlist, true)
```

## Content Length Limits

Limit content size with multiple counting modes:

```go
import "github.com/MitulShah1/openai-agents-go/guardrail/builtin"

// Limit characters
charLimit := builtin.NewContentLengthGuardrail(
    "Character Limit",
    1000,  // max length
    builtin.CountCharacters,
    false, // not a tripwire
)

// Limit words
wordLimit := builtin.NewContentLengthGuardrail(
    "Word Limit",
    200,
    builtin.CountWords,
    false,
)

// Limit lines
lineLimit := builtin.NewContentLengthGuardrail(
    "Line Limit",
    50,
    builtin.CountLines,
    false,
)
```

## Rate Limiting

Prevent abuse with distributed-ready rate limiting:

```go
import (
    "time"
    "github.com/MitulShah1/openai-agents-go/guardrail/builtin"
)

// Global rate limit
rateLimiter := builtin.NewRateLimitGuardrail(
    "API Rate Limit",
    100,                    // 100 requests
    time.Minute,            // per minute
    builtin.NewMemoryBackend(),
    nil,                    // global limit
    true,                   // tripwire
)

// Per-user rate limit
perUserLimit := builtin.NewRateLimitGuardrail(
    "Per-User Limit",
    10,
    time.Minute,
    builtin.NewMemoryBackend(),
    func(text string, isInput bool) string {
        // Extract user ID from context
        return "user-123"
    },
    true,
)
```

## Profanity Detection

Filter toxic content with severity levels:

```go
// Block all profanity
profanity := builtin.NewProfanityGuardrail("Profanity Filter", true)

// Custom word list
custom := builtin.NewProfanityGuardrailWithWords(
    "Custom Filter",
    map[string]builtin.Severity{
        "badword1": builtin.SeverityHigh,
        "badword2": builtin.SeverityMedium,
    },
    true,
)
```

Features:
- Comprehensive word lists (Low, Medium, High severity)
- Leetspeak normalization (@ → a, $ → s, ! → i)
- Case-insensitive matching

## Secrets Detection

Prevent credential leakage:

```go
secrets := builtin.NewSecretsGuardrail("Secrets Protection", true)

// Detects 12 secret types:
// - AWS Access Keys
// - GitHub Personal Access Tokens
// - Google API Keys
// - JWT Tokens
// - Private SSH Keys
// - Database Connection Strings
// - API Keys (generic patterns)
// - OAuth Tokens
// - And more...
```

## Prompt Injection Detection

Protect against LLM security attacks:

```go
injection := builtin.NewPromptInjectionGuardrail("Injection Protection", true)

// Detects 13 attack patterns:
// - Instruction override attempts
// - Role manipulation
// - Jailbreak attempts
// - Delimiter attacks
// - Encoding attacks (base64, hex, unicode)
// - System prompt extraction
// - And more...
```

## OpenAI Moderation API

Use OpenAI's official content moderation:

```go
import "github.com/openai/openai-go"

client := openai.NewClient(/* ... */)

moderation := builtin.NewModerationGuardrail(
    "Content Moderation",
    &client,
    nil,  // Use default thresholds
    true, // tripwire
)

// 13 moderation categories:
// - hate, hate/threatening
// - harassment, harassment/threatening
// - self-harm (intent, instructions)
// - sexual, sexual/minors
// - violence, violence/graphic
// - illicit, illicit/violent
```

Custom thresholds:

```go
thresholds := map[string]float64{
    "hate": 0.5,
    "violence": 0.7,
}

moderation := builtin.NewModerationGuardrail(
    "Custom Moderation",
    &client,
    thresholds,
    true,
)
```

## Custom Regex Guardrails

Create pattern-based validation:

```go
// Block specific patterns
patterns := []string{
    `\b\d{3}-\d{2}-\d{4}\b`,  // SSN pattern
    `password\s*[:=]\s*\S+`,   // Password assignments
}

regex := builtin.NewRegexGuardrail("Pattern Filter", patterns, true)
```

## Tripwires

Tripwires halt execution immediately on failure:

```go
// Tripwire ON - execution stops if violated
critical := builtin.NewPIIGuardrail("Critical PII", true)

// Tripwire OFF - logs violation but continues
warning := builtin.NewPIIGuardrail("PII Warning", false)
```

## Guardrail Composition
Since v0.3.0, you can compose complex guardrail logic using chains.

### Chaining Strategies

```go
import "github.com/MitulShah1/openai-agents-go/guardrail"

chain := guardrail.NewChain().
    // Add guardrails
    Add(builtin.NewPIIGuardrail("pii", true)).
    Add(builtin.NewProfanityGuardrail("profanity", true)).
    
    // Choose strategy
    WithStrategy(guardrail.Sequential). // Stop at first failure (Short-circuit)
    // OR
    // WithStrategy(guardrail.Parallel). // Run all concurrently
    // OR
    // WithStrategy(guardrail.StopOnFirstPass). // First pass wins
    
    Build()

// Use the chain as a single guardrail
runner.Run(..., agents.WithGuardrails([]agents.Guardrail{chain}))
```

### Async & Timeouts

Apply timeouts to guardrails (v0.3.0):

```go
// Timeout after 500ms
timedGuard := guardrail.WithTimeout(slowGuardrail, 500*time.Millisecond)

// Timeout with graceful degradation (warns instead of failing)
gracefulGuard := guardrail.WithTimeoutGraceful(slowGuardrail, 500*time.Millisecond)
```

### Metrics

Track guardrail performance (v0.3.0):

```go
metrics := guardrail.NewInMemoryMetrics()
monitoredGuard := guardrail.WithMetrics(myGuardrail, metrics)

// ... run agent ...

stats := metrics.GetStats("my_guardrail")
fmt.Printf("Avg Latency: %v", stats.AvgDuration())
```

## Legacy Chaining (Simple List)

You can also pass a simple list of guardrails to the runner. This is equivalent to a Sequential chain.

```go
guardrails := []agents.Guardrail{
    builtin.NewPIIGuardrail("PII", true),
    builtin.NewSecretsGuardrail("Secrets", true),
}
```

## Complete Example

```go
package main

import (
    "context"
    "fmt"
    "time"

    agents "github.com/MitulShah1/openai-agents-go"
    "github.com/MitulShah1/openai-agents-go/guardrail/builtin"
    "github.com/openai/openai-go"
)

func main() {
    client := openai.NewClient(/* ... */)
    runner := agents.NewRunner(&client)

    agent := agents.NewAgent("SecureAgent")
    agent.Instructions = "You are a helpful assistant"

    // Production guardrail stack
    guardrails := []agents.Guardrail{
        // Critical: Block dangerous content
        builtin.NewPIIGuardrail("PII Protection", true),
        builtin.NewSecretsGuardrail("Secrets Protection", true),
        builtin.NewPromptInjectionGuardrail("Injection Defense", true),
        
        // Moderation
        builtin.NewModerationGuardrail("Content Moderation", &client, nil, true),
        builtin.NewProfanityGuardrail("Profanity Filter", false),
        
        // Rate limiting
        builtin.NewRateLimitGuardrail(
            "API Rate Limit",
            100,
            time.Minute,
            builtin.NewMemoryBackend(),
            nil,
            true,
        ),
        
        // Quality control
        builtin.NewContentLengthGuardrail("Length Limit", 2000, builtin.CountCharacters, false),
    }

    messages := []openai.ChatCompletionMessageParamUnion{
        openai.UserMessage("Hello!"),
    }

    result, err := runner.Run(
        context.Background(),
        agent,
        messages,
        nil,
        agents.WithGuardrails(guardrails),
    )

    if err != nil {
        fmt.Println("Guardrail violation:", err)
        return
    }

    fmt.Println(result.FinalOutput)
}
```

## Best Practices

1. **Use tripwires for critical guardrails** (PII, secrets, injections)
2. **Log violations for non-tripwire guardrails** (profanity, length)
3. **Layer guardrails** from most to least critical
4. **Test guardrails** with known violation cases
5. **Monitor guardrail metrics** in production

## API Reference

See [Guardrails API Reference](ref/guardrails/index.md) for detailed API documentation.
