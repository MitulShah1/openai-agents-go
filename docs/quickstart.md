# Quickstart Guide

This guide will help you get started with the OpenAI Agents Go SDK in minutes.

## Prerequisites

- Go 1.24 or higher
- OpenAI API key ([Get one here](https://platform.openai.com/api-keys))

## Installation

```bash
go get github.com/MitulShah1/openai-agents-go@latest
```

## Your First Agent

Let's create a simple agent that can answer questions:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    agents "github.com/MitulShah1/openai-agents-go"
    "github.com/openai/openai-go/v3"
    "github.com/openai/openai-go/v3/option"
)

func main() {
    // 1. Initialize OpenAI client
    client := openai.NewClient(option.WithAPIKey(os.Getenv("OPENAI_API_KEY")))
    runner := agents.NewRunner(&client)

    // 2. Create an agent
    agent := agents.NewAgent("Assistant")
    agent.Instructions = "You are a helpful assistant who answers questions concisely"
    agent.Model = openai.ChatModelGPT4o // Optional: defaults to gpt-4o

    // 3. Create user message
    messages := []openai.ChatCompletionMessageParamUnion{
        openai.UserMessage("What is the capital of France?"),
    }

    // 4. Run the agent
    result, err := runner.Run(context.Background(), agent, messages, nil, nil)
    if err != nil {
        log.Fatal(err)
    }

    // 5. Print the response
    fmt.Println(result.FinalOutput)
    // Output: Paris is the capital of France.
}
```

## Adding Tools

Agents can call Go functions as tools. Here's how to add a weather tool:

```go
package main

import (
    "context"
    "fmt"

    agents "github.com/MitulShah1/openai-agents-go"
    "github.com/MitulShah1/openai-agents-go/tools"
    "github.com/openai/openai-go/v3"
)

func main() {
    client := openai.NewClient(/* ... */)
    runner := agents.NewRunner(&client)

    // 1. Define a tool function
    weatherTool := tools.New(
        "get_weather",
        "Get the current weather for a city",
        map[string]any{
            "type": "object",
            "properties": map[string]any{
                "city": map[string]any{
                    "type":        "string",
                    "description": "The city name",
                },
            },
            "required": []string{"city"},
        },
        func(args map[string]any, ctx tools.ContextVariables) (any, error) {
            city := args["city"].(string)
            // In real app, call weather API here
            return fmt.Sprintf("The weather in %s is sunny, 72°F", city), nil
        },
    )

    // 2. Add tool to agent
    agent := agents.NewAgent("Weather Agent")
    agent.Instructions = "You help users check the weather"
    agent.Tools = []tools.Tool{weatherTool}

    // 3. Run agent with tool
    messages := []openai.ChatCompletionMessageParamUnion{
        openai.UserMessage("What's the weather in Tokyo?"),
    }

    result, _ := runner.Run(context.Background(), agent, messages, nil, nil)
    fmt.Println(result.FinalOutput)
    // Output: The weather in Tokyo is sunny, 72°F
}
```

## Agent Handoffs

Hand off between specialized agents:

```go
package main

import (
    agents "github.com/MitulShah1/openai-agents-go"
    "github.com/MitulShah1/openai-agents-go/handoff"
    "github.com/MitulShah1/openai-agents-go/tools"
)

func main() {
    // 1. Create specialized agent
    weatherAgent := agents.NewAgent("Weather Specialist")
    weatherAgent.Instructions = "You are an expert at weather information"
    weatherAgent.Tools = []tools.Tool{weatherTool} // from previous example

    // 2. Create main agent that can hand off
    mainAgent := agents.NewAgent("Main Assistant")
    mainAgent.Instructions = "You coordinate with specialists. Transfer weather questions to the weather specialist."
    mainAgent.Tools = []tools.Tool{
        handoff.New(weatherAgent).ToTool(),
    }

    // 3. Run - handoffs happen automatically
    messages := []openai.ChatCompletionMessageParamUnion{
        openai.UserMessage("What's the weather in Paris?"),
    }

    result, _ := runner.Run(ctx, mainAgent, messages, nil, nil)
    // The runner automatically handles the handoff to weatherAgent
}
```

## Structured Outputs

Get type-safe JSON responses with schema validation:

```go
package main

import (
    "encoding/json"
    "github.com/MitulShah1/openai-agents-go/jsonschema"
)

func main() {
    // 1. Define JSON schema
    schema := jsonschema.Object().
        WithProperty("answer", jsonschema.Integer().WithDescription("The calculated answer")).
        WithProperty("reasoning", jsonschema.String().WithDescription("Step-by-step reasoning")).
        WithRequired("answer", "reasoning")

    // 2. Create agent with structured output
    agent := agents.NewAgent("Math Tutor")
    agent.Instructions = "Solve math problems and show your reasoning"
    agent.ResponseFormat = jsonschema.JSONSchema("math_response", schema)

    // 3. Run agent
    messages := []openai.ChatCompletionMessageParamUnion{
        openai.UserMessage("What is 15 + 27?"),
    }

    result, _ := runner.Run(ctx, agent, messages, nil, nil)

    // 4. Parse structured output
    type MathResponse struct {
        Answer    int    `json:"answer"`
        Reasoning string `json:"reasoning"`
    }

    var response MathResponse
    json.Unmarshal([]byte(result.FinalOutput), &response)
    
    fmt.Printf("Answer: %d\nReasoning: %s\n", response.Answer, response.Reasoning)
}
```

## Adding Guardrails

Protect your agents with input/output validation:

```go
package main

import (
    "github.com/MitulShah1/openai-agents-go/guardrail/security"
)

func main() {
    // 1. Create guardrails
    piiGuardrail := security.NewPII(security.WithTripwire(true))
    urlGuardrail := security.NewURLFilter(
        security.WithURLTripwire(true),
    )

    // 2. Run agent with guardrails
    result, err := runner.Run(
        ctx,
        agent,
        messages,
        nil,
        agents.WithGuardrails([]agents.Guardrail{piiGuardrail, urlGuardrail}),
    )

    // Guardrails automatically validate input before sending to API
    // and validate output before returning to user
}
```

## Using Sessions

Persist conversation history across runs:

```go
package main

import (
    "github.com/MitulShah1/openai-agents-go/session"
)

func main() {
    // 1. Create session (in-memory for testing, file-based for production)
    sess := session.NewMemorySession()

    // 2. Run agent with session
    result1, _ := runner.Run(
        ctx,
        agent,
        []openai.ChatCompletionMessageParamUnion{
            openai.UserMessage("My name is Alice"),
        },
        nil,
        agents.WithSession(sess, "user-123"),
    )

    // 3. Next run remembers context
    result2, _ := runner.Run(
        ctx,
        agent,
        []openai.ChatCompletionMessageParamUnion{
            openai.UserMessage("What's my name?"),
        },
        nil,
        agents.WithSession(sess, "user-123"),
    )
    
    fmt.Println(result2.FinalOutput)
    // Output: Your name is Alice.
}
```

## Configuration Options

Control agent execution with `RunConfig`:

```go
package main

import (
    "time"
    agents "github.com/MitulShah1/openai-agents-go"
)

func main() {
    // Set maximum turns and timeout
    result, err := runner.Run(
        ctx,
        agent,
        messages,
        &agents.RunConfig{
            MaxTurns: 5,                    // Limit agent loops
            Timeout:  30 * time.Second,     // Overall timeout
            Debug:    true,                  // Enable debug logging
        },
        nil,
    )
}
```

## Next Steps

Now that you've learned the basics, explore more advanced topics:

- **[Agents](agents.md)** - Deep dive into agent configuration
- **[Tools](tools.md)** - Advanced tool patterns and custom tools
- **[Guardrails](guardrails/index.md)** - Comprehensive input/output validation
- **[Sessions](sessions/index.md)** - Persistent conversation management
- **[Structured Outputs](structured_outputs.md)** - Complex JSON schemas

Or check out the [Examples](https://github.com/MitulShah1/openai-agents-go/tree/main/examples) on GitHub for complete working code.
