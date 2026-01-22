# Guardrail Composition

Since v0.3.0, you can compose complex guardrail logic using chains.

## Chaining Strategies

```go
import "github.com/MitulShah1/openai-agents-go/guardrail"

chain := guardrail.NewChain().
    // Add guardrails
    Add(security.NewPII(security.WithTripwire(true))).
    Add(moderation.NewProfanity(moderation.ProfanityConfig{Tripwire: true})).
    
    // Choose strategy
    WithStrategy(guardrail.Sequential). // Stop at first failure (Short-circuit)
    // OR
    // WithStrategy(guardrail.Parallel). // Run all concurrently
    // OR
    // WithStrategy(guardrail.StopOnFirstPass). // First pass wins
    
    Build()

// Use the chain as a single guardrail
runner.Run(..., agents.WithGuardrails([]agents.Guardrail{chain}))
```

## Async & Timeouts

Apply timeouts to guardrails (v0.3.0):

```go
// Timeout after 500ms
timedGuard := guardrail.WithTimeout(slowGuardrail, 500*time.Millisecond)

// Timeout with graceful degradation (warns instead of failing)
gracefulGuard := guardrail.WithTimeoutGraceful(slowGuardrail, 500*time.Millisecond)
```

## Metrics

Track guardrail performance (v0.3.0):

```go
metrics := guardrail.NewInMemoryMetrics()
monitoredGuard := guardrail.WithMetrics(myGuardrail, metrics)

// ... run agent ...

stats := metrics.GetStats("my_guardrail")
fmt.Printf("Avg Latency: %v", stats.AvgDuration())
```

## Legacy Chaining (Simple List)

You can also pass a simple list of guardrails to the runner. This is equivalent to a Sequential chain.

```go
guardrails := []agents.Guardrail{
    security.NewPII(security.WithTripwire(true)),
    security.NewSecrets(security.SecretsConfig{Tripwire: true}),
}
```
