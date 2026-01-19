// Package jsonschema provides utilities for building JSON schemas for OpenAI's Structured Outputs.
package jsonschema

import (
	"encoding/json"
	"fmt"
)

// Type represents a JSON schema type.
type Type string

// JSON schema type constants define the available types for schema fields.
const (
	// TypeString represents string data type
	TypeString Type = "string"
	// TypeNumber represents numeric data type with decimals
	TypeNumber Type = "number"
	// TypeInteger represents whole number data type
	TypeInteger Type = "integer"
	// TypeBoolean represents true/false data type
	TypeBoolean Type = "boolean"
	// TypeObject represents an object/map data type
	TypeObject Type = "object"
	// TypeArray represents an array/list data type
	TypeArray Type = "array"
	// TypeNull represents a null value
	TypeNull Type = "null"
)

// Schema represents a JSON schema.
type Schema struct {
	Type                 Type               `json:"type,omitempty"`
	Description          string             `json:"description,omitempty"`
	Properties           map[string]*Schema `json:"properties,omitempty"`
	Required             []string           `json:"required,omitempty"`
	Items                *Schema            `json:"items,omitempty"`
	Enum                 []any              `json:"enum,omitempty"`
	AdditionalProperties *bool              `json:"additionalProperties,omitempty"`
	MinLength            *int               `json:"minLength,omitempty"`
	MaxLength            *int               `json:"maxLength,omitempty"`
	Minimum              *float64           `json:"minimum,omitempty"`
	Maximum              *float64           `json:"maximum,omitempty"`
	Pattern              string             `json:"pattern,omitempty"`
}

// NewSchema creates a new JSON schema with the given type.
func NewSchema(t Type) *Schema {
	return &Schema{Type: t}
}

// String creates a string type schema.
func String() *Schema {
	return &Schema{Type: TypeString}
}

// Number creates a number type schema.
func Number() *Schema {
	return &Schema{Type: TypeNumber}
}

// Integer creates an integer type schema.
func Integer() *Schema {
	return &Schema{Type: TypeInteger}
}

// Boolean creates a boolean type schema.
func Boolean() *Schema {
	return &Schema{Type: TypeBoolean}
}

// Object creates an object type schema.
func Object() *Schema {
	falseVal := false
	return &Schema{
		Type:                 TypeObject,
		Properties:           make(map[string]*Schema),
		AdditionalProperties: &falseVal,
	}
}

// Array creates an array type schema.
func Array(items *Schema) *Schema {
	return &Schema{
		Type:  TypeArray,
		Items: items,
	}
}

// WithDescription sets the description of the schema.
func (s *Schema) WithDescription(desc string) *Schema {
	s.Description = desc
	return s
}

// WithProperty adds a property to an object schema.
func (s *Schema) WithProperty(name string, schema *Schema) *Schema {
	if s.Properties == nil {
		s.Properties = make(map[string]*Schema)
	}
	s.Properties[name] = schema
	return s
}

// WithRequired marks fields as required in an object schema.
func (s *Schema) WithRequired(fields ...string) *Schema {
	s.Required = append(s.Required, fields...)
	return s
}

// WithEnum sets the allowed enum values.
func (s *Schema) WithEnum(values ...any) *Schema {
	s.Enum = values
	return s
}

// WithAdditionalProperties controls whether additional properties are allowed.
func (s *Schema) WithAdditionalProperties(allowed bool) *Schema {
	s.AdditionalProperties = &allowed
	return s
}

// WithMinLength sets the minimum string length.
func (s *Schema) WithMinLength(minLen int) *Schema {
	s.MinLength = &minLen
	return s
}

// WithMaxLength sets the maximum string length.
func (s *Schema) WithMaxLength(maxLen int) *Schema {
	s.MaxLength = &maxLen
	return s
}

// WithMinimum sets the minimum number value.
func (s *Schema) WithMinimum(minVal float64) *Schema {
	s.Minimum = &minVal
	return s
}

// WithMaximum sets the maximum number value.
func (s *Schema) WithMaximum(maxVal float64) *Schema {
	s.Maximum = &maxVal
	return s
}

// WithPattern sets a regex pattern for string validation.
func (s *Schema) WithPattern(pattern string) *Schema {
	s.Pattern = pattern
	return s
}

// ToJSON converts the schema to JSON string.
func (s *Schema) ToJSON() (string, error) {
	data, err := json.Marshal(s)
	if err != nil {
		return "", fmt.Errorf("failed to marshal schema: %w", err)
	}
	return string(data), nil
}

// ToMap converts the schema to a map[string]any for use with OpenAI API.
func (s *Schema) ToMap() (map[string]any, error) {
	data, err := json.Marshal(s)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal schema: %w", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal schema: %w", err)
	}

	return result, nil
}

// Validate performs basic validation on the schema.
func (s *Schema) Validate() error {
	if s.Type == "" {
		return fmt.Errorf("schema type is required")
	}

	if s.Type == TypeObject && s.Properties == nil {
		return fmt.Errorf("object schema must have properties")
	}

	if s.Type == TypeArray && s.Items == nil {
		return fmt.Errorf("array schema must have items")
	}

	// Validate nested schemas
	if s.Properties != nil {
		for name, prop := range s.Properties {
			if err := prop.Validate(); err != nil {
				return fmt.Errorf("invalid property %q: %w", name, err)
			}
		}
	}

	if s.Items != nil {
		if err := s.Items.Validate(); err != nil {
			return fmt.Errorf("invalid items schema: %w", err)
		}
	}

	return nil
}
