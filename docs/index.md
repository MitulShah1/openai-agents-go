# OpenAI Agents Go SDK

[![CI](https://github.com/MitulShah1/openai-agents-go/workflows/CI/badge.svg)](https://github.com/MitulShah1/openai-agents-go/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/MitulShah1/openai-agents-go)](https://goreportcard.com/report/github.com/MitulShah1/openai-agents-go)
[![GoDoc](https://pkg.go.dev/badge/github.com/MitulShah1/openai-agents-go.svg)](https://pkg.go.dev/github.com/MitulShah1/openai-agents-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A lightweight, powerful Go framework for building multi-agent workflows with OpenAI's API. Build production-ready AI agents with tool calling, handoffs, structured outputs, and comprehensive guardrails.

!!! note "Community Project"
    This is an **unofficial** community-maintained SDK inspired by the official [OpenAI Agents Python SDK](https://github.com/openai/openai-agents-python) and [OpenAI Agents JavaScript SDK](https://github.com/openai/openai-agents-js), but is not affiliated with or endorsed by OpenAI.

## Why Use the Agents SDK?

The Agents SDK simplifies building multi-agent AI applications with these key advantages:

- **🚀 Zero Dependencies**: Core SDK has no external dependencies beyond the official OpenAI Go SDK
- **🔒 Type Safety**: Full Go compile-time type safety with generics support
- **🛡️ Production-Ready Guardrails**: 9+ built-in guardrails including PII detection, rate limiting, profanity filtering, and prompt injection detection
- **💾 Flexible Sessions**: Multiple storage backends (in-memory, file-based, cloud)
- **🔧 Powerful Tool System**: Seamlessly integrate Go functions as agent tools
- **🤝 Multi-Agent Workflows**: Hand off between specialized agents dynamically
- **📊 Structured Outputs**: Schema-validated JSON responses with fluent API
- **⚡ Performance**: Leverages Go's concurrency and performance characteristics

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

## Supported Features

| Feature | Status | Version |
|---------|--------|---------|
| Multi-Agent Workflows | ✅ Complete | v0.1.0 |
| Tool Integration | ✅ Complete | v0.1.0 |
| Handoffs | ✅ Complete | v0.1.0 |
| Structured Outputs | ✅ Complete | v0.1.0 |
| Lifecycle Hooks | ✅ Complete | v0.1.0 |
| **Guardrails** | ✅ Complete | v0.2.0+ |
| **Sessions** | ✅ Complete | v0.2.0+ |
| Context Variables | ✅ Complete | v0.1.0 |
| Usage Tracking | ✅ Complete | v0.1.0 |
| Error Handling | ✅ Enhanced | v0.2.3 |
| Streaming | 🔮 Planned | v0.4.0 |
| Tracing & Observability | 🔮 Planned | v0.3.0 |
| Database Session Backends | 🔮 Planned | v0.3.0 |

See the [CHANGELOG](https://github.com/MitulShah1/openai-agents-go/blob/main/CHANGELOG.md) for detailed release notes.

## Next Steps

<div class="grid cards" markdown>

-   :material-clock-fast:{ .lg .middle } __Quickstart__

    ---

    Get started quickly with our comprehensive quickstart guide

    [:octicons-arrow-right-24: Get Started](quickstart.md)

-   :material-book-open-variant:{ .lg .middle } __Core Concepts__

    ---

    Learn about agents, tools, handoffs, and more

    [:octicons-arrow-right-24: Learn Concepts](agents.md)

-   :material-code-braces:{ .lg .middle } __API Reference__

    ---

    Detailed API documentation for all types and functions

    [:octicons-arrow-right-24: API Docs](ref/index.md)

-   :material-github:{ .lg .middle } __Examples__

    ---

    Browse 11+ working examples on GitHub

    [:octicons-arrow-right-24: View Examples](examples.md)

</div>

## Comparison with Official SDKs

| Feature | [Python SDK](https://github.com/openai/openai-agents-python) | [JavaScript SDK](https://github.com/openai/openai-agents-js) | **Go SDK (This)** |
|---------|---------------|--------------|---------------| | Agents | ✅ | ✅ | ✅ |
| Tools | ✅ | ✅ | ✅ |
| Handoffs | ✅ | ✅ | ✅ |
| Structured Outputs | ✅ | ✅ | ✅ |
| Streaming | ✅ | ✅ | 🔮 Planned |
| Guardrails | ✅ | ✅ | ✅ (9+ guardrails) |
| Tracing | ✅ | ✅ | 🔮 Planned |
| Voice Agents | ❌ | ✅ | 🔮 Future |
| **Type Safety** | ⚠️ Runtime | ⚠️ TypeScript | ✅ Compile-time |
| **Zero Dependencies** | ❌ | ❌ | ✅ (core only) |

## Community & Support

- 🐛 [Report Issues](https://github.com/MitulShah1/openai-agents-go/issues)
- 💬 [Discussions](https://github.com/MitulShah1/openai-agents-go/discussions)
- 📖 [Roadmap](https://github.com/MitulShah1/openai-agents-go/blob/main/ROADMAP.md)
- ⭐ Star the repo if you find it useful!

---

**Made with ❤️ by the Go community**
