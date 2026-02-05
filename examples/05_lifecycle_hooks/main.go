// Package main demonstrates lifecycle hooks mechanism.
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

// This example demonstrates lifecycle hooks - functions that run before and after agent execution.
// Useful for logging, setup/teardown, tracing, etc.
func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Please set OPENAI_API_KEY environment variable.")
		return
	}

	client := openai.NewClient(option.WithAPIKey(apiKey))
	runner := agents.NewRunner(&client)

	// Define a simple tool
	myTool := tools.New(
		"my_tool",
		"A test tool",
		map[string]any{},
		func(_ map[string]any, _ agents.ContextVariables) (any, error) {
			return "tool result", nil
		},
	)

	// Create agent
	agent := agents.NewAgent("LifecycleAgent")
	agent.Instructions = "You are a helpful assistant."
	agent.Tools = []tools.Tool{myTool}

	// Track execution time
	var startTime time.Time

	// Add Hooks
	agent.OnBeforeRun = func(_ context.Context, agent *agents.Agent) error {
		startTime = time.Now()
		fmt.Printf("Starting execution for agent: %s at %v\n", agent.Name, startTime)
		return nil
	}

	agent.OnAfterRun = func(_ context.Context, agent *agents.Agent) error {
		duration := time.Since(startTime)
		fmt.Printf("Finished execution for agent: %s. Duration: %v\n", agent.Name, duration)
		return nil
	}

	// Run
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Hello, please call my_tool."),
	}

	ctx := context.Background()
	_, err := runner.Run(ctx, agent, messages)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
