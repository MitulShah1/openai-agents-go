// Package main demonstrates how to return multimodal content (images) from tools.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"

	agents "github.com/MitulShah1/openai-agents-go"
	"github.com/MitulShah1/openai-agents-go/tools"
)

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Please set OPENAI_API_KEY environment variable.")
		return
	}

	client := openai.NewClient(option.WithAPIKey(apiKey))

	// Define an agent that uses multimodal tools
	agent := agents.NewAgent("MultimodalBot")
	agent.Instructions = "You are a helpful assistant that can generate images and data files."

	// Tool 1: Generate an image (returns an image URL)
	imageGenTool := tools.New(
		"generate_image",
		"Generates an image based on a prompt",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"prompt": map[string]any{
					"type":        "string",
					"description": "The prompt to generate image from",
				},
			},
		},
		func(args map[string]any, _ agents.ContextVariables) (any, error) {
			prompt, _ := args["prompt"].(string)
			fmt.Printf("Generating image for: %s\n", prompt)

			// In a real app, this would call DALL-E
			// Here we return a dummy image URL wrapped in ImageContent
			return tools.ImageContent("https://example.com/generated-image.png", "high"), nil
		},
	)

	// Tool 2: Analyze data (returns a plot/file)
	dataAnalysisTool := tools.New(
		"analyze_data",
		"Analyzes dataset and returns a plot",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"dataset": map[string]any{
					"type": "string",
				},
			},
		},
		func(args map[string]any, _ agents.ContextVariables) (any, error) {
			dataset, _ := args["dataset"].(string)
			fmt.Printf("Analyzing dataset: %s\n", dataset)

			// Simulate generating a CSV file
			csvData := []byte("id,value\n1,100\n2,200")
			return tools.FileContent(csvData, "analysis.csv", "text/csv"), nil
		},
	)

	// Tool 3: Simple text response (backward compatibility)
	getWeatherTool := tools.New(
		"get_weather",
		"Get current weather",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"city": map[string]any{"type": "string"},
			},
		},
		func(_ map[string]any, _ agents.ContextVariables) (any, error) {
			return "Sunny and warm", nil
		},
	)

	// Tool 4: Generate report (retained for completeness from original, fixed)
	createReportTool := tools.New(
		"create_report",
		"Creates a report",
		map[string]any{"type": "object"},
		func(_ map[string]any, _ agents.ContextVariables) (any, error) {
			return tools.TextContent("Report created."), nil
		},
	)

	agent.Tools = []tools.Tool{imageGenTool, dataAnalysisTool, getWeatherTool, createReportTool}

	// Create runner
	runner := agents.NewRunner(&client)

	// Run the agent
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Generate an image of a sunset and analyze the sales dataset."),
	}

	result, err := runner.Run(
		context.Background(), // Context needs import
		agent,
		messages,
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Final output: %s\n", result.FinalOutput)
}
