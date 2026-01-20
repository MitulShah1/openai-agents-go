// Package main demonstrates type-safe tool usage.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"

	agents "github.com/MitulShah1/openai-agents-go"
	"github.com/MitulShah1/openai-agents-go/function"
)

// WeatherArgs defines the arguments for the weather tool
type WeatherArgs struct {
	City string `json:"city" jsonschema:"description=The city name"`
	Unit string `json:"unit,omitempty" jsonschema:"enum=celsius|fahrenheit,description=Temperature unit"`
}

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY is not set")
	}

	client := openai.NewClient(option.WithAPIKey(apiKey))

	// Define a strongly-typed tool using a Go function and struct
	weatherTool := function.FromFunc("get_weather", "Get current weather", func(args WeatherArgs) (any, error) {
		unit := args.Unit
		if unit == "" {
			unit = "celsius"
		}
		return fmt.Sprintf("The weather in %s is 25 degrees %s", args.City, unit), nil
	})

	// Create the agent with the typed tool
	agent := agents.NewAgent("WeatherBot")
	agent.Model = openai.ChatModelGPT4o
	agent.Instructions = "You are a helpful weather assistant."
	agent.Tools = []agents.Tool{weatherTool}

	runner := agents.NewRunner(&client)

	// Run the agent
	ctx := context.Background()
	_, err := runner.Run(ctx, agent, []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("What's the weather in Tokyo in fahrenheit?"),
	})
	if err != nil {
		log.Fatalf("Error running agent: %v", err)
	}
}
