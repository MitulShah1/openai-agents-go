package tools

import (
	"testing"
)

type weatherArgs struct {
	City string `json:"city"`
	Unit string `json:"unit,omitempty"`
}

func TestFromFunc(t *testing.T) {
	// Define a function
	getWeather := func(args weatherArgs) (any, error) {
		return "Sunny in " + args.City, nil
	}

	// Create tool
	tool := FromFunc("get_weather", "Get weather", getWeather)

	// Verify name and description
	if tool.Name != "get_weather" {
		t.Errorf("expected name 'get_weather', got %q", tool.Name)
	}
	if tool.Description != "Get weather" {
		t.Errorf("expected description 'Get weather', got %q", tool.Description)
	}

	// Verify params generated
	if tool.Parameters == nil {
		t.Fatal("expected parameters to be generated")
	}

	// Test Execution
	args := `{"city": "Paris"}`
	result, err := tool.Execute(args, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != "Sunny in Paris" {
		t.Errorf("expected 'Sunny in Paris', got %v", result)
	}

	// Test generic generic unmarshalling
	// Verify that optional fields work
	argsFull := `{"city": "London", "unit": "celsius"}`
	_, err = tool.Execute(argsFull, nil)
	if err != nil {
		t.Errorf("unexpected error with full args: %v", err)
	}
}
