// Package main demonstrates guardrails usage for input/output validation.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"

	agents "github.com/MitulShah1/openai-agents-go"
	"github.com/MitulShah1/openai-agents-go/guardrail"
	"github.com/MitulShah1/openai-agents-go/guardrail/builtin"
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
		builtin.NewPIIGuardrail(
			builtin.WithTripwire(true), // Halt if PII detected
			builtin.WithEmailDetection(true),
			builtin.WithPhoneDetection(true),
			builtin.WithSSNDetection(true),
		),

		// Block URLs from untrusted domains
		builtin.NewURLFilterGuardrail(
			builtin.WithBlocklist("evil.com", "*.malicious.org"),
			builtin.WithURLTripwire(true),
		),

		// Block forbidden keywords
		builtin.NewRegexGuardrail(
			`\b(password|secret|token)\b`,
			builtin.WithMustMatch(false), // Pattern must NOT match
			builtin.WithRegexTripwire(true),
			builtin.WithRegexMessage("Please don't share sensitive credentials"),
		),
	}

	// Add output guardrails - validate agent responses
	agent.OutputGuardrails = []*guardrail.Guardrail{
		// Ensure agent doesn't leak PII in responses
		builtin.NewPIIGuardrail(
			builtin.WithTripwire(true),
		),
	}

	ctx := context.Background()

	// Example 1: Valid input
	fmt.Println("=== Example 1: Valid Input ===")
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("What is the capital of France?"),
	}

	result, err := runner.Run(ctx, agent, messages, nil, nil, nil, "")
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

	result, err = runner.Run(ctx, agent, messages, nil, nil, nil, "")
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

	result, err = runner.Run(ctx, agent, messages, nil, nil, nil, "")
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

	result, err = runner.Run(ctx, agent, messages, nil, nil, nil, "")
	if err != nil {
		fmt.Printf("✅ Guardrail blocked: %v\n\n", err)
	} else {
		fmt.Printf("Response: %s\n\n", result.FinalOutput)
	}

	fmt.Println("=== Guardrails Demo Complete ===")
	fmt.Println("Guardrails protect both user inputs and agent outputs!")
}
