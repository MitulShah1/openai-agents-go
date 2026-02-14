# Models

The `models` package provides a flexible abstraction layer for LLM interactions. It allows agents to use different model backends (OpenAI, Azure, compatible APIs) and testing mocks while maintaining a consistent interface.

## Core Concepts

### Model Interface

The `Model` interface defines the contract for interacting with an LLM.

```go
type Model interface {
    // GetResponse generates a single completion
    GetResponse(ctx context.Context, req *ModelRequest) (*ModelResponse, error)
    
    // StreamResponse streams the completion chunks
    StreamResponse(ctx context.Context, req *ModelRequest) (*ssestream.Stream[ModelResponse], error)
    
    // ModelName returns the identifier of the model (e.g., "gpt-4o")
    ModelName() string
}
```

### Model Provider

A `ModelProvider` is a factory that creates `Model` instances based on a model name.

```go
type ModelProvider interface {
    GetModel(modelName string) (Model, error)
}
```

The SDK comes with a built-in `OpenAIProvider` which wraps the official `openai-go` client.

## Usage

### Default Behavior

By default, the `Runner` uses an `OpenAIProvider` initialized with the client you pass to `NewRunner`.

```go
client := openai.NewClient(...)
runner := agents.NewRunner(client) // Automatically creates OpenAIProvider
```

### Custom Providers

You can inject a custom provider using `NewRunnerWithProvider`. This is useful for using Azure OpenAI, other compatible APIs, or for testing.

```go
// Example: Using a mock provider for testing
mockProvider := NewMockProvider()
runner := agents.NewRunnerWithProvider(mockProvider)
```

### Agent-Level Overrides

You can specify a `ModelProvider` on a specific agent. This takes precedence over the runner's provider.

```go
agent := agents.NewAgent("SpecialAgent")
agent.ModelProvider = mySpecialProvider // Only this agent uses this provider
```

## Resolution Logic

When an agent needs to call a model, the `Runner` resolves the `Model` implementation using the following precedence:

1.  **Agent Provider**: If `agent.ModelProvider` is set, it is used.
2.  **Runner Provider**: If not, `runner.ModelProvider` is used.
3.  **Fallback**: If neither is set (legacy mode), it falls back to the `openai.Client` inside the runner.

## Settings

Model behavior is controlled by `ModelSettings`. These settings are derived from the agent's configuration (Temperature, MaxTokens, ResponseFormat) and passed to the model.

```go
type ModelSettings struct {
    Model          string
    Temperature    *float64
    MaxTokens      *int
    Stop           []string
    ResponseFormat *jsonschema.ResponseFormat
    ParallelToolCalls bool
    // ...
}
```

The `Model` implementation is responsible for applying these settings to the underlying API request.
