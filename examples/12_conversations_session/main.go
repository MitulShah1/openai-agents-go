// Package main demonstrates using ConversationsSession for cloud-based conversation persistence.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"

	agents "github.com/MitulShah1/openai-agents-go"
	"github.com/MitulShah1/openai-agents-go/session"
)

// This example demonstrates the OpenAI Conversations API session backend.
func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Please set OPENAI_API_KEY environment variable.")
		return
	}

	client := openai.NewClient(option.WithAPIKey(apiKey))
	runner := agents.NewRunner(&client)

	// Create agent
	agent := agents.NewAgent("CloudAssistant")
	agent.Instructions = "You are a helpful assistant with cloud-based persistent memory. Remember conversation context across sessions."

	ctx := context.Background()

	// ===================================================================
	// Example 1: Cloud-Based Conversations Session
	// ===================================================================
	fmt.Println("=== Example 1: Cloud-Based Conversations Session ===")
	fmt.Println("Using OpenAI Conversations API for storage")
	fmt.Println()

	// Create conversations session (cloud-based persistence)
	convSession := session.NewConversationsSession(&client)
	// In a real app, you would persist the mapping of sessionID -> conversationID
	// somewhere (database, Redis, etc.) to access the same conversation later
	// or from another device. The in-memory session only holds it for the process lifetime.
	sessionID := "user_alice_device1"

	// First conversation - store in cloud
	fmt.Println("Device 1 - First conversation:")
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Remember that my favorite color is blue and I live in San Francisco"),
	}

	result, err := runner.Run(
		ctx,
		agent,
		messages,
		agents.WithSession(convSession, sessionID),
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Agent: %s\n\n", result.FinalOutput)

	// Second turn - conversation persists
	fmt.Println("Device 1 - Second message:")
	messages = []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("What's my favorite color?"),
	}

	result, err = runner.Run(
		ctx,
		agent,
		messages,
		agents.WithSession(convSession, sessionID),
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Agent: %s\n\n", result.FinalOutput)

	// ===================================================================
	// Example 2: Multiple Sessions (Simulation)
	// ===================================================================
	fmt.Println("=== Example 2: Multiple Sessions (Simulation) ===")
	fmt.Println("Demonstrating independent session handling")
	fmt.Println()

	// NOTE: Since the basic ConversationsSession stores the ID mapping in memory,
	// creating a new instance simulates a scenario where the mapping is known/unknown.
	// Here we use a different session ID to represent a completely new conversation.
	sessionID2 := "user_bob_device2"

	fmt.Println("Device 2 - New conversation (Bob):")
	messages = []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Where do I live?"),
	}

	// This starts a fresh conversation in the cloud
	result, err = runner.Run(
		ctx,
		agent,
		messages,
		agents.WithSession(convSession, sessionID2),
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Agent: %s\n\n", result.FinalOutput)

	// ===================================================================
	// Example 3: Session Management
	// ===================================================================
	fmt.Println("=== Example 3: Session Management ==")

	// View conversation history
	// Note: We need a short delay as the API is eventually consistent
	time.Sleep(1 * time.Second)

	history, err := convSession.Get(ctx, sessionID)
	if err != nil {
		fmt.Printf("Error getting history: %v\n", err)
		return
	}
	fmt.Printf("Alice's Conversation has %d messages\n", len(history))

	// Delete session completely
	err = convSession.Delete(ctx, sessionID)
	if err != nil {
		fmt.Printf("Error deleting session: %v\n", err)
		return
	}
	fmt.Println("Alice's session deleted from cloud")

	// Verify it's gone (or empty if creating new on next access)
	history, _ = convSession.Get(ctx, sessionID)
	if len(history) == 0 {
		fmt.Println("Verified: Alice's history is empty")
	} else {
		fmt.Printf("Warning: History still exists (%d messages)\n", len(history))
	}

	fmt.Println()
	fmt.Println("=== Conversations Session Demo Complete ===")
	fmt.Println()
	fmt.Println("Benefits of Conversations API Session:")
	fmt.Println("✓ Cloud-based persistence")
	fmt.Println("✓ Built-in pagination and management")
	fmt.Println("✓ Direct integration with OpenAI infrastructure")
}
