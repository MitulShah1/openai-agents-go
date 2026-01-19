package agents

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"time"
)

// RateLimitError represents an API rate limit error with retry information.
type RateLimitError struct {
	// Message describes the rate limit error
	Message string
	// RetryAfter indicates when the request can be retried
	RetryAfter time.Duration
	// Limit is the rate limit threshold that was exceeded
	Limit int
	// Remaining is the number of requests remaining (usually 0)
	Remaining int
}

func (e *RateLimitError) Error() string {
	if e.RetryAfter > 0 {
		return fmt.Sprintf("rate limit exceeded: %s (retry after %v)", e.Message, e.RetryAfter)
	}
	return fmt.Sprintf("rate limit exceeded: %s", e.Message)
}

// TimeoutError represents a request timeout error.
type TimeoutError struct {
	// Message describes the timeout
	Message string
	// Duration is how long the request took before timing out
	Duration time.Duration
	// Timeout is the configured timeout value
	Timeout time.Duration
}

func (e *TimeoutError) Error() string {
	return fmt.Sprintf("request timeout after %v (limit: %v): %s", e.Duration, e.Timeout, e.Message)
}

// NetworkError represents a network-level failure.
type NetworkError struct {
	// Message describes the network error
	Message string
	// Cause is the underlying error
	Cause error
	// Retryable indicates if the error is transient and can be retried
	Retryable bool
}

func (e *NetworkError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("network error: %s: %v", e.Message, e.Cause)
	}
	return fmt.Sprintf("network error: %s", e.Message)
}

func (e *NetworkError) Unwrap() error {
	return e.Cause
}

// ErrorContext wraps an error with contextual information about the agent execution.
type ErrorContext struct {
	// AgentName is the name of the agent that encountered the error
	AgentName string
	// StepNumber is the execution step where the error occurred
	StepNumber int
	// SessionID is the session identifier (if applicable)
	SessionID string
	// Err is the underlying error
	Err error
}

func (e *ErrorContext) Error() string {
	context := fmt.Sprintf("[agent=%s, step=%d", e.AgentName, e.StepNumber)
	if e.SessionID != "" {
		context += fmt.Sprintf(", session=%s", e.SessionID)
	}
	context += "]"
	return fmt.Sprintf("%s %s", context, e.Err.Error())
}

func (e *ErrorContext) Unwrap() error {
	return e.Err
}

// WrapError wraps an error with execution context.
func WrapError(err error, agentName string, stepNumber int, sessionID string) error {
	if err == nil {
		return nil
	}
	return &ErrorContext{
		AgentName:  agentName,
		StepNumber: stepNumber,
		SessionID:  sessionID,
		Err:        err,
	}
}

// BackoffStrategy defines the interface for retry backoff strategies.
type BackoffStrategy interface {
	// NextDelay returns the next delay duration for the given attempt number (0-indexed).
	NextDelay(attempt int) time.Duration
}

// FixedBackoff implements a constant delay between retries.
type FixedBackoff struct {
	Delay time.Duration
}

// NextDelay returns the configured fixed delay for any attempt number.
func (b *FixedBackoff) NextDelay(_ int) time.Duration {
	return b.Delay
}

// LinearBackoff implements a linear increase in delay.
type LinearBackoff struct {
	// Initial is the starting delay
	Initial time.Duration
	// Increment is added to the delay for each attempt
	Increment time.Duration
	// MaxDelay caps the maximum delay
	MaxDelay time.Duration
}

// NextDelay returns a linearly increasing delay based on the attempt number.
func (b *LinearBackoff) NextDelay(attempt int) time.Duration {
	delay := b.Initial + time.Duration(attempt)*b.Increment
	if b.MaxDelay > 0 && delay > b.MaxDelay {
		return b.MaxDelay
	}
	return delay
}

// ExponentialBackoff implements exponential backoff with optional jitter.
type ExponentialBackoff struct {
	// Initial is the starting delay
	Initial time.Duration
	// Multiplier is the exponential growth factor (typically 2.0)
	Multiplier float64
	// MaxDelay caps the maximum delay
	MaxDelay time.Duration
	// Jitter adds randomness to prevent thundering herd (0.0 to 1.0)
	Jitter float64
}

// NextDelay returns an exponentially increasing delay with optional jitter.
func (b *ExponentialBackoff) NextDelay(attempt int) time.Duration {
	// Calculate exponential delay: initial * (multiplier ^ attempt)
	delay := float64(b.Initial) * math.Pow(b.Multiplier, float64(attempt))

	// Apply max delay cap
	if b.MaxDelay > 0 && time.Duration(delay) > b.MaxDelay {
		delay = float64(b.MaxDelay)
	}

	// Apply jitter if configured
	if b.Jitter > 0 {
		jitterAmount := delay * b.Jitter
		// Use crypto/rand for secure randomness
		var randBytes [8]byte
		if _, err := rand.Read(randBytes[:]); err == nil {
			randFloat := float64(binary.BigEndian.Uint64(randBytes[:])) / float64(^uint64(0))
			delay = delay - jitterAmount + (randFloat * 2 * jitterAmount)
		}
	}

	return time.Duration(delay)
}

// CustomBackoff allows users to define their own backoff function.
type CustomBackoff struct {
	// Fn is a user-defined function that returns the delay for a given attempt
	Fn func(attempt int) time.Duration
}

// NextDelay calls the user-defined backoff function with the attempt number.
func (b *CustomBackoff) NextDelay(attempt int) time.Duration {
	return b.Fn(attempt)
}

// RetryConfig configures retry behavior for transient failures.
type RetryConfig struct {
	// MaxAttempts is the maximum number of retry attempts (0 = no retries)
	MaxAttempts int
	// Backoff is the backoff strategy to use
	Backoff BackoffStrategy
	// RetryableErrors is a function that determines if an error should be retried
	RetryableErrors func(error) bool
}

// DefaultRetryConfig returns a sensible default retry configuration.
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts: 3,
		Backoff: &ExponentialBackoff{
			Initial:    100 * time.Millisecond,
			Multiplier: 2.0,
			MaxDelay:   10 * time.Second,
			Jitter:     0.1,
		},
		RetryableErrors: func(err error) bool {
			// Retry network errors and rate limits by default
			var netErr *NetworkError
			var rateLimitErr *RateLimitError
			var timeoutErr *TimeoutError

			if errors.As(err, &netErr) && netErr.Retryable {
				return true
			}
			if errors.As(err, &rateLimitErr) {
				return true
			}
			if errors.As(err, &timeoutErr) {
				return true
			}

			return false
		},
	}
}

// RetryWithBackoff executes a function with retry logic using the configured backoff strategy.
func RetryWithBackoff(ctx context.Context, config *RetryConfig, fn func() error) error {
	if config == nil {
		config = DefaultRetryConfig()
	}

	var lastErr error
	for attempt := 0; attempt <= config.MaxAttempts; attempt++ {
		// Execute the function
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if we should retry this error
		if config.RetryableErrors != nil && !config.RetryableErrors(err) {
			return err
		}

		// Don't sleep after the last attempt
		if attempt == config.MaxAttempts {
			break
		}

		// Calculate backoff delay
		delay := config.Backoff.NextDelay(attempt)

		// Handle rate limit retry-after if present
		var rateLimitErr *RateLimitError
		if errors.As(err, &rateLimitErr) && rateLimitErr.RetryAfter > 0 {
			delay = rateLimitErr.RetryAfter
		}

		// Wait with context cancellation support
		select {
		case <-ctx.Done():
			return fmt.Errorf("retry cancelled: %w", ctx.Err())
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	return fmt.Errorf("max retry attempts (%d) exceeded: %w", config.MaxAttempts, lastErr)
}
