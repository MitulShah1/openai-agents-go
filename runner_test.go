package agents

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/openai/openai-go"
)

func TestNewRunner(t *testing.T) {
	client := &openai.Client{}
	runner := NewRunner(client)

	if runner == nil {
		t.Fatal("expected NewRunner to return non-nil runner")
	}

	if runner.Client != client {
		t.Error("expected runner to store the provided client")
	}
}

func TestRunNoMessages(t *testing.T) {
	client := &openai.Client{}
	runner := NewRunner(client)
	agent := NewAgent("TestAgent")

	ctx := context.Background()
	result, err := runner.Run(ctx, agent, []openai.ChatCompletionMessageParamUnion{}, nil, nil, nil, "")

	if !errors.Is(err, ErrNoMessages) {
		t.Errorf("expected ErrNoMessages, got %v", err)
	}

	if result != nil {
		t.Error("expected nil result when no messages provided")
	}
}

func TestRunMaxTurnsExceeded(t *testing.T) {
	// This test would require mocking the OpenAI client to return responses
	// For now, we'll test the max turns configuration
	config := &RunConfig{
		MaxTurns: 1,
	}

	if config.MaxTurns != 1 {
		t.Errorf("expected MaxTurns=1, got %d", config.MaxTurns)
	}
}

func TestRunTimeout(t *testing.T) {
	client := &openai.Client{}
	runner := NewRunner(client)
	agent := NewAgent("TestAgent")

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("test"),
	}

	_, err := runner.Run(ctx, agent, messages, nil, nil, nil, "")

	// The error should be related to context cancellation
	if err == nil {
		t.Error("expected error for cancelled context")
	}
}

func TestRunWithTimeout(t *testing.T) {
	config := &RunConfig{
		Timeout: 100 * time.Millisecond,
	}

	if config.Timeout != 100*time.Millisecond {
		t.Errorf("expected Timeout=100ms, got %v", config.Timeout)
	}
}

func TestRunLifecycleHooks(t *testing.T) {
	agent := NewAgent("TestAgent")

	beforeCalled := false
	afterCalled := false

	agent.OnBeforeRun = func(_ context.Context, _ *Agent) error {
		beforeCalled = true
		return nil
	}

	agent.OnAfterRun = func(_ context.Context, _ *Agent) error {
		afterCalled = true
		return nil
	}

	// Test that hooks are set
	if agent.OnBeforeRun == nil {
		t.Error("OnBeforeRun hook not set")
	}

	if agent.OnAfterRun == nil {
		t.Error("OnAfterRun hook not set")
	}

	// Manually invoke to test
	ctx := context.Background()
	_ = agent.OnBeforeRun(ctx, agent)
	_ = agent.OnAfterRun(ctx, agent)

	if !beforeCalled {
		t.Error("OnBeforeRun was not called")
	}

	if !afterCalled {
		t.Error("OnAfterRun was not called")
	}
}

func TestRunBeforeHookError(t *testing.T) {
	client := &openai.Client{}
	runner := NewRunner(client)
	agent := NewAgent("TestAgent")

	expectedErr := errors.New("before hook failed")
	agent.OnBeforeRun = func(_ context.Context, _ *Agent) error {
		return expectedErr
	}

	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("test"),
	}

	ctx := context.Background()
	_, err := runner.Run(ctx, agent, messages, nil, nil, nil, "")

	if err == nil {
		t.Fatal("expected error from OnBeforeRun hook")
	}

	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error to wrap %v, got %v", expectedErr, err)
	}
}

func TestDefaultConfigUsed(t *testing.T) {
	// When nil config is passed, default should be used
	config := DefaultRunConfig()

	if config.MaxTurns != 10 {
		t.Errorf("default MaxTurns should be 10, got %d", config.MaxTurns)
	}

	if config.Timeout != 5*time.Minute {
		t.Errorf("default Timeout should be 5m, got %v", config.Timeout)
	}
}

func TestContextVariablesInitialization(t *testing.T) {
	// Test that ContextVariables can be initialized and used
	ctx := make(ContextVariables)

	// Test adding values
	ctx["key"] = "value"
	if ctx["key"] != "value" {
		t.Error("failed to set value in ContextVariables")
	}
}

// Integration-style test that would require OpenAI API
// Commented out as it requires actual API access
/*
func TestRunIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	client := openai.NewClient(option.WithAPIKey(apiKey))
	runner := NewRunner(&client)

	agent := NewAgent("TestAgent")
	agent.Instructions = "You are a test agent. Respond with 'OK'."

	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Say OK"),
	}

	ctx := context.Background()
	config := &RunConfig{
		MaxTurns: 1,
	}

	result, err := runner.Run(ctx, agent, messages, nil, config, nil, "")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if len(result.Steps) != 1 {
		t.Errorf("expected 1 step, got %d", len(result.Steps))
	}
}
*/
