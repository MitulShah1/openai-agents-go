// Package main demonstrates real-time function call argument streaming.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"

	agents "github.com/MitulShah1/openai-agents-go"
	"github.com/MitulShah1/openai-agents-go/stream"
	"github.com/MitulShah1/openai-agents-go/tools"
)

// WeatherArgs represents the arguments for the get_weather tool
type WeatherArgs struct {
	Location string `json:"location" jsonschema:"description=The city and state, e.g. San Francisco, CA"`
	Unit     string `json:"unit,omitempty" jsonschema:"enum=celsius,enum=fahrenheit,description=Temperature unit"`
}

func main() {
	// Create OpenAI client
	client := openai.NewClient(option.WithAPIKey(os.Getenv("OPENAI_API_KEY")))

	// Create weather tool
	weatherTool := tools.New(
		"get_weather",
		"Get the current weather for a location",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"location": map[string]any{
					"type":        "string",
					"description": "The city and state, e.g. San Francisco, CA",
				},
				"unit": map[string]any{
					"type":        "string",
					"enum":        []string{"celsius", "fahrenheit"},
					"description": "Temperature unit",
				},
			},
			"required": []string{"location"},
		},
		func(args map[string]any, _ agents.ContextVariables) (any, error) {
			location := args["location"].(string)
			return fmt.Sprintf("The weather in %s is 72°F and sunny", location), nil
		},
	)

	// Create agent with tool
	agent := &agents.Agent{
		Name:         "WeatherBot",
		Model:        "gpt-4o-mini",
		Instructions: "You are a helpful weather assistant. Use the get_weather tool to answer questions.",
		Tools:        []tools.Tool{weatherTool},
	}

	// Create runner
	runner := agents.NewRunner(&client)

	// Initial message
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("What's the weather like in San Francisco and New York?"),
	}

	// Execute with streaming
	result, err := runner.StreamWithResult(context.Background(), agent, messages)
	if err != nil {
		fmt.Printf("Error starting stream: %v\n", err)
		return
	}

	fmt.Println("=== Streaming Function Call Arguments ===")
	fmt.Println()

	currentFunctionArgs := make(map[int]string)

	// Stream events
	for event, err := range result.StreamEvents(context.Background()) {
		if err != nil {
			fmt.Printf("\nError: %v\n", err)
			return
		}

		// Handle raw response events for function call argument streaming
		if raw, ok := event.(*stream.RawResponseEvent); ok {
			switch raw.Type {
			case "response.function_call_arguments.delta":
				// Function arguments are streaming in real-time!
				if data, ok := raw.Data.(map[string]any); ok {
					outputIndex := int(data["output_index"].(float64))
					delta := data["delta"].(string)

					// Accumulate arguments
					currentFunctionArgs[outputIndex] += delta

					fmt.Printf("📝 Function arg delta [%d]: %s\n", outputIndex, delta)
				}

			case "response.output_item.done":
				// Function call complete
				if data, ok := raw.Data.(map[string]any); ok {
					if item, ok := data["item"].(map[string]any); ok {
						if item["type"] == "function_call" {
							outputIndex := int(data["output_index"].(float64))
							name := item["name"].(string)
							args := item["arguments"].(string)

							fmt.Printf("\n✅ Function call complete [%d]:\n", outputIndex)
							fmt.Printf("   Name: %s\n", name)
							fmt.Printf("   Arguments: %s\n", args)

							// Parse and pretty-print arguments
							var argsMap map[string]any
							if err := json.Unmarshal([]byte(args), &argsMap); err == nil {
								prettyArgs, _ := json.MarshalIndent(argsMap, "   ", "  ")
								fmt.Printf("   Parsed:\n   %s\n\n", string(prettyArgs))
							}
						}
					}
				}
			}
		}

		// Handle semantic events
		if runItem, ok := event.(*stream.RunItemEvent); ok {
			switch runItem.Name {
			case string(stream.ToolCalled):
				fmt.Println("🔧 Tool called")
			case string(stream.ToolOutput):
				fmt.Println("📤 Tool output received")
			case string(stream.MessageOutputCreated):
				fmt.Println("💬 Message output created")
			}
		}
	}

	fmt.Println("\n=== Stream Complete ===")
	fmt.Printf("Final output: %v\n", result.FinalOutput)
}
