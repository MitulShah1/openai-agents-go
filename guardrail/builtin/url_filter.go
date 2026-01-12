package builtin

import (
	"context"
	"net/url"
	"strings"

	"github.com/MitulShah1/openai-agents-go/guardrail"
)

// URLFilterConfig configures the URL filtering guardrail.
type URLFilterConfig struct {
	// Tripwire determines if detection should halt execution
	Tripwire bool

	// Blocklist contains URL patterns that should be blocked
	// Supports wildcards: *.example.com blocks all subdomains
	Blocklist []string

	// Allowlist contains URL patterns that are explicitly allowed
	// If set, only these URLs are permitted
	Allowlist []string
}

// URLFilterOption is a functional option for URL filtering.
type URLFilterOption func(*URLFilterConfig)

// WithURLTripwire enables tripwire mode.
func WithURLTripwire(enabled bool) URLFilterOption {
	return func(c *URLFilterConfig) {
		c.Tripwire = enabled
	}
}

// WithBlocklist sets the URL blocklist.
func WithBlocklist(patterns ...string) URLFilterOption {
	return func(c *URLFilterConfig) {
		c.Blocklist = append(c.Blocklist, patterns...)
	}
}

// WithAllowlist sets the URL allowlist.
func WithAllowlist(patterns ...string) URLFilterOption {
	return func(c *URLFilterConfig) {
		c.Allowlist = append(c.Allowlist, patterns...)
	}
}

// NewURLFilterGuardrail creates a guardrail that filters URLs based on block/allow lists.
func NewURLFilterGuardrail(opts ...URLFilterOption) *guardrail.Guardrail {
	config := &URLFilterConfig{
		Tripwire: true,
	}

	for _, opt := range opts {
		opt(config)
	}

	return guardrail.NewGuardrail("url_filter", func(_ context.Context, input string) (*guardrail.Result, error) {
		// Extract URLs from input
		urls := extractURLs(input)
		if len(urls) == 0 {
			return &guardrail.Result{
				Passed:            true,
				TripwireTriggered: false,
				Message:           "No URLs found",
			}, nil
		}

		var blocked []string
		var violations []string

		for _, u := range urls {
			// Check allowlist first (if configured)
			if len(config.Allowlist) > 0 {
				if !matchesAnyPattern(u, config.Allowlist) {
					blocked = append(blocked, u)
					violations = append(violations, u+" (not in allowlist)")
				}
				continue
			}

			// Check blocklist
			if len(config.Blocklist) > 0 {
				if matchesAnyPattern(u, config.Blocklist) {
					blocked = append(blocked, u)
					violations = append(violations, u+" (blocked)")
				}
			}
		}

		if len(blocked) > 0 {
			return &guardrail.Result{
				Passed:            false,
				TripwireTriggered: config.Tripwire,
				Message:           "Blocked URLs: " + strings.Join(violations, ", "),
				Metadata: map[string]any{
					"blocked_urls": blocked,
					"total_urls":   len(urls),
				},
			}, nil
		}

		return &guardrail.Result{
			Passed:            true,
			TripwireTriggered: false,
			Message:           "All URLs passed validation",
		}, nil
	})
}

// extractURLs finds URLs in the input text.
func extractURLs(text string) []string {
	var urls []string
	words := strings.Fields(text)

	for _, word := range words {
		// Try to parse as URL
		if u, err := url.Parse(word); err == nil {
			if u.Scheme != "" && u.Host != "" {
				urls = append(urls, u.Host)
			}
		}

		// Also check for common URL patterns without scheme
		if strings.Contains(word, ".com") || strings.Contains(word, ".org") ||
			strings.Contains(word, ".net") || strings.Contains(word, ".io") {
			// Extract domain
			cleaned := strings.TrimPrefix(word, "http://")
			cleaned = strings.TrimPrefix(cleaned, "https://")
			cleaned = strings.Split(cleaned, "/")[0]
			if cleaned != "" && strings.Contains(cleaned, ".") {
				urls = append(urls, cleaned)
			}
		}
	}

	return urls
}

// matchesAnyPattern checks if a URL matches any of the patterns.
// Supports wildcards like *.example.com
func matchesAnyPattern(urlHost string, patterns []string) bool {
	for _, pattern := range patterns {
		if matchesPattern(urlHost, pattern) {
			return true
		}
	}
	return false
}

// matchesPattern checks if a URL matches a pattern with wildcard support.
func matchesPattern(urlHost, pattern string) bool {
	// Exact match
	if urlHost == pattern {
		return true
	}

	// Wildcard match (*.example.com matches sub.example.com)
	if strings.HasPrefix(pattern, "*.") {
		suffix := pattern[2:]
		if urlHost == suffix || strings.HasSuffix(urlHost, "."+suffix) {
			return true
		}
	}

	return false
}
