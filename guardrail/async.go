package guardrail

import (
	"context"
	"fmt"
	"time"
)

// WithTimeout wraps a guardrail with a timeout.
// If the guardrail execution exceeds the timeout, an error is returned.
func WithTimeout(g *Guardrail, timeout time.Duration) *Guardrail {
	return &Guardrail{
		Name:          fmt.Sprintf("%s_with_timeout", g.Name),
		RunInParallel: g.RunInParallel,
		Func: func(ctx context.Context, input string) (*Result, error) {
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			resultChan := make(chan *Result, 1)
			errChan := make(chan error, 1)

			go func() {
				result, err := g.Func(ctx, input)
				if err != nil {
					errChan <- err
					return
				}
				resultChan <- result
			}()

			select {
			case result := <-resultChan:
				return result, nil
			case err := <-errChan:
				return nil, err
			case <-ctx.Done():
				return nil, fmt.Errorf("guardrail %s timed out after %v", g.Name, timeout)
			}
		},
	}
}

// WithContext makes a guardrail respect context cancellation.
// This is useful when you want guardrails to stop gracefully on request.
func WithContext(g *Guardrail) *Guardrail {
	return &Guardrail{
		Name:          g.Name,
		RunInParallel: g.RunInParallel,
		Func: func(ctx context.Context, input string) (*Result, error) {
			// Check context before execution
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("guardrail %s cancelled: %w", g.Name, ctx.Err())
			default:
			}

			// Execute with context
			resultChan := make(chan *Result, 1)
			errChan := make(chan error, 1)

			go func() {
				result, err := g.Func(ctx, input)
				if err != nil {
					errChan <- err
					return
				}
				resultChan <- result
			}()

			select {
			case result := <-resultChan:
				return result, nil
			case err := <-errChan:
				return nil, err
			case <-ctx.Done():
				return nil, fmt.Errorf("guardrail %s cancelled: %w", g.Name, ctx.Err())
			}
		},
	}
}

// WithTimeoutGraceful wraps a guardrail with a timeout and allows graceful degradation.
// If the guardrail times out, it returns a passing result with a warning instead of an error.
func WithTimeoutGraceful(g *Guardrail, timeout time.Duration) *Guardrail {
	return &Guardrail{
		Name:          fmt.Sprintf("%s_with_graceful_timeout", g.Name),
		RunInParallel: g.RunInParallel,
		Func: func(ctx context.Context, input string) (*Result, error) {
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			resultChan := make(chan *Result, 1)
			errChan := make(chan error, 1)

			go func() {
				result, err := g.Func(ctx, input)
				if err != nil {
					errChan <- err
					return
				}
				resultChan <- result
			}()

			select {
			case result := <-resultChan:
				return result, nil
			case err := <-errChan:
				return nil, err
			case <-ctx.Done():
				// Graceful degradation: pass with warning
				return &Result{
					Passed:            true,
					TripwireTriggered: false,
					Message:           fmt.Sprintf("guardrail %s timed out after %v (gracefully degraded)", g.Name, timeout),
					Metadata: map[string]any{
						"timeout":  true,
						"duration": timeout,
						"degraded": true,
					},
				}, nil
			}
		},
	}
}
