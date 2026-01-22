package ratelimit

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/MitulShah1/openai-agents-go/guardrail"
)

func TestInMemoryRateLimiter(t *testing.T) {
	// 5 requests per 1 second, with burst of 5
	limiter := NewInMemoryRateLimiter(5, 5, time.Second)
	ctx := context.Background()
	key := "test-key"

	// Consume all 5 tokens
	for i := 0; i < 5; i++ {
		allowed, err := limiter.Allow(ctx, key)
		if err != nil {
			t.Fatalf("Unexpected error on request %d: %v", i, err)
		}
		if !allowed {
			t.Fatalf("Request %d should be allowed", i)
		}
	}

	// 6th request should fail
	allowed, err := limiter.Allow(ctx, key)
	if err != nil {
		t.Fatalf("Unexpected error on 6th request: %v", err)
	}
	if allowed {
		t.Fatal("6th request should be rate limited")
	}

	// Refill: Wait for 1 second (plus a tiny buffer)
	// We use a slightly shorter sleep for "partial" refill if we wanted,
	// but let's wait full window for simplicity
	time.Sleep(1100 * time.Millisecond)

	// Should be allowed again
	allowed, err = limiter.Allow(ctx, key)
	if err != nil {
		t.Fatalf("Unexpected error after refill: %v", err)
	}
	if !allowed {
		t.Fatal("Request after refill should be allowed")
	}
}

func TestGuardrail_Validation(t *testing.T) {
	tests := []struct {
		name            string
		config          Config
		requests        int
		expectedFailIdx int // -1 if all pass
		expectTripwire  bool
	}{
		{
			name: "basic_limits",
			config: Config{
				Rate:   2,
				Burst:  2,
				Window: time.Second,
			},
			requests:        3,
			expectedFailIdx: 2, // 3rd request (index 2) fails
			expectTripwire:  false,
		},
		{
			name: "tripwire_enabled",
			config: Config{
				Rate:     1,
				Burst:    1,
				Window:   time.Second,
				Tripwire: true,
			},
			requests:        2,
			expectedFailIdx: 1, // 2nd request fails
			expectTripwire:  true,
		},
		{
			name: "custom_key_func_logic",
			config: Config{
				Rate:   1,
				Burst:  1,
				Window: time.Second,
				KeyFunc: func(input string) string {
					return input // rate limit by input content
				},
			},
			requests:        2,  // sending different inputs
			expectedFailIdx: -1, // both should pass because keys are different
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			guard := New(tt.config)
			ctx := context.Background()

			for i := 0; i < tt.requests; i++ {
				input := "static-input"
				if tt.config.KeyFunc != nil {
					// vary input for key func test
					if i == 0 {
						input = "user1"
					} else {
						input = "user2"
					}
				}

				err := guard.Validate(ctx, input)

				if tt.expectedFailIdx == i {
					if err == nil {
						t.Errorf("Request %d expected to fail, passed", i)
					} else {
						// Check error type
						_, isTripwire := err.(*guardrail.InputGuardrailTripwireError)
						if tt.expectTripwire && !isTripwire {
							t.Errorf("Expected tripwire error, got: %T", err)
						}
						if !tt.expectTripwire && isTripwire {
							t.Errorf("Expected standard error, got tripwire")
						}
						if !strings.Contains(err.Error(), "rate limit exceeded") {
							t.Errorf("Unexpected error message: %v", err)
						}
					}
				} else {
					if err != nil {
						t.Errorf("Request %d expected to pass, failed: %v", i, err)
					}
				}
			}
		})
	}
}

func TestGuardrail_Defaults(t *testing.T) {
	// Test that New() handles defaults (nil limiter, nil KeyFunc)
	g := New(Config{Rate: 10, Burst: 10, Window: time.Minute})
	if g.keyFunc == nil {
		t.Error("Default KeyFunc not set")
	}
	if g.limiter == nil {
		t.Error("Default Limiter not set")
	}

	// Verify default key is "global"
	if g.keyFunc("input") != "global" {
		t.Error("Default KeyFunc should return 'global'")
	}
}
