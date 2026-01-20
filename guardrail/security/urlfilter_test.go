package security

import (
	"context"
	"testing"
)

func TestNewURLFilter(t *testing.T) {
	tests := []struct {
		name             string
		opts             []URLFilterOption
		input            string
		expectedPass     bool
		expectedTripwire bool
	}{
		{
			name:         "no_urls_passes_by_default",
			opts:         nil,
			input:        "Just some text without links",
			expectedPass: true,
		},
		{
			name:             "blocked_domain_exact_match",
			opts:             []URLFilterOption{WithBlocklist("malware.com")},
			input:            "Do not visit http://malware.com please",
			expectedPass:     false,
			expectedTripwire: true, // Default true
		},
		{
			name:             "blocked_domain_wildcard",
			opts:             []URLFilterOption{WithBlocklist("*.bad.com")},
			input:            "Visit sub.bad.com now",
			expectedPass:     false,
			expectedTripwire: true,
		},
		{
			name:             "allowed_domain_passes",
			opts:             []URLFilterOption{WithAllowlist("good.com"), WithBlocklist("bad.com")},
			input:            "Visit good.com",
			expectedPass:     true,
			expectedTripwire: false,
		},
		{
			name:  "block_all_except_allowlist",
			opts:  []URLFilterOption{WithAllowlist("good.com")},
			input: "Visit example.com", // should be blocked because allowlist is present so implicit block-others?
			// Implementation check:
			// if len(Allowlist)>0 { if !match(allow) { block } }
			expectedPass:     false,
			expectedTripwire: true,
		},
		{
			name:             "multiple_urls_all_allowed",
			opts:             []URLFilterOption{WithAllowlist("good.com", "ok.com")},
			input:            "http://good.com and https://ok.com",
			expectedPass:     true,
			expectedTripwire: false,
		},
		{
			name:             "multiple_urls_one_blocked",
			opts:             []URLFilterOption{WithAllowlist("good.com")},
			input:            "good.com and bad.com",
			expectedPass:     false,
			expectedTripwire: true,
		},
		{
			name:             "url_without_protocol",
			opts:             []URLFilterOption{WithBlocklist("evil.com")},
			input:            "Go to evil.com",
			expectedPass:     false,
			expectedTripwire: true,
		},
		{
			name: "url_with_subdomain",
			opts: []URLFilterOption{WithBlocklist("example.com")},
			// Wait, logic says: exact match OR wildcard match.
			// exact "example.com" does NOT match "sub.example.com" unless wildcard used.
			input:        "sub.example.com",
			expectedPass: true, // Should pass if no wildcard used
		},
		{
			name:             "wildcard_matches_all_subdomains",
			opts:             []URLFilterOption{WithBlocklist("*.example.com")},
			input:            "deep.nested.sub.example.com",
			expectedPass:     false,
			expectedTripwire: true,
		},
		{
			name:             "explicit_tripwire_off",
			opts:             []URLFilterOption{WithBlocklist("bad.com"), WithURLTripwire(false)},
			input:            "bad.com",
			expectedPass:     false,
			expectedTripwire: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			guard := NewURLFilter(tt.opts...)

			if guard == nil {
				t.Fatal("NewURLFilter returned nil")
			}

			ctx := context.Background()
			result, err := guard.Func(ctx, tt.input)

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result == nil {
				t.Fatal("Result is nil")
			}

			if result.Passed != tt.expectedPass {
				t.Errorf("Expected Passed=%v, got %v. Message: %s",
					tt.expectedPass, result.Passed, result.Message)
			}

			if result.TripwireTriggered != tt.expectedTripwire {
				t.Errorf("Expected TripwireTriggered=%v, got %v",
					tt.expectedTripwire, result.TripwireTriggered)
			}
		})
	}
}

func TestURLFilterOptions(t *testing.T) {
	t.Run("WithBlocklist", func(t *testing.T) {
		c := &URLFilterConfig{}
		WithBlocklist("a.com")(c)
		if len(c.Blocklist) != 1 {
			t.Error("Blocklist not updated")
		}
	})
	// ... other options
}
