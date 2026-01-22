// Package main demonstrates RunConfig usage and token usage tracking.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"

	agents "github.com/MitulShah1/openai-agents-go"
	"github.com/MitulShah1/openai-agents-go/tools"
)

// This example demonstrates RunConfig usage and token usage tracking.
func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Please set OPENAI_API_KEY environment variable.")
		return
	}

	client := openai.NewClient(option.WithAPIKey(apiKey))
	runner := agents.NewRunner(&client)

	// Define a simple tool
	weatherTool := tools.New(
		"get_weather",
		"Get weather",
		map[string]any{"type": "object"},
		func(_ map[string]any, _ agents.ContextVariables) (any, error) {
			return "Sunny, 25C", nil
		},
	)

	// A recursive tool that calls itself (to test MaxTurns)
	recursiveTool := tools.New(
		"recursive_call",
		"Calls itself indefinitely",
		map[string]any{"type": "object"},
		func(_ map[string]any, _ agents.ContextVariables) (any, error) {
			return "calling again...", nil
		},
	)

	agent := agents.NewAgent("ConfiguredAgent")
	agent.Instructions = "You are a helpful assistant. Provide detailed answers and use the get_info tool when needed."
	agent.Tools = []tools.Tool{weatherTool, recursiveTool}

	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Tell me about Go programming language."),
	}

	ctx := context.Background()

	// 1. Run with MaxTokens limit and Temperature
	fmt.Println("--- Run 1: Limited Tokens ---")
	maxTokens := 50
	temp := 0.2
	config1 := &agents.RunConfig{
		MaxTurns:    5,
		MaxTokens:   &maxTokens, // Fix pointer
		Temperature: &temp,      // Fix pointer
	}
	result1, err := runner.Run(
		ctx,
		agent,
		messages,
		agents.WithConfig(config1),
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Output: %s\n", result1.FinalOutput)
		fmt.Printf("Usage: %+v\n", result1.Usage)
	}

	// 2. Run with Timeout
	fmt.Println("\n--- Run 2: Timeout ---")
	config2 := &agents.RunConfig{
		Timeout: 5 * time.Second,
	}
	_, err = runner.Run(
		ctx,
		agent,
		messages,
		agents.WithConfig(config2),
	)
	if err != nil {
		fmt.Printf("Run finished (possibly with error): %v\n", err)
	}

	// 3. Run with MaxTurns check (using recursive loop)
	fmt.Println("\n--- Run 3: MaxTurns Limit ---")
	agent.Instructions = "Call recursive_call tool repeatedly."
	config3 := &agents.RunConfig{
		MaxTurns: 3,
	}
	result3, err := runner.Run(
		ctx,
		agent,
		[]openai.ChatCompletionMessageParamUnion{openai.UserMessage("Call recursive_call please.")},
		agents.WithConfig(config3),
	)
	if err != nil {
		fmt.Printf("Expected error or success: %v\n", err)
	} else {
		fmt.Printf("Turns used: %d\n", len(result3.Steps))
	}
}
