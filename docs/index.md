# OpenAI Agents Go SDK

[![CI](https://github.com/MitulShah1/openai-agents-go/workflows/CI/badge.svg)](https://github.com/MitulShah1/openai-agents-go/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/MitulShah1/openai-agents-go)](https://goreportcard.com/report/github.com/MitulShah1/openai-agents-go)
[![GoDoc](https://pkg.go.dev/badge/github.com/MitulShah1/openai-agents-go.svg)](https://pkg.go.dev/github.com/MitulShah1/openai-agents-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**A production-ready, type-safe Go framework for building intelligent multi-agent workflows with comprehensive security guardrails, persistent sessions, and enterprise-grade observability.**

!!! note "Community Project"
    This is an **unofficial** community-maintained SDK inspired by the official [OpenAI Agents Python SDK](https://github.com/openai/openai-agents-python) and [OpenAI Agents JavaScript SDK](https://github.com/openai/openai-agents-js), but is not affiliated with or endorsed by OpenAI.

## Why Use the Agents SDK?

### **🔒 Security First**

The SDK provides **comprehensive security guardrails** to protect your applications:

- **PII Detection & Protection**: Automatically detect and mask 10+ types of sensitive data including SSNs, credit cards, emails, phone numbers, and more
- **9+ Built-in Guardrails**: Rate limiting, profanity filtering, prompt injection detection, URL validation, secret scanning
- **Content Moderation**: OpenAI Moderation API integration with 13 safety categories
- **Tool Approvals**: Human-in-the-loop safety workflows for dangerous operations
- **Input/Output Validation**: Protect both incoming requests and agent responses

[:octicons-arrow-right-24: Learn about Guardrails](guardrails/index.md)

### **🚀 Production Ready**

Built for enterprise deployment with battle-tested features:

- **Zero Dependencies**: Core SDK has no external dependencies beyond the official OpenAI Go SDK
- **Type Safety**: Full compile-time type checking with Go generics for reliable code
- **Session Management**: Persistent conversations with multiple backends (memory, file, SQLite, PostgreSQL, Redis)
- **Observability**: OpenTelemetry tracing for monitoring agent behavior and performance
- **Error Handling**: Comprehensive error types with built-in retry strategies
- **High Test Coverage**: Extensively tested with >85% code coverage

### **⚡ Developer Experience**

Designed for Go developers who value clean, maintainable code:

- **Idiomatic Go**: Clean patterns following Go best practices
- **Rich Examples**: 25+ working examples covering all features
- **Comprehensive Docs**: Full documentation site with guides, tutorials, and API reference
- **Multi-Agent Workflows**: Seamless agent handoffs and composition
- **Streaming Support**: Real-time token-by-token and event-based streaming
- **Flexible Architecture**: Model provider abstraction for multi-LLM support

## Quick Example

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

## Installation

```bash
go get github.com/MitulShah1/openai-agents-go@latest
```

**Requirements**: Go 1.24 or higher

## Comprehensive Feature Set

| Feature | Status | Description |
|---------|--------|-------------|
| **Multi-Agent Workflows** | ✅ Complete | Compose and orchestrate multiple specialized agents |
| **Tool Integration** | ✅ Complete | Seamlessly call Go functions from agent responses |
| **Runtime Agent Skills** | ✅ Complete | Reusable capability bundles (instructions + tools + guardrails) |
| **Handoffs** | ✅ Complete | Dynamic agent-to-agent transfers during execution |
| **Structured Outputs** | ✅ Complete | Schema-validated JSON responses with fluent API |
| **Lifecycle Hooks** | ✅ Complete | Execute code before/after agent runs |
| **Context Variables** | ✅ Complete | Pass state between agents and tools |
| **Usage Tracking** | ✅ Complete | Monitor token consumption and costs |
| **Error Handling** | ✅ Enhanced | Comprehensive error types for robust applications |
| | | |
| **🛡️ Security & Safety** | | |
| **Guardrails Framework** | ✅ Complete | Extensible input/output validation system |
| **PII Detection** | ✅ Complete | 10+ detector types for sensitive data protection |
| **Content Moderation** | ✅ Complete | OpenAI Moderation API (13 categories) |
| **Rate Limiting** | ✅ Complete | Token bucket algorithm with burst support |
| **Profanity Filtering** | ✅ Complete | Customizable word list filtering |
| **Prompt Injection Detection** | ✅ Complete | Protect against malicious prompts |
| **URL Validation** | ✅ Complete | Domain allowlists and blocklists |
| **Secret Detection** | ✅ Complete | Find API keys, tokens, passwords |
| **Regex Filtering** | ✅ Complete | Custom pattern-based validation |
| **Guardrail Composition** | ✅ Complete | Chain guardrails sequentially or in parallel |
| **Tool Approvals** | ✅ Complete | Human-in-the-loop approval workflows |
| | | |
| **💾 Data & Persistence** | | |
| **Sessions Framework** | ✅ Complete | Persistent conversation history |
| **Memory Sessions** | ✅ Complete | In-memory storage for development |
| **File Sessions** | ✅ Complete | File-based persistence |
| **SQLite Sessions** | ✅ Complete | Embedded database backend |
| **PostgreSQL Sessions** | ✅ Complete | Production database backend (build tag) |
| **Redis Sessions** | ✅ Complete | Distributed cache backend (build tag) |
| **Cloud Sessions** | ✅ Complete | OpenAI Conversations API integration |
| **Session Encryption** | ✅ Complete | Secure sensitive data at rest |
| **Message Compaction** | ✅ Complete | Automatic history pruning |
| | | |
| **⚡ Performance & Scale** | | |
| **Streaming** | ✅ Complete | Token-by-token and object-based streaming |
| **Parallel Tools** | ✅ Complete | Concurrent tool execution |
| **Tracing & Observability** | ✅ Complete | OpenTelemetry integration |
| **Multimodal Support** | ✅ Complete | Images, files, and rich content |
| **Model Abstraction** | ✅ Complete | Pluggable LLM providers |
| **Prompts API** | ✅ Complete | Static and dynamic prompt management |
| | | |
| **🔮 Advanced Integrations** | | |
| **MCP Support** | ✅ Complete | Model Context Protocol integration |
| **Computer Use** | ✅ Complete | Browser/desktop automation interface |
| **Diff Tool** | ✅ Complete | Structured code change application |

See the [CHANGELOG](https://github.com/MitulShah1/openai-agents-go/blob/main/CHANGELOG.md) for detailed release history.

## Core Concepts

### Agents
LLMs configured with instructions, tools, and behavior settings. Each agent is a specialized worker optimized for specific tasks.

[:octicons-arrow-right-24: Learn about Agents](agents.md)

### Tools
Go functions that agents can call to perform actions, fetch data, or interact with external systems. Type-safe and easy to define.

[:octicons-arrow-right-24: Learn about Tools](tools.md)

### Guardrails
Input/output validation for safety and compliance. Protect against PII leaks, harmful content, rate abuse, and policy violations.

[:octicons-arrow-right-24: Learn about Guardrails](guardrails/index.md)

### Sessions
Persistent conversation history and memory management. Support multiple storage backends for different deployment scenarios.

[:octicons-arrow-right-24: Learn about Sessions](sessions/index.md)

### Handoffs
Transfer control between specialized agents dynamically based on task requirements or user requests.

[:octicons-arrow-right-24: Learn about Handoffs](handoffs.md)

### Structured Outputs
Schema-validated JSON responses for reliable parsing. Define expected output format and get compile-time safety.

[:octicons-arrow-right-24: Learn about Structured Outputs](structured_outputs.md)

### Streaming
Real-time token-by-token or event-based streaming for responsive user experiences.

[:octicons-arrow-right-24: Learn about Streaming](streaming.md)

### Tracing
OpenTelemetry integration for monitoring, debugging, and performance analysis.

[:octicons-arrow-right-24: Learn about Tracing](tracing.md)

## Next Steps

<div class="grid cards" markdown>

-   :material-clock-fast:{ .lg .middle } __Quickstart__

    ---

    Get started quickly with our comprehensive quickstart guide

    [:octicons-arrow-right-24: Get Started](quickstart.md)

-   :material-shield-check:{ .lg .middle } __Guardrails Guide__

    ---

    Learn how to protect your applications with PII detection and security guardrails

    [:octicons-arrow-right-24: Secure Your Agents](guardrails/index.md)

-   :material-book-open-variant:{ .lg .middle } __Core Concepts__

    ---

    Deep dive into agents, tools, handoffs, and more

    [:octicons-arrow-right-24: Learn Concepts](agents.md)

-   :material-code-braces:{ .lg .middle } __API Reference__

    ---

    Detailed API documentation for all types and functions

    [:octicons-arrow-right-24: API Docs](ref/index.md)

-   :material-github:{ .lg .middle } __Examples__

    ---

    Browse 25+ working examples on GitHub

    [:octicons-arrow-right-24: View Examples](examples.md)

-   :material-database:{ .lg .middle } __Sessions__

    ---

    Persist conversation history with flexible storage backends

    [:octicons-arrow-right-24: Manage Sessions](sessions/index.md)

</div>

## Comparison with Official SDKs

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
| MCP Support | ✅ | ✅ | ✅ |
| Computer Use | ✅ | ✅ | ✅ |
| Tool Approvals | ✅ | ❌ | ✅ |
| Prompts API | ✅ | ❌ | ✅ |
| Model Abstraction | ❌ | ❌ | ✅ |
| Voice Agents | ❌ | ✅ | 🔮 Planned |
| **Type Safety** | ⚠️ Runtime | ⚠️ TypeScript | ✅ **Compile-time** |
| **Zero Dependencies** | ❌ | ❌ | ✅ **(core only)** |
| **Performance** | ~ | ~ | ✅ **Native Go** |

## Community & Support

- 🐛 [Report Issues](https://github.com/MitulShah1/openai-agents-go/issues)
- 💬 [Discussions](https://github.com/MitulShah1/openai-agents-go/discussions)
- ⭐ Star the repo if you find it useful!
- 📖 [Full Documentation](https://mitulshah1.github.io/openai-agents-go/)

---

**Made with ❤️ by the Go community**
