// Package content provides content validation guardrails (length, regex).
package content

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

// Config configures the content length guardrail.
type Config struct {
	// Mode determines how to count content (characters, words, or lines)
	Mode CountMode
	// Min is the minimum allowed count (0 = no minimum)
	Min int
	// Max is the maximum allowed count (0 = no maximum)
	Max int
	// Tripwire halts execution if content length is invalid
	Tripwire bool
}

// NewLength creates a new content length guardrail.
func NewLength(config Config) *guardrail.Guardrail {
	mode := config.Mode
	if mode == "" {
		mode = CountModeCharacters // Default to character counting
	}
	minVal := config.Min
	maxVal := config.Max
	tripwire := config.Tripwire

	return guardrail.NewGuardrail("content_length", func(_ context.Context, input string) (*guardrail.Result, error) {
		count := countContent(input, mode)

		// Check minimum
		if minVal > 0 && count < minVal {
			return &guardrail.Result{
				Passed:            false,
				TripwireTriggered: tripwire,
				Message:           fmt.Sprintf("content too short: %d %s (minimum: %d)", count, mode, minVal),
				Metadata: map[string]any{
					"count": count,
					"mode":  mode,
					"min":   minVal,
				},
			}, nil
		}

		// Check maximum
		if maxVal > 0 && count > maxVal {
			return &guardrail.Result{
				Passed:            false,
				TripwireTriggered: tripwire,
				Message:           fmt.Sprintf("content too long: %d %s (maximum: %d)", count, mode, maxVal),
				Metadata: map[string]any{
					"count": count,
					"mode":  mode,
					"max":   maxVal,
				},
			}, nil
		}

		return &guardrail.Result{
			Passed:            true,
			TripwireTriggered: false,
			Message:           "content length within limits",
			Metadata: map[string]any{
				"count": count,
			},
		}, nil
	})
}

// countContent returns the count based on the configured mode.
func countContent(input string, mode CountMode) int {
	switch mode {
	case CountModeWords:
		// Trim whitespace and split on whitespace
		trimmed := strings.TrimSpace(input)
		if trimmed == "" {
			return 0
		}
		words := strings.Fields(trimmed)
		return len(words)
	case CountModeLines:
		if input == "" {
			return 0
		}
		lines := strings.Split(input, "\n")
		return len(lines)
	case CountModeCharacters:
		fallthrough
	default:
		return len(input)
	}
}
