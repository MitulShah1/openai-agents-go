// Package guardrail provides input/output validation for agents.
package guardrail

import (
	"context"
	"fmt"
	"sync"
)

// Strategy defines how guardrails are executed in a chain.
type Strategy int

const (
	// Sequential runs guardrails one by one, stopping on first failure.
	Sequential Strategy = iota
	// Parallel runs all guardrails concurrently and aggregates results.
	Parallel
	// StopOnFirstPass runs guardrails until one passes (OR logic).
	StopOnFirstPass
)

// String returns the string representation of the Strategy.
func (s Strategy) String() string {
	switch s {
	case Sequential:
		return "Sequential"
	case Parallel:
		return "Parallel"
	case StopOnFirstPass:
		return "StopOnFirstPass"
	default:
		return "Unknown"
	}
}

// Chain represents a composition of multiple guardrails.
type Chain struct {
	guardrails []*Guardrail
	strategy   Strategy
	name       string
}

// ChainBuilder provides a fluent API for building guardrail chains.
type ChainBuilder struct {
	chain *Chain
}

// NewChain creates a new ChainBuilder for composing guardrails.
func NewChain() *ChainBuilder {
	return &ChainBuilder{
		chain: &Chain{
			guardrails: make([]*Guardrail, 0),
			strategy:   Sequential, // Default strategy
			name:       "guardrail_chain",
		},
	}
}

// Add appends a guardrail to the chain.
func (b *ChainBuilder) Add(g *Guardrail) *ChainBuilder {
	if g != nil {
		b.chain.guardrails = append(b.chain.guardrails, g)
	}
	return b
}

// WithStrategy sets the execution strategy for the chain.
func (b *ChainBuilder) WithStrategy(strategy Strategy) *ChainBuilder {
	b.chain.strategy = strategy
	return b
}

// WithName sets a custom name for the chain guardrail.
func (b *ChainBuilder) WithName(name string) *ChainBuilder {
	b.chain.name = name
	return b
}

// Build returns a Guardrail that executes the chain.
func (b *ChainBuilder) Build() *Guardrail {
	chain := b.chain
	return &Guardrail{
		Name: chain.name,
		Func: func(ctx context.Context, input string) (*Result, error) {
			return chain.execute(ctx, input)
		},
		RunInParallel: false, // Chains always run sequentially
	}
}

// execute runs the chain according to its strategy.
func (c *Chain) execute(ctx context.Context, input string) (*Result, error) {
	if len(c.guardrails) == 0 {
		return &Result{
			Passed:            true,
			TripwireTriggered: false,
			Message:           "no guardrails in chain",
			Metadata:          map[string]any{},
		}, nil
	}

	switch c.strategy {
	case Sequential:
		return c.executeSequential(ctx, input)
	case Parallel:
		return c.executeParallel(ctx, input)
	case StopOnFirstPass:
		return c.executeStopOnFirstPass(ctx, input)
	default:
		return nil, fmt.Errorf("unknown strategy: %v", c.strategy)
	}
}

// executeSequential runs guardrails one by one, stopping on first failure.
func (c *Chain) executeSequential(ctx context.Context, input string) (*Result, error) {
	results := make([]*Result, 0, len(c.guardrails))

	for _, g := range c.guardrails {
		result, err := g.Func(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("guardrail %s failed: %w", g.Name, err)
		}

		results = append(results, result)

		// Short-circuit on first failure
		if !result.Passed {
			return &Result{
				Passed:            false,
				TripwireTriggered: result.TripwireTriggered,
				Message:           fmt.Sprintf("guardrail %s failed: %s", g.Name, result.Message),
				Metadata: map[string]any{
					"failed_guardrail": g.Name,
					"strategy":         "Sequential",
					"results":          results,
				},
			}, nil
		}
	}

	// All passed
	return &Result{
		Passed:            true,
		TripwireTriggered: false,
		Message:           fmt.Sprintf("all %d guardrails passed", len(c.guardrails)),
		Metadata: map[string]any{
			"strategy": "Sequential",
			"results":  results,
		},
	}, nil
}

// executeParallel runs all guardrails concurrently and aggregates results.
func (c *Chain) executeParallel(ctx context.Context, input string) (*Result, error) {
	type guardResult struct {
		name   string
		result *Result
		err    error
	}

	resultChan := make(chan guardResult, len(c.guardrails))
	var wg sync.WaitGroup

	// Launch all guardrails concurrently
	for _, g := range c.guardrails {
		wg.Add(1)
		go func(guard *Guardrail) {
			defer wg.Done()
			result, err := guard.Func(ctx, input)
			resultChan <- guardResult{
				name:   guard.Name,
				result: result,
				err:    err,
			}
		}(g)
	}

	// Wait for all to complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	results := make([]*Result, 0, len(c.guardrails))
	failures := make([]string, 0)
	hasError := false
	hasTripwire := false

	for gr := range resultChan {
		if gr.err != nil {
			return nil, fmt.Errorf("guardrail %s failed: %w", gr.name, gr.err)
		}

		results = append(results, gr.result)

		if !gr.result.Passed {
			hasError = true
			failures = append(failures, gr.name)
		}

		if gr.result.TripwireTriggered {
			hasTripwire = true
		}
	}

	// Aggregate results
	if hasError {
		return &Result{
			Passed:            false,
			TripwireTriggered: hasTripwire,
			Message:           fmt.Sprintf("%d/%d guardrails failed: %v", len(failures), len(c.guardrails), failures),
			Metadata: map[string]any{
				"strategy":          "Parallel",
				"failed_count":      len(failures),
				"failed_guardrails": failures,
				"results":           results,
			},
		}, nil
	}

	return &Result{
		Passed:            true,
		TripwireTriggered: false,
		Message:           fmt.Sprintf("all %d guardrails passed", len(c.guardrails)),
		Metadata: map[string]any{
			"strategy": "Parallel",
			"results":  results,
		},
	}, nil
}

// executeStopOnFirstPass runs guardrails until one passes (OR logic).
func (c *Chain) executeStopOnFirstPass(ctx context.Context, input string) (*Result, error) {
	results := make([]*Result, 0, len(c.guardrails))

	for _, g := range c.guardrails {
		result, err := g.Func(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("guardrail %s failed: %w", g.Name, err)
		}

		results = append(results, result)

		// Stop on first pass
		if result.Passed {
			return &Result{
				Passed:            true,
				TripwireTriggered: false,
				Message:           fmt.Sprintf("guardrail %s passed", g.Name),
				Metadata: map[string]any{
					"strategy":         "StopOnFirstPass",
					"passed_guardrail": g.Name,
					"results":          results,
				},
			}, nil
		}
	}

	// All failed
	return &Result{
		Passed:            false,
		TripwireTriggered: false,
		Message:           fmt.Sprintf("all %d guardrails failed", len(c.guardrails)),
		Metadata: map[string]any{
			"strategy": "StopOnFirstPass",
			"results":  results,
		},
	}, nil
}
