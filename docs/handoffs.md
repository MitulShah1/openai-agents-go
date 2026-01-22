# Handoffs

Handoffs allow an agent to transfer control of a conversation to another agent. This is fundamental for building multi-agent systems where specialized agents handle different parts of a complex workflow.

The Go SDK provides a type-safe, declarative API for configuring handoffs that matches the capabilities of the Python SDK.

## Basic Usage

The simplest way to create a handoff is to use `handoff.New()` with a target agent. This automatically generates a tool that, when called by the model, transfers control to that agent.

```go
import (
    "github.com/MitulShah1/openai-agents-go/handoff"
    "github.com/MitulShah1/openai-agents-go"
)

// Create agents
triageAgent := agents.NewAgent("Triage")
supportAgent := agents.NewAgent("Support")

// Create a handoff tool
// By default, this creates a tool named "transfer_to_support"
transferToSupport := handoff.New(supportAgent).ToTool()

// Register the tool with the source agent
triageAgent.Tools = []tools.Tool{transferToSupport}
```

## Configuration Options

The `handoff.New()` function accepts functional options to customize behavior.

### Custom Tool Name and Description

You can override the automatically generated tool name and description to better guide the model on when to use the handoff.

```go
transfer := handoff.New(
    billingAgent,
    handoff.WithToolName("escalate_payment_issue"),
    handoff.WithDescription("Use this when the user has a problem with a credit card charge."),
).ToTool()
```

### Input Filtering

Input filters allow you to inspect or modify the conversation state before the handoff occurs. This is useful for redacting sensitive data, adding context variables, or validation.

```go
filter := func(ctx context.Context, data handoff.InputData) (handoff.InputData, error) {
    // Add a context variable for the next agent
    data.ContextVars["referring_agent"] = "triage"
    return data, nil
}

transfer := handoff.New(
    salesAgent,
    handoff.WithInputFilter(filter),
).ToTool()
```

### History Nesting (Summarization)

By default, the entire conversation history is passed to the next agent. For long conversations, this can consume many tokens. History nesting summarizes the conversation history into a single message for the next agent.

```go
transfer := handoff.New(
    archiveAgent,
    handoff.WithHistoryNesting(true), // Content will be summarized
).ToTool()
```

### Dynamic Enablement

You can control when a handoff is available using a predicate function. This is evaluated at runtime.

```go
// Only allow handoff during business hours
predicate := func(ctx context.Context, agent *agents.Agent, vars agents.ContextVariables) (bool, error) {
    hour := time.Now().Hour()
    return hour >= 9 && hour < 17, nil
}

transfer := handoff.New(
    liveAgent,
    handoff.WithEnabledPredicate(predicate),
).ToTool()
```

## Advanced Patterns

### Context Variables

Handoffs preserve `ContextVariables`. Any variables set by previous agents are available to the target agent.

```go
// In Agent A
ctxVars["customer_id"] = "123"
// ... handoff to Agent B ...
// In Agent B
id := ctxVars["customer_id"] // "123"
```

### Multi-Hop Handoffs

Agents can hand off to other agents that can, in turn, hand off again. The SDK manages the call stack and history automatically.

## API Reference

See the [GoDoc](https://pkg.go.dev/github.com/MitulShah1/openai-agents-go/handoff) for the full API reference.
