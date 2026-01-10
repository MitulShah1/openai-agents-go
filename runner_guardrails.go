package agents

import (
	"context"
	"fmt"

	"github.com/MitulShah1/openai-agents-go/guardrail"
)

// Returns an InputGuardrailTripwireError if any guardrail triggers its tripwire.
func (r *Runner) runInputGuardrails(ctx context.Context, agent *Agent, input string) error {
	for _, gr := range agent.InputGuardrails {
		result, err := gr.Func(ctx, input)
		if err != nil {
			return fmt.Errorf("guardrail '%s' failed: %w", gr.Name, err)
		}

		if result.TripwireTriggered {
			return &guardrail.InputGuardrailTripwireError{
				GuardrailName: gr.Name,
				Message:       result.Message,
				Result:        result,
			}
		}
	}
	return nil
}

// runOutputGuardrails executes output validation guardrails.
// Returns an OutputGuardrailTripwireError if any guardrail triggers its tripwire.
func (r *Runner) runOutputGuardrails(ctx context.Context, agent *Agent, output string) error {
	for _, gr := range agent.OutputGuardrails {
		result, err := gr.Func(ctx, output)
		if err != nil {
			return fmt.Errorf("guardrail '%s' failed: %w", gr.Name, err)
		}

		if result.TripwireTriggered {
			return &guardrail.OutputGuardrailTripwireError{
				GuardrailName: gr.Name,
				Message:       result.Message,
				Result:        result,
			}
		}
	}
	return nil
}
