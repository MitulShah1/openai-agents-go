package files

import (
	"context"
	"fmt"
	"io"

	"github.com/openai/openai-go/v3"
)

// Client wraps the OpenAI Files API.
type Client struct {
	client *openai.Client
}

// New creates a new Files client.
func New(client *openai.Client) *Client {
	return &Client{
		client: client,
	}
}

// FileObject represents a file object from the OpenAI API.
// We expect the type to be exposed, usually File or FileObject.
// Since openai.File is a function, the struct might be FileObject.
// Let's rely on type inference or use "any" temporarily if unsure, but FileObject is common.
type FileObject = openai.FileObject

// Upload uploads a file to OpenAI.
// purpose is usually "assistants" or "fine-tune".
func (c *Client) Upload(ctx context.Context, file io.Reader, filename string, purpose string) (*FileObject, error) {
	if purpose == "" {
		purpose = "assistants"
	}

	// openai.File is the function to create the file param
	// Note: The error said openai.File is func(io.Reader, string, string) ...
	fileParam := openai.File(file, filename, "application/octet-stream")

	params := openai.FileNewParams{
		File:    fileParam,
		Purpose: openai.FilePurpose(purpose),
	}

	f, err := c.client.Files.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	return f, nil
}

// Get retrieves metadata for a file.
func (c *Client) Get(ctx context.Context, fileID string) (*FileObject, error) {
	f, err := c.client.Files.Get(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get file %s: %w", fileID, err)
	}
	return f, nil
}

// List lists all uploaded files.
func (c *Client) List(ctx context.Context) ([]FileObject, error) {
	iter := c.client.Files.ListAutoPaging(ctx, openai.FileListParams{})
	var files []FileObject
	for iter.Next() {
		files = append(files, iter.Current())
	}
	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}
	return files, nil
}

// Delete deletes a file.
func (c *Client) Delete(ctx context.Context, fileID string) error {
	_, err := c.client.Files.Delete(ctx, fileID)
	if err != nil {
		return fmt.Errorf("failed to delete file %s: %w", fileID, err)
	}
	return nil
}

// Content retrieves the content of a file.
// Returns a ReadCloser that must be closed by the caller.
func (c *Client) Content(ctx context.Context, fileID string) (io.ReadCloser, error) {
	resp, err := c.client.Files.Content(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get file content: %w", err)
	}
	return resp.Body, nil
}
