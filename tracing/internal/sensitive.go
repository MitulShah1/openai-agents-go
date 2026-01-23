package internal

// RedactSensitive redacts sensitive data if includeSensitive is false.
func RedactSensitive(v any, includeSensitive bool) any {
	if includeSensitive {
		return v
	}
	if v == nil {
		return nil
	}
	return "[REDACTED]"
}
