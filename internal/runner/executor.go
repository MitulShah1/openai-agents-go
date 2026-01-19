package runner

import (
	"context"
	"fmt"
	"time"
)

// Executor manages the main execution loop
type Executor struct {
	maxTurns int
	timeout  time.Duration
}

// NewExecutor creates a new executor with the given configuration
func NewExecutor(maxTurns int, timeout time.Duration) *Executor {
	return &Executor{
		maxTurns: maxTurns,
		timeout:  timeout,
	}
}

// CheckContext verifies that the context is still valid (not cancelled or timed out)
// Returns an appropriate error if the context is done
func CheckContext(ctx context.Context) error {
	select {
	case <-ctx.Done():
		err := ctx.Err()
		if err == context.DeadlineExceeded {
			// Return a more specific timeout error
			return fmt.Errorf("%w: %v", ErrTimeout, err)
		}
		return err
	default:
		return nil
	}
}

// ShouldContinueExecution checks if execution should continue
// Returns false if max turns exceeded or context is done
func (e *Executor) ShouldContinueExecution(ctx context.Context, turnCount int) (bool, error) {
	// Check max turns
	if e.maxTurns > 0 && turnCount >= e.maxTurns {
		return false, fmt.Errorf("%w: %d", ErrMaxTurnsExceeded, e.maxTurns)
	}

	// Check context cancellation/timeout
	if err := CheckContext(ctx); err != nil {
		return false, err
	}

	return true, nil
}

// ExtractFinalOutput extracts the final output from the last message
func ExtractFinalOutput(_ any) string {
	// This will need to be implemented with the actual message type
	// For now, return a placeholder
	return ""
}

// ApplyTimeout applies a timeout to the context if configured
func (e *Executor) ApplyTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if e.timeout > 0 {
		return context.WithTimeout(ctx, e.timeout)
	}
	// Return a no-op cancel function
	return ctx, func() {}
}
