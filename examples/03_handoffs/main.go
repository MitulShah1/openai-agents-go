// Package main demonstrates agent handoffs - transferring conversations between specialized agents.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"

	agents "github.com/MitulShah1/openai-agents-go"
)

// This example demonstrates agent handoffs - transferring a conversation from one agent to another.
// This is useful for creating specialized agents that handle different parts of a conversation.
func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Please set OPENAI_API_KEY environment variable.")
		return
	}

	client := openai.NewClient(option.WithAPIKey(apiKey))
	runner := agents.NewRunner(&client)

	// Create specialized agents
	salesAgent := agents.NewAgent("SalesAgent")
	salesAgent.Instructions = "You are a sales agent. Help customers with purchases and product information. If the customer has a technical question, transfer them to technical support."

	supportAgent := agents.NewAgent("TechnicalSupport")
	supportAgent.Instructions = "You are a technical support agent. Help customers solve technical problems. Be detailed and patient."

	// Create handoff tool for sales agent to transfer to support
	transferToSupport := agents.FunctionTool(
		"transfer_to_support",
		"Transfer the customer to technical support for technical questions",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"reason": map[string]any{
					"type":        "string",
					"description": "The reason for the transfer",
				},
			},
			"required": []any{"reason"},
		},
		func(args map[string]any, _ agents.ContextVariables) (any, error) {
			reason := args["reason"].(string)
			fmt.Printf("📞 Transferring to support. Reason: %s\n\n", reason)
			// Return the support agent to trigger handoff
			return supportAgent, nil
		},
	)

	// Create handoff tool for support agent to transfer back to sales
	transferToSales := agents.FunctionTool(
		"transfer_to_sales",
		"Transfer the customer back to sales for purchase assistance",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"reason": map[string]any{
					"type":        "string",
					"description": "The reason for the transfer",
				},
			},
			"required": []any{"reason"},
		},
		func(args map[string]any, _ agents.ContextVariables) (any, error) {
			reason := args["reason"].(string)
			fmt.Printf("📞 Transferring to sales. Reason: %s\n\n", reason)
			return salesAgent, nil
		},
	)

	// Assign tools to agents
	salesAgent.Tools = []agents.Tool{transferToSupport}
	supportAgent.Tools = []agents.Tool{transferToSales}

	// Start with sales agent
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Hi! I'm interested in buying a laptop, but I'm having trouble connecting to Wi-Fi on my current one. Can you help?"),
	}

	ctx := context.Background()
	result, err := runner.Run(ctx, salesAgent, messages, nil, nil, nil, "")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Print the conversation
	fmt.Printf("=== Conversation ===\n")
	fmt.Printf("Final Agent: %s\n", result.Agent.Name)
	fmt.Printf("Response: %s\n\n", result.FinalOutput)

	// Show execution trace
	fmt.Printf("=== Execution Trace ===\n")
	for i, step := range result.Steps {
		fmt.Printf("Step %d - Agent: %s\n", i+1, step.AgentName)
		for _, toolCall := range step.ToolCalls {
			if toolCall.ToolName == "transfer_to_support" || toolCall.ToolName == "transfer_to_sales" {
				fmt.Printf("  → Handoff: %s\n", toolCall.ToolName)
			}
		}
	}

	fmt.Printf("\nTotal tokens: %d\n", result.Usage.TotalTokens)
}
