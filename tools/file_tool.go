package tools

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/MitulShah1/openai-agents-go/pkg/files"
)

// NewUploadFileTool creates a tool that allows agents to upload files.
func NewUploadFileTool(client *files.Client) Tool {
	return FromFunc(
		"upload_file",
		"Upload a file to OpenAI. Returns the file object.",
		func(args struct {
			Filename string `json:"filename" jsonschema:"description=Name of the file to upload"`
			Content  string `json:"content" jsonschema:"description=Text content of the file"`
			Purpose  string `json:"purpose,omitempty" jsonschema:"description=Purpose of the file (assistants, fine-tune), defaults to assistants"`
		}) (any, error) {
			ctx := context.Background() // Tools don't strictly receive context in FromFunc adapter yet, assuming background or TODO: fix FromFunc to pass context?
			// Actually tool.Callback receives contextVariables, but not context.Context directly.
			// But FromFunc doesn't expose context.Context.
			// We should probably rely on the runner context, but since FromFunc interface is rigid, we use Background for now.
			// Ideally tools.FromFunc should support context if the inner func asks for it, but for now:

			return client.Upload(ctx, strings.NewReader(args.Content), args.Filename, args.Purpose)
		},
	)
}

// NewListFilesTool creates a tool that lists uploaded files.
func NewListFilesTool(client *files.Client) Tool {
	return FromFunc(
		"list_files",
		"List all files uploaded to OpenAI.",
		func(_ struct{}) (any, error) {
			ctx := context.Background()
			return client.List(ctx)
		},
	)
}

// NewGetFileTool creates a tool that retrieves file metadata.
func NewGetFileTool(client *files.Client) Tool {
	return FromFunc(
		"get_file",
		"Get metadata for a specific file.",
		func(args struct {
			FileID string `json:"file_id" jsonschema:"description=ID of the file to retrieve"`
		}) (any, error) {
			ctx := context.Background()
			return client.Get(ctx, args.FileID)
		},
	)
}

// NewGetFileContentTool creates a tool that retrieves file content.
func NewGetFileContentTool(client *files.Client) Tool {
	return FromFunc(
		"get_file_content",
		"Get the content of a specific file.",
		func(args struct {
			FileID string `json:"file_id" jsonschema:"description=ID of the file to retrieve content from"`
		}) (any, error) {
			ctx := context.Background()
			rc, err := client.Content(ctx, args.FileID)
			if err != nil {
				return nil, err
			}
			defer func() { _ = rc.Close() }()

			content, err := io.ReadAll(rc)
			if err != nil {
				return nil, fmt.Errorf("failed to read file content: %w", err)
			}
			return string(content), nil
		},
	)
}

// NewDeleteFileTool creates a tool that deletes a file.
func NewDeleteFileTool(client *files.Client) Tool {
	return FromFunc(
		"delete_file",
		"Delete a file from OpenAI.",
		func(args struct {
			FileID string `json:"file_id" jsonschema:"description=ID of the file to delete"`
		}) (any, error) {
			ctx := context.Background()
			err := client.Delete(ctx, args.FileID)
			if err != nil {
				return nil, err
			}
			return map[string]string{"status": "deleted", "id": args.FileID}, nil
		},
	)
}
