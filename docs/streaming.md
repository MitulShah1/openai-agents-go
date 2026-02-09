# Streaming

Streaming lets you subscribe to updates of the agent run as it proceeds. This is useful for showing end-user progress updates and partial responses in real-time.

## Quick Start

To stream, call `StreamWithResult()`, which returns a `*stream.Result`. Use the `StreamEvents()` method to get an iterator over streaming events:

```go
result, err := runner.StreamWithResult(ctx, agent, messages)
if err != nil {
    return err
}

for event, err := range result.StreamEvents(ctx) {
    if err != nil {
        return err
    }
    // Handle event
}
```

## Event Types

There are three main types of streaming events:

### 1. Raw Response Events

`RawResponseEvent` contains raw events passed directly from the LLM in OpenAI Responses API format. These are useful for streaming response messages to the user as they're generated.

**Example - Streaming text token-by-token:**

```go
for event, err := range result.StreamEvents(ctx) {
    if err != nil {
        return err
    }
    
    if raw, ok := event.(*stream.RawResponseEvent); ok {
        if raw.Type == "response.output_text.delta" {
            if data, ok := raw.Data.(map[string]any); ok {
                if delta, ok := data["delta"].(string); ok {
                    fmt.Print(delta) // Print each token as it arrives
                }
            }
        }
    }
}
```

**Common raw event types:**
- `response.created` - Response started
- `response.output_text.delta` - Text content delta
- `response.function_call_arguments.delta` - Function argument delta
- `response.output_item.done` - Output item completed
- `response.completed` - Response finished

### 2. Run Item Events

`RunItemEvent` represents high-level semantic events that inform you when an item has been fully generated. This allows progress updates at the level of "message generated", "tool ran", etc.

**Example - Tracking agent actions:**

```go
for event, err := range result.StreamEvents(ctx) {
    if err != nil {
        return err
    }
    
    if runItem, ok := event.(*stream.RunItemEvent); ok {
        switch runItem.Name {
        case string(stream.MessageOutputCreated):
            fmt.Println("💬 Agent responded")
        case string(stream.ToolCalled):
            fmt.Println("🔧 Tool called")
        case string(stream.ToolOutput):
            fmt.Println("📤 Tool completed")
        case string(stream.HandoffRequested):
            fmt.Println("🔄 Handoff requested")
        case string(stream.HandoffOccurred):
            fmt.Println("✅ Handoff completed")
        }
    }
}
```

**Available RunItemEvent names:**
- `MessageOutputCreated` - New message output
- `ToolCalled` - Tool was called
- `ToolOutput` - Tool output received
- `HandoffRequested` - Handoff requested
- `HandoffOccurred` - Handoff completed
- `ReasoningItemCreated` - Reasoning item created

### 3. Approval Required Events

`ApprovalRequiredEvent` is emitted when a tool call requires human approval before execution can proceed. The stream will terminate with a `ToolApprovalRequiredError` after this event. Use `Runner.Resume()` to continue.

```go
for event, err := range result.StreamEvents(ctx) {
    if err != nil {
        var approvalErr *agents.ToolApprovalRequiredError
        if errors.As(err, &approvalErr) {
            // Collect approval decisions, then resume
            result, err = runner.Resume(ctx, approvalErr.State, approvals)
        }
        return err
    }

    if approval, ok := event.(*stream.ApprovalRequiredEvent); ok {
        fmt.Printf("⚠️ Tool approval needed: %v\n", approval.Requests)
    }
}
```

### 4. Agent Updated Events

`AgentUpdatedEvent` provides updates when the current agent changes (e.g., during handoffs).

**Example:**

```go
for event, err := range result.StreamEvents(ctx) {
    if err != nil {
        return err
    }
    
    if agentUpdate, ok := event.(*stream.AgentUpdatedEvent); ok {
        fmt.Printf("👤 Now talking to: %v\n", agentUpdate.NewAgent)
    }
}
```

## Cancellation

You can cancel a streaming run using the `Cancel()` method with two modes:

### Immediate Cancellation

Stops immediately, cancels all tasks, and clears event queues:

```go
result.Cancel(stream.CancelImmediate)
```

### Graceful Cancellation

Completes the current turn gracefully before stopping:

```go
result.Cancel(stream.CancelAfterTurn)
```

**Example with context:**

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

result, err := runner.StreamWithResult(ctx, agent, messages)

// Cancel after 5 seconds
go func() {
    time.Sleep(5 * time.Second)
    result.Cancel(stream.CancelAfterTurn)
}()

for event, err := range result.StreamEvents(ctx) {
    // Handle events
}
```

## Examples

### Example 1: Basic Text Streaming

See [examples/18_streaming_basic](https://github.com/MitulShah1/openai-agents-go/tree/main/examples/18_streaming_basic) for a complete example of streaming text responses.

### Example 2: Function Argument Streaming

See [examples/19_streaming_function_args](https://github.com/MitulShah1/openai-agents-go/tree/main/examples/19_streaming_function_args) for real-time function call argument streaming.

### Example 3: Semantic Event Streaming

See [examples/20_streaming_semantic](https://github.com/MitulShah1/openai-agents-go/tree/main/examples/20_streaming_semantic) for high-level progress tracking with semantic events.

## Best Practices

### 1. Choose the Right Event Type

- **Raw events**: Use when you need token-by-token updates or real-time function arguments
- **Semantic events**: Use for high-level progress indicators and UI state updates
- **Agent events**: Use to track multi-agent conversations

### 2. Handle Errors Properly

Always check for errors in the event loop:

```go
for event, err := range result.StreamEvents(ctx) {
    if err != nil {
        log.Printf("Stream error: %v", err)
        return err
    }
    // Handle event
}
```

### 3. Use Context for Timeout Control

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

result, err := runner.StreamWithResult(ctx, agent, messages)
```

### 4. Filter Events for Performance

If you only need certain events, filter early:

```go
for event, err := range result.StreamEvents(ctx) {
    if err != nil {
        return err
    }
    
    // Ignore raw events if you only care about semantic events
    if _, ok := event.(*stream.RawResponseEvent); ok {
        continue
    }
    
    // Process only semantic events
    if runItem, ok := event.(*stream.RunItemEvent); ok {
        // Handle
    }
}
```

## Comparison with Python SDK

The Go streaming API is designed for parity with the Python SDK:

| Feature | Python | Go |
|---------|--------|-----|
| Raw events | `RawResponsesStreamEvent` | `*stream.RawResponseEvent` |
| Semantic events | `RunItemStreamEvent` | `*stream.RunItemEvent` |
| Approval events | — | `*stream.ApprovalRequiredEvent` |
| Agent updates | `AgentUpdatedStreamEvent` | `*stream.AgentUpdatedEvent` |
| Iteration | `async for event in result.stream_events()` | `for event, err := range result.StreamEvents(ctx)` |
| Cancellation | `result.cancel(mode="after_turn")` | `result.Cancel(stream.CancelAfterTurn)` |

## Technical Details

### Iterator Pattern

The Go SDK uses Go 1.23+ iterators (`iter.Seq2[Event, error]`) for idiomatic iteration:

```go
func (r *Result) StreamEvents(ctx context.Context) iter.Seq2[Event, error]
```

This provides:
- Automatic cleanup when iteration stops
- Context cancellation support
- Error handling in the iteration
- Memory-efficient streaming

### Thread Safety

All `stream.Result` methods are thread-safe and can be called concurrently.

### Event Ordering

Events include sequence numbers to maintain ordering even when generated concurrently.
