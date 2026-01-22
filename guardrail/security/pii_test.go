package security

import (
	"context"
	"regexp"
	"testing"
)

func TestNewPII(t *testing.T) {
	tests := []struct {
		name             string
		opts             []PIIOption
		input            string
		expectedPass     bool
		expectedTripwire bool
	}{
		{
			name:         "no_pii_detected_default_config",
			opts:         nil,
			input:        "Hello, how can I help you today?",
			expectedPass: true,
		},
		{
			name:             "email_detected_default_config_tripwire_on",
			opts:             nil, // Default is all on + tripwire
			input:            "Contact me at john@example.com",
			expectedPass:     false,
			expectedTripwire: true,
		},
		{
			name:             "email_detected_explicit_tripwire",
			opts:             []PIIOption{WithEmailDetection(true), WithTripwire(true)},
			input:            "Contact me at john@example.com",
			expectedPass:     false,
			expectedTripwire: true,
		},
		{
			name:         "email_in_text_detection_disabled",
			opts:         []PIIOption{WithEmailDetection(false)},
			input:        "Email: test@domain.com",
			expectedPass: true,
		},
		{
			name:             "phone_number_detected",
			opts:             []PIIOption{WithPhoneDetection(true)},
			input:            "Call me at 555-123-4567",
			expectedPass:     false,
			expectedTripwire: true,
		},
		{
			name:             "phone_number_with_parentheses",
			opts:             []PIIOption{WithPhoneDetection(true)},
			input:            "My number is (555) 123-4567",
			expectedPass:     false,
			expectedTripwire: true,
		},
		{
			name:             "ssn_detected",
			opts:             []PIIOption{WithSSNDetection(true)},
			input:            "SSN: 123-45-6789",
			expectedPass:     false,
			expectedTripwire: true,
		},
		{
			name:             "credit_card_detected",
			opts:             []PIIOption{WithCreditCardDetection(true)},
			input:            "Card: 4532-1234-5678-9010",
			expectedPass:     false,
			expectedTripwire: true,
		},
		{
			name:         "credit_card_format_without_detection",
			opts:         []PIIOption{WithCreditCardDetection(false)},
			input:        "Card: 4532-1234-5678-9010",
			expectedPass: true,
		},
		{
			name:             "multiple_pii_types",
			opts:             []PIIOption{WithEmailDetection(true), WithPhoneDetection(true)},
			input:            "Email: user@example.com Phone: 555-1234", // 555-1234 might not match strict phone pattern
			expectedPass:     false,
			expectedTripwire: true,
		},
		{
			name:             "custom_pii_pattern",
			opts:             []PIIOption{WithCustomPattern("passport", regexp.MustCompile(`[A-Z]{2}\d{7}`))},
			input:            "Passport: AB1234567",
			expectedPass:     false,
			expectedTripwire: true,
		},
		{
			name: "explicitly_disable_all_detection",
			opts: []PIIOption{
				WithEmailDetection(false),
				WithPhoneDetection(false),
				WithSSNDetection(false),
				WithCreditCardDetection(false),
			},
			input:        "user@example.com and 555-123-4567",
			expectedPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			guard := NewPII(tt.opts...)

			if guard == nil {
				t.Fatal("NewPII returned nil")
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

func TestPIIOptions(t *testing.T) {
	t.Run("WithEmailDetection", func(t *testing.T) {
		config := &PIIConfig{}
		WithEmailDetection(true)(config)
		if !config.DetectEmail {
			t.Error("Email detection should be enabled")
		}
	})

	t.Run("WithPhoneDetection", func(t *testing.T) {
		config := &PIIConfig{}
		WithPhoneDetection(true)(config)
		if !config.DetectPhone {
			t.Error("Phone detection should be enabled")
		}
	})

	t.Run("WithSSNDetection", func(t *testing.T) {
		config := &PIIConfig{}
		WithSSNDetection(true)(config)
		if !config.DetectSSN {
			t.Error("SSN detection should be enabled")
		}
	})

	t.Run("WithCreditCardDetection", func(t *testing.T) {
		config := &PIIConfig{}
		WithCreditCardDetection(true)(config)
		if !config.DetectCreditCard {
			t.Error("Credit card detection should be enabled")
		}
	})

	t.Run("WithTripwire", func(t *testing.T) {
		config := &PIIConfig{}
		WithTripwire(true)(config)
		if !config.Tripwire {
			t.Error("Tripwire should be enabled")
		}
	})

	t.Run("WithCustomPattern", func(t *testing.T) {
		config := &PIIConfig{}
		WithCustomPattern("test", regexp.MustCompile(`test`))(config)
		if len(config.CustomPatterns) != 1 {
			t.Error("Custom pattern should be added")
		}
	})
}
