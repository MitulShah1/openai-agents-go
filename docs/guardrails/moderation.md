# Moderation Guardrails

## OpenAI Moderation API

Use OpenAI's official content moderation:

```go
import "github.com/openai/openai-go"
import "github.com/MitulShah1/openai-agents-go/guardrail/moderation"

client := openai.NewClient(/* ... */)

openAIMod := moderation.NewOpenAI(
    &client,
    moderation.WithModerationTripwire(true),
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
openAIMod := moderation.NewOpenAI(
    &client,
    moderation.WithModerationThreshold(0.5),
    moderation.WithModerationTripwire(true),
)
```

## Profanity Detection

Filter toxic content with severity levels:

```go
import "github.com/MitulShah1/openai-agents-go/guardrail/moderation"

// Block all profanity
profanity := moderation.NewProfanity(moderation.ProfanityConfig{
    Tripwire: true,
})

// Custom word list
custom := moderation.NewProfanity(moderation.ProfanityConfig{
    WordList: map[string]moderation.SeverityLevel{
        "badword1": moderation.SeverityHigh,
        "badword2": moderation.SeverityMedium,
    },
    Tripwire: true,
})
```

Features:
- Comprehensive word lists (Low, Medium, High severity)
- Leetspeak normalization (@ → a, $ → s, ! → i)
- Case-insensitive matching

## Prompt Injection Detection

Protect against LLM security attacks:

```go
import "github.com/MitulShah1/openai-agents-go/guardrail/moderation"

injection := moderation.NewInjection(moderation.PromptInjectionConfig{
    Tripwire: true,
})

// Detects 13 attack patterns:
// - Instruction override attempts
// - Role manipulation
// - Jailbreak attempts
// - Delimiter attacks
// - Encoding attacks (base64, hex, unicode)
// - System prompt extraction
// - And more...
```
