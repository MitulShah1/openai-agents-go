// Package main demonstrates guardrail composition features introduced in v0.3.0.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"

	agents "github.com/MitulShah1/openai-agents-go"
	"github.com/MitulShah1/openai-agents-go/guardrail"
	"github.com/MitulShah1/openai-agents-go/guardrail/content"
	"github.com/MitulShah1/openai-agents-go/guardrail/moderation"
	"github.com/MitulShah1/openai-agents-go/guardrail/security"
)

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	client := openai.NewClient(option.WithAPIKey(apiKey))
	runner := agents.NewRunner(&client)

	agent := agents.NewAgent("helpful_assistant")
	agent.Instructions = "You are a helpful assistant."
	agent.Model = "gpt-4o-mini"

	// Example 1: Sequential Chain (fail-fast)
	fmt.Println("=== Example 1: Sequential Chain ===")
	fmt.Println("Guardrails run in order and stop at the first failure")

	sequentialChain := guardrail.NewChain().
		Add(content.NewLength(content.Config{Min: 10, Max: 500})).
		Add(moderation.NewProfanity(moderation.ProfanityConfig{Tripwire: true})).
		WithStrategy(guardrail.Sequential).
		WithName("sequential_validation").
		Build()

	messages1 := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Hello, how are you?"),
	}

	// Attach guardrail to agent
	agent.InputGuardrails = []*guardrail.Guardrail{sequentialChain}

	result, err := runner.Run(context.Background(), agent, messages1)
	if err != nil {
		log.Printf("Expected behavior - guardrail may fail: %v\n", err)
	} else {
		fmt.Printf("✓ Response: %s\n\n", result.FinalOutput)
	}

	// Example 2: Parallel Execution
	fmt.Println("=== Example 2: Parallel Execution ===")
	fmt.Println("All guardrails run concurrently, results are aggregated")

	parallelChain := guardrail.NewChain().
		Add(content.NewLength(content.Config{Min: 5, Max: 1000})).
		Add(security.NewURLFilter()).
		Add(security.NewSecrets(security.SecretsConfig{Tripwire: true})).
		WithStrategy(guardrail.Parallel).
		WithName("parallel_validation").
		Build()

	messages2 := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("What's the weather like?"),
	}

	agent.InputGuardrails = []*guardrail.Guardrail{parallelChain}

	result, err = runner.Run(context.Background(), agent, messages2)
	if err != nil {
		log.Printf("Some guardrails failed: %v\n", err)
	} else {
		fmt.Printf("✓ All guardrails passed\n")
		fmt.Printf("✓ Response: %s\n\n", result.FinalOutput)
	}

	// Example 3: StopOnFirstPass (OR logic)
	fmt.Println("=== Example 3: StopOnFirstPass Strategy ===")
	fmt.Println("Stops as soon as one guardrail passes (useful for bypass logic)")

	// Create multiple guardrails where at least one should pass
	bypassChain := guardrail.NewChain().
		Add(content.NewLength(content.Config{Min: 1, Max: 10})).  // This will likely fail
		Add(content.NewLength(content.Config{Min: 1, Max: 500})). // This should pass
		WithStrategy(guardrail.StopOnFirstPass).
		WithName("bypass_validation").
		Build()

	messages3 := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Tell me a joke"),
	}

	agent.InputGuardrails = []*guardrail.Guardrail{bypassChain}

	result, err = runner.Run(context.Background(), agent, messages3)
	if err != nil {
		log.Printf("No guardrail passed: %v\n", err)
	} else {
		fmt.Printf("✓ At least one guardrail passed\n")
		fmt.Printf("✓ Response: %s\n\n", result.FinalOutput)
	}

	// Example 4: Timeout Handling
	fmt.Println("=== Example 4: Timeout Handling ===")
	fmt.Println("Guardrails with timeout protection")

	// Wrap a guardrail with timeout
	timeoutGuard := guardrail.WithTimeout(
		content.NewLength(content.Config{Min: 5, Max: 500}),
		100*time.Millisecond,
	)

	messages4 := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("What's 2+2?"),
	}

	agent.InputGuardrails = []*guardrail.Guardrail{timeoutGuard}

	result, err = runner.Run(context.Background(), agent, messages4)
	if err != nil {
		log.Printf("Guardrail error (possibly timeout): %v\n", err)
	} else {
		fmt.Printf("✓ Guardrail completed within timeout\n")
		fmt.Printf("✓ Response: %s\n\n", result.FinalOutput)
	}

	// Example 5: Graceful Degradation
	fmt.Println("=== Example 5: Graceful Timeout Degradation ===")
	fmt.Println("Allows execution to continue with warning on timeout")

	gracefulGuard := guardrail.WithTimeoutGraceful(
		content.NewLength(content.Config{Min: 5, Max: 500}),
		50*time.Millisecond,
	)

	messages5 := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Hello!"),
	}

	agent.InputGuardrails = []*guardrail.Guardrail{gracefulGuard}

	result, err = runner.Run(context.Background(), agent, messages5)
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("✓ Request proceeded despite potential timeout\n")
		fmt.Printf("✓ Response: %s\n\n", result.FinalOutput)
	}

	// Example 6: Metrics Collection
	fmt.Println("=== Example 6: Metrics Collection ===")
	fmt.Println("Track guardrail performance and statistics")

	// Create metrics collector
	metrics := guardrail.NewInMemoryMetrics()

	// Wrap guardrails with metrics
	metricsGuard := guardrail.WithMetrics(
		content.NewLength(content.Config{Min: 5, Max: 500}),
		metrics,
	)

	// Run multiple times to collect data
	agent.InputGuardrails = []*guardrail.Guardrail{metricsGuard}

	for i := 0; i < 5; i++ {
		messagesLoop := []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(fmt.Sprintf("Test message %d", i+1)),
		}

		_, _ = runner.Run(context.Background(), agent, messagesLoop)
	}

	// Print metrics
	stats := metrics.GetStats("content_length")
	if stats != nil {
		fmt.Printf("✓ Guardrail Metrics:\n")
		fmt.Printf("  Total executions: %d\n", stats.TotalCount)
		fmt.Printf("  Passed: %d\n", stats.PassedCount)
		fmt.Printf("  Failed: %d\n", stats.FailedCount)
		fmt.Printf("  Avg duration: %v\n", stats.AvgDuration())
		fmt.Printf("  P95 duration: %v\n", stats.P95Duration())
		fmt.Printf("  P99 duration: %v\n", stats.P99Duration())
	}
	fmt.Println()

	// Example 7: Production Chain with Metrics
	fmt.Println("=== Example 7: Production-Ready Chain ===")
	fmt.Println("Complete validation pipeline with metrics tracking")

	productionChain := guardrail.NewChain().
		Add(guardrail.WithMetrics(content.NewLength(content.Config{Min: 10, Max: 1000}), metrics)).
		Add(guardrail.WithMetrics(moderation.NewProfanity(moderation.ProfanityConfig{}), metrics)).
		Add(guardrail.WithMetrics(security.NewSecrets(security.SecretsConfig{}), metrics)).
		Add(guardrail.WithMetrics(security.NewURLFilter(), metrics)).
		WithStrategy(guardrail.Sequential).
		WithName("production_validation").
		Build()

	messages7 := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Can you help me with a task?"),
	}

	agent.InputGuardrails = []*guardrail.Guardrail{productionChain}

	result, err = runner.Run(context.Background(), agent, messages7)
	if err != nil {
		log.Printf("Validation failed: %v\n", err)
	} else {
		fmt.Printf("✓ All validations passed\n")
		fmt.Printf("✓ Response: %s\n\n", result.FinalOutput)
	}

	// Print comprehensive metrics summary
	fmt.Println("\n=== Metrics Summary ===")
	allStats := metrics.GetAllStats()
	for name, stats := range allStats {
		if stats.TotalCount > 0 {
			fmt.Printf("\n%s:\n", name)
			fmt.Printf("  Executions: %d (Passed: %d, Failed: %d)\n",
				stats.TotalCount, stats.PassedCount, stats.FailedCount)
			fmt.Printf("  Avg: %v, P95: %v, P99: %v\n",
				stats.AvgDuration(), stats.P95Duration(), stats.P99Duration())
		}
	}

	fmt.Println("\n✅ Guardrail Composition Demo Complete!")
	fmt.Println("\nKey Features Demonstrated:")
	fmt.Println("  • Sequential chain execution (fail-fast)")
	fmt.Println("  • Parallel guardrail execution")
	fmt.Println("  • StopOnFirstPass strategy (OR logic)")
	fmt.Println("  • Timeout protection with graceful degradation")
	fmt.Println("  • Metrics collection and analysis")
	fmt.Println("  • Production-ready validation pipeline")
}
