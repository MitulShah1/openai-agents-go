# OpenAI Agents Go SDK

[![CI](https://github.com/MitulShah1/openai-agents-go/workflows/CI/badge.svg)](https://github.com/MitulShah1/openai-agents-go/actions)
[![CodeQL](https://github.com/MitulShah1/openai-agents-go/actions/workflows/codeql.yml/badge.svg)](https://github.com/MitulShah1/openai-agents-go/actions/workflows/codeql.yml)
[![codecov](https://codecov.io/gh/MitulShah1/openai-agents-go/branch/main/graph/badge.svg)](https://codecov.io/gh/MitulShah1/openai-agents-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/MitulShah1/openai-agents-go)](https://goreportcard.com/report/github.com/MitulShah1/openai-agents-go)
[![GoDoc](https://pkg.go.dev/badge/github.com/MitulShah1/openai-agents-go.svg)](https://pkg.go.dev/github.com/MitulShah1/openai-agents-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

> **Community Project**: This is an **unofficial** community-maintained Go SDK for building AI agents with OpenAI's API. Inspired by the official [Python](https://github.com/openai/openai-agents-python) and [JavaScript](https://github.com/openai/openai-agents-js) SDKs, but not affiliated with or endorsed by OpenAI.

**A production-ready, type-safe Go framework for building intelligent multi-agent workflows with comprehensive security guardrails, persistent sessions, and enterprise-grade observability.**

---

## 🎯 Why Choose This SDK?

### **🔒 Security First**
- **PII Detection & Protection**: Automatically detect and mask sensitive data (SSNs, credit cards, emails, phone numbers)
- **9+ Built-in Guardrails**: Rate limiting, profanity filtering, prompt injection detection, URL validation, secret scanning
- **Content Moderation**: OpenAI Moderation API integration with 13 safety categories
- **Tool Approvals**: Human-in-the-loop safety for dangerous operations

### **🚀 Production Ready**
- **Zero Dependencies**: Core SDK has no external dependencies (beyond OpenAI Go SDK)
- **Type Safety**: Full compile-time type checking with Go generics
- **Session Management**: Persistent conversations with multiple backends (memory, file, SQLite, PostgreSQL, Redis)
- **Observability**: OpenTelemetry tracing for monitoring agent behavior
- **Error Handling**: Comprehensive error types with retry strategies

### **⚡ Developer Experience**
- **Idiomatic Go**: Clean, maintainable code following Go best practices
- **Rich Examples**: 25+ working examples covering all features
- **Comprehensive Docs**: Full documentation site with guides and API reference
- **Multi-Agent Workflows**: Seamless agent handoffs and composition
- **Streaming Support**: Real-time token-by-token and event-based streaming

---

## 📦 Installation

```bash
go get github.com/MitulShah1/openai-agents-go@latest
```

**Requirements**: Go 1.24 or higher

---

## ⚡ Quick Start

### Hello World

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    agents "github.com/MitulShah1/openai-agents-go"
    "github.com/openai/openai-go"
    "github.com/openai/openai-go/option"
)

func main() {
    // Initialize OpenAI client
    client := openai.NewClient(option.WithAPIKey(os.Getenv("OPENAI_API_KEY")))
    runner := agents.NewRunner(&client)

    // Create an agent
    agent := agents.NewAgent("Assistant")
    agent.Instructions = "You are a helpful assistant"

    // Run the agent
    messages := []openai.ChatCompletionMessageParamUnion{
        openai.UserMessage("Write a haiku about Go programming"),
    }

    result, err := runner.Run(context.Background(), agent, messages, nil, nil)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(result.FinalOutput)
}
```

---

## 🛡️ Guardrails: Protect Your AI Applications

The SDK provides comprehensive security and safety guardrails to protect your applications from malicious inputs, data leaks, and policy violations.

### **PII Detection & Protection**

Automatically detect and mask 10+ types of sensitive data:

```go
package main

import (
    "github.com/MitulShah1/openai-agents-go/guardrail/security"
)

func main() {
    // Create PII guardrail
    piiGuardrail := security.NewPII(
        security.WithTripwire(true),        // Block instead of mask
        security.WithDetectors(              // Customize detectors
            security.PIITypeSSN,
            security.PIITypeCreditCard,
            security.PIITypeEmail,
            security.PIITypePhone,
        ),
    )

    // Run agent with PII protection
    result, err := runner.Run(
        ctx,
        agent,
        messages,
        nil,
        agents.WithGuardrails([]agents.Guardrail{piiGuardrail}),
    )
}
```

**Supported PII Types:**
- Social Security Numbers (SSN)
- Credit Card Numbers
- Email Addresses
- Phone Numbers
- IP Addresses
- API Keys & Secrets
- Dates of Birth
- Passport Numbers
- Driver's License Numbers
- Medical Record Numbers

### **Content Moderation**

Block harmful content using OpenAI's Moderation API:

```go
import "github.com/MitulShah1/openai-agents-go/guardrail/moderation"

moderationGuardrail := moderation.NewModeration(
    client,
    moderation.WithThresholds(0.7), // Custom threshold
)
```

**Detects 13 Categories:**
- Hate speech
- Harassment
- Self-harm
- Sexual content
- Violence
- And more...

### **Rate Limiting**

Protect against abuse and control costs:

```go
import "github.com/MitulShah1/openai-agents-go/guardrail/ratelimit"

rateLimiter := ratelimit.NewRateLimiter(
    100,                    // 100 requests
    time.Minute,            // per minute
    ratelimit.WithBurst(10), // Allow bursts
)
```

### **Additional Guardrails**

```go
import (
    "github.com/MitulShah1/openai-agents-go/guardrail/content"
    "github.com/MitulShah1/openai-agents-go/guardrail/security"
)

// Block profanity
profanityFilter := content.NewProfanityFilter()

// Detect prompt injection attacks
injectionDetector := security.NewPromptInjection()

// Validate URLs
urlFilter := security.NewURLFilter(
    security.WithAllowedDomains([]string{"example.com"}),
)

// Custom regex validation
customFilter := content.NewRegexFilter(
    `\b[A-Z]{3}\d{3}\b`, // Match pattern
    true,                 // Block on match
)
```

### **Guardrail Composition**

Chain multiple guardrails with sequential or parallel execution:

```go
import "github.com/MitulShah1/openai-agents-go/guardrail"

// Sequential: Stop at first violation
sequential := guardrail.NewSequence(
    piiGuardrail,
    moderationGuardrail,
    profanityFilter,
)

// Parallel: Check all simultaneously
parallel := guardrail.NewParallel(
    rateLimiter,
    urlFilter,
    injectionDetector,
)

// Combine
combined := guardrail.NewSequence(parallel, sequential)
```

[**→ Full Guardrails Documentation**](https://mitulshah1.github.io/openai-agents-go/guardrails/)

---

## 💾 Sessions: Persistent Conversation Memory

Maintain conversation context across multiple interactions with flexible storage backends.

### **In-Memory Sessions** (Development/Testing)

```go
import "github.com/MitulShah1/openai-agents-go/session"

sess := session.NewMemorySession()

result, _ := runner.Run(
    ctx,
    agent,
    messages,
    nil,
    agents.WithSession(sess, "user-123"),
)
```

### **File-Based Sessions** (Simple Production)

```go
sess := session.NewFileSession("./sessions")
```

### **Database Sessions** (Enterprise)

```go
// SQLite (built-in)
sess, _ := session.NewSQLiteSession("./sessions.db")

// PostgreSQL (requires build tag: -tags postgres)
sess, _ := session.NewPostgresSession(
    "postgres://user:pass@localhost/db",
)

// Redis (requires build tag: -tags redis)
sess, _ := session.NewRedisSession("localhost:6379", 0)
```

### **Session Features**

- **Message History**: Automatic conversation tracking
- **Context Variables**: Pass state between agent runs
- **Encryption**: Secure sensitive data at rest
- **Compaction**: Automatic history pruning
- **Cloud Backend**: OpenAI Conversations API integration

[**→ Full Sessions Documentation**](https://mitulshah1.github.io/openai-agents-go/sessions/)

---

## 🔧 Tools: Extend Agent Capabilities

Agents can call Go functions as tools to perform actions, fetch data, or interact with external systems.

### **Basic Tool**

```go
import "github.com/MitulShah1/openai-agents-go/tools"

weatherTool := tools.New(
    "get_weather",
    "Get current weather for a city",
    map[string]any{
        "type": "object",
        "properties": map[string]any{
            "city": map[string]any{
                "type":        "string",
                "description": "City name",
            },
        },
        "required": []string{"city"},
    },
    func(args map[string]any, ctx tools.ContextVariables) (any, error) {
        city := args["city"].(string)
        // Call weather API
        return fmt.Sprintf("Weather in %s: Sunny, 72°F", city), nil
    },
)

agent.Tools = []tools.Tool{weatherTool}
```

### **Tool Approvals** (Human-in-the-Loop)

Require approval before executing dangerous operations:

```go
deleteFileTool := tools.New(
    "delete_file",
    "Delete a file from the system",
    schema,
    deleteFileFunc,
)

// Static approval requirement
deleteFileTool.NeedsApproval = true

// Or dynamic approval logic
deleteFileTool.ApprovalFunc = func(args map[string]any) (bool, error) {
    filename := args["filename"].(string)
    return strings.HasPrefix(filename, "/tmp/"), nil // Only approve /tmp files
}

// Handle approvals inline
result, err := runner.Run(
    ctx,
    agent,
    messages,
    nil,
    agents.WithApprovalHandler(func(req tools.ApprovalRequest) (bool, error) {
        fmt.Printf("Approve %s with args %v? (y/n): ", req.Name, req.Args)
        var response string
        fmt.Scanln(&response)
        return response == "y", nil
    }),
)

// Or pause/resume workflow
if errors.As(err, &agents.ToolApprovalRequiredError{}) {
    approvalErr := err.(agents.ToolApprovalRequiredError)
    
    // Review requests
    for i, req := range approvalErr.Requests {
        approved := getUserApproval(req)
        approvalErr.Requests[i].Approved = approved
    }
    
    // Resume execution
    result, err = runner.Resume(ctx, approvalErr.State, approvalErr.Requests)
}
```

### **Multimodal Tools**

Return images, files, and rich content:

```go
func generateChart(args map[string]any, ctx tools.ContextVariables) (any, error) {
    // Generate chart image
    imageData := createChartImage(args)
    
    return tools.MultimodalResponse{
        Text: "Here's your chart:",
        Images: []tools.ImageContent{
            {Data: imageData, MimeType: "image/png"},
        },
    }, nil
}
```

### **Parallel Tool Execution**

Execute multiple tools concurrently for better performance:

```go
config := &agents.RunConfig{
    MaxToolConcurrency: 5, // Run up to 5 tools in parallel
}
```

[**→ Full Tools Documentation**](https://mitulshah1.github.io/openai-agents-go/tools/)

---

## 🤝 Multi-Agent Workflows

Create specialized agents and orchestrate handoffs for complex workflows.

### **Agent Handoffs**

```go
import "github.com/MitulShah1/openai-agents-go/handoff"

// Specialized agents
weatherAgent := agents.NewAgent("Weather Specialist")
weatherAgent.Instructions = "You are an expert at weather information"
weatherAgent.Tools = []tools.Tool{weatherTool}

newsAgent := agents.NewAgent("News Specialist")
newsAgent.Instructions = "You provide the latest news"
newsAgent.Tools = []tools.Tool{newsTool}

// Main coordinator agent
mainAgent := agents.NewAgent("Main Assistant")
mainAgent.Instructions = `You coordinate with specialists.
Transfer weather questions to the weather specialist.
Transfer news questions to the news specialist.`
mainAgent.Tools = []tools.Tool{
    handoff.New(weatherAgent).ToTool(),
    handoff.New(newsAgent).ToTool(),
}

// Handoffs happen automatically during execution
result, _ := runner.Run(ctx, mainAgent, messages, nil, nil)
```

### **Handoff Features**

- **Input Filtering**: Control what data passes to target agent
- **History Nesting**: Isolate or merge conversation context
- **Dynamic Enablement**: Conditional handoff availability
- **Bidirectional Transfers**: Agents can hand back control

[**→ Full Handoffs Documentation**](https://mitulshah1.github.io/openai-agents-go/handoffs/)

---

## 📊 Structured Outputs

Get schema-validated JSON responses for reliable parsing and type safety.

### **Simple Structured Output**

```go
import "github.com/MitulShah1/openai-agents-go/jsonschema"

// Define schema
schema := jsonschema.Object().
    WithProperty("answer", jsonschema.Integer()).
    WithProperty("reasoning", jsonschema.String()).
    WithRequired("answer", "reasoning")

// Create agent with structured output
agent := agents.NewAgent("Math Tutor")
agent.ResponseFormat = jsonschema.JSONSchema("math_response", schema)

messages := []openai.ChatCompletionMessageParamUnion{
    openai.UserMessage("What is 15 + 27?"),
}

result, _ := runner.Run(ctx, agent, messages, nil, nil)

// Parse response
type MathResponse struct {
    Answer    int    `json:"answer"`
    Reasoning string `json:"reasoning"`
}

var response MathResponse
json.Unmarshal([]byte(result.FinalOutput), &response)
```

### **Complex Schemas**

```go
// Nested objects
personSchema := jsonschema.Object().
    WithProperty("name", jsonschema.String()).
    WithProperty("age", jsonschema.Integer()).
    WithProperty("address", jsonschema.Object().
        WithProperty("street", jsonschema.String()).
        WithProperty("city", jsonschema.String()).
        WithRequired("street", "city"),
    ).
    WithRequired("name", "age")

// Arrays
listSchema := jsonschema.Object().
    WithProperty("items", jsonschema.Array(jsonschema.String())).
    WithRequired("items")

// Enums
statusSchema := jsonschema.String().
    WithEnum([]string{"pending", "approved", "rejected"})
```

[**→ Full Structured Outputs Documentation**](https://mitulshah1.github.io/openai-agents-go/structured_outputs/)

---

## 📡 Streaming

Stream responses token-by-token or event-by-event for real-time applications.

### **Basic Streaming**

```go
stream := runner.Stream(ctx, agent, messages, nil, nil)

for event := range stream.Events() {
    switch e := event.(type) {
    case *stream.TextDeltaEvent:
        fmt.Print(e.Delta)
    case *stream.TextDoneEvent:
        fmt.Println("\nComplete:", e.Text)
    case *stream.FinalEvent:
        fmt.Println("Final output:", e.Result.FinalOutput)
    }
}
```

### **Advanced Streaming**

```go
streamResult := runner.StreamWithResult(ctx, agent, messages, nil, nil)

// Process events
for event := range streamResult.StreamEvents() {
    switch e := event.(type) {
    case *stream.AgentMessageEvent:
        fmt.Println("Agent:", e.Message)
    case *stream.ToolCallEvent:
        fmt.Println("Calling tool:", e.Name)
    case *stream.ToolResultEvent:
        fmt.Println("Tool result:", e.Result)
    case *stream.ApprovalRequiredEvent:
        fmt.Println("Approval needed:", e.Requests)
    }
}

// Get final result
result := streamResult.Result()
```

[**→ Full Streaming Documentation**](https://mitulshah1.github.io/openai-agents-go/streaming/)

---

## 🔍 Observability & Tracing

![Tracing Dashboard](./docs/assets/tracing.png)

Monitor agent behavior with OpenTelemetry integration.

```go
import (
    "github.com/MitulShah1/openai-agents-go/tracing"
    "go.opentelemetry.io/otel"
)

// Enable tracing
tracer := otel.Tracer("my-app")
runner := agents.NewRunner(&client)

// Run with tracing context
ctx := tracing.StartAgentSpan(ctx, tracer, "main-agent")
defer tracing.EndSpan(ctx)

result, _ := runner.Run(ctx, agent, messages, nil, nil)
```

**Traced Operations:**
- Agent runs (start, end, turns)
- Tool calls (arguments, results, duration)
- Session operations (load, save)
- Guardrail checks (input, output, violations)

[**→ Full Tracing Documentation**](https://mitulshah1.github.io/openai-agents-go/tracing/)

---

## 🎓 Examples

The repository includes **25+ comprehensive examples** covering all features:

### **Getting Started**
- [01_basic](./examples/01_basic) - Hello world agent
- [02_tools](./examples/02_tools) - Tool integration
- [03_handoffs](./examples/03_handoffs) - Agent handoffs
- [05_lifecycle_hooks](./examples/05_lifecycle_hooks) - Execution hooks
- [06_config_usage](./examples/06_config_usage) - Run configuration

### **Security & Safety**
- [09_guardrails_demo](./examples/09_guardrails_demo) - Comprehensive guardrails
- [14_guardrail_composition](./examples/14_guardrail_composition) - Chaining guardrails
- [23_tool_approvals](./examples/23_tool_approvals) - Human-in-the-loop approvals

### **Data & Persistence**
- [10_sessions_demo](./examples/10_sessions_demo) - Session management
- [15_session_backends](./examples/15_session_backends) - Database backends
- [12_conversations_session](./examples/12_conversations_session) - Cloud sessions

### **Advanced Features**
- [07_structured_output](./examples/07_structured_output) - Structured JSON
- [19_streaming_basic](./examples/19_streaming_basic) - Streaming responses
- [22_parallel_tools](./examples/22_parallel_tools) - Parallel execution
- [16_tracing](./examples/16_tracing) - OpenTelemetry tracing
- [24_prompts_demo](./examples/24_prompts_demo) - Prompts API
- [25_multi_provider](./examples/25_multi_provider) - Multiple LLM providers

### **Production Ready**
- [17_production_chatbot](./examples/17_production_chatbot) - Complete chatbot
- [18_type_safe_tools](./examples/18_type_safe_tools) - Type-safe tools

**Run any example:**

```bash
export OPENAI_API_KEY="your-key-here"
cd examples/09_guardrails_demo
go run main.go
```

[**→ Full Examples Guide**](./examples/README.md)

---

## 📚 Documentation

### **Getting Started**
- 📖 [**Documentation Site**](https://mitulshah1.github.io/openai-agents-go/) - Complete guides
- 🚀 [Quickstart Guide](https://mitulshah1.github.io/openai-agents-go/quickstart/)
- 📋 [Examples Directory](./examples)

### **Core Concepts**
- 🤖 [Agents](https://mitulshah1.github.io/openai-agents-go/agents/)
- 🔧 [Tools](https://mitulshah1.github.io/openai-agents-go/tools/)
- 🤝 [Handoffs](https://mitulshah1.github.io/openai-agents-go/handoffs/)
- 📊 [Structured Outputs](https://mitulshah1.github.io/openai-agents-go/structured_outputs/)

### **Production Features**
- 🛡️ [Guardrails](https://mitulshah1.github.io/openai-agents-go/guardrails/)
- 💾 [Sessions](https://mitulshah1.github.io/openai-agents-go/sessions/)
- 📡 [Streaming](https://mitulshah1.github.io/openai-agents-go/streaming/)
- 🔍 [Tracing](https://mitulshah1.github.io/openai-agents-go/tracing/)

### **Advanced Topics**
- 🎯 [Models & Providers](https://mitulshah1.github.io/openai-agents-go/models/)
- 📝 [Prompts API](https://mitulshah1.github.io/openai-agents-go/prompts/)
- 🏃 [Running Agents](https://mitulshah1.github.io/openai-agents-go/running_agents/)

### **Reference**
- 📚 [API Documentation](https://pkg.go.dev/github.com/MitulShah1/openai-agents-go) - GoDoc
- 📝 [CHANGELOG](./CHANGELOG.md) - Release history

---

## 🚀 Feature Comparison

| Feature | Python SDK | JavaScript SDK | **Go SDK (This)** |
|---------|------------|----------------|-------------------|
| Agents | ✅ | ✅ | ✅ |
| Tools | ✅ | ✅ | ✅ |
| Handoffs | ✅ | ✅ | ✅ |
| Structured Outputs | ✅ | ✅ | ✅ |
| Streaming | ✅ | ✅ | ✅ |
| **Guardrails** | ✅ | ✅ | ✅ **9+ types** |
| **PII Detection** | ✅ | ✅ | ✅ **10+ detectors** |
| Tracing | ✅ | ✅ | ✅ OpenTelemetry |
| Tool Approvals | ✅ | ❌ | ✅ |
| Prompts API | ✅ | ❌ | ✅ |
| Model Abstraction | ❌ | ❌ | ✅ |
| MCP Support | ✅ | ✅ | ✅ |
| Computer Use | ✅ | ✅ | ✅ |
| Voice Agents | ❌ | ✅ | 🔮 Planned |
| **Type Safety** | ⚠️ Runtime | ⚠️ TypeScript | ✅ **Compile-time** |
| **Zero Dependencies** | ❌ | ❌ | ✅ **Core only** |
| **Performance** | ~ | ~ | ✅ **Native Go** |

---

## 🏗️ Development

### **Prerequisites**

- Go 1.24+
- golangci-lint
- goimports

### **Running Tests**

```bash
# Run all tests
go test -v ./...

# With coverage
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run linter
golangci-lint run ./...

# Or use Makefile
make check
make test
make lint
```

### **Building Documentation**

```bash
# Install MkDocs
pip install mkdocs-material

# Serve locally
mkdocs serve

# Build site
mkdocs build
```

---

## 🤝 Contributing

Contributions are welcome! This is a community project.

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Write tests for your changes
4. Ensure all tests pass (`make check`)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

**Please ensure:**
- All tests pass
- Code is formatted (`gofmt`, `goimports`)
- Documentation is updated
- Examples demonstrate new features

---

## 🙏 Acknowledgements

This project is inspired by:
- [OpenAI Agents Python SDK](https://github.com/openai/openai-agents-python)
- [OpenAI Agents JavaScript SDK](https://github.com/openai/openai-agents-js)

Built with:
- [OpenAI Go SDK](https://github.com/openai/openai-go)

---

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

---

## 💬 Support & Community

- 🐛 [Report Issues](https://github.com/MitulShah1/openai-agents-go/issues)
- 💬 [Discussions](https://github.com/MitulShah1/openai-agents-go/discussions)
- ⭐ Star the repo if you find it useful!
- 📖 [Full Documentation](https://mitulshah1.github.io/openai-agents-go/)

---

<div align="center">

**Made with ❤️ by the Go community**

[Documentation](https://mitulshah1.github.io/openai-agents-go/) • 
[Examples](./examples) • 
[API Reference](https://pkg.go.dev/github.com/MitulShah1/openai-agents-go) •
[GitHub](https://github.com/MitulShah1/openai-agents-go)

</div>
