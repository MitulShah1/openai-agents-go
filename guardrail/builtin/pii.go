// Package builtin provides ready-to-use guardrail implementations.
package builtin

import (
	"context"
	"regexp"
	"strings"

	"github.com/MitulShah1/openai-agents-go/guardrail"
)

// PIIPattern represents a pattern for detecting personally identifiable information.
type PIIPattern struct {
	Name    string
	Pattern *regexp.Regexp
}

// Common PII patterns
var (
	EmailPattern = PIIPattern{
		Name:    "email",
		Pattern: regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`),
	}
	PhonePattern = PIIPattern{
		Name:    "phone",
		Pattern: regexp.MustCompile(`\b(\+\d{1,2}\s?)?(\(?\d{3}\)?[\s.-]?)?\d{3}[\s.-]?\d{4}\b`),
	}
	SSNPattern = PIIPattern{
		Name:    "ssn",
		Pattern: regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`),
	}
	CreditCardPattern = PIIPattern{
		Name:    "credit_card",
		Pattern: regexp.MustCompile(`\b\d{4}[\s-]?\d{4}[\s-]?\d{4}[\s-]?\d{4}\b`),
	}
)

// PIIConfig configures the PII detection guardrail.
type PIIConfig struct {
	// Tripwire determines if detection should halt execution
	Tripwire bool

	// DetectEmail enables email detection
	DetectEmail bool

	// DetectPhone enables phone number detection
	DetectPhone bool

	// DetectSSN enables SSN detection
	DetectSSN bool

	// DetectCreditCard enables credit card detection
	DetectCreditCard bool

	// CustomPatterns allows adding custom PII patterns
	CustomPatterns []PIIPattern
}

// PIIOption is a functional option for configuring PII guardrail.
type PIIOption func(*PIIConfig)

// WithTripwire enables tripwire mode (halts execution on detection).
func WithTripwire(enabled bool) PIIOption {
	return func(c *PIIConfig) {
		c.Tripwire = enabled
	}
}

// WithEmailDetection enables/disables email detection.
func WithEmailDetection(enabled bool) PIIOption {
	return func(c *PIIConfig) {
		c.DetectEmail = enabled
	}
}

// WithPhoneDetection enables/disables phone detection.
func WithPhoneDetection(enabled bool) PIIOption {
	return func(c *PIIConfig) {
		c.DetectPhone = enabled
	}
}

// WithSSNDetection enables/disables SSN detection.
func WithSSNDetection(enabled bool) PIIOption {
	return func(c *PIIConfig) {
		c.DetectSSN = enabled
	}
}

// WithCreditCardDetection enables/disables credit card detection.
func WithCreditCardDetection(enabled bool) PIIOption {
	return func(c *PIIConfig) {
		c.DetectCreditCard = enabled
	}
}

// WithCustomPattern adds a custom PII pattern.
func WithCustomPattern(name string, pattern *regexp.Regexp) PIIOption {
	return func(c *PIIConfig) {
		c.CustomPatterns = append(c.CustomPatterns, PIIPattern{
			Name:    name,
			Pattern: pattern,
		})
	}
}

// NewPIIGuardrail creates a guardrail that detects personally identifiable information.
func NewPIIGuardrail(opts ...PIIOption) *guardrail.Guardrail {
	config := &PIIConfig{
		Tripwire:         true, // Default to blocking
		DetectEmail:      true,
		DetectPhone:      true,
		DetectSSN:        true,
		DetectCreditCard: true,
	}

	for _, opt := range opts {
		opt(config)
	}

	return guardrail.NewGuardrail("pii_detection", func(_ context.Context, input string) (*guardrail.Result, error) {
		var detected []string
		var patterns []PIIPattern

		// Build pattern list based on config
		if config.DetectEmail {
			patterns = append(patterns, EmailPattern)
		}
		if config.DetectPhone {
			patterns = append(patterns, PhonePattern)
		}
		if config.DetectSSN {
			patterns = append(patterns, SSNPattern)
		}
		if config.DetectCreditCard {
			patterns = append(patterns, CreditCardPattern)
		}
		patterns = append(patterns, config.CustomPatterns...)

		// Check for PII
		for _, pattern := range patterns {
			if pattern.Pattern.MatchString(input) {
				detected = append(detected, pattern.Name)
			}
		}

		if len(detected) > 0 {
			return &guardrail.Result{
				Passed:            false,
				TripwireTriggered: config.Tripwire,
				Message:           "Detected PII: " + strings.Join(detected, ", "),
				Metadata: map[string]any{
					"detected_types": detected,
				},
			}, nil
		}

		return &guardrail.Result{
			Passed:            true,
			TripwireTriggered: false,
			Message:           "No PII detected",
		}, nil
	})
}
