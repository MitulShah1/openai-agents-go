package builtin

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/MitulShah1/openai-agents-go/guardrail"
)

// RateLimiter defines the interface for pluggable rate limiting backends.
// This interface allows for distributed rate limiting in multi-agent systems.
type RateLimiter interface {
	// Allow checks if a request is allowed for the given key.
	// Returns true if allowed, false if rate limited, and any error.
	Allow(ctx context.Context, key string) (bool, error)

	// Reset resets the rate limit state for the given key.
	Reset(ctx context.Context, key string) error

	// Close cleans up any resources held by the rate limiter.
	Close() error
}

// TokenBucket implements a token bucket rate limiter algorithm.
type TokenBucket struct {
	tokens     float64   // Current number of tokens
	capacity   float64   // Maximum tokens (burst size)
	refillRate float64   // Tokens added per second
	lastRefill time.Time // Last time tokens were refilled
	mu         sync.Mutex
}

// InMemoryRateLimiter implements thread-safe in-memory rate limiting.
// This is the v0.3.0 built-in implementation.
type InMemoryRateLimiter struct {
	rate    int                     // Tokens per time window
	burst   int                     // Maximum burst size
	window  time.Duration           // Time window for rate limiting
	buckets map[string]*TokenBucket // Key to token bucket mapping
	mu      sync.RWMutex            // Protects buckets map
}

// NewInMemoryRateLimiter creates a new in-memory rate limiter.
func NewInMemoryRateLimiter(rate int, burst int, window time.Duration) *InMemoryRateLimiter {
	return &InMemoryRateLimiter{
		rate:    rate,
		burst:   burst,
		window:  window,
		buckets: make(map[string]*TokenBucket),
	}
}

// Allow checks if a request is allowed for the given key.
func (r *InMemoryRateLimiter) Allow(_ context.Context, key string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Get or create bucket for this key
	bucket, exists := r.buckets[key]
	if !exists {
		refillRate := float64(r.rate) / r.window.Seconds()
		bucket = &TokenBucket{
			tokens:     float64(r.burst),
			capacity:   float64(r.burst),
			refillRate: refillRate,
			lastRefill: time.Now(),
		}
		r.buckets[key] = bucket
	}

	return bucket.consume(), nil
}

// Reset resets the rate limit state for the given key.
func (r *InMemoryRateLimiter) Reset(_ context.Context, key string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.buckets, key)
	return nil
}

// Close cleans up any resources (no-op for in-memory).
func (r *InMemoryRateLimiter) Close() error {
	return nil
}

// consume attempts to consume one token from the bucket.
func (tb *TokenBucket) consume() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// Refill tokens based on time elapsed
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.tokens += elapsed * tb.refillRate

	// Cap at max capacity
	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}

	tb.lastRefill = now

	// Try to consume one token
	if tb.tokens >= 1.0 {
		tb.tokens -= 1.0
		return true
	}

	return false
}

// RateLimitGuardrail implements rate limiting for multi-agent distributed systems.
type RateLimitGuardrail struct {
	limiter  RateLimiter               // Pluggable backend (in-memory or distributed)
	keyFunc  func(input string) string // Function to extract rate limit key
	tripwire bool                      // If true, halt execution on rate limit
}

// RateLimitConfig configures the rate limit guardrail.
type RateLimitConfig struct {
	// Rate is the number of tokens per time window
	Rate int
	// Burst is the maximum burst size
	Burst int
	// Window is the time window for rate limiting
	Window time.Duration
	// KeyFunc extracts the rate limit key from input (default: global key)
	KeyFunc func(input string) string
	// Limiter is a custom rate limiter backend (default: in-memory)
	Limiter RateLimiter
	// Tripwire halts execution if rate limit is exceeded
	Tripwire bool
}

// NewRateLimitGuardrail creates a new rate limiting guardrail.
func NewRateLimitGuardrail(config RateLimitConfig) *RateLimitGuardrail {
	// Default to global key
	keyFunc := config.KeyFunc
	if keyFunc == nil {
		keyFunc = func(_ string) string {
			return "global"
		}
	}

	// Default to in-memory limiter
	limiter := config.Limiter
	if limiter == nil {
		limiter = NewInMemoryRateLimiter(config.Rate, config.Burst, config.Window)
	}

	return &RateLimitGuardrail{
		limiter:  limiter,
		keyFunc:  keyFunc,
		tripwire: config.Tripwire,
	}
}

// Name returns the guardrail name.
func (g *RateLimitGuardrail) Name() string {
	return "rate_limit"
}

// Validate checks if the input passes the rate limit.
func (g *RateLimitGuardrail) Validate(ctx context.Context, input string) error {
	key := g.keyFunc(input)
	allowed, err := g.limiter.Allow(ctx, key)
	if err != nil {
		return fmt.Errorf("rate limiter error: %w", err)
	}

	if !allowed {
		msg := fmt.Sprintf("rate limit exceeded for key: %s", key)
		if g.tripwire {
			return &guardrail.InputGuardrailTripwireError{
				GuardrailName: g.Name(),
				Message:       msg,
			}
		}
		return errors.New(msg)
	}

	return nil
}

// IsTripwire returns whether this guardrail halts execution on failure.
func (g *RateLimitGuardrail) IsTripwire() bool {
	return g.tripwire
}

// Close cleans up the rate limiter resources.
func (g *RateLimitGuardrail) Close() error {
	if g.limiter != nil {
		return g.limiter.Close()
	}
	return nil
}
