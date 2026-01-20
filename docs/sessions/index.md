# Sessions

Sessions provide conversation persistence, allowing agents to remember context across multiple interactions.

## Overview

The session framework enables:
- **Persistent conversation history** across runs
- **Pluggable storage backends** (in-memory, file, cloud)
- **Automatic history management** integrated into Runner
- **Thread-safe concurrent access**

## Available Backends

| Backend | Use Case | Dependencies | Version |
|---------|----------|--------------|---------|
| Memory | Testing, development | None | v0.2.0 |
| File | Single-server production | None | v0.2.0 |
| Conversations API | Cloud, distributed | OpenAI API key | v0.2.2 |
| SQLite | Single-server persistence | None (pure Go) | v0.3.0 |
| Redis | Scalable/Distributed | Redis client | v0.3.5 |
| PostgreSQL | Enterprise/Compliance | PostgreSQL driver | v0.3.5 |

## Quick Start

```go
import "github.com/MitulShah1/openai-agents-go/session"

// Create session
sess := session.NewMemorySession()

// Run agent with session
result, err := runner.Run(
    ctx,
    agent,
    messages,
    nil,
    agents.WithSession(sess, "user-123"), // Session ID
)
```

## Session Backends

### In-Memory Session

Best for development and testing:

```go
sess := session.NewMemorySession()

// Thread-safe concurrent access
// Data lost when process ends
// Zero external dependencies
```

### File-Based Session

Best for single-server production:

```go
import "github.com/MitulShah1/openai-agents-go/session"

sess, err := session.NewFileSession("./sessions")
if err != nil {
    panic(err)
}

// Persists to disk as JSON files
// Atomic writes prevent corruption
// Works across process restarts
```

### Conversations API

Best for cloud and distributed systems:

```go
import (
    "github.com/MitulShah1/openai-agents-go/session"
    "github.com/openai/openai-go"
)

client := openai.NewClient(/* ... */)

sess, err := session.NewConversationsSession(&client)
if err != nil {
    panic(err)
}

// Cloud-based persistence via OpenAI
// Distributed-ready
// Cloud-based persistence via OpenAI
// Distributed-ready
// Automatic synchronization
```

### SQLite Session (New in v0.3.0)

Best for robust single-server persistence without external database servers:

```go
import "github.com/MitulShah1/openai-agents-go/session"

sess, err := session.NewSQLite("./sessions.db")
if err != nil {
    panic(err)
}

// Uses modernc.org/sqlite (Pure Go, CGO-free)
// Automatic schema migration
// Thread-safe connection pooling
```

## Complete Example

```go
package main

import (
    "context"
    "fmt"

    agents "github.com/MitulShah1/openai-agents-go"
    "github.com/MitulShah1/openai-agents-go/session"
    "github.com/openai/openai-go"
)

func main() {
    client := openai.NewClient(/* ... */)
    runner := agents.NewRunner(&client)

    agent := agents.NewAgent("Assistant")
    agent.Instructions = "You are a helpful assistant who remembers context"

    // Create session
    sess := session.NewMemorySession()
    sessionID := "user-alice"

    // First interaction
    result1, _ := runner.Run(
        context.Background(),
        agent,
        []openai.ChatCompletionMessageParamUnion{
            openai.UserMessage("My name is Alice and I like Go programming"),
        },
        nil,
        agents.WithSession(sess, sessionID),
    )
    fmt.Println("Response 1:", result1.FinalOutput)

    // Second interaction - agent remembers context
    result2, _ := runner.Run(
        context.Background(),
        agent,
        []openai.ChatCompletionMessageParamUnion{
            openai.UserMessage("What's my name and what do I like?"),
        },
        nil,
        agents.WithSession(sess, sessionID),
    )
    fmt.Println("Response 2:", result2.FinalOutput)
    // Output: Your name is Alice and you like Go programming.
}
```

## Session Interface

All backends implement the `Session` interface:

```go
type Session interface {
    // Get retrieves all messages for a session ID
    Get(ctx context.Context, sessionID string) ([]openai.ChatCompletionMessageParamUnion, error)
    
    // Append adds messages to a session
    Append(ctx context.Context, sessionID string, messages []openai.ChatCompletionMessageParamUnion) error
    
    // Clear removes all messages from a session
    Clear(ctx context.Context, sessionID string) error
    
    // Delete removes a session completely
    Delete(ctx context.Context, sessionID string) error
}
```

## How Sessions Work

1. **Before agent run**: Runner calls `session.Get(ctx, sessionID)` to get history
2. **History prepended**: Loaded messages are added before new messages
3. **Agent executes**: With full context from previous interactions
4. **After completion**: Runner calls `session.Append(ctx, sessionID, newMessages)` to persist

```
User Message → Get History → [History + New Message] → Agent → Append New History
```

## Choosing a Backend

| Scenario | Recommended Backend |
|----------|-------------------|
| Local development | Memory Session |
| Unit tests | Memory Session |
| Single-server app | File Session |
| Multi-server app | Conversations API |
| Cloud deployment | Conversations API |
| High scalability | Redis (v0.3.0) |
| Enterprise | PostgreSQL (v0.3.0) |

## Best Practices

### Session IDs

Use meaningful, unique session IDs:

```go
// ✅ Good: User-specific
sessionID := fmt.Sprintf("user-%s", userID)

// ✅ Good: Conversation-specific
sessionID := fmt.Sprintf("conv-%s", conversationID)

// ❌ Bad: Not unique
sessionID := "session"
```

### Memory Management

For file-based sessions, consider periodic cleanup:

```go
// Delete old sessions
sess.Delete(oldSessionID)
```

For memory sessions, clear when done:

```go
// Clear all sessions
sess = session.NewMemorySession()
```

### Error Handling

Always check session errors:

```go
result, err := runner.Run(
    ctx, agent, messages, nil,
    agents.WithSession(sess, sessionID),
)
if err != nil {
    // Session save/load errors are wrapped
    fmt.Println("Session error:", err)
}
```

## Session Utilities (New in v0.3.0)

Enhance any session backend with transparent compression and encryption.

### Compression

Reduce storage size by compressing message history using GZIP.

```go
// Wrap any session with compression
sess = session.WithCompression(baseSession)
```

### Encryption

Secure your conversation history with AES-GCM encryption.

```go
key := []byte("your-32-byte-secret-key-12345678")
// Wrap session with encryption (requires 16, 24, or 32 byte key)
sess = session.WithEncryption(baseSession, key)
```

### Composition

Utilities can be composed. For "Compress then Encrypt" strategy (recommended):

```go
// Data -> Compress -> Encrypt -> Store
store, _ := session.NewSQLite("chats.db")

// 1. Encryption wraps the store (Inner)
encrypted := session.WithEncryption(store, key)

// 2. Compression wraps the encrypted session (Outer)
finalSession := session.WithCompression(encrypted)
```

## Advanced Topics

### Manual Session Control

You can manually load/save sessions:

```go
// Load manually
history, err := sess.Get(context.Background(), "user-123")

// Modify history
history = append(history, openai.UserMessage("New message"))

// Append manually (or use Append to add just new messages)
err = sess.Append(context.Background(), "user-123", []openai.ChatCompletionMessageParamUnion{
    openai.UserMessage("New message"),
})
```

### Session Migration

Coming in v0.3.0: Migrate between backends

```go
// Future API (v0.3.0)
err := session.Migrate(fileSession, redisSession, "user-123")
```

## API Reference

See [Sessions API Reference](../ref/sessions/index.md) for detailed API documentation.
