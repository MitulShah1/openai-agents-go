// Package main demonstrates structured outputs with OpenAI agents.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"

	agents "github.com/MitulShah1/openai-agents-go"
	"github.com/MitulShah1/openai-agents-go/internal/jsonschema"
)

// MathReasoning represents the structured output we expect
type MathReasoning struct {
	Steps      []string `json:"steps"`
	Answer     int      `json:"answer"`
	Confidence string   `json:"confidence"`
}

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable not set")
	}

	// Initialize OpenAI client
	client := openai.NewClient(option.WithAPIKey(apiKey))
	runner := agents.NewRunner(&client)

	// Define JSON schema for math reasoning
	reasoningSchema := jsonschema.Object().
		WithDescription("Step-by-step math reasoning").
		WithProperty("steps", jsonschema.Array(jsonschema.String()).
			WithDescription("Array of reasoning steps")).
		WithProperty("answer", jsonschema.Integer().
			WithDescription("The final numerical answer")).
		WithProperty("confidence", jsonschema.String().
			WithEnum("low", "medium", "high").
			WithDescription("Confidence level in the answer")).
		WithRequired("steps", "answer", "confidence")

	// Create agent with structured output
	agent := agents.NewAgent("Math Tutor")
	agent.Instructions = "You are a helpful math tutor. Solve problems step by step."
	agent.ResponseFormat = jsonschema.JSONSchema("math_reasoning", reasoningSchema).
		WithDescription("Structured mathematical reasoning")

	// Problem to solve
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("What is 157 + 2584? Show your reasoning."),
	}

	ctx := context.Background()
	result, err := runner.Run(ctx, agent, messages, nil, nil)
	if err != nil {
		log.Fatalf("Error running agent: %v", err)
	}

	// Parse the structured response
	if len(result.Messages) > 0 {
		// Get the final output which should be structured JSON
		if result.FinalOutput != "" {
			var reasoning MathReasoning
			if err := json.Unmarshal([]byte(result.FinalOutput), &reasoning); err != nil {
				log.Fatalf("Error parsing response: %v", err)
			}

			fmt.Println("Math Reasoning (Structured Output):")
			fmt.Println("====================================")
			fmt.Println("\nSteps:")
			for i, step := range reasoning.Steps {
				fmt.Printf("%d. %s\n", i+1, step)
			}
			fmt.Printf("\nFinal Answer: %d\n", reasoning.Answer)
			fmt.Printf("Confidence: %s\n", reasoning.Confidence)
		}
	}

	fmt.Printf("\nTokens used: %d\n", result.Usage.TotalTokens)
}
