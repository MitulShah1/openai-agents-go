package builtin

import (
	"context"
	"fmt"
	"strings"

	"github.com/MitulShah1/openai-agents-go/guardrail"
)

// CountMode determines how content length is measured.
type CountMode string

const (
	// CountModeCharacters counts total characters
	CountModeCharacters CountMode = "characters"
	// CountModeWords counts words (whitespace-separated tokens)
	CountModeWords CountMode = "words"
	// CountModeLines counts lines (newline-separated)
	CountModeLines CountMode = "lines"
)

// ContentLengthGuardrail validates content length constraints.
type ContentLengthGuardrail struct {
	mode     CountMode
	min      int
	max      int
	tripwire bool
}

// ContentLengthConfig configures the content length guardrail.
type ContentLengthConfig struct {
	// Mode determines how to count content (characters, words, or lines)
	Mode CountMode
	// Min is the minimum allowed count (0 = no minimum)
	Min int
	// Max is the maximum allowed count (0 = no maximum)
	Max int
	// Tripwire halts execution if content length is invalid
	Tripwire bool
}

// NewContentLengthGuardrail creates a new content length guardrail.
func NewContentLengthGuardrail(config ContentLengthConfig) *ContentLengthGuardrail {
	mode := config.Mode
	if mode == "" {
		mode = CountModeCharacters // Default to character counting
	}

	return &ContentLengthGuardrail{
		mode:     mode,
		min:      config.Min,
		max:      config.Max,
		tripwire: config.Tripwire,
	}
}

// Name returns the guardrail name.
func (g *ContentLengthGuardrail) Name() string {
	return "content_length"
}

// Validate checks if the content meets length constraints.
func (g *ContentLengthGuardrail) Validate(_ context.Context, input string) error {
	count := g.count(input)

	// Check minimum
	if g.min > 0 && count < g.min {
		msg := fmt.Sprintf("content too short: %d %s (minimum: %d)", count, g.mode, g.min)
		if g.tripwire {
			return &guardrail.InputGuardrailTripwireError{
				GuardrailName: g.Name(),
				Message:       msg,
			}
		}
		return fmt.Errorf("%s", msg)
	}

	// Check maximum
	if g.max > 0 && count > g.max {
		msg := fmt.Sprintf("content too long: %d %s (maximum: %d)", count, g.mode, g.max)
		if g.tripwire {
			return &guardrail.InputGuardrailTripwireError{
				GuardrailName: g.Name(),
				Message:       msg,
			}
		}
		return fmt.Errorf("%s", msg)
	}

	return nil
}

// IsTripwire returns whether this guardrail halts execution on failure.
func (g *ContentLengthGuardrail) IsTripwire() bool {
	return g.tripwire
}

// count returns the count based on the configured mode.
func (g *ContentLengthGuardrail) count(input string) int {
	switch g.mode {
	case CountModeWords:
		return g.countWords(input)
	case CountModeLines:
		return g.countLines(input)
	case CountModeCharacters:
		fallthrough
	default:
		return len(input)
	}
}

// countWords counts whitespace-separated words.
func (g *ContentLengthGuardrail) countWords(input string) int {
	// Trim whitespace and split on whitespace
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return 0
	}

	words := strings.Fields(trimmed)
	return len(words)
}

// countLines counts newline-separated lines.
func (g *ContentLengthGuardrail) countLines(input string) int {
	if input == "" {
		return 0
	}

	lines := strings.Split(input, "\n")
	return len(lines)
}
