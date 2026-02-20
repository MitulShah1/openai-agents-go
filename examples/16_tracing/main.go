// Package main demonstrates the usage of tracing in the openai-agents-go SDK.
// This example specifically showcases how complex agent interactions, such as
// multi-agent handoffs and tool executions, are captured in the OpenAI Tracing Dashboard.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"

	agents "github.com/MitulShah1/openai-agents-go"
	"github.com/MitulShah1/openai-agents-go/handoff"
	"github.com/MitulShah1/openai-agents-go/tools"
	"github.com/MitulShah1/openai-agents-go/tracing"
	"github.com/MitulShah1/openai-agents-go/tracing/exporter"
	"github.com/MitulShah1/openai-agents-go/tracing/processor"
)

func tracingPreflight() string {
	apiKey := strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))
	if apiKey == "" {
		log.Fatal("missing required env: OPENAI_API_KEY")
	}

	org := strings.TrimSpace(os.Getenv("OPENAI_ORG_ID"))
	project := strings.TrimSpace(os.Getenv("OPENAI_PROJECT_ID"))

	fmt.Println("=== Tracing Requirements Preflight ===")
	fmt.Printf("OPENAI_API_KEY: set (required)\n")

	if org == "" {
		fmt.Println("OPENAI_ORG_ID: not set (optional)")
	} else {
		fmt.Println("OPENAI_ORG_ID: set (optional)")
	}

	if project == "" {
		fmt.Println("OPENAI_PROJECT_ID: not set (optional)")
	} else {
		fmt.Println("OPENAI_PROJECT_ID: set (optional)")
	}

	fmt.Println("Exporter mode: backend (sends traces to OpenAI dashboard)")
	fmt.Println()

	return apiKey
}

func main() {

	// Create OpenAI client
	apiKey := tracingPreflight()

	client := openai.NewClient(option.WithAPIKey(apiKey))
	runner := agents.NewRunner(&client)

	// Configure tracing with backend exporter (sends to OpenAI dashboard)
	// We use the backend exporter so that you can view the execution paths natively
	// in the OpenAI platform under the Observability -> Traces tab.
	exp := exporter.NewBackendExporter()
	proc := processor.NewBatch(exp)
	provider := tracing.NewProvider(proc)
	tracing.SetProvider(provider)
	ctx := context.Background()

	// IMPORTANT: Deferred shutdown to flush pending traces before process exit
	defer func() { _ = tracing.Shutdown(ctx) }()

	// Create the math agent which contains a math tool
	mathAgent := &agents.Agent{
		Name:         "Math Agent",
		Instructions: "You are a helpful calculator assistant. If someone asks a math question, use your tools to solve it.",
		Model:        "gpt-4o-mini",
		Tools: []tools.Tool{
			{
				Name:        "add",
				Description: "Add two numbers",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"a": map[string]any{"type": "number"},
						"b": map[string]any{"type": "number"},
					},
					"required": []string{"a", "b"},
				},
				Callback: func(args map[string]any, _ tools.ContextVariables) (any, error) {
					a := args["a"].(float64)
					b := args["b"].(float64)
					fmt.Printf("[Math Agent] -> Tool 'add' executed: %v + %v\n", a, b)
					return a + b, nil
				},
			},
		},
	}

	// Create a triage agent that acts as a router
	triageAgent := &agents.Agent{
		Name:         "Triage Agent",
		Instructions: "You are a customer service router. If the user asks a math question, you must IMMEDIATELY hand off to the Math Agent.",
		Model:        "gpt-4o-mini",
		Tools: []tools.Tool{
			handoff.New(mathAgent, handoff.WithDescription("Transfer the user to the Math Agent for any calculation queries.")).ToTool(),
		},
	}

	// Example: Visualizing Handoffs and Tools in the Dashboard
	fmt.Println("=== Example: Tracking Handoffs and Tool Calls ===")
	fmt.Println("This run starts with the Triage Agent.")
	fmt.Println("The output trace will show a nested hierarchy: Workflow -> Triage Agent -> Handoff -> Math Agent -> Tool -> Generation")
	fmt.Println()

	customConfig := &agents.RunConfig{
		MaxTurns:          5,
		TraceWorkflowName: "Multi-Agent Math Support",
		TraceMetadata: map[string]any{
			"example_type": "handoffs_and_tools",
		},
	}

	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Can you help me? I need to know what 244 + 511 is."),
	}

	result, err := runner.Run(ctx, triageAgent, messages, agents.WithConfig(customConfig))
	if err != nil {
		log.Printf("Runner error: %v", err)
	} else {
		fmt.Printf("\nFinal output: %s\n\n", result.FinalOutput)
		fmt.Println("Traces are now being flushed to the OpenAI Dashboard...")
		fmt.Println("Check the OpenAI platform under Observability -> Traces to view.")
		fmt.Println("Note: It may take up to 60 seconds for traces to appear on the server.")
	}
}
