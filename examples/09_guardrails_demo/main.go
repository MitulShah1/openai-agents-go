// Package main demonstrates guardrails usage for input/output validation.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"

	agents "github.com/MitulShah1/openai-agents-go"
	"github.com/MitulShah1/openai-agents-go/guardrail"
	"github.com/MitulShah1/openai-agents-go/guardrail/content"
	"github.com/MitulShah1/openai-agents-go/guardrail/security"
)

// This example demonstrates how to use guardrails to validate agent inputs and outputs.
func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Please set OPENAI_API_KEY environment variable.")
		return
	}

	client := openai.NewClient(option.WithAPIKey(apiKey))
	runner := agents.NewRunner(&client)

	// Create agent with guardrails
	agent := agents.NewAgent("SafeAssistant")
	agent.Instructions = "You are a helpful assistant. Answer questions concisely."

	// Add input guardrails - validate user input before processing
	agent.InputGuardrails = []*guardrail.Guardrail{
		// Detect PII in user input
		security.NewPII(
			security.WithTripwire(true), // Halt if PII detected
			security.WithEmailDetection(true),
			security.WithPhoneDetection(true),
			security.WithSSNDetection(true),
		),

		// Block URLs from untrusted domains
		security.NewURLFilter(
			security.WithBlocklist("evil.com", "*.malicious.org"),
			security.WithURLTripwire(true),
		),

		// Block forbidden keywords
		content.NewRegex(
			`\b(password|secret|token)\b`,
			content.WithMustMatch(false), // Pattern must NOT match
			content.WithRegexTripwire(true),
			content.WithRegexMessage("Please don't share sensitive credentials"),
		),
	}

	// Add output guardrails - validate agent responses
	agent.OutputGuardrails = []*guardrail.Guardrail{
		// Ensure agent doesn't leak PII in responses
		security.NewPII(
			security.WithTripwire(true),
		),
	}

	ctx := context.Background()

	// Example 1: Valid input
	fmt.Println("=== Example 1: Valid Input ===")
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("What is the capital of France?"),
	}

	result, err := runner.Run(ctx, agent, messages)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Response: %s\n\n", result.FinalOutput)
	}

	// Example 2: Input with PII (should be blocked)
	fmt.Println("=== Example 2: Input with PII (Blocked) ===")
	messages = []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("My email is john@example.com, can you help?"),
	}

	result, err = runner.Run(ctx, agent, messages)
	if err != nil {
		fmt.Printf("✅ Guardrail blocked: %v\n\n", err)
	} else {
		fmt.Printf("Response: %s\n\n", result.FinalOutput)
	}

	// Example 3: Input with forbidden keyword (should be blocked)
	fmt.Println("=== Example 3: Forbidden Keyword (Blocked) ===")
	messages = []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("What's my password for the account?"),
	}

	result, err = runner.Run(ctx, agent, messages)
	if err != nil {
		fmt.Printf("✅ Guardrail blocked: %v\n\n", err)
	} else {
		fmt.Printf("Response: %s\n\n", result.FinalOutput)
	}

	// Example 4: Input with blocked URL
	fmt.Println("=== Example 4: Blocked URL (Blocked) ===")
	messages = []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Check out this link: https://evil.com/malware"),
	}

	result, err = runner.Run(ctx, agent, messages)
	if err != nil {
		fmt.Printf("✅ Guardrail blocked: %v\n\n", err)
	} else {
		fmt.Printf("Response: %s\n\n", result.FinalOutput)
	}

	fmt.Println("=== Guardrails Demo Complete ===")
	fmt.Println("Guardrails protect both user inputs and agent outputs!")
}
