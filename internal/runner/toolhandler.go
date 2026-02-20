package runner

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/openai/openai-go/v3"

	"github.com/MitulShah1/openai-agents-go/tracing"
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
	ctx context.Context,
	toolCalls []openai.ChatCompletionMessageToolCallUnion,
	toolMap ToolMap,
	contextParams map[string]any,
	isHandoffFunc func(result any) (any, bool),
	parallel bool,
	maxConcurrency int,
) ([]openai.ChatCompletionMessageParamUnion, []ToolCallResult, any) {
	// Pre-allocate results slice to ensure order preservation
	results := make([]ToolCallResult, len(toolCalls))
	messages := make([]openai.ChatCompletionMessageParamUnion, len(toolCalls))
	var recordedToolCalls []ToolCallResult
	var nextAgent any

	if parallel && len(toolCalls) > 1 {
		// Parallel execution
		var wg sync.WaitGroup
		wg.Add(len(toolCalls))

		// Semaphore for limiting concurrency
		var sem chan struct{}
		if maxConcurrency > 0 {
			sem = make(chan struct{}, maxConcurrency)
		}

		for i, tc := range toolCalls {
			go func(i int, tc openai.ChatCompletionMessageToolCallUnion) {
				defer wg.Done()

				// Acquire semaphore
				if sem != nil {
					sem <- struct{}{}
					defer func() { <-sem }()
				}

				// Shallow copy context params to prevent data race
				localParams := make(map[string]any, len(contextParams))
				for k, v := range contextParams {
					localParams[k] = v
				}

				res, msg, _ := executeSingleTool(ctx, tc, toolMap, localParams, isHandoffFunc)

				// Capture results safely
				// Since we pre-allocated slices and have unique 'i', no mutex needed for slice access
				// BUT: Go memory model guarantees? Yes, modifying distinct indices is safe if no resizing.
				// However, 'messages[i] = msg' involves interface assignment which might not be atomic?
				// Actually, slice assignment of distinct elements is safe in Go.
				results[i] = res
				messages[i] = msg
			}(i, tc)
		}

		wg.Wait()
	} else {
		// Sequential execution
		for i, tc := range toolCalls {
			res, msg, next := executeSingleTool(ctx, tc, toolMap, contextParams, isHandoffFunc)
			results[i] = res
			messages[i] = msg
			if next != nil {
				nextAgent = next
			}
		}
	}

	// Re-construct recordedToolCalls and check for handoffs from results
	for _, res := range results {
		recordedToolCalls = append(recordedToolCalls, res)
		// Check for handoff if we didn't capture it in sequential loop
		if parallel && isHandoffFunc != nil {
			if extracted, ok := isHandoffFunc(res.Result); ok {
				nextAgent = extracted
			}
		}
	}

	return messages, recordedToolCalls, nextAgent
}

// executeSingleTool helper to isolate execution logic
func executeSingleTool(
	ctx context.Context,
	toolCall openai.ChatCompletionMessageToolCallUnion,
	toolMap ToolMap,
	contextParams map[string]any,
	isHandoffFunc func(result any) (any, bool),
) (ToolCallResult, openai.ChatCompletionMessageParamUnion, any) {
	toolName := toolCall.Function.Name

	toolStart := time.Now()
	args := toolCall.Function.Arguments

	tool, found := toolMap[toolName]
	var result any
	var err error

	if !found {
		available := make([]string, 0, len(toolMap))
		for name := range toolMap {
			available = append(available, name)
		}
		result = fmt.Sprintf("Error: Tool %s not found. Available tools: %v", toolName, available)
		err = fmt.Errorf("tool %s not found (available: %v)", toolName, available)
	} else {
		// Start function span
		// Redaction is handled automatically
		ctx2, sp, _ := tracing.StartFunctionSpan(ctx, toolName, tracing.WithArguments(args))
		result, err = tool.Execute(args, contextParams)
		if err != nil {
			sp.RecordError(err)
		}

		traceOutput := result
		if isHandoffFunc != nil {
			if _, ok := isHandoffFunc(result); ok {
				traceOutput = "Transferred to agent"
			}
		}

		sp.SetAttributes(map[string]any{
			"output": traceOutput,
		})
		sp.End(ctx2)

		if err != nil {
			var sb strings.Builder
			sb.WriteString("Error executing tool ")
			sb.WriteString(toolName)
			sb.WriteString(": ")
			sb.WriteString(err.Error())
			result = sb.String()
		}
	}

	// Record tool call
	record := ToolCallResult{
		ToolName:  toolName,
		Arguments: args,
		Result:    result,
		Error:     err,
		Duration:  time.Since(toolStart),
	}

	var nextAgent any
	if isHandoffFunc != nil {
		if extractedAgent, ok := isHandoffFunc(result); ok {
			nextAgent = extractedAgent
			result = "Transferred to agent"
		}
	}

	toolCallID := toolCall.ID
	if len(toolCallID) > MaxToolCallIDLength {
		toolCallID = toolCallID[:MaxToolCallIDLength]
	}
	resultStr := fmt.Sprintf("%v", result)
	msg := openai.ToolMessage(resultStr, toolCallID)

	return record, msg, nextAgent
}
