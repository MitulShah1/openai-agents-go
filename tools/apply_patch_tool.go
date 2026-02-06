package tools

import (
	"fmt"

	"github.com/MitulShah1/openai-agents-go/diff"
)

// NewApplyPatchTool creates a tool that applies a patch to a file.
// Note: This implementation assumes the agent manages reading/writing the file separately (e.g. read, then apply, then write).
// Or we could pass a FileSystem interface?
// For simplicity, we just take the content and diff, and return modified content.
// The agent is responsible for saving it.
func NewApplyPatchTool() Tool {
	return FromFunc(
		"apply_patch",
		"Apply a V4A-formatted patch to the given file content.",
		func(args struct {
			Content string `json:"content" jsonschema:"description=The original content of the file"`
			Patch   string `json:"patch" jsonschema:"description=The V4A format diff to apply"`
		}) (any, error) {
			// Parse the patch
			diffs, err := diff.Parse(args.Patch)
			if err != nil {
				return nil, fmt.Errorf("failed to parse patch: %w", err)
			}
			if len(diffs) == 0 {
				return args.Content, nil // No changes
			}

			// Apply the first diff (assuming patch is for this content)
			// If multiple files update in patch, we only process the first one that matches?
			// Usually agent sends patch for one file.
			result, err := diff.Apply(args.Content, diffs[0])
			if err != nil {
				return nil, fmt.Errorf("failed to apply patch: %w", err)
			}

			return result, nil
		},
	)
}
