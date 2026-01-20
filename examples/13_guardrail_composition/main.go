package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	agents "github.com/MitulShah1/openai-agents-go"
	"github.com/MitulShah1/openai-agents-go/guardrail"
	"github.com/MitulShah1/openai-agents-go/guardrail/builtin"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	client := openai.NewClient(option.WithAPIKey(apiKey))
	runner := agents.NewRunner(&client)

	// Create agent
	agent := agents.NewAgent("content_moderator")
	agent.Instructions = "You are a helpful assistant that generates content based on user requests."
	agent.Model = "gpt-4o-mini"

	// Example 1: Sequential chain (fail-fast)
	fmt.Println("=== Example 1: Sequential Chain (Fail-Fast) ===")
	sequentialChain := guardrail.NewChain().
		Add(builtin.NewPIIGuardrail(builtin.WithTripwire(false))).
		Add(builtin.NewURLFilterGuardrail(
			builtin.WithBlocklist("badsite.com", "spam.com"),
			builtin.WithURLTripwire(false),
		)).
		Add(builtin.NewSecretsGuardrail(builtin.WithSecretsTripwire(false))).
		WithStrategy(guardrail.Sequential).
		WithName("sequential_moderation").
		Build()

	runWithGuardrails(runner, agent, "Tell me a short story about Paris", []*guardrail.Guardrail{sequentialChain})

	// Example 2: Parallel execution (run all)
	fmt.Println("\n=== Example 2: Parallel Execution ===")
	parallelChain := guardrail.NewChain().
		Add(builtin.NewPIIGuardrail()).
		Add(builtin.NewSecretsGuardrail()).
		Add(builtin.NewProfanityGuardrail()).
		WithStrategy(guardrail.Parallel).
		WithName("parallel_safety_checks").
		Build()

	runWithGuardrails(runner, agent, "Generate a motivational quote", []*guardrail.Guardrail{parallelChain})

	// Example 3: StopOnFirstPass (OR logic)
	fmt.Println("\n=== Example 3: StopOnFirstPass (OR Logic) ===")
	g1 := guardrail.NewGuardrail("short_enough", func(ctx context.Context, input string) (*guardrail.Result, error) {
		if len(input) < 50 {
			return &guardrail.Result{Passed: true, Message: "Input is short"}, nil
		}
		return &guardrail.Result{Passed: false, Message: "Input too long"}, nil
	})

	g2 := guardrail.NewGuardrail("contains_keyword", func(ctx context.Context, input string) (*guardrail.Result, error) {
		keywords := []string{"urgent", "important", "critical"}
		for _, kw := range keywords {
			if strings.Contains(strings.ToLower(input), kw) {
				return &guardrail.Result{Passed: true, Message: fmt.Sprintf("Contains keyword: %s", kw)}, nil
			}
		}
		return &guardrail.Result{Passed: false, Message: "No keywords found"}, nil
	})

	orChain := guardrail.NewChain().
		Add(g1).
		Add(g2).
		WithStrategy(guardrail.StopOnFirstPass).
		WithName("short_or_important").
		Build()

	runWithGuardrails(runner, agent, "This is urgent: please help!", []*guardrail.Guardrail{orChain})

	// Example 4: Timeout and metrics
	fmt.Println("\n=== Example 4: Timeout and Metrics ===")
	metrics := guardrail.NewInMemoryMetrics()

	slowGuardrail := guardrail.NewGuardrail("slow_check", func(ctx context.Context, input string) (*guardrail.Result, error) {
		time.Sleep(2 * time.Second)
		return &guardrail.Result{Passed: true, Message: "Slow check passed"}, nil
	})

	// Wrap with timeout and metrics
	timedGuardrail := guardrail.WithTimeout(slowGuardrail, 500*time.Millisecond)
	metricedGuardrail := guardrail.WithMetrics(timedGuardrail, metrics)

	fmt.Println("Running guardrail with 500ms timeout (will timeout)...")
	result, err := metricedGuardrail.Func(context.Background(), "test input")
	if err != nil {
		fmt.Printf("  Error (expected): %v\n", err)
	} else {
		fmt.Printf("  Result: %s\n", result.Message)
	}

	// Check metrics
	stats := metrics.GetStats("slow_check_with_timeout")
	if stats != nil {
		fmt.Printf("  Metrics - Total: %d, Errors: %d\n", stats.TotalCount, stats.ErrorCount)
	}

	// Example 5: Graceful degradation
	fmt.Println("\n=== Example 5: Graceful Degradation ===")
	gracefulGuardrail := guardrail.WithTimeoutGraceful(slowGuardrail, 500*time.Millisecond)

	fmt.Println("Running guardrail with graceful timeout (continues gracefully)...")
	result, err = gracefulGuardrail.Func(context.Background(), "test input")
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
	} else {
		fmt.Printf("  Result: %s (degraded: %v)\n", result.Message, result.Metadata["degraded"])
	}

	// Example 6: Full composite chain with metrics
	fmt.Println("\n=== Example 6: Production-Ready Chain ===")
	productionMetrics := guardrail.NewInMemoryMetrics()

	// Build comprehensive validation chain
	piiCheck := guardrail.WithMetrics(builtin.NewPIIGuardrail(), productionMetrics)
	secretsCheck := guardrail.WithMetrics(builtin.NewSecretsGuardrail(), productionMetrics)
	profanityCheck := guardrail.WithMetrics(builtin.NewProfanityGuardrail(), productionMetrics)

	productionChain := guardrail.NewChain().
		Add(piiCheck).
		Add(secretsCheck).
		Add(profanityCheck).
		WithStrategy(guardrail.Sequential).
		WithName("production_validation").
		Build()

	runWithGuardrails(runner, agent, "Write a professional email about project updates", []*guardrail.Guardrail{productionChain})

	// Print metrics summary
	fmt.Println("\n=== Metrics Summary ===")
	for name, stats := range productionMetrics.GetAllStats() {
		fmt.Printf("%s:\n", name)
		fmt.Printf("  Total: %d, Passed: %d, Failed: %d\n", stats.TotalCount, stats.PassedCount, stats.FailedCount)
		fmt.Printf("  Avg Duration: %v\n", stats.AvgDuration())
		if len(stats.Durations) > 0 {
			fmt.Printf("  P95 Duration: %v, P99 Duration: %v\n", stats.P95Duration(), stats.P99Duration())
		}
	}
}

func runWithGuardrails(runner *agents.Runner, agent *agents.Agent, input string, guards []*guardrail.Guardrail) {
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage(input),
	}

	result, err := runner.Run(
		context.Background(),
		agent,
		messages,
		agents.WithGuardrails(guards...),
	)

	if err != nil {
		fmt.Printf("  Input: %s\n", input)
		fmt.Printf("  Error: %v\n", err)
		return
	}

	fmt.Printf("  Input: %s\n", input)
	fmt.Printf("  Output: %s\n", result.FinalOutput)
	fmt.Printf("  Tokens: %d\n", result.Usage.TotalTokens)
}
