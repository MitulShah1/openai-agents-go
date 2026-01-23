package internal

import (
	"encoding/hex"
	"fmt"
	"strings"
)

// ValidationError represents a validation error.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error for %s: %s", e.Field, e.Message)
}

// IsValidTraceID checks if a trace ID has the correct format: trace_<32 hex chars>
func IsValidTraceID(id string) bool {
	if len(id) != 38 || !strings.HasPrefix(id, "trace_") {
		return false
	}
	_, err := hex.DecodeString(id[6:])
	return err == nil
}

// IsValidGroupID checks if a group ID has the correct format: group_<24 hex chars>
func IsValidGroupID(id string) bool {
	if len(id) != 30 || !strings.HasPrefix(id, "group_") {
		return false
	}
	_, err := hex.DecodeString(id[6:])
	return err == nil
}

// IsValidSpanID checks if a span ID has the correct format: span_<24 hex chars>
func IsValidSpanID(id string) bool {
	if len(id) != 29 || !strings.HasPrefix(id, "span_") {
		return false
	}
	_, err := hex.DecodeString(id[5:])
	return err == nil
}
