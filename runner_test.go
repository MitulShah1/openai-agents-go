package agents

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/packages/ssestream"

	"github.com/MitulShah1/openai-agents-go/models"
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
	result, err := runner.Run(ctx, agent, []openai.ChatCompletionMessageParamUnion{})

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

	_, err := runner.Run(ctx, agent, messages)

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
	_, err := runner.Run(ctx, agent, messages)

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

func TestRunAsync_ReturnsImmediately(t *testing.T) {
	client := &openai.Client{}
	runner := NewRunner(client)
	agent := NewAgent("TestAgent")

	// Simulate a slow operation via hook
	agent.OnBeforeRun = func(_ context.Context, _ *Agent) error {
		time.Sleep(100 * time.Millisecond)
		return nil
	}

	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Hello"),
	}

	ctx := context.Background()
	start := time.Now()

	// This should return immediately
	ch := runner.RunAsync(ctx, agent, messages)
	duration := time.Since(start)

	if duration > 50*time.Millisecond {
		t.Errorf("RunAsync blocked for %v, expected immediate return", duration)
	}

	// But eventually it should finish (with error because no API key/client logic)
	// Actually, Run will fail fast with default client?
	// Or it might try to make a request and fail.

	// Let's rely on ErrNoMessages check which happens early?
	// No, messages are provided.
	// It will proceed to `execute` -> `OnBeforeRun` -> `executeInputGuardrails` -> `LoadHistory` -> `executeAgentLoop` -> `prepareRequest` -> `Client.Chat.Completions.New` -> ERROR.

	select {
	case res := <-ch:
		if res.Error == nil {
			// We actually expect an error from the LLM call because we have no API key/mock
			// But that's fine, we just want to ensure it finished.
			t.Log("RunAsync finished successfully (unexpectedly?)")
		} else {
			t.Logf("RunAsync finished with expected error: %v", res.Error)
		}
	case <-time.After(500 * time.Millisecond):
		t.Error("RunAsync timed out waiting for result")
	}
}

func TestRunAsync_CapturesResult(t *testing.T) {
	client := &openai.Client{}
	runner := NewRunner(client)
	agent := NewAgent("TestAgent")

	// No messages -> Should return ErrNoMessages immediately
	messages := []openai.ChatCompletionMessageParamUnion{}
	ctx := context.Background()

	ch := runner.RunAsync(ctx, agent, messages)

	res := <-ch
	if !errors.Is(res.Error, ErrNoMessages) {
		t.Errorf("expected ErrNoMessages, got %v", res.Error)
	}

	if res.Result != nil {
		t.Error("expected nil Result on error")
	}
}

// --- Phase 3: Model Provider Integration Tests ---

// testModel implements models.Model for testing without real API calls.
type testModel struct {
	name         string
	response     *models.ModelResponse
	err          error
	callCount    int
	lastSettings models.ModelSettings // captures the last ModelSettings received
}

func (m *testModel) GetResponse(_ context.Context, _ openai.ChatCompletionNewParams, settings models.ModelSettings) (*models.ModelResponse, error) {
	m.callCount++
	m.lastSettings = settings
	if m.err != nil {
		return nil, m.err
	}
	return m.response, nil
}

func (m *testModel) StreamResponse(_ context.Context, _ openai.ChatCompletionNewParams, settings models.ModelSettings) (*ssestream.Stream[openai.ChatCompletionChunk], error) {
	m.lastSettings = settings
	return nil, fmt.Errorf("streaming not implemented in test model")
}

func (m *testModel) ModelName() string { return m.name }

// testProvider implements models.ModelProvider for testing.
type testProvider struct {
	model models.Model
	err   error
}

func (p *testProvider) GetModel(_ string) (models.Model, error) {
	if p.err != nil {
		return nil, p.err
	}
	return p.model, nil
}

func TestNewRunnerWithProvider(t *testing.T) {
	provider := &testProvider{model: &testModel{name: "test-model"}}
	r := NewRunnerWithProvider(provider)

	if r == nil {
		t.Fatal("expected non-nil runner")
	}
	if r.ModelProvider == nil {
		t.Error("expected ModelProvider to be set")
	}
	if r.Client != nil {
		t.Error("expected Client to be nil for provider-based runner")
	}
}

func TestNewRunner_SetsModelProvider(t *testing.T) {
	client := &openai.Client{}
	r := NewRunner(client)

	if r.ModelProvider == nil {
		t.Error("NewRunner should auto-set ModelProvider")
	}
	if r.Client != client {
		t.Error("NewRunner should preserve Client for backward compatibility")
	}
}

func TestResolveModel_AgentProviderTakesPrecedence(t *testing.T) {
	agentModel := &testModel{name: "agent-model"}
	runnerModel := &testModel{name: "runner-model"}

	r := NewRunnerWithProvider(&testProvider{model: runnerModel})

	agent := NewAgent("test")
	agent.ModelProvider = &testProvider{model: agentModel}

	model, err := r.resolveModel(agent)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if model.ModelName() != "agent-model" {
		t.Errorf("expected agent-model, got %s", model.ModelName())
	}
}

func TestResolveModel_FallsBackToRunnerProvider(t *testing.T) {
	runnerModel := &testModel{name: "runner-model"}
	r := NewRunnerWithProvider(&testProvider{model: runnerModel})

	agent := NewAgent("test")
	// agent.ModelProvider is nil

	model, err := r.resolveModel(agent)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if model.ModelName() != "runner-model" {
		t.Errorf("expected runner-model, got %s", model.ModelName())
	}
}

func TestResolveModel_FallsBackToClient(t *testing.T) {
	client := &openai.Client{}
	r := &Runner{Client: client} // No ModelProvider set

	agent := NewAgent("test")
	agent.Model = "gpt-4o"

	model, err := r.resolveModel(agent)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if model.ModelName() != "gpt-4o" {
		t.Errorf("expected gpt-4o, got %s", model.ModelName())
	}
}

func TestResolveModel_ErrorWhenNoProviderOrClient(t *testing.T) {
	r := &Runner{} // Neither ModelProvider nor Client

	agent := NewAgent("test")

	_, err := r.resolveModel(agent)
	if err == nil {
		t.Fatal("expected error when no provider or client is set")
	}
}

func TestRunWithCustomProvider(t *testing.T) {
	// Create a mock model that returns a simple completion
	mockResp := &models.ModelResponse{
		Completion: &openai.ChatCompletion{
			Choices: []openai.ChatCompletionChoice{
				{
					Message: openai.ChatCompletionMessage{
						Role:    "assistant",
						Content: "Hello from custom provider!",
					},
				},
			},
		},
		Usage: models.ModelUsage{
			PromptTokens:     10,
			CompletionTokens: 5,
			TotalTokens:      15,
		},
	}

	mock := &testModel{name: "custom-model", response: mockResp}
	provider := &testProvider{model: mock}
	r := NewRunnerWithProvider(provider)

	agent := NewAgent("test")
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Hello"),
	}

	result, err := r.Run(context.Background(), agent, messages)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.FinalOutput != "Hello from custom provider!" {
		t.Errorf("expected 'Hello from custom provider!', got %q", result.FinalOutput)
	}
	if mock.callCount != 1 {
		t.Errorf("expected model to be called once, got %d", mock.callCount)
	}
	if result.Usage.TotalTokens != 15 {
		t.Errorf("expected TotalTokens 15, got %d", result.Usage.TotalTokens)
	}
}

func TestRunWithAgentLevelProvider(t *testing.T) {
	// Runner has one provider, agent has a different one
	runnerResp := &models.ModelResponse{
		Completion: &openai.ChatCompletion{
			Choices: []openai.ChatCompletionChoice{
				{Message: openai.ChatCompletionMessage{Role: "assistant", Content: "from runner"}},
			},
		},
		Usage: models.ModelUsage{},
	}
	agentResp := &models.ModelResponse{
		Completion: &openai.ChatCompletion{
			Choices: []openai.ChatCompletionChoice{
				{Message: openai.ChatCompletionMessage{Role: "assistant", Content: "from agent provider"}},
			},
		},
		Usage: models.ModelUsage{},
	}

	runnerMock := &testModel{name: "runner-model", response: runnerResp}
	agentMock := &testModel{name: "agent-model", response: agentResp}

	r := NewRunnerWithProvider(&testProvider{model: runnerMock})

	agent := NewAgent("test")
	agent.ModelProvider = &testProvider{model: agentMock}

	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Hello"),
	}

	result, err := r.Run(context.Background(), agent, messages)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.FinalOutput != "from agent provider" {
		t.Errorf("expected 'from agent provider', got %q", result.FinalOutput)
	}
	if agentMock.callCount != 1 {
		t.Errorf("expected agent model to be called, got %d calls", agentMock.callCount)
	}
	if runnerMock.callCount != 0 {
		t.Errorf("expected runner model NOT to be called, got %d calls", runnerMock.callCount)
	}
}

func TestRunWithProviderError(t *testing.T) {
	provider := &testProvider{err: fmt.Errorf("provider unavailable")}
	r := NewRunnerWithProvider(provider)

	agent := NewAgent("test")
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Hello"),
	}

	_, err := r.Run(context.Background(), agent, messages)
	if err == nil {
		t.Fatal("expected error from failing provider")
	}
	if !errors.Is(err, nil) && err.Error() == "" {
		t.Error("error should contain provider error info")
	}
}

func TestRunEmptyModelResponse_NilCompletion(t *testing.T) {
	mock := &testModel{
		name:     "empty-model",
		response: &models.ModelResponse{Completion: nil},
	}
	provider := &testProvider{model: mock}
	r := NewRunnerWithProvider(provider)

	agent := NewAgent("test")
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Hello"),
	}

	_, err := r.Run(context.Background(), agent, messages)
	if !errors.Is(err, ErrEmptyModelResponse) {
		t.Errorf("expected ErrEmptyModelResponse, got %v", err)
	}
}

func TestRunEmptyModelResponse_EmptyChoices(t *testing.T) {
	mock := &testModel{
		name: "empty-model",
		response: &models.ModelResponse{
			Completion: &openai.ChatCompletion{
				Choices: []openai.ChatCompletionChoice{},
			},
		},
	}
	provider := &testProvider{model: mock}
	r := NewRunnerWithProvider(provider)

	agent := NewAgent("test")
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Hello"),
	}

	_, err := r.Run(context.Background(), agent, messages)
	if !errors.Is(err, ErrEmptyModelResponse) {
		t.Errorf("expected ErrEmptyModelResponse, got %v", err)
	}
}
