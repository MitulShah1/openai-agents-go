# Models

The Model Provider abstraction decouples the runner from a specific LLM backend. It allows agents to use different providers (OpenAI, Anthropic, Google, etc.) without changing runner or agent code.

## Overview

The `models` package provides three key interfaces:

- **`Model`** — Represents a single LLM that can generate responses
- **`ModelProvider`** — Resolves model name strings to `Model` instances
- **`MultiProvider`** — Routes model names to different providers via prefix conventions

## Basic Usage

By default, `NewRunner(client)` wraps the OpenAI client in an `OpenAIProvider`. Existing code continues to work without changes.

```go
// This still works exactly as before
client := openai.NewClient(option.WithAPIKey(apiKey))
runner := agents.NewRunner(&client)
```

## Using Providers Explicitly

For more control, use `NewRunnerWithProvider`:

```go
import "github.com/MitulShah1/openai-agents-go/models"

provider := models.NewOpenAIProvider(&client)
runner := agents.NewRunnerWithProvider(provider)
```

## Per-Agent Providers

Each agent can have its own `ModelProvider`, overriding the runner's default:

```go
agent := agents.NewAgent("Premium Agent")
agent.ModelProvider = models.NewOpenAIProvider(&premiumClient)
```

The resolution order is:

1. `Agent.ModelProvider` (if set)
2. `Runner.ModelProvider` (if set)
3. `Runner.Client` wrapped in `OpenAIProvider` (backward compatibility)

## MultiProvider

The `MultiProvider` routes model names to different providers using a prefix convention:

```go
openaiProvider := models.NewOpenAIProvider(&openaiClient)

multi := models.NewMultiProvider(openaiProvider) // default
multi.AddProvider("openai", openaiProvider)
// multi.AddProvider("anthropic", anthropicProvider)

runner := agents.NewRunnerWithProvider(multi)

// No prefix → default provider
agent1 := agents.NewAgent("GPT Agent")
agent1.Model = "gpt-4o"

// Prefix → routes to specific provider
agent2 := agents.NewAgent("GPT Mini")
agent2.Model = "openai/gpt-4o-mini"
```

## Custom Providers

Implement the `models.Model` interface to add support for any LLM:

```go
type Model interface {
    GetResponse(ctx context.Context, params openai.ChatCompletionNewParams, settings ModelSettings) (*ModelResponse, error)
    StreamResponse(ctx context.Context, params openai.ChatCompletionNewParams, settings ModelSettings) (*ssestream.Stream[openai.ChatCompletionChunk], error)
    ModelName() string
}
```

A custom provider implements `ModelProvider`:

```go
type ModelProvider interface {
    GetModel(modelName string) (Model, error)
}
```

## ModelSettings

`ModelSettings` carries LLM parameters (temperature, max tokens, etc.) that can be specified at the agent or run level:

```go
type ModelSettings struct {
    Temperature       *float64
    MaxTokens         *int
    TopP              *float64
    Stop              []string
    FrequencyPenalty  *float64
    PresencePenalty   *float64
    Seed              *int64
    ParallelToolCalls *bool
    ResponseFormat    any
    Prompt            *prompts.Prompt
}
```

The `Resolve` method merges two settings, with overrides taking precedence:

```go
base := &models.ModelSettings{Temperature: float64Ptr(0.5)}
override := &models.ModelSettings{Temperature: float64Ptr(0.9)}
merged := base.Resolve(override) // Temperature = 0.9
```

## ModelResponse

All model implementations return a normalized `ModelResponse`:

```go
type ModelResponse struct {
    Completion *openai.ChatCompletion  // The underlying completion
    Usage      ModelUsage              // Token usage statistics
    ResponseID string                  // Provider-specific response ID
}
```

Non-OpenAI providers should convert their responses to `*openai.ChatCompletion` format for compatibility with the runner.
