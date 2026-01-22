package handoff

import (
	"context"
	"errors"
	"testing"

	"github.com/openai/openai-go/v3"

	agents "github.com/MitulShah1/openai-agents-go"
	"github.com/MitulShah1/openai-agents-go/handoff"
)

// Test Execute function

func TestExecute_Success(t *testing.T) {
	agent := agents.NewAgent("TestAgent")
	h := handoff.New(agent)

	inputData := handoff.InputData{
		Agent:       agent,
		NewItems:    []openai.ChatCompletionMessageParamUnion{},
		ContextVars: agents.ContextVariables{},
	}

	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Hello"),
	}

	result, err := Execute(context.Background(), h, inputData, messages)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if result.TargetAgent != h {
		t.Error("Expected TargetAgent to be the handoff")
	}

	if result.HistoryNested {
		t.Error("Expected HistoryNested to be false by default")
	}

	if result.FilterApplied {
		t.Error("Expected FilterApplied to be false when no filter")
	}
}

func TestExecute_WithInputFilter(t *testing.T) {
	agent := agents.NewAgent("TestAgent")

	filterCalled := false
	h := handoff.New(agent, handoff.WithInputFilter(func(_ context.Context, data handoff.InputData) (handoff.InputData, error) {
		filterCalled = true
		// Add a marker to context vars
		data.ContextVars["filtered"] = true
		return data, nil
	}))

	inputData := handoff.InputData{
		Agent:       agent,
		NewItems:    []openai.ChatCompletionMessageParamUnion{},
		ContextVars: agents.ContextVariables{},
	}

	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Test"),
	}

	result, err := Execute(context.Background(), h, inputData, messages)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !filterCalled {
		t.Error("Expected input filter to be called")
	}

	if !result.FilterApplied {
		t.Error("Expected FilterApplied to be true")
	}
}

func TestExecute_InputFilterError(t *testing.T) {
	agent := agents.NewAgent("TestAgent")
	expectedErr := errors.New("filter error")

	h := handoff.New(agent, handoff.WithInputFilter(func(_ context.Context, data handoff.InputData) (handoff.InputData, error) {
		return data, expectedErr
	}))

	inputData := handoff.InputData{
		Agent:       agent,
		NewItems:    []openai.ChatCompletionMessageParamUnion{},
		ContextVars: agents.ContextVariables{},
	}

	messages := []openai.ChatCompletionMessageParamUnion{}

	_, err := Execute(context.Background(), h, inputData, messages)

	if err == nil {
		t.Error("Expected error from filter, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Errorf("Expected error to wrap filter error, got %v", err)
	}
}

func TestExecute_WithHistoryNesting(t *testing.T) {
	agent := agents.NewAgent("TestAgent")
	h := handoff.New(agent, handoff.WithHistoryNesting(true))

	inputData := handoff.InputData{
		Agent:       agent,
		NewItems:    []openai.ChatCompletionMessageParamUnion{},
		ContextVars: agents.ContextVariables{},
	}

	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Hello"),
		openai.AssistantMessage("Hi there"),
		openai.UserMessage("How are you?"),
	}

	result, err := Execute(context.Background(), h, inputData, messages)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !result.HistoryNested {
		t.Error("Expected HistoryNested to be true")
	}

	// TransformedMessages should be returned (nested/summarized)
	if len(result.TransformedMessages) == 0 {
		t.Error("Expected transformed messages to be returned")
	}

	// Should be summarized to 1 message
	if len(result.TransformedMessages) != 1 {
		t.Errorf("Expected 1 summary message, got %d", len(result.TransformedMessages))
	}
}

func TestApplyInputFilter_NoFilter(t *testing.T) {
	inputData := handoff.InputData{
		Agent:       agents.NewAgent("Test"),
		NewItems:    []openai.ChatCompletionMessageParamUnion{},
		ContextVars: agents.ContextVariables{},
	}

	result, err := ApplyInputFilter(context.Background(), nil, inputData)

	if err != nil {
		t.Errorf("Expected no error with nil filter, got %v", err)
	}

	// Should return the input unchanged
	if len(result.NewItems) != 0 {
		t.Error("Expected input data to be unchanged")
	}
}

func TestApplyInputFilter_WithFilter(t *testing.T) {
	filter := func(_ context.Context, data handoff.InputData) (handoff.InputData, error) {
		data.ContextVars["modified"] = true
		return data, nil
	}

	inputData := handoff.InputData{
		Agent:       agents.NewAgent("Test"),
		NewItems:    []openai.ChatCompletionMessageParamUnion{},
		ContextVars: agents.ContextVariables{},
	}

	result, err := ApplyInputFilter(context.Background(), filter, inputData)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result.ContextVars["modified"] != true {
		t.Error("Expected filter to modify context vars")
	}
}

func TestApplyInputFilter_Error(t *testing.T) {
	expectedErr := errors.New("filter failed")
	filter := func(_ context.Context, data handoff.InputData) (handoff.InputData, error) {
		return data, expectedErr
	}

	inputData := handoff.InputData{
		Agent:       agents.NewAgent("Test"),
		NewItems:    []openai.ChatCompletionMessageParamUnion{},
		ContextVars: agents.ContextVariables{},
	}

	_, err := ApplyInputFilter(context.Background(), filter, inputData)

	if err == nil {
		t.Error("Expected error from filter")
	}

	if !errors.Is(err, expectedErr) {
		t.Errorf("Expected error to be filter error, got %v", err)
	}
}

// Test CheckEnabled

func TestCheckEnabled_NoPredicate(t *testing.T) {
	// Pass a non-nil Handoff struct, even if empty, to avoid nil pointer dereference
	h := &handoff.Handoff{}
	enabled, err := CheckEnabled(context.Background(), nil, h)

	if err != nil {
		t.Errorf("Expected no error with nil predicate, got %v", err)
	}

	if !enabled {
		t.Error("Expected enabled to be true when no predicate")
	}
}

func TestCheckEnabled_WithPredicate_True(t *testing.T) {
	agent := agents.NewAgent("Test")
	h := handoff.New(agent)
	predicate := func(_ context.Context, _ *agents.Agent, _ agents.ContextVariables) (bool, error) {
		return true, nil
	}

	enabled, err := CheckEnabled(context.Background(), predicate, h)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !enabled {
		t.Error("Expected enabled to be true")
	}
}

func TestCheckEnabled_WithPredicate_False(t *testing.T) {
	agent := agents.NewAgent("Test")
	h := handoff.New(agent)
	predicate := func(_ context.Context, _ *agents.Agent, _ agents.ContextVariables) (bool, error) {
		return false, nil
	}

	enabled, err := CheckEnabled(context.Background(), predicate, h)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if enabled {
		t.Error("Expected enabled to be false")
	}
}

func TestCheckEnabled_PredicateError(t *testing.T) {
	agent := agents.NewAgent("Test")
	h := handoff.New(agent)
	expectedErr := errors.New("predicate error")
	predicate := func(_ context.Context, _ *agents.Agent, _ agents.ContextVariables) (bool, error) {
		return false, expectedErr
	}

	_, err := CheckEnabled(context.Background(), predicate, h)

	if err == nil {
		t.Error("Expected error from predicate")
	}

	if !errors.Is(err, expectedErr) {
		t.Errorf("Expected error to be predicate error, got %v", err)
	}
}

// Test ApplyHistoryNesting

func TestApplyHistoryNesting_Disabled(t *testing.T) {
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Message 1"),
		openai.UserMessage("Message 2"),
	}

	// When not nesting, use FlattenHistoryMapper
	result, err := handoff.FlattenHistoryMapper(context.Background(), messages)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Should return messages unchanged
	if len(result) != len(messages) {
		t.Errorf("Expected %d messages, got %d", len(messages), len(result))
	}
}

func TestApplyHistoryNesting_Enabled(t *testing.T) {
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Message 1"),
		openai.AssistantMessage("Response 1"),
		openai.UserMessage("Message 2"),
	}

	// ApplyHistoryNesting always uses DefaultHistoryMapper
	result, err := ApplyHistoryNesting(context.Background(), messages)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Should summarize to 1 message
	if len(result) != 1 {
		t.Errorf("Expected 1 summary message, got %d", len(result))
	}
}

func TestApplyHistoryNesting_EmptyMessages(t *testing.T) {
	messages := []openai.ChatCompletionMessageParamUnion{}

	// ApplyHistoryNesting should handle empty messages
	result, err := ApplyHistoryNesting(context.Background(), messages)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Expected 0 messages, got %d", len(result))
	}
}

// Test combined scenarios

func TestExecute_AllFeaturesEnabled(t *testing.T) {
	agent := agents.NewAgent("TestAgent")

	filterCalled := false

	h := handoff.New(
		agent,
		handoff.WithInputFilter(func(_ context.Context, data handoff.InputData) (handoff.InputData, error) {
			filterCalled = true
			return data, nil
		}),
		handoff.WithHistoryNesting(true),
	)

	inputData := handoff.InputData{
		Agent:       agent,
		NewItems:    []openai.ChatCompletionMessageParamUnion{},
		ContextVars: agents.ContextVariables{},
	}

	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Test message"),
	}

	result, err := Execute(context.Background(), h, inputData, messages)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !filterCalled {
		t.Error("Expected filter to be called")
	}

	if !result.FilterApplied {
		t.Error("Expected FilterApplied to be true")
	}

	if !result.HistoryNested {
		t.Error("Expected HistoryNested to be true")
	}
}

func TestExecute_ContextCancellation(t *testing.T) {
	agent := agents.NewAgent("TestAgent")
	h := handoff.New(agent, handoff.WithInputFilter(func(ctx context.Context, data handoff.InputData) (handoff.InputData, error) {
		// Check if context is cancelled
		if ctx.Err() != nil {
			return data, ctx.Err()
		}
		return data, nil
	}))

	inputData := handoff.InputData{
		Agent:       agent,
		NewItems:    []openai.ChatCompletionMessageParamUnion{},
		ContextVars: agents.ContextVariables{},
	}

	messages := []openai.ChatCompletionMessageParamUnion{}

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := Execute(ctx, h, inputData, messages)

	if err == nil {
		t.Error("Expected error from cancelled context")
	}

	if !errors.Is(err, context.Canceled) {
		t.Errorf("Expected context.Canceled error, got %v", err)
	}
}
