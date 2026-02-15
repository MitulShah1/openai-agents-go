// Package main demonstrates the Prompts API for dynamic prompt configuration.
//
// The Prompts API lets agents use centrally managed prompts from OpenAI's
// Prompts API, enabling prompt versioning, A/B testing, and runtime
// customization without code changes.
//
// This example shows three patterns:
//  1. Static prompts — a fixed Prompt with ID, version, and variables
//  2. Dynamic prompts — runtime prompt selection based on context
//  3. Prompt variables — template variable substitution
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"

	agents "github.com/MitulShah1/openai-agents-go"
	"github.com/MitulShah1/openai-agents-go/prompts"
)

func main() {
	fmt.Println("=== Prompts API Demo ===")
	fmt.Println()

	demoStaticPrompt()
	fmt.Println()
	demoDynamicPrompt()
	fmt.Println()
	demoPromptVariables()
}

// demoStaticPrompt shows how to configure a fixed prompt on an agent.
func demoStaticPrompt() {
	fmt.Println("--- 1. Static Prompt ---")

	agent := agents.NewAgent("Assistant")
	agent.Instructions = "You are a helpful assistant."

	// Set a static prompt — the ID references a prompt in OpenAI's Prompts API
	agent.Prompt = &prompts.Prompt{
		ID:      "prompt_helpful_assistant",
		Version: "v2",
	}

	// Resolve the prompt (normally done by the runner automatically)
	resolved, err := agent.GetPrompt(nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Resolved prompt: ID=%s, Version=%s\n", resolved.ID, resolved.Version)
	fmt.Println("The runner passes this prompt to the model via ModelSettings.")
	fmt.Println()

	// To actually run with an API key:
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey != "" {
		client := openai.NewClient(option.WithAPIKey(apiKey))
		runner := agents.NewRunner(&client)

		messages := []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("Hello! What can you help me with?"),
		}

		result, err := runner.Run(context.Background(), agent, messages)
		if err != nil {
			fmt.Printf("Run error (expected if prompt ID doesn't exist): %v\n", err)
			return
		}
		fmt.Printf("Response: %s\n", result.FinalOutput)
	} else {
		fmt.Println("(Set OPENAI_API_KEY to run with a real API)")
	}
}

// demoDynamicPrompt shows runtime prompt selection based on context.
func demoDynamicPrompt() {
	fmt.Println("--- 2. Dynamic Prompt ---")

	agent := agents.NewAgent("Support")
	agent.Instructions = "You are a customer support agent."

	// Use a dynamic prompt that selects based on context variables
	agent.Prompt = prompts.DynamicPromptFunc(func(data prompts.DynamicPromptData) (*prompts.Prompt, error) {
		tier, _ := data.ContextVariables["tier"].(string)

		switch tier {
		case "premium":
			return &prompts.Prompt{
				ID:      "prompt_premium_support",
				Version: "v3",
			}, nil
		case "enterprise":
			return &prompts.Prompt{
				ID:      "prompt_enterprise_support",
				Version: "v1",
			}, nil
		default:
			return &prompts.Prompt{
				ID: "prompt_free_support",
			}, nil
		}
	})

	// Simulate different user tiers
	for _, tier := range []string{"free", "premium", "enterprise"} {
		vars := map[string]any{"tier": tier}
		resolved, err := agent.GetPrompt(vars)
		if err != nil {
			fmt.Printf("  %s tier: Error: %v\n", tier, err)
			continue
		}
		fmt.Printf("  %s tier → prompt ID=%s", tier, resolved.ID)
		if resolved.Version != "" {
			fmt.Printf(", version=%s", resolved.Version)
		}
		fmt.Println()
	}

	fmt.Println()
	fmt.Println("With the runner, pass context variables via WithContextVariables:")
	fmt.Println(`  runner.Run(ctx, agent, messages, agents.WithContextVariables(agents.ContextVariables{"tier": "premium"}))`)
}

// demoPromptVariables shows how to use template variables in prompts.
func demoPromptVariables() {
	fmt.Println("--- 3. Prompt with Variables ---")

	agent := agents.NewAgent("Personalized")
	agent.Instructions = "You are a personalized assistant."

	// Static prompt with template variables
	agent.Prompt = &prompts.Prompt{
		ID:      "prompt_personalized",
		Version: "v1",
		Variables: map[string]any{
			"user_name": "Alice",
			"language":  "English",
			"tone":      "friendly",
		},
	}

	resolved, err := agent.GetPrompt(nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Prompt ID: %s (version: %s)\n", resolved.ID, resolved.Version)
	fmt.Println("Variables:")
	for k, v := range resolved.Variables {
		fmt.Printf("  %s = %v\n", k, v)
	}
	fmt.Println()
	fmt.Println("These variables are substituted in the prompt template by the API.")
}
