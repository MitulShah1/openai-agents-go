# Basic Streaming Example

This example demonstrates basic streaming with the OpenAI Agents Go SDK.

## What it does

- Creates a simple storyteller agent
- Streams the response in real-time
- Prints text deltas as they arrive from the LLM
- Shows how to handle `RawResponseEvent` for text streaming

## Running the example

```bash
export OPENAI_API_KEY=your-api-key
go run main.go
```

## Expected output

You'll see the story being written in real-time, token by token, as the LLM generates it.

## Key concepts

- **StreamWithResult()**: Returns a `stream.Result` object
- **StreamEvents()**: Go 1.23+ iterator for consuming events
- **RawResponseEvent**: Raw events from the LLM in OpenAI Responses API format
- **response.output_text.delta**: Event type for text deltas
