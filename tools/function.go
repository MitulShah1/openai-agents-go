// Package tools provides primitives for defining and using tools.
package tools

import (
	"encoding/json"
	"fmt"

	"github.com/MitulShah1/openai-agents-go/jsonschema"
)

// FromFunc creates a Tool from a Go function and a struct defining its arguments.
// T is the type of the arguments struct.
// functionName: name of the tool
// description: description of the tool
// fn: the function to execute. It receives the arguments struct and returns (any, error)
func FromFunc[T any](
	functionName string,
	description string,
	fn func(args T) (any, error),
) Tool {
	var zero T
	schema := jsonschema.GenerateSchema(zero)

	// Convert schema to map needed by Tool
	paramsMap, _ := schema.ToMap()

	return Tool{
		Name:        functionName,
		Description: description,
		Parameters:  paramsMap,
		Callback: func(argsMap map[string]any, _ ContextVariables) (any, error) {
			// Convert generic map to typed struct
			var args T

			// We need to re-marshal to JSON and unmarshal into the struct to handle types correctly
			// This incurs some overhead but ensures correctness with standard json tags
			jsonBytes, err := json.Marshal(argsMap)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal args: %w", err)
			}

			if err := json.Unmarshal(jsonBytes, &args); err != nil {
				return nil, fmt.Errorf("failed to unmarshal args into struct: %w", err)
			}

			return fn(args)
		},
	}
}
