// Package main demonstrates the tool approval workflow.
//
// This example shows three patterns:
//  1. Static approval (NeedsApproval = true) - always requires approval
//  2. Dynamic approval (ApprovalFunc) - conditionally requires approval
//  3. Inline ApprovalHandler - auto-approves/rejects without pause
//
// Since the pause/resume flow cannot call the real OpenAI API in a demo,
// this example focuses on demonstrating the API surface and approval logic.
package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/openai/openai-go/v3"

	agents "github.com/MitulShah1/openai-agents-go"
	"github.com/MitulShah1/openai-agents-go/tools"
)

func main() {
	fmt.Println("=== Tool Approvals Demo ===")
	fmt.Println()

	demoPauseResumeWorkflow()
	fmt.Println()
	demoDynamicApproval()
	fmt.Println()
	demoInlineHandler()
}

// demoPauseResumeWorkflow shows how the pause/resume pattern works.
// When no ApprovalHandler is provided and a tool needs approval,
// Run() returns a ToolApprovalRequiredError with a RunState snapshot.
func demoPauseResumeWorkflow() {
	fmt.Println("--- 1. Pause/Resume Workflow ---")

	deleteTool := tools.New(
		"delete_database",
		"Delete a database by name",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"db_name": map[string]any{
					"type":        "string",
					"description": "Name of the database to delete",
				},
			},
			"required": []string{"db_name"},
		},
		func(args map[string]any, _ tools.ContextVariables) (any, error) {
			return fmt.Sprintf("Database '%s' deleted", args["db_name"]), nil
		},
	)
	deleteTool.NeedsApproval = true

	agent := agents.NewAgent("DBA Agent")
	agent.Instructions = "You help manage databases. Use tools when asked."
	agent.Tools = []tools.Tool{deleteTool}

	fmt.Printf("  Tool '%s' has NeedsApproval=%v\n", deleteTool.Name, deleteTool.NeedsApproval)
	fmt.Println()

	// Simulate what happens when the model calls the tool:
	// The runner would return a ToolApprovalRequiredError.
	// Here we show how to handle it.

	approvalErr := &agents.ToolApprovalRequiredError{
		Requests: []tools.ApprovalRequest{
			{
				ToolName: "delete_database",
				CallID:   "call_abc123",
				Args:     map[string]any{"db_name": "production"},
				Context:  tools.ContextVariables{"user": "admin"},
			},
		},
		State: &agents.RunState{
			Agent:     agent,
			TurnCount: 1,
			Config:    agents.DefaultRunConfig(),
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.UserMessage("Delete the production database"),
			},
		},
	}

	fmt.Printf("  ToolApprovalRequiredError: %s\n", approvalErr.Error())
	fmt.Printf("  Pending requests: %d\n", len(approvalErr.Requests))

	for _, req := range approvalErr.Requests {
		fmt.Printf("    Tool: %s, CallID: %s, Args: %v\n",
			req.ToolName, req.CallID, req.Args)
	}

	// In a real app, you would present this to the user and collect decisions:
	//
	//   approvals := map[string]*tools.ApprovalResponse{
	//       "call_abc123": {Approved: true},
	//   }
	//   result, err := runner.Resume(ctx, approvalErr.State, approvals)
	//
	// Or to reject:
	//
	//   approvals := map[string]*tools.ApprovalResponse{
	//       "call_abc123": {Approved: false, Reason: "not authorized"},
	//   }
	//   result, err := runner.Resume(ctx, approvalErr.State, approvals)

	fmt.Println()
	fmt.Println("  Usage pattern:")
	fmt.Println("    result, err := runner.Run(ctx, agent, messages)")
	fmt.Println("    var approvalErr *agents.ToolApprovalRequiredError")
	fmt.Println("    if errors.As(err, &approvalErr) {")
	fmt.Println("        // collect decisions from user...")
	fmt.Println("        result, err = runner.Resume(ctx, approvalErr.State, approvals)")
	fmt.Println("    }")
}

// demoDynamicApproval shows how ApprovalFunc enables conditional approval.
func demoDynamicApproval() {
	fmt.Println("--- 2. Dynamic Approval (ApprovalFunc) ---")

	transferTool := tools.New(
		"transfer_funds",
		"Transfer funds between accounts",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"amount": map[string]any{
					"type":        "number",
					"description": "Amount to transfer",
				},
				"to": map[string]any{
					"type":        "string",
					"description": "Destination account",
				},
			},
			"required": []string{"amount", "to"},
		},
		func(args map[string]any, _ tools.ContextVariables) (any, error) {
			return fmt.Sprintf("Transferred $%.2f to %s", args["amount"], args["to"]), nil
		},
	)

	// Only require approval for large transfers
	transferTool.ApprovalFunc = func(args map[string]any, _ string, _ tools.ContextVariables) (bool, error) {
		amount, ok := args["amount"].(float64)
		if !ok {
			return false, nil
		}
		return amount > 1000, nil
	}

	// Test with different amounts
	testCases := []map[string]any{
		{"amount": float64(50), "to": "savings"},
		{"amount": float64(5000), "to": "external"},
	}

	for _, args := range testCases {
		needs, _ := transferTool.RequiresApproval(args, "call_1", nil)
		fmt.Printf("  Transfer $%.0f to %s → needs approval: %v\n",
			args["amount"], args["to"], needs)
	}
}

// demoInlineHandler shows the WithApprovalHandler pattern.
func demoInlineHandler() {
	fmt.Println("--- 3. Inline ApprovalHandler ---")

	execTool := tools.New(
		"execute_command",
		"Execute a shell command",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"command": map[string]any{
					"type":        "string",
					"description": "The command to execute",
				},
			},
			"required": []string{"command"},
		},
		func(args map[string]any, _ tools.ContextVariables) (any, error) {
			return fmt.Sprintf("Executed: %s", args["command"]), nil
		},
	)
	execTool.NeedsApproval = true

	// Create a handler that auto-approves safe commands, rejects dangerous ones
	handler := func(req tools.ApprovalRequest) (*tools.ApprovalResponse, error) {
		cmd, _ := req.Args["command"].(string)
		fmt.Printf("  Handler received: tool=%s, command=%q\n", req.ToolName, cmd)

		dangerous := []string{"rm", "sudo", "mkfs", "dd"}
		for _, d := range dangerous {
			if strings.Contains(cmd, d) {
				fmt.Printf("    → REJECTED (contains '%s')\n", d)
				return &tools.ApprovalResponse{
					Approved: false,
					Reason:   fmt.Sprintf("command contains dangerous keyword: %s", d),
				}, nil
			}
		}

		fmt.Println("    → APPROVED")
		return &tools.ApprovalResponse{Approved: true}, nil
	}

	// Simulate approval checks
	testCommands := []string{"ls -la", "sudo rm -rf /", "cat file.txt"}
	for _, cmd := range testCommands {
		req := tools.ApprovalRequest{
			ToolName: "execute_command",
			CallID:   "call_1",
			Args:     map[string]any{"command": cmd},
		}
		resp, _ := handler(req)
		_ = resp
	}

	fmt.Println()
	fmt.Println("  Usage:")
	fmt.Println("    runner.Run(ctx, agent, messages,")
	fmt.Println("        agents.WithApprovalHandler(handler),")
	fmt.Println("    )")

	// Verify errors.As works
	var err error = &agents.ToolApprovalRequiredError{
		Requests: []tools.ApprovalRequest{{ToolName: "test"}},
	}
	var approvalErr *agents.ToolApprovalRequiredError
	if errors.As(err, &approvalErr) {
		fmt.Println()
		fmt.Printf("  errors.As works correctly: %s\n", approvalErr.Error())
	}

	_ = context.Background()
}
