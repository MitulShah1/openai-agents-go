# Parallel Tools Example

This example demonstrates the parallel tool execution feature in the OpenAI Agents Go SDK.

## Overview

The SDK supports three execution modes:
1. **Parallel Execution** (default) - Tools execute concurrently using goroutines
2. **Sequential Execution** - Tools execute one at a time
3. **Limited Concurrency** - Maximum N tools execute simultaneously

## Running the Example

### With OpenAI API

```bash
export OPENAI_API_KEY=your_api_key_here
go run main.go
```

### Without API Key (Demo Mode)

```bash
go run main.go
```

The demo mode simulates parallel vs sequential execution without making actual API calls.

## What It Demonstrates

1. **Parallel Execution** - 3 tools (weather, news, stocks) execute simultaneously
2. **Sequential Execution** - Same 3 tools execute one after another
3. **Limited Concurrency** - Maximum 2 tools execute at once
4. **Performance Comparison** - Shows speedup from parallel execution

## Expected Output

```
============================================================
🚀 PARALLEL TOOL EXECUTION DEMO
============================================================

📊 Demo 1: Parallel Execution (Default)
Tools will execute concurrently using goroutines
------------------------------------------------------------
🌤️  Fetching weather data...
📰 Fetching latest news...
📈 Fetching stock prices...

✅ Parallel Execution Complete!
⏱️  Duration: ~2s
📝 Response: [Agent response with all information]
🔧 Tools Called: 3

============================================================
📊 Demo 2: Sequential Execution
Tools will execute one at a time
------------------------------------------------------------
🌤️  Fetching weather data...
📰 Fetching latest news...
📈 Fetching stock prices...

✅ Sequential Execution Complete!
⏱️  Duration: ~6s
📝 Response: [Agent response with all information]
🔧 Tools Called: 3

============================================================
📈 PERFORMANCE COMPARISON
============================================================
Parallel:    ~2s
Sequential:  ~6s (3.0x slower)
Limited (2): ~4s (2.0x slower)

🚀 Parallel execution is 3.0x faster!
```

## Key Takeaways

- **Parallel execution** is ideal for I/O-bound tools (API calls, database queries)
- **Sequential execution** is better for stateful tools with dependencies
- **Concurrency limiting** prevents resource exhaustion with many tools
- Results maintain original tool call order regardless of execution mode
