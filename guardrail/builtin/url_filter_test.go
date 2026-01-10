package builtin

import (
	"context"
	"strings"
	"testing"
)

func TestURLFilterGuardrail_Blocklist(t *testing.T) {
	gr := NewURLFilterGuardrail(
		WithBlocklist("evil.com", "*.bad.org"),
	)

	tests := []struct {
		name       string
		input      string
		shouldPass bool
	}{
		{"no URLs", "Hello world", true},
		{"safe URL", "Visit https://example.com", true},
		{"blocked exact", "Go to evil.com", false},
		{"blocked wildcard", "Check sub.bad.org", false},
		{"blocked with scheme", "Visit https://evil.com/page", false},
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

func TestURLFilterGuardrail_Allowlist(t *testing.T) {
	gr := NewURLFilterGuardrail(
		WithAllowlist("example.com", "*.safe.org"),
	)

	tests := []struct {
		name       string
		input      string
		shouldPass bool
	}{
		{"no URLs", "Hello world", true},
		{"allowed exact", "Visit example.com", true},
		{"allowed wildcard", "Check sub.safe.org", true},
		{"not in allowlist", "Visit other.com", false},
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

func TestURLFilterGuardrail_Tripwire(t *testing.T) {
	tests := []struct {
		name            string
		tripwire        bool
		input           string
		expectTriggered bool
	}{
		{"tripwire enabled with blocked URL", true, "evil.com", true},
		{"tripwire disabled with blocked URL", false, "evil.com", false},
		{"no blocked URLs", true, "safe.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gr := NewURLFilterGuardrail(
				WithBlocklist("evil.com"),
				WithURLTripwire(tt.tripwire),
			)
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

func TestRegexGuardrail_MustNotMatch(t *testing.T) {
	// Pattern that should NOT appear in input
	gr := NewRegexGuardrail(
		`\b(password|secret|token)\b`,
		WithMustMatch(false),
	)

	tests := []struct {
		name       string
		input      string
		shouldPass bool
	}{
		{"no forbidden words", "Hello world", true},
		{"has password", "My password is 123", false},
		{"has secret", "The secret code", false},
		{"has token", "API token here", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := gr.Func(context.Background(), tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.Passed != tt.shouldPass {
				t.Errorf("expected passed=%v, got %v for: %s", tt.shouldPass, result.Passed, tt.input)
			}

			if !tt.shouldPass && !strings.Contains(result.Message, "forbidden") {
				t.Errorf("expected message to mention 'forbidden', got: %s", result.Message)
			}
		})
	}
}

func TestRegexGuardrail_MustMatch(t *testing.T) {
	// Pattern that MUST appear in input
	gr := NewRegexGuardrail(
		`^[A-Z][a-z]+$`, // Must be a capitalized word
		WithMustMatch(true),
	)

	tests := []struct {
		name       string
		input      string
		shouldPass bool
	}{
		{"valid format", "Hello", true},
		{"lowercase", "hello", false},
		{"multiple words", "Hello World", false},
		{"has numbers", "Hello123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := gr.Func(context.Background(), tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.Passed != tt.shouldPass {
				t.Errorf("expected passed=%v, got %v for: %s", tt.shouldPass, result.Passed, tt.input)
			}

			if !tt.shouldPass && !strings.Contains(result.Message, "required") {
				t.Errorf("expected message to mention 'required', got: %s", result.Message)
			}
		})
	}
}

func TestRegexGuardrail_CustomMessage(t *testing.T) {
	customMsg := "Input contains profanity"
	gr := NewRegexGuardrail(
		`\b(badword1|badword2)\b`,
		WithRegexMessage(customMsg),
	)

	result, err := gr.Func(context.Background(), "This has badword1 in it")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Passed {
		t.Error("expected to fail")
	}

	if result.Message != customMsg {
		t.Errorf("expected custom message '%s', got '%s'", customMsg, result.Message)
	}
}
