// Package main demonstrates a production-ready chatbot combining all v0.3.0 features.
// Features: SQLite persistence, Guardrail composition with metrics, Multimodal tool outputs.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"

	agents "github.com/MitulShah1/openai-agents-go"
	"github.com/MitulShah1/openai-agents-go/guardrail"
	"github.com/MitulShah1/openai-agents-go/guardrail/content"
	"github.com/MitulShah1/openai-agents-go/guardrail/moderation"
	"github.com/MitulShah1/openai-agents-go/guardrail/security"
	"github.com/MitulShah1/openai-agents-go/session"
)

func main() {
	// 1. Setup OpenAI Client
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}
	client := openai.NewClient(option.WithAPIKey(apiKey))

	// 2. Setup Database Session Backend (SQLite)
	// In production, this persists conversation history to a file.
	sessionBackend, err := session.NewSQLite("./production_sessions.db")
	if err != nil {
		log.Fatalf("Failed to initialize SQLite backend: %v", err)
	}
	// Ensure we close connections on exit (though in a server, this would be on shutdown)
	if closer, ok := sessionBackend.(interface{ Close() error }); ok {
		defer func() { _ = closer.Close() }()
	}

	// 3. Setup Guardrails with Metrics
	metrics := guardrail.NewInMemoryMetrics()

	// Create a validation chain:
	// - Must be reasonable length (5-2000 chars)
	// - No profanity
	// - No secrets (PII logic)
	// - All wrapped with metrics tracking
	inputGuardrail := guardrail.NewChain().
		Add(guardrail.WithMetrics(content.NewLength(content.Config{
			Mode:     content.CountModeCharacters,
			Min:      5,
			Max:      2000,
			Tripwire: true,
		}), metrics)).
		Add(guardrail.WithMetrics(moderation.NewProfanity(moderation.ProfanityConfig{
			Tripwire: true,
		}), metrics)).
		Add(guardrail.WithMetrics(security.NewSecrets(security.SecretsConfig{
			Tripwire: true,
		}), metrics)).
		WithStrategy(guardrail.Sequential).
		WithName("chatbot_input_validation").
		Build()

	// 4. Setup Multimodal Tools
	// Example tool: Generate a chart (returns an image)
	chartTool := agents.FunctionTool(
		"generate_chart",
		"Generates a sales chart for a given year",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"year": map[string]any{
					"type":        "integer",
					"description": "Year to generate chart for",
				},
				"title": map[string]any{
					"type":        "string",
					"description": "Chart title",
				},
			},
			"required": []string{"year"},
		},
		func(args map[string]any, _ agents.ContextVariables) (any, error) {
			year := int(args["year"].(float64))
			title := "Sales Chart"
			if t, ok := args["title"].(string); ok {
				title = t
			}

			// In a real app, this would generate a real image.
			// Here we use a placeholder service.
			url := fmt.Sprintf("https://via.placeholder.com/800x400.png?text=%s+%d", title, year)

			// Return rich content (Image)
			return agents.ImageContent(url, "high"), nil
		},
	)

	// 5. Setup Agent
	agent := agents.NewAgent("ProductionBot")
	agent.Model = openai.ChatModelGPT4oMini
	agent.Instructions = `You are a capable production assistant.
	- You can generate charts using the generate_chart tool.
	- Your inputs are validated for safety.
	- Your memory is persistent.`
	agent.Tools = []agents.Tool{chartTool}
	// Attach guardrails to the agent
	agent.InputGuardrails = []*guardrail.Guardrail{inputGuardrail}

	// 6. Create Runner
	runner := agents.NewRunner(&client)

	// 7. Simulate User Interaction
	sessionID := "user_session_001"
	ctx := context.Background()

	// Clear previous session for cleanliness in this demo
	_ = sessionBackend.Clear(ctx, sessionID)

	fmt.Println("=== Production Chatbot Demo ===")

	// Turn 1: User asks for a chart
	userMsg1 := "Can you show me the sales chart for 2024?"
	fmt.Printf("\nUser: %s\n", userMsg1)

	result1, err := runner.Run(ctx, agent,
		[]openai.ChatCompletionMessageParamUnion{openai.UserMessage(userMsg1)},
		agents.WithSession(sessionBackend, sessionID),
	)
	if err != nil {
		log.Printf("Run 1 failed: %v", err)
		return
	}
	fmt.Printf("Bot:  %s\n", result1.FinalOutput)

	// Turn 2: User follows up (testing memory)
	userMsg2 := "What about for 2025?"
	fmt.Printf("\nUser: %s\n", userMsg2)

	result2, err := runner.Run(ctx, agent,
		[]openai.ChatCompletionMessageParamUnion{openai.UserMessage(userMsg2)},
		agents.WithSession(sessionBackend, sessionID),
	)
	if err != nil {
		log.Printf("Run 2 failed: %v", err)
		return
	}
	fmt.Printf("Bot:  %s\n", result2.FinalOutput)

	// Turn 3: Guardrail Check (Profanity)
	userMsg3 := "You are a damn fool" // Should be caught by profanity filter
	fmt.Printf("\nUser: %s\n", userMsg3)

	_, err = runner.Run(ctx, agent,
		[]openai.ChatCompletionMessageParamUnion{openai.UserMessage(userMsg3)},
		agents.WithSession(sessionBackend, sessionID),
	)
	if err != nil {
		fmt.Printf("Bot:  [Blocked by Guardrail] %v\n", err)
	} else {
		// Should not happen if guardrail works
		fmt.Printf("Bot:  (Unexpectedly allowed)\n")
	}

	// 8. Print Metrics
	fmt.Println("\n=== Performance Metrics ===")
	stats := metrics.GetAllStats()
	for name, s := range stats {
		if s.TotalCount > 0 {
			fmt.Printf("[%s] Executions: %d | Pass: %d | Fail: %d | Avg Latency: %v\n",
				name, s.TotalCount, s.PassedCount, s.FailedCount, s.AvgDuration())
		}
	}
}
