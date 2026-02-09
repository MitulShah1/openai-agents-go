package agents

import (
	"context"
	"errors"
	"testing"

	"github.com/openai/openai-go/v3"

	"github.com/MitulShah1/openai-agents-go/tools"
)

func TestToolApprovalRequiredError_SingleTool(t *testing.T) {
	err := &ToolApprovalRequiredError{
		Requests: []tools.ApprovalRequest{
			{ToolName: "delete_file", CallID: "call_1"},
		},
	}

	expected := "tool approval required: delete_file"
	if err.Error() != expected {
		t.Errorf("got %q, want %q", err.Error(), expected)
	}
}

func TestToolApprovalRequiredError_MultipleTools(t *testing.T) {
	err := &ToolApprovalRequiredError{
		Requests: []tools.ApprovalRequest{
			{ToolName: "delete_file", CallID: "call_1"},
			{ToolName: "send_email", CallID: "call_2"},
		},
	}

	expected := "tool approval required: delete_file, send_email"
	if err.Error() != expected {
		t.Errorf("got %q, want %q", err.Error(), expected)
	}
}

func TestToolApprovalRequiredError_AsError(t *testing.T) {
	originalErr := &ToolApprovalRequiredError{
		Requests: []tools.ApprovalRequest{
			{ToolName: "rm_rf", CallID: "call_1"},
		},
		State: &RunState{
			Agent:     NewAgent("test"),
			TurnCount: 3,
		},
	}

	var err error = originalErr
	var approvalErr *ToolApprovalRequiredError
	if !errors.As(err, &approvalErr) {
		t.Fatal("expected errors.As to succeed")
	}

	if approvalErr.State.TurnCount != 3 {
		t.Errorf("got TurnCount=%d, want 3", approvalErr.State.TurnCount)
	}
}

func TestRunState_Fields(t *testing.T) {
	agent := NewAgent("test-agent")
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("hello"),
	}
	config := DefaultRunConfig()
	ctxVars := ContextVariables{"key": "value"}

	state := &RunState{
		Agent:            agent,
		Messages:         messages,
		TurnCount:        5,
		ContextVariables: ctxVars,
		Config:           config,
		PendingToolCalls: []openai.ChatCompletionMessageToolCallUnion{
			{
				ID:   "call_abc",
				Type: "function",
				Function: openai.ChatCompletionMessageFunctionToolCallFunction{
					Name:      "test_tool",
					Arguments: `{"key":"value"}`,
				},
			},
		},
	}

	if state.Agent.Name != "test-agent" {
		t.Errorf("got Agent.Name=%q, want %q", state.Agent.Name, "test-agent")
	}
	if state.TurnCount != 5 {
		t.Errorf("got TurnCount=%d, want 5", state.TurnCount)
	}
	if len(state.PendingToolCalls) != 1 {
		t.Fatalf("got %d pending tool calls, want 1", len(state.PendingToolCalls))
	}
	if state.PendingToolCalls[0].Function.Name != "test_tool" {
		t.Errorf("got tool name %q, want %q", state.PendingToolCalls[0].Function.Name, "test_tool")
	}
}

func TestCheckToolApprovals_NoApprovalNeeded(t *testing.T) {
	r := NewRunner(&openai.Client{})
	agent := NewAgent("test")
	agent.Tools = []tools.Tool{
		tools.New("safe_tool", "safe", nil, func(_ map[string]any, _ tools.ContextVariables) (any, error) {
			return "ok", nil
		}),
	}

	toolCalls := []openai.ChatCompletionMessageToolCallUnion{
		{
			ID:   "call_1",
			Type: "function",
			Function: openai.ChatCompletionMessageFunctionToolCallFunction{
				Name:      "safe_tool",
				Arguments: "{}",
			},
		},
	}

	err := r.checkToolApprovals(toolCalls, agent, nil, nil, nil, 0, DefaultRunConfig())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestCheckToolApprovals_NoHandler_ReturnsError(t *testing.T) {
	r := NewRunner(&openai.Client{})
	agent := NewAgent("test")
	dangerousTool := tools.New("delete_db", "delete database", nil,
		func(_ map[string]any, _ tools.ContextVariables) (any, error) {
			return "deleted", nil
		})
	dangerousTool.NeedsApproval = true
	agent.Tools = []tools.Tool{dangerousTool}

	toolCalls := []openai.ChatCompletionMessageToolCallUnion{
		{
			ID:   "call_1",
			Type: "function",
			Function: openai.ChatCompletionMessageFunctionToolCallFunction{
				Name:      "delete_db",
				Arguments: `{"target":"prod"}`,
			},
		},
	}

	err := r.checkToolApprovals(toolCalls, agent, nil, nil, nil, 1, DefaultRunConfig())
	if err == nil {
		t.Fatal("expected ToolApprovalRequiredError, got nil")
	}

	var approvalErr *ToolApprovalRequiredError
	if !errors.As(err, &approvalErr) {
		t.Fatalf("expected ToolApprovalRequiredError, got %T", err)
	}

	if len(approvalErr.Requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(approvalErr.Requests))
	}
	if approvalErr.Requests[0].ToolName != "delete_db" {
		t.Errorf("expected tool name 'delete_db', got %q", approvalErr.Requests[0].ToolName)
	}
	if approvalErr.Requests[0].Args["target"] != "prod" {
		t.Errorf("expected args[target]='prod', got %v", approvalErr.Requests[0].Args["target"])
	}
	if approvalErr.State == nil {
		t.Fatal("expected non-nil State")
	}
	if approvalErr.State.TurnCount != 1 {
		t.Errorf("expected TurnCount=1, got %d", approvalErr.State.TurnCount)
	}
}

func TestCheckToolApprovals_HandlerApproves(t *testing.T) {
	r := NewRunner(&openai.Client{})
	agent := NewAgent("test")
	tool := tools.New("send_email", "send email", nil,
		func(_ map[string]any, _ tools.ContextVariables) (any, error) {
			return "sent", nil
		})
	tool.NeedsApproval = true
	agent.Tools = []tools.Tool{tool}

	toolCalls := []openai.ChatCompletionMessageToolCallUnion{
		{
			ID:   "call_1",
			Type: "function",
			Function: openai.ChatCompletionMessageFunctionToolCallFunction{
				Name:      "send_email",
				Arguments: `{}`,
			},
		},
	}

	handler := func(_ tools.ApprovalRequest) (*tools.ApprovalResponse, error) {
		return &tools.ApprovalResponse{Approved: true}, nil
	}

	err := r.checkToolApprovals(toolCalls, agent, nil, handler, nil, 0, DefaultRunConfig())
	if err != nil {
		t.Fatalf("expected nil error when handler approves, got: %v", err)
	}
}

func TestCheckToolApprovals_HandlerRejects(t *testing.T) {
	r := NewRunner(&openai.Client{})
	agent := NewAgent("test")
	tool := tools.New("delete_user", "delete user", nil,
		func(_ map[string]any, _ tools.ContextVariables) (any, error) {
			return "deleted", nil
		})
	tool.NeedsApproval = true
	agent.Tools = []tools.Tool{tool}

	toolCalls := []openai.ChatCompletionMessageToolCallUnion{
		{
			ID:   "call_1",
			Type: "function",
			Function: openai.ChatCompletionMessageFunctionToolCallFunction{
				Name:      "delete_user",
				Arguments: `{}`,
			},
		},
	}

	handler := func(_ tools.ApprovalRequest) (*tools.ApprovalResponse, error) {
		return &tools.ApprovalResponse{Approved: false, Reason: "too dangerous"}, nil
	}

	err := r.checkToolApprovals(toolCalls, agent, nil, handler, nil, 0, DefaultRunConfig())
	if err == nil {
		t.Fatal("expected error when handler rejects")
	}

	var approvalErr *ToolApprovalRequiredError
	if !errors.As(err, &approvalErr) {
		t.Fatalf("expected ToolApprovalRequiredError, got %T: %v", err, err)
	}
}

func TestCheckToolApprovals_DynamicApprovalFunc(t *testing.T) {
	r := NewRunner(&openai.Client{})
	agent := NewAgent("test")
	tool := tools.New("transfer_funds", "transfer", nil,
		func(_ map[string]any, _ tools.ContextVariables) (any, error) {
			return "transferred", nil
		})
	tool.ApprovalFunc = func(args map[string]any, _ string, _ tools.ContextVariables) (bool, error) {
		amount, ok := args["amount"].(float64)
		if !ok {
			return false, nil
		}
		return amount > 1000, nil
	}
	agent.Tools = []tools.Tool{tool}

	t.Run("below threshold no approval needed", func(t *testing.T) {
		toolCalls := []openai.ChatCompletionMessageToolCallUnion{
			{
				ID:   "call_1",
				Type: "function",
				Function: openai.ChatCompletionMessageFunctionToolCallFunction{
					Name:      "transfer_funds",
					Arguments: `{"amount":500}`,
				},
			},
		}

		err := r.checkToolApprovals(toolCalls, agent, nil, nil, nil, 0, DefaultRunConfig())
		if err != nil {
			t.Fatalf("expected no error for small amount, got: %v", err)
		}
	})

	t.Run("above threshold approval needed", func(t *testing.T) {
		toolCalls := []openai.ChatCompletionMessageToolCallUnion{
			{
				ID:   "call_2",
				Type: "function",
				Function: openai.ChatCompletionMessageFunctionToolCallFunction{
					Name:      "transfer_funds",
					Arguments: `{"amount":5000}`,
				},
			},
		}

		err := r.checkToolApprovals(toolCalls, agent, nil, nil, nil, 0, DefaultRunConfig())
		if err == nil {
			t.Fatal("expected approval error for large amount")
		}

		var approvalErr *ToolApprovalRequiredError
		if !errors.As(err, &approvalErr) {
			t.Fatalf("expected ToolApprovalRequiredError, got %T", err)
		}
	})
}

func TestCheckToolApprovals_MultipleToolsMixedApproval(t *testing.T) {
	r := NewRunner(&openai.Client{})
	agent := NewAgent("test")

	safeTool := tools.New("read_file", "read", nil,
		func(_ map[string]any, _ tools.ContextVariables) (any, error) {
			return "content", nil
		})
	dangerousTool := tools.New("write_file", "write", nil,
		func(_ map[string]any, _ tools.ContextVariables) (any, error) {
			return "written", nil
		})
	dangerousTool.NeedsApproval = true

	agent.Tools = []tools.Tool{safeTool, dangerousTool}

	toolCalls := []openai.ChatCompletionMessageToolCallUnion{
		{
			ID:   "call_1",
			Type: "function",
			Function: openai.ChatCompletionMessageFunctionToolCallFunction{
				Name:      "read_file",
				Arguments: `{}`,
			},
		},
		{
			ID:   "call_2",
			Type: "function",
			Function: openai.ChatCompletionMessageFunctionToolCallFunction{
				Name:      "write_file",
				Arguments: `{}`,
			},
		},
	}

	err := r.checkToolApprovals(toolCalls, agent, nil, nil, nil, 0, DefaultRunConfig())
	if err == nil {
		t.Fatal("expected approval error when batch contains tool needing approval")
	}

	var approvalErr *ToolApprovalRequiredError
	if !errors.As(err, &approvalErr) {
		t.Fatalf("expected ToolApprovalRequiredError, got %T", err)
	}
	if len(approvalErr.Requests) != 1 {
		t.Errorf("expected 1 request (only the dangerous tool), got %d", len(approvalErr.Requests))
	}
	if approvalErr.Requests[0].ToolName != "write_file" {
		t.Errorf("expected 'write_file', got %q", approvalErr.Requests[0].ToolName)
	}
}

func TestCheckToolApprovals_UnknownToolSkipped(t *testing.T) {
	r := NewRunner(&openai.Client{})
	agent := NewAgent("test")
	agent.Tools = []tools.Tool{}

	toolCalls := []openai.ChatCompletionMessageToolCallUnion{
		{
			ID:   "call_1",
			Type: "function",
			Function: openai.ChatCompletionMessageFunctionToolCallFunction{
				Name:      "nonexistent_tool",
				Arguments: `{}`,
			},
		},
	}

	err := r.checkToolApprovals(toolCalls, agent, nil, nil, nil, 0, DefaultRunConfig())
	if err != nil {
		t.Fatalf("expected nil error for unknown tool, got: %v", err)
	}
}

func TestParseToolArgs(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"empty string", "", 0},
		{"empty object", "{}", 0},
		{"valid args", `{"key":"value","count":42}`, 2},
		{"invalid json", "not json", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseToolArgs(tt.input)
			if len(result) != tt.expected {
				t.Errorf("expected %d keys, got %d", tt.expected, len(result))
			}
		})
	}
}

func TestResumeNilState(t *testing.T) {
	r := NewRunner(&openai.Client{})
	_, err := r.Resume(context.Background(), nil, nil)
	if err == nil {
		t.Fatal("expected error for nil state")
	}
}

func TestJoinNames(t *testing.T) {
	tests := []struct {
		input    []string
		expected string
	}{
		{nil, ""},
		{[]string{}, ""},
		{[]string{"a"}, "a"},
		{[]string{"a", "b"}, "a, b"},
		{[]string{"a", "b", "c"}, "a, b, c"},
	}

	for _, tt := range tests {
		result := joinNames(tt.input)
		if result != tt.expected {
			t.Errorf("joinNames(%v) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestWithApprovalHandler(t *testing.T) {
	called := false
	handler := func(_ tools.ApprovalRequest) (*tools.ApprovalResponse, error) {
		called = true
		return &tools.ApprovalResponse{Approved: true}, nil
	}

	opts := &runOptions{}
	WithApprovalHandler(handler)(opts)

	if opts.approvalHandler == nil {
		t.Fatal("expected approvalHandler to be set")
	}

	_, _ = opts.approvalHandler(tools.ApprovalRequest{ToolName: "test"})
	if !called {
		t.Error("expected handler to be called")
	}
}

func TestCheckToolApprovals_ContextVariablesPassed(t *testing.T) {
	r := NewRunner(&openai.Client{})
	agent := NewAgent("test")
	tool := tools.New("protected_tool", "protected", nil,
		func(_ map[string]any, _ tools.ContextVariables) (any, error) {
			return "ok", nil
		})
	tool.NeedsApproval = true
	agent.Tools = []tools.Tool{tool}

	ctxVars := ContextVariables{"user_role": "admin"}
	toolCalls := []openai.ChatCompletionMessageToolCallUnion{
		{
			ID:   "call_1",
			Type: "function",
			Function: openai.ChatCompletionMessageFunctionToolCallFunction{
				Name:      "protected_tool",
				Arguments: `{}`,
			},
		},
	}

	var receivedCtx tools.ContextVariables
	handler := func(req tools.ApprovalRequest) (*tools.ApprovalResponse, error) {
		receivedCtx = req.Context
		return &tools.ApprovalResponse{Approved: true}, nil
	}

	err := r.checkToolApprovals(toolCalls, agent, ctxVars, handler, nil, 0, DefaultRunConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedCtx["user_role"] != "admin" {
		t.Errorf("expected context variable 'user_role'='admin', got %v", receivedCtx["user_role"])
	}
}

func TestCheckToolApprovals_HandlerError(t *testing.T) {
	r := NewRunner(&openai.Client{})
	agent := NewAgent("test")
	tool := tools.New("risky_tool", "risky", nil,
		func(_ map[string]any, _ tools.ContextVariables) (any, error) {
			return "ok", nil
		})
	tool.NeedsApproval = true
	agent.Tools = []tools.Tool{tool}

	toolCalls := []openai.ChatCompletionMessageToolCallUnion{
		{
			ID:   "call_1",
			Type: "function",
			Function: openai.ChatCompletionMessageFunctionToolCallFunction{
				Name:      "risky_tool",
				Arguments: `{}`,
			},
		},
	}

	handlerErr := errors.New("approval service unavailable")
	handler := func(_ tools.ApprovalRequest) (*tools.ApprovalResponse, error) {
		return nil, handlerErr
	}

	err := r.checkToolApprovals(toolCalls, agent, nil, handler, nil, 0, DefaultRunConfig())
	if err == nil {
		t.Fatal("expected error when handler returns error")
	}
	if !errors.Is(err, handlerErr) {
		t.Errorf("expected wrapped handler error, got: %v", err)
	}
}

func TestCheckToolApprovals_StateCapture(t *testing.T) {
	r := NewRunner(&openai.Client{})
	agent := NewAgent("test-agent")
	tool := tools.New("dangerous_tool", "danger", nil,
		func(_ map[string]any, _ tools.ContextVariables) (any, error) {
			return "ok", nil
		})
	tool.NeedsApproval = true
	agent.Tools = []tools.Tool{tool}

	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("hello"),
		openai.UserMessage("do something dangerous"),
	}
	ctxVars := ContextVariables{"session": "abc123"}
	config := &RunConfig{MaxTurns: 5}

	toolCalls := []openai.ChatCompletionMessageToolCallUnion{
		{
			ID:   "call_xyz",
			Type: "function",
			Function: openai.ChatCompletionMessageFunctionToolCallFunction{
				Name:      "dangerous_tool",
				Arguments: `{"target":"production"}`,
			},
		},
	}

	err := r.checkToolApprovals(toolCalls, agent, ctxVars, nil, messages, 3, config)
	if err == nil {
		t.Fatal("expected ToolApprovalRequiredError")
	}

	var approvalErr *ToolApprovalRequiredError
	if !errors.As(err, &approvalErr) {
		t.Fatalf("expected ToolApprovalRequiredError, got %T", err)
	}

	state := approvalErr.State
	if state.Agent.Name != "test-agent" {
		t.Errorf("state.Agent.Name = %q, want 'test-agent'", state.Agent.Name)
	}
	if state.TurnCount != 3 {
		t.Errorf("state.TurnCount = %d, want 3", state.TurnCount)
	}
	if len(state.Messages) != 2 {
		t.Errorf("state.Messages length = %d, want 2", len(state.Messages))
	}
	if state.ContextVariables["session"] != "abc123" {
		t.Errorf("state.ContextVariables[session] = %v, want 'abc123'", state.ContextVariables["session"])
	}
	if state.Config.MaxTurns != 5 {
		t.Errorf("state.Config.MaxTurns = %d, want 5", state.Config.MaxTurns)
	}
	if len(state.PendingToolCalls) != 1 {
		t.Fatalf("state.PendingToolCalls length = %d, want 1", len(state.PendingToolCalls))
	}
	if state.PendingToolCalls[0].ID != "call_xyz" {
		t.Errorf("state.PendingToolCalls[0].ID = %q, want 'call_xyz'", state.PendingToolCalls[0].ID)
	}
}
