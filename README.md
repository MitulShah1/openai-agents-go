# OpenAI Agents Go SDK

[![CI](https://github.com/MitulShah1/openai-agents-go/workflows/CI/badge.svg)](https://github.com/MitulShah1/openai-agents-go/actions)
[![CodeQL](https://github.com/MitulShah1/openai-agents-go/actions/workflows/codeql.yml/badge.svg)](https://github.com/MitulShah1/openai-agents-go/actions/workflows/codeql.yml)
[![codecov](https://codecov.io/gh/MitulShah1/openai-agents-go/branch/main/graph/badge.svg)](https://codecov.io/gh/MitulShah1/openai-agents-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/MitulShah1/openai-agents-go)](https://goreportcard.com/report/github.com/MitulShah1/openai-agents-go)
[![GoDoc](https://pkg.go.dev/badge/github.com/MitulShah1/openai-agents-go.svg)](https://pkg.go.dev/github.com/MitulShah1/openai-agents-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

> **Note**: This is an **unofficial** community-maintained Go SDK for building AI agents with OpenAI's API. It is inspired by the official [OpenAI Agents Python SDK](https://github.com/openai/openai-agents-python) and [OpenAI Agents JavaScript SDK](https://github.com/openai/openai-agents-js), but is not affiliated with or endorsed by OpenAI.

A lightweight, powerful Go framework for building multi-agent workflows with OpenAI's API. Build production-ready AI agents with tool calling, handoffs, structured outputs, and more.

---

## Core Concepts

1. **Agents**: LLMs configured with instructions, tools, and behavior settings
2. **Tools**: Functions that agents can call to perform actions
3. **Runtime Skills**: Reusable capability bundles (instructions + tools + guardrails) on agents
4. **Handoffs**: Transfer control between specialized agents dynamically
5. **Structured Outputs**: Schema-validated JSON responses for reliable parsing
6. **Guardrails**: Input/output validation for safety and compliance
7. **Sessions**: Persistent conversation history and memory management
8. **Run Configuration**: Control execution flow with max turns, timeouts, and more

Explore the [`examples/`](./examples) directory to see the SDK in action.

---

## Supported Features

- ✅ **Multi-Agent Workflows**: Compose and orchestrate multiple agents
- ✅ **Tool Integration**: Seamlessly call Go functions from agent responses
- ✅ **Runtime Agent Skills**: Reusable capability composition via `Skill` + `AddSkill`
- ✅ **Handoffs**: Dynamic agent-to-agent transfers during execution
- ✅ **Structured Outputs**: Schema-validated JSON responses with fluent API
- ✅ **Guardrails**: Input/output validation with PII detection, URL filtering, custom regex, and OpenAI Moderation API
- ✅ **Sessions**: Persistent conversation history with memory, file-based, and cloud-based (Conversations API) storage
- ✅ **Lifecycle Hooks**: Execute code before/after agent runs
- ✅ **Context Variables**: Pass state between agents and tools
- ✅ **Usage Tracking**: Monitor token consumption and costs
- ✅ **Error Handling**: Comprehensive error types for debugging
- ✅ **Type Safety**: Full Go type safety with generics support
- ✅ **Multimodal Tools**: Return images and files from tools (v0.3.0)
- ✅ **Guardrail Composition**: Chain guardrails with sequential/parallel execution, async validation, and metrics (v0.3.0)
- ✅ **Database Sessions**: SQLite support for production-ready persistence (v0.3.0)
- ✅ **Handoff Parity**: Full feature parity with Python SDK (v0.3.5)
- ✅ **Streaming**: Token-by-token and object-based streaming (v0.4.0)
- ✅ **Tracing & Observability**: OpenTelemetry integration (v0.4.0)
- ✅ **Advanced DB Backends**: Redis, PostgreSQL (v0.4.0)
- ✅ **Tool Approvals**: Human-in-the-loop approval workflows with pause/resume (v0.6.0)

---


## Project Agent Skills

This repo includes project-local skills in `.agents/skills`.
These are for **developer workflow guidance** when using Codex in this repository.

See [.agents/skills/README.md](.agents/skills/README.md) for available project skills.

## Runtime Agent Skills (SDK API)

The Go SDK also provides a runtime `Skill` type you can attach to an `Agent` with `AddSkill`/`AddSkills`.
These runtime skills compose instructions, tools, and guardrails directly into agent behavior.

---

## Installation

```bash
go get github.com/MitulShah1/openai-agents-go@latest
```

**Requirements**: Go 1.24 or higher

---

## Quick Start

<details>
<summary><b>Hello World</b></summary>

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
        openai.UserMessage("Write a haiku about recursion in programming"),
    }

    result, err := runner.Run(context.Background(), agent, messages, nil, nil)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(result.FinalOutput)
}
```

</details>

<details>
<summary><b>Tools Example</b></summary>

```go
package main

import (
    "context"
    "fmt"

    agents "github.com/MitulShah1/openai-agents-go"
    "github.com/openai/openai-go"
)

func main() {
    client := openai.NewClient(/* ... */)
    runner := agents.NewRunner(&client)

    // Define a tool
    weatherTool := agents.FunctionTool(
        "get_weather",
        "Get the current weather for a city",
        map[string]any{
            "type": "object",
            "properties": map[string]any{
                "city": map[string]any{
                    "type": "string",
                    "description": "The city name",
                },
            },
            "required": []string{"city"},
        },
        func(args map[string]any, ctx agents.ContextVariables) (any, error) {
            city := args["city"].(string)
            return fmt.Sprintf("The weather in %s is sunny", city), nil
        },
    )

    // Create agent with tool
    agent := agents.NewAgent("Weather Agent")
    agent.Instructions = "You help users check the weather"
    agent.Tools = []agents.Tool{weatherTool}

    messages := []openai.ChatCompletionMessageParamUnion{
        openai.UserMessage("What's the weather in Tokyo?"),
    }

    result, _ := runner.Run(context.Background(), agent, messages, nil, nil)
    fmt.Println(result.FinalOutput)
}
```

</details>

<details>
<summary><b>Handoffs Example</b></summary>

```go
package main

import (
    agents "github.com/MitulShah1/openai-agents-go"
)

func main() {
    // Specialized weather agent
    weatherAgent := agents.NewAgent("Weather Specialist")
    weatherAgent.Instructions = "You are an expert at weather information"
    weatherAgent.Tools = []agents.Tool{weatherTool}

    // Main agent that hands off to specialist
    mainAgent := agents.NewAgent("Main Assistant")
    mainAgent.Instructions = "You coordinate with specialists"
    mainAgent.Tools = []agents.Tool{
        agents.HandoffTool(weatherAgent, "Transfer to weather specialist"),
    }

    // Running will automatically handle handoffs
    result, _ := runner.Run(ctx, mainAgent, messages, nil, nil)
}
```

</details>

<details>
<summary><b>Structured Outputs</b></summary>

```go
package main

import (
    "github.com/MitulShah1/openai-agents-go/jsonschema"
)

func main() {
    // Define JSON schema
    schema := jsonschema.Object().
        WithProperty("answer", jsonschema.Integer()).
        WithProperty("reasoning", jsonschema.String()).
        WithRequired("answer", "reasoning")

    // Create agent with structured output
    agent := agents.NewAgent("Math Tutor")
    agent.ResponseFormat = jsonschema.JSONSchema("math_response", schema)

    // Response will be valid JSON matching the schema
    result, _ := runner.Run(ctx, agent, messages, nil, nil)
    
    var response MathResponse
    json.Unmarshal([]byte(result.FinalOutput), &response)
}
```

</details>

---

## Running Examples

The [`examples/`](./examples) directory contains comprehensive examples:

```bash
# Set your OpenAI API key
export OPENAI_API_KEY="your-key-here"

# Basic agent
cd examples/01_basic && go run main.go

# Tools and function calling
cd examples/02_tools && go run main.go

# Agent handoffs
cd examples/03_handoffs && go run main.go

# Lifecycle hooks
cd examples/04_lifecycle_hooks && go run main.go

# Run configuration
cd examples/05_config_usage && go run main.go

# Structured outputs
cd examples/06_structured_output && go run main.go

# Complex nested schemas
cd examples/07_complex_schema && go run main.go
# Guardrails and sessions
cd examples/08_guardrails_demo && go run main.go

# Session management
cd examples/09_sessions_demo && go run main.go

# Production chatbot (combines guardrails + sessions)
cd examples/10_advanced_v02 && go run main.go

# Cloud-based Conversations session (v0.2.2)
cd examples/11_conversations_session && go run main.go

# Multimodal tools (Image/File) (v0.3.0)
cd examples/12_multimodal_tools && go run main.go

# Guardrail composition & async (v0.3.0)
cd examples/13_guardrail_composition && go run main.go

# Session backends (SQLite) (v0.3.0)
cd examples/14_session_backends && go run main.go

# Production chatbot (v0.3.0)
cd examples/16_production_chatbot && go run main.go

# Tool approvals (human-in-the-loop)
cd examples/23_tool_approvals && go run main.go
```

---

## Documentation

- 📖 **[Documentation Site](https://mitulshah1.github.io/openai-agents-go/)** - Comprehensive guides and tutorials
- 📚 [API Documentation](https://pkg.go.dev/github.com/MitulShah1/openai-agents-go) - GoDoc reference
- 📝 [CHANGELOG](./CHANGELOG.md) - Version history and release notes
- 🤖 [AI Assistant Guide](./AGENT.md) - For Claude, Copilot, etc.
- 🗺️ [Development Roadmap](./ROADMAP.md)
- 📋 [Examples Directory](./examples)

---

## Development

### Prerequisites

- Go 1.24+
- golangci-lint
- goimports

### Running Tests

```bash
# Run all tests
go test -v ./...

# Run tests with coverage
go test -v -coverprofile=coverage.out ./...

# Run linter
golangci-lint run ./...

# Or use the Makefile
make check
```

---

## Comparison with Official SDKs

| Feature | [Python SDK](https://github.com/openai/openai-agents-python) | [JavaScript SDK](https://github.com/openai/openai-agents-js) | **Go SDK (This)** |
|---------|---------------|--------------|---------------|
| Agents | ✅ | ✅ | ✅ |
| Tools | ✅ | ✅ | ✅ |
| Handoffs | ✅ | ✅ | ✅ |
| Structured Outputs | ✅ | ✅ | ✅ |
| Streaming | ✅ | ✅ | ✅ (v0.4.0) |
| Guardrails | ✅ | ✅ | ✅ (v0.2.1) |
| Tracing | ✅ | ✅ | ✅ (v0.4.0) |
| Tool Approvals | ✅ | ❌ | ✅ (v0.6.0) |
| Voice Agents | ❌ | ✅ | 🔮 Future |
| **Type Safety** | ⚠️ Runtime | ⚠️ TypeScript | ✅ Compile-time |
| **Zero Dependencies** | ❌ | ❌ | ✅ (core only) |

---

## Contributing

Contributions are welcome! This is a community project.

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

Please ensure:
- All tests pass (`make check`)
- Code is formatted (`gofmt`, `goimports`)
- Documentation is updated

---

## Acknowledgements

This project is inspired by:
- [OpenAI Agents Python SDK](https://github.com/openai/openai-agents-python)
- [OpenAI Agents JavaScript SDK](https://github.com/openai/openai-agents-js)

Built with:
- [OpenAI Go SDK](https://github.com/openai/openai-go)

---

## License

MIT License - see [LICENSE](LICENSE) file for details.

---

## Support

- 🐛 [Report Issues](https://github.com/MitulShah1/openai-agents-go/issues)
- 💬 [Discussions](https://github.com/MitulShah1/openai-agents-go/discussions)
- ⭐ Star the repo if you find it useful!

---

**Made with ❤️ by the Go community**