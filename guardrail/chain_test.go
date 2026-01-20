package guardrail

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

// Helper function to create a test guardrail
func testGuardrail(name string, passed bool, tripwire bool, delay time.Duration) *Guardrail {
	return &Guardrail{
		Name: name,
		Func: func(ctx context.Context, input string) (*Result, error) {
			if delay > 0 {
				time.Sleep(delay)
			}
			return &Result{
				Passed:            passed,
				TripwireTriggered: tripwire,
				Message:           fmt.Sprintf("%s result", name),
				Metadata:          map[string]any{"name": name},
			}, nil
		},
	}
}

// Helper function to create a test guardrail that returns an error
func testGuardrailWithError(name string, err error) *Guardrail {
	return &Guardrail{
		Name: name,
		Func: func(ctx context.Context, input string) (*Result, error) {
			return nil, err
		},
	}
}

func TestNewChain(t *testing.T) {
	builder := NewChain()
	if builder == nil {
		t.Fatal("Expected non-nil builder")
	}
	if builder.chain == nil {
		t.Fatal("Expected non-nil chain")
	}
	if builder.chain.strategy != Sequential {
		t.Errorf("Expected default strategy to be Sequential, got %v", builder.chain.strategy)
	}
}

func TestChainBuilder_Add(t *testing.T) {
	g1 := testGuardrail("test1", true, false, 0)
	g2 := testGuardrail("test2", true, false, 0)

	builder := NewChain().Add(g1).Add(g2)

	if len(builder.chain.guardrails) != 2 {
		t.Errorf("Expected 2 guardrails, got %d", len(builder.chain.guardrails))
	}

	// Test adding nil guardrail
	builder.Add(nil)
	if len(builder.chain.guardrails) != 2 {
		t.Error("Adding nil should not increase count")
	}
}

func TestChainBuilder_WithStrategy(t *testing.T) {
	tests := []struct {
		name     string
		strategy Strategy
	}{
		{"Sequential", Sequential},
		{"Parallel", Parallel},
		{"StopOnFirstPass", StopOnFirstPass},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewChain().WithStrategy(tt.strategy)
			if builder.chain.strategy != tt.strategy {
				t.Errorf("Expected strategy %v, got %v", tt.strategy, builder.chain.strategy)
			}
		})
	}
}

func TestChainBuilder_WithName(t *testing.T) {
	customName := "my_custom_chain"
	builder := NewChain().WithName(customName)

	if builder.chain.name != customName {
		t.Errorf("Expected name %q, got %q", customName, builder.chain.name)
	}
}

func TestChain_Sequential_AllPass(t *testing.T) {
	g1 := testGuardrail("guard1", true, false, 0)
	g2 := testGuardrail("guard2", true, false, 0)
	g3 := testGuardrail("guard3", true, false, 0)

	chain := NewChain().
		Add(g1).
		Add(g2).
		Add(g3).
		WithStrategy(Sequential).
		Build()

	result, err := chain.Func(context.Background(), "test input")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result.Passed {
		t.Error("Expected chain to pass")
	}
	if result.TripwireTriggered {
		t.Error("Expected no tripwire")
	}
}

func TestChain_Sequential_FailFast(t *testing.T) {
	g1 := testGuardrail("guard1", true, false, 0)
	g2 := testGuardrail("guard2", false, false, 0) // This will fail
	g3 := testGuardrail("guard3", true, false, 0)  // Should not be executed

	chain := NewChain().
		Add(g1).
		Add(g2).
		Add(g3).
		WithStrategy(Sequential).
		Build()

	result, err := chain.Func(context.Background(), "test input")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result.Passed {
		t.Error("Expected chain to fail")
	}

	// Verify it stopped after guard2
	results, ok := result.Metadata["results"].([]*Result)
	if !ok || len(results) != 2 {
		t.Errorf("Expected 2 results in metadata (short-circuit), got %d", len(results))
	}
}

func TestChain_Sequential_WithTripwire(t *testing.T) {
	g1 := testGuardrail("guard1", true, false, 0)
	g2 := testGuardrail("guard2", false, true, 0) // Fail with tripwire

	chain := NewChain().
		Add(g1).
		Add(g2).
		WithStrategy(Sequential).
		Build()

	result, err := chain.Func(context.Background(), "test input")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result.Passed {
		t.Error("Expected chain to fail")
	}
	if !result.TripwireTriggered {
		t.Error("Expected tripwire to be triggered")
	}
}

func TestChain_Parallel_AllPass(t *testing.T) {
	g1 := testGuardrail("guard1", true, false, 10*time.Millisecond)
	g2 := testGuardrail("guard2", true, false, 10*time.Millisecond)
	g3 := testGuardrail("guard3", true, false, 10*time.Millisecond)

	chain := NewChain().
		Add(g1).
		Add(g2).
		Add(g3).
		WithStrategy(Parallel).
		Build()

	start := time.Now()
	result, err := chain.Func(context.Background(), "test input")
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result.Passed {
		t.Error("Expected chain to pass")
	}

	// Should complete in ~10ms (parallel) not ~30ms (sequential)
	if elapsed > 25*time.Millisecond {
		t.Errorf("Expected parallel execution to be faster, took %v", elapsed)
	}
}

func TestChain_Parallel_SomeFail(t *testing.T) {
	g1 := testGuardrail("guard1", true, false, 0)
	g2 := testGuardrail("guard2", false, false, 0)
	g3 := testGuardrail("guard3", false, true, 0) // Fail with tripwire

	chain := NewChain().
		Add(g1).
		Add(g2).
		Add(g3).
		WithStrategy(Parallel).
		Build()

	result, err := chain.Func(context.Background(), "test input")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result.Passed {
		t.Error("Expected chain to fail")
	}
	if !result.TripwireTriggered {
		t.Error("Expected tripwire to be triggered")
	}

	// All guardrails should have executed
	results, ok := result.Metadata["results"].([]*Result)
	if !ok || len(results) != 3 {
		t.Errorf("Expected 3 results in metadata (all executed), got %d", len(results))
	}
}

func TestChain_StopOnFirstPass_FirstPasses(t *testing.T) {
	g1 := testGuardrail("guard1", true, false, 0) // This passes, should stop here
	g2 := testGuardrail("guard2", false, false, 0)
	g3 := testGuardrail("guard3", false, false, 0)

	chain := NewChain().
		Add(g1).
		Add(g2).
		Add(g3).
		WithStrategy(StopOnFirstPass).
		Build()

	result, err := chain.Func(context.Background(), "test input")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result.Passed {
		t.Error("Expected chain to pass")
	}

	// Should stop after first pass
	results, ok := result.Metadata["results"].([]*Result)
	if !ok || len(results) != 1 {
		t.Errorf("Expected 1 result in metadata (stopped on first pass), got %d", len(results))
	}
}

func TestChain_StopOnFirstPass_AllFail(t *testing.T) {
	g1 := testGuardrail("guard1", false, false, 0)
	g2 := testGuardrail("guard2", false, false, 0)
	g3 := testGuardrail("guard3", false, false, 0)

	chain := NewChain().
		Add(g1).
		Add(g2).
		Add(g3).
		WithStrategy(StopOnFirstPass).
		Build()

	result, err := chain.Func(context.Background(), "test input")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result.Passed {
		t.Error("Expected chain to fail")
	}

	// All guardrails should have executed
	results, ok := result.Metadata["results"].([]*Result)
	if !ok || len(results) != 3 {
		t.Errorf("Expected 3 results in metadata (all executed), got %d", len(results))
	}
}

func TestChain_EmptyChain(t *testing.T) {
	chain := NewChain().Build()

	result, err := chain.Func(context.Background(), "test input")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result.Passed {
		t.Error("Expected empty chain to pass")
	}
}

func TestChain_ErrorHandling(t *testing.T) {
	testErr := errors.New("guardrail error")
	g1 := testGuardrail("guard1", true, false, 0)
	g2 := testGuardrailWithError("guard2", testErr)

	chain := NewChain().
		Add(g1).
		Add(g2).
		WithStrategy(Sequential).
		Build()

	_, err := chain.Func(context.Background(), "test input")
	if err == nil {
		t.Fatal("Expected error from chain")
	}
	if !errors.Is(err, testErr) {
		t.Errorf("Expected error to wrap test error, got %v", err)
	}
}

func TestStrategy_String(t *testing.T) {
	tests := []struct {
		strategy Strategy
		expected string
	}{
		{Sequential, "Sequential"},
		{Parallel, "Parallel"},
		{StopOnFirstPass, "StopOnFirstPass"},
		{Strategy(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if tt.strategy.String() != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, tt.strategy.String())
			}
		})
	}
}
