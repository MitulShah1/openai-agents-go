package builtin

import (
	"context"
	"regexp"
	"strings"
	"testing"
)

func TestPIIGuardrail_Email(t *testing.T) {
	gr := NewPIIGuardrail(
		WithEmailDetection(true),
		WithPhoneDetection(false),
		WithSSNDetection(false),
		WithCreditCardDetection(false),
	)

	tests := []struct {
		name       string
		input      string
		shouldPass bool
	}{
		{"no email", "Hello world", true},
		{"has email", "Contact me at test@example.com", false},
		{"multiple emails", "Send to foo@bar.com or baz@qux.org", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := gr.Func(context.Background(), tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.Passed != tt.shouldPass {
				t.Errorf("expected passed=%v, got %v", tt.shouldPass, result.Passed)
			}

			if !tt.shouldPass && !strings.Contains(result.Message, "email") {
				t.Errorf("expected message to mention 'email', got: %s", result.Message)
			}
		})
	}
}

func TestPIIGuardrail_Phone(t *testing.T) {
	gr := NewPIIGuardrail(
		WithEmailDetection(false),
		WithPhoneDetection(true),
		WithSSNDetection(false),
		WithCreditCardDetection(false),
	)

	tests := []struct {
		name       string
		input      string
		shouldPass bool
	}{
		{"no phone", "Hello world", true},
		{"has phone", "Call me at 555-123-4567", false},
		{"phone with parens", "Call (555) 123-4567", false},
		{"phone no dashes", "Call 5551234567", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := gr.Func(context.Background(), tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.Passed != tt.shouldPass {
				t.Errorf("expected passed=%v, got %v for input: %s", tt.shouldPass, result.Passed, tt.input)
			}
		})
	}
}

func TestPIIGuardrail_SSN(t *testing.T) {
	gr := NewPIIGuardrail(
		WithEmailDetection(false),
		WithPhoneDetection(false),
		WithSSNDetection(true),
		WithCreditCardDetection(false),
	)

	tests := []struct {
		name       string
		input      string
		shouldPass bool
	}{
		{"no SSN", "Hello world", true},
		{"has SSN", "My SSN is 123-45-6789", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := gr.Func(context.Background(), tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.Passed != tt.shouldPass {
				t.Errorf("expected passed=%v, got %v", tt.shouldPass, result.Passed)
			}
		})
	}
}

func TestPIIGuardrail_CreditCard(t *testing.T) {
	gr := NewPIIGuardrail(
		WithEmailDetection(false),
		WithPhoneDetection(false),
		WithSSNDetection(false),
		WithCreditCardDetection(true),
	)

	tests := []struct {
		name       string
		input      string
		shouldPass bool
	}{
		{"no card", "Hello world", true},
		{"has card with spaces", "Card: 1234 5678 9012 3456", false},
		{"has card with dashes", "Card: 1234-5678-9012-3456", false},
		{"has card no separators", "Card: 1234567890123456", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := gr.Func(context.Background(), tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.Passed != tt.shouldPass {
				t.Errorf("expected passed=%v, got %v", tt.shouldPass, result.Passed)
			}
		})
	}
}

func TestPIIGuardrail_CustomPattern(t *testing.T) {
	customPattern := regexp.MustCompile(`\b[A-Z]{3}-\d{3}\b`)
	gr := NewPIIGuardrail(
		WithEmailDetection(false),
		WithPhoneDetection(false),
		WithSSNDetection(false),
		WithCreditCardDetection(false),
		WithCustomPattern("employee_id", customPattern),
	)

	tests := []struct {
		name       string
		input      string
		shouldPass bool
	}{
		{"no pattern", "Hello world", true},
		{"has pattern", "Employee ABC-123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := gr.Func(context.Background(), tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.Passed != tt.shouldPass {
				t.Errorf("expected passed=%v, got %v", tt.shouldPass, result.Passed)
			}
		})
	}
}

func TestPIIGuardrail_Tripwire(t *testing.T) {
	tests := []struct {
		name            string
		tripwire        bool
		input           string
		expectTriggered bool
	}{
		{"tripwire enabled with PII", true, "Email: test@example.com", true},
		{"tripwire disabled with PII", false, "Email: test@example.com", false},
		{"no PII", true, "Hello world", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gr := NewPIIGuardrail(WithTripwire(tt.tripwire))
			result, err := gr.Func(context.Background(), tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.TripwireTriggered != tt.expectTriggered {
				t.Errorf("expected tripwire=%v, got %v", tt.expectTriggered, result.TripwireTriggered)
			}
		})
	}
}

func TestPIIGuardrail_MultipleDetections(t *testing.T) {
	gr := NewPIIGuardrail() // All detection enabled by default

	input := "Contact john@example.com at 555-1234 or use SSN 123-45-6789"
	result, err := gr.Func(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Passed {
		t.Error("expected to fail with multiple PII types")
	}

	detected, ok := result.Metadata["detected_types"].([]string)
	if !ok {
		t.Fatal("expected detected_types in metadata")
	}

	if len(detected) < 2 {
		t.Errorf("expected at least 2 detections, got %d: %v", len(detected), detected)
	}
}
