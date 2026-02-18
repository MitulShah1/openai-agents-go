# OpenAI Agents Go - Examples

This directory contains **26+ comprehensive examples** demonstrating all features of the OpenAI Agents Go SDK. Each example is a complete, runnable Go program showcasing specific capabilities.

## Prerequisites

```bash
# Set your OpenAI API key
export OPENAI_API_KEY="sk-..."

# Navigate to any example directory
cd examples/01_basic

# Run the example
go run main.go
```

---

## 📋 Quick Navigation

### 🚀 Getting Started
- [01_basic](#01_basic---hello-world) - Your first agent
- [02_tools](#02_tools---tool-integration) - Function calling basics
- [03_handoffs](#03_handoffs---agent-handoffs) - Multi-agent workflows

### 🛡️ Security & Safety
- [09_guardrails_demo](#09_guardrails_demo---comprehensive-guardrails) - **PII detection, rate limiting, content filtering**
- [14_guardrail_composition](#14_guardrail_composition---guardrail-composition) - **Chaining guardrails**
- [23_tool_approvals](#23_tool_approvals---human-in-the-loop-approvals) - **Human approval workflows**

### 💾 Data & Persistence
- [10_sessions_demo](#10_sessions_demo---session-management) - Conversation memory
- [15_session_backends](#15_session_backends---database-backends) - **SQLite, PostgreSQL, Redis**
- [12_conversations_session](#12_conversations_session---cloud-sessions) - OpenAI Conversations API

### ⚡ Advanced Features
- [07_structured_output](#07_structured_output---structured-json) - Schema-validated responses
- [19_streaming_basic](#19_streaming_basic---streaming-responses) - Real-time streaming
- [22_parallel_tools](#22_parallel_tools---parallel-execution) - Concurrent tool calls
- [16_tracing](#16_tracing---observability) - OpenTelemetry integration
- [26_multi_agent_trace](#26_multi_agent_trace---shared-trace-context) - **Multi-agent handoffs with shared trace**

### 🏭 Production Ready
- [17_production_chatbot](#17_production_chatbot---complete-chatbot) - **Full-featured chatbot**
- [18_type_safe_tools](#18_type_safe_tools---type-safety) - Go generics for tools

---

## 📚 Example Details

### 01_basic - Hello World

**The simplest possible agent** - A basic conversation with no tools.

```bash
cd examples/01_basic && go run main.go
```

**Demonstrates:**
- Creating a basic agent with instructions
- Running a simple conversation
- Viewing token usage and costs
- Basic error handling

**Use Case:** Starting point for any agent application

---

### 02_tools - Tool Integration

**Agents with function calling** - Multiple tools (weather and time lookup).

```bash
cd examples/02_tools && go run main.go
```

**Demonstrates:**
- Defining custom tools with parameters
- Multiple tools on a single agent
- Tool execution tracking and logging
- Detailed execution trace
- Parsing tool arguments

**Use Case:** Building agents that interact with external APIs or perform calculations

---

### 03_handoffs - Agent Handoffs

**Multi-agent collaboration** - Transfer conversations between specialized agents (sales and support).

```bash
cd examples/03_handoffs && go run main.go
```

**Demonstrates:**
- Multiple specialized agents
- Agent-to-agent handoffs
- Bidirectional transfers
- Execution trace showing agent transitions
- Context preservation across handoffs

**Use Case:** Complex workflows requiring different expertise areas

---

### 04_advanced_handoffs - Complex Handoff Patterns

**Advanced multi-agent patterns** - Conditional transfers, data filtering, and nested handoffs.

```bash
cd examples/04_advanced_handoffs && go run main.go
```

**Demonstrates:**
- Conditional handoff logic
- Input filtering for handoffs
- History nesting and isolation
- Dynamic handoff enablement
- Complex multi-agent orchestration

**Use Case:** Enterprise workflows with strict data isolation requirements

---

### 05_lifecycle_hooks - Execution Hooks

**Lifecycle management** - Using OnBeforeRun and OnAfterRun hooks.

```bash
cd examples/05_lifecycle_hooks && go run main.go
```

**Demonstrates:**
- OnBeforeRun hook for initialization
- OnAfterRun hook for cleanup
- Error handling in hooks
- Execution time tracking
- Logging and auditing patterns

**Use Case:** Monitoring, logging, and validation in production systems

---

### 06_config_usage - Advanced Configuration

**RunConfig options** - Control agent behavior with comprehensive settings.

```bash
cd examples/06_config_usage && go run main.go
```

**Demonstrates:**
- MaxTurns enforcement to prevent infinite loops
- Temperature and other model parameters
- Timeout handling for long-running tasks
- Usage tracking across multiple calls
- Cost estimation and budget control
- Debug mode for troubleshooting

**Use Case:** Production deployments requiring fine-grained control

---

### 07_structured_output - Structured JSON

**Type-safe responses** - Schema-validated JSON for reliable parsing.

```bash
cd examples/07_structured_output && go run main.go
```

**Demonstrates:**
- Defining JSON schemas with the fluent API
- Schema validation enforcement
- Parsing structured responses into Go types
- Required vs optional fields
- Type safety for agent outputs

**Use Case:** Applications requiring reliable data extraction and parsing

---

### 08_complex_schema - Advanced Schemas

**Complex data structures** - Nested objects, arrays, enums, and unions.

```bash
cd examples/08_complex_schema && go run main.go
```

**Demonstrates:**
- Nested object schemas
- Array and enum types
- Complex validation rules
- Schema composition
- Real-world data modeling

**Use Case:** E-commerce, data extraction, structured analysis

---

### 09_guardrails_demo - Comprehensive Guardrails

**Security and safety** - **PII detection, content moderation, rate limiting, and more**.

```bash
cd examples/09_guardrails_demo && go run main.go
```

**Demonstrates:**
- **PII detection (SSN, credit cards, emails, phones)**
- **URL filtering and domain validation**
- **Profanity filtering**
- **Rate limiting (token bucket)**
- **Custom regex patterns**
- Input and output validation
- Guardrail violation handling

**Use Case:** Production systems handling sensitive data or user-generated content

---

### 10_sessions_demo - Session Management

**Persistent conversations** - Managing conversation state across multiple runs.

```bash
cd examples/10_sessions_demo && go run main.go
```

**Demonstrates:**
- Creating and using sessions
- Conversation history persistence
- Context variables for state management
- Session isolation by user ID
- Memory vs file-based sessions

**Use Case:** Chatbots, customer support systems, conversational applications

---

### 11_advanced_v02 - Feature Showcase

**Combined capabilities** - Multiple features working together.

```bash
cd examples/11_advanced_v02 && go run main.go
```

**Demonstrates:**
- Guardrails + Sessions + Tools
- Real-world integration patterns
- Production-ready configuration

**Use Case:** Reference architecture for complex applications

---

### 12_conversations_session - Cloud Sessions

**OpenAI Conversations API** - Cloud-based session persistence.

```bash
cd examples/12_conversations_session && go run main.go
```

**Demonstrates:**
- Cloud session storage
- Cross-device conversation continuity
- Serverless deployment patterns
- No local storage required

**Use Case:** Distributed applications, mobile apps, multi-device experiences

---

### 13_multimodal_tools - Images and Files

**Rich content** - Tools that return images, files, and mixed content.

```bash
cd examples/13_multimodal_tools && go run main.go
```

**Demonstrates:**
- Returning images from tools
- File attachments in responses
- Mixed text and media content
- MIME type handling
- Base64 encoding for images

**Use Case:** Data visualization, document generation, image processing

---

### 14_guardrail_composition - Guardrail Composition

**Chaining guardrails** - Sequential and parallel guardrail execution.

```bash
cd examples/14_guardrail_composition && go run main.go
```

**Demonstrates:**
- Sequential guardrail chains (fail-fast)
- Parallel guardrail execution (check all)
- Combining different guardrail types
- Custom composition patterns
- Performance optimization

**Use Case:** Complex compliance requirements, layered security

---

### 15_session_backends - Database Backends

**Production persistence** - Using SQLite, PostgreSQL, and Redis for sessions.

```bash
cd examples/15_session_backends && go run main.go
```

**Demonstrates:**
- SQLite embedded database
- PostgreSQL for distributed systems
- Redis for caching and speed
- Backend selection strategies
- Connection pooling and management

**Use Case:** Scalable production deployments, high-availability systems

**Note:** PostgreSQL and Redis require build tags:
```bash
go run -tags postgres main.go
go run -tags redis main.go
```

---

### 16_tracing - Observability

**OpenTelemetry integration** - Monitoring and debugging agent behavior.

```bash
cd examples/16_tracing && go run main.go
```

**Demonstrates:**
- OpenTelemetry setup
- Tracing agent runs
- Tool call spans
- Performance monitoring
- Custom instrumentation

**Use Case:** Production monitoring, performance optimization, debugging

---

### 17_production_chatbot - Complete Chatbot

**Full-featured application** - A complete chatbot with all production features.

```bash
cd examples/17_production_chatbot && go run main.go
```

**Demonstrates:**
- Guardrails for safety
- Sessions for memory
- Tools for capabilities
- Error handling
- Logging and monitoring
- User experience best practices

**Use Case:** Production-ready chatbot template

---

### 18_type_safe_tools - Type Safety

**Go generics** - Type-safe tool definitions with compile-time checking.

```bash
cd examples/18_type_safe_tools && go run main.go
```

**Demonstrates:**
- Generic tool wrappers
- Compile-time type safety
- Automatic serialization
- Reduced runtime errors
- Clean tool interfaces

**Use Case:** Large codebases requiring strong type safety

---

### 19_streaming_basic - Streaming Responses

**Real-time output** - Token-by-token streaming for responsive UIs.

```bash
cd examples/19_streaming_basic && go run main.go
```

**Demonstrates:**
- Basic streaming setup
- Processing text deltas
- Handling streaming events
- Building responsive UIs
- Error handling in streams

**Use Case:** Chat applications, real-time interfaces

---

### 20_streaming_function_args - Streaming Tool Calls

**Streaming tool execution** - Watch tool calls and arguments stream in.

```bash
cd examples/20_streaming_function_args && go run main.go
```

**Demonstrates:**
- Streaming function arguments
- Partial argument processing
- Tool call progress tracking
- Advanced streaming patterns

**Use Case:** Long-running tool calls, progress indication

---

### 21_streaming_semantic - Semantic Streaming

**High-level streaming** - Event-based streaming with semantic chunks.

```bash
cd examples/21_streaming_semantic && go run main.go
```

**Demonstrates:**
- Semantic event types
- Structured streaming data
- Complex event handling
- State management in streams

**Use Case:** Complex UIs, multi-panel displays

---

### 22_parallel_tools - Parallel Execution

**Concurrent tool calls** - Execute multiple tools simultaneously.

```bash
cd examples/22_parallel_tools && go run main.go
```

**Demonstrates:**
- Parallel tool execution
- Concurrency configuration
- Performance optimization
- Safe concurrent access
- Result aggregation

**Use Case:** Performance-critical applications, batch processing

---

### 23_tool_approvals - Human-in-the-Loop Approvals

**Safety workflows** - **Require human approval before executing dangerous tools**.

```bash
cd examples/23_tool_approvals && go run main.go
```

**Demonstrates:**
- **Static approval (`NeedsApproval = true`)**
- **Dynamic approval (`ApprovalFunc` with conditional logic)**
- **Inline approval handler for synchronous decisions**
- **Pause/resume workflow with `Runner.Resume()`**
- **Streaming approval events**
- **Parallel batch safety (no partial execution)**

**Use Case:** File deletion, database operations, financial transactions, any dangerous operation

---

### 24_prompts_demo - Prompts API

**Centralized prompt management** - Using OpenAI Prompts API for externally managed prompts.

```bash
cd examples/24_prompts_demo && go run main.go
```

**Demonstrates:**
- Static prompts with ID and version
- Dynamic prompt selection based on context
- Prompt template variable substitution
- A/B testing prompts
- Centralized prompt governance

**Use Case:** Large teams, prompt versioning, governance, experimentation

---

### 25_multi_provider - Model Providers

**Multi-LLM support** - Using multiple model providers in one application.

```bash
cd examples/25_multi_provider && go run main.go
```

**Demonstrates:**
- Default OpenAI provider
- Explicit provider configuration with `NewRunnerWithProvider`
- Per-agent provider overrides
- `MultiProvider` for prefix-based routing (e.g., `anthropic/claude-3`)
- Mixing multiple LLM providers

**Use Case:** Cost optimization, feature access, vendor diversification

---

### 26_multi_agent_trace - Shared Trace Context

**Multi-agent handoffs with a single trace** - All agents in a handoff chain appear under one trace ID in your observability dashboard.

```bash
cd examples/26_multi_agent_trace && go run main.go
```

**Demonstrates:**
- Starting a shared trace before the first `runner.Run` call
- Carrying the trace through triage → billing / technical handoffs
- Custom workflow-level spans that wrap multiple agent runs
- Summary spans for recording cross-run metadata
- How the runner reuses an existing trace from context (no duplicate traces)

**Use Case:** Customer support pipelines, multi-step workflows, production observability

---

## 🎯 Common Patterns

### Creating an Agent

```go
agent := agents.NewAgent("MyAgent")
agent.Instructions = "You are a helpful assistant specialized in Go programming."
agent.Model = openai.ChatModelGPT4o
```

### Adding Tools

```go
import "github.com/MitulShah1/openai-agents-go/tools"

tool := tools.New(
    "tool_name",
    "Tool description for the model",
    map[string]any{
        "type": "object",
        "properties": map[string]any{
            "param": map[string]any{
                "type": "string",
                "description": "Parameter description",
            },
        },
        "required": []string{"param"},
    },
    func(args map[string]any, ctx tools.ContextVariables) (any, error) {
        // Tool implementation
        return "result", nil
    },
)

agent.Tools = []tools.Tool{tool}
```

### Using Guardrails

```go
import (
    "github.com/MitulShah1/openai-agents-go/guardrail/security"
    "github.com/MitulShah1/openai-agents-go/guardrail/content"
)

piiGuardrail := security.NewPII(security.WithTripwire(true))
profanityFilter := content.NewProfanityFilter()

result, err := runner.Run(
    ctx,
    agent,
    messages,
    nil,
    agents.WithGuardrails([]agents.Guardrail{piiGuardrail, profanityFilter}),
)
```

### Configuring Execution

```go
import "time"

temp := 0.7
config := &agents.RunConfig{
    MaxTurns:            5,
    Temperature:         &temp,
    Timeout:             2 * time.Minute,
    Debug:               true,
    MaxToolConcurrency:  3,
}

result, err := runner.Run(ctx, agent, messages, config, nil)
```

### Using Sessions

```go
import "github.com/MitulShah1/openai-agents-go/session"

sess := session.NewMemorySession()  // or NewFileSession, NewSQLiteSession, etc.

result, err := runner.Run(
    ctx,
    agent,
    messages,
    nil,
    agents.WithSession(sess, "user-123"),
)
```

### Agent Handoffs

```go
import "github.com/MitulShah1/openai-agents-go/handoff"

specialistAgent := agents.NewAgent("Specialist")
specialistAgent.Instructions = "You are an expert in specific domain"

mainAgent := agents.NewAgent("Main")
mainAgent.Tools = []tools.Tool{
    handoff.New(specialistAgent).ToTool(),
}
```

---

## 💡 Tips for Running Examples

### Environment Setup

```bash
# Required
export OPENAI_API_KEY="sk-..."

# Optional for specific examples
export ANTHROPIC_API_KEY="..."  # For multi-provider examples
export POSTGRES_URL="postgres://..."  # For PostgreSQL session backends
export REDIS_URL="localhost:6379"  # For Redis session backends
```

### Build Tags

Some examples require build tags for optional dependencies:

```bash
# PostgreSQL sessions
go run -tags postgres examples/15_session_backends/main.go

# Redis sessions
go run -tags redis examples/15_session_backends/main.go

# Both
go run -tags "postgres redis" examples/15_session_backends/main.go
```

### Troubleshooting

**API Key Issues:**
```bash
# Verify your API key is set
echo $OPENAI_API_KEY

# Test with a simple example
cd examples/01_basic && go run main.go
```

**Import Issues:**
```bash
# Ensure dependencies are downloaded
go mod download

# Update dependencies
go get -u github.com/MitulShah1/openai-agents-go@latest
```

**Build Errors:**
```bash
# Clean module cache
go clean -modcache

# Re-download dependencies
go mod tidy
```

---

## 📖 Learn More

### Documentation
- [Full Documentation Site](https://mitulshah1.github.io/openai-agents-go/)
- [Quickstart Guide](https://mitulshah1.github.io/openai-agents-go/quickstart/)
- [API Reference](https://pkg.go.dev/github.com/MitulShah1/openai-agents-go)

### Core Concepts
- [Agents](https://mitulshah1.github.io/openai-agents-go/agents/)
- [Tools](https://mitulshah1.github.io/openai-agents-go/tools/)
- [Guardrails](https://mitulshah1.github.io/openai-agents-go/guardrails/)
- [Sessions](https://mitulshah1.github.io/openai-agents-go/sessions/)

### Advanced Topics
- [Handoffs](https://mitulshah1.github.io/openai-agents-go/handoffs/)
- [Streaming](https://mitulshah1.github.io/openai-agents-go/streaming/)
- [Tracing](https://mitulshah1.github.io/openai-agents-go/tracing/)
- [Models & Providers](https://mitulshah1.github.io/openai-agents-go/models/)

---

## 🤝 Contributing Examples

Have an interesting use case? We welcome example contributions!

1. Create a new directory: `examples/XX_your_feature/`
2. Add a complete `main.go` with comments
3. Add a README.md explaining the example
4. Update this file with your example
5. Submit a pull request

---

## 📦 Import Path

All examples use the same import path:

```go
import (
    agents "github.com/MitulShah1/openai-agents-go"
    "github.com/MitulShah1/openai-agents-go/tools"
    "github.com/MitulShah1/openai-agents-go/guardrail/security"
    "github.com/MitulShah1/openai-agents-go/session"
    // ... etc
)
```

---

**Made with ❤️ by the Go community**
