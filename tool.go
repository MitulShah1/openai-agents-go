package agents

import (
	"encoding/json"
	"fmt"

	"github.com/openai/openai-go/v3"
)

// Tool represents a function that can be called by an agent.
type Tool struct {
	// Name is the name of the tool.
	Name string
	// Description is the description of the tool.
	Description string
	// Parameters is the JSON schema for the tool parameters.
	Parameters map[string]any
	// Callback is the function to execute when the tool is called.
	// It receives the arguments as a map and context variables.
	Callback func(args map[string]any, ctx ContextVariables) (any, error)

	// IsHandoffTool indicates this tool performs an agent handoff.
	// When true, the Callback is expected to return *Agent.
	// This field is set automatically by the handoff package.
	IsHandoffTool bool
}

// ContentType represents the type of content in a tool response.
type ContentType string

const (
	// ContentTypeText represents plain text content.
	ContentTypeText ContentType = "text"
	// ContentTypeImage represents image content.
	ContentTypeImage ContentType = "image"
	// ContentTypeFile represents file/document content.
	ContentTypeFile ContentType = "file"
)

// Content represents multimodal content that can be returned by tools.
// It supports text, images, and files to enable rich tool outputs.
type Content struct {
	// Type indicates the content type (text, image, file).
	Type ContentType
	// Text contains the text content (for ContentTypeText).
	Text string
	// ImageURL contains the image URL (for ContentTypeImage).
	ImageURL string
	// ImageDetail specifies the detail level for image processing ("low", "high", "auto").
	ImageDetail string
	// FileData contains the raw file bytes (for ContentTypeFile).
	FileData []byte
	// FileName is the name of the file (for ContentTypeFile).
	FileName string
	// MimeType is the MIME type of the file (for ContentTypeFile).
	MimeType string
}

// TextContent creates a text Content object.
// This is the most common content type for tool responses.
func TextContent(text string) Content {
	return Content{
		Type: ContentTypeText,
		Text: text,
	}
}

// ImageContent creates an image Content object.
// The imageURL should be a publicly accessible URL or data URI.
// The detail parameter controls image processing quality: "low", "high", or "auto".
func ImageContent(imageURL string, detail string) Content {
	if detail == "" {
		detail = "auto"
	}
	return Content{
		Type:        ContentTypeImage,
		ImageURL:    imageURL,
		ImageDetail: detail,
	}
}

// FileContent creates a file Content object.
// The data parameter contains the raw file bytes.
// The filename and mimeType help identify and process the file.
func FileContent(data []byte, filename string, mimeType string) Content {
	return Content{
		Type:     ContentTypeFile,
		FileData: data,
		FileName: filename,
		MimeType: mimeType,
	}
}

// ToParam converts the Tool to an openai.ChatCompletionToolUnionParam.
func (t Tool) ToParam() openai.ChatCompletionToolUnionParam {
	// If parameters are empty, default to empty object
	params := t.Parameters
	if params == nil {
		params = map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		}
	}

	// Use v3 API helper function
	return openai.ChatCompletionFunctionTool(openai.FunctionDefinitionParam{
		Name:        t.Name,
		Description: openai.String(t.Description),
		Parameters:  openai.FunctionParameters(params),
	})
}

// Execute runs the tool's callback with the provided arguments.
func (t Tool) Execute(argsJSON string, ctx ContextVariables) (any, error) {
	// Handle empty args - default to empty JSON object
	if argsJSON == "" {
		argsJSON = "{}"
	}

	var args map[string]any
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return nil, fmt.Errorf("failed to unmarshal arguments: %w", err)
	}

	// Validate callback exists
	if t.Callback == nil {
		return nil, fmt.Errorf("tool %s has no callback function", t.Name)
	}

	return t.Callback(args, ctx)
}

// FunctionTool is a helper to create a Tool from a simpler definition.
// For now, it accepts manual schema. In the future, we could use reflection.
func FunctionTool(name, description string, params map[string]any, callback func(map[string]any, ContextVariables) (any, error)) Tool {
	if name == "" {
		panic("tool name cannot be empty")
	}
	if callback == nil {
		panic("tool callback cannot be nil")
	}

	return Tool{
		Name:        name,
		Description: description,
		Parameters:  params,
		Callback:    callback,
	}
}

// IsHandoff checks if the result is an Agent, indicating a handoff.
func IsHandoff(result any) (*Agent, bool) {
	a, ok := result.(*Agent)
	return a, ok
}

// IsContent checks if the result is a Content object.
func IsContent(result any) (*Content, bool) {
	switch v := result.(type) {
	case Content:
		return &v, true
	case *Content:
		return v, true
	default:
		return nil, false
	}
}

// String returns a string representation of the Content.
// For text content, it returns the text directly.
// For image/file content, it returns a description.
func (c Content) String() string {
	switch c.Type {
	case ContentTypeText:
		return c.Text
	case ContentTypeImage:
		return fmt.Sprintf("[Image: %s (detail: %s)]", c.ImageURL, c.ImageDetail)
	case ContentTypeFile:
		return fmt.Sprintf("[File: %s (%s, %d bytes)]", c.FileName, c.MimeType, len(c.FileData))
	default:
		return ""
	}
}
