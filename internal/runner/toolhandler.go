package runner

import (
	"fmt"
	"strings"
	"time"

	"github.com/openai/openai-go"
)

// ToolExecutor defines the interface for executing tools
type ToolExecutor interface {
	Execute(arguments string, contextVariables map[string]any) (any, error)
}

// ToolMap is a map of tool names to tool executors
type ToolMap map[string]ToolExecutor

// TruncateToolCallIDs truncates tool call IDs to MaxToolCallIDLength
// This modifies the message in place for efficiency
func TruncateToolCallIDs(message *openai.ChatCompletionMessage) {
	if len(message.ToolCalls) == 0 {
		return
	}

	for i := range message.ToolCalls {
		if len(message.ToolCalls[i].ID) > MaxToolCallIDLength {
			message.ToolCalls[i].ID = message.ToolCalls[i].ID[:MaxToolCallIDLength]
		}
	}
}

// HandleToolCalls executes tool calls and returns the results
// Returns: tool messages, recorded tool calls, and next agent (if handoff occurred)
func HandleToolCalls(
	toolCalls []openai.ChatCompletionMessageToolCall,
	toolMap ToolMap,
	contextParams map[string]any,
	isHandoffFunc func(result any) (any, bool),
) ([]openai.ChatCompletionMessageParamUnion, []ToolCallResult, any) {
	messages := make([]openai.ChatCompletionMessageParamUnion, 0, len(toolCalls))
	recordedToolCalls := make([]ToolCallResult, 0, len(toolCalls))
	var nextAgent any

	for _, toolCall := range toolCalls {
		toolStart := time.Now()
		toolName := toolCall.Function.Name
		args := toolCall.Function.Arguments

		tool, found := toolMap[toolName]
		var result any
		var err error

		if !found {
			// Provide helpful error with available tools
			available := make([]string, 0, len(toolMap))
			for name := range toolMap {
				available = append(available, name)
			}
			result = fmt.Sprintf("Error: Tool %s not found. Available tools: %v", toolName, available)
			err = fmt.Errorf("tool %s not found (available: %v)", toolName, available)
		} else {
			result, err = tool.Execute(args, contextParams)
			if err != nil {
				// Use strings.Builder for efficient string concatenation
				var sb strings.Builder
				sb.WriteString("Error executing tool ")
				sb.WriteString(toolName)
				sb.WriteString(": ")
				sb.WriteString(err.Error())
				result = sb.String()
			}
		}

		// Record tool call
		recordedToolCalls = append(recordedToolCalls, ToolCallResult{
			ToolName:  toolName,
			Arguments: args,
			Result:    result,
			Error:     err,
			Duration:  time.Since(toolStart),
		})

		// Check for Handoff
		if isHandoffFunc != nil {
			if extractedAgent, ok := isHandoffFunc(result); ok {
				nextAgent = extractedAgent
				// Note: We'll need to get the agent name from outside
				result = "Transferred to agent" // Will be updated by caller
			}
		}

		// Add tool output to history
		toolCallID := toolCall.ID
		if len(toolCallID) > MaxToolCallIDLength {
			toolCallID = toolCallID[:MaxToolCallIDLength]
		}
		resultStr := fmt.Sprintf("%v", result)
		messages = append(messages, openai.ToolMessage(resultStr, toolCallID))
	}

	return messages, recordedToolCalls, nextAgent
}
