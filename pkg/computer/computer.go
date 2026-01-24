package computer

import (
	"context"
)

// Button represents a mouse button.
type Button string

const (
	Left    Button = "left"
	Right   Button = "right"
	Middle  Button = "middle"
	Wheel   Button = "wheel" // Usually implies scroll but maybe button click?
	Back    Button = "back"
	Forward Button = "forward"
)

// Environment represents the operating environment.
type Environment string

const (
	EnvMac     Environment = "mac"
	EnvWindows Environment = "windows"
	EnvUbuntu  Environment = "ubuntu"
	EnvBrowser Environment = "browser"
)

// Point represents a 2D coordinate.
type Point struct {
	X int
	Y int
}

// Computer is the synchronous interface for computer interaction.
type Computer interface {
	// Screenshot captures the screen or viewport and returns a base64 encoded string.
	Screenshot() (string, error)

	// Click performs a mouse click at the given coordinates with the specified button.
	Click(x, y int, button Button) error

	// DoubleClick performs a double click at the given coordinates.
	DoubleClick(x, y int) error

	// Scroll performs a scroll action.
	Scroll(x, y, scrollX, scrollY int) error

	// Type types the given text.
	Type(text string) error

	// Wait waits for UI updates (e.g. implicitly or explicit sleep).
	Wait() error

	// Move moves the mouse to the given coordinates.
	Move(x, y int) error

	// Keypress presses the given keys (e.g. ["Ctrl", "C"]).
	Keypress(keys []string) error

	// Drag drags from start to a sequence of points.
	// Implementation might assume starting drag from current position or explicit start.
	Drag(path []Point) error
}

// AsyncComputer is the asynchronous interface with context support.
type AsyncComputer interface {
	Screenshot(ctx context.Context) (string, error)
	Click(ctx context.Context, x, y int, button Button) error
	DoubleClick(ctx context.Context, x, y int) error
	Scroll(ctx context.Context, x, y, scrollX, scrollY int) error
	Type(ctx context.Context, text string) error
	Wait(ctx context.Context) error
	Move(ctx context.Context, x, y int) error
	Keypress(ctx context.Context, keys []string) error
	Drag(ctx context.Context, path []Point) error
}
