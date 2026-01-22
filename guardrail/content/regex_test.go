package content

import (
	"context"
	"testing"
)

func TestNewRegex(t *testing.T) {
	tests := []struct {
		name             string
		pattern          string
		opts             []RegexOption
		input            string
		expectedPass     bool
		expectedTripwire bool
		expectPanic      bool
	}{
		{
			name:         "must_not_match_forbidden_pattern_not_present",
			pattern:      `\b(password|secret)\b`,
			opts:         []RegexOption{WithMustMatch(false)},
			input:        "Hello World",
			expectedPass: true,
		},
		{
			name:             "must_not_match_forbidden_pattern_present",
			pattern:          `\b(password|secret)\b`,
			opts:             []RegexOption{WithMustMatch(false), WithRegexTripwire(true)},
			input:            "My password is 123",
			expectedPass:     false,
			expectedTripwire: true,
		},
		{
			name:         "must_match_required_pattern_present",
			pattern:      `^[A-Z]`,
			opts:         []RegexOption{WithMustMatch(true)},
			input:        "Hello",
			expectedPass: true,
		},
		{
			name:             "must_match_required_pattern_absent",
			pattern:          `^[A-Z]`,
			opts:             []RegexOption{WithMustMatch(true), WithRegexTripwire(true)},
			input:            "hello",
			expectedPass:     false,
			expectedTripwire: true,
		},
		{
			name:             "custom_message_on_failure",
			pattern:          `^\d+$`,
			opts:             []RegexOption{WithMustMatch(true), WithRegexMessage("Input must be numbers only"), WithRegexTripwire(false)},
			input:            "abc123",
			expectedPass:     false,
			expectedTripwire: false,
		},
		{
			name:             "email_pattern_validation",
			pattern:          `\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`,
			opts:             []RegexOption{WithMustMatch(false), WithRegexTripwire(true)},
			input:            "Contact: user@example.com",
			expectedPass:     false,
			expectedTripwire: true,
		},
		{
			name:         "case_sensitive_pattern",
			pattern:      `SECRET`,
			opts:         []RegexOption{WithMustMatch(false)},
			input:        "secret",
			expectedPass: true,
		},
		{
			name:             "multiline_pattern",
			pattern:          `(?m)^confidential$`,
			opts:             []RegexOption{WithMustMatch(false), WithRegexTripwire(false)},
			input:            "This is\nconfidential\ninformation",
			expectedPass:     false,
			expectedTripwire: false,
		},
		{
			name:             "empty_input",
			pattern:          `\S+`,
			opts:             []RegexOption{WithMustMatch(true), WithRegexTripwire(false)},
			input:            "",
			expectedPass:     false,
			expectedTripwire: false,
		},
		{
			name:             "unicode_in_pattern",
			pattern:          `[你好世界]`,
			opts:             []RegexOption{WithMustMatch(false), WithRegexTripwire(false)},
			input:            "Hello 世界",
			expectedPass:     false,
			expectedTripwire: false,
		},
		{
			name:             "default_options_must_not_match_with_tripwire",
			pattern:          `danger`,
			input:            "This contains danger",
			expectedPass:     false,
			expectedTripwire: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Error("Expected panic but didn't get one")
					}
				}()
			}

			guard := NewRegex(tt.pattern, tt.opts...)

			if guard == nil {
				t.Fatal("NewRegex returned nil")
			}

			ctx := context.Background()
			result, err := guard.Func(ctx, tt.input)

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

			// Verify metadata
			if result.Metadata != nil {
				if pattern, ok := result.Metadata["pattern"]; ok {
					if pattern != tt.pattern {
						t.Errorf("Metadata pattern mismatch")
					}
				}
			}
		})
	}
}

func TestRegexOptions(t *testing.T) {
	t.Run("WithRegexTripwire", func(t *testing.T) {
		config := &RegexConfig{}
		WithRegexTripwire(true)(config)
		if !config.Tripwire {
			t.Error("Tripwire should be true")
		}

		WithRegexTripwire(false)(config)
		if config.Tripwire {
			t.Error("Tripwire should be false")
		}
	})

	t.Run("WithMustMatch", func(t *testing.T) {
		config := &RegexConfig{}
		WithMustMatch(true)(config)
		if !config.MustMatch {
			t.Error("MustMatch should be true")
		}

		WithMustMatch(false)(config)
		if config.MustMatch {
			t.Error("MustMatch should be false")
		}
	})

	t.Run("WithRegexMessage", func(t *testing.T) {
		config := &RegexConfig{}
		msg := "Custom error message"
		WithRegexMessage(msg)(config)
		if config.Message != msg {
			t.Errorf("Expected message='%s', got '%s'", msg, config.Message)
		}
	})
}

func TestRegexGuardrail_InvalidPattern(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for invalid regex pattern")
		}
	}()

	// This should panic because of invalid regex
	NewRegex(`[invalid(`)
}

func TestRegexGuardrail_Integration(t *testing.T) {
	// Test blocking forbidden keywords
	guard := NewRegex(
		`\b(password|secret|token|key)\b`,
		WithMustMatch(false),
		WithRegexTripwire(true),
		WithRegexMessage("Please don't share sensitive credentials"),
	)

	if guard.Name != "regex_validation" {
		t.Errorf("Expected name='regex_validation', got '%s'", guard.Name)
	}

	// Test valid input (no forbidden keywords)
	ctx := context.Background()
	result, err := guard.Func(ctx, "Hello, how can I help you?")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result.Passed {
		t.Error("Expected valid input to pass")
	}

	// Test invalid input (contains forbidden keyword)
	result, err = guard.Func(ctx, "What is my password?")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result.Passed {
		t.Error("Expected invalid input to fail")
	}
	if !result.TripwireTriggered {
		t.Error("Expected tripwire to be triggered")
	}
	if result.Message != "Please don't share sensitive credentials" {
		t.Errorf("Expected custom message, got: %s", result.Message)
	}
}

func TestRegexGuardrail_MustMatchScenarios(t *testing.T) {
	// Test requiring specific format
	guard := NewRegex(
		`^[A-Z][a-z]+ [A-Z][a-z]+$`, // Full name format
		WithMustMatch(true),
		WithRegexMessage("Please provide your full name in proper format"),
	)

	ctx := context.Background()

	// Valid full name
	result, _ := guard.Func(ctx, "John Doe")
	if !result.Passed {
		t.Error("Valid full name should pass")
	}

	// Invalid format
	result, _ = guard.Func(ctx, "john")
	if result.Passed {
		t.Error("Invalid format should fail")
	}
	if result.Message != "Please provide your full name in proper format" {
		t.Errorf("Unexpected message: %s", result.Message)
	}
}
