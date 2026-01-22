package jsonschema

import (
	"reflect"
	"strings"
)

// GenerateSchema creates a JSON schema from a Go struct using reflection.
// It supports `json` tags for field names and `jsonschema` tags for metadata.
// Supported jsonschema tags:
// - description: Set the field description
// - enum: Comma-separated list of allowed values
// - required: Mark field as required (true/false) - defaults to false unless "required" is present in json tag or pointer is nil
func GenerateSchema(v any) *Schema {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return generateTypeSchema(t)
}

func generateTypeSchema(t reflect.Type) *Schema {
	switch t.Kind() {
	case reflect.String:
		return String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return Integer()
	case reflect.Float32, reflect.Float64:
		return Number()
	case reflect.Bool:
		return Boolean()
	case reflect.Slice, reflect.Array:
		elemSchema := generateTypeSchema(t.Elem())
		return Array(elemSchema)
	case reflect.Struct:
		return generateStructSchema(t)
	case reflect.Map:
		// Go maps are generally strictly typed in this context, but specialized object with additionalProperties typically
		// For now, treat as object with generic values or strict string keys
		// valSchema := generateTypeSchema(t.Elem()) // Unused for now as we can't strongly type additionalProperties yet
		s := Object()

		// schema.go currently defines AdditionalProperties as *bool.
		// Ideally we update Schema to support schema for AdditionalProperties, but for now let's assume loose maps are "object"
		// If we want strict typed maps, we need to update Schema struct.
		// For simple tool use, let's treat generic maps as freeform objects
		trueVal := true
		s.AdditionalProperties = &trueVal
		return s
	default:
		// Default to string for unknown types for safety? Or nil?
		return String()
	}
}

func generateStructSchema(t reflect.Type) *Schema {
	s := Object()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if field.PkgPath != "" {
			continue
		}

		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}

		name := field.Name
		required := false

		// Parse json tag
		if jsonTag != "" {
			parts := strings.Split(jsonTag, ",")
			name = parts[0]
			// Check for omitempty
			if len(parts) > 1 {
				for _, p := range parts[1:] {
					if p == "omitempty" {
						required = false
					}
				}
			}
		}

		// If NOT a pointer and omitempty NOT found, assume required
		if field.Type.Kind() != reflect.Ptr && jsonTag != "" {
			if !strings.Contains(jsonTag, "omitempty") {
				required = true
			}
		} else if field.Type.Kind() != reflect.Ptr {
			// No json tag, default required
			required = true
		}

		propSchema := generateTypeSchema(field.Type)

		// Parse jsonschema tag
		schemaTag := field.Tag.Get("jsonschema")
		if schemaTag != "" {
			parseSchemaTag(propSchema, schemaTag)
		}

		s.WithProperty(name, propSchema)
		if required {
			s.WithRequired(name)
		}
	}

	return s
}

func parseSchemaTag(s *Schema, tag string) {
	// Simple comma separated k=v parser
	// limititation: doesn't handle values with commas
	parts := strings.Split(tag, ",")
	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 {
			key, val := strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1])
			switch key {
			case "description":
				s.WithDescription(val)
			case "enum":
				enums := strings.Split(val, "|")
				anyEnums := make([]any, len(enums))
				for i, v := range enums {
					anyEnums[i] = v
				}
				s.WithEnum(anyEnums...)
			}
		}
	}
}
