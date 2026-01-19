# Sessions API Reference

Complete API reference for the session framework and backends.

## Session Interface

```go
type Session interface {
    Load(sessionID string) ([]openai.ChatCompletionMessageParamUnion, error)
    Save(sessionID string, messages []openai.ChatCompletionMessageParamUnion) error
}
```

## Memory Session

In-memory session storage (thread-safe, zero dependencies).

```go
func NewMemorySession() Session
```

**Use cases**: Development, testing, single-instance apps

## File Session

File-based JSON storage with atomic writes.

```go
func NewFileSession(directory string) (Session, error)
```

**Parameters**:
- `directory`: Path to store session files

**Use cases**: Single-server production, persistent storage

## Conversations Session

Cloud-based storage using OpenAI Conversations API.

```go
func NewConversationsSession(client *openai.Client) (Session, error)
```

**Parameters**:
- `client`: OpenAI client with API key

**Use cases**: Distributed systems, cloud deployments

## Run Option

Use sessions with agents:

```go
func WithSession(session Session, sessionID string) RunOption
```

**Parameters**:
- `session`: Session backend implementation
- `sessionID`: Unique identifier for the conversation

## Example

```go
sess := session.NewMemorySession()
result, err := runner.Run(
    ctx, agent, messages, nil,
    agents.WithSession(sess, "user-123"),
)
```

## See Also

- [Sessions Guide](../sessions/index.md)
- [Examples](https://github.com/MitulShah1/openai-agents-go/tree/main/examples/09_sessions_demo)
