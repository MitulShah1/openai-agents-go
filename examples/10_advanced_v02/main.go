// Package main demonstrates combining guardrails and sessions for production-ready agents.
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"

	agents "github.com/MitulShah1/openai-agents-go"
	"github.com/MitulShah1/openai-agents-go/guardrail"
	"github.com/MitulShah1/openai-agents-go/guardrail/builtin"
	"github.com/MitulShah1/openai-agents-go/session"
)

// This example demonstrates a production-ready chatbot with both guardrails and sessions.
func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Please set OPENAI_API_KEY environment variable.")
		return
	}

	client := openai.NewClient(option.WithAPIKey(apiKey))
	runner := agents.NewRunner(&client)

	// Create a production-ready agent with safety guardrails
	agent := agents.NewAgent("ProductionChatbot")
	agent.Instructions = `You are a helpful customer service assistant.
- Be professional and courteous
- Answer questions accurately
- Never share or ask for sensitive personal information
- Keep responses concise`

	// Configure input guardrails for safety
	agent.InputGuardrails = []*guardrail.Guardrail{
		// Protect against PII leakage
		builtin.NewPIIGuardrail(
			builtin.WithEmailDetection(true),
			builtin.WithPhoneDetection(true),
			builtin.WithSSNDetection(true),
			builtin.WithCreditCardDetection(true),
			builtin.WithTripwire(false), // Log but don't block
		),

		// Block malicious URLs
		builtin.NewURLFilterGuardrail(
			builtin.WithBlocklist("*.malware.com", "phishing.net"),
			builtin.WithURLTripwire(true),
		),
	}

	// Configure output guardrails
	agent.OutputGuardrails = []*guardrail.Guardrail{
		// Ensure agent never leaks PII
		builtin.NewPIIGuardrail(
			builtin.WithTripwire(true), // Strict for outputs
		),

		// Ensure professional responses (no profanity)
		builtin.NewRegexGuardrail(
			`\b(damn|hell|crap)\b`,
			builtin.WithMustMatch(false),
			builtin.WithRegexMessage("Response contains unprofessional language"),
		),
	}

	// Setup persistent file-based session
	sessionsDir := filepath.Join(os.TempDir(), "production-sessions")
	fileSession, err := session.NewFileSession(sessionsDir)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer os.RemoveAll(sessionsDir)

	ctx := context.Background()

	// Simulate a multi-turn customer conversation
	fmt.Println("=== Production Chatbot Demo ===\n")

	userID := "customer_001"

	// Turn 1: Customer asks about product
	fmt.Println("Customer: What products do you offer?")
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("What products do you offer?"),
	}
	result, err := runner.Run(ctx, agent, messages, nil, nil, fileSession, userID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Agent: %s\n\n", result.FinalOutput)

	// Turn 2: Customer asks follow-up (uses session memory)
	fmt.Println("Customer: How much does the first one cost?")
	messages = []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("How much does the first one cost?"),
	}
	result, err = runner.Run(ctx, agent, messages, nil, nil, fileSession, userID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Agent: %s\n\n", result.FinalOutput)

	// Turn 3: Customer tries to share PII (should be caught)
	fmt.Println("Customer: My email is customer@example.com, can you send me details?")
	messages = []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("My email is customer@example.com, can you send me details?"),
	}
	result, err = runner.Run(ctx, agent, messages, nil, nil, fileSession, userID)
	if err != nil {
		// Guardrail blocked the request
		fmt.Printf("⚠️  Safety Alert: %v\n", err)
		fmt.Println("(In production, log this and ask user to rephrase)\n")
	} else {
		fmt.Printf("Agent: %s\n\n", result.FinalOutput)
	}

	// Turn 4: Customer asks valid question
	fmt.Println("Customer: What's your return policy?")
	messages = []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("What's your return policy?"),
	}
	result, err = runner.Run(ctx, agent, messages, nil, nil, fileSession, userID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Agent: %s\n\n", result.FinalOutput)

	// Show session management
	fmt.Println("=== Session Management ===")
	history, _ := fileSession.Get(ctx, userID)
	fmt.Printf("Session '%s' has %d messages stored\n", userID, len(history))
	fmt.Printf("Conversation persisted to: %s/%s.json\n\n", sessionsDir, userID)

	// Show usage tracking
	fmt.Println("=== Usage Tracking ===")
	fmt.Printf("Total Tokens Used: %d\n", result.Usage.TotalTokens)
	fmt.Printf("  - Prompt Tokens: %d\n", result.Usage.PromptTokens)
	fmt.Printf("  - Completion Tokens: %d\n", result.Usage.CompletionTokens)

	estimatedCost := float64(result.Usage.TotalTokens) * 0.00002 // Rough estimate
	fmt.Printf("Estimated Cost: $%.6f\n\n", estimatedCost)

	fmt.Println("=== Production Features Demonstrated ===")
	fmt.Println("✅ Input guardrails (PII detection, URL filtering)")
	fmt.Println("✅ Output guardrails (PII protection, profanity filter)")
	fmt.Println("✅ Persistent sessions (conversation memory)")
	fmt.Println("✅ Multi-turn conversations")
	fmt.Println("✅ Usage tracking and cost estimation")
	fmt.Println("\nThis chatbot is production-ready!")
}
