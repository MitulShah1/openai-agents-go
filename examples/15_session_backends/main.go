// Package main demonstrates database session backends introduced in v0.3.0.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"

	agents "github.com/MitulShah1/openai-agents-go"
	"github.com/MitulShah1/openai-agents-go/session"
)

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	client := openai.NewClient(option.WithAPIKey(apiKey))
	runner := agents.NewRunner(&client)

	// Example 1: SQLite backend with file persistence
	fmt.Println("=== Example 1: SQLite File Backend ===")
	sqliteSession, err := session.NewSQLite("./sessions.db")
	if err != nil {
		log.Fatalf("Failed to create SQLite session: %v", err)
	}
	defer func() {
		if sess, ok := sqliteSession.(*session.SQLiteSession); ok {
			if err := sess.Close(); err != nil {
				log.Printf("Error closing SQLite session: %v", err)
			}
		}
	}()

	agent := agents.NewAgent("helpful_assistant")
	agent.Instructions = "You are a helpful assistant. Keep responses concise."
	agent.Model = openai.ChatModelGPT4oMini

	userID := "user123"

	// First conversation
	fmt.Println("\nFirst conversation:")
	messages1 := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Remember: my favorite color is blue"),
	}
	result, err := runner.Run(context.Background(), agent, messages1, agents.WithSession(sqliteSession, userID))
	if err != nil {
		log.Printf("Run failed: %v", err)
		return
	}
	fmt.Printf("  Agent: %s\n", result.FinalOutput)

	// Second conversation (should remember)
	fmt.Println("\nSecond conversation:")
	messages2 := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("What's my favorite color?"),
	}
	result, err = runner.Run(context.Background(), agent, messages2, agents.WithSession(sqliteSession, userID))
	if err != nil {
		log.Printf("Run failed: %v", err)
		return
	}
	fmt.Printf("  Agent: %s\n", result.FinalOutput)

	// Example 2: In-memory SQLite (no persistence)
	fmt.Println("\n=== Example 2: SQLite In-Memory Backend ===")
	memSession, err := session.NewSQLite(":memory:")
	if err != nil {
		log.Printf("Failed to create in-memory session: %v", err)
		return
	}
	defer func() {
		if sess, ok := memSession.(*session.SQLiteSession); ok {
			if err := sess.Close(); err != nil {
				log.Printf("Error closing in-memory session: %v", err)
			}
		}
	}()

	tempUserID := "temp_user"

	messages3 := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("This conversation won't persist after the program ends"),
	}
	result, err = runner.Run(context.Background(), agent, messages3, agents.WithSession(memSession, tempUserID))
	if err != nil {
		log.Printf("Run failed: %v", err)
		return
	}
	fmt.Printf("  Agent: %s\n", result.FinalOutput)

	// Example 3: Using the registry pattern
	fmt.Println("\n=== Example 3: Session Registry Pattern ===")

	// Create session via registry
	config := map[string]any{
		"path": "./file_sessions.db",
	}
	fileSession, err := session.Create("sqlite", config)
	if err != nil {
		log.Printf("Failed to create session from registry: %v", err)
		return
	}

	registryUserID := "registry_user"

	messages4 := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("This uses the sqlite backend via registry"),
	}
	result, err = runner.Run(context.Background(), agent, messages4, agents.WithSession(fileSession, registryUserID))
	if err != nil {
		log.Printf("Run failed: %v", err)
		return
	}
	fmt.Printf("  Agent: %s\n", result.FinalOutput)

	// Example 4: Session management operations
	fmt.Println("\n=== Example 4: Session Management ===")

	// Get session history
	messages, err := sqliteSession.Get(context.Background(), userID)
	if err != nil {
		log.Printf("Failed to get session: %v", err)
		return
	}
	fmt.Printf("  Session history for %s: %d messages\n", userID, len(messages))

	// Clear session
	err = sqliteSession.Clear(context.Background(), userID)
	if err != nil {
		log.Printf("Failed to clear session: %v", err)
		return
	}
	fmt.Printf("  Cleared session for %s\n", userID)

	// Verify cleared
	messages, err = sqliteSession.Get(context.Background(), userID)
	if err != nil {
		log.Printf("Failed to get session: %v", err)
		return
	}
	fmt.Printf("  Session history after clear: %d messages\n", len(messages))

	// Delete session
	err = sqliteSession.Delete(context.Background(), userID)
	if err != nil {
		log.Printf("Failed to delete session: %v", err)
		return
	}
	fmt.Printf("  Deleted session for %s\n", userID)

	fmt.Println("\n✅ Session Backends Demo Complete!")
	fmt.Println("\nKey Features Demonstrated:")
	fmt.Println("  • SQLite file-based persistence")
	fmt.Println("  • SQLite in-memory mode")
	fmt.Println("  • Registry pattern for backend selection")
	fmt.Println("  • Session CRUD operations (Get, Clear, Delete)")
	fmt.Println("  • Multi-user session management")
}
