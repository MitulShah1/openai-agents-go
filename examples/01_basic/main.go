// Package main demonstrates the simplest possible agent - a basic conversation with no tools.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"

	agents "github.com/MitulShah1/openai-agents-go"
)

// This is the simplest possible agent - just a basic conversation with no tools.
func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Please set OPENAI_API_KEY environment variable.")
		return
	}

	// Initialize the OpenAI client
	client := openai.NewClient(option.WithAPIKey(apiKey))
	runner := agents.NewRunner(&client)

	// Create a basic agent with default settings
	agent := agents.NewAgent("BasicAgent")
	agent.Instructions = "You are a helpful assistant. Be concise and friendly."

	// Run the agent with a simple question
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("What is the capital of France?"),
	}

	ctx := context.Background()
	result, err := runner.Run(ctx, agent, messages, nil, nil, nil, "")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Print the response
	fmt.Printf("Agent: %s\n", result.FinalOutput)
	fmt.Printf("\nTokens used: %d\n", result.Usage.TotalTokens)
}
