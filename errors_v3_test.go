package agents

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRateLimitError(t *testing.T) {
	t.Run("with retry after", func(t *testing.T) {
		err := &RateLimitError{
			Message:    "too many requests",
			RetryAfter: 5 * time.Second,
			Limit:      100,
			Remaining:  0,
		}

		expected := "rate limit exceeded: too many requests (retry after 5s)"
		if err.Error() != expected {
			t.Errorf("expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("without retry after", func(t *testing.T) {
		err := &RateLimitError{
			Message:   "quota exceeded",
			Limit:     1000,
			Remaining: 0,
		}

		expected := "rate limit exceeded: quota exceeded"
		if err.Error() != expected {
			t.Errorf("expected %q, got %q", expected, err.Error())
		}
	})
}

func TestTimeoutError(t *testing.T) {
	err := &TimeoutError{
		Message:  "API call timed out",
		Duration: 35 * time.Second,
		Timeout:  30 * time.Second,
	}

	if !errors.Is(err, err) {
		t.Error("TimeoutError should match itself")
	}

	expected := "request timeout after 35s (limit: 30s): API call timed out"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestNetworkError(t *testing.T) {
	t.Run("with cause", func(t *testing.T) {
		cause := errors.New("connection refused")
		err := &NetworkError{
			Message:   "failed to connect",
			Cause:     cause,
			Retryable: true,
		}

		if !errors.Is(err, cause) {
			t.Error("NetworkError should unwrap to cause")
		}
	})

	t.Run("without cause", func(t *testing.T) {
		err := &NetworkError{
			Message:   "dns resolution failed",
			Retryable: false,
		}

		expected := "network error: dns resolution failed"
		if err.Error() != expected {
			t.Errorf("expected %q, got %q", expected, err.Error())
		}
	})
}

func TestErrorContext(t *testing.T) {
	t.Run("with session ID", func(t *testing.T) {
		baseErr := errors.New("tool execution failed")
		err := &ErrorContext{
			AgentName:  "assistant",
			StepNumber: 5,
			SessionID:  "sess_123",
			Err:        baseErr,
		}

		expected := "[agent=assistant, step=5, session=sess_123] tool execution failed"
		if err.Error() != expected {
			t.Errorf("expected %q, got %q", expected, err.Error())
		}

		if !errors.Is(err, baseErr) {
			t.Error("ErrorContext should unwrap to base error")
		}
	})

	t.Run("without session ID", func(t *testing.T) {
		baseErr := errors.New("validation failed")
		err := WrapError(baseErr, "validator", 1, "")

		expected := "[agent=validator, step=1] validation failed"
		if err.Error() != expected {
			t.Errorf("expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("wrap nil error", func(t *testing.T) {
		err := WrapError(nil, "test", 1, "")
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})
}

func TestFixedBackoff(t *testing.T) {
	backoff := &FixedBackoff{Delay: 1 * time.Second}

	for i := 0; i < 5; i++ {
		delay := backoff.NextDelay(i)
		if delay != 1*time.Second {
			t.Errorf("attempt %d: expected 1s, got %v", i, delay)
		}
	}
}

func TestLinearBackoff(t *testing.T) {
	backoff := &LinearBackoff{
		Initial:   100 * time.Millisecond,
		Increment: 50 * time.Millisecond,
		MaxDelay:  500 * time.Millisecond,
	}

	expected := []time.Duration{
		100 * time.Millisecond, // attempt 0
		150 * time.Millisecond, // attempt 1
		200 * time.Millisecond, // attempt 2
		250 * time.Millisecond, // attempt 3
		300 * time.Millisecond, // attempt 4
		350 * time.Millisecond, // attempt 5
		400 * time.Millisecond, // attempt 6
		450 * time.Millisecond, // attempt 7
		500 * time.Millisecond, // attempt 8 (capped)
		500 * time.Millisecond, // attempt 9 (capped)
	}

	for i, exp := range expected {
		delay := backoff.NextDelay(i)
		if delay != exp {
			t.Errorf("attempt %d: expected %v, got %v", i, exp, delay)
		}
	}
}

func TestExponentialBackoff(t *testing.T) {
	t.Run("without jitter", func(t *testing.T) {
		backoff := &ExponentialBackoff{
			Initial:    100 * time.Millisecond,
			Multiplier: 2.0,
			MaxDelay:   10 * time.Second,
			Jitter:     0.0,
		}

		expected := []time.Duration{
			100 * time.Millisecond,  // 100 * 2^0
			200 * time.Millisecond,  // 100 * 2^1
			400 * time.Millisecond,  // 100 * 2^2
			800 * time.Millisecond,  // 100 * 2^3
			1600 * time.Millisecond, // 100 * 2^4
		}

		for i, exp := range expected {
			delay := backoff.NextDelay(i)
			if delay != exp {
				t.Errorf("attempt %d: expected %v, got %v", i, exp, delay)
			}
		}
	})

	t.Run("with max delay", func(t *testing.T) {
		backoff := &ExponentialBackoff{
			Initial:    1 * time.Second,
			Multiplier: 2.0,
			MaxDelay:   5 * time.Second,
			Jitter:     0.0,
		}

		// Should cap at 5 seconds
		delay := backoff.NextDelay(10)
		if delay != 5*time.Second {
			t.Errorf("expected 5s (capped), got %v", delay)
		}
	})

	t.Run("with jitter", func(t *testing.T) {
		backoff := &ExponentialBackoff{
			Initial:    1 * time.Second,
			Multiplier: 2.0,
			MaxDelay:   0,
			Jitter:     0.1, // 10% jitter
		}

		// With jitter, the delay should be within bounds
		delay := backoff.NextDelay(3)                           // 1s * 2^3 = 8s
		minDelay := time.Duration(float64(8*time.Second) * 0.9) // 8s - 10%
		maxDelay := time.Duration(float64(8*time.Second) * 1.1) // 8s + 10%

		if delay < minDelay || delay > maxDelay {
			t.Errorf("delay %v not in range [%v, %v]", delay, minDelay, maxDelay)
		}
	})
}

func TestCustomBackoff(t *testing.T) {
	// Custom function: always return attempt number in seconds
	backoff := &CustomBackoff{
		Fn: func(attempt int) time.Duration {
			return time.Duration(attempt) * time.Second
		},
	}

	for i := 0; i < 5; i++ {
		delay := backoff.NextDelay(i)
		expected := time.Duration(i) * time.Second
		if delay != expected {
			t.Errorf("attempt %d: expected %v, got %v", i, expected, delay)
		}
	}
}

func TestRetryWithBackoff(t *testing.T) {
	t.Run("succeeds on first attempt", func(t *testing.T) {
		attempts := 0
		fn := func() error {
			attempts++
			return nil
		}

		config := &RetryConfig{
			MaxAttempts: 3,
			Backoff:     &FixedBackoff{Delay: 10 * time.Millisecond},
		}

		err := RetryWithBackoff(context.Background(), config, fn)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if attempts != 1 {
			t.Errorf("expected 1 attempt, got %d", attempts)
		}
	})

	t.Run("succeeds after retries", func(t *testing.T) {
		attempts := 0
		fn := func() error {
			attempts++
			if attempts < 3 {
				return &NetworkError{Message: "temporary failure", Retryable: true}
			}
			return nil
		}

		config := &RetryConfig{
			MaxAttempts: 5,
			Backoff:     &FixedBackoff{Delay: 1 * time.Millisecond},
			RetryableErrors: func(err error) bool {
				var netErr *NetworkError
				return errors.As(err, &netErr) && netErr.Retryable
			},
		}

		err := RetryWithBackoff(context.Background(), config, fn)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if attempts != 3 {
			t.Errorf("expected 3 attempts, got %d", attempts)
		}
	})

	t.Run("fails after max attempts", func(t *testing.T) {
		attempts := 0
		fn := func() error {
			attempts++
			return &NetworkError{Message: "persistent failure", Retryable: true}
		}

		config := &RetryConfig{
			MaxAttempts: 2,
			Backoff:     &FixedBackoff{Delay: 1 * time.Millisecond},
			RetryableErrors: func(err error) bool {
				var netErr *NetworkError
				return errors.As(err, &netErr) && netErr.Retryable
			},
		}

		err := RetryWithBackoff(context.Background(), config, fn)
		if err == nil {
			t.Error("expected error, got nil")
		}
		if attempts != 3 { // initial + 2 retries
			t.Errorf("expected 3 attempts, got %d", attempts)
		}
	})

	t.Run("non-retryable error fails immediately", func(t *testing.T) {
		attempts := 0
		fn := func() error {
			attempts++
			return errors.New("non-retryable error")
		}

		config := DefaultRetryConfig()

		err := RetryWithBackoff(context.Background(), config, fn)
		if err == nil {
			t.Error("expected error, got nil")
		}
		if attempts != 1 {
			t.Errorf("expected 1 attempt, got %d", attempts)
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		attempts := 0
		fn := func() error {
			attempts++
			if attempts == 2 {
				cancel() // Cancel after first retry
			}
			return &NetworkError{Message: "failure", Retryable: true}
		}

		config := &RetryConfig{
			MaxAttempts: 10,
			Backoff:     &FixedBackoff{Delay: 10 * time.Millisecond},
			RetryableErrors: func(err error) bool {
				var netErr *NetworkError
				return errors.As(err, &netErr) && netErr.Retryable
			},
		}

		err := RetryWithBackoff(ctx, config, fn)
		if err == nil {
			t.Error("expected error, got nil")
		}
		if !errors.Is(err, context.Canceled) {
			t.Errorf("expected context.Canceled, got %v", err)
		}
	})

	t.Run("rate limit retry after", func(t *testing.T) {
		attempts := 0
		start := time.Now()
		fn := func() error {
			attempts++
			if attempts == 1 {
				return &RateLimitError{
					Message:    "rate limited",
					RetryAfter: 50 * time.Millisecond,
				}
			}
			return nil
		}

		config := DefaultRetryConfig()

		err := RetryWithBackoff(context.Background(), config, fn)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		elapsed := time.Since(start)
		if elapsed < 50*time.Millisecond {
			t.Errorf("expected at least 50ms delay, got %v", elapsed)
		}
	})
}
