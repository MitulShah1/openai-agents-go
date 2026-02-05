// Package main demonstrates parallel tool execution in the OpenAI Agents Go SDK.
// It shows three execution modes: parallel, sequential, and limited concurrency.
package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"

	agents "github.com/MitulShah1/openai-agents-go"
	"github.com/MitulShah1/openai-agents-go/tools"
)

// Simulated slow API call tools
func weatherTool(_ map[string]any, _ tools.ContextVariables) (any, error) {
	fmt.Println("🌤️  Fetching weather data...")
	time.Sleep(2 * time.Second) // Simulate API call
	return "Weather: 72°F, Sunny", nil
}

func newsTool(_ map[string]any, _ tools.ContextVariables) (any, error) {
	fmt.Println("📰 Fetching latest news...")
	time.Sleep(2 * time.Second) // Simulate API call
	return "News: Tech stocks rise 5%", nil
}

func stocksTool(_ map[string]any, _ tools.ContextVariables) (any, error) {
	fmt.Println("📈 Fetching stock prices...")
	time.Sleep(2 * time.Second) // Simulate API call
	return "Stocks: AAPL $180, GOOGL $140", nil
}

func main() {
	// Initialize OpenAI client
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("❌ OPENAI_API_KEY environment variable not set")
		fmt.Println("This example demonstrates parallel tool execution without making actual API calls")
		fmt.Println("Set OPENAI_API_KEY to see the full example in action")
		demonstrateParallelExecution()
		return
	}

	client := openai.NewClient(option.WithAPIKey(apiKey))
	runner := agents.NewRunner(&client)

	// Create agent with multiple tools
	agent := agents.NewAgent("info-agent")
	agent.Instructions = "You are a helpful assistant that can fetch weather, news, and stock information. Use all available tools to provide comprehensive information."
	agent.Model = openai.ChatModelGPT4oMini
	agent.ParallelToolCalls = true // Enable parallel execution (default)
	agent.Tools = []tools.Tool{
		tools.New("get_weather", "Get current weather information", nil, weatherTool),
		tools.New("get_news", "Get latest news headlines", nil, newsTool),
		tools.New("get_stocks", "Get current stock prices", nil, stocksTool),
	}

	ctx := context.Background()

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🚀 PARALLEL TOOL EXECUTION DEMO")
	fmt.Println(strings.Repeat("=", 60))

	// Demo 1: Parallel Execution (Default)
	fmt.Println("\n📊 Demo 1: Parallel Execution (Default)")
	fmt.Println("Tools will execute concurrently using goroutines")
	fmt.Println(strings.Repeat("-", 60))

	start := time.Now()
	result, err := runner.Run(
		ctx,
		agent,
		[]openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("Give me the weather, latest news, and stock prices"),
		},
	)
	parallelDuration := time.Since(start)

	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	fmt.Printf("\n✅ Parallel Execution Complete!\n")
	fmt.Printf("⏱️  Duration: %v\n", parallelDuration)
	fmt.Printf("📝 Response: %s\n", result.FinalOutput)
	fmt.Printf("🔧 Tools Called: %d\n", len(result.Steps[0].ToolCalls))

	// Demo 2: Sequential Execution
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("📊 Demo 2: Sequential Execution")
	fmt.Println("Tools will execute one at a time")
	fmt.Println(strings.Repeat("-", 60))

	start = time.Now()
	result, err = runner.Run(
		ctx,
		agent,
		[]openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("Give me the weather, latest news, and stock prices"),
		},
		agents.WithConfig(&agents.RunConfig{
			ParallelToolCalls: boolPtr(false), // Disable parallel execution
		}),
	)
	sequentialDuration := time.Since(start)

	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	fmt.Printf("\n✅ Sequential Execution Complete!\n")
	fmt.Printf("⏱️  Duration: %v\n", sequentialDuration)
	fmt.Printf("📝 Response: %s\n", result.FinalOutput)
	fmt.Printf("🔧 Tools Called: %d\n", len(result.Steps[0].ToolCalls))

	// Demo 3: Concurrency Limiting
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("📊 Demo 3: Limited Concurrency")
	fmt.Println("Maximum 2 tools executing at once")
	fmt.Println(strings.Repeat("-", 60))

	start = time.Now()
	result, err = runner.Run(
		ctx,
		agent,
		[]openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("Give me the weather, latest news, and stock prices"),
		},
		agents.WithConfig(&agents.RunConfig{
			ParallelToolCalls:  boolPtr(true),
			MaxToolConcurrency: 2, // Limit to 2 concurrent tools
		}),
	)
	limitedDuration := time.Since(start)

	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	fmt.Printf("\n✅ Limited Concurrency Complete!\n")
	fmt.Printf("⏱️  Duration: %v\n", limitedDuration)
	fmt.Printf("📝 Response: %s\n", result.FinalOutput)
	fmt.Printf("🔧 Tools Called: %d\n", len(result.Steps[0].ToolCalls))

	// Performance Comparison
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("📈 PERFORMANCE COMPARISON")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Parallel:    %v\n", parallelDuration)
	fmt.Printf("Sequential:  %v (%.1fx slower)\n", sequentialDuration, float64(sequentialDuration)/float64(parallelDuration))
	fmt.Printf("Limited (2): %v (%.1fx slower)\n", limitedDuration, float64(limitedDuration)/float64(parallelDuration))

	speedup := float64(sequentialDuration) / float64(parallelDuration)
	fmt.Printf("\n🚀 Parallel execution is %.1fx faster!\n", speedup)
}

// demonstrateParallelExecution shows the concept without OpenAI API
func demonstrateParallelExecution() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎯 PARALLEL EXECUTION CONCEPT DEMO")
	fmt.Println(strings.Repeat("=", 60))

	// Simulate parallel execution
	fmt.Println("\n📊 Simulating Parallel Execution...")
	start := time.Now()

	// Execute 3 tools in parallel
	done := make(chan bool, 3)
	go func() { _, _ = weatherTool(nil, nil); done <- true }()
	go func() { _, _ = newsTool(nil, nil); done <- true }()
	go func() { _, _ = stocksTool(nil, nil); done <- true }()

	// Wait for all to complete
	for i := 0; i < 3; i++ {
		<-done
	}
	parallelDuration := time.Since(start)

	// Simulate sequential execution
	fmt.Println("\n📊 Simulating Sequential Execution...")
	start = time.Now()
	_, _ = weatherTool(nil, nil)
	_, _ = newsTool(nil, nil)
	_, _ = stocksTool(nil, nil)
	sequentialDuration := time.Since(start)

	// Show results
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("📈 RESULTS")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Parallel:   %v (3 tools running simultaneously)\n", parallelDuration)
	fmt.Printf("Sequential: %v (3 tools running one by one)\n", sequentialDuration)
	fmt.Printf("\n🚀 Parallel is %.1fx faster!\n", float64(sequentialDuration)/float64(parallelDuration))

	fmt.Println("\n💡 Key Takeaways:")
	fmt.Println("  • Parallel execution uses goroutines for concurrent tool calls")
	fmt.Println("  • Best for I/O-bound tools (API calls, database queries)")
	fmt.Println("  • Can limit concurrency to prevent resource exhaustion")
	fmt.Println("  • Results maintain original tool call order")
}

func boolPtr(b bool) *bool {
	return &b
}

// Helper for string repetition
