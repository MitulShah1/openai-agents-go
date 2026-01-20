package moderation

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/MitulShah1/openai-agents-go/guardrail"
)

// SeverityLevel represents the severity of detected profanity.
type SeverityLevel int

const (
	// SeverityLow represents mild profanity
	SeverityLow SeverityLevel = 1
	// SeverityMedium represents moderate profanity
	SeverityMedium SeverityLevel = 2
	// SeverityHigh represents severe profanity
	SeverityHigh SeverityLevel = 3
)

// ProfanityGuardrail detects profanity and toxic language.
type ProfanityGuardrail struct {
	wordList           map[string]SeverityLevel
	minSeverity        SeverityLevel
	normalizeLeetspeak bool
	tripwire           bool
	leetSpeakMap       map[rune]rune
}

// ProfanityConfig configures the profanity guardrail.
type ProfanityConfig struct {
	// WordList is a map of profane words to their severity levels
	WordList map[string]SeverityLevel
	// MinSeverity is the minimum severity level to block (default: SeverityLow)
	MinSeverity SeverityLevel
	// NormalizeLeetspeak converts leetspeak to normal characters (default: true)
	NormalizeLeetspeak bool
	// Tripwire halts execution if profanity is detected
	Tripwire bool
}

// NewProfanity creates a new profanity detection guardrail.
func NewProfanity(config ProfanityConfig) *guardrail.Guardrail {
	wordList := config.WordList
	if wordList == nil {
		wordList = getDefaultWordList()
	}

	minSeverity := config.MinSeverity
	if minSeverity == 0 {
		minSeverity = SeverityLow // Block all by default
	}

	normalizeLeetspeak := config.NormalizeLeetspeak
	// If WordList is provided but NormalizeLeetspeak wasn't explicitly set, default to false
	// If using default word list, default to true
	if config.WordList == nil {
		normalizeLeetspeak = true // Default to true for default word list
	}

	tripwire := config.Tripwire
	leetSpeakMap := getLeetSpeakMap()

	return guardrail.NewGuardrail("profanity", func(_ context.Context, input string) (*guardrail.Result, error) {
		// Normalize input
		normalized := strings.ToLower(input)

		if normalizeLeetspeak {
			var result strings.Builder
			result.Grow(len(normalized))
			for _, r := range normalized {
				if replacement, found := leetSpeakMap[r]; found {
					result.WriteRune(replacement)
				} else {
					result.WriteRune(r)
				}
			}
			normalized = result.String()
		}

		// Check for profanity
		words := regexp.MustCompile(`\w+`).FindAllString(normalized, -1)

		var detectedWords []string
		highestSeverity := SeverityLevel(0)

		for _, word := range words {
			if severity, found := wordList[word]; found {
				if severity >= minSeverity {
					detectedWords = append(detectedWords, word)
					if severity > highestSeverity {
						highestSeverity = severity
					}
				}
			}
		}

		if len(detectedWords) > 0 {
			return &guardrail.Result{
				Passed:            false,
				TripwireTriggered: tripwire,
				Message:           fmt.Sprintf("profanity detected (severity: %d): %d word(s) found", highestSeverity, len(detectedWords)),
				Metadata: map[string]any{
					"severity":       highestSeverity,
					"detected_count": len(detectedWords),
				},
			}, nil
		}

		return &guardrail.Result{
			Passed:            true,
			TripwireTriggered: false,
			Message:           "no profanity detected",
		}, nil
	})
}

// Name returns the guardrail name.
func (g *ProfanityGuardrail) Name() string {
	return "profanity"
}

// Validate checks if the input contains profanity.
func (g *ProfanityGuardrail) Validate(_ context.Context, input string) error {
	// Normalize input
	normalized := strings.ToLower(input)

	if g.normalizeLeetspeak {
		normalized = g.normalizeLeet(normalized)
	}

	// Check for profanity
	words := regexp.MustCompile(`\w+`).FindAllString(normalized, -1)

	var detectedWords []string
	highestSeverity := SeverityLevel(0)

	for _, word := range words {
		if severity, found := g.wordList[word]; found {
			if severity >= g.minSeverity {
				detectedWords = append(detectedWords, word)
				if severity > highestSeverity {
					highestSeverity = severity
				}
			}
		}
	}

	if len(detectedWords) > 0 {
		msg := fmt.Sprintf("profanity detected (severity: %d): %d word(s) found",
			highestSeverity, len(detectedWords))

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
func (g *ProfanityGuardrail) IsTripwire() bool {
	return g.tripwire
}

// normalizeLeet converts leetspeak characters to normal characters.
func (g *ProfanityGuardrail) normalizeLeet(input string) string {
	var result strings.Builder
	result.Grow(len(input))

	for _, r := range input {
		if replacement, found := g.leetSpeakMap[r]; found {
			result.WriteRune(replacement)
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// getLeetSpeakMap returns common leetspeak character substitutions.
func getLeetSpeakMap() map[rune]rune {
	return map[rune]rune{
		'0': 'o',
		'1': 'i',
		'3': 'e',
		'4': 'a',
		'5': 's',
		'7': 't',
		'8': 'b',
		'@': 'a',
		'$': 's',
		'!': 'i',
	}
}

// getDefaultWordList returns a basic English profanity word list.
// This is a minimal list for demonstration - production use should have comprehensive lists.
func getDefaultWordList() map[string]SeverityLevel {
	return map[string]SeverityLevel{
		// High severity
		"fuck":    SeverityHigh,
		"fucking": SeverityHigh,
		"fucked":  SeverityHigh,
		"shit":    SeverityHigh,
		"shitty":  SeverityHigh,
		"bitch":   SeverityHigh,
		"asshole": SeverityHigh,
		"bastard": SeverityHigh,
		"damn":    SeverityMedium,

		// Medium severity
		"crap":   SeverityMedium,
		"crappy": SeverityMedium,
		"hell":   SeverityMedium,
		"piss":   SeverityMedium,
		"pissed": SeverityMedium,
		"ass":    SeverityMedium,

		// Low severity (mild)
		"stupid": SeverityLow,
		"dumb":   SeverityLow,
		"idiot":  SeverityLow,
	}
}
