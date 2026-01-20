# Session Backends Example

This example demonstrates the **database session backends** feature introduced in v0.3.0.

## Features Demonstrated

### 1. SQLite File Backend
Persistent session storage using SQLite database:
- Automatic table creation and migrations
- Connection pooling
- Full CRUD operations
- Session survives application restarts

### 2. SQLite In-Memory Backend
Fast, temporary session storage:
-Use `:memory:` as the database path
- No disk I/O overhead
- Perfect for testing and temporary workflows
- Data lost when application closes

### 3. Session Registry Pattern
Plugin-based backend selection:
- Register custom backends
- Create sessions via `session.Create("backend_name", config)`
- Switch backends without code changes
- Pre-registered backends: memory, file, conversations, sqlite

### 4. Session Management
Full CRUD operations on session data:
- **Get**: Retrieve conversation history
- **Append**: Add messages to session
- **Clear**: Reset session while keeping the ID
- **Delete**: Remove session completely
- **Multi-user**: Isolate sessions by user ID

## Running the Example

```bash
export OPENAI_API_KEY="your-api-key"
go run main.go
```

The example will:
1. Create a SQLite database (`sessions.db`) in the current directory
2. Demonstrate persistence across multiple agent runs
3. Show in-memory session usage
4. Use the registry pattern to create file-based sessions
5. Perform session management operations

## Key APIs

### Creating SQLite Backend

```go
// File-based (persistent)
session, err := session.NewSQLite("./sessions.db")
defer session.Close()

// In-memory (temporary)
session, err := session.NewSQLite(":memory:")
```

### Using Registry

```go
// Create via registry
config := map[string]any{
    "directory": "./file_sessions",
}
fileSession, err := session.Create("file", config)
```

### Session Operations

```go
// Get history
messages, err := session.Get(ctx, "user123")

// Clear session
err = session.Clear(ctx, "user123")

// Delete session
err = session.Delete(ctx, "user123")
```

### Attaching to Agent

```go
agent.SessionBackend = sqliteSession
runner.Run(ctx, agent, messages, agents.WithSessionID("user123"))
```

## Storage Details

**SQLite Schema:**
```sql
CREATE TABLE sessions (
    session_id TEXT PRIMARY KEY,
    messages TEXT NOT NULL,
    updated_at INTEGER NOT NULL
)
```

- Messages stored as JSON
- Automatic timestamp tracking
- Efficient querying by session ID

## Production Considerations

**File Backend:**
- ✅ Persistent across restarts
- ✅ Good for single-server deployments
- ✅ No external dependencies
- ⚠️ Not suitable for distributed systems

**In-Memory Backend:**
- ✅ Fastest performance
- ✅ Perfect for testing
- ⚠️ Data lost on restart
- ⚠️ Not suitable for production

**For Production Multi-Server:**
Use Redis or PostgreSQL backends (coming in future releases) for distributed session management.

## Learn More

- See `session/sqlite.go` for implementation details
- See `session/registry.go` for plugin system
- See tests in `session/sqlite_test.go` for more examples
