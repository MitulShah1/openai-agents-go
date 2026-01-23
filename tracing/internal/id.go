// Package internal contains internal utilities for the tracing package.
package internal

import (
	"crypto/rand"
	"encoding/hex"
)

// GenTraceID generates a new trace ID.
// Format: trace_<32 hex characters>
func GenTraceID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return "trace_" + hex.EncodeToString(b[:])
}

// GenSpanID generates a new span ID.
// Format: span_<24 hex characters>
func GenSpanID() string {
	var b [12]byte
	_, _ = rand.Read(b[:])
	return "span_" + hex.EncodeToString(b[:])
}

// GenGroupID generates a new group ID.
// Format: group_<24 hex characters>
func GenGroupID() string {
	var b [12]byte
	_, _ = rand.Read(b[:])
	return "group_" + hex.EncodeToString(b[:])
}
