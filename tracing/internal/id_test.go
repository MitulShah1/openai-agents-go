package internal

import (
	"strings"
	"testing"
)

func TestGenTraceID(t *testing.T) {
	id := GenTraceID()

	if id == "" {
		t.Error("GenTraceID returned empty string")
	}

	if !strings.HasPrefix(id, "trace_") {
		t.Errorf("Expected trace ID to start with 'trace_', got %q", id)
	}

	// Should be trace_ + 32 hex characters
	if len(id) != 38 { // "trace_" (6) + 32 hex chars
		t.Errorf("Expected trace ID length 38, got %d", len(id))
	}
}

func TestGenTraceIDUnique(t *testing.T) {
	id1 := GenTraceID()
	id2 := GenTraceID()

	if id1 == id2 {
		t.Error("GenTraceID generated duplicate IDs")
	}
}

func TestGenSpanID(t *testing.T) {
	id := GenSpanID()

	if id == "" {
		t.Error("GenSpanID returned empty string")
	}

	if !strings.HasPrefix(id, "span_") {
		t.Errorf("Expected span ID to start with 'span_', got %q", id)
	}

	// Should be span_ + 24 hex characters
	if len(id) != 29 { // "span_" (5) + 24 hex chars
		t.Errorf("Expected span ID length 29, got %d", len(id))
	}
}

func TestGenSpanIDUnique(t *testing.T) {
	id1 := GenSpanID()
	id2 := GenSpanID()

	if id1 == id2 {
		t.Error("GenSpanID generated duplicate IDs")
	}
}

func TestGenGroupID(t *testing.T) {
	id := GenGroupID()

	if id == "" {
		t.Error("GenGroupID returned empty string")
	}

	if !strings.HasPrefix(id, "group_") {
		t.Errorf("Expected group ID to start with 'group_', got %q", id)
	}

	// Should be group_ + 24 hex characters
	if len(id) != 30 { // "group_" (6) + 24 hex chars
		t.Errorf("Expected group ID length 30, got %d", len(id))
	}
}

func TestGenGroupIDUnique(t *testing.T) {
	id1 := GenGroupID()
	id2 := GenGroupID()

	if id1 == id2 {
		t.Error("GenGroupID generated duplicate IDs")
	}
}
