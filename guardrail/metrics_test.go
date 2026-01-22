package guardrail

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestInMemoryMetrics(t *testing.T) {
	t.Run("track_success_and_failure", func(t *testing.T) {
		metrics := NewInMemoryMetrics()
		guard := NewGuardrail("test_guard", func(_ context.Context, input string) (*Result, error) {
			if input == "pass" {
				return &Result{Passed: true}, nil
			}
			return &Result{Passed: false, Message: "failed"}, nil
		})

		wrapped := WithMetrics(guard, metrics)

		// Execute passing validations
		for i := 0; i < 5; i++ {
			_, _ = wrapped.Func(context.Background(), "pass")
		}

		// Execute failing validations
		for i := 0; i < 3; i++ {
			_, _ = wrapped.Func(context.Background(), "fail")
		}

		stats := metrics.GetStats("test_guard")
		if stats == nil {
			t.Fatal("expected stats but got nil")
		}

		if stats.TotalCount != 8 {
			t.Errorf("expected 8 total executions, got %d", stats.TotalCount)
		}

		if stats.PassedCount != 5 {
			t.Errorf("expected 5 passed, got %d", stats.PassedCount)
		}

		if stats.FailedCount != 3 {
			t.Errorf("expected 3 failed, got %d", stats.FailedCount)
		}
	})

	t.Run("track_duration_statistics", func(t *testing.T) {
		metrics := NewInMemoryMetrics()
		guard := NewGuardrail("timing_guard", func(_ context.Context, _ string) (*Result, error) {
			// Simulate varying execution times
			time.Sleep(10 * time.Millisecond)
			return &Result{Passed: true}, nil
		})

		wrapped := WithMetrics(guard, metrics)

		// Execute multiple times
		for i := 0; i < 10; i++ {
			_, _ = wrapped.Func(context.Background(), "test")
		}

		stats := metrics.GetStats("timing_guard")
		if stats == nil {
			t.Fatal("expected stats but got nil")
		}

		avgDuration := stats.AvgDuration()
		if avgDuration == 0 {
			t.Error("expected non-zero average duration")
		}

		if avgDuration < 5*time.Millisecond {
			t.Errorf("average duration too low: %v", avgDuration)
		}
	})

	t.Run("calculate_percentiles", func(t *testing.T) {
		metrics := NewInMemoryMetrics()

		// Create a guardrail with controlled timing
		guard := NewGuardrail("percentile_guard", func(_ context.Context, input string) (*Result, error) {
			// Parse input as sleep duration
			var sleep time.Duration
			switch input {
			case "fast":
				sleep = 10 * time.Millisecond
			case "medium":
				sleep = 50 * time.Millisecond
			case "slow":
				sleep = 100 * time.Millisecond
			}
			time.Sleep(sleep)
			return &Result{Passed: true}, nil
		})

		wrapped := WithMetrics(guard, metrics)

		// Create a distribution with known percentiles
		inputs := []string{"fast", "fast", "fast", "fast", "fast", "fast", "fast", "medium", "medium", "slow"}
		for _, input := range inputs {
			_, _ = wrapped.Func(context.Background(), input)
		}

		stats := metrics.GetStats("percentile_guard")
		if stats == nil {
			t.Fatal("expected stats but got nil")
		}

		p95 := stats.P95Duration()
		p99 := stats.P99Duration()

		// P95 should be close to medium or slow
		if p95 < 40*time.Millisecond {
			t.Errorf("P95 duration too low: %v", p95)
		}

		// P99 should be close to slow
		if p99 < 80*time.Millisecond {
			t.Errorf("P99 duration too low: %v", p99)
		}

		// P99 should be >= P95
		if p99 < p95 {
			t.Errorf("P99 (%v) should be >= P95 (%v)", p99, p95)
		}
	})

	t.Run("track_multiple_guardrails", func(t *testing.T) {
		metrics := NewInMemoryMetrics()

		guard1 := NewGuardrail("guard1", func(_ context.Context, _ string) (*Result, error) {
			return &Result{Passed: true}, nil
		})

		guard2 := NewGuardrail("guard2", func(_ context.Context, _ string) (*Result, error) {
			return &Result{Passed: false, Message: "always fails"}, nil
		})

		wrapped1 := WithMetrics(guard1, metrics)
		wrapped2 := WithMetrics(guard2, metrics)

		// Execute both guards
		_, _ = wrapped1.Func(context.Background(), "test")
		_, _ = wrapped1.Func(context.Background(), "test")
		_, _ = wrapped2.Func(context.Background(), "test")

		allStats := metrics.GetAllStats()
		if len(allStats) != 2 {
			t.Errorf("expected 2 guardrail stats, got %d", len(allStats))
		}

		stats1 := metrics.GetStats("guard1")
		if stats1 == nil || stats1.TotalCount != 2 {
			t.Error("guard1 stats incorrect")
		}

		stats2 := metrics.GetStats("guard2")
		if stats2 == nil || stats2.TotalCount != 1 {
			t.Error("guard2 stats incorrect")
		}
	})

	t.Run("handle_errors_in_validation", func(t *testing.T) {
		metrics := NewInMemoryMetrics()
		guard := NewGuardrail("error_guard", func(_ context.Context, input string) (*Result, error) {
			if input == "error" {
				return nil, errors.New("validation error")
			}
			return &Result{Passed: true}, nil
		})

		wrapped := WithMetrics(guard, metrics)

		// Execute with error
		_, err := wrapped.Func(context.Background(), "error")
		if err == nil {
			t.Error("expected error")
		}

		// Execute successfully
		_, _ = wrapped.Func(context.Background(), "success")

		stats := metrics.GetStats("error_guard")
		if stats == nil {
			t.Fatal("expected stats but got nil")
		}

		// Both executions should be counted
		if stats.TotalCount != 2 {
			t.Errorf("expected 2 executions, got %d", stats.TotalCount)
		}
	})
}

func TestMetricsThreadSafety(t *testing.T) {
	t.Run("concurrent_access_is_safe", func(t *testing.T) {
		metrics := NewInMemoryMetrics()
		guard := NewGuardrail("concurrent_guard", func(_ context.Context, _ string) (*Result, error) {
			time.Sleep(1 * time.Millisecond)
			return &Result{Passed: true}, nil
		})

		wrapped := WithMetrics(guard, metrics)

		// Run concurrent validations
		done := make(chan bool)
		for i := 0; i < 10; i++ {
			go func() {
				for j := 0; j < 10; j++ {
					_, _ = wrapped.Func(context.Background(), "test")
				}
				done <- true
			}()
		}

		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			<-done
		}

		stats := metrics.GetStats("concurrent_guard")
		if stats == nil {
			t.Fatal("expected stats but got nil")
		}

		if stats.TotalCount != 100 {
			t.Errorf("expected 100 executions, got %d", stats.TotalCount)
		}
	})
}

func TestMetricsWithChain(t *testing.T) {
	t.Run("track_metrics_for_each_guardrail_in_chain", func(t *testing.T) {
		metrics := NewInMemoryMetrics()

		guard1 := NewGuardrail("chain_guard1", func(_ context.Context, _ string) (*Result, error) {
			time.Sleep(10 * time.Millisecond)
			return &Result{Passed: true}, nil
		})

		guard2 := NewGuardrail("chain_guard2", func(_ context.Context, _ string) (*Result, error) {
			time.Sleep(20 * time.Millisecond)
			return &Result{Passed: true}, nil
		})

		chain := NewChain().
			Add(WithMetrics(guard1, metrics)).
			Add(WithMetrics(guard2, metrics)).
			WithStrategy(Sequential).
			Build()

		// Execute chain
		_, _ = chain.Func(context.Background(), "test")

		// Check both guardrails were tracked
		stats1 := metrics.GetStats("chain_guard1")
		stats2 := metrics.GetStats("chain_guard2")

		if stats1 == nil || stats1.TotalCount != 1 {
			t.Error("chain_guard1 not tracked correctly")
		}

		if stats2 == nil || stats2.TotalCount != 1 {
			t.Error("chain_guard2 not tracked correctly")
		}

		// Verify durations are different
		avg1 := stats1.AvgDuration()
		avg2 := stats2.AvgDuration()

		if avg2 <= avg1 {
			t.Errorf("expected guard2 (%v) to be slower than guard1 (%v)", avg2, avg1)
		}
	})

	t.Run("short_circuit_affects_metrics", func(t *testing.T) {
		metrics := NewInMemoryMetrics()

		guard1 := NewGuardrail("shortcircuit_guard1", func(_ context.Context, _ string) (*Result, error) {
			return &Result{Passed: false, Message: "first fails"}, nil
		})

		guard2 := NewGuardrail("shortcircuit_guard2", func(_ context.Context, _ string) (*Result, error) {
			return &Result{Passed: true}, nil
		})

		chain := NewChain().
			Add(WithMetrics(guard1, metrics)).
			Add(WithMetrics(guard2, metrics)).
			WithStrategy(Sequential). // Sequential short-circuits
			Build()

		// Execute chain (should stop at guard1)
		_, _ = chain.Func(context.Background(), "test")

		// Only guard1 should have been executed
		stats1 := metrics.GetStats("shortcircuit_guard1")
		stats2 := metrics.GetStats("shortcircuit_guard2")

		if stats1 == nil || stats1.TotalCount != 1 {
			t.Error("shortcircuit_guard1 should have 1 execution")
		}

		if stats2 != nil && stats2.TotalCount > 0 {
			t.Error("shortcircuit_guard2 should not have been executed due to short-circuit")
		}
	})
}

// MockMetricsCollector is a helper for testing
type MockMetricsCollector struct {
	StartFunc    func(name string)
	CompleteFunc func(name string, result *Result, duration time.Duration)
	ErrorFunc    func(name string, err error)
}

func (m *MockMetricsCollector) OnStart(name string) {
	if m.StartFunc != nil {
		m.StartFunc(name)
	}
}

func (m *MockMetricsCollector) OnComplete(name string, result *Result, duration time.Duration) {
	if m.CompleteFunc != nil {
		m.CompleteFunc(name, result, duration)
	}
}

func (m *MockMetricsCollector) OnError(name string, err error) {
	if m.ErrorFunc != nil {
		m.ErrorFunc(name, err)
	}
}

func TestMetricsCollectorInterface(t *testing.T) {
	t.Run("custom_metrics_collector", func(t *testing.T) {
		var startCount, completeCount, errorCount int

		collector := &MockMetricsCollector{
			StartFunc: func(_ string) {
				startCount++
			},
			CompleteFunc: func(_ string, _ *Result, _ time.Duration) {
				completeCount++
			},
			ErrorFunc: func(_ string, _ error) {
				errorCount++
			},
		}

		// Verify it implements the interface
		var _ MetricsCollector = collector

		// Use it
		collector.OnStart("test")
		collector.OnComplete("test", &Result{Passed: true}, time.Second)
		collector.OnError("test", errors.New("err"))

		if startCount != 1 {
			t.Error("OnStart not called")
		}
		if completeCount != 1 {
			t.Error("OnComplete not called")
		}
		if errorCount != 1 {
			t.Error("OnError not called")
		}
	})
}
