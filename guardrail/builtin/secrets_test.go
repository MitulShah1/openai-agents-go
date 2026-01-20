package builtin

import (
	"context"
	"regexp"
	"testing"
)

func TestSecretsGuardrail_APIKeys(t *testing.T) {
	guard := NewSecretsGuardrail(SecretsConfig{})
	ctx := context.Background()

	t.Run("detects generic API key", func(t *testing.T) {
		t.Skip("Skipped: GitHub secret scanning blocks realistic test data")
	})

	t.Run("detects Google API key", func(t *testing.T) {
		t.Skip("Skipped: GitHub secret scanning blocks realistic test data")
	})

	t.Run("clean text passes", func(t *testing.T) {
		result, err := guard.Func(ctx, "This is clean text without secrets")
		if err != nil {
			t.Fatalf("unexpected execution error: %v", err)
		}
		if !result.Passed {
			t.Errorf("expected clean text to pass, got: %s", result.Message)
		}
	})
}

func TestSecretsGuardrail_AWS(t *testing.T) {
	guard := NewSecretsGuardrail(SecretsConfig{})
	ctx := context.Background()

	t.Run("detects AWS access key", func(t *testing.T) {
		result, err := guard.Func(ctx, "AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE")
		if err != nil {
			t.Fatalf("unexpected execution error: %v", err)
		}
		if result.Passed {
			t.Error("expected AWS access key to fail validation")
		}
	})

	t.Run("detects AWS secret key", func(t *testing.T) {
		t.Skip("Skipped: GitHub secret scanning blocks realistic test data")
	})
}

func TestSecretsGuardrail_GitHub(t *testing.T) {
	t.Run("detects GitHub personal access token", func(t *testing.T) {
		t.Skip("Skipped: GitHub secret scanning blocks realistic test data")
	})

	t.Run("detects GitHub OAuth token", func(t *testing.T) {
		t.Skip("Skipped: GitHub secret scanning blocks realistic test data")
	})
}

func TestSecretsGuardrail_Tokens(t *testing.T) {
	guard := NewSecretsGuardrail(SecretsConfig{})
	ctx := context.Background()

	t.Run("detects Bearer token", func(t *testing.T) {
		result, err := guard.Func(ctx, "Authorization: Bearer abcdef1234567890abcdef1234567890")
		if err != nil {
			t.Fatalf("unexpected execution error: %v", err)
		}
		if result.Passed {
			t.Error("expected Bearer token to fail validation")
		}
	})

	t.Run("detects JWT token", func(t *testing.T) {
		jwt := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"
		result, err := guard.Func(ctx, jwt)
		if err != nil {
			t.Fatalf("unexpected execution error: %v", err)
		}
		if result.Passed {
			t.Error("expected JWT token to fail validation")
		}
	})

	t.Run("detects Slack token", func(t *testing.T) {
		t.Skip("Skipped: GitHub secret scanning blocks realistic test data")
	})
}

func TestSecretsGuardrail_Passwords(t *testing.T) {
	guard := NewSecretsGuardrail(SecretsConfig{})
	ctx := context.Background()

	t.Run("detects password in URL", func(t *testing.T) {
		result, err := guard.Func(ctx, "postgres://user:my_secret_password@localhost:5432/db")
		if err != nil {
			t.Fatalf("unexpected execution error: %v", err)
		}
		if result.Passed {
			t.Error("expected password in URL to fail validation")
		}
	})

	t.Run("detects password assignment", func(t *testing.T) {
		result, err := guard.Func(ctx, "password=MySuperSecret123!")
		if err != nil {
			t.Fatalf("unexpected execution error: %v", err)
		}
		if result.Passed {
			t.Error("expected password assignment to fail validation")
		}
	})

	t.Run("short password not detected", func(t *testing.T) {
		// Passwords under 8 characters are not flagged
		result, err := guard.Func(ctx, "pwd=short")
		if err != nil {
			t.Fatalf("unexpected execution error: %v", err)
		}
		if !result.Passed {
			t.Errorf("short password should pass validation, got: %v", result.Message)
		}
	})
}

func TestSecretsGuardrail_PrivateKeys(t *testing.T) {
	guard := NewSecretsGuardrail(SecretsConfig{})
	ctx := context.Background()

	t.Run("detects private key", func(t *testing.T) {
		result, err := guard.Func(ctx, "-----BEGIN PRIVATE KEY-----\\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASC...")
		if err != nil {
			t.Fatalf("unexpected execution error: %v", err)
		}
		if result.Passed {
			t.Error("expected private key to fail validation")
		}
	})

	t.Run("detects RSA private key", func(t *testing.T) {
		result, err := guard.Func(ctx, "-----BEGIN RSA PRIVATE KEY-----")
		if err != nil {
			t.Fatalf("unexpected execution error: %v", err)
		}
		if result.Passed {
			t.Error("expected RSA private key to fail validation")
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
		result, err := guard.Func(ctx, "My secret: CUSTOM_SECRET_ABCD1234EFGH5678")
		if err != nil {
			t.Fatalf("unexpected execution error: %v", err)
		}
		if result.Passed {
			t.Error("expected custom secret to fail validation")
		}
	})

	t.Run("still detects default patterns", func(t *testing.T) {
		result, err := guard.Func(ctx, "AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE")
		if err != nil {
			t.Fatalf("unexpected execution error: %v", err)
		}
		if result.Passed {
			t.Error("expected default pattern to fail validation")
		}
	})
}

func TestSecretsGuardrail_Tripwire(t *testing.T) {
	guard := NewSecretsGuardrail(SecretsConfig{
		Tripwire: true,
	})

	ctx := context.Background()
	result, err := guard.Func(ctx, "api_key: sk_test_1234567890abcdef1234")
	if err != nil {
		t.Fatalf("unexpected execution error: %v", err)
	}
	if result.Passed {
		t.Error("expected tripwire to fail validation")
	}
	if !result.TripwireTriggered {
		t.Error("expected TripwireTriggered to be true")
	}
}

func TestSecretsGuardrail_Name(t *testing.T) {
	guard := NewSecretsGuardrail(SecretsConfig{})

	if guard.Name != "secrets" {
		t.Errorf("expected name 'secrets', got '%s'", guard.Name)
	}
}

func TestSecretsGuardrail_MultipleSecrets(t *testing.T) {
	guard := NewSecretsGuardrail(SecretsConfig{})
	ctx := context.Background()

	//nolint:gosec // Test data with example credentials
	input := `
		AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE
		password=MySecretPassword123!
	`

	result, err := guard.Func(ctx, input)
	if err != nil {
		t.Fatalf("unexpected execution error: %v", err)
	}
	if result.Passed {
		t.Error("expected multiple secrets to fail validation")
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
			result, err := guard.Func(ctx, tc.input)
			if err != nil {
				t.Fatalf("unexpected execution error: %v", err)
			}
			if result.Passed {
				t.Errorf("expected %s to fail validation", tc.name)
			}
		})
	}
}
