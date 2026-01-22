package moderation

import (
	"context"
	"testing"
)

func TestNewProfanity(t *testing.T) {
	tests := []struct {
		name             string
		config           ProfanityConfig
		input            string
		expectedPass     bool
		expectedTripwire bool
	}{
		{
			name:         "clean_text_passes",
			config:       ProfanityConfig{},
			input:        "Hello world, this is a clean message",
			expectedPass: true,
		},
		{
			name:             "profanity_detected_default_severe_word",
			config:           ProfanityConfig{Tripwire: true},
			input:            "This is fucking crazy", // Use a word definitely in the default list
			expectedPass:     false,
			expectedTripwire: true,
		},
		{
			name:             "custom_blocked_word",
			config:           ProfanityConfig{WordList: map[string]SeverityLevel{"custombadword": SeverityHigh}},
			input:            "Don't use the word custombadword please",
			expectedPass:     false,
			expectedTripwire: false,
		},
		{
			name:             "severity_filtering_ignore_low_severity",
			config:           ProfanityConfig{MinSeverity: SeverityMedium, WordList: map[string]SeverityLevel{"sucks": SeverityLow}},
			input:            "This sucks",
			expectedPass:     true,
			expectedTripwire: false,
		},
		// Leetspeak test with supported characters (0, 1, 3, 4, 5, 7, 8, @, $, !)
		{
			name:             "leetspeak_detection",
			config:           ProfanityConfig{WordList: map[string]SeverityLevel{"bitch": SeverityHigh}, NormalizeLeetspeak: true},
			input:            "You are a b!tch",
			expectedPass:     false,
			expectedTripwire: false,
		},
		{
			name:             "multiple_bad_words",
			config:           ProfanityConfig{Tripwire: true, WordList: map[string]SeverityLevel{"damn": SeverityMedium, "shit": SeverityHigh}},
			input:            "Damn, this is shit",
			expectedPass:     false,
			expectedTripwire: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			guard := NewProfanity(tt.config)

			if guard == nil {
				t.Fatal("NewProfanity returned nil")
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
		})
	}
}
