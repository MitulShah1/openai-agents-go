// Package main demonstrates the usage of tracing in the openai-agents-go SDK.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"

	agents "github.com/MitulShah1/openai-agents-go"
	"github.com/MitulShah1/openai-agents-go/tools"
	"github.com/MitulShah1/openai-agents-go/tracing"
	"github.com/MitulShah1/openai-agents-go/tracing/exporter"
	"github.com/MitulShah1/openai-agents-go/tracing/processor"
)

func main() {
	// Create OpenAI client
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	client := openai.NewClient(option.WithAPIKey(apiKey))
	runner := agents.NewRunner(&client)

	// Configure tracing with console exporter
	// This will print trace information to stdout
	exp := exporter.NewConsoleExporter()
	proc := processor.NewBatch(exp)
	provider := tracing.NewProvider(proc)
	tracing.SetProvider(provider)

	// Create a simple agent with a tool
	agent := &agents.Agent{
		Name:         "Calculator Agent",
		Instructions: "You are a helpful calculator assistant.",
		Model:        "gpt-4o-mini",
		Tools: []tools.Tool{
			{
				Name:        "add",
				Description: "Add two numbers",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"a": map[string]any{"type": "number"},
						"b": map[string]any{"type": "number"},
					},
					"required": []string{"a", "b"},
				},
				Callback: func(args map[string]any, _ tools.ContextVariables) (any, error) {
					a := args["a"].(float64)
					b := args["b"].(float64)
					return a + b, nil
				},
			},
		},
	}

	// Example 1: Basic run with default tracing
	fmt.Println("=== Example 1: Basic Run with Default Tracing ===")

	ctx := context.Background()
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("What is 15 + 27?"),
	}

	result, err := runner.Run(ctx, agent, messages)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Final output: %s\n\n", result.FinalOutput)

	// Example 2: Custom trace configuration
	fmt.Println("=== Example 2: Custom Trace Configuration ===")

	customConfig := &agents.RunConfig{
		MaxTurns:          5,
		TraceWorkflowName: "Calculator Workflow",
		TraceGroupID:      "example-group-1",
		TraceMetadata: map[string]any{
			"user_id":    "user-123",
			"session_id": "session-456",
		},
	}

	messages = []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Calculate 42 + 58"),
	}

	result, err = runner.Run(ctx, agent, messages,
		agents.WithConfig(customConfig),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Final output: %s\n\n", result.FinalOutput)

	// Example 3: Disabling tracing
	fmt.Println("=== Example 3: Tracing Disabled ===")

	tracingDisabled := false
	disabledConfig := &agents.RunConfig{
		MaxTurns:     5,
		TraceEnabled: &tracingDisabled,
	}

	messages = []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("What is 10 + 20?"),
	}

	result, err = runner.Run(ctx, agent, messages,
		agents.WithConfig(disabledConfig),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Final output: %s\n", result.FinalOutput)
	fmt.Println("(No trace output expected)")

	// Shutdown tracing provider to flush remaining spans
	if err := provider.Shutdown(ctx); err != nil {
		log.Printf("Error shutting down tracing: %v", err)
	}
}
