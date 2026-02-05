// Package main demonstrates basic streaming with the OpenAI Agents Go SDK.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"

	agents "github.com/MitulShah1/openai-agents-go"
	"github.com/MitulShah1/openai-agents-go/stream"
)

func main() {
	// Create OpenAI client
	client := openai.NewClient(option.WithAPIKey(os.Getenv("OPENAI_API_KEY")))

	// Create a simple agent
	agent := &agents.Agent{
		Name:         "Storyteller",
		Model:        "gpt-4o-mini",
		Instructions: "You are a creative storyteller. Tell engaging short stories.",
	}

	// Create runner
	runner := agents.NewRunner(&client)

	// Initial message
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Tell me a short story about a robot learning to paint."),
	}

	// Execute with streaming
	result, err := runner.StreamWithResult(context.Background(), agent, messages)
	if err != nil {
		fmt.Printf("Error starting stream: %v\n", err)
		return
	}

	fmt.Println("=== Streaming Story ===")
	fmt.Println()

	// Stream events and print text as it arrives
	for event, err := range result.StreamEvents(context.Background()) {
		if err != nil {
			fmt.Printf("\nError: %v\n", err)
			return
		}

		// Handle raw response events for text streaming
		if raw, ok := event.(*stream.RawResponseEvent); ok {
			// Print text deltas as they arrive
			if raw.Type == "response.output_text.delta" {
				if data, ok := raw.Data.(map[string]any); ok {
					if delta, ok := data["delta"].(string); ok {
						fmt.Print(delta)
					}
				}
			}
		}
	}

	fmt.Println("\n\n=== Stream Complete ===")
	fmt.Printf("Final output: %v\n", result.FinalOutput)
}
