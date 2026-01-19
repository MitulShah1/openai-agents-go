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
| SQLite | Coming soon | None (pure Go) | v0.3.0 |
| Redis | Coming soon | Redis client | v0.3.0 |
| PostgreSQL | Coming soon | PostgreSQL driver | v0.3.0 |

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
// Automatic synchronization
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
    // Load conversation history for a session ID
    Load(sessionID string) ([]openai.ChatCompletionMessageParamUnion, error)
    
    // Save conversation history for a session ID
    Save(sessionID string, messages []openai.ChatCompletionMessageParamUnion) error
}
```

## How Sessions Work

1. **Before agent run**: Runner calls `session.Load(sessionID)` to get history
2. **History prepended**: Loaded messages are added before new messages
3. **Agent executes**: With full context from previous interactions
4. **After completion**: Runner calls `session.Save(sessionID, allMessages)` to persist

```
User Message → Load History → [History + New Message] → Agent → Save Updated History
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

## Advanced Topics

### Manual Session Control

You can manually load/save sessions:

```go
// Load manually
history, err := sess.Load("user-123")

// Modify history
history = append(history, openai.UserMessage("New message"))

// Save manually
err = sess.Save("user-123", history)
```

### Session Migration

Coming in v0.3.0: Migrate between backends

```go
// Future API (v0.3.0)
err := session.Migrate(fileSession, redisSession, "user-123")
```

## API Reference

See [Sessions API Reference](../ref/sessions/index.md) for detailed API documentation.
