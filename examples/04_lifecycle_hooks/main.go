// Package main demonstrates lifecycle hooks - OnBeforeRun and OnAfterRun for logging and validation.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"

	agents "github.com/MitulShah1/openai-agents-go"
)

// This example demonstrates lifecycle hooks - functions that run before and after agent execution.
// Lifecycle hooks are useful for logging, metrics, validation, and other cross-cutting concerns.
func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Please set OPENAI_API_KEY environment variable.")
		return
	}

	client := openai.NewClient(option.WithAPIKey(apiKey))
	runner := agents.NewRunner(&client)

	// Create agent
	agent := agents.NewAgent("HookedAgent")
	agent.Instructions = "You are a helpful assistant. Answer questions concisely."

	// Track execution time
	var startTime time.Time

	// OnBeforeRun - executes before the agent starts
	agent.OnBeforeRun = func(_ context.Context, a *agents.Agent) error {
		startTime = time.Now()
		fmt.Printf("🚀 Starting agent: %s\n", a.Name)
		fmt.Printf("📝 Instructions: %s\n", a.Instructions)
		fmt.Printf("🤖 Model: %s\n\n", a.Model)
		return nil
	}

	// OnAfterRun - executes after the agent completes
	agent.OnAfterRun = func(_ context.Context, a *agents.Agent) error {
		elapsed := time.Since(startTime)
		fmt.Printf("\n✅ Agent completed: %s\n", a.Name)
		fmt.Printf("⏱️  Total execution time: %v\n", elapsed)
		return nil
	}

	// Simple tool for demonstration
	calculator := agents.FunctionTool(
		"calculate",
		"Perform a simple calculation",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"expression": map[string]any{
					"type":        "string",
					"description": "The math expression to evaluate, e.g. '2 + 2'",
				},
			},
			"required": []any{"expression"},
		},
		func(args map[string]any, _ agents.ContextVariables) (any, error) {
			expression := args["expression"].(string)
			// Simplified - in real app, use a proper math parser
			result := "4" // Hardcoded for demo
			return fmt.Sprintf("%s = %s", expression, result), nil
		},
	)

	agent.Tools = []agents.Tool{calculator}

	// Run the agent
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("What is 2 + 2?"),
	}

	ctx := context.Background()
	result, err := runner.Run(ctx, agent, messages, nil, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Print results
	fmt.Printf("\n=== Results ===\n")
	fmt.Printf("Response: %s\n", result.FinalOutput)
	fmt.Printf("Steps: %d\n", len(result.Steps))
	fmt.Printf("Tokens: %d\n", result.Usage.TotalTokens)

	// Demonstrate error handling in hooks
	fmt.Printf("\n=== Testing Error Handling in Hooks ===\n")

	errorAgent := agents.NewAgent("ErrorAgent")
	errorAgent.Instructions = "Test agent"

	// This hook returns an error - execution will fail before running
	errorAgent.OnBeforeRun = func(_ context.Context, _ *agents.Agent) error {
		return fmt.Errorf("validation failed: agent not ready")
	}

	_, err = runner.Run(ctx, errorAgent, messages, nil, nil)
	if err != nil {
		fmt.Printf("❌ Expected error: %v\n", err)
	}
}
