package tools

import "fmt"

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
