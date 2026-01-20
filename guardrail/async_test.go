package guardrail

import (
	"context"
	"errors"
	"testing"
	"time"
)

// Mock slow guardrail for timeout testing
func slowGuardrail(delay time.Duration, shouldPass bool) *Guardrail {
	return NewGuardrail("slow_guardrail", func(ctx context.Context, _ string) (*Result, error) {
		select {
		case <-time.After(delay):
			if shouldPass {
				return &Result{Passed: true}, nil
			}
			return &Result{Passed: false, Message: "validation failed"}, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	})
}

func TestWithTimeout(t *testing.T) {
	tests := []struct {
		name          string
		guardrail     *Guardrail
		timeout       time.Duration
		input         string
		expectError   bool
		expectTimeout bool
	}{
		{
			name:          "completes within timeout",
			guardrail:     slowGuardrail(50*time.Millisecond, true),
			timeout:       200 * time.Millisecond,
			input:         "test input",
			expectError:   false,
			expectTimeout: false,
		},
		{
			name:          "exceeds timeout",
			guardrail:     slowGuardrail(200*time.Millisecond, true),
			timeout:       50 * time.Millisecond,
			input:         "test input",
			expectError:   true,
			expectTimeout: true,
		},
		{
			name:          "fast failing guardrail",
			guardrail:     slowGuardrail(10*time.Millisecond, false),
			timeout:       100 * time.Millisecond,
			input:         "test input",
			expectError:   false,
			expectTimeout: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrapped := WithTimeout(tt.guardrail, tt.timeout)
			result, err := wrapped.Func(context.Background(), tt.input)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				if tt.expectTimeout && !errors.Is(err, context.DeadlineExceeded) {
					t.Errorf("expected timeout error, got: %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result == nil {
					t.Error("expected result but got nil")
				}
			}
		})
	}
}

func TestWithTimeoutGraceful(t *testing.T) {
	tests := []struct {
		name       string
		guardrail  *Guardrail
		timeout    time.Duration
		input      string
		expectPass bool
		expectWarn bool
	}{
		{
			name:       "completes within timeout",
			guardrail:  slowGuardrail(50*time.Millisecond, true),
			timeout:    200 * time.Millisecond,
			input:      "test input",
			expectPass: true,
			expectWarn: false,
		},
		{
			name:       "timeout triggers graceful pass",
			guardrail:  slowGuardrail(200*time.Millisecond, true),
			timeout:    50 * time.Millisecond,
			input:      "test input",
			expectPass: true,
			expectWarn: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrapped := WithTimeoutGraceful(tt.guardrail, tt.timeout)
			result, err := wrapped.Func(context.Background(), tt.input)

			if err != nil {
				t.Errorf("graceful mode should not return error: %v", err)
			}

			if result == nil {
				t.Fatal("expected result but got nil")
			}

			if tt.expectPass && !result.Passed {
				t.Error("expected validation to pass")
			}

			if tt.expectWarn && result.Message == "" {
				t.Error("expected warning message")
			}
		})
	}
}

func TestWithContext(t *testing.T) {
	t.Run("context cancellation propagates", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		slowGuard := slowGuardrail(200*time.Millisecond, true)
		wrapped := WithContext(slowGuard)

		// Cancel context during validation
		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel()
		}()

		_, err := wrapped.Func(ctx, "test input")

		if err == nil {
			t.Error("expected error due to context cancellation")
		}

		if !errors.Is(err, context.Canceled) {
			t.Errorf("expected context.Canceled error, got: %v", err)
		}
	})

	t.Run("normal completion without cancellation", func(t *testing.T) {
		ctx := context.Background()

		fastGuard := slowGuardrail(10*time.Millisecond, true)
		wrapped := WithContext(fastGuard)

		result, err := wrapped.Func(ctx, "test input")

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if result == nil || !result.Passed {
			t.Error("expected valid result")
		}
	})
}

func TestTimeoutChaining(t *testing.T) {
	t.Run("timeout on multiple guardrails in chain", func(t *testing.T) {
		fast := slowGuardrail(10*time.Millisecond, true)
		slow := slowGuardrail(200*time.Millisecond, true)

		chain := NewChain().
			Add(WithTimeout(fast, 50*time.Millisecond)).
			Add(WithTimeout(slow, 50*time.Millisecond)).
			WithStrategy(Sequential).
			Build()

		_, err := chain.Func(context.Background(), "test")

		if err == nil {
			t.Error("expected timeout error in chain")
		}
	})

	t.Run("graceful timeout in chain allows continuation", func(t *testing.T) {
		fast := slowGuardrail(10*time.Millisecond, true)
		slow := slowGuardrail(200*time.Millisecond, true)

		chain := NewChain().
			Add(WithTimeout(fast, 50*time.Millisecond)).
			Add(WithTimeoutGraceful(slow, 50*time.Millisecond)).
			WithStrategy(Sequential).
			Build()

		result, err := chain.Func(context.Background(), "test")

		if err != nil {
			t.Errorf("graceful timeout should allow continuation: %v", err)
		}

		if result == nil || !result.Passed {
			t.Error("expected valid result despite timeout")
		}
	})
}

func TestContextWithDeadline(t *testing.T) {
	t.Run("respects context deadline", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		slow := slowGuardrail(200*time.Millisecond, true)
		wrapped := WithContext(slow)

		start := time.Now()
		_, err := wrapped.Func(ctx, "test input")
		duration := time.Since(start)

		if err == nil {
			t.Error("expected deadline exceeded error")
		}

		if duration > 100*time.Millisecond {
			t.Errorf("validation should have stopped at deadline, took %v", duration)
		}
	})
}
