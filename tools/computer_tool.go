package tools

import (
	"fmt"

	"github.com/MitulShah1/openai-agents-go/pkg/computer"
)

// NewComputerTool creates a tool that allows agents to interact with a computer.
// It exposes actions like screenshot, click, type, move, etc.
func NewComputerTool(comp computer.Computer) Tool {
	return FromFunc(
		"computer",
		"Interact with the computer (screen, mouse, keyboard).",
		func(args struct {
			Action      string `json:"action" jsonschema:"description=The action to perform: screenshot, click, type, key, mouse_move, left_click, right_click, double_click, scroll"`
			Coordinates []int  `json:"coordinates,omitempty" jsonschema:"description=The (x, y) coordinates for mouse actions"`
			Text        string `json:"text,omitempty" jsonschema:"description=The text to type or keys to press"`
		}) (any, error) {
			switch args.Action {
			case "screenshot":
				b64, err := comp.Screenshot()
				if err != nil {
					return nil, err
				}
				// Return truncated or full? Usually full Base64 for the model to see.
				// But we return output. The model receives it as tool output.
				// If it's an image, the model needs to handle it.
				// For now return text saying "screenshot_base64: ..."
				// But the model can ingest images if using Vision.
				// This tool returns "image_url" content block?
				// For now, return the raw base64 string or an object.
				return map[string]string{"type": "image", "source": b64}, nil

			case "click", "left_click":
				if len(args.Coordinates) != 2 {
					return nil, fmt.Errorf("coordinates required for click")
				}
				return nil, comp.Click(args.Coordinates[0], args.Coordinates[1], computer.Left)

			case "right_click":
				if len(args.Coordinates) != 2 {
					return nil, fmt.Errorf("coordinates required for right_click")
				}
				return nil, comp.Click(args.Coordinates[0], args.Coordinates[1], computer.Right)

			case "double_click":
				if len(args.Coordinates) != 2 {
					return nil, fmt.Errorf("coordinates required for double_click")
				}
				return nil, comp.DoubleClick(args.Coordinates[0], args.Coordinates[1])

			case "mouse_move":
				if len(args.Coordinates) != 2 {
					return nil, fmt.Errorf("coordinates required for mouse_move")
				}
				return nil, comp.Move(args.Coordinates[0], args.Coordinates[1])

			case "type":
				return nil, comp.Type(args.Text)

			case "key":
				// assuming args.Text contains comma separated keys or just one key?
				// Simple implementation: single key sequence?
				// Or parse args.Text
				return nil, comp.Keypress([]string{args.Text})

			case "scroll":
				// Needs scroll amount. Arg schema might need enhancement or hijack "coordinates"?
				// Assume coordinates [dx, dy]?
				if len(args.Coordinates) != 2 {
					return nil, fmt.Errorf("coordinates [scrollX, scrollY] required for scroll")
				}
				// We don't have current mouse pos in args, but Scroll API takes x, y.
				// We assume 0,0 relative? Or we need absolute pos.
				// For simplicity, we assume generic scroll.
				return nil, comp.Scroll(0, 0, args.Coordinates[0], args.Coordinates[1])

			default:
				return nil, fmt.Errorf("unknown action: %s", args.Action)
			}
		},
	)
}
