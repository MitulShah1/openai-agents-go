package exporter

import "fmt"

// ExportError represents an error during trace/span export.
type ExportError struct {
	StatusCode int
	Body       string
	Err        error
	Retryable  bool
}

func (e *ExportError) Error() string {
	if e.Err != nil {
		if e.Body != "" {
			return fmt.Sprintf("trace export failed (status=%d, retryable=%v): %v: %s",
				e.StatusCode, e.Retryable, e.Err, e.Body)
		}
		return fmt.Sprintf("trace export failed (status=%d, retryable=%v): %v",
			e.StatusCode, e.Retryable, e.Err)
	}
	if e.Body != "" {
		return fmt.Sprintf("trace export failed (status=%d, retryable=%v): %s",
			e.StatusCode, e.Retryable, e.Body)
	}
	return fmt.Sprintf("trace export failed (status=%d, retryable=%v)",
		e.StatusCode, e.Retryable)
}

func (e *ExportError) Unwrap() error { return e.Err }

// IsRetryable returns true if the error is retryable.
func (e *ExportError) IsRetryable() bool { return e.Retryable }
