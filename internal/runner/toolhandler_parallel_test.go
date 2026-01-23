package runner

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/openai/openai-go/v3"
)

const (
	testResult1 = "result1"
	testResult2 = "result2"
)

// funcToolExecutor implements ToolExecutor with a function for testing parallel execution
type funcToolExecutor struct {
	fn func(arguments string, contextVariables map[string]any) (any, error)
}

func (f *funcToolExecutor) Execute(arguments string, contextVariables map[string]any) (any, error) {
	return f.fn(arguments, contextVariables)
}

// TestParallelToolExecution verifies that tools execute in parallel
func TestParallelToolExecution(t *testing.T) {
	ctx := context.Background()

	// Track execution times
	var mu sync.Mutex
	execTimes := make(map[string]time.Time)

	// Create tools that sleep and record execution time
	tool1 := &funcToolExecutor{
		fn: func(_ string, _ map[string]any) (any, error) {
			mu.Lock()
			execTimes["tool1"] = time.Now()
			mu.Unlock()
			time.Sleep(100 * time.Millisecond)
			return testResult1, nil
		},
	}

	tool2 := &funcToolExecutor{
		fn: func(_ string, _ map[string]any) (any, error) {
			mu.Lock()
			execTimes["tool2"] = time.Now()
			mu.Unlock()
			time.Sleep(100 * time.Millisecond)
			return testResult2, nil
		},
	}

	toolMap := ToolMap{
		"tool1": tool1,
		"tool2": tool2,
	}

	toolCalls := []openai.ChatCompletionMessageToolCallUnion{
		{
			ID:   "call1",
			Type: "function",
			Function: openai.ChatCompletionMessageFunctionToolCallFunction{
				Name:      "tool1",
				Arguments: "{}",
			},
		},
		{
			ID:   "call2",
			Type: "function",
			Function: openai.ChatCompletionMessageFunctionToolCallFunction{
				Name:      "tool2",
				Arguments: "{}",
			},
		},
	}

	// Test parallel execution
	start := time.Now()
	messages, results, _ := HandleToolCalls(ctx, toolCalls, toolMap, nil, nil, true, 0)
	duration := time.Since(start)

	// Verify parallel execution (should be ~100ms, not ~200ms)
	if duration > 150*time.Millisecond {
		t.Errorf("Expected parallel execution ~100ms, got %v", duration)
	}

	// Verify both tools executed
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	if len(messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(messages))
	}

	// Verify execution times overlap (parallel)
	mu.Lock()
	defer mu.Unlock()

	t1, ok1 := execTimes["tool1"]
	t2, ok2 := execTimes["tool2"]

	if !ok1 || !ok2 {
		t.Fatal("Not all tools executed")
	}

	timeDiff := t2.Sub(t1)
	if timeDiff < 0 {
		timeDiff = -timeDiff
	}

	// Tools should start within 50ms of each other (parallel)
	if timeDiff > 50*time.Millisecond {
		t.Errorf("Tools did not execute in parallel, time diff: %v", timeDiff)
	}
}

// TestSequentialToolExecution verifies that tools execute sequentially when parallel=false
func TestSequentialToolExecution(t *testing.T) {
	ctx := context.Background()

	// Track execution times
	var mu sync.Mutex
	execTimes := make(map[string]time.Time)

	tool1 := &funcToolExecutor{
		fn: func(_ string, _ map[string]any) (any, error) {
			mu.Lock()
			execTimes["tool1"] = time.Now()
			mu.Unlock()
			time.Sleep(50 * time.Millisecond)
			return testResult1, nil
		},
	}

	tool2 := &funcToolExecutor{
		fn: func(_ string, _ map[string]any) (any, error) {
			mu.Lock()
			execTimes["tool2"] = time.Now()
			mu.Unlock()
			time.Sleep(50 * time.Millisecond)
			return testResult2, nil
		},
	}

	toolMap := ToolMap{
		"tool1": tool1,
		"tool2": tool2,
	}

	toolCalls := []openai.ChatCompletionMessageToolCallUnion{
		{
			ID:   "call1",
			Type: "function",
			Function: openai.ChatCompletionMessageFunctionToolCallFunction{
				Name:      "tool1",
				Arguments: "{}",
			},
		},
		{
			ID:   "call2",
			Type: "function",
			Function: openai.ChatCompletionMessageFunctionToolCallFunction{
				Name:      "tool2",
				Arguments: "{}",
			},
		},
	}

	// Test sequential execution
	start := time.Now()
	messages, results, _ := HandleToolCalls(ctx, toolCalls, toolMap, nil, nil, false, 0)
	duration := time.Since(start)

	// Verify sequential execution (should be ~100ms)
	if duration < 90*time.Millisecond {
		t.Errorf("Expected sequential execution ~100ms, got %v (too fast)", duration)
	}

	// Verify both tools executed
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	if len(messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(messages))
	}

	// Verify execution times are sequential
	mu.Lock()
	defer mu.Unlock()

	t1, ok1 := execTimes["tool1"]
	t2, ok2 := execTimes["tool2"]

	if !ok1 || !ok2 {
		t.Fatal("Not all tools executed")
	}

	// Tool2 should start after tool1 completes (50ms+ apart)
	timeDiff := t2.Sub(t1)
	if timeDiff < 40*time.Millisecond {
		t.Errorf("Tools executed in parallel, expected sequential. Time diff: %v", timeDiff)
	}
}

// TestConcurrencyLimiting verifies that MaxConcurrency limits parallel execution
func TestConcurrencyLimiting(t *testing.T) {
	ctx := context.Background()

	// Track concurrent executions
	var mu sync.Mutex
	var maxConcurrent int
	var currentConcurrent int

	createTool := func(name string) *funcToolExecutor {
		return &funcToolExecutor{
			fn: func(_ string, _ map[string]any) (any, error) {
				mu.Lock()
				currentConcurrent++
				if currentConcurrent > maxConcurrent {
					maxConcurrent = currentConcurrent
				}
				mu.Unlock()

				time.Sleep(50 * time.Millisecond)

				mu.Lock()
				currentConcurrent--
				mu.Unlock()

				return "result_" + name, nil
			},
		}
	}

	toolMap := ToolMap{
		"tool1": createTool("1"),
		"tool2": createTool("2"),
		"tool3": createTool("3"),
		"tool4": createTool("4"),
	}

	toolCalls := []openai.ChatCompletionMessageToolCallUnion{
		{ID: "call1", Type: "function", Function: openai.ChatCompletionMessageFunctionToolCallFunction{Name: "tool1", Arguments: "{}"}},
		{ID: "call2", Type: "function", Function: openai.ChatCompletionMessageFunctionToolCallFunction{Name: "tool2", Arguments: "{}"}},
		{ID: "call3", Type: "function", Function: openai.ChatCompletionMessageFunctionToolCallFunction{Name: "tool3", Arguments: "{}"}},
		{ID: "call4", Type: "function", Function: openai.ChatCompletionMessageFunctionToolCallFunction{Name: "tool4", Arguments: "{}"}},
	}

	// Test with max concurrency of 2
	messages, results, _ := HandleToolCalls(ctx, toolCalls, toolMap, nil, nil, true, 2)

	// Verify all tools executed
	if len(results) != 4 {
		t.Errorf("Expected 4 results, got %d", len(results))
	}

	if len(messages) != 4 {
		t.Errorf("Expected 4 messages, got %d", len(messages))
	}

	// Verify max concurrency was respected
	mu.Lock()
	defer mu.Unlock()

	if maxConcurrent > 2 {
		t.Errorf("Expected max concurrency of 2, got %d", maxConcurrent)
	}

	if maxConcurrent < 1 {
		t.Error("Expected at least 1 concurrent execution")
	}
}

// TestOrderPreservation verifies that results maintain tool call order
func TestOrderPreservation(t *testing.T) {
	ctx := context.Background()

	// Create tools with varying execution times
	tool1 := &funcToolExecutor{
		fn: func(_ string, _ map[string]any) (any, error) {
			time.Sleep(100 * time.Millisecond)
			return "result1", nil
		},
	}

	tool2 := &funcToolExecutor{
		fn: func(_ string, _ map[string]any) (any, error) {
			time.Sleep(10 * time.Millisecond)
			return "result2", nil
		},
	}

	tool3 := &funcToolExecutor{
		fn: func(_ string, _ map[string]any) (any, error) {
			time.Sleep(50 * time.Millisecond)
			return "result3", nil
		},
	}

	toolMap := ToolMap{
		"tool1": tool1,
		"tool2": tool2,
		"tool3": tool3,
	}

	toolCalls := []openai.ChatCompletionMessageToolCallUnion{
		{ID: "call1", Type: "function", Function: openai.ChatCompletionMessageFunctionToolCallFunction{Name: "tool1", Arguments: "{}"}},
		{ID: "call2", Type: "function", Function: openai.ChatCompletionMessageFunctionToolCallFunction{Name: "tool2", Arguments: "{}"}},
		{ID: "call3", Type: "function", Function: openai.ChatCompletionMessageFunctionToolCallFunction{Name: "tool3", Arguments: "{}"}},
	}

	// Execute in parallel
	messages, results, _ := HandleToolCalls(ctx, toolCalls, toolMap, nil, nil, true, 0)

	// Verify order is preserved despite different execution times
	if len(results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(results))
	}

	expectedResults := []string{"result1", "result2", "result3"}
	for i, expected := range expectedResults {
		if results[i].Result != expected {
			t.Errorf("Result %d: expected %s, got %v", i, expected, results[i].Result)
		}
	}

	// Verify message order
	if len(messages) != 3 {
		t.Fatalf("Expected 3 messages, got %d", len(messages))
	}
}

// TestParallelErrorHandling verifies that errors in one tool don't block others
func TestParallelErrorHandling(t *testing.T) {
	ctx := context.Background()

	tool1 := &funcToolExecutor{
		fn: func(_ string, _ map[string]any) (any, error) {
			time.Sleep(50 * time.Millisecond)
			return nil, errors.New("tool1 execution failed")
		},
	}

	tool2 := &funcToolExecutor{
		fn: func(_ string, _ map[string]any) (any, error) {
			time.Sleep(50 * time.Millisecond)
			return "success", nil
		},
	}

	toolMap := ToolMap{
		"tool1": tool1,
		"tool2": tool2,
	}

	toolCalls := []openai.ChatCompletionMessageToolCallUnion{
		{ID: "call1", Type: "function", Function: openai.ChatCompletionMessageFunctionToolCallFunction{Name: "tool1", Arguments: "{}"}},
		{ID: "call2", Type: "function", Function: openai.ChatCompletionMessageFunctionToolCallFunction{Name: "tool2", Arguments: "{}"}},
	}

	// Execute in parallel
	messages, results, _ := HandleToolCalls(ctx, toolCalls, toolMap, nil, nil, true, 0)

	// Verify both tools executed
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	// Verify tool1 has error
	if results[0].Error == nil {
		t.Error("Expected tool1 to have error")
	}

	// Verify tool2 succeeded
	if results[1].Error != nil {
		t.Errorf("Expected tool2 to succeed, got error: %v", results[1].Error)
	}

	if results[1].Result != "success" {
		t.Errorf("Expected tool2 result 'success', got %v", results[1].Result)
	}

	// Verify both messages were created
	if len(messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(messages))
	}
}
