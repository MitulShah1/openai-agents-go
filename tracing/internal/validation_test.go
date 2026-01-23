package internal

import (
	"strings"
	"testing"
)

func TestIsValidTraceID(t *testing.T) {
	tests := []struct {
		id    string
		valid bool
	}{
		{"trace_" + strings.Repeat("a", 32), true},       // Valid
		{"trace_0123456789abcdef0123456789abcdef", true}, // Valid hex
		{"trace_", false},                           // Too short
		{"trace_abc", false},                        // Too short
		{"span_" + strings.Repeat("a", 32), false},  // Wrong prefix
		{"trace_" + strings.Repeat("g", 32), false}, // Invalid hex (g)
		{"trace_" + strings.Repeat("a", 31), false}, // Too short
		{"trace_" + strings.Repeat("a", 33), false}, // Too long
		{"", false},        // Empty
		{"invalid", false}, // No prefix
	}

	for _, tt := range tests {
		result := IsValidTraceID(tt.id)
		if result != tt.valid {
			t.Errorf("IsValidTraceID(%q) = %v, want %v", tt.id, result, tt.valid)
		}
	}
}

func TestIsValidSpanID(t *testing.T) {
	tests := []struct {
		id    string
		valid bool
	}{
		{"span_" + strings.Repeat("a", 24), true},   // Valid (24 hex chars)
		{"span_0123456789abcdef01234567", true},     // Valid hex (24 chars)
		{"span_", false},                            // Too short
		{"span_abc", false},                         // Too short
		{"trace_" + strings.Repeat("a", 24), false}, // Wrong prefix
		{"span_" + strings.Repeat("g", 24), false},  // Invalid hex
		{"span_" + strings.Repeat("a", 23), false},  // Too short
		{"span_" + strings.Repeat("a", 25), false},  // Too long
		{"", false}, // Empty
	}

	for _, tt := range tests {
		result := IsValidSpanID(tt.id)
		if result != tt.valid {
			t.Errorf("IsValidSpanID(%q) = %v, want %v", tt.id, result, tt.valid)
		}
	}
}

func TestIsValidGroupID(t *testing.T) {
	tests := []struct {
		id    string
		valid bool
	}{
		{"group_" + strings.Repeat("a", 24), true},  // Valid
		{"group_0123456789abcdef01234567", true},    // Valid hex
		{"group_", false},                           // Too short
		{"group_abc", false},                        // Too short
		{"trace_" + strings.Repeat("a", 24), false}, // Wrong prefix
		{"group_" + strings.Repeat("g", 24), false}, // Invalid hex
		{"group_" + strings.Repeat("a", 23), false}, // Too short
		{"group_" + strings.Repeat("a", 25), false}, // Too long
		{"", false}, // Empty
	}

	for _, tt := range tests {
		result := IsValidGroupID(tt.id)
		if result != tt.valid {
			t.Errorf("IsValidGroupID(%q) = %v, want %v", tt.id, result, tt.valid)
		}
	}
}
