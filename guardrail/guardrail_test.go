package guardrail

import (
	"context"
	"testing"
)

func TestNewGuardrail(t *testing.T) {
	fn := func(_ context.Context, input string) (*Result, error) {
		return &Result{
			Passed:            true,
			TripwireTriggered: false,
			Message:           "test",
		}, nil
	}

	g := NewGuardrail("test", fn)

	if g.Name != "test" {
		t.Errorf("expected name 'test', got '%s'", g.Name)
	}

	if g.Func == nil {
		t.Error("expected func to be set")
	}

	if g.RunInParallel {
		t.Error("expected RunInParallel to be false by default")
	}
}

func TestGuardrailWithParallel(t *testing.T) {
	fn := func(ctx context.Context, input string) (*Result, error) {
		return &Result{Passed: true}, nil
	}

	g := NewGuardrail("test", fn).WithParallel(true)

	if !g.RunInParallel {
		t.Error("expected RunInParallel to be true")
	}
}

func TestInputGuardrailTripwireError(t *testing.T) {
	err := &InputGuardrailTripwireError{
		GuardrailName: "test",
		Message:       "failed validation",
	}

	expected := "input guardrail 'test' triggered: failed validation"
	if err.Error() != expected {
		t.Errorf("expected error message '%s', got '%s'", expected, err.Error())
	}

	// Test without message
	err2 := &InputGuardrailTripwireError{
		GuardrailName: "test",
	}

	expected2 := "input guardrail 'test' triggered"
	if err2.Error() != expected2 {
		t.Errorf("expected error message '%s', got '%s'", expected2, err2.Error())
	}
}

func TestOutputGuardrailTripwireError(t *testing.T) {
	err := &OutputGuardrailTripwireError{
		GuardrailName: "test",
		Message:       "failed validation",
	}

	expected := "output guardrail 'test' triggered: failed validation"
	if err.Error() != expected {
		t.Errorf("expected error message '%s', got '%s'", expected, err.Error())
	}

	// Test without message
	err2 := &OutputGuardrailTripwireError{
		GuardrailName: "test",
	}

	expected2 := "output guardrail 'test' triggered"
	if err2.Error() != expected2 {
		t.Errorf("expected error message '%s', got '%s'", expected2, err2.Error())
	}
}

func TestGuardrailFunc(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantPass bool
	}{
		{"pass", "valid input", true},
		{"fail", "invalid input", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := func(ctx context.Context, input string) (*Result, error) {
				return &Result{
					Passed:            input == "valid input",
					TripwireTriggered: input != "valid input",
					Message:           "validation result",
				}, nil
			}

			g := NewGuardrail("test", fn)
			result, err := g.Func(context.Background(), tt.input)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.Passed != tt.wantPass {
				t.Errorf("expected pass=%v, got %v", tt.wantPass, result.Passed)
			}
		})
	}
}
