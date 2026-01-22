package handoff

import (
	"context"
	"fmt"
	"strings"

	"github.com/openai/openai-go/v3"

	agents "github.com/MitulShah1/openai-agents-go"
	"github.com/MitulShah1/openai-agents-go/tools"
)

// ToTool converts the Handoff into an agents.Tool that can be registered with an agent.
// This is the bridge between declarative handoff configuration and executable tools.
//
// The tool's callback function will:
//  1. Check if the handoff is enabled (via predicate)
//  2. Apply input filtering if configured
//  3. Return the target agent to trigger the handoff
//
// Example:
//
//	handoffTool := handoff.New(supportAgent).ToTool()
//	triageAgent.Tools = []tools.Tool{handoffTool}
func (h *Handoff) ToTool() tools.Tool {
	toolName := h.getToolName()
	description := h.getDescription()

	// Create the tool with a callback that executes the handoff
	tool := tools.Tool{
		Name:          toolName,
		Description:   description,
		IsHandoffTool: true, // Mark as handoff tool
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"summary": map[string]any{
					"type":        "string",
					"description": "A brief summary of the conversation so far to provide context to the next agent",
				},
			},
			"required": []string{},
		},
		Callback: h.createCallback(),
	}

	return tool
}

// getToolName returns the tool name, generating one if not explicitly set.
func (h *Handoff) getToolName() string {
	if h.toolName != "" {
		return h.toolName
	}
	return h.generateToolName()
}

// getDescription returns the description, generating one if not explicitly set.
func (h *Handoff) getDescription() string {
	if h.description != "" {
		return h.description
	}
	return h.generateDescription()
}

// generateToolName creates a tool name from the agent name.
// Format: "transfer_to_<agent_name_lowercase>"
func (h *Handoff) generateToolName() string {
	agentName := h.agent.Name
	// Convert to lowercase and replace spaces with underscores
	normalized := strings.ToLower(agentName)
	normalized = strings.ReplaceAll(normalized, " ", "_")
	return fmt.Sprintf("transfer_to_%s", normalized)
}

// generateDescription creates a description from the agent name and instructions.
func (h *Handoff) generateDescription() string {
	agentName := h.agent.Name

	// Get agent instructions for context
	instructions := ""
	if instrStr, ok := h.agent.Instructions.(string); ok {
		instructions = instrStr
		// Truncate if too long
		if len(instructions) > 100 {
			instructions = instructions[:97] + "..."
		}
	}

	if instructions != "" {
		return fmt.Sprintf("Transfer the conversation to %s. %s", agentName, instructions)
	}
	return fmt.Sprintf("Transfer the conversation to %s", agentName)
}

// createCallback creates the tool callback function that executes the handoff.
func (h *Handoff) createCallback() func(args map[string]any, ctx agents.ContextVariables) (any, error) {
	return func(args map[string]any, contextVars agents.ContextVariables) (any, error) {
		// Create context for the handoff execution
		execCtx := context.Background()

		// Check if handoff is enabled via predicate
		if h.isEnabled != nil {
			enabled, err := h.isEnabled(execCtx, h.agent, contextVars)
			if err != nil {
				return nil, fmt.Errorf("error checking handoff enablement: %w", err)
			}
			if !enabled {
				return nil, fmt.Errorf("handoff to %s is currently disabled", h.agent.Name)
			}
		}

		// Create input data for filtering
		inputData := InputData{
			Agent:       h.agent,
			NewItems:    []openai.ChatCompletionMessageParamUnion{}, // Will be populated by runner
			ContextVars: contextVars,
		}

		// Apply input filter if configured
		if h.inputFilter != nil {
			_, err := h.inputFilter(execCtx, inputData)
			if err != nil {
				return nil, fmt.Errorf("input filter failed: %w", err)
			}
		}

		// Extract summary from args if provided
		summary := ""
		if summaryVal, ok := args["summary"]; ok {
			if summaryStr, ok := summaryVal.(string); ok {
				summary = summaryStr
			}
		}

		// Store summary in context vars for the next agent
		if summary != "" {
			contextVars["_handoff_summary"] = summary
		}

		// Return the target agent to signal handoff
		return h.agent, nil
	}
}
