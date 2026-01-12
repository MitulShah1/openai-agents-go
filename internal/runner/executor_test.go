package runner

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestNewExecutor(t *testing.T) {
	maxTurns := 10
	timeout := 5 * time.Minute

	executor := NewExecutor(maxTurns, timeout)

	if executor == nil {
		t.Fatal("expected NewExecutor to return non-nil executor")
	}

	if executor.maxTurns != maxTurns {
		t.Errorf("expected maxTurns=%d, got %d", maxTurns, executor.maxTurns)
	}

	if executor.timeout != timeout {
		t.Errorf("expected timeout=%v, got %v", timeout, executor.timeout)
	}
}

func TestCheckContext(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		wantErr bool
		errType error
	}{
		{
			name:    "valid context",
			ctx:     context.Background(),
			wantErr: false,
		},
		{
			name: "cancelled context",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			wantErr: true,
			errType: context.Canceled,
		},
		{
			name: "timeout context",
			ctx: func() context.Context {
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
				defer cancel()
				time.Sleep(2 * time.Millisecond)
				return ctx
			}(),
			wantErr: true,
			errType: context.DeadlineExceeded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckContext(tt.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckContext() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errType != nil {
				if !errors.Is(err, tt.errType) {
					t.Errorf("CheckContext() error type = %v, want %v", err, tt.errType)
				}
			}
		})
	}
}

func TestExecutor_ShouldContinueExecution(t *testing.T) {
	tests := []struct {
		name         string
		maxTurns     int
		turnCount    int
		ctx          context.Context
		wantContinue bool
		wantErr      bool
	}{
		{
			name:         "should continue - below max turns",
			maxTurns:     10,
			turnCount:    5,
			ctx:          context.Background(),
			wantContinue: true,
			wantErr:      false,
		},
		{
			name:         "should continue - no max turns limit",
			maxTurns:     0,
			turnCount:    100,
			ctx:          context.Background(),
			wantContinue: true,
			wantErr:      false,
		},
		{
			name:         "should stop - max turns exceeded",
			maxTurns:     5,
			turnCount:    5,
			ctx:          context.Background(),
			wantContinue: false,
			wantErr:      true,
		},
		{
			name:      "should stop - context cancelled",
			maxTurns:  10,
			turnCount: 2,
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			wantContinue: false,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := NewExecutor(tt.maxTurns, 0)

			shouldContinue, err := executor.ShouldContinueExecution(tt.ctx, tt.turnCount)

			if shouldContinue != tt.wantContinue {
				t.Errorf("ShouldContinueExecution() shouldContinue = %v, want %v", shouldContinue, tt.wantContinue)
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("ShouldContinueExecution() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExecutor_ApplyTimeout(t *testing.T) {
	tests := []struct {
		name           string
		timeout        time.Duration
		expectDeadline bool
	}{
		{
			name:           "with timeout",
			timeout:        100 * time.Millisecond,
			expectDeadline: true,
		},
		{
			name:           "without timeout",
			timeout:        0,
			expectDeadline: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := NewExecutor(0, tt.timeout)

			ctx := context.Background()
			newCtx, cancel := executor.ApplyTimeout(ctx)
			defer cancel()

			_, hasDeadline := newCtx.Deadline()
			if hasDeadline != tt.expectDeadline {
				t.Errorf("ApplyTimeout() deadline presence = %v, want %v", hasDeadline, tt.expectDeadline)
			}

			// Verify the timeout is actually applied
			if tt.expectDeadline {
				deadline, _ := newCtx.Deadline()
				expectedDeadline := time.Now().Add(tt.timeout)
				// Allow 10ms tolerance for timing
				if deadline.After(expectedDeadline.Add(10*time.Millisecond)) ||
					deadline.Before(expectedDeadline.Add(-10*time.Millisecond)) {
					t.Errorf("ApplyTimeout() deadline not correctly set")
				}
			}
		})
	}
}

func TestExtractFinalOutput(t *testing.T) {
	// Test the placeholder implementation
	result := ExtractFinalOutput(nil)
	if result != "" {
		t.Errorf("ExtractFinalOutput() = %v, want empty string", result)
	}

	// Test with some data
	result = ExtractFinalOutput("test message")
	if result != "" {
		t.Errorf("ExtractFinalOutput() = %v, want empty string (placeholder)", result)
	}
}
