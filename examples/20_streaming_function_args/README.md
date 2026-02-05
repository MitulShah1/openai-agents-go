# Function Argument Streaming Example

This example demonstrates real-time streaming of function call arguments.

## What it does

- Creates an agent with a weather tool
- Streams function call arguments as they're generated
- Shows how arguments build up token-by-token
- Demonstrates the difference between raw events and semantic events

## Running the example

```bash
export OPENAI_API_KEY=your-api-key
go run main.go
```

## Expected output

You'll see:
1. Function arguments streaming in real-time (character by character)
2. Complete function call when done
3. Tool execution events
4. Final response

## Key concepts

- **response.function_call_arguments.delta**: Real-time argument streaming
- **response.output_item.done**: Function call completion
- **RunItemEvent**: High-level semantic events for tool calls
- **Argument accumulation**: Building complete arguments from deltas

## Why this matters

Real-time function argument streaming allows you to:
- Show progress to users during long function calls
- Validate arguments as they're being generated
- Provide early feedback if the LLM is generating incorrect arguments
