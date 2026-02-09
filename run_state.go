package agents

import (
	"github.com/openai/openai-go/v3"

	"github.com/MitulShah1/openai-agents-go/tools"
)

// RunState captures the execution state of a runner at the point of interruption.
// It enables pause/resume workflows, particularly for tool approval flows where
// human-in-the-loop decisions are required before continuing execution.
//
// RunState is returned as part of ToolApprovalRequiredError when a tool call
// requires approval and no ApprovalHandler is configured.
//
// Example (pause/resume):
//
//	result, err := runner.Run(ctx, agent, messages)
//	var approvalErr *agents.ToolApprovalRequiredError
//	if errors.As(err, &approvalErr) {
//	    // Present approval requests to user...
//	    approvals := map[string]*tools.ApprovalResponse{
//	        approvalErr.Requests[0].CallID: {Approved: true},
//	    }
//	    result, err = runner.Resume(ctx, approvalErr.State, approvals)
//	}
type RunState struct {
	// Agent is the agent that was executing when the interruption occurred.
	Agent *Agent

	// Messages contains the conversation history up to and including
	// the assistant message with pending tool calls.
	Messages []openai.ChatCompletionMessageParamUnion

	// TurnCount is the number of completed turns before interruption.
	// Used to enforce MaxTurns across pause/resume boundaries.
	TurnCount int

	// PendingToolCalls contains the tool calls awaiting approval decisions.
	PendingToolCalls []openai.ChatCompletionMessageToolCallUnion

	// ContextVariables holds the context variables at the point of interruption.
	ContextVariables ContextVariables

	// Config is the run configuration that was in effect.
	Config *RunConfig
}

// ToolApprovalRequiredError is returned when one or more tool calls require
// human approval before execution and no ApprovalHandler is configured.
//
// The caller should inspect the Requests to determine which tools need approval,
// make approval/rejection decisions, and then call Runner.Resume with the
// embedded State and approval decisions.
type ToolApprovalRequiredError struct {
	// Requests contains the approval requests for each tool call that needs approval.
	Requests []tools.ApprovalRequest

	// State captures the runner state at the point of interruption.
	// Pass this to Runner.Resume along with approval decisions to continue execution.
	State *RunState
}

func (e *ToolApprovalRequiredError) Error() string {
	if len(e.Requests) == 1 {
		return "tool approval required: " + e.Requests[0].ToolName
	}
	names := make([]string, len(e.Requests))
	for i, r := range e.Requests {
		names[i] = r.ToolName
	}
	return "tool approval required: " + joinNames(names)
}

func joinNames(names []string) string {
	switch len(names) {
	case 0:
		return ""
	case 1:
		return names[0]
	default:
		result := names[0]
		for _, n := range names[1:] {
			result += ", " + n
		}
		return result
	}
}
