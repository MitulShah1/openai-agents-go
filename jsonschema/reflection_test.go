package jsonschema

import (
	"encoding/json"
	"testing"
)

type TestStruct struct {
	Name        string   `json:"name" jsonschema:"description=The name of the item"`
	Count       int      `json:"count"`
	Tags        []string `json:"tags,omitempty"`
	Category    string   `json:"category" jsonschema:"enum=A|B|C,description=The category"`
	Description string   `json:"desc,omitempty"`
}

func TestGenerateSchema(t *testing.T) {
	schema := GenerateSchema(TestStruct{})

	jsonBytes, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal schema: %v", err)
	}

	t.Logf("Generated Schema:\n%s", string(jsonBytes))

	if schema.Type != TypeObject {
		t.Errorf("Expected type object, got %s", schema.Type)
	}

	if len(schema.Required) < 3 {
		t.Errorf("Expected at least 3 required fields, got %d", len(schema.Required))
	}

	catProp, ok := schema.Properties["category"]
	if !ok {
		t.Error("Missing category property")
	} else if len(catProp.Enum) != 3 {
		t.Errorf("Expected 3 enum values for category, got %d", len(catProp.Enum))
	}
}
