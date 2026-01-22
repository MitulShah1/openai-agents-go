// Package handoff provides internal execution logic for handoffs.
// This package is used by the runner to process handoffs and should not
// be imported directly by user code.
package handoff

import (
	"context"
	"fmt"

	"github.com/openai/openai-go/v3"

	"github.com/MitulShah1/openai-agents-go/handoff"
)

// Result contains metadata and state from handoff execution.
type Result struct {
	// TargetAgent is the agent to hand off to
	TargetAgent *handoff.Handoff

	// HistoryNested indicates if history was nested/summarized
	HistoryNested bool

	// FilterApplied indicates if an input filter was applied
	FilterApplied bool

	// TransformedMessages are the messages after filtering/nesting
	TransformedMessages []openai.ChatCompletionMessageParamUnion
}

// Execute processes a handoff and returns execution metadata.
// This is called by the runner when a handoff tool is executed.
func Execute(
	ctx context.Context,
	h *handoff.Handoff,
	inputData handoff.InputData,
	currentMessages []openai.ChatCompletionMessageParamUnion,
) (*Result, error) {
	result := &Result{
		TargetAgent:         h,
		HistoryNested:       false,
		FilterApplied:       false,
		TransformedMessages: currentMessages,
	}

	// Apply input filter if configured
	if h.InputFilter() != nil {
		_, err := ApplyInputFilter(ctx, h.InputFilter(), inputData)
		if err != nil {
			return nil, fmt.Errorf("failed to apply input filter: %w", err)
		}
		result.FilterApplied = true
	}

	// Apply history nesting if enabled
	if h.NestHistory() {
		nested, err := ApplyHistoryNesting(ctx, currentMessages)
		if err != nil {
			return nil, fmt.Errorf("failed to apply history nesting: %w", err)
		}
		result.TransformedMessages = nested
		result.HistoryNested = true
	}

	return result, nil
}

// ApplyInputFilter applies the input filter to the handoff data.
func ApplyInputFilter(
	ctx context.Context,
	filter handoff.InputFilterFunc,
	data handoff.InputData,
) (handoff.InputData, error) {
	if filter == nil {
		return data, nil
	}

	return filter(ctx, data)
}

// CheckEnabled verifies if a handoff is currently enabled.
// Returns true if enabled, false otherwise.
// An error indicates the check itself failed (different from returning false).
func CheckEnabled(
	ctx context.Context,
	predicate handoff.EnabledPredicateFunc,
	h *handoff.Handoff,
) (bool, error) {
	if predicate == nil {
		// No predicate means always enabled
		return true, nil
	}

	// Note: We need access to context variables, but they're not available here
	// This will be called from the tool callback where context vars are available
	// For now, we'll just call the predicate with what we have
	return predicate(ctx, h.TargetAgent(), nil)
}

// ApplyHistoryNesting applies the default history mapper to nest/summarize messages.
func ApplyHistoryNesting(
	ctx context.Context,
	messages []openai.ChatCompletionMessageParamUnion,
) ([]openai.ChatCompletionMessageParamUnion, error) {
	return handoff.DefaultHistoryMapper(ctx, messages)
}
