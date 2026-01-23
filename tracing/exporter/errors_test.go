package exporter

import (
	"errors"
	"testing"
)

func TestExportError(t *testing.T) {
	err := &ExportError{
		Err:        errors.New("test error"),
		StatusCode: 500,
		Body:       "internal server error",
		Retryable:  true,
	}

	// Test Error() method
	errMsg := err.Error()
	if errMsg == "" {
		t.Error("Error() returned empty string")
	}

	// Test Unwrap() method
	unwrapped := err.Unwrap()
	if unwrapped == nil {
		t.Error("Unwrap() returned nil")
	}
	if unwrapped.Error() != "test error" {
		t.Errorf("Expected unwrapped error 'test error', got %q", unwrapped.Error())
	}
}

func TestExportErrorWithoutWrappedError(t *testing.T) {
	err := &ExportError{
		StatusCode: 400,
		Body:       "bad request",
		Retryable:  false,
	}

	// Should still have error message
	errMsg := err.Error()
	if errMsg == "" {
		t.Error("Error() returned empty string")
	}

	// Unwrap should return nil
	if err.Unwrap() != nil {
		t.Error("Expected Unwrap() to return nil when no wrapped error")
	}
}
