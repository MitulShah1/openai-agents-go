package runner

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/openai/openai-go/v3"
)

// mockToolExecutor is a mock implementation of ToolExecutor
type mockToolExecutor struct {
	result any
	err    error
}

func (m *mockToolExecutor) Execute(_ string, _ map[string]any) (any, error) {
	return m.result, m.err
}

func TestTruncateToolCallIDs(t *testing.T) {
	tests := []struct {
		name          string
		message       *openai.ChatCompletionMessage
		expectedIDLen int
	}{
		{
			name: "truncate long ID",
			message: &openai.ChatCompletionMessage{
				ToolCalls: []openai.ChatCompletionMessageToolCallUnion{
					{
						ID: strings.Repeat("a", 50), // 50 chars, should be truncated to 40
						Function: openai.ChatCompletionMessageFunctionToolCallFunction{
							Name: "test_tool",
						},
					},
				},
			},
			expectedIDLen: 40,
		},
		{
			name: "keep short ID",
			message: &openai.ChatCompletionMessage{
				ToolCalls: []openai.ChatCompletionMessageToolCallUnion{
					{
						ID: "short_id",
						Function: openai.ChatCompletionMessageFunctionToolCallFunction{
							Name: "test_tool",
						},
					},
				},
			},
			expectedIDLen: 8, // length of "short_id"
		},
		{
			name: "no tool calls",
			message: &openai.ChatCompletionMessage{
				ToolCalls: []openai.ChatCompletionMessageToolCallUnion{},
			},
			expectedIDLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			TruncateToolCallIDs(tt.message)

			if len(tt.message.ToolCalls) > 0 {
				actualLen := len(tt.message.ToolCalls[0].ID)
				if actualLen != tt.expectedIDLen {
					t.Errorf("TruncateToolCallIDs() ID length = %d, want %d", actualLen, tt.expectedIDLen)
				}
			}
		})
	}
}

func TestHandleToolCalls(t *testing.T) {
	tests := []struct {
		name            string
		toolCalls       []openai.ChatCompletionMessageToolCallUnion
		toolMap         ToolMap
		contextParams   map[string]any
		isHandoffFunc   func(result any) (any, bool)
		wantMsgCount    int
		wantResultCount int
		wantNextAgent   any
	}{
		{
			name: "successful tool execution",
			toolCalls: []openai.ChatCompletionMessageToolCallUnion{
				{
					ID: "call_1",
					Function: openai.ChatCompletionMessageFunctionToolCallFunction{
						Name:      "test_tool",
						Arguments: `{"key": "value"}`,
					},
				},
			},
			toolMap: ToolMap{
				"test_tool": &mockToolExecutor{
					result: "success",
					err:    nil,
				},
			},
			contextParams:   map[string]any{},
			isHandoffFunc:   nil,
			wantMsgCount:    1,
			wantResultCount: 1,
			wantNextAgent:   nil,
		},
		{
			name: "tool not found",
			toolCalls: []openai.ChatCompletionMessageToolCallUnion{
				{
					ID: "call_1",
					Function: openai.ChatCompletionMessageFunctionToolCallFunction{
						Name:      "unknown_tool",
						Arguments: `{}`,
					},
				},
			},
			toolMap: ToolMap{
				"test_tool": &mockToolExecutor{
					result: "success",
					err:    nil,
				},
			},
			contextParams:   map[string]any{},
			isHandoffFunc:   nil,
			wantMsgCount:    1,
			wantResultCount: 1,
			wantNextAgent:   nil,
		},
		{
			name: "tool execution error",
			toolCalls: []openai.ChatCompletionMessageToolCallUnion{
				{
					ID: "call_1",
					Function: openai.ChatCompletionMessageFunctionToolCallFunction{
						Name:      "test_tool",
						Arguments: `{}`,
					},
				},
			},
			toolMap: ToolMap{
				"test_tool": &mockToolExecutor{
					result: nil,
					err:    errors.New("execution failed"),
				},
			},
			contextParams:   map[string]any{},
			isHandoffFunc:   nil,
			wantMsgCount:    1,
			wantResultCount: 1,
			wantNextAgent:   nil,
		},
		{
			name: "agent handoff",
			toolCalls: []openai.ChatCompletionMessageToolCallUnion{
				{
					ID: "call_1",
					Function: openai.ChatCompletionMessageFunctionToolCallFunction{
						Name:      "test_tool",
						Arguments: `{}`,
					},
				},
			},
			toolMap: ToolMap{
				"test_tool": &mockToolExecutor{
					result: "agent_result",
					err:    nil,
				},
			},
			contextParams: map[string]any{},
			isHandoffFunc: func(result any) (any, bool) {
				if result == "agent_result" {
					return "next_agent", true
				}
				return nil, false
			},
			wantMsgCount:    1,
			wantResultCount: 1,
			wantNextAgent:   "next_agent",
		},
		{
			name: "multiple tool calls",
			toolCalls: []openai.ChatCompletionMessageToolCallUnion{
				{
					ID: "call_1",
					Function: openai.ChatCompletionMessageFunctionToolCallFunction{
						Name:      "tool_1",
						Arguments: `{}`,
					},
				},
				{
					ID: "call_2",
					Function: openai.ChatCompletionMessageFunctionToolCallFunction{
						Name:      "tool_2",
						Arguments: `{}`,
					},
				},
			},
			toolMap: ToolMap{
				"tool_1": &mockToolExecutor{result: "result_1", err: nil},
				"tool_2": &mockToolExecutor{result: "result_2", err: nil},
			},
			contextParams:   map[string]any{},
			isHandoffFunc:   nil,
			wantMsgCount:    2,
			wantResultCount: 2,
			wantNextAgent:   nil,
		},
		{
			name: "long tool call ID truncation",
			toolCalls: []openai.ChatCompletionMessageToolCallUnion{
				{
					ID: strings.Repeat("x", 50), // 50 chars, should be truncated
					Function: openai.ChatCompletionMessageFunctionToolCallFunction{
						Name:      "test_tool",
						Arguments: `{}`,
					},
				},
			},
			toolMap: ToolMap{
				"test_tool": &mockToolExecutor{
					result: "success",
					err:    nil,
				},
			},
			contextParams:   map[string]any{},
			isHandoffFunc:   nil,
			wantMsgCount:    1,
			wantResultCount: 1,
			wantNextAgent:   nil,
		},
		{
			name: "parallel execution",
			toolCalls: []openai.ChatCompletionMessageToolCallUnion{
				{
					ID: "call_1",
					Function: openai.ChatCompletionMessageFunctionToolCallFunction{
						Name:      "slow_tool",
						Arguments: `{}`,
					},
				},
				{
					ID: "call_2",
					Function: openai.ChatCompletionMessageFunctionToolCallFunction{
						Name:      "slow_tool",
						Arguments: `{}`,
					},
				},
			},
			toolMap: ToolMap{
				"slow_tool": &mockToolExecutor{
					result: "done",
					err:    nil,
				},
			},
			contextParams:   map[string]any{},
			isHandoffFunc:   nil,
			wantMsgCount:    2,
			wantResultCount: 2,
			wantNextAgent:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			messages, results, nextAgent := HandleToolCalls(
				context.Background(),
				tt.toolCalls,
				tt.toolMap,
				tt.contextParams,
				tt.isHandoffFunc,
				false,
				0, // maxConcurrency
			)

			if len(messages) != tt.wantMsgCount {
				t.Errorf("HandleToolCalls() message count = %d, want %d", len(messages), tt.wantMsgCount)
			}

			if len(results) != tt.wantResultCount {
				t.Errorf("HandleToolCalls() result count = %d, want %d", len(results), tt.wantResultCount)
			}

			if nextAgent != tt.wantNextAgent {
				t.Errorf("HandleToolCalls() nextAgent = %v, want %v", nextAgent, tt.wantNextAgent)
			}

			// Verify tool call results have proper structure
			for _, result := range results {
				if result.ToolName == "" {
					t.Error("HandleToolCalls() tool call result missing ToolName")
				}
				if result.Duration == 0 {
					t.Error("HandleToolCalls() tool call result missing Duration")
				}
			}
		})
	}
}

func TestHandleToolCalls_ErrorMessage(t *testing.T) {
	toolCalls := []openai.ChatCompletionMessageToolCallUnion{
		{
			ID: "call_1",
			Function: openai.ChatCompletionMessageFunctionToolCallFunction{
				Name:      "test_tool",
				Arguments: `{}`,
			},
		},
	}

	toolMap := ToolMap{
		"test_tool": &mockToolExecutor{
			result: nil,
			err:    errors.New("test error"),
		},
	}

	messages, results, _ := HandleToolCalls(context.Background(), toolCalls, toolMap, map[string]any{}, nil, false, 0)

	// Check that error message is properly formatted
	if len(results) != 1 {
		t.Fatal("Expected 1 result")
	}

	resultStr, ok := results[0].Result.(string)
	if !ok {
		t.Fatal("Expected result to be a string")
	}

	if !strings.Contains(resultStr, "Error executing tool test_tool") {
		t.Errorf("Expected error message to contain tool name, got: %s", resultStr)
	}

	if !strings.Contains(resultStr, "test error") {
		t.Errorf("Expected error message to contain original error, got: %s", resultStr)
	}

	// Verify message was created
	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}
}
