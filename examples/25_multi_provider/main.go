// Package main demonstrates the Model Provider Abstraction.
//
// The models package decouples the runner from a specific LLM backend,
// allowing agents to use different providers (OpenAI, Anthropic, etc.)
// without changing runner or agent code.
//
// This example shows:
//  1. Default provider — NewRunner(client) auto-creates an OpenAIProvider
//  2. Explicit provider — NewRunnerWithProvider for custom providers
//  3. Per-agent providers — different agents using different backends
//  4. MultiProvider — prefix-based routing (e.g., "openai/gpt-4o")
package main

import (
	"fmt"
	"os"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"

	agents "github.com/MitulShah1/openai-agents-go"
	"github.com/MitulShah1/openai-agents-go/models"
)

func main() {
	fmt.Println("=== Model Provider Abstraction Demo ===")
	fmt.Println()

	demoDefaultProvider()
	fmt.Println()
	demoExplicitProvider()
	fmt.Println()
	demoPerAgentProvider()
	fmt.Println()
	demoMultiProvider()
}

// demoDefaultProvider shows that NewRunner(client) works exactly as before.
func demoDefaultProvider() {
	fmt.Println("--- 1. Default Provider (Backward Compatible) ---")

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		apiKey = "sk-placeholder"
	}
	client := openai.NewClient(option.WithAPIKey(apiKey))
	runner := agents.NewRunner(&client)

	fmt.Printf("Runner created with client: ModelProvider is %T\n", runner.ModelProvider)
	fmt.Println("All existing code continues to work unchanged.")
}

// demoExplicitProvider shows NewRunnerWithProvider for custom providers.
func demoExplicitProvider() {
	fmt.Println("--- 2. Explicit Provider ---")

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		apiKey = "sk-placeholder"
	}
	client := openai.NewClient(option.WithAPIKey(apiKey))

	provider := models.NewOpenAIProvider(&client)
	runner := agents.NewRunnerWithProvider(provider)

	fmt.Printf("Runner created with provider: %T\n", runner.ModelProvider)
	fmt.Println("No Client field needed — the provider handles all model resolution.")

	agent := agents.NewAgent("Assistant")
	agent.Model = "gpt-4o-mini"
	fmt.Printf("Agent model %q will be resolved by the provider\n", agent.Model)

	_ = runner
}

// demoPerAgentProvider shows how different agents can use different providers.
func demoPerAgentProvider() {
	fmt.Println("--- 3. Per-Agent Providers ---")

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		apiKey = "sk-placeholder"
	}
	client := openai.NewClient(option.WithAPIKey(apiKey))

	runner := agents.NewRunner(&client)

	defaultAgent := agents.NewAgent("DefaultAgent")
	defaultAgent.Instructions = "You use the runner's default provider."
	fmt.Printf("  %s: uses runner's default provider\n", defaultAgent.Name)

	customProvider := models.NewOpenAIProvider(&client)
	customAgent := agents.NewAgent("CustomAgent")
	customAgent.Instructions = "You use a custom provider."
	customAgent.ModelProvider = customProvider
	fmt.Printf("  %s: uses its own provider (%T)\n", customAgent.Name, customAgent.ModelProvider)

	_ = runner

	fmt.Println()
	fmt.Println("Resolution order: Agent.ModelProvider → Runner.ModelProvider → Runner.Client")
}

// demoMultiProvider shows prefix-based routing for multiple backends.
func demoMultiProvider() {
	fmt.Println("--- 4. MultiProvider (Prefix Routing) ---")

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		apiKey = "sk-placeholder"
	}
	client := openai.NewClient(option.WithAPIKey(apiKey))

	openaiProvider := models.NewOpenAIProvider(&client)

	multi := models.NewMultiProvider(openaiProvider,
		models.WithProviderPrefix("openai", openaiProvider),
	)

	runner := agents.NewRunnerWithProvider(multi)

	agent1 := agents.NewAgent("GPT Agent")
	agent1.Model = "gpt-4o"
	fmt.Printf("  %s: model=%q → default provider (OpenAI)\n", agent1.Name, agent1.Model)

	agent2 := agents.NewAgent("GPT Mini Agent")
	agent2.Model = "openai/gpt-4o-mini"
	fmt.Printf("  %s: model=%q → openai provider\n", agent2.Name, agent2.Model)

	_ = runner

	fmt.Println()
	fmt.Println("MultiProvider enables mixing models from different backends in one application.")
}
