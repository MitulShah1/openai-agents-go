// Package main demonstrates the Prompts API for agent configuration.
//
// This example shows three patterns:
//  1. Static prompt - Fixed prompt configuration
//  2. Dynamic prompt - Runtime-resolved prompt based on context
//  3. No prompt - Traditional instructions-only approach
//
// The Prompts API allows agents to use centrally managed prompt configurations
// from OpenAI's Prompts API, enabling prompt versioning, A/B testing, and
// runtime customization without code changes.
//
// Note: This example demonstrates the API surface. In production, prompts
// would be fetched from OpenAI's Prompts API endpoint.
package main

import (
	"fmt"

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

// demoStaticPrompt shows how to configure an agent with a fixed prompt.
func demoStaticPrompt() {
	fmt.Println("--- 1. Static Prompt ---")

	// A static prompt has a fixed ID and optional version
	prompt := &prompts.Prompt{
		ID:      "prompt_helpful_assistant",
		Version: "v2",
	}

	fmt.Printf("  Prompt ID: %s\n", prompt.ID)
	fmt.Printf("  Version: %s\n", prompt.Version)
	fmt.Println()
	fmt.Println("  Usage:")
	fmt.Println("    agent := agents.NewAgent(\"Assistant\")")
	fmt.Println("    agent.Prompt = &prompts.Prompt{")
	fmt.Println("        ID:      \"prompt_helpful_assistant\",")
	fmt.Println("        Version: \"v2\",")
	fmt.Println("    }")
}

// demoDynamicPrompt shows how to select prompts based on runtime context.
func demoDynamicPrompt() {
	fmt.Println("--- 2. Dynamic Prompt ---")

	// A dynamic prompt function receives context and returns the appropriate prompt
	dynamicPrompt := prompts.DynamicPromptFunc(func(data prompts.DynamicPromptData) (*prompts.Prompt, error) {
		// Select prompt based on user tier from context variables
		tier, _ := data.ContextVariables["tier"].(string)

		switch tier {
		case "premium":
			return &prompts.Prompt{
				ID:      "prompt_premium_assistant",
				Version: "v3",
			}, nil
		case "enterprise":
			return &prompts.Prompt{
				ID:      "prompt_enterprise_assistant",
				Version: "v1",
			}, nil
		default:
			return &prompts.Prompt{
				ID: "prompt_free_assistant",
			}, nil
		}
	})

	// Test with different context variables
	testCases := []map[string]any{
		{"tier": "free"},
		{"tier": "premium"},
		{"tier": "enterprise"},
	}

	for _, ctxVars := range testCases {
		data := prompts.DynamicPromptData{
			Agent:            prompts.AgentInfo{Name: "Assistant", Model: "gpt-4o"},
			ContextVariables: ctxVars,
		}

		prompt, _ := dynamicPrompt(data)
		fmt.Printf("  tier=%q → prompt=%s (version=%s)\n",
			ctxVars["tier"], prompt.ID, prompt.Version)
	}

	fmt.Println()
	fmt.Println("  Usage:")
	fmt.Println("    agent.Prompt = prompts.DynamicPromptFunc(func(data prompts.DynamicPromptData) (*prompts.Prompt, error) {")
	fmt.Println("        if data.ContextVariables[\"tier\"] == \"premium\" {")
	fmt.Println("            return &prompts.Prompt{ID: \"prompt_premium\"}, nil")
	fmt.Println("        }")
	fmt.Println("        return &prompts.Prompt{ID: \"prompt_default\"}, nil")
	fmt.Println("    })")
}

// demoPromptVariables shows how to pass template variables to prompts.
func demoPromptVariables() {
	fmt.Println("--- 3. Prompt Variables ---")

	// Variables are substituted in the prompt template at runtime
	prompt := &prompts.Prompt{
		ID:      "prompt_personalized",
		Version: "v1",
		Variables: map[string]any{
			"user_name":  "Alice",
			"language":   "Spanish",
			"tone":       "friendly",
			"max_length": 100,
		},
	}

	fmt.Printf("  Prompt ID: %s\n", prompt.ID)
	fmt.Println("  Variables:")
	for k, v := range prompt.Variables {
		fmt.Printf("    %s: %v\n", k, v)
	}

	fmt.Println()
	fmt.Println("  Usage:")
	fmt.Println("    agent.Prompt = &prompts.Prompt{")
	fmt.Println("        ID: \"prompt_personalized\",")
	fmt.Println("        Variables: map[string]any{")
	fmt.Println("            \"user_name\": userName,")
	fmt.Println("            \"language\":  preferredLang,")
	fmt.Println("        },")
	fmt.Println("    }")

	// Dynamic prompts can also include variables
	fmt.Println()
	fmt.Println("  Dynamic prompts with variables:")

	dynamicWithVars := prompts.DynamicPromptFunc(func(data prompts.DynamicPromptData) (*prompts.Prompt, error) {
		userName, _ := data.ContextVariables["user_name"].(string)
		return &prompts.Prompt{
			ID: "prompt_greeting",
			Variables: map[string]any{
				"name":  userName,
				"agent": data.Agent.Name,
			},
		}, nil
	})

	data := prompts.DynamicPromptData{
		Agent:            prompts.AgentInfo{Name: "Greeter", Model: "gpt-4o"},
		ContextVariables: map[string]any{"user_name": "Bob"},
	}

	resolved, _ := dynamicWithVars(data)
	fmt.Printf("    Resolved: ID=%s, Variables=%v\n", resolved.ID, resolved.Variables)
}
