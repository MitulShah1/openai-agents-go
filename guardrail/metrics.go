package guardrail

import (
	"context"
	"sync"
	"time"
)

// MetricsCollector defines the interface for collecting guardrail telemetry.
type MetricsCollector interface {
	// OnStart is called when a guardrail starts execution.
	OnStart(name string)
	// OnComplete is called when a guardrail completes successfully.
	OnComplete(name string, result *Result, duration time.Duration)
	// OnError is called when a guardrail encounters an error.
	OnError(name string, _ error)
}

// GuardrailStats contains statistics for a single guardrail.
//
//nolint:revive // GuardrailStats is intentionally prefixed for clarity in external usage
type GuardrailStats struct {
	TotalCount    int
	PassedCount   int
	FailedCount   int
	TripwireCount int
	ErrorCount    int
	TotalDuration time.Duration
	MinDuration   time.Duration
	MaxDuration   time.Duration
	Durations     []time.Duration // For percentile calculations
}

// AvgDuration returns the average execution duration.
func (s *GuardrailStats) AvgDuration() time.Duration {
	if s.TotalCount == 0 {
		return 0
	}
	return s.TotalDuration / time.Duration(s.TotalCount)
}

// P95Duration returns the 95th percentile duration.
func (s *GuardrailStats) P95Duration() time.Duration {
	return s.percentile(0.95)
}

// P99Duration returns the 99th percentile duration.
func (s *GuardrailStats) P99Duration() time.Duration {
	return s.percentile(0.99)
}

func (s *GuardrailStats) percentile(p float64) time.Duration {
	if len(s.Durations) == 0 {
		return 0
	}

	// Simple percentile calculation (not sorted, but good enough for monitoring)
	// In production, consider using a proper streaming percentile algorithm
	sorted := make([]time.Duration, len(s.Durations))
	copy(sorted, s.Durations)

	// Simple bubble sort for small datasets
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	idx := int(float64(len(sorted)) * p)
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

// InMemoryMetrics is a thread-safe in-memory metrics collector.
type InMemoryMetrics struct {
	mu    sync.RWMutex
	stats map[string]*GuardrailStats
}

// NewInMemoryMetrics creates a new in-memory metrics collector.
func NewInMemoryMetrics() *InMemoryMetrics {
	return &InMemoryMetrics{
		stats: make(map[string]*GuardrailStats),
	}
}

// OnStart records the start of a guardrail execution.
func (m *InMemoryMetrics) OnStart(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.stats[name]; !exists {
		m.stats[name] = &GuardrailStats{
			Durations: make([]time.Duration, 0, 100),
		}
	}
}

// OnComplete records the completion of a guardrail execution.
func (m *InMemoryMetrics) OnComplete(name string, result *Result, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	stats, exists := m.stats[name]
	if !exists {
		stats = &GuardrailStats{
			Durations: make([]time.Duration, 0, 100),
		}
		m.stats[name] = stats
	}

	stats.TotalCount++
	stats.TotalDuration += duration
	stats.Durations = append(stats.Durations, duration)

	if stats.MinDuration == 0 || duration < stats.MinDuration {
		stats.MinDuration = duration
	}
	if duration > stats.MaxDuration {
		stats.MaxDuration = duration
	}

	if result.Passed {
		stats.PassedCount++
	} else {
		stats.FailedCount++
	}

	if result.TripwireTriggered {
		stats.TripwireCount++
	}
}

// OnError records an error during guardrail execution.
func (m *InMemoryMetrics) OnError(name string, _ error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	stats, exists := m.stats[name]
	if !exists {
		stats = &GuardrailStats{
			Durations: make([]time.Duration, 0, 100),
		}
		m.stats[name] = stats
	}

	stats.ErrorCount++
	stats.TotalCount++
}

// GetStats returns a copy of the statistics for a guardrail.
func (m *InMemoryMetrics) GetStats(name string) *GuardrailStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats, exists := m.stats[name]
	if !exists {
		return nil
	}

	// Return a copy to avoid race conditions
	statsCopy := *stats
	statsCopy.Durations = make([]time.Duration, len(stats.Durations))
	copy(statsCopy.Durations, stats.Durations)

	return &statsCopy
}

// GetAllStats returns a copy of all statistics.
func (m *InMemoryMetrics) GetAllStats() map[string]*GuardrailStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	allStats := make(map[string]*GuardrailStats, len(m.stats))
	for name, stats := range m.stats {
		statsCopy := *stats
		statsCopy.Durations = make([]time.Duration, len(stats.Durations))
		copy(statsCopy.Durations, stats.Durations)
		allStats[name] = &statsCopy
	}

	return allStats
}

// Reset clears all metrics.
func (m *InMemoryMetrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stats = make(map[string]*GuardrailStats)
}

// WithMetrics wraps a guardrail with metrics collection.
func WithMetrics(g *Guardrail, collector MetricsCollector) *Guardrail {
	return &Guardrail{
		Name:          g.Name,
		RunInParallel: g.RunInParallel,
		Func: func(ctx context.Context, input string) (*Result, error) {
			collector.OnStart(g.Name)
			start := time.Now()

			result, err := g.Func(ctx, input)
			duration := time.Since(start)

			if err != nil {
				collector.OnError(g.Name, err)
				return nil, err
			}

			collector.OnComplete(g.Name, result, duration)
			return result, nil
		},
	}
}
