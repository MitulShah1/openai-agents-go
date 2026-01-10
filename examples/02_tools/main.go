// Package main demonstrates how to use tools with agents.
package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"

	agents "github.com/MitulShah1/openai-agents-go"
)

// This example demonstrates how to use tools with agents.
// Tools allow agents to perform actions like getting the weather, searching the web, etc.
func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Please set OPENAI_API_KEY environment variable.")
		return
	}

	client := openai.NewClient(option.WithAPIKey(apiKey))
	runner := agents.NewRunner(&client)

	// Define a weather tool
	getWeather := agents.FunctionTool(
		"get_weather",
		"Get the current weather in a given location",
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
					"description": "The temperature unit to use",
				},
			},
			"required": []any{"location"},
		},
		func(args map[string]any, _ agents.ContextVariables) (any, error) {
			location := args["location"].(string)
			unit, ok := args["unit"].(string)
			if !ok {
				unit = "fahrenheit"
			}

			// Simulate API call
			temp := "72"
			if unit == "celsius" {
				temp = "22"
			}

			return fmt.Sprintf("The weather in %s is %s°%s and sunny.", location, temp, strings.ToUpper(unit[:1])), nil
		},
	)

	// Define a time tool
	getTime := agents.FunctionTool(
		"get_current_time",
		"Get the current time in a given timezone",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"timezone": map[string]any{
					"type":        "string",
					"description": "The timezone, e.g. America/New_York",
				},
			},
			"required": []any{"timezone"},
		},
		func(args map[string]any, _ agents.ContextVariables) (any, error) {
			timezone := args["timezone"].(string)

			// For simplicity, just return current UTC time
			// In real app, you'd use time.LoadLocation()
			now := time.Now().UTC()
			return fmt.Sprintf("Current time in %s: %s", timezone, now.Format("15:04:05 MST")), nil
		},
	)

	// Create agent with multiple tools
	agent := agents.NewAgent("ToolsAgent")
	agent.Instructions = "You are a helpful assistant with access to weather and time information. Use the tools to answer user questions."
	agent.Tools = []agents.Tool{getWeather, getTime}

	// Run with a question that requires tool usage
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("What's the weather in Paris and what time is it in New York right now?"),
	}

	ctx := context.Background()
	result, err := runner.Run(ctx, agent, messages, nil, nil, nil, "")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Print the conversation
	fmt.Printf("Agent: %s\n\n", result.FinalOutput)

	// Show execution details
	fmt.Printf("=== Execution Details ===\n")
	fmt.Printf("Total Steps: %d\n", len(result.Steps))
	fmt.Printf("Tokens Used: %d\n", result.Usage.TotalTokens)

	fmt.Printf("\n=== Tool Calls ===\n")
	for _, step := range result.Steps {
		for _, toolCall := range step.ToolCalls {
			fmt.Printf("- %s: %v\n", toolCall.ToolName, toolCall.Result)
		}
	}
}
