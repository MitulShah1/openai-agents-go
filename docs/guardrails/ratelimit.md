# Rate Limiting

Prevent abuse with distributed-ready rate limiting:

```go
import (
    "time"
    "github.com/MitulShah1/openai-agents-go/guardrail/ratelimit"
)

// Global rate limit
rateLimiter := ratelimit.New(ratelimit.Config{
    Rate:     100,
    Burst:    100,
    Window:   time.Minute,
    Tripwire: true,
})

// Per-user rate limit
perUserLimit := ratelimit.New(ratelimit.Config{
    Rate:   10,
    Burst:  10,
    Window: time.Minute,
    KeyFunc: func(text string) string {
        // Extract user ID from input or default to custom key
        return "user-123"
    },
    Tripwire: true,
})
```
