package computer

import "context"

// MockComputer implements Computer for testing.
type MockComputer struct {
	ScreenshotFunc func() (string, error)
	ClickFunc      func(x, y int, button Button) error
	TypeFunc       func(text string) error
	MoveFunc       func(x, y int) error
	// ... add others as needed
}

func (m *MockComputer) Screenshot() (string, error) {
	if m.ScreenshotFunc != nil {
		return m.ScreenshotFunc()
	}
	return "mock_base64_image", nil
}

func (m *MockComputer) Click(x, y int, button Button) error {
	if m.ClickFunc != nil {
		return m.ClickFunc(x, y, button)
	}
	return nil
}

func (m *MockComputer) DoubleClick(x, y int) error {
	return nil
}

func (m *MockComputer) Scroll(x, y, scrollX, scrollY int) error {
	return nil
}

func (m *MockComputer) Type(text string) error {
	if m.TypeFunc != nil {
		return m.TypeFunc(text)
	}
	return nil
}

func (m *MockComputer) Wait() error {
	return nil
}

func (m *MockComputer) Move(x, y int) error {
	if m.MoveFunc != nil {
		return m.MoveFunc(x, y)
	}
	return nil
}

func (m *MockComputer) Keypress(keys []string) error {
	return nil
}

func (m *MockComputer) Drag(path []Point) error {
	return nil
}

// AsyncMockComputer implements AsyncComputer for testing.
type AsyncMockComputer struct {
	MockComputer
}

func (m *AsyncMockComputer) Screenshot(ctx context.Context) (string, error) {
	return m.MockComputer.Screenshot()
}

func (m *AsyncMockComputer) Click(ctx context.Context, x, y int, button Button) error {
	return m.MockComputer.Click(x, y, button)
}

func (m *AsyncMockComputer) DoubleClick(ctx context.Context, x, y int) error {
	return m.MockComputer.DoubleClick(x, y)
}

func (m *AsyncMockComputer) Scroll(ctx context.Context, x, y, scrollX, scrollY int) error {
	return m.MockComputer.Scroll(x, y, scrollX, scrollY)
}

func (m *AsyncMockComputer) Type(ctx context.Context, text string) error {
	return m.MockComputer.Type(text)
}

func (m *AsyncMockComputer) Wait(ctx context.Context) error {
	return m.MockComputer.Wait()
}

func (m *AsyncMockComputer) Move(ctx context.Context, x, y int) error {
	return m.MockComputer.Move(x, y)
}

func (m *AsyncMockComputer) Keypress(ctx context.Context, keys []string) error {
	return m.MockComputer.Keypress(keys)
}

func (m *AsyncMockComputer) Drag(ctx context.Context, path []Point) error {
	return m.MockComputer.Drag(path)
}
