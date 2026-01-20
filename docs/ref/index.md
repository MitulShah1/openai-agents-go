# API Reference

Complete API reference for the OpenAI Agents Go SDK.

## Package Structure

```
github.com/MitulShah1/openai-agents-go
├── agents              # Core package
│   ├── Agent          # Agent type and configuration
│   ├── Runner         # Agent execution orchestration
│   ├── Tool           # Tool interface and implementations
│   ├── Result         # Execution results and steps
│   └── ...
├── session/            # Session backends
│   ├── MemorySession
│   ├── FileSession
│   └── ConversationsSession
├── guardrail/          # Guardrail framework
│   ├── content/       # Content validation
│   ├── moderation/    # Content moderation
│   ├── ratelimit/     # Rate limiting
│   └── security/      # Security checks
└── internal/           # Internal packages
    └── jsonschema/    # JSON schema builder
```

## Core Types

### [Agent](#)
The central type representing an AI agent with configured behavior.

```go
type Agent struct {
    Name            string
    Instructions    string
    Model           string
    Temperature     float64
    MaxTokens       int
    Tools           []Tool
    ResponseFormat  any
    OnBeforeRun     func(context.Context, *Agent) error
    OnAfterRun      func(context.Context, *Agent, *Result) error
    InputGuardrails []*guardrail.Guardrail
    OutputGuardrails []*guardrail.Guardrail
}
```

_Detailed documentation coming soon_

### [Runner](#)
Orchestrates agent execution and manages the interaction loop.

```go
type Runner struct {
    // Contains filtered or unexported fields
}

func NewRunner(client *openai.Client) *Runner
func (*Runner) Run(
    ctx context.Context,
    agent *Agent,
    messages []openai.ChatCompletionMessageParamUnion,
    config *RunConfig,
    options ...RunOption,
) (*Result, error)
```

_Detailed documentation coming soon_

### [Tool](#)
Interface for functions that agents can call.

```go
type Tool interface {
    ToParam() openai.ChatCompletionToolParam
    Execute(args string, ctx ContextVariables) (any, error)
}

// Factory functions
func FunctionTool(name, description string, schema map[string]any, fn ToolFunction) Tool
func HandoffTool(agent *Agent, description string) Tool
```

See [Tools Guide](../tools.md) for details.

### [Result & Types](#)
Execution results, steps, usage tracking, and error types.

```go
type Result struct {
    FinalOutput  string
    Agent        *Agent
    Steps        []Step
    Usage        Usage
    ContextVars  ContextVariables
}

type Step struct {
    Index       int
    AgentName   string
    Type        string
    Content     string
    ToolCalls   []ToolCall
    Timestamp   time.Time
}

type Usage struct {
    PromptTokens     int
    CompletionTokens int
    TotalTokens      int
}
```

_Detailed API documentation coming soon_

## Advanced Features

### [Guardrails](guardrails/index.md)
Input/output validation framework with 9+ built-in guardrails.

```go
type Guardrail interface {
    Validate(text string, isInput bool) error
    Name() string
    IsTripwire() bool
}

// Built-in guardrails (subpackages)
func security.NewPII(opts ...PIIOption) *guardrail.Guardrail
func moderation.NewOpenAI(client *openai.Client, opts ...Option) *guardrail.Guardrail
func ratelimit.New(config Config) *guardrail.Guardrail
// ... and more in content, security, moderation, ratelimit
```

[View Guardrails documentation →](guardrails/index.md)

### [Sessions](sessions/index.md)
Conversation persistence with multiple storage backends.

```go
type Session interface {
    Load(sessionID string) ([]openai.ChatCompletionMessageParamUnion, error)
    Save(sessionID string, messages []openai.ChatCompletionMessageParamUnion) error
}

// Implementations
func NewMemorySession() Session
func NewFileSession(directory string) (Session, error)
func NewConversationsSession(client *openai.Client) (Session, error)
```

[View Sessions documentation →](sessions/index.md)

## Configuration

### RunConfig
Control agent execution behavior.

```go
type RunConfig struct {
    MaxTurns int           // Maximum tool call iterations
    Timeout  time.Duration // Overall timeout
    Debug    bool          // Enable debug logging
}
```

### RunOptions
Functional options for Run method.

```go
// Available options
func WithSession(session Session, sessionID string) RunOption
func WithGuardrails(guardrails []Guardrail) RunOption
```

## Quick Reference

### Creating an Agent
```go
agent := agents.NewAgent("MyAgent")
agent.Instructions = "You are helpful"
agent.Model = "gpt-4"
```

### Running an Agent
```go
runner := agents.NewRunner(&client)
result, err := runner.Run(ctx, agent, messages, config, options...)
```

### Using Tools
```go
tool := agents.FunctionTool("name", "desc", schema, func(args map[string]any, ctx agents.ContextVariables) (any, error) {
    return result, nil
})
agent.Tools = []agents.Tool{tool}
```

### Using Guardrails
```go
guardrails := []agents.Guardrail{
    security.NewPII(security.WithTripwire(true)),
}
result, err := runner.Run(ctx, agent, messages, nil, agents.WithGuardrails(guardrails))
```

### Using Sessions
```go
sess := session.NewMemorySession()
result, err := runner.Run(ctx, agent, messages, nil, agents.WithSession(sess, "user-id"))
```

## External Dependencies

The SDK minimizes dependencies:

### Required
- `github.com/openai/openai-go` - Official OpenAI Go SDK

### Optional
- None for core functionality
- Database drivers for future session backends (v0.3.0+)

## Version Compatibility

| SDK Version | Go Version | OpenAI SDK |
|-------------|------------|------------|
| v0.2.3 | 1.24+ | v3.16.0+ |
| v0.2.0 | 1.24+ | v3.16.0+ |
| v0.1.0 | 1.24+ | v3.16.0+ |

## See Also

- [GoDoc](https://pkg.go.dev/github.com/MitulShah1/openai-agents-go) - Auto-generated Go documentation
- [GitHub Examples](https://github.com/MitulShah1/openai-agents-go/tree/main/examples) - Working code samples
- [CHANGELOG](https://github.com/MitulShah1/openai-agents-go/blob/main/CHANGELOG.md) - Version history
