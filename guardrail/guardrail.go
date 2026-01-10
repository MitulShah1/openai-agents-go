// Package guardrail provides input/output validation for agents.
package guardrail

import (
	"context"
	"fmt"
)

// Func is a function that validates input/output and returns a result.
// It receives the input string and returns a Result or an error.
type Func func(ctx context.Context, input string) (*Result, error)

// Result contains the outcome of a guardrail validation.
type Result struct {
	// Passed indicates whether the validation passed
	Passed bool

	// TripwireTriggered indicates if this guardrail should halt execution
	TripwireTriggered bool

	// Message provides details about the validation result
	Message string

	// Metadata contains additional information about the validation
	Metadata map[string]any
}

// Guardrail wraps a validation function with configuration.
type Guardrail struct {
	// Name identifies this guardrail for error reporting
	Name string

	// Func is the validation function to execute
	Func Func

	// RunInParallel determines if this guardrail can run concurrently
	// Only applicable for input guardrails; output guardrails always run sequentially
	RunInParallel bool
}

// InputGuardrailTripwireError is raised when an input guardrail's tripwire is triggered.
type InputGuardrailTripwireError struct {
	GuardrailName string
	Message       string
	Result        *Result
}

// Error implements the error interface.
func (e *InputGuardrailTripwireError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("input guardrail '%s' triggered: %s", e.GuardrailName, e.Message)
	}
	return fmt.Sprintf("input guardrail '%s' triggered", e.GuardrailName)
}

// OutputGuardrailTripwireError is raised when an output guardrail's tripwire is triggered.
type OutputGuardrailTripwireError struct {
	GuardrailName string
	Message       string
	Result        *Result
}

// Error implements the error interface.
func (e *OutputGuardrailTripwireError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("output guardrail '%s' triggered: %s", e.GuardrailName, e.Message)
	}
	return fmt.Sprintf("output guardrail '%s' triggered", e.GuardrailName)
}

// NewGuardrail creates a new guardrail with the given name and validation function.
func NewGuardrail(name string, fn Func) *Guardrail {
	return &Guardrail{
		Name:          name,
		Func:          fn,
		RunInParallel: false, // Default to blocking
	}
}

// WithParallel sets the guardrail to run in parallel (for input guardrails only).
func (g *Guardrail) WithParallel(parallel bool) *Guardrail {
	g.RunInParallel = parallel
	return g
}
