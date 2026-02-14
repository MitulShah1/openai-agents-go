// Package main demonstrates the Model Provider abstraction.
//
// This example shows three patterns:
//  1. Default provider - Using NewRunner with an OpenAI client
//  2. Custom provider - Using NewRunnerWithProvider for explicit control
//  3. Multi-provider routing - Using MultiProvider for prefix-based model selection
//
// The Model Provider abstraction allows:
//   - Swapping LLM backends (OpenAI, Anthropic, Google, etc.)
//   - Agent-level provider overrides
//   - Prefix-based routing to different providers in the same app
//
// Note: This example demonstrates the API surface without making real API calls.
package main

import (
	"context"
	"fmt"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/packages/ssestream"

	"github.com/MitulShah1/openai-agents-go/models"
)

func main() {
	fmt.Println("=== Model Provider Abstraction Demo ===")
	fmt.Println()

	demoDefaultProvider()
	fmt.Println()
	demoCustomProvider()
	fmt.Println()
	demoMultiProvider()
	fmt.Println()
	demoAgentLevelProvider()
}

// demoDefaultProvider shows the backward-compatible NewRunner approach.
func demoDefaultProvider() {
	fmt.Println("--- 1. Default Provider (Backward Compatible) ---")

	// NewRunner automatically wraps the client in an OpenAIProvider
	client := openai.NewClient(option.WithAPIKey("sk-demo-key"))
	_ = client // Would use: runner := agents.NewRunner(&client)

	fmt.Println("  The classic NewRunner(client) pattern still works:")
	fmt.Println()
	fmt.Println("    client := openai.NewClient(option.WithAPIKey(apiKey))")
	fmt.Println("    runner := agents.NewRunner(&client)")
	fmt.Println()
	fmt.Println("  Internally, this wraps the client in an OpenAIProvider.")
}

// demoCustomProvider shows explicit provider creation.
func demoCustomProvider() {
	fmt.Println("--- 2. Custom Provider (Explicit) ---")

	// Create an OpenAI provider explicitly
	client := openai.NewClient(option.WithAPIKey("sk-demo-key"))
	provider := models.NewOpenAIProvider(&client)

	// Get a model from the provider
	model, err := provider.GetModel("gpt-4o")
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
		return
	}

	fmt.Printf("  Provider created model: %s\n", model.ModelName())
	fmt.Println()
	fmt.Println("  Usage:")
	fmt.Println("    provider := models.NewOpenAIProvider(&client)")
	fmt.Println("    runner := agents.NewRunnerWithProvider(provider)")
	fmt.Println()
	fmt.Println("  Or access the underlying client for direct API calls:")
	fmt.Println("    openaiProvider.Client().Chat.Completions.New(...)")
}

// demoMultiProvider shows prefix-based routing to different providers.
func demoMultiProvider() {
	fmt.Println("--- 3. Multi-Provider Routing ---")

	// Create providers for different backends
	openaiClient := openai.NewClient(option.WithAPIKey("sk-openai-key"))
	openaiProvider := models.NewOpenAIProvider(&openaiClient)

	// For demo: create a mock provider for Anthropic
	// In production, you'd implement models.ModelProvider for Anthropic's API
	anthropicProvider := &mockProvider{name: "anthropic"}
	googleProvider := &mockProvider{name: "google"}

	// Create a multi-provider with prefix routing
	multiProvider := models.NewMultiProvider(
		openaiProvider, // Default (no prefix)
		models.WithProviderPrefix("anthropic", anthropicProvider),
		models.WithProviderPrefix("google", googleProvider),
	)

	// Test model name routing
	testModels := []string{
		"gpt-4o",                    // → openaiProvider (default)
		"gpt-4o-mini",               // → openaiProvider (default)
		"anthropic/claude-3-sonnet", // → anthropicProvider
		"google/gemini-pro",         // → googleProvider
	}

	for _, modelName := range testModels {
		model, err := multiProvider.GetModel(modelName)
		if err != nil {
			fmt.Printf("  %s → error: %v\n", modelName, err)
			continue
		}
		fmt.Printf("  %s → resolved to %s\n", modelName, model.ModelName())
	}

	fmt.Println()
	fmt.Println("  Usage:")
	fmt.Println("    mp := models.NewMultiProvider(")
	fmt.Println("        openaiProvider,  // default for unprefixed names")
	fmt.Println("        models.WithProviderPrefix(\"anthropic\", anthropicProvider),")
	fmt.Println("        models.WithProviderPrefix(\"google\", googleProvider),")
	fmt.Println("    )")
	fmt.Println("    runner := agents.NewRunnerWithProvider(mp)")
	fmt.Println()
	fmt.Println("  Then set agent.Model to route to different providers:")
	fmt.Println("    agent.Model = \"gpt-4o\"                    // OpenAI")
	fmt.Println("    agent.Model = \"anthropic/claude-3-sonnet\" // Anthropic")
	fmt.Println("    agent.Model = \"google/gemini-pro\"         // Google")
}

// demoAgentLevelProvider shows per-agent provider overrides.
func demoAgentLevelProvider() {
	fmt.Println("--- 4. Agent-Level Provider Override ---")

	fmt.Println("  Agents can override the runner's default provider:")
	fmt.Println()
	fmt.Println("    // Runner uses OpenAI by default")
	fmt.Println("    runner := agents.NewRunnerWithProvider(openaiProvider)")
	fmt.Println()
	fmt.Println("    // Most agents use the default")
	fmt.Println("    agent1 := agents.NewAgent(\"Assistant\")")
	fmt.Println("    agent1.Model = \"gpt-4o\"")
	fmt.Println()
	fmt.Println("    // This agent uses a different provider")
	fmt.Println("    agent2 := agents.NewAgent(\"Specialist\")")
	fmt.Println("    agent2.Model = \"claude-3-sonnet\"")
	fmt.Println("    agent2.ModelProvider = anthropicProvider")
	fmt.Println()
	fmt.Println("  Resolution order:")
	fmt.Println("    1. Agent.ModelProvider (if set)")
	fmt.Println("    2. Runner's default provider")
	fmt.Println("    3. Fallback to wrapped OpenAI client (backward compat)")
}

// mockProvider is a simple mock for demonstrating multi-provider routing.
type mockProvider struct {
	name string
}

func (p *mockProvider) GetModel(modelName string) (models.Model, error) {
	return &mockModel{name: modelName, provider: p.name}, nil
}

type mockModel struct {
	name     string
	provider string
}

func (m *mockModel) ModelName() string {
	return fmt.Sprintf("%s (via %s)", m.name, m.provider)
}

func (m *mockModel) GetResponse(_ context.Context, _ openai.ChatCompletionNewParams, _ models.ModelSettings) (*models.ModelResponse, error) {
	return nil, fmt.Errorf("mock model: not implemented")
}

func (m *mockModel) StreamResponse(_ context.Context, _ openai.ChatCompletionNewParams, _ models.ModelSettings) (*ssestream.Stream[openai.ChatCompletionChunk], error) {
	return nil, fmt.Errorf("mock model: not implemented")
}
