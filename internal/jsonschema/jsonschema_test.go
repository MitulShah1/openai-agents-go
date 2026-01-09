package jsonschema

import (
	"encoding/json"
	"testing"
)

func TestNewSchema(t *testing.T) {
	s := NewSchema(TypeString)
	if s.Type != TypeString {
		t.Errorf("expected type %s, got %s", TypeString, s.Type)
	}
}

func TestTypeHelpers(t *testing.T) {
	tests := []struct {
		name     string
		schema   *Schema
		wantType Type
	}{
		{"String", String(), TypeString},
		{"Number", Number(), TypeNumber},
		{"Integer", Integer(), TypeInteger},
		{"Boolean", Boolean(), TypeBoolean},
		{"Object", Object(), TypeObject},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.schema.Type != tt.wantType {
				t.Errorf("expected type %s, got %s", tt.wantType, tt.schema.Type)
			}
		})
	}
}

func TestArray(t *testing.T) {
	s := Array(String())
	if s.Type != TypeArray {
		t.Errorf("expected type array, got %s", s.Type)
	}
	if s.Items == nil {
		t.Error("expected items to be set")
	}
	if s.Items.Type != TypeString {
		t.Errorf("expected items type string, got %s", s.Items.Type)
	}
}

func TestFluentAPI(t *testing.T) {
	s := Object().
		WithDescription("A person object").
		WithProperty("name", String().WithDescription("Person's name")).
		WithProperty("age", Integer().WithDescription("Person's age").WithMinimum(0).WithMaximum(150)).
		WithRequired("name", "age")

	if s.Description != "A person object" {
		t.Errorf("expected description to be set")
	}

	if len(s.Properties) != 2 {
		t.Errorf("expected 2 properties, got %d", len(s.Properties))
	}

	if len(s.Required) != 2 {
		t.Errorf("expected 2 required fields, got %d", len(s.Required))
	}

	nameSchema := s.Properties["name"]
	if nameSchema == nil || nameSchema.Type != TypeString {
		t.Error("name property not set correctly")
	}

	ageSchema := s.Properties["age"]
	if ageSchema == nil || ageSchema.Type != TypeInteger {
		t.Error("age property not set correctly")
	}
	if ageSchema.Minimum == nil || *ageSchema.Minimum != 0 {
		t.Error("age minimum not set correctly")
	}
	if ageSchema.Maximum == nil || *ageSchema.Maximum != 150 {
		t.Error("age maximum not set correctly")
	}
}

func TestStringConstraints(t *testing.T) {
	s := String().
		WithMinLength(5).
		WithMaxLength(100).
		WithPattern("^[a-z]+$")

	if s.MinLength == nil || *s.MinLength != 5 {
		t.Error("minLength not set correctly")
	}
	if s.MaxLength == nil || *s.MaxLength != 100 {
		t.Error("maxLength not set correctly")
	}
	if s.Pattern != "^[a-z]+$" {
		t.Error("pattern not set correctly")
	}
}

func TestEnum(t *testing.T) {
	s := String().WithEnum("red", "green", "blue")

	if len(s.Enum) != 3 {
		t.Errorf("expected 3 enum values, got %d", len(s.Enum))
	}
}

func TestAdditionalProperties(t *testing.T) {
	s := Object().WithAdditionalProperties(false)

	if s.AdditionalProperties == nil || *s.AdditionalProperties != false {
		t.Error("additionalProperties not set correctly")
	}
}

func TestToJSON(t *testing.T) {
	s := Object().
		WithProperty("name", String()).
		WithRequired("name")

	jsonStr, err := s.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	// Verify it's valid JSON
	var result map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if result["type"] != string(TypeObject) {
		t.Error("type not serialized correctly")
	}
}

func TestToMap(t *testing.T) {
	s := Object().
		WithProperty("count", Integer()).
		WithRequired("count")

	m, err := s.ToMap()
	if err != nil {
		t.Fatalf("ToMap failed: %v", err)
	}

	if m["type"] != string(TypeObject) {
		t.Error("type not in map")
	}

	props, ok := m["properties"].(map[string]any)
	if !ok {
		t.Fatal("properties not in map")
	}

	if props["count"] == nil {
		t.Error("count property not in map")
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		schema  *Schema
		wantErr bool
	}{
		{
			name:    "valid string schema",
			schema:  String(),
			wantErr: false,
		},
		{
			name:    "valid object schema",
			schema:  Object().WithProperty("name", String()),
			wantErr: false,
		},
		{
			name:    "valid array schema",
			schema:  Array(String()),
			wantErr: false,
		},
		{
			name:    "empty schema",
			schema:  &Schema{},
			wantErr: true,
		},
		{
			name:    "object without properties",
			schema:  &Schema{Type: TypeObject},
			wantErr: true,
		},
		{
			name:    "array without items",
			schema:  &Schema{Type: TypeArray},
			wantErr: true,
		},
		{
			name:    "object with invalid nested property",
			schema:  Object().WithProperty("bad", &Schema{}),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.schema.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestComplexNestedSchema(t *testing.T) {
	// Create a complex nested schema: array of objects with nested properties
	addressSchema := Object().
		WithDescription("Physical address").
		WithProperty("street", String()).
		WithProperty("city", String()).
		WithProperty("zipCode", String().WithPattern("^\\d{5}$")).
		WithRequired("street", "city")

	personSchema := Object().
		WithDescription("Person with address").
		WithProperty("name", String()).
		WithProperty("age", Integer().WithMinimum(0)).
		WithProperty("address", addressSchema).
		WithProperty("tags", Array(String())).
		WithRequired("name")

	listSchema := Array(personSchema)

	// Validate
	if err := listSchema.Validate(); err != nil {
		t.Fatalf("complex schema validation failed: %v", err)
	}

	// Convert to map
	m, err := listSchema.ToMap()
	if err != nil {
		t.Fatalf("ToMap failed: %v", err)
	}

	if m["type"] != string(TypeArray) {
		t.Error("root type should be array")
	}

	items, ok := m["items"].(map[string]any)
	if !ok {
		t.Fatal("items should be a map")
	}

	if items["type"] != string(TypeObject) {
		t.Error("items type should be object")
	}
}
