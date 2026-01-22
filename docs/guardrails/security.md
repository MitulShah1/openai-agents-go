# Security Guardrails

## PII Detection

Detects and blocks personally identifiable information:

```go
import "github.com/MitulShah1/openai-agents-go/guardrail/security"

pii := security.NewPII(security.WithTripwire(true))

// Detects:
// - Email addresses: user@example.com
// - Phone numbers: (555) 123-4567, +1-555-123-4567
// - SSNs: 123-45-6789
// - Credit cards: 4111-1111-1111-1111
```

## URL Filtering

Control which URLs are allowed:

```go
import "github.com/MitulShah1/openai-agents-go/guardrail/security"

// Block specific domains
urlGuard := security.NewURLFilter(
    security.WithBlocklist("malicious.com", "spam.net"),
    security.WithURLTripwire(true),
)

// Allow only specific domains
urlGuardAllowed := security.NewURLFilter(
    security.WithAllowlist("example.com", "trusted.org"),
    security.WithURLTripwire(true),
)
```

## Secrets Detection

Prevent credential leakage:

```go
import "github.com/MitulShah1/openai-agents-go/guardrail/security"

secrets := security.NewSecrets(security.SecretsConfig{
    Tripwire: true,
})

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
