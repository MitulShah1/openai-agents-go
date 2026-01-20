package function

import (
	"encoding/json"
	"testing"
)

type WeatherArgs struct {
	City string `json:"city" jsonschema:"description=The city name"`
	Unit string `json:"unit,omitempty" jsonschema:"enum=celsius|fahrenheit"`
}

func TestFromFunc(t *testing.T) {
	tool := FromFunc("get_weather", "Get weather", func(args WeatherArgs) (any, error) {
		if args.City == "" {
			return nil, nil // Error?
		}
		return "Sunny in " + args.City, nil
	})

	if tool.Name != "get_weather" {
		t.Errorf("Expected name get_weather, got %s", tool.Name)
	}

	// Check schema
	params, _ := json.Marshal(tool.Parameters)
	t.Logf("Schema: %s", string(params))

	// Execute with valid args
	args := map[string]any{
		"city": "London",
	}

	res, err := tool.Callback(args, nil)
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	if res != "Sunny in London" {
		t.Errorf("Expected Sunny in London, got %v", res)
	}
}
