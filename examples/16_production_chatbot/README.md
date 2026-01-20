# Production Chatbot Example

This example demonstrates a complete production-ready configuration using all major features introduced in v0.3.0.

## Features Integrated

1.  **Database Persistence**: Uses the **SQLite** backend to persist conversation history across restarts. The database file (`production_sessions.db`) is automatically managed.
2.  **Guardrail Composition**: Implements a validation chain that runs the following checks sequentially:
    *   **Content Length**: Ensures messages are substantive (5-2000 chars).
    *   **Profanity Filter**: Blocks offensive language.
    *   **Secrets Detector**: Prevents PII/credentials leakage.
3.  **Metrics Collection**: Tracks the performance (latency) and success/failure rates of all guardrails using `usage.InMemoryMetrics`.
4.  **Multimodal Tools**: Demonstrates a tool (`generate_chart`) that returns **Image Content** directly to the agent, which the agent then references in its response.

## Running the Example

```bash
export OPENAI_API_KEY="your-api-key"
go run main.go
```

## Expected Output

1.  **Conversation**: The bot handles a multi-turn conversation about sales charts.
2.  **Persistence**: History is saved to `production_sessions.db`.
3.  **Guardrail Action**: The third message ("You are a damn fool") triggers the profanity filter and is blocked.
4.  **Metrics**: A clean summary of guardrail performance is printed at the end.
