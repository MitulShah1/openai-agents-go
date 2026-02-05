# Semantic Event Streaming Example

This example demonstrates high-level semantic event streaming, focusing on agent actions rather than raw LLM output.

## What it does

- Creates a multi-agent system (support + sales)
- Streams semantic events (tool calls, handoffs, messages)
- Ignores raw response events for cleaner progress tracking
- Shows agent transitions during handoffs

## Running the example

```bash
export OPENAI_API_KEY=your-api-key
go run main.go
```

## Expected output

You'll see high-level progress updates:
- "Agent is responding..."
- "Tool called"
- "Tool completed"
- "Handoff requested"
- "Handoff completed"
- "Now talking to: SalesAgent"

## Key concepts

- **RunItemEvent**: High-level semantic events
- **AgentUpdatedEvent**: Agent transition notifications
- **Event filtering**: Ignoring raw events for cleaner UX
- **Progress tracking**: Using semantic events for UI updates

## Use cases

This pattern is perfect for:
- Building user-facing progress indicators
- Showing "Agent is thinking..." states
- Displaying tool execution status
- Tracking multi-agent conversations
- Creating conversational UIs

## Comparison with raw events

- **Raw events**: Token-by-token, function argument deltas
- **Semantic events**: "Tool called", "Message created", "Handoff occurred"

Use semantic events when you want to show **what's happening** rather than **how it's being generated**.
