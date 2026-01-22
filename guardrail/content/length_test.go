package content

import (
	"context"
	"testing"
)

func TestNewLength(t *testing.T) {
	tests := []struct {
		name             string
		config           Config
		input            string
		expectedPass     bool
		expectedTripwire bool
		expectError      bool
	}{
		{
			name: "character_count_within_limits",
			config: Config{
				Mode: CountModeCharacters,
				Min:  5,
				Max:  20,
			},
			input:        "Hello World",
			expectedPass: true,
		},
		{
			name: "character_count_too_short",
			config: Config{
				Mode:     CountModeCharacters,
				Min:      10,
				Tripwire: true,
			},
			input:            "Hi",
			expectedPass:     false,
			expectedTripwire: true,
		},
		{
			name: "character_count_too_long",
			config: Config{
				Mode:     CountModeCharacters,
				Max:      5,
				Tripwire: true,
			},
			input:            "Hello World",
			expectedPass:     false,
			expectedTripwire: true,
		},
		{
			name: "word_count_within_limits",
			config: Config{
				Mode: CountModeWords,
				Min:  2,
				Max:  5,
			},
			input:        "Hello World Test",
			expectedPass: true,
		},
		{
			name: "word_count_too_few",
			config: Config{
				Mode:     CountModeWords,
				Min:      3,
				Tripwire: false,
			},
			input:            "Hello World",
			expectedPass:     false,
			expectedTripwire: false,
		},
		{
			name: "word_count_too_many",
			config: Config{
				Mode: CountModeWords,
				Max:  2,
			},
			input:        "Hello World Test Case",
			expectedPass: false,
		},
		{
			name: "line_count_within_limits",
			config: Config{
				Mode: CountModeLines,
				Min:  2,
				Max:  4,
			},
			input:        "Line 1\nLine 2\nLine 3",
			expectedPass: true,
		},
		{
			name: "empty_string_with_words_mode",
			config: Config{
				Mode: CountModeWords,
				Min:  1,
			},
			input:        "",
			expectedPass: false,
		},
		{
			name: "empty_string_with_characters_mode",
			config: Config{
				Mode: CountModeCharacters,
				Min:  0,
				Max:  100,
			},
			input:        "",
			expectedPass: true,
		},
		{
			name: "whitespace_only_with_words_mode",
			config: Config{
				Mode: CountModeWords,
				Min:  1,
			},
			input:        "   \n\t  ",
			expectedPass: false,
		},
		{
			name: "single_line_with_lines_mode",
			config: Config{
				Mode: CountModeLines,
				Min:  1,
				Max:  1,
			},
			input:        "Single line",
			expectedPass: true,
		},
		{
			name: "unicode_characters",
			config: Config{
				Mode: CountModeCharacters,
				Max:  20,
			},
			input:        "Hello 世界 🌍",
			expectedPass: true,
		},
		{
			name: "default_mode_is_characters",
			config: Config{
				Min: 5,
				Max: 20,
			},
			input:        "Hello World",
			expectedPass: true,
		},
		{
			name: "no_limits_set_always_passes",
			config: Config{
				Mode: CountModeCharacters,
			},
			input:        "Any length string should pass",
			expectedPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			guard := NewLength(tt.config)

			if guard == nil {
				t.Fatal("NewLength returned nil")
			}

			ctx := context.Background()
			result, err := guard.Func(ctx, tt.input)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result == nil {
				t.Fatal("Result is nil")
			}

			if result.Passed != tt.expectedPass {
				t.Errorf("Expected Passed=%v, got %v. Message: %s",
					tt.expectedPass, result.Passed, result.Message)
			}

			if result.TripwireTriggered != tt.expectedTripwire {
				t.Errorf("Expected TripwireTriggered=%v, got %v",
					tt.expectedTripwire, result.TripwireTriggered)
			}

			// Verify metadata contains count
			if result.Metadata == nil {
				t.Error("Metadata is nil")
			} else if _, ok := result.Metadata["count"]; !ok {
				t.Error("Metadata missing 'count' field")
			}
		})
	}
}

func TestCountContent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		mode     CountMode
		expected int
	}{
		{
			name:     "count_characters",
			input:    "Hello",
			mode:     CountModeCharacters,
			expected: 5,
		},
		{
			name:     "count_words_simple",
			input:    "Hello World",
			mode:     CountModeWords,
			expected: 2,
		},
		{
			name:     "count_words_with_extra_whitespace",
			input:    "  Hello   World  ",
			mode:     CountModeWords,
			expected: 2,
		},
		{
			name:     "count_words_empty_string",
			input:    "",
			mode:     CountModeWords,
			expected: 0,
		},
		{
			name:     "count_lines_single",
			input:    "Hello",
			mode:     CountModeLines,
			expected: 1,
		},
		{
			name:     "count_lines_multiple",
			input:    "Line 1\nLine 2\nLine 3",
			mode:     CountModeLines,
			expected: 3,
		},
		{
			name:     "count_lines_empty",
			input:    "",
			mode:     CountModeLines,
			expected: 0,
		},
		{
			name:     "unicode_characters",
			input:    "Hello 世界",
			mode:     CountModeCharacters,
			expected: 12, // "Hello " (6) + 世界 (6 bytes in UTF-8) = 12 bytes
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := countContent(tt.input, tt.mode)
			if result != tt.expected {
				t.Errorf("Expected count=%d, got %d", tt.expected, result)
			}
		})
	}
}

func TestLengthGuardrail_Integration(t *testing.T) {
	// Test that guard can be properly integrated into a chain
	guard := NewLength(Config{
		Mode:     CountModeWords,
		Min:      2,
		Max:      10,
		Tripwire: true,
	})

	if guard.Name != "content_length" {
		t.Errorf("Expected name='content_length', got '%s'", guard.Name)
	}

	// Test valid input
	ctx := context.Background()
	result, err := guard.Func(ctx, "This is a test message")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result.Passed {
		t.Error("Expected valid input to pass")
	}

	// Test invalid input triggers tripwire
	result, err = guard.Func(ctx, "Short")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result.Passed {
		t.Error("Expected invalid input to fail")
	}
	if !result.TripwireTriggered {
		t.Error("Expected tripwire to be triggered")
	}
}
