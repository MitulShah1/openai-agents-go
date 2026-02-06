package computer_test

import (
	"context"
	"testing"

	"github.com/MitulShah1/openai-agents-go/computer"
)

func TestMockComputer(t *testing.T) {
	mock := &computer.MockComputer{
		ScreenshotFunc: func() (string, error) {
			return "img123", nil
		},
		ClickFunc: func(x, y int, button computer.Button) error {
			if x != 10 || y != 20 {
				t.Errorf("unexpected coords: %d, %d", x, y)
			}
			if button != computer.Left {
				t.Errorf("unexpected button: %s", button)
			}
			return nil
		},
	}

	snap, err := mock.Screenshot()
	if err != nil {
		t.Fatalf("screenshot failed: %v", err)
	}
	if snap != "img123" {
		t.Errorf("unexpected screenshot: %s", snap)
	}

	if err := mock.Click(10, 20, computer.Left); err != nil {
		t.Fatalf("click failed: %v", err)
	}
}

func TestAsyncMockComputer(t *testing.T) {
	mock := &computer.AsyncMockComputer{}
	// Just verify method existence and compilation
	if err := mock.Wait(context.Background()); err != nil {
		t.Fatalf("wait failed: %v", err)
	}
}
