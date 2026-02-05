// Package main demonstrates advanced handoff patterns including input filtering,
// history nesting, and dynamic enablement.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"

	agents "github.com/MitulShah1/openai-agents-go"
	"github.com/MitulShah1/openai-agents-go/handoff"
	"github.com/MitulShah1/openai-agents-go/tools"
)

// Agent definitions
var (
	triageAgent  = agents.NewAgent("Triage Agent")
	salesAgent   = agents.NewAgent("Sales Agent")
	supportAgent = agents.NewAgent("Support Agent")
)

func main() {
	client := openai.NewClient(
		option.WithAPIKey(os.Getenv("OPENAI_API_KEY")),
	)

	configureAgents()

	// Note: NewRunner expects a pointer, but NewClient returns a struct in v3
	runner := agents.NewRunner(&client)

	// Simulate a conversation flow
	// 1. User asks for help -> Triage
	// 2. Triage determines it's a sales question -> Sales (with input filtering)
	// 3. User provides details -> Sales
	// 4. Sales determines it's actually a support issue -> Support (with history nesting)

	fmt.Println("--- Starting Advanced Handoff Demo ---")

	// Initial request
	ctx := context.Background()

	// Create initial messages
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("I'm interested in buying your enterprise plan, but I have a technical question about API limits first."),
	}

	result, err := runner.Run(ctx, triageAgent, messages)
	if err != nil {
		log.Fatal(err)
	}
	printLastMessage(result)

	// Update messages with the result from the first run
	messages = result.Messages
	messages = append(messages, openai.UserMessage("My current system does 5000 requests per second. Can you handle that?"))

	// Follow-up
	// Note: We continue with the agent returned from the last run (which might be Sales or Support)
	result, err = runner.Run(ctx, result.Agent, messages)
	if err != nil {
		log.Fatal(err)
	}
	printLastMessage(result)
}

func configureAgents() {
	// 1. Triage Agent Instructions
	triageAgent.Instructions = "You are a triage agent. Route the user to sales for purchase inquiries or support for technical help."

	// 2. Sales Agent Instructions
	salesAgent.Instructions = "You are a sales agent. Answer questions about pricing and plans. If the user has deep technical questions, transfer them to support."

	// 3. Support Agent Instructions
	supportAgent.Instructions = "You are a technical support agent. Help users with API limits, integration issues, and bugs."

	// --- Define Handoffs ---

	// Triage -> Sales (Standard handoff)
	toSales := handoff.New(salesAgent).ToTool()

	// Triage -> Support (Standard handoff)
	toSupport := handoff.New(supportAgent).ToTool()

	// Sales -> Support (Advanced handoff with Input Filtering and History Nesting)
	// We want to pass context but summarize the sales conversation to save tokens
	toSupportFromSales := handoff.New(
		supportAgent,
		handoff.WithToolName("escalate_technical_issue"),
		handoff.WithDescription("Transfer to support for complex technical questions."),
		handoff.WithHistoryNesting(true), // Summarize history!
		// Add a context variable to indicate source
		handoff.WithInputFilter(func(_ context.Context, data handoff.InputData) (handoff.InputData, error) {
			data.ContextVars["transferred_from"] = "sales"
			data.ContextVars["priority"] = "high" // Potential enterprise customer
			return data, nil
		}),
	).ToTool()

	// Register tools
	triageAgent.Tools = []tools.Tool{toSales, toSupport}
	salesAgent.Tools = []tools.Tool{toSupportFromSales}
	// Support agent is a leaf node, no tools
}

func printLastMessage(result *agents.Result) {
	if len(result.Messages) > 0 {
		// Use the FinalOutput helper from the result which extracts the last message content
		fmt.Printf("[%s]: %s\n", result.Agent.Name, result.FinalOutput)
	}
}
