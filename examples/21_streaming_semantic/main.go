// Package main demonstrates semantic event streaming for high-level progress tracking.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"

	agents "github.com/MitulShah1/openai-agents-go"
	"github.com/MitulShah1/openai-agents-go/stream"
	"github.com/MitulShah1/openai-agents-go/tools"
)

func main() {
	// Create OpenAI client
	client := openai.NewClient(option.WithAPIKey(os.Getenv("OPENAI_API_KEY")))

	// Create a calculator tool
	calculatorTool := tools.New(
		"calculate",
		"Perform a calculation",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"expression": map[string]any{
					"type":        "string",
					"description": "The mathematical expression to evaluate",
				},
			},
			"required": []string{"expression"},
		},
		func(args map[string]any, _ agents.ContextVariables) (any, error) {
			expression := args["expression"].(string)
			return fmt.Sprintf("Result: 42 (evaluated: %s)", expression), nil
		},
	)

	// Create agent with tool
	agent := &agents.Agent{
		Name:         "Assistant",
		Model:        "gpt-4o-mini",
		Instructions: "You are a helpful assistant. Use the calculator when needed.",
		Tools:        []tools.Tool{calculatorTool},
	}

	// Create runner
	runner := agents.NewRunner(&client)

	// Initial message
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Can you calculate 25 * 17 for me?"),
	}

	// Execute with streaming
	result, err := runner.StreamWithResult(context.Background(), agent, messages)
	if err != nil {
		fmt.Printf("Error starting stream: %v\n", err)
		return
	}

	fmt.Println("=== Streaming Semantic Events ===")
	fmt.Println()

	// Track progress with semantic events
	for event, err := range result.StreamEvents(context.Background()) {
		if err != nil {
			fmt.Printf("\nError: %v\n", err)
			return
		}

		// Ignore raw response events - focus on semantic events
		if _, ok := event.(*stream.RawResponseEvent); ok {
			continue
		}

		// Handle semantic events
		switch e := event.(type) {
		case *stream.RunItemEvent:
			switch e.Name {
			case string(stream.MessageOutputCreated):
				fmt.Println("💬 Agent is responding...")

			case string(stream.ToolCalled):
				fmt.Println("🔧 Tool called")
				// In a real app, you could show a loading indicator

			case string(stream.ToolOutput):
				fmt.Println("📤 Tool completed")
				// In a real app, you could update the UI with the result
			}

		case *stream.AgentUpdatedEvent:
			// Agent changed during handoff
			if agentMap, ok := e.NewAgent.(map[string]any); ok {
				if name, ok := agentMap["name"].(string); ok {
					fmt.Printf("👤 Now talking to: %s\n", name)
				}
			}
		}
	}

	fmt.Println("\n=== Stream Complete ===")
	fmt.Printf("Final output: %v\n", result.FinalOutput)
	fmt.Printf("Current agent: %v\n", result.CurrentAgent)
	fmt.Printf("Total turns: %d\n", result.CurrentTurn)
}
