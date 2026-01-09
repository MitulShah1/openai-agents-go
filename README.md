# OpenAI Agents Go SDK (Unofficial)

[![CI](https://github.com/MitulShah1/openai-agents-go/workflows/CI/badge.svg)](https://github.com/MitulShah1/openai-agents-go/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/MitulShah1/openai-agents-go)](https://goreportcard.com/report/github.com/MitulShah1/openai-agents-go)
[![GoDoc](https://pkg.go.dev/badge/github.com/MitulShah1/openai-agents-go.svg)](https://pkg.go.dev/github.com/MitulShah1/openai-agents-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

> **Note**: This is an **unofficial** community-maintained Go SDK for building AI agents with OpenAI's API. It is inspired by the official [OpenAI Agents Python SDK](https://github.com/openai/openai-agents-python) and [OpenAI Agents JavaScript SDK](https://github.com/openai/openai-agents-js), but is not affiliated with or endorsed by OpenAI.

A lightweight, powerful Go framework for building multi-agent workflows with OpenAI's API. Build production-ready AI agents with tool calling, handoffs, structured outputs, and more.

---

## Core Concepts

1. **Agents**: LLMs configured with instructions, tools, and behavior settings
2. **Tools**: Functions that agents can call to perform actions
3. **Handoffs**: Transfer control between specialized agents dynamically
4. **Structured Outputs**: Schema-validated JSON responses for reliable parsing
5. **Run Configuration**: Control execution flow with max turns, timeouts, and more

Explore the [`examples/`](./examples) directory to see the SDK in action.

---

## Supported Features

- ✅ **Multi-Agent Workflows**: Compose and orchestrate multiple agents
- ✅ **Tool Integration**: Seamlessly call Go functions from agent responses
- ✅ **Handoffs**: Dynamic agent-to-agent transfers during execution
- ✅ **Structured Outputs**: Schema-validated JSON responses with fluent API
- ✅ **Lifecycle Hooks**: Execute code before/after agent runs
- ✅ **Context Variables**: Pass state between agents and tools
- ✅ **Usage Tracking**: Monitor token consumption and costs
- ✅ **Error Handling**: Comprehensive error types for debugging
- ✅ **Type Safety**: Full Go type safety with generics support
- 🔮 **Streaming** (Coming soon - see [ROADMAP.md](./ROADMAP.md))
- 🔮 **Tracing & Debugging** (Planned)
- 🔮 **Guardrails** (Planned)

---

## Installation

```bash
go get github.com/MitulShah1/openai-agents-go@latest
```

**Requirements**: Go 1.24 or higher

---

## Quick Start

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
        openai.UserMessage("Write a haiku about recursion in programming"),
    }

    result, err := runner.Run(context.Background(), agent, messages, nil, nil)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(result.FinalOutput)
}
```

### Tools Example

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

### Handoffs Example

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

### Structured Outputs

```go
package main

import (
    "github.com/MitulShah1/openai-agents-go/internal/jsonschema"
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
```

---

## Documentation

- 📚 [API Documentation](https://pkg.go.dev/github.com/MitulShah1/openai-agents-go)
- 🗺️ [Development Roadmap](./ROADMAP.md)
- 📝 [Examples Directory](./examples)
- 🔧 [OpenAI Go SDK](https://github.com/openai/openai-go)

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
| Streaming | ✅ | ✅ | 🔮 Planned |
| Guardrails | ✅ | ✅ | 🔮 Planned |
| Tracing | ✅ | ✅ | 🔮 Planned |
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