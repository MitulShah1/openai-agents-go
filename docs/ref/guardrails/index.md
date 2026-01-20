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

### PII Detection (security)
```go
func NewPII(opts ...PIIOption) *guardrail.Guardrail
```

Detects: emails, phones, SSNs, credit cards

### URL Filtering (security)
```go
func NewURLFilter(opts ...URLFilterOption) *guardrail.Guardrail
```

### Regex (content)
```go
func NewRegex(pattern string, opts ...RegexOption) *guardrail.Guardrail
```

### Moderation API (moderation - OpenAI)
```go
func NewOpenAI(client *openai.Client, opts ...Option) *guardrail.Guardrail
```

### Content Length (content)
```go
func NewLength(config Config) *guardrail.Guardrail

type CountMode string
const (
    CountModeCharacters CountMode = "characters"
    CountModeWords      CountMode = "words"
    CountModeLines      CountMode = "lines"
)
```

### Rate Limiting (ratelimit)
```go
func New(config Config) *guardrail.Guardrail

type RateLimiter interface {
    Allow(ctx context.Context, key string) (bool, error)
    Reset(ctx context.Context, key string) error
    Close() error
}
```

### Profanity Detection (moderation)
```go
func NewProfanity(config ProfanityConfig) *guardrail.Guardrail

type SeverityLevel string
const (
    SeverityLow    SeverityLevel = "low"
    SeverityMedium SeverityLevel = "medium"
    SeverityHigh   SeverityLevel = "high"
)
```

### Secrets Detection (security)
```go
func NewSecrets(config SecretsConfig) *guardrail.Guardrail
```

### Prompt Injection Detection (moderation)
```go
func NewInjection(config PromptInjectionConfig) *guardrail.Guardrail
```

## See Also

- [Guardrails Guide](../../guardrails/index.md)
- [Examples](https://github.com/MitulShah1/openai-agents-go/tree/main/examples/08_guardrails_demo)
