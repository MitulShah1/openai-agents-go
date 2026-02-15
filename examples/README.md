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

### 04_advanced_handoffs - Advanced Handoffs
Complex handoff patterns including conditional transfers and data passing.

```bash
cd examples/04_advanced_handoffs
go run main.go
```

---

### 05_lifecycle_hooks - Lifecycle Hooks
Using OnBeforeRun and OnAfterRun hooks for logging and validation.

```bash
cd examples/05_lifecycle_hooks
go run main.go
```

**Demonstrates:**
- OnBeforeRun hook for initialization
- OnAfterRun hook for cleanup
- Error handling in hooks
- Execution time tracking

---

### 06_config_usage - Configuration & Usage Tracking
Advanced RunConfig usage and token tracking.

```bash
cd examples/06_config_usage
go run main.go
```

**Demonstrates:**
- RunConfig options (MaxTurns, Temperature, Timeout)
- Max turns enforcement
- Timeout handling
- Usage tracking across multiple calls
- Cost estimation

---

### 07_structured_output - Structured Output
Parsing agent responses into structured Go types.

```bash
cd examples/07_structured_output
go run main.go
```

---

### 08_complex_schema - Complex Schema
Handling complex JSON schemas for tools and outputs.

```bash
cd examples/08_complex_schema
go run main.go
```

---

### 09_guardrails_demo - Guardrails
Implementing safety guardrails for inputs and outputs.

```bash
cd examples/09_guardrails_demo
go run main.go
```

---

### 10_sessions_demo - Sessions
Managing conversation sessions and state.

```bash
cd examples/10_sessions_demo
go run main.go
```

---

### 11_advanced_v02 - Advanced Features (v0.2)
Showcase of advanced v0.2 features.

```bash
cd examples/11_advanced_v02
go run main.go
```

---

### 12_conversations_session - Conversation Sessions
Deep dive into conversation management.

```bash
cd examples/12_conversations_session
go run main.go
```

---

### 13_multimodal_tools - Multimodal Tools
Using tools with image inputs.

```bash
cd examples/13_multimodal_tools
go run main.go
```

---

### 14_guardrail_composition - Guardrail Composition
Composing multiple guardrails.

```bash
cd examples/14_guardrail_composition
go run main.go
```

---

### 15_session_backends - Session Backends
Using different backends (SQLite, etc.) for session storage.

```bash
cd examples/15_session_backends
go run main.go
```

---

### 16_tracing - Tracing
OpenTelemetry tracing support.

```bash
cd examples/16_tracing
go run main.go
```

---

### 17_production_chatbot - Production Chatbot
A complete chatbot example suitable for production.

```bash
cd examples/17_production_chatbot
go run main.go
```

---

### 18_type_safe_tools - Type Safe Tools
Using Go generics for type-safe tool definitions.

```bash
cd examples/18_type_safe_tools
go run main.go
```

---

### 19_streaming_basic - Basic Streaming
Streaming responses from the agent.

```bash
cd examples/19_streaming_basic
go run main.go
```

---

### 20_streaming_function_args - Streaming Function Args
Streaming tool calls and arguments.

```bash
cd examples/20_streaming_function_args
go run main.go
```

---

### 21_streaming_semantic - Semantic Streaming
Advanced streaming features.

```bash
cd examples/21_streaming_semantic
go run main.go
```

---

### 22_parallel_tools - Parallel Tools
Executing tools in parallel.

```bash
cd examples/22_parallel_tools
go run main.go
```

---

### 23_tool_approvals - Tool Approvals
Human-in-the-loop approval workflows for dangerous tool operations.

```bash
cd examples/23_tool_approvals
go run main.go
```

**Demonstrates:**
- Static approval (`NeedsApproval = true`)
- Dynamic approval (`ApprovalFunc` with conditional logic)
- Inline `ApprovalHandler` for synchronous approval decisions
- Pause/resume workflow with `ToolApprovalRequiredError` and `Runner.Resume()`
- `StreamEventApprovalRequired` for streaming approval
- Parallel batch safety (no partial execution)

---

### 24_prompts_demo - Prompts API
Using static and dynamic prompts for externally managed prompt configurations.

```bash
cd examples/24_prompts_demo
go run main.go
```

**Demonstrates:**
- Static prompts with ID, version, and variables
- Dynamic prompt selection based on context variables
- Prompt template variable substitution

---

### 25_multi_provider - Model Providers
Using multiple model providers in a single application.

```bash
cd examples/25_multi_provider
go run main.go
```

**Demonstrates:**
- Default provider (backward compatible `NewRunner`)
- Explicit provider with `NewRunnerWithProvider`
- Per-agent provider overrides
- `MultiProvider` prefix-based routing

---

## Quick Start

Run all examples:

```bash
export OPENAI_API_KEY="sk-..."

cd examples/01_basic && go run main.go
cd ../02_tools && go run main.go
cd ../03_handoffs && go run main.go
cd ../04_advanced_handoffs && go run main.go
cd ../05_lifecycle_hooks && go run main.go
cd ../06_config_usage && go run main.go
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
