package builtin

import (
	"context"
	"strings"
	"testing"
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
		result, err := guard.Func(ctx, "Hello, World!")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.Passed {
			t.Errorf("expected pass, got failure: %s", result.Message)
		}
	})

	t.Run("below minimum", func(t *testing.T) {
		guard := NewContentLengthGuardrail(ContentLengthConfig{
			Mode: CountModeCharacters,
			Min:  10,
		})

		ctx := context.Background()

		result, err := guard.Func(ctx, "Short")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Passed {
			t.Error("expected failure for content below minimum")
		}
	})

	t.Run("above maximum", func(t *testing.T) {
		guard := NewContentLengthGuardrail(ContentLengthConfig{
			Mode: CountModeCharacters,
			Max:  10,
		})

		ctx := context.Background()

		result, err := guard.Func(ctx, "This is a very long string")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Passed {
			t.Error("expected failure for content above maximum")
		}
	})

	t.Run("exact minimum", func(t *testing.T) {
		guard := NewContentLengthGuardrail(ContentLengthConfig{
			Mode: CountModeCharacters,
			Min:  5,
		})

		ctx := context.Background()

		result, err := guard.Func(ctx, "Hello")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.Passed {
			t.Error("expected pass for exact minimum")
		}
	})

	t.Run("exact maximum", func(t *testing.T) {
		guard := NewContentLengthGuardrail(ContentLengthConfig{
			Mode: CountModeCharacters,
			Max:  5,
		})

		ctx := context.Background()

		result, err := guard.Func(ctx, "Hello")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.Passed {
			t.Error("expected pass for exact maximum")
		}
	})

	t.Run("empty input with no minimum", func(t *testing.T) {
		guard := NewContentLengthGuardrail(ContentLengthConfig{
			Mode: CountModeCharacters,
			Max:  100,
		})

		ctx := context.Background()

		result, err := guard.Func(ctx, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.Passed {
			t.Error("expected pass for empty input")
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
		result, err := guard.Func(ctx, "This is a test")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.Passed {
			t.Error("expected pass")
		}
	})

	t.Run("too few words", func(t *testing.T) {
		guard := NewContentLengthGuardrail(ContentLengthConfig{
			Mode: CountModeWords,
			Min:  5,
		})

		ctx := context.Background()

		// 2 words
		result, err := guard.Func(ctx, "Hello World")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Passed {
			t.Error("expected failure for too few words")
		}
	})

	t.Run("too many words", func(t *testing.T) {
		guard := NewContentLengthGuardrail(ContentLengthConfig{
			Mode: CountModeWords,
			Max:  3,
		})

		ctx := context.Background()

		// 5 words
		result, err := guard.Func(ctx, "This is a long sentence")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Passed {
			t.Error("expected failure for too many words")
		}
	})

	t.Run("empty input", func(t *testing.T) {
		guard := NewContentLengthGuardrail(ContentLengthConfig{
			Mode: CountModeWords,
			Min:  1,
		})

		ctx := context.Background()

		result, err := guard.Func(ctx, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Passed {
			t.Error("expected failure for empty input with word minimum")
		}
	})

	t.Run("whitespace only", func(t *testing.T) {
		guard := NewContentLengthGuardrail(ContentLengthConfig{
			Mode: CountModeWords,
			Min:  1,
		})

		ctx := context.Background()

		result, err := guard.Func(ctx, "   \t\n   ")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Passed {
			t.Error("expected failure for whitespace-only input")
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
		result, err := guard.Func(ctx, "Line 1\nLine 2\nLine 3")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.Passed {
			t.Error("expected pass")
		}
	})

	t.Run("too few lines", func(t *testing.T) {
		guard := NewContentLengthGuardrail(ContentLengthConfig{
			Mode: CountModeLines,
			Min:  3,
		})

		ctx := context.Background()

		// 1 line
		result, err := guard.Func(ctx, "Single line")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Passed {
			t.Error("expected failure for too few lines")
		}
	})

	t.Run("too many lines", func(t *testing.T) {
		guard := NewContentLengthGuardrail(ContentLengthConfig{
			Mode: CountModeLines,
			Max:  2,
		})

		ctx := context.Background()

		// 3 lines
		result, err := guard.Func(ctx, "Line 1\nLine 2\nLine 3")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Passed {
			t.Error("expected failure for too many lines")
		}
	})

	t.Run("empty input", func(t *testing.T) {
		guard := NewContentLengthGuardrail(ContentLengthConfig{
			Mode: CountModeLines,
			Min:  1,
		})

		ctx := context.Background()

		result, err := guard.Func(ctx, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Passed {
			t.Error("expected failure for empty input with line minimum")
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
		result, err := guard.Func(ctx, "Line 1\n")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.Passed {
			t.Error("expected pass")
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
	result, err := guard.Func(ctx, "Testing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Passed {
		t.Error("expected pass")
	}
}

func TestContentLengthGuardrail_Tripwire(t *testing.T) {
	guard := NewContentLengthGuardrail(ContentLengthConfig{
		Mode:     CountModeCharacters,
		Max:      5,
		Tripwire: true,
	})

	// Check tripwire by triggering it
	ctx := context.Background()
	result, err := guard.Func(ctx, "Too long content")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Passed {
		t.Error("expected failure")
	}
	if !result.TripwireTriggered {
		t.Error("expected TripwireTriggered to be true")
	}
}

func TestContentLengthGuardrail_Name(t *testing.T) {
	guard := NewContentLengthGuardrail(ContentLengthConfig{})

	if guard.Name != "content_length" {
		t.Errorf("expected name 'content_length', got '%s'", guard.Name)
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
		result, err := guard.Func(ctx, tc)
		if err != nil {
			t.Errorf("unexpected error for no-limits guardrail: %v", err)
		}
		if !result.Passed {
			t.Errorf("expected pass for valid input, got failure: %s", result.Message)
		}
	}
}
