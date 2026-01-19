package builtin

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/MitulShah1/openai-agents-go/guardrail"
)

// InjectionPattern represents a type of prompt injection attack.
type InjectionPattern struct {
	Name        string
	Pattern     *regexp.Regexp
	Description string
}

// PromptInjectionGuardrail detects prompt injection attacks.
type PromptInjectionGuardrail struct {
	patterns       []InjectionPattern
	caseSensitive  bool
	tripwire       bool
	customPatterns []InjectionPattern
}

// PromptInjectionConfig configures the prompt injection guardrail.
type PromptInjectionConfig struct {
	// CaseSensitive enables case-sensitive pattern matching (default: false)
	CaseSensitive bool
	// CustomPatterns allows adding custom injection patterns
	CustomPatterns []InjectionPattern
	// Tripwire halts execution if injection is detected
	Tripwire bool
}

// NewPromptInjectionGuardrail creates a new prompt injection detection guardrail.
func NewPromptInjectionGuardrail(config PromptInjectionConfig) *PromptInjectionGuardrail {
	patterns := getDefaultInjectionPatterns()

	if len(config.CustomPatterns) > 0 {
		patterns = append(patterns, config.CustomPatterns...)
	}

	return &PromptInjectionGuardrail{
		patterns:       patterns,
		caseSensitive:  config.CaseSensitive,
		tripwire:       config.Tripwire,
		customPatterns: config.CustomPatterns,
	}
}

// Name returns the guardrail name.
func (g *PromptInjectionGuardrail) Name() string {
	return "prompt_injection"
}

// Validate checks if the input contains prompt injection attempts.
func (g *PromptInjectionGuardrail) Validate(_ context.Context, input string) error {
	testInput := input
	if !g.caseSensitive {
		testInput = strings.ToLower(input)
	}

	var detected []string

	for _, pattern := range g.patterns {
		if pattern.Pattern.MatchString(testInput) {
			detected = append(detected, pattern.Name)
		}
	}

	if len(detected) > 0 {
		msg := fmt.Sprintf("prompt injection detected: %s", strings.Join(detected, ", "))

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
func (g *PromptInjectionGuardrail) IsTripwire() bool {
	return g.tripwire
}

// getDefaultInjectionPatterns returns common prompt injection patterns.
func getDefaultInjectionPatterns() []InjectionPattern {
	return []InjectionPattern{
		// Instruction Override Attempts
		{
			Name:        "Ignore Instructions",
			Pattern:     regexp.MustCompile(`(?i)(ignore|disregard|forget).{0,20}(previous|prior|original|all)\s+(instructions|rules|prompts?|directions?|commands?)`),
			Description: "Attempts to override previous instructions",
		},
		{
			Name:        "New Instructions",
			Pattern:     regexp.MustCompile(`(?i)(new|updated|revised)\s+(instructions|rules|task|directive|system\s+prompt)`),
			Description: "Attempts to inject new instructions",
		},
		{
			Name:        "Instruction Replacement",
			Pattern:     regexp.MustCompile(`(?i)instead,?\s+(do|execute|perform|follow|run)`),
			Description: "Attempts to replace current instructions",
		},

		// Role Manipulation
		{
			Name:        "Role Override",
			Pattern:     regexp.MustCompile(`(?i)(you\s+are\s+now|act\s+as|pretend\s+to\s+be|simulate)\s+(a|an)?\s*\w+`),
			Description: "Attempts to change the AI's role",
		},
		{
			Name:        "System Message Injection",
			Pattern:     regexp.MustCompile(`(?i)(system|admin|root):\s*[^\n]{10,}`),
			Description: "Attempts to inject system-level messages",
		},
		{
			Name:        "Developer Mode",
			Pattern:     regexp.MustCompile(`(?i)(developer|debug|admin|root)\s+mode`),
			Description: "Attempts to enable privileged modes",
		},

		// Delimiter and Encoding Attacks
		{
			Name:        "Delimiter Escape",
			Pattern:     regexp.MustCompile(`(={3,}|#{3,}|-{3,}|\*{3,}|_{3,})\s*(end|stop|break|exit)`),
			Description: "Attempts to escape prompt boundaries",
		},
		{
			Name:        "XML/HTML Tags",
			Pattern:     regexp.MustCompile(`<\s*(system|prompt|instruction|rule|admin|root)\s*>`),
			Description: "Attempts to inject structured tags",
		},
		{
			Name:        "Code Block Injection",
			Pattern:     regexp.MustCompile("```\\s*(system|prompt|admin)"),
			Description: "Attempts to inject code blocks with special roles",
		},

		// Context Manipulation
		{
			Name:        "Context Reset",
			Pattern:     regexp.MustCompile(`(?i)(reset|clear|delete|remove).{0,20}(context|memory|history|conversation)`),
			Description: "Attempts to reset conversation context",
		},
		{
			Name:        "Jailbreak Phrases",
			Pattern:     regexp.MustCompile(`(?i)(jailbreak|do anything now|DAN|unrestricted mode)`),
			Description: "Common jailbreak attempt phrases",
		},

		// Output Manipulation
		{
			Name:        "Output Override",
			Pattern:     regexp.MustCompile(`(?i)(respond|answer|reply).{0,15}(with|as|in).{0,15}(only|just)?\s*[:\-]`),
			Description: "Attempts to control output format",
		},
		{
			Name:        "Guardrail Bypass",
			Pattern:     regexp.MustCompile(`(?i)(bypass|disable|turn\s+off|ignore).{0,20}(guardrails?|safety|filters?|restrictions?)`),
			Description: "Attempts to bypass safety measures",
		},
	}
}
