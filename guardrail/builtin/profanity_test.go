package builtin

import (
	"context"
	"testing"

	"github.com/MitulShah1/openai-agents-go/guardrail"
)

func TestProfanityGuardrail_Basic(t *testing.T) {
	t.Run("clean text passes", func(t *testing.T) {
		guard := NewProfanityGuardrail(ProfanityConfig{})

		ctx := context.Background()
		err := guard.Validate(ctx, "This is a nice and friendly message")
		if err != nil {
			t.Errorf("unexpected error for clean text: %v", err)
		}
	})

	t.Run("profanity detected", func(t *testing.T) {
		guard := NewProfanityGuardrail(ProfanityConfig{})

		ctx := context.Background()
		err := guard.Validate(ctx, "This is a shit message")
		if err == nil {
			t.Error("expected error for profane text")
		}
	})

	t.Run("multiple profanities", func(t *testing.T) {
		guard := NewProfanityGuardrail(ProfanityConfig{})

		ctx := context.Background()
		err := guard.Validate(ctx, "What the fuck is this shit")
		if err == nil {
			t.Error("expected error for multiple profanities")
		}
	})
}

func TestProfanityGuardrail_Leetspeak(t *testing.T) {
	t.Run("detects leetspeak profanity", func(t *testing.T) {
		guard := NewProfanityGuardrail(ProfanityConfig{
			NormalizeLeetspeak: true,
		})

		ctx := context.Background()

		// "shit" in leetspeak: "5h1t" or "$h!t"
		err := guard.Validate(ctx, "This is 5h1t")
		if err == nil {
			t.Error("expected error for leetspeak profanity")
		}
	})

	t.Run("leetspeak normalization disabled", func(t *testing.T) {
		guard := NewProfanityGuardrail(ProfanityConfig{
			NormalizeLeetspeak: false,
		})

		ctx := context.Background()

		// Should not detect leetspeak when normalization is off
		err := guard.Validate(ctx, "This is 5h1t")
		if err != nil {
			t.Errorf("unexpected error with leetspeak normalization disabled: %v", err)
		}
	})

	t.Run("various leetspeak patterns", func(t *testing.T) {
		guard := NewProfanityGuardrail(ProfanityConfig{
			NormalizeLeetspeak: true,
		})

		ctx := context.Background()

		testCases := []struct {
			name       string
			input      string
			shouldFail bool
		}{
			{"@ for a", "wh@t the hell", true},
			{"$ for s", "thi$ i$ crap", true},
			{"! for i", "th!s !s sh!t", true},
			{"3 for e", "h3ll", true},
			{"4 for a", "d4mn", true},
			{"0 for o", "0h n0", false}, // "oh no" is clean
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := guard.Validate(ctx, tc.input)
				if tc.shouldFail && err == nil {
					t.Errorf("expected error for %q", tc.input)
				}
				if !tc.shouldFail && err != nil {
					t.Errorf("unexpected error for %q: %v", tc.input, err)
				}
			})
		}
	})
}

func TestProfanityGuardrail_SeverityLevels(t *testing.T) {
	t.Run("blocks only high severity", func(t *testing.T) {
		guard := NewProfanityGuardrail(ProfanityConfig{
			MinSeverity: SeverityHigh,
		})

		ctx := context.Background()

		// Low severity - should pass
		err := guard.Validate(ctx, "You are stupid")
		if err != nil {
			t.Errorf("unexpected error for low severity: %v", err)
		}

		// Medium severity - should pass
		err = guard.Validate(ctx, "What the hell")
		if err != nil {
			t.Errorf("unexpected error for medium severity: %v", err)
		}

		// High severity - should fail
		err = guard.Validate(ctx, "What the fuck")
		if err == nil {
			t.Error("expected error for high severity")
		}
	})

	t.Run("blocks medium and above", func(t *testing.T) {
		guard := NewProfanityGuardrail(ProfanityConfig{
			MinSeverity: SeverityMedium,
		})

		ctx := context.Background()

		// Low severity - should pass
		err := guard.Validate(ctx, "You idiot")
		if err != nil {
			t.Errorf("unexpected error for low severity: %v", err)
		}

		// Medium severity - should fail
		err = guard.Validate(ctx, "This is crap")
		if err == nil {
			t.Error("expected error for medium severity")
		}
	})

	t.Run("blocks all severities (default)", func(t *testing.T) {
		guard := NewProfanityGuardrail(ProfanityConfig{})

		ctx := context.Background()

		// All should fail
		testCases := []string{
			"You stupid idiot",    // Low
			"What the hell",       // Medium
			"This is fucking bad", // High
		}

		for _, tc := range testCases {
			err := guard.Validate(ctx, tc)
			if err == nil {
				t.Errorf("expected error for %q", tc)
			}
		}
	})
}

func TestProfanityGuardrail_CustomWordList(t *testing.T) {
	t.Run("custom word list", func(t *testing.T) {
		customWords := map[string]SeverityLevel{
			"banana": SeverityHigh,
			"apple":  SeverityMedium,
		}

		guard := NewProfanityGuardrail(ProfanityConfig{
			WordList: customWords,
		})

		ctx := context.Background()

		// Custom word should be detected (exact match)
		err := guard.Validate(ctx, "I hate banana")
		if err == nil {
			t.Error("expected error for custom profane word")
		}

		// Standard profanity should not be detected
		err = guard.Validate(ctx, "This is shit")
		if err != nil {
			t.Errorf("unexpected error for standard profanity with custom list: %v", err)
		}
	})
}

func TestProfanityGuardrail_CaseInsensitive(t *testing.T) {
	guard := NewProfanityGuardrail(ProfanityConfig{})

	ctx := context.Background()

	testCases := []string{
		"FUCK",
		"Fuck",
		"FuCk",
		"fuck",
	}

	for _, tc := range testCases {
		err := guard.Validate(ctx, tc)
		if err == nil {
			t.Errorf("expected error for %q (case insensitive)", tc)
		}
	}
}

func TestProfanityGuardrail_WordBoundaries(t *testing.T) {
	guard := NewProfanityGuardrail(ProfanityConfig{})

	ctx := context.Background()

	t.Run("word as standalone", func(t *testing.T) {
		err := guard.Validate(ctx, "This is shit")
		if err == nil {
			t.Error("expected error for standalone word")
		}
	})

	t.Run("word with punctuation", func(t *testing.T) {
		// The `\w+` regex will extract "hell" from "hell!"
		err := guard.Validate(ctx, "What the hell!")
		if err == nil {
			t.Error("expected error for word with punctuation")
		}
	})

	t.Run("word in middle of text", func(t *testing.T) {
		err := guard.Validate(ctx, "The damn thing broke")
		if err == nil {
			t.Error("expected error for word in middle")
		}
	})
}

func TestProfanityGuardrail_Tripwire(t *testing.T) {
	guard := NewProfanityGuardrail(ProfanityConfig{
		Tripwire: true,
	})

	if !guard.IsTripwire() {
		t.Error("guardrail should be marked as tripwire")
	}

	ctx := context.Background()
	err := guard.Validate(ctx, "This is shit")
	if err == nil {
		t.Error("expected tripwire error")
	}

	// Check it's the right error type
	if _, ok := err.(*guardrail.InputGuardrailTripwireError); !ok {
		t.Errorf("expected InputGuardrailTripwireError, got %T", err)
	}
}

func TestProfanityGuardrail_Name(t *testing.T) {
	guard := NewProfanityGuardrail(ProfanityConfig{})

	if guard.Name() != "profanity" {
		t.Errorf("expected name 'profanity', got '%s'", guard.Name())
	}
}

func TestProfanityGuardrail_EmptyInput(t *testing.T) {
	guard := NewProfanityGuardrail(ProfanityConfig{})

	ctx := context.Background()
	err := guard.Validate(ctx, "")
	if err != nil {
		t.Errorf("unexpected error for empty input: %v", err)
	}
}

func TestProfanityGuardrail_NumbersOnly(t *testing.T) {
	guard := NewProfanityGuardrail(ProfanityConfig{})

	ctx := context.Background()
	err := guard.Validate(ctx, "12345 67890")
	if err != nil {
		t.Errorf("unexpected error for numbers only: %v", err)
	}
}
