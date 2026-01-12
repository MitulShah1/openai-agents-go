// Package agents provides the core types and functionality for the OpenAI Agents Go SDK.
package agents

import (
	"context"

	"github.com/MitulShah1/openai-agents-go/guardrail"
	"github.com/MitulShah1/openai-agents-go/internal/jsonschema"
)

const (
	// DefaultModel is the default OpenAI model used for agents
	DefaultModel = "gpt-4o"

	// DefaultInstructions is the default instruction for agents
	DefaultInstructions = "You are a helpful agent."
)

// LifecycleFunc is called before/after agent execution
type LifecycleFunc func(ctx context.Context, agent *Agent) error

// Agent represents an entity that can process messages and use tools.
type Agent struct {
	// Name is the name of the agent.
	Name string

	// Model is the OpenAI model to use (default: "gpt-4o").
	Model string

	// Instructions can be a string or a function that returns a string.
	// Function signature: func(context.Context) string or func() string.
	Instructions any

	// Tools is a list of tools available to the agent.
	Tools []Tool

	// ParallelToolCalls determines if tools can be called in parallel.
	// Can be overridden by RunConfig.
	ParallelToolCalls bool

	// Temperature controls randomness (0.0 to 2.0)
	// If nil, uses model default
	Temperature *float64

	// MaxTokens limits response length
	// If nil, uses model default
	MaxTokens *int

	// ResponseFormat defines the structure of the response (for structured outputs)
	// If nil, responses will be unstructured text
	ResponseFormat *jsonschema.ResponseFormat

	// OnBeforeRun is called before the agent starts execution
	OnBeforeRun LifecycleFunc

	// OnAfterRun is called after the agent completes execution
	OnAfterRun LifecycleFunc

	// InputGuardrails validate user input before agent execution
	// These run on the first agent in a handoff chain
	InputGuardrails []*guardrail.Guardrail

	// OutputGuardrails validate agent output after execution
	// These run on the final agent in a handoff chain
	OutputGuardrails []*guardrail.Guardrail
}

// NewAgent creates a new Agent with default values.
func NewAgent(name string) *Agent {
	return &Agent{
		Name:              name,
		Model:             DefaultModel,
		Instructions:      DefaultInstructions,
		ParallelToolCalls: true,
	}
}

// GetInstructions returns the instructions for the agent, resolving functions if necessary.
func (a *Agent) GetInstructions(ctx context.Context) string {
	switch v := a.Instructions.(type) {
	case string:
		return v
	case func() string:
		return v()
	case func(context.Context) string:
		return v(ctx)
	default:
		return DefaultInstructions
	}
}
