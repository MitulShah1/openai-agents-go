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

// Screenshot returns a mock base64 image or the result of ScreenshotFunc.
func (m *MockComputer) Screenshot() (string, error) {
	if m.ScreenshotFunc != nil {
		return m.ScreenshotFunc()
	}
	return "mock_base64_image", nil
}

// Click delegates to ClickFunc or returns nil.
func (m *MockComputer) Click(x, y int, button Button) error {
	if m.ClickFunc != nil {
		return m.ClickFunc(x, y, button)
	}
	return nil
}

// DoubleClick is a no-op for the mock.
func (m *MockComputer) DoubleClick(_, _ int) error {
	return nil
}

// Scroll is a no-op for the mock.
func (m *MockComputer) Scroll(_, _ int, _, _ int) error {
	return nil
}

// Type delegates to TypeFunc or returns nil.
func (m *MockComputer) Type(text string) error {
	if m.TypeFunc != nil {
		return m.TypeFunc(text)
	}
	return nil
}

// Wait is a no-op for the mock.
func (m *MockComputer) Wait() error {
	return nil
}

// Move delegates to MoveFunc or returns nil.
func (m *MockComputer) Move(x, y int) error {
	if m.MoveFunc != nil {
		return m.MoveFunc(x, y)
	}
	return nil
}

// Keypress is a no-op for the mock.
func (m *MockComputer) Keypress(_ []string) error {
	return nil
}

// Drag is a no-op for the mock.
func (m *MockComputer) Drag(_ []Point) error {
	return nil
}

// AsyncMockComputer implements AsyncComputer for testing.
type AsyncMockComputer struct {
	MockComputer
}

// Screenshot delegates to MockComputer.Screenshot.
func (m *AsyncMockComputer) Screenshot(_ context.Context) (string, error) {
	return m.MockComputer.Screenshot()
}

// Click delegates to MockComputer.Click.
func (m *AsyncMockComputer) Click(_ context.Context, x, y int, button Button) error {
	return m.MockComputer.Click(x, y, button)
}

// DoubleClick delegates to MockComputer.DoubleClick.
func (m *AsyncMockComputer) DoubleClick(_ context.Context, x, y int) error {
	return m.MockComputer.DoubleClick(x, y)
}

// Scroll delegates to MockComputer.Scroll.
func (m *AsyncMockComputer) Scroll(_ context.Context, x, y, scrollX, scrollY int) error {
	return m.MockComputer.Scroll(x, y, scrollX, scrollY)
}

// Type delegates to MockComputer.Type.
func (m *AsyncMockComputer) Type(_ context.Context, text string) error {
	return m.MockComputer.Type(text)
}

// Wait delegates to MockComputer.Wait.
func (m *AsyncMockComputer) Wait(_ context.Context) error {
	return m.MockComputer.Wait()
}

// Move delegates to MockComputer.Move.
func (m *AsyncMockComputer) Move(_ context.Context, x, y int) error {
	return m.MockComputer.Move(x, y)
}

// Keypress delegates to MockComputer.Keypress.
func (m *AsyncMockComputer) Keypress(_ context.Context, keys []string) error {
	return m.MockComputer.Keypress(keys)
}

// Drag delegates to MockComputer.Drag.
func (m *AsyncMockComputer) Drag(_ context.Context, path []Point) error {
	return m.MockComputer.Drag(path)
}
