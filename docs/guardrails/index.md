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

| Guardrail | Purpose | Package |
|-----------|---------|---------|
| [PII Detection](security.md#pii-detection) | Detect emails, phones, SSNs, credit cards | `security` |
| [URL Filtering](security.md#url-filtering) | Block/allow specific URLs | `security` |
| [Secrets Detection](security.md#secrets-detection) | Prevent credential leakage | `security` |
| [Profanity Detection](moderation.md#profanity-detection) | Filter toxic content | `moderation` |
| [Moderation API](moderation.md#openai-moderation-api) | OpenAI's content moderation | `moderation` |
| [Prompt Injection](moderation.md#prompt-injection-detection) | Detect LLM security attacks | `moderation` |
| [Content Length](content.md#content-length-limits) | Limit characters, words, or lines | `content` |
| [Custom Regex](content.md#custom-regex-guardrails) | Pattern-based validation | `content` |
| [Rate Limiting](ratelimit.md) | Prevent abuse | `ratelimit` |

## Basic Usage

```go
import "github.com/MitulShah1/openai-agents-go/guardrail/security"

func main() {
    // Create guardrails
    piiGuardrail := security.NewPII(security.WithTripwire(true))
    
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

## Tripwires

Tripwires halt execution immediately on failure:

```go
// Tripwire ON - execution stops if violated
critical := security.NewPII(security.WithTripwire(true))

// Tripwire OFF - logs violation but continues
warning := security.NewPII(security.WithTripwire(false))
```

## Best Practices

1. **Use tripwires for critical guardrails** (PII, secrets, injections)
2. **Log violations for non-tripwire guardrails** (profanity, length)
3. **Layer guardrails** from most to least critical
4. **Test guardrails** with known violation cases
5. **Monitor guardrail metrics** in production
