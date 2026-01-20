package agents

import (
	"testing"
)

func TestTextContent(t *testing.T) {
	text := "Hello, world!"
	content := TextContent(text)

	if content.Type != ContentTypeText {
		t.Errorf("Expected type %s, got %s", ContentTypeText, content.Type)
	}
	if content.Text != text {
		t.Errorf("Expected text %q, got %q", text, content.Text)
	}
	if content.String() != text {
		t.Errorf("Expected String() to return %q, got %q", text, content.String())
	}
}

func TestImageContent(t *testing.T) {
	url := "https://example.com/image.png"
	detail := "high"
	content := ImageContent(url, detail)

	if content.Type != ContentTypeImage {
		t.Errorf("Expected type %s, got %s", ContentTypeImage, content.Type)
	}
	if content.ImageURL != url {
		t.Errorf("Expected URL %q, got %q", url, content.ImageURL)
	}
	if content.ImageDetail != detail {
		t.Errorf("Expected detail %q, got %q", detail, content.ImageDetail)
	}

	// Test default detail
	contentAuto := ImageContent(url, "")
	if contentAuto.ImageDetail != "auto" {
		t.Errorf("Expected default detail 'auto', got %q", contentAuto.ImageDetail)
	}
}

func TestFileContent(t *testing.T) {
	data := []byte("file content here")
	filename := "test.txt"
	mimeType := "text/plain"
	content := FileContent(data, filename, mimeType)

	if content.Type != ContentTypeFile {
		t.Errorf("Expected type %s, got %s", ContentTypeFile, content.Type)
	}
	if content.FileName != filename {
		t.Errorf("Expected filename %q, got %q", filename, content.FileName)
	}
	if content.MimeType != mimeType {
		t.Errorf("Expected mime type %q, got %q", mimeType, content.MimeType)
	}
	if len(content.FileData) != len(data) {
		t.Errorf("Expected data length %d, got %d", len(data), len(content.FileData))
	}
}

func TestIsContent(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected bool
	}{
		{
			name:     "Content value",
			input:    TextContent("test"),
			expected: true,
		},
		{
			name:     "Content pointer",
			input:    &Content{Type: ContentTypeText, Text: "test"},
			expected: true,
		},
		{
			name:     "String",
			input:    "test",
			expected: false,
		},
		{
			name:     "Integer",
			input:    42,
			expected: false,
		},
		{
			name:     "Nil",
			input:    nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, ok := IsContent(tt.input)
			if ok != tt.expected {
				t.Errorf("Expected IsContent to return %v, got %v", tt.expected, ok)
			}
			if ok && content == nil {
				t.Error("Expected non-nil content pointer when ok is true")
			}
			if !ok && content != nil {
				t.Error("Expected nil content pointer when ok is false")
			}
		})
	}
}

func TestContentString(t *testing.T) {
	tests := []struct {
		name     string
		content  Content
		contains string // substring that should be in the result
	}{
		{
			name:     "Text content",
			content:  TextContent("Hello, world!"),
			contains: "Hello, world!",
		},
		{
			name:     "Image content",
			content:  ImageContent("https://example.com/img.png", "high"),
			contains: "Image:",
		},
		{
			name:     "File content",
			content:  FileContent([]byte("data"), "test.txt", "text/plain"),
			contains: "File:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.content.String()
			if result == "" {
				t.Error("Expected non-empty string result")
			}
			// For text, it should be exact match
			if tt.content.Type == ContentTypeText && result != tt.contains {
				t.Errorf("Expected exact match %q, got %q", tt.contains, result)
			}
			// For image/file, just check it contains expected substring
			if tt.content.Type != ContentTypeText && !contains(result, tt.contains) {
				t.Errorf("Expected result to contain %q, got %q", tt.contains, result)
			}
		})
	}
}

func TestToolWithMultimodalCallback(t *testing.T) {
	// Test that tools can return Content objects
	tool := FunctionTool(
		"image_generator",
		"Generates an image",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"prompt": map[string]any{"type": "string"},
			},
		},
		func(_ map[string]any, _ ContextVariables) (any, error) {
			return ImageContent("https://example.com/generated.png", "high"), nil
		},
	)

	result, err := tool.Execute(`{"prompt":"sunset"}`, ContextVariables{})
	if err != nil {
		t.Fatalf("Tool execution failed: %v", err)
	}

	content, ok := IsContent(result)
	if !ok {
		t.Fatal("Expected result to be Content")
	}
	if content.Type != ContentTypeImage {
		t.Errorf("Expected image content, got %s", content.Type)
	}
}

func TestToolBackwardCompatibility(t *testing.T) {
	// Test that tools can still return strings (backward compatibility)
	tool := FunctionTool(
		"old_tool",
		"Legacy tool returning string",
		map[string]any{"type": "object"},
		func(_ map[string]any, _ ContextVariables) (any, error) {
			return "plain string result", nil
		},
	)

	result, err := tool.Execute("{}", ContextVariables{})
	if err != nil {
		t.Fatalf("Tool execution failed: %v", err)
	}

	// Should still work with plain strings
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	str, ok := result.(string)
	if !ok {
		t.Fatalf("Expected string result, got %T", result)
	}
	if str != "plain string result" {
		t.Errorf("Expected %q, got %q", "plain string result", str)
	}
}

// Helper function to check substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
