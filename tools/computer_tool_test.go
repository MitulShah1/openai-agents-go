package tools_test

import (
	"testing"

	"github.com/MitulShah1/openai-agents-go/computer"
	"github.com/MitulShah1/openai-agents-go/tools"
)

func TestComputerTool(t *testing.T) {
	mock := &computer.MockComputer{
		ScreenshotFunc: func() (string, error) {
			return "base64data", nil
		},
		ClickFunc: func(_, _ int, _ computer.Button) error {
			return nil
		},
	}

	tool := tools.NewComputerTool(mock)

	// Test screenshot action
	args := `{"action": "screenshot"}`
	res, err := tool.Execute(args, nil)
	if err != nil {
		t.Fatalf("execute screenshot failed: %v", err)
	}

	// Check result map
	m, ok := res.(map[string]string)
	if !ok {
		t.Fatalf("expected map[string]string result, got %T", res)
	}
	if m["source"] != "base64data" {
		t.Errorf("expected base64data, got %s", m["source"])
	}

	// Test click action
	args = `{"action": "click", "coordinates": [100, 200]}`
	_, err = tool.Execute(args, nil)
	if err != nil {
		t.Fatalf("execute click failed: %v", err)
	}
}
