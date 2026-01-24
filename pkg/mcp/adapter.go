package mcp

import (
	"context"

	"github.com/MitulShah1/openai-agents-go/tools"
)

// ToAgentTool converts an MCP Tool to an SDK Tool.
func (t *Tool) ToAgentTool(client *Client) tools.Tool {
	return tools.New(
		t.Name,
		t.Description,
		t.Schema,
		func(args map[string]any, _ tools.ContextVariables) (any, error) {
			// Adapter uses background context as Tool callback doesn't provide it yet
			return client.CallTool(context.Background(), t.Name, args)
		},
	)
}
