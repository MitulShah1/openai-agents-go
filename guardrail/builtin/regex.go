package builtin

import (
	"context"
	"regexp"

	"github.com/MitulShah1/openai-agents-go/guardrail"
)

// RegexConfig configures the regex guardrail.
type RegexConfig struct {
	// Pattern is the regex pattern to match
	Pattern *regexp.Regexp

	// MustMatch determines if the input must match (true) or must NOT match (false)
	MustMatch bool

	// Tripwire determines if violation should halt execution
	Tripwire bool

	// Message is the custom message on failure
	Message string
}

// RegexOption is a functional option for regex guardrail.
type RegexOption func(*RegexConfig)

// WithRegexTripwire enables tripwire mode.
func WithRegexTripwire(enabled bool) RegexOption {
	return func(c *RegexConfig) {
		c.Tripwire = enabled
	}
}

// WithMustMatch sets whether pattern must match or must NOT match.
func WithMustMatch(must bool) RegexOption {
	return func(c *RegexConfig) {
		c.MustMatch = must
	}
}

// WithRegexMessage sets a custom failure message.
func WithRegexMessage(msg string) RegexOption {
	return func(c *RegexConfig) {
		c.Message = msg
	}
}

// NewRegexGuardrail creates a guardrail that validates input against a regex pattern.
func NewRegexGuardrail(pattern string, opts ...RegexOption) *guardrail.Guardrail {
	compiled := regexp.MustCompile(pattern)

	config := &RegexConfig{
		Pattern:   compiled,
		MustMatch: false, // Default: pattern must NOT match
		Tripwire:  true,
		Message:   "Pattern validation failed",
	}

	for _, opt := range opts {
		opt(config)
	}

	return guardrail.NewGuardrail("regex_validation", func(_ context.Context, input string) (*guardrail.Result, error) {
		matches := config.Pattern.MatchString(input)

		// Check if validation passes
		passed := (config.MustMatch && matches) || (!config.MustMatch && !matches)

		if !passed {
			msg := config.Message
			if msg == "Pattern validation failed" {
				if config.MustMatch {
					msg = "Input does not match required pattern"
				} else {
					msg = "Input matches forbidden pattern"
				}
			}

			return &guardrail.Result{
				Passed:            false,
				TripwireTriggered: config.Tripwire,
				Message:           msg,
				Metadata: map[string]any{
					"pattern":     config.Pattern.String(),
					"must_match":  config.MustMatch,
					"input_match": matches,
				},
			}, nil
		}

		return &guardrail.Result{
			Passed:            true,
			TripwireTriggered: false,
			Message:           "Pattern validation passed",
		}, nil
	})
}
