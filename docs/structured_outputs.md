# Structured Outputs

Get type-safe JSON responses from agents with schema validation.

## Overview

Structured outputs ensure agents return JSON that matches your schema, enabling reliable parsing and type safety.

## Quick Example

```go
import "github.com/MitulShah1/openai-agents-go/internal/jsonschema"

// Define schema
schema := jsonschema.Object().
    WithProperty("answer", jsonschema.Integer()).
    WithProperty("reasoning", jsonschema.String()).
    WithRequired("answer", "reasoning")

// Create agent with structured output
agent := agents.NewAgent("Math Tutor")
agent.ResponseFormat = jsonschema.JSONSchema("math_response", schema)

// Parse response
type MathResponse struct {
    Answer    int    `json:"answer"`
    Reasoning string `json:"reasoning"`
}
```

## Schema Builder

The `jsonschema` package provides a fluent API for building schemas:

```go
// Object with properties
schema := jsonschema.Object().
    WithProperty("name", jsonschema.String()).
    WithProperty("age", jsonschema.Integer()).
    WithRequired("name")

// Arrays
jsonschema.Array(jsonschema.String())

// Enums
jsonschema.String().WithEnum("small", "medium", "large")
```

## Complete Examples

See examples:- `examples/06_structured_output/` - Basic usage
- `examples/07_complex_schema/` - Nested schemas

## API Reference

See [API Reference](ref/types.md) for detailed documentation.
