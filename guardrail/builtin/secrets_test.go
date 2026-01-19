package builtin

import (
	"context"
	"regexp"
	"strings"
	"testing"

	"github.com/MitulShah1/openai-agents-go/guardrail"
)

func TestSecretsGuardrail_APIKeys(t *testing.T) {
	guard := NewSecretsGuardrail(SecretsConfig{})
	ctx := context.Background()

	t.Run("detects generic API key", func(t *testing.T) {
		// Using obviously fake key for testing
		err := guard.Validate(ctx, "My api_key is sk_test_EXAMPLE1234567890abcdefghijklmnop")
		if err == nil {
			t.Error("expected error for API key")
		}
	})

	t.Run("detects Google API key", func(t *testing.T) {
		err := guard.Validate(ctx, "Google key: AIzaSyDaGmWKa4JsXZ-HjGw7ISLn_3namBGewQe")
		if err == nil {
			t.Error("expected error for Google API key")
		}
	})

	t.Run("clean text passes", func(t *testing.T) {
		err := guard.Validate(ctx, "This is clean text without secrets")
		if err != nil {
			t.Errorf("unexpected error for clean text: %v", err)
		}
	})
}

func TestSecretsGuardrail_AWS(t *testing.T) {
	guard := NewSecretsGuardrail(SecretsConfig{})
	ctx := context.Background()

	t.Run("detects AWS access key", func(t *testing.T) {
		err := guard.Validate(ctx, "AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE")
		if err == nil {
			t.Error("expected error for AWS access key")
		}
	})

	t.Run("detects AWS secret key", func(t *testing.T) {
		err := guard.Validate(ctx, "aws_secret_key: wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY")
		if err == nil {
			t.Error("expected error for AWS secret key")
		}
	})
}

func TestSecretsGuardrail_GitHub(t *testing.T) {
	guard := NewSecretsGuardrail(SecretsConfig{})
	ctx := context.Background()

	t.Run("detects GitHub personal access token", func(t *testing.T) {
		err := guard.Validate(ctx, "token: ghp_1234567890abcdefghijklmnopqrstuv123")
		if err == nil {
			t.Error("expected error for GitHub token")
		}
	})

	t.Run("detects GitHub OAuth token", func(t *testing.T) {
		err := guard.Validate(ctx, "oauth: gho_1234567890abcdefghijklmnopqrstuv123")
		if err == nil {
			t.Error("expected error for GitHub OAuth")
		}
	})
}

func TestSecretsGuardrail_Tokens(t *testing.T) {
	guard := NewSecretsGuardrail(SecretsConfig{})
	ctx := context.Background()

	t.Run("detects Bearer token", func(t *testing.T) {
		err := guard.Validate(ctx, "Authorization: Bearer abcdef1234567890abcdef1234567890")
		if err == nil {
			t.Error("expected error for Bearer token")
		}
	})

	t.Run("detects JWT token", func(t *testing.T) {
		jwt := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"
		err := guard.Validate(ctx, jwt)
		if err == nil {
			t.Error("expected error for JWT token")
		}
	})

	t.Run("detects Slack token", func(t *testing.T) {
		// Using TEST_ prefix to avoid GitHub secret scanning
		err := guard.Validate(ctx, "xoxb-TEST-FAKE-SLACKTOKEN")
		if err == nil {
			t.Error("expected error for Slack token")
		}
	})
}

func TestSecretsGuardrail_Passwords(t *testing.T) {
	guard := NewSecretsGuardrail(SecretsConfig{})
	ctx := context.Background()

	t.Run("detects password in URL", func(t *testing.T) {
		err := guard.Validate(ctx, "postgres://user:my_secret_password@localhost:5432/db")
		if err == nil {
			t.Error("expected error for password in URL")
		}
	})

	t.Run("detects password assignment", func(t *testing.T) {
		err := guard.Validate(ctx, "password=MySuperSecret123!")
		if err == nil {
			t.Error("expected error for password assignment")
		}
	})

	t.Run("short password not detected", func(t *testing.T) {
		// Passwords under 8 characters are not flagged
		err := guard.Validate(ctx, "pwd=short")
		if err != nil {
			t.Errorf("unexpected error for short password: %v", err)
		}
	})
}

func TestSecretsGuardrail_PrivateKeys(t *testing.T) {
	guard := NewSecretsGuardrail(SecretsConfig{})
	ctx := context.Background()

	t.Run("detects private key", func(t *testing.T) {
		err := guard.Validate(ctx, "-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASC...")
		if err == nil {
			t.Error("expected error for private key")
		}
	})

	t.Run("detects RSA private key", func(t *testing.T) {
		err := guard.Validate(ctx, "-----BEGIN RSA PRIVATE KEY-----")
		if err == nil {
			t.Error("expected error for RSA private key")
		}
	})
}

func TestSecretsGuardrail_CustomPatterns(t *testing.T) {
	customPattern := SecretPattern{
		Type:    SecretTypeCustom,
		Name:    "Custom Secret",
		Pattern: regexp.MustCompile(`CUSTOM_SECRET_[A-Z0-9]{16}`),
	}

	guard := NewSecretsGuardrail(SecretsConfig{
		CustomPatterns: []SecretPattern{customPattern},
	})

	ctx := context.Background()

	t.Run("detects custom secret", func(t *testing.T) {
		err := guard.Validate(ctx, "My secret: CUSTOM_SECRET_ABCD1234EFGH5678")
		if err == nil {
			t.Error("expected error for custom secret")
		}
	})

	t.Run("still detects default patterns", func(t *testing.T) {
		err := guard.Validate(ctx, "AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE")
		if err == nil {
			t.Error("expected error for default pattern")
		}
	})
}

func TestSecretsGuardrail_Tripwire(t *testing.T) {
	guard := NewSecretsGuardrail(SecretsConfig{
		Tripwire: true,
	})

	if !guard.IsTripwire() {
		t.Error("guardrail should be marked as tripwire")
	}

	ctx := context.Background()
	err := guard.Validate(ctx, "api_key: sk_test_1234567890abcdef1234")
	if err == nil {
		t.Error("expected tripwire error")
	}

	// Check it's the right error type
	if _, ok := err.(*guardrail.InputGuardrailTripwireError); !ok {
		t.Errorf("expected InputGuardrailTripwireError, got %T", err)
	}
}

func TestSecretsGuardrail_Name(t *testing.T) {
	guard := NewSecretsGuardrail(SecretsConfig{})

	if guard.Name() != "secrets" {
		t.Errorf("expected name 'secrets', got '%s'", guard.Name())
	}
}

func TestSecretsGuardrail_MultipleSecrets(t *testing.T) {
	guard := NewSecretsGuardrail(SecretsConfig{})
	ctx := context.Background()

	//nolint:gosec // Test data with example credentials
	input := `
		AWS_ACCESS_KEY_ID=TEST_AKIAIOSFODNN7EXAMPLE
		AWS_SECRET_ACCESS_KEY=TEST_wJalrXUtnFEMI_K7MDENG_EXAMPLE
		api_key: TEST_sk_FAKEAPIKEY1234567890
	`

	err := guard.Validate(ctx, input)
	if err == nil {
		t.Error("expected error for multiple secrets")
	}

	// Should mention multiple detections
	if !strings.Contains(err.Error(), "secrets detected") {
		t.Errorf("error should mention secrets detected: %v", err)
	}
}

func TestSecretsGuardrail_CaseSensitivity(t *testing.T) {
	guard := NewSecretsGuardrail(SecretsConfig{})
	ctx := context.Background()

	testCases := []struct {
		name  string
		input string
	}{
		{"lowercase", "api_key: sk_test_1234567890abcdef1234"},
		{"uppercase", "API_KEY: sk_test_1234567890abcdef1234"},
		{"mixed", "Api_Key: sk_test_1234567890abcdef1234"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := guard.Validate(ctx, tc.input)
			if err == nil {
				t.Errorf("expected error for %s", tc.name)
			}
		})
	}
}
