# OpenAI Agents Go - Examples

This directory contains examples demonstrating various features of the OpenAI Agents Go SDK.

## Prerequisites

```bash
export OPENAI_API_KEY="sk-..."
```

## Examples

### 01_basic - Hello World
The simplest possible agent - just a basic conversation with no tools.

```bash
cd examples/01_basic
go run main.go
```

**Demonstrates:**
- Creating a basic agent
- Running a simple conversation
- Viewing token usage

---

### 02_tools - Using Tools
Agents with multiple tools (weather and time).

```bash
cd examples/02_tools
go run main.go
```

**Demonstrates:**
- Defining custom tools
- Multiple tools on one agent
- Tool execution tracking
- Detailed execution trace

---

### 03_handoffs - Agent Handoffs
Multi-agent conversation with transfers between sales and support.

```bash
cd examples/03_handoffs
go run main.go
```

**Demonstrates:**
- Multiple specialized agents
- Agent-to-agent handoffs
- Bidirectional transfers
- Execution trace showing handoffs

---

### 04_lifecycle_hooks - Lifecycle Hooks
Using OnBeforeRun and OnAfterRun hooks for logging and validation.

```bash
cd examples/04_lifecycle_hooks
go run main.go
```

**Demonstrates:**
- OnBeforeRun hook for initialization
- OnAfterRun hook for cleanup
- Error handling in hooks
- Execution time tracking

---

### 05_config_usage - Configuration & Usage Tracking
Advanced RunConfig usage and token tracking.

```bash
cd examples/05_config_usage
go run main.go
```

**Demonstrates:**
- RunConfig options (MaxTurns, Temperature, Timeout)
- Max turns enforcement
- Timeout handling
- Usage tracking across multiple calls
- Cost estimation

---

## Quick Start

Run all examples:

```bash
export OPENAI_API_KEY="sk-..."

cd examples/01_basic && go run main.go
cd ../02_tools && go run main.go
cd ../03_handoffs && go run main.go
cd ../04_lifecycle_hooks && go run main.go
cd ../05_config_usage && go run main.go
```

## Common Patterns

### Creating an Agent
```go
agent := agents.NewAgent("MyAgent")
agent.Instructions = "You are a helpful assistant."
agent.Model = agents.DefaultModel
```

### Adding Tools
```go
tool := agents.FunctionTool(
    "tool_name",
    "Tool description",
    parametersSchema,
    callbackFunction,
)
agent.Tools = []agents.Tool{tool}
```

### Configuring Execution
```go
temp := 0.7
config := &agents.RunConfig{
    MaxTurns:    5,
    Temperature: &temp,
    Timeout:     2 * time.Minute,
}

result, err := runner.Run(ctx, agent, messages, nil, config)
```

### Agent Handoffs
```go
// In tool callback, return another agent
return supportAgent, nil
```

## Import Path

```go
import "github.com/MitulShah1/openai-agents-go"
```

## What's Next?

- **v0.2.0**: Guardrails and Sessions
- **v0.3.0**: Tracing and Observability
- **v0.4.0**: Streaming and Advanced Features

See [ROADMAP.md](../ROADMAP.md) for details.
