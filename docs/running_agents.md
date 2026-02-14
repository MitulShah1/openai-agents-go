# Running Agents

The `Runner` is the engine that drives agent execution. It manages the conversation loop, interacts with the OpenAI API, executes tools, and handles agent handoffs.

## The Runner

First, initialize a `Runner` with an OpenAI client:

```go
import (
    "github.com/openai/openai-go"
    "github.com/openai/openai-go/option"
    agents "github.com/MitulShah1/openai-agents-go"
)

client := openai.NewClient(option.WithAPIKey("your-api-key"))
runner := agents.NewRunner(&client)
```

### Custom Model Providers

For explicit control over model providers, use `NewRunnerWithProvider`:

```go
import (
    agents "github.com/MitulShah1/openai-agents-go"
    "github.com/MitulShah1/openai-agents-go/models"
)

// Create an explicit OpenAI provider
provider := models.NewOpenAIProvider(&client)
runner := agents.NewRunnerWithProvider(provider)
```

This is useful when you want to:
- Use multiple LLM providers in one application
- Implement custom providers for Anthropic, Google, or other backends
- Control model resolution behavior

See [Models](models.md) for details on the provider abstraction.

## Basic Execution

To run an agent, call the `Run` method with a context, the agent, and the initial messages.

```go
messages := []openai.ChatCompletionMessageParamUnion{
    openai.UserMessage("Hello! Can you help me?"),
}

result, err := runner.Run(context.Background(), myAgent, messages)
if err != nil {
    log.Fatal(err)
}

fmt.Println(result.FinalOutput)
```


## Asynchronous Execution

For non-blocking operations, use `RunAsync`. This runs the agent loop in a separate goroutine and returns a channel that receives the result.

```go
// Start execution immediately
resultChan := runner.RunAsync(ctx, agent, messages)

// Perform other work while the agent thinks and acts...
fmt.Println("Agent is running...")

// Wait for the result
asyncResult := <-resultChan
if asyncResult.Error != nil {
    log.Fatal(asyncResult.Error)
}

fmt.Println("Result:", asyncResult.Result.FinalOutput)
```

## The Execution Loop

When `Run()` is called, the runner executes a continuous loop:

1.  **Context Preparation**: It resolves the current agent's instructions (calling the function if dynamic).
2.  **LLM Request**: It sends the conversation history and available tools to the OpenAI API.
3.  **Model Decision**: The model returns either a text response or tool calls.
4.  **Tool Execution**:
    *   If **Text**: The loop typically ends (unless configured otherwise).
    *   If **Tool Calls**: The runner executes the Go functions for the requested tools.
5.  **Handoff Check**: If a tool returns a new `*Agent`, the runner updates the `Current Agent` for the next turn.
6.  **Loop**: The process repeats with the updated history (including tool results) and the potentially new agent.

The loop terminates when:
*   The model produces a final text response and no further tools are called.
*   The maximum number of turns is reached.
*   A timeout occurs.
*   A tool/execution error occurs (depending on handling).

## Configuration

You can customize the execution behavior using functional options passed to `Run()`.

### Run Configuration (`RunConfig`)

Control operational limits and debug settings:

```go
config := &agents.RunConfig{
    MaxTurns:          30,              // Max conversation turns (default: 30)
    Timeout:           2 * time.Minute, // Total execution timeout
    Temperature:       nil,             // Override agent temperature
    MaxTokens:         nil,             // Override agent max tokens
    ParallelToolCalls: true,            // Enable/disable parallel execution
    Debug:             true,            // Print debug logs to stdout
}

runner.Run(ctx, agent, messages, agents.WithConfig(config))
```

### Context Variables

Pass data into tools without polluting the global scope or agent definition. These are available to every tool execution in the run.

```go
ctxVars := agents.ContextVariables{
    "user_id": "user_123",
    "region":  "us-east-1",
}

runner.Run(ctx, agent, messages, agents.WithContextVariables(ctxVars))
```

### Sessions

Use sessions to automatically load and save conversation history from a storage backend.

```go
// Example: Using an in-memory session backend
sess := session.NewInMemorySession()

runner.Run(ctx, agent, messages, agents.WithSession(sess, "unique-session-id"))
```

The runner will load previous messages for "unique-session-id" before starting and save the new interactions after finishing.

## Inspecting Results

The `Run()` method returns a `Result` struct containing comprehensive details about the execution.

```go
type Result struct {
    // The final text response from the agent
    FinalOutput string
    
    // The full conversation history, including initial messages and new interactions
    Messages []openai.ChatCompletionMessageParamUnion
    
    // The specific agent that was active at the end of the run
    Agent *Agent
    
    // A simplified trace of steps (Agent Name, Duration, Tools Called)
    Steps []Step
    
    // Token usage statistics for the entire run
    Usage Usage
}
```

You can use the `Steps` slice to visualize the agent's thought process or debug handoffs.

```go
for i, step := range result.Steps {
    fmt.Printf("Step %d [%s]: %v\n", i, step.AgentName, step.ToolCalls)
}
```

## Concurrency Model

The runner is designed for safe concurrent execution:

*   **Thread Safety**: The `Runner` instance is safe for concurrent use. You can share a single `Runner` across multiple goroutines or HTTP requests.
*   **Parallel Tools**: By default, multiple tool calls in a single turn are executed in parallel.
    *   This uses `sync.WaitGroup` to ensure all tools complete.
    *   **Robustness**: If one tool fails, others continue to execute. The runner aggregates all results.
    *   You can disable this by setting `ParallelToolCalls: false` in `RunConfig` or the `Agent` configuration.

## Tool Approvals

The runner supports human-in-the-loop approval for sensitive tool operations. Tools can be marked with `NeedsApproval` or use `ApprovalFunc` for conditional approval.

### Inline Handler

For synchronous approval, provide a handler via `WithApprovalHandler`:

```go
result, err := runner.Run(ctx, agent, messages,
    agents.WithApprovalHandler(func(req tools.ApprovalRequest) (*tools.ApprovalResponse, error) {
        fmt.Printf("Approve %s? ", req.ToolName)
        return &tools.ApprovalResponse{Approved: true}, nil
    }),
)
```

### Pause/Resume

When no handler is set, `Run()` returns a `ToolApprovalRequiredError`:

```go
result, err := runner.Run(ctx, agent, messages)

var approvalErr *agents.ToolApprovalRequiredError
if errors.As(err, &approvalErr) {
    approvals := map[string]*tools.ApprovalResponse{
        approvalErr.Requests[0].CallID: {Approved: true},
    }
    result, err = runner.Resume(ctx, approvalErr.State, approvals)
}
```

See [Tools - Tool Approvals](tools.md#tool-approvals) for complete documentation.
