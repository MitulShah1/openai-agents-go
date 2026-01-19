package builtin

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestTokenBucket_Consume(t *testing.T) {
	t.Run("allows requests up to burst", func(t *testing.T) {
		tb := &TokenBucket{
			tokens:     5.0,
			capacity:   5.0,
			refillRate: 1.0, // 1 token per second
			lastRefill: time.Now(),
		}

		// Should allow 5 requests (burst size)
		for i := 0; i < 5; i++ {
			if !tb.consume() {
				t.Errorf("request %d should be allowed", i+1)
			}
		}

		// 6th request should be denied
		if tb.consume() {
			t.Error("request beyond burst should be denied")
		}
	})

	t.Run("refills tokens over time", func(t *testing.T) {
		tb := &TokenBucket{
			tokens:     0.0,
			capacity:   10.0,
			refillRate: 10.0,                             // 10 tokens per second
			lastRefill: time.Now().Add(-1 * time.Second), // 1 second ago
		}

		// After 1 second, should have 10 tokens refilled
		if !tb.consume() {
			t.Error("should allow request after refill")
		}
	})

	t.Run("caps at max capacity", func(t *testing.T) {
		tb := &TokenBucket{
			tokens:     5.0,
			capacity:   10.0,
			refillRate: 100.0,                             // Very high refill rate
			lastRefill: time.Now().Add(-10 * time.Second), // 10 seconds ago
		}

		// Even with 1000 tokens refilled, should cap at 10
		tb.consume() // Trigger refill

		// Lock to check internal state
		tb.mu.Lock()
		if tb.tokens > tb.capacity {
			t.Errorf("tokens %f should not exceed capacity %f", tb.tokens, tb.capacity)
		}
		tb.mu.Unlock()
	})
}

func TestInMemoryRateLimiter(t *testing.T) {
	t.Run("creates bucket on first request", func(t *testing.T) {
		limiter := NewInMemoryRateLimiter(10, 5, time.Second)

		allowed, err := limiter.Allow(context.Background(), "user123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !allowed {
			t.Error("first request should be allowed")
		}

		// Check bucket was created
		limiter.mu.RLock()
		if _, exists := limiter.buckets["user123"]; !exists {
			t.Error("bucket should be created for key")
		}
		limiter.mu.RUnlock()
	})

	t.Run("enforces rate limit", func(t *testing.T) {
		limiter := NewInMemoryRateLimiter(10, 2, time.Second) // 2 burst
		ctx := context.Background()

		// First 2 requests should succeed (burst)
		for i := 0; i < 2; i++ {
			allowed, err := limiter.Allow(ctx, "user456")
			if err != nil {
				t.Fatalf("unexpected error on request %d: %v", i+1, err)
			}
			if !allowed {
				t.Errorf("request %d should be allowed (within burst)", i+1)
			}
		}

		// 3rd request should be denied (burst exhausted)
		allowed, err := limiter.Allow(ctx, "user456")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if allowed {
			t.Error("request beyond burst should be denied")
		}
	})

	t.Run("isolates keys", func(t *testing.T) {
		limiter := NewInMemoryRateLimiter(10, 1, time.Second)
		ctx := context.Background()

		// Exhaust user1's quota
		_, _ = limiter.Allow(ctx, "user1")

		// user2 should still be allowed
		allowed, err := limiter.Allow(ctx, "user2")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !allowed {
			t.Error("different key should have separate quota")
		}
	})

	t.Run("reset clears bucket", func(t *testing.T) {
		limiter := NewInMemoryRateLimiter(10, 1, time.Second)
		ctx := context.Background()

		// Exhaust quota
		_, _ = limiter.Allow(ctx, "user789")
		allowed, _ := limiter.Allow(ctx, "user789")
		if allowed {
			t.Error("should be rate limited")
		}

		// Reset quota
		err := limiter.Reset(ctx, "user789")
		if err != nil {
			t.Fatalf("reset error: %v", err)
		}

		// Should be allowed again
		allowed, err = limiter.Allow(ctx, "user789")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !allowed {
			t.Error("should be allowed after reset")
		}
	})

	t.Run("thread safety", func(t *testing.T) {
		limiter := NewInMemoryRateLimiter(100, 50, time.Second)
		ctx := context.Background()

		const goroutines = 10
		const requestsPerGoroutine = 10

		var wg sync.WaitGroup
		wg.Add(goroutines)

		for i := 0; i < goroutines; i++ {
			go func(id int) {
				defer wg.Done()
				key := fmt.Sprintf("user%d", id)

				for j := 0; j < requestsPerGoroutine; j++ {
					_, err := limiter.Allow(ctx, key)
					if err != nil {
						t.Errorf("goroutine %d request %d error: %v", id, j, err)
					}
				}
			}(i)
		}

		wg.Wait()

		// Verify all keys were created
		limiter.mu.RLock()
		if len(limiter.buckets) != goroutines {
			t.Errorf("expected %d buckets, got %d", goroutines, len(limiter.buckets))
		}
		limiter.mu.RUnlock()
	})
}

func TestRateLimitGuardrail(t *testing.T) {
	t.Run("allows requests within limit", func(t *testing.T) {
		guard := NewRateLimitGuardrail(RateLimitConfig{
			Rate:   10,
			Burst:  5,
			Window: time.Second,
		})
		defer func() { _ = guard.Close() }()

		ctx := context.Background()

		// First 5 should be allowed
		for i := 0; i < 5; i++ {
			err := guard.Validate(ctx, "test input")
			if err != nil {
				t.Errorf("request %d should be allowed: %v", i+1, err)
			}
		}
	})

	t.Run("blocks requests over limit", func(t *testing.T) {
		guard := NewRateLimitGuardrail(RateLimitConfig{
			Rate:   10,
			Burst:  2,
			Window: time.Second,
		})
		defer func() { _ = guard.Close() }()

		ctx := context.Background()

		// Exhaust burst
		_ = guard.Validate(ctx, "test input")
		_ = guard.Validate(ctx, "test input")

		// Should be blocked
		err := guard.Validate(ctx, "test input")
		if err == nil {
			t.Error("expected rate limit error")
		}
	})

	t.Run("custom key function", func(t *testing.T) {
		guard := NewRateLimitGuardrail(RateLimitConfig{
			Rate:   10,
			Burst:  1,
			Window: time.Second,
			KeyFunc: func(input string) string {
				// Use input as key (per-message limiting)
				return input
			},
		})
		defer func() { _ = guard.Close() }()

		ctx := context.Background()

		// Different inputs should have separate limits
		err1 := guard.Validate(ctx, "input1")
		err2 := guard.Validate(ctx, "input2")

		if err1 != nil || err2 != nil {
			t.Error("different keys should have separate quotas")
		}

		// Same input should be rate limited
		err := guard.Validate(ctx, "input1")
		if err == nil {
			t.Error("same key should be rate limited")
		}
	})

	t.Run("tripwire mode", func(t *testing.T) {
		guard := NewRateLimitGuardrail(RateLimitConfig{
			Rate:     10,
			Burst:    1,
			Window:   time.Second,
			Tripwire: true,
		})
		defer func() { _ = guard.Close() }()

		if !guard.IsTripwire() {
			t.Error("guardrail should be marked as tripwire")
		}

		ctx := context.Background()

		// Exhaust burst
		_ = guard.Validate(ctx, "test")

		// Should return TripwireError
		err := guard.Validate(ctx, "test")
		if err == nil {
			t.Error("expected tripwire error")
		}

		// TODO: Check for TripwireError type once imported
	})

	t.Run("name", func(t *testing.T) {
		guard := NewRateLimitGuardrail(RateLimitConfig{
			Rate:   10,
			Burst:  5,
			Window: time.Second,
		})
		defer func() { _ = guard.Close() }()

		if guard.Name() != "rate_limit" {
			t.Errorf("expected name 'rate_limit', got '%s'", guard.Name())
		}
	})
}

func TestRateLimitGuardrail_PerAgent(t *testing.T) {
	// Simulate per-agent rate limiting
	guard := NewRateLimitGuardrail(RateLimitConfig{
		Rate:   100,
		Burst:  10,
		Window: time.Second,
		KeyFunc: func(_ string) string {
			// Extract agent name from input (simplified)
			// In real usage, this would parse context variables
			return "agent1"
		},
	})
	defer func() { _ = guard.Close() }()

	ctx := context.Background()

	// First 10 requests for agent1 should succeed
	for i := 0; i < 10; i++ {
		err := guard.Validate(ctx, "test")
		if err != nil {
			t.Errorf("request %d failed: %v", i+1, err)
		}
	}

	// 11th should be blocked
	err := guard.Validate(ctx, "test")
	if err == nil {
		t.Error("expected rate limit for agent1")
	}
}
