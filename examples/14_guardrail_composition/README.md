# Guardrail Composition Example

This example demonstrates the **guardrail composition features** introduced in v0.3.0.

## Features Demonstrated

### 1. Chain Execution Strategies
- **Sequential**: Run guardrails in order, fail-fast on first failure
- **Parallel**: Run all guardrails concurrently, aggregate results
- **StopOnFirstPass**: OR logic, stop on first pass (useful for bypass scenarios)

### 2. Async Capabilities
- **WithTimeout()**: Add execution timeout to guardrails
- **WithTimeoutGraceful()**: Graceful degradation when timeout occurs
- **WithContext()**: Context-aware cancellation support

### 3. Metrics Collection
- **MetricsCollector** interface for telemetry
- **InMemoryMetrics** with latency tracking
- **WithMetrics()** wrapper for instrumentation
- P95/P99 percentile calculations

## Examples Covered

### Example 1: Sequential Chain
Demonstrates fail-fast behavior where guardrails run in order and stop at the first failure.

### Example 2: Parallel Execution
Shows concurrent execution of multiple guardrails with aggregated results.

### Example 3: StopOnFirstPass (OR Logic)
Implements "bypass" logic where execution stops as soon as one guardrail passes.

### Example 4: Timeout and Metrics
Demonstrates timeout handling and metrics collection for slow guardrails.

### Example 5: Graceful Degradation
Shows how to gracefully handle timeouts without failing the entire validation.

### Example 6: Production-Ready Chain
Complete example with metrics tracking for a production validation pipeline.

## Running the Example

```bash
export OPENAI_API_KEY="your-api-key"
go run main.go
```

## Key APIs

```go
// Build a chain with fluent API
chain := guardrail.NewChain().
    Add(guard1).
    Add(guard2).
    WithStrategy(guardrail.Sequential).
    WithName("my_chain").
    Build()

// Add timeout
timedGuard := guardrail.WithTimeout(guard, 500*time.Millisecond)

// Add metrics
metrics := guardrail.NewInMemoryMetrics()
metricedGuard := guardrail.WithMetrics(guard, metrics)

// Get metrics
stats := metrics.GetStats("guard_name")
fmt.Printf("P95: %v, P99: %v\n", stats.P95Duration(), stats.P99Duration())
```

## Learn More

- See `guardrail/chain.go` for chain implementation
- See `guardrail/async.go` for timeout support
- See `guardrail/metrics.go` for metrics collection
- Check tests for comprehensive examples
