package jsonschema

import (
	"fmt"
)

// ResponseFormat defines how the LLM should structure its response.
type ResponseFormat struct {
	// Type is either "text" or "json_schema"
	Type string

	// JSONSchema is the schema definition (only if Type is "json_schema")
	JSONSchema *JSONSchemaFormat
}

// JSONSchemaFormat defines the JSON schema specification for structured outputs.
//
//nolint:revive // JSONSchemaFormat is clearer than Format
type JSONSchemaFormat struct {
	// Name of the schema (required by OpenAI)
	Name string

	// Description of what the schema represents
	Description string

	// Schema is the JSON schema definition
	Schema *Schema

	// Strict enables strict schema adherence (recommended for OpenAI)
	Strict bool
}

// Text creates a text response format (default behavior).
func Text() *ResponseFormat {
	return &ResponseFormat{
		Type: "text",
	}
}

// JSONSchema creates a structured JSON response format.
func JSONSchema(name string, schema *Schema) *ResponseFormat {
	return &ResponseFormat{
		Type: "json_schema",
		JSONSchema: &JSONSchemaFormat{
			Name:   name,
			Schema: schema,
			Strict: true, // Default to strict mode
		},
	}
}

// WithDescription sets the description for the JSON schema.
func (r *ResponseFormat) WithDescription(desc string) *ResponseFormat {
	if r.JSONSchema != nil {
		r.JSONSchema.Description = desc
	}
	return r
}

// WithStrict sets the strict mode for the JSON schema.
func (r *ResponseFormat) WithStrict(strict bool) *ResponseFormat {
	if r.JSONSchema != nil {
		r.JSONSchema.Strict = strict
	}
	return r
}

// Validate checks if the response format is valid.
func (r *ResponseFormat) Validate() error {
	if r.Type != "text" && r.Type != "json_schema" {
		return fmt.Errorf("invalid response format type: %s (must be 'text' or 'json_schema')", r.Type)
	}

	if r.Type == "json_schema" {
		if r.JSONSchema == nil {
			return fmt.Errorf("json_schema type requires JSONSchema to be set")
		}
		if r.JSONSchema.Name == "" {
			return fmt.Errorf("json_schema requires a name")
		}
		if r.JSONSchema.Schema == nil {
			return fmt.Errorf("json_schema requires a schema")
		}
		if err := r.JSONSchema.Schema.Validate(); err != nil {
			return fmt.Errorf("invalid schema: %w", err)
		}
	}

	return nil
}

// ToOpenAIParam converts the ResponseFormat to OpenAI's format.
func (r *ResponseFormat) ToOpenAIParam() (any, error) {
	if err := r.Validate(); err != nil {
		return nil, fmt.Errorf("invalid response format: %w", err)
	}

	if r.Type == "text" {
		return map[string]any{
			"type": "text",
		}, nil
	}

	// Convert schema to map
	schemaMap, err := r.JSONSchema.Schema.ToMap()
	if err != nil {
		return nil, fmt.Errorf("failed to convert schema: %w", err)
	}

	result := map[string]any{
		"type": "json_schema",
		"json_schema": map[string]any{
			"name":   r.JSONSchema.Name,
			"schema": schemaMap,
			"strict": r.JSONSchema.Strict,
		},
	}

	if r.JSONSchema.Description != "" {
		result["json_schema"].(map[string]any)["description"] = r.JSONSchema.Description
	}

	return result, nil
}
