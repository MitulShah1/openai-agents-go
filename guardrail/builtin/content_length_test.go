package builtin

import (
	"context"
	"strings"
	"testing"

	"github.com/MitulShah1/openai-agents-go/guardrail"
)

func TestContentLengthGuardrail_Characters(t *testing.T) {
	t.Run("within limits", func(t *testing.T) {
		guard := NewContentLengthGuardrail(ContentLengthConfig{
			Mode: CountModeCharacters,
			Min:  5,
			Max:  20,
		})

		ctx := context.Background()

		// Valid input
		err := guard.Validate(ctx, "Hello, World!")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("below minimum", func(t *testing.T) {
		guard := NewContentLengthGuardrail(ContentLengthConfig{
			Mode: CountModeCharacters,
			Min:  10,
		})

		ctx := context.Background()

		err := guard.Validate(ctx, "Short")
		if err == nil {
			t.Error("expected error for content below minimum")
		}
	})

	t.Run("above maximum", func(t *testing.T) {
		guard := NewContentLengthGuardrail(ContentLengthConfig{
			Mode: CountModeCharacters,
			Max:  10,
		})

		ctx := context.Background()

		err := guard.Validate(ctx, "This is a very long string")
		if err == nil {
			t.Error("expected error for content above maximum")
		}
	})

	t.Run("exact minimum", func(t *testing.T) {
		guard := NewContentLengthGuardrail(ContentLengthConfig{
			Mode: CountModeCharacters,
			Min:  5,
		})

		ctx := context.Background()

		err := guard.Validate(ctx, "Hello")
		if err != nil {
			t.Errorf("unexpected error for exact minimum: %v", err)
		}
	})

	t.Run("exact maximum", func(t *testing.T) {
		guard := NewContentLengthGuardrail(ContentLengthConfig{
			Mode: CountModeCharacters,
			Max:  5,
		})

		ctx := context.Background()

		err := guard.Validate(ctx, "Hello")
		if err != nil {
			t.Errorf("unexpected error for exact maximum: %v", err)
		}
	})

	t.Run("empty input with no minimum", func(t *testing.T) {
		guard := NewContentLengthGuardrail(ContentLengthConfig{
			Mode: CountModeCharacters,
			Max:  100,
		})

		ctx := context.Background()

		err := guard.Validate(ctx, "")
		if err != nil {
			t.Errorf("unexpected error for empty input: %v", err)
		}
	})
}

func TestContentLengthGuardrail_Words(t *testing.T) {
	t.Run("counts words correctly", func(t *testing.T) {
		guard := NewContentLengthGuardrail(ContentLengthConfig{
			Mode: CountModeWords,
			Min:  3,
			Max:  5,
		})

		ctx := context.Background()

		// 4 words - valid
		err := guard.Validate(ctx, "This is a test")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("too few words", func(t *testing.T) {
		guard := NewContentLengthGuardrail(ContentLengthConfig{
			Mode: CountModeWords,
			Min:  5,
		})

		ctx := context.Background()

		// 2 words
		err := guard.Validate(ctx, "Hello World")
		if err == nil {
			t.Error("expected error for too few words")
		}
	})

	t.Run("too many words", func(t *testing.T) {
		guard := NewContentLengthGuardrail(ContentLengthConfig{
			Mode: CountModeWords,
			Max:  3,
		})

		ctx := context.Background()

		// 5 words
		err := guard.Validate(ctx, "This is a long sentence")
		if err == nil {
			t.Error("expected error for too many words")
		}
	})

	t.Run("empty input", func(t *testing.T) {
		guard := NewContentLengthGuardrail(ContentLengthConfig{
			Mode: CountModeWords,
			Min:  1,
		})

		ctx := context.Background()

		err := guard.Validate(ctx, "")
		if err == nil {
			t.Error("expected error for empty input with word minimum")
		}
	})

	t.Run("whitespace only", func(t *testing.T) {
		guard := NewContentLengthGuardrail(ContentLengthConfig{
			Mode: CountModeWords,
			Min:  1,
		})

		ctx := context.Background()

		err := guard.Validate(ctx, "   \t\n   ")
		if err == nil {
			t.Error("expected error for whitespace-only input")
		}
	})
}

func TestContentLengthGuardrail_Lines(t *testing.T) {
	t.Run("counts lines correctly", func(t *testing.T) {
		guard := NewContentLengthGuardrail(ContentLengthConfig{
			Mode: CountModeLines,
			Min:  2,
			Max:  4,
		})

		ctx := context.Background()

		// 3 lines - valid
		err := guard.Validate(ctx, "Line 1\nLine 2\nLine 3")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("too few lines", func(t *testing.T) {
		guard := NewContentLengthGuardrail(ContentLengthConfig{
			Mode: CountModeLines,
			Min:  3,
		})

		ctx := context.Background()

		// 1 line
		err := guard.Validate(ctx, "Single line")
		if err == nil {
			t.Error("expected error for too few lines")
		}
	})

	t.Run("too many lines", func(t *testing.T) {
		guard := NewContentLengthGuardrail(ContentLengthConfig{
			Mode: CountModeLines,
			Max:  2,
		})

		ctx := context.Background()

		// 3 lines
		err := guard.Validate(ctx, "Line 1\nLine 2\nLine 3")
		if err == nil {
			t.Error("expected error for too many lines")
		}
	})

	t.Run("empty input", func(t *testing.T) {
		guard := NewContentLengthGuardrail(ContentLengthConfig{
			Mode: CountModeLines,
			Min:  1,
		})

		ctx := context.Background()

		err := guard.Validate(ctx, "")
		if err == nil {
			t.Error("expected error for empty input with line minimum")
		}
	})

	t.Run("single line with newline", func(t *testing.T) {
		guard := NewContentLengthGuardrail(ContentLengthConfig{
			Mode: CountModeLines,
			Min:  1,
			Max:  2,
		})

		ctx := context.Background()

		// 2 lines (second is empty after newline)
		err := guard.Validate(ctx, "Line 1\n")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestContentLengthGuardrail_DefaultMode(t *testing.T) {
	// Should default to character mode
	guard := NewContentLengthGuardrail(ContentLengthConfig{
		Min: 5,
		Max: 10,
	})

	ctx := context.Background()

	// 7 characters - valid
	err := guard.Validate(ctx, "Testing")
	if err != nil {
		t.Errorf("unexpected error with default mode: %v", err)
	}
}

func TestContentLengthGuardrail_Tripwire(t *testing.T) {
	guard := NewContentLengthGuardrail(ContentLengthConfig{
		Mode:     CountModeCharacters,
		Max:      5,
		Tripwire: true,
	})

	if !guard.IsTripwire() {
		t.Error("guardrail should be marked as tripwire")
	}

	ctx := context.Background()

	err := guard.Validate(ctx, "Too long content")
	if err == nil {
		t.Error("expected tripwire error")
	}

	// Check it's the right error type
	if _, ok := err.(*guardrail.InputGuardrailTripwireError); !ok {
		t.Errorf("expected InputGuardrailTripwireError, got %T", err)
	}
}

func TestContentLengthGuardrail_Name(t *testing.T) {
	guard := NewContentLengthGuardrail(ContentLengthConfig{})

	if guard.Name() != "content_length" {
		t.Errorf("expected name 'content_length', got '%s'", guard.Name())
	}
}

func TestContentLengthGuardrail_NoLimits(t *testing.T) {
	// No min or max - should always pass
	guard := NewContentLengthGuardrail(ContentLengthConfig{
		Mode: CountModeCharacters,
	})

	ctx := context.Background()

	testCases := []string{
		"",
		"short",
		"a very long string with lots of characters",
		strings.Repeat("x", 10000),
	}

	for _, tc := range testCases {
		err := guard.Validate(ctx, tc)
		if err != nil {
			t.Errorf("unexpected error for no-limits guardrail: %v", err)
		}
	}
}
