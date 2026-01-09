// Package main demonstrates RunConfig usage and token tracking for cost estimation.
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

// This example demonstrates RunConfig usage and token usage tracking.
// RunConfig allows you to control agent execution with max turns, temperature, timeouts, etc.
func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Please set OPENAI_API_KEY environment variable.")
		return
	}

	client := openai.NewClient(option.WithAPIKey(apiKey))
	runner := agents.NewRunner(&client)

	// Create agent with a tool that requires multiple turns
	recursiveTool := agents.FunctionTool(
		"get_info",
		"Get information iteratively (demonstrates max turns)",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"query": map[string]any{
					"type":        "string",
					"description": "The query to process",
				},
			},
			"required": []any{"query"},
		},
		func(args map[string]any, _ agents.ContextVariables) (any, error) {
			query := args["query"].(string)
			return fmt.Sprintf("Info about: %s", query), nil
		},
	)

	agent := agents.NewAgent("ConfiguredAgent")
	agent.Instructions = "You are a helpful assistant. Provide detailed answers and use the get_info tool when needed."
	agent.Tools = []agents.Tool{recursiveTool}

	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Tell me about Go programming language."),
	}

	// Example 1: Using default config
	fmt.Println("=== Example 1: Default Config ===")
	ctx := context.Background()
	result, err := runner.Run(ctx, agent, messages, nil, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		printResults("Default Config", result)
	}

	// Example 2: Custom temperature and max turns
	fmt.Println("\n=== Example 2: Custom Temperature & Max Turns ===")
	temp := 0.3 // Lower temperature for more focused responses
	config := &agents.RunConfig{
		MaxTurns:    3, // Limit to 3 turns
		Temperature: &temp,
		Debug:       true,
	}

	result, err = runner.Run(ctx, agent, messages, nil, config)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		printResults("Custom Config", result)
	}

	// Example 3: Max turns exceeded
	fmt.Println("\n=== Example 3: Max Turns Exceeded ===")
	veryLowTurns := &agents.RunConfig{
		MaxTurns: 1, // Very low limit - will likely exceed
	}

	result, err = runner.Run(ctx, agent, messages, nil, veryLowTurns)
	if err != nil {
		fmt.Printf("❌ Expected error: %v\n", err)
	} else {
		printResults("Low Max Turns", result)
	}

	// Example 4: Timeout
	fmt.Println("\n=== Example 4: Timeout Configuration ===")
	timeoutConfig := &agents.RunConfig{
		Timeout: 10 * time.Second, // 10 second timeout
		Debug:   true,
	}

	result, err = runner.Run(ctx, agent, messages, nil, timeoutConfig)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		printResults("With Timeout", result)
	}

	// Example 5: Usage tracking across multiple calls
	fmt.Println("\n=== Example 5: Usage Tracking Across Multiple Calls ===")
	var totalUsage agents.Usage

	queries := []string{
		"What is 2+2?",
		"What year was Go created?",
		"Who created Go?",
	}

	simpleAgent := agents.NewAgent("SimpleAgent")
	simpleAgent.Instructions = "Answer questions concisely."

	for i, query := range queries {
		msgs := []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(query),
		}

		result, err := runner.Run(ctx, simpleAgent, msgs, nil, nil)
		if err != nil {
			fmt.Printf("Error on query %d: %v\n", i+1, err)
			continue
		}

		totalUsage.Add(result.Usage)
		fmt.Printf("%d. %s → %s\n", i+1, query, result.FinalOutput)
	}

	fmt.Printf("\n📊 Total Usage Across All Queries:\n")
	fmt.Printf("   Prompt Tokens: %d\n", totalUsage.PromptTokens)
	fmt.Printf("   Completion Tokens: %d\n", totalUsage.CompletionTokens)
	fmt.Printf("   Total Tokens: %d\n", totalUsage.TotalTokens)

	// Estimate cost (example rates - adjust based on actual OpenAI pricing)
	promptCost := float64(totalUsage.PromptTokens) * 0.00001         // $0.01 per 1K tokens
	completionCost := float64(totalUsage.CompletionTokens) * 0.00003 // $0.03 per 1K tokens
	totalCost := promptCost + completionCost

	fmt.Printf("   Estimated Cost: $%.6f\n", totalCost)
}

func printResults(label string, result *agents.Result) {
	fmt.Printf("📊 %s Results:\n", label)
	fmt.Printf("   Response: %s\n", result.FinalOutput)
	fmt.Printf("   Steps: %d\n", len(result.Steps))
	fmt.Printf("   Total Tokens: %d (Prompt: %d, Completion: %d)\n",
		result.Usage.TotalTokens,
		result.Usage.PromptTokens,
		result.Usage.CompletionTokens)

	// Show step details
	for i, step := range result.Steps {
		fmt.Printf("   Step %d (%s): %d tool calls, %v\n",
			i+1, step.AgentName, len(step.ToolCalls), step.Duration)
	}
}
