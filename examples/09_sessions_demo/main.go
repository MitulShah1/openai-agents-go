// Package main demonstrates session management for persistent conversations.
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"

	agents "github.com/MitulShah1/openai-agents-go"
	"github.com/MitulShah1/openai-agents-go/session"
)

// This example demonstrates how to use sessions to maintain conversation history.
func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Please set OPENAI_API_KEY environment variable.")
		return
	}

	client := openai.NewClient(option.WithAPIKey(apiKey))
	runner := agents.NewRunner(&client)

	// Create agent
	agent := agents.NewAgent("Assistant")
	agent.Instructions = "You are a helpful assistant. Remember conversation context and answer concisely."

	ctx := context.Background()

	// ===================================================================
	// Example 1: In-Memory Session (lost when program exits)
	// ===================================================================
	fmt.Println("=== Example 1: In-Memory Session ===")
	memSession := session.NewMemorySession()
	sessionID := "user_123"

	// First turn
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("My name is Alice"),
	}
	result, err := runner.Run(ctx, agent, messages, nil, nil, memSession, sessionID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Turn 1: %s\n", result.FinalOutput)

	// Second turn - agent remembers the name
	messages = []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("What's my name?"),
	}
	result, err = runner.Run(ctx, agent, messages, nil, nil, memSession, sessionID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Turn 2: %s\n\n", result.FinalOutput)

	// ===================================================================
	// Example 2: File-Based Session (persistent across restarts)
	// ===================================================================
	fmt.Println("=== Example 2: File-Based Session ===")

	// Create sessions directory
	sessionsDir := filepath.Join(os.TempDir(), "agent-sessions")
	fileSession, err := session.NewFileSession(sessionsDir)
	if err != nil {
		fmt.Printf("Error creating file session: %v\n", err)
		return
	}
	defer os.RemoveAll(sessionsDir)

	sessionID2 := "user_456"

	// First conversation
	messages = []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("I live in San Francisco"),
	}
	result, err = runner.Run(ctx, agent, messages, nil, nil, fileSession, sessionID2)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Turn 1: %s\n", result.FinalOutput)

	// Second conversation - uses saved history
	messages = []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("What city did I mention?"),
	}
	result, err = runner.Run(ctx, agent, messages, nil, nil, fileSession, sessionID2)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Turn 2: %s\n\n", result.FinalOutput)

	// ===================================================================
	// Example 3: Multiple Sessions (Isolated Conversations)
	// ===================================================================
	fmt.Println("=== Example 3: Multiple Isolated Sessions ===")

	// Session A
	sessionA := "alice_session"
	messages = []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("I work at Google"),
	}
	resultA, _ := runner.Run(ctx, agent, messages, nil, nil, fileSession, sessionA)
	fmt.Printf("Alice: %s\n", resultA.FinalOutput)

	// Session B (different user, different history)
	sessionB := "bob_session"
	messages = []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("I work at Microsoft"),
	}
	resultB, _ := runner.Run(ctx, agent, messages, nil, nil, fileSession, sessionB)
	fmt.Printf("Bob: %s\n", resultB.FinalOutput)

	// Ask same question to both - they should remember their own company
	messages = []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Where do I work?"),
	}

	resultA, _ = runner.Run(ctx, agent, messages, nil, nil, fileSession, sessionA)
	fmt.Printf("Alice remembers: %s\n", resultA.FinalOutput)

	resultB, _ = runner.Run(ctx, agent, messages, nil, nil, fileSession, sessionB)
	fmt.Printf("Bob remembers: %s\n\n", resultB.FinalOutput)

	// ===================================================================
	// Example 4: Session Management Operations
	// ===================================================================
	fmt.Println("=== Example 4: Session Management ===")

	sessionID3 := "user_789"

	// Create some history
	messages = []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Remember: my favorite color is blue"),
	}
	runner.Run(ctx, agent, messages, nil, nil, fileSession, sessionID3)

	// Get session history
	history, err := fileSession.Get(ctx, sessionID3)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Session has %d messages\n", len(history))

	// Clear session
	err = fileSession.Clear(ctx, sessionID3)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println("Session cleared")

	// Verify it's empty
	history, _ = fileSession.Get(ctx, sessionID3)
	fmt.Printf("After clear: %d messages\n", len(history))

	// Delete session completely
	err = fileSession.Delete(ctx, sessionID3)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println("Session deleted")

	fmt.Println("\n=== Sessions Demo Complete ===")
	fmt.Println("Sessions enable multi-turn conversations with persistent memory!")
}
