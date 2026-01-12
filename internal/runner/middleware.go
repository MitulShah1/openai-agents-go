package runner

import (
	"context"
	"fmt"
)

// GuardrailFunc is a function that validates input/output
type GuardrailFunc func(ctx context.Context, text string) (GuardrailResult, error)

// GuardrailResult represents the result of a guardrail check
type GuardrailResult struct {
	TripwireTriggered bool
	Message           string
}

// Guardrail represents a validation rule
type Guardrail struct {
	Name string
	Func GuardrailFunc
}

// LifecycleHook is a function called before/after execution
// This matches the agents.LifecycleFunc signature
type LifecycleHook = func(ctx context.Context, agent any) error

// GuardrailExecutor runs input or output guardrails
type GuardrailExecutor struct {
	guardrails []*Guardrail
	errorType  string // "input" or "output"
}

// NewGuardrailExecutor creates a new guardrail executor
func NewGuardrailExecutor(guardrails []*Guardrail, errorType string) *GuardrailExecutor {
	return &GuardrailExecutor{
		guardrails: guardrails,
		errorType:  errorType,
	}
}

// Execute runs all guardrails and returns an error if any tripwire is triggered
func (ge *GuardrailExecutor) Execute(ctx context.Context, text string) error {
	if len(ge.guardrails) == 0 {
		return nil
	}

	for _, gr := range ge.guardrails {
		result, err := gr.Func(ctx, text)
		if err != nil {
			return fmt.Errorf("guardrail '%s' failed: %w", gr.Name, err)
		}

		if result.TripwireTriggered {
			// Return a structured error that will be handled by the caller
			return &GuardrailError{
				Type:          ge.errorType,
				GuardrailName: gr.Name,
				Message:       result.Message,
				Result:        result,
			}
		}
	}

	return nil
}

// GuardrailError represents a guardrail violation
type GuardrailError struct {
	Type          string // "input" or "output"
	GuardrailName string
	Message       string
	Result        GuardrailResult
}

// Error implements the error interface
func (e *GuardrailError) Error() string {
	return fmt.Sprintf("%s guardrail '%s' triggered: %s", e.Type, e.GuardrailName, e.Message)
}

// ExecuteLifecycleHook executes a lifecycle hook if it's not nil
func ExecuteLifecycleHook(ctx context.Context, hook LifecycleHook, agent any, hookName string) error {
	if hook == nil {
		return nil
	}

	if err := hook(ctx, agent); err != nil {
		return fmt.Errorf("%s hook failed: %w", hookName, err)
	}

	return nil
}
