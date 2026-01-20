// Package function provides helpers to create type-safe tools from Go functions.
package function

import (
	"encoding/json"
	"fmt"

	agents "github.com/MitulShah1/openai-agents-go"
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
) agents.Tool {
	var zero T
	schema := jsonschema.GenerateSchema(zero)

	// Convert schema to map needed by Tool
	paramsMap, _ := schema.ToMap()

	return agents.Tool{
		Name:        functionName,
		Description: description,
		Parameters:  paramsMap,
		Callback: func(argsMap map[string]any, _ agents.ContextVariables) (any, error) {
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

// Runnable is a function that takes arguments and returns a result
type Runnable[T any] func(T) (any, error)

// Example usage:
//
// type WeatherArgs struct {
//     City string `json:"city" jsonschema:"description=The city name"`
// }
//
// tool := function.FromFunc("get_weather", "Get weather", func(args WeatherArgs) (any, error) {
//     return "Sunny", nil
// })
