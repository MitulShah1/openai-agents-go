package agents

import (
	"testing"
	"time"
)

func TestDefaultRunConfig(t *testing.T) {
	config := DefaultRunConfig()

	if config.MaxTurns != 10 {
		t.Errorf("expected MaxTurns=10, got %d", config.MaxTurns)
	}

	if config.Debug != false {
		t.Error("expected Debug=false")
	}

	if config.Timeout != 5*time.Minute {
		t.Errorf("expected Timeout=5m, got %v", config.Timeout)
	}

	if config.Temperature != nil {
		t.Error("expected Temperature=nil")
	}

	if config.MaxTokens != nil {
		t.Error("expected MaxTokens=nil")
	}
}

func TestRunConfigMerge(t *testing.T) {
	tests := []struct {
		name     string
		base     *RunConfig
		override *RunConfig
		validate func(*testing.T, *RunConfig)
	}{
		{
			name:     "nil override returns base",
			base:     &RunConfig{MaxTurns: 5},
			override: nil,
			validate: func(t *testing.T, result *RunConfig) {
				if result.MaxTurns != 5 {
					t.Errorf("expected MaxTurns=5, got %d", result.MaxTurns)
				}
			},
		},
		{
			name:     "override MaxTurns",
			base:     &RunConfig{MaxTurns: 5},
			override: &RunConfig{MaxTurns: 10},
			validate: func(t *testing.T, result *RunConfig) {
				if result.MaxTurns != 10 {
					t.Errorf("expected MaxTurns=10, got %d", result.MaxTurns)
				}
			},
		},
		{
			name:     "override Temperature",
			base:     &RunConfig{MaxTurns: 5},
			override: &RunConfig{Temperature: floatPtr(0.7)},
			validate: func(t *testing.T, result *RunConfig) {
				if result.Temperature == nil || *result.Temperature != 0.7 {
					t.Errorf("expected Temperature=0.7, got %v", result.Temperature)
				}
			},
		},
		{
			name:     "override MaxTokens",
			base:     &RunConfig{},
			override: &RunConfig{MaxTokens: intPtr(100)},
			validate: func(t *testing.T, result *RunConfig) {
				if result.MaxTokens == nil || *result.MaxTokens != 100 {
					t.Errorf("expected MaxTokens=100, got %v", result.MaxTokens)
				}
			},
		},
		{
			name:     "override Debug",
			base:     &RunConfig{Debug: false},
			override: &RunConfig{Debug: true},
			validate: func(t *testing.T, result *RunConfig) {
				if !result.Debug {
					t.Error("expected Debug=true")
				}
			},
		},
		{
			name:     "override Timeout",
			base:     &RunConfig{},
			override: &RunConfig{Timeout: 2 * time.Minute},
			validate: func(t *testing.T, result *RunConfig) {
				if result.Timeout != 2*time.Minute {
					t.Errorf("expected Timeout=2m, got %v", result.Timeout)
				}
			},
		},
		{
			name:     "zero values don't override",
			base:     &RunConfig{MaxTurns: 10},
			override: &RunConfig{MaxTurns: 0}, // Zero value
			validate: func(t *testing.T, result *RunConfig) {
				if result.MaxTurns != 10 {
					t.Errorf("expected MaxTurns=10 (not overridden), got %d", result.MaxTurns)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.base.Merge(tt.override)
			tt.validate(t, result)
		})
	}
}

// Helper functions
func floatPtr(f float64) *float64 {
	return &f
}

func intPtr(i int) *int {
	return &i
}
