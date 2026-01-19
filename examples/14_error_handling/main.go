// Example demonstrating error handling with retry strategies
package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	agents "github.com/MitulShah1/openai-agents-go"
)

func main() {
	fmt.Println("=== Error Handling & Retry Strategies Demo ===")
	fmt.Println()

	// 1. Fixed Backoff Strategy
	fmt.Println("1️⃣  Fixed Backoff (500ms delay)")
	fixedConfig := &agents.RetryConfig{
		MaxAttempts: 3,
		Backoff: &agents.FixedBackoff{
			Delay: 500 * time.Millisecond,
		},
	}

	attempt := 0
	err := agents.RetryWithBackoff(context.Background(), fixedConfig, func() error {
		attempt++
		fmt.Printf("   Attempt %d at %s\n", attempt, time.Now().Format("15:04:05.000"))
		if attempt < 3 {
			return &agents.NetworkError{Message: "connection refused", Retryable: true}
		}
		return nil
	})
	fmt.Printf("   Result: %v\n\n", err)

	// 2. Linear Backoff Strategy
	fmt.Println("2️⃣  Linear Backoff (100ms initial, +200ms per attempt)")
	linearConfig := &agents.RetryConfig{
		MaxAttempts: 4,
		Backoff: &agents.LinearBackoff{
			Initial:   100 * time.Millisecond,
			Increment: 200 * time.Millisecond,
			MaxDelay:  1 * time.Second,
		},
	}

	attempt = 0
	err = agents.RetryWithBackoff(context.Background(), linearConfig, func() error {
		attempt++
		fmt.Printf("   Attempt %d at %s\n", attempt, time.Now().Format("15:04:05.000"))
		if attempt < 4 {
			return &agents.TimeoutError{Message: "query timeout", Duration: time.Second}
		}
		return nil
	})
	fmt.Printf("   Result: %v\n\n", err)

	// 3. Exponential Backoff with Jitter
	fmt.Println("3️⃣  Exponential Backoff (100ms initial, 2x multiplier, 20% jitter)")
	expConfig := &agents.RetryConfig{
		MaxAttempts: 5,
		Backoff: &agents.ExponentialBackoff{
			Initial:    100 * time.Millisecond,
			Multiplier: 2.0,
			MaxDelay:   2 * time.Second,
			Jitter:     0.2, // 20% randomization
		},
	}

	attempt = 0
	startTime := time.Now()
	err = agents.RetryWithBackoff(context.Background(), expConfig, func() error {
		attempt++
		elapsed := time.Since(startTime)
		fmt.Printf("   Attempt %d at +%dms\n", attempt, elapsed.Milliseconds())
		if attempt < 4 {
			return &agents.NetworkError{Message: "temporary failure", Retryable: true}
		}
		return nil
	})
	fmt.Printf("   Result: %v\n\n", err)

	// 4. Rate Limit with RetryAfter
	fmt.Println("4️⃣  Rate Limit Error (respects RetryAfter)")
	rateLimitConfig := &agents.RetryConfig{
		MaxAttempts: 3,
		Backoff: &agents.FixedBackoff{
			Delay: 100 * time.Millisecond,
		},
	}

	attempt = 0
	err = agents.RetryWithBackoff(context.Background(), rateLimitConfig, func() error {
		attempt++
		fmt.Printf("   Attempt %d at %s\n", attempt, time.Now().Format("15:04:05.000"))
		if attempt == 1 {
			// First attempt hits rate limit
			return &agents.RateLimitError{
				Message:    "Too many requests",
				RetryAfter: 800 * time.Millisecond,
				Limit:      100,
				Remaining:  0,
			}
		}
		return nil
	})
	fmt.Printf("   Result: %v (retried after 800ms)\n\n", err)

	// 5. Custom Backoff Strategy
	fmt.Println("5️⃣  Custom Backoff (Fibonacci-like: 100, 200, 300, 500, 800ms)")
	customConfig := &agents.RetryConfig{
		MaxAttempts: 5,
		Backoff: &agents.CustomBackoff{
			Fn: func(attempt int) time.Duration {
				// Fibonacci-like sequence
				delays := []int{100, 200, 300, 500, 800}
				if attempt < len(delays) {
					return time.Duration(delays[attempt]) * time.Millisecond
				}
				return 1 * time.Second
			},
		},
	}

	attempt = 0
	startTime = time.Now()
	err = agents.RetryWithBackoff(context.Background(), customConfig, func() error {
		attempt++
		elapsed := time.Since(startTime)
		fmt.Printf("   Attempt %d at +%dms\n", attempt, elapsed.Milliseconds())
		if attempt < 5 {
			return &agents.NetworkError{Message: "retry", Retryable: true}
		}
		return nil
	})
	fmt.Printf("   Result: %v\n\n", err)

	// 6. Non-retryable Error
	fmt.Println("6️⃣  Non-Retryable Error (fails immediately)")
	err = agents.RetryWithBackoff(context.Background(), fixedConfig, func() error {
		fmt.Println("   Single attempt (non-retryable error)")
		return errors.New("invalid input") // Regular errors are not retried
	})
	fmt.Printf("   Result: %v (no retries)\n\n", err)

	// 7. Context Cancellation
	fmt.Println("7️⃣  Context Cancellation (respects timeout)")
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	attempt = 0
	err = agents.RetryWithBackoff(ctx, linearConfig, func() error {
		attempt++
		fmt.Printf("   Attempt %d\n", attempt)
		time.Sleep(300 * time.Millisecond)
		return &agents.TimeoutError{Message: "slow operation", Duration: time.Second}
	})
	fmt.Printf("   Result: %v (context canceled after %d attempts)\n\n", err, attempt)

	fmt.Println("✨ Error handling demonstration complete!")
	fmt.Println("\nKey takeaways:")
	fmt.Println("  🔄 Different backoff strategies for different scenarios")
	fmt.Println("  ⏱️  Rate limit errors respect RetryAfter")
	fmt.Println("  🚫 Non-retryable errors fail immediately")
	fmt.Println("  ⏰ Context cancellation stops retries")
	fmt.Println("  🎲 Jitter prevents thundering herd")
}
