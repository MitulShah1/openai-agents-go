# Guardrails API Reference

Complete API reference for the guardrail framework and built-in guardrails.

## Guardrail Interface

```go
type Guardrail interface {
    Validate(text string, isInput bool) error
    Name() string
    IsTripwire() bool
}
```

## Built-in Guardrails

### PII Detection
```go
func NewPIIGuardrail(name string, isTripwire bool) Guardrail
```

Detects: emails, phones, SSNs, credit cards

### URL Filtering
```go
func NewURLGuardrail(name string, blocklist, allowlist []string, isTripwire bool) Guardrail
```

### Custom Regex
```go
func NewRegexGuardrail(name string, patterns []string, isTripwire bool) Guardrail
```

### Moderation API
```go
func NewModerationGuardrail(name string, client *openai.Client, thresholds map[string]float64, isTripwire bool) Guardrail
```

### Content Length
```go
func NewContentLengthGuardrail(name string, maxLength int, mode CountingMode, isTripwire bool) Guardrail

type CountingMode int
const (
    CountCharacters CountingMode = iota
    CountWords
    CountLines
)
```

### Rate Limiting
```go
func NewRateLimitGuardrail(name string, limit int, window time.Duration, backend Backend, keyFunc KeyFunc, isTripwire bool) Guardrail

type Backend interface {
    Allow(key string, limit int, window time.Duration) (bool, error)
}

type KeyFunc func(text string, isInput bool) string
```

### Profanity Detection
```go
func NewProfanityGuardrail(name string, isTripwire bool) Guardrail
func NewProfanityGuardrailWithWords(name string, words map[string]Severity, isTripwire bool) Guardrail

type Severity int
const (
    SeverityLow Severity = iota
    SeverityMedium
    SeverityHigh
)
```

### Secrets Detection
```go
func NewSecretsGuardrail(name string, isTripwire bool) Guardrail
```

### Prompt Injection Detection
```go
func NewPromptInjectionGuardrail(name string, isTripwire bool) Guardrail
```

## See Also

- [Guardrails Guide](../../guardrails.md)
- [Examples](https://github.com/MitulShah1/openai-agents-go/tree/main/examples/08_guardrails_demo)
