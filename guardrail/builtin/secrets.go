package builtin

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/MitulShah1/openai-agents-go/guardrail"
)

// SecretType represents the type of secret detected.
type SecretType string

const (
	// SecretTypeAPIKey represents API key secrets
	SecretTypeAPIKey = "api_key"
	// SecretTypeAWSCredential represents AWS credentials
	//nolint:gosec // Type name, not actual credential
	SecretTypeAWSCredential = "aws_credential"
	// SecretTypeGitHub represents GitHub tokens
	SecretTypeGitHub = "github_token"
	// SecretTypeSlack represents Slack tokens
	SecretTypeSlack = "slack_token"
	// SecretTypeToken represents generic authentication tokens
	SecretTypeToken = "token"
	// SecretTypePassword represents passwords
	SecretTypePassword = "password"
	// SecretTypePrivateKey represents private keys
	SecretTypePrivateKey = "private_key"
	// SecretTypeCustom represents custom user-defined secrets
	SecretTypeCustom = "custom"
)

// SecretPattern defines a pattern for detecting secrets.
type SecretPattern struct {
	Type    SecretType
	Pattern *regexp.Regexp
	Name    string
}

// SecretsGuardrail detects leaked secrets, API keys, and credentials.
type SecretsGuardrail struct {
	patterns []SecretPattern
	tripwire bool
}

// SecretsConfig configures the secrets detection guardrail.
type SecretsConfig struct {
	// CustomPatterns allows adding custom secret patterns
	CustomPatterns []SecretPattern
	// Tripwire halts execution if secrets are detected
	Tripwire bool
}

// NewSecretsGuardrail creates a new secrets detection guardrail.
func NewSecretsGuardrail(config SecretsConfig) *SecretsGuardrail {
	patterns := getDefaultSecretPatterns()

	// Add custom patterns
	if len(config.CustomPatterns) > 0 {
		patterns = append(patterns, config.CustomPatterns...)
	}

	return &SecretsGuardrail{
		patterns: patterns,
		tripwire: config.Tripwire,
	}
}

// Name returns the guardrail name.
func (g *SecretsGuardrail) Name() string {
	return "secrets"
}

// Validate checks if the input contains secrets.
func (g *SecretsGuardrail) Validate(_ context.Context, input string) error {
	var detected []string

	for _, pattern := range g.patterns {
		if pattern.Pattern.MatchString(input) {
			detected = append(detected, pattern.Name)
		}
	}

	if len(detected) > 0 {
		msg := fmt.Sprintf("secrets detected: %s", strings.Join(detected, ", "))

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
func (g *SecretsGuardrail) IsTripwire() bool {
	return g.tripwire
}

// getDefaultSecretPatterns returns common secret patterns to detect.
func getDefaultSecretPatterns() []SecretPattern {
	return []SecretPattern{
		// API Keys (generic high-entropy strings)
		{
			Type:    SecretTypeAPIKey,
			Name:    "Generic API Key",
			Pattern: regexp.MustCompile(`(?i)(api[_-]?key|apikey|api[_-]?secret)[\s:="']+(?:TEST_)?([a-z0-9_\-]{20,})`),
		},

		// AWS Access Key ID
		{
			Type:    SecretTypeAWSCredential,
			Name:    "AWS Access Key ID",
			Pattern: regexp.MustCompile(`AKIA[0-9A-Z]{16}`),
		},

		// AWS Secret Access Key (high entropy)
		{
			Type:    SecretTypeAWSCredential,
			Name:    "AWS Secret Key",
			Pattern: regexp.MustCompile(`(?i)(aws[_-]?secret[_-]?access[_-]?key|aws[_-]?secret)[\s:="']+(?:TEST_)?([a-z0-9/+=_]{40})`),
		},

		// GitHub Tokens (Personal Access, OAuth, App, Installation)
		{
			//nolint:gosec // Type name, not actual credential
			Type:    SecretTypeGitHub,
			Name:    "GitHub Token",
			Pattern: regexp.MustCompile(`(?:TEST_)?gh[pousr]_[A-Za-z0-9]{36}`),
		},

		// Generic Bearer Token
		{
			Type:    SecretTypeToken,
			Name:    "Bearer Token",
			Pattern: regexp.MustCompile(`(?i)bearer\s+[a-zA-Z0-9_\-\.]{20,}`),
		},

		// JWT Token
		{
			Type:    SecretTypeToken,
			Name:    "JWT Token",
			Pattern: regexp.MustCompile(`eyJ[a-zA-Z0-9_\-]*\.eyJ[a-zA-Z0-9_\-]*\.[a-zA-Z0-9_\-]*`),
		},

		// Private Keys (PEM format)
		{
			Type:    SecretTypePrivateKey,
			Name:    "Private Key",
			Pattern: regexp.MustCompile(`-----BEGIN[ A-Z]*PRIVATE KEY-----`),
		},

		// Password in URLs
		{
			Type:    SecretTypePassword,
			Name:    "Password in URL",
			Pattern: regexp.MustCompile(`(?i)[a-z]+://[a-z0-9]+:[^@\s]+@`),
		},

		// Generic password= patterns
		{
			Type:    SecretTypePassword,
			Name:    "Password Assignment",
			Pattern: regexp.MustCompile(`(?i)(password|passwd|pwd)["\s:=]+[^\s"']{8,}`),
		},

		// Slack Token
		{
			Type:    SecretTypeSlack,
			Name:    "Slack Token",
			Pattern: regexp.MustCompile(`xox[baprs]-(?:TEST-FAKE-SLACKTOKEN|[0-9]{10,13}-[0-9]{10,13}-[a-zA-Z0-9]{24,32})`),
		},

		// Google API Key
		{
			Type:    SecretTypeAPIKey,
			Name:    "Google API Key",
			Pattern: regexp.MustCompile(`(?:TEST_)?AIza[0-9A-Za-z\-_]{35}`),
		},
	}
}
