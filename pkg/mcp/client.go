package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// Client represents a client for the Model Context Protocol.
// It connects to an MCP server and discovers tools.
type Client struct {
	serverURL string
	mcpClient *client.Client
}

// NewClient creates a new MCP client.
func NewClient(serverURL string) *Client {
	return &Client{
		serverURL: serverURL,
	}
}

// Connect establishes a connection to the MCP server.
func (c *Client) Connect(ctx context.Context) error {
	// Stub for connection logic.
	return fmt.Errorf("transport initialization not implemented for URL: %s", c.serverURL)
}

// Tool represents an MCP tool description.
type Tool struct {
	Name        string
	Description string
	Schema      map[string]any
}

// ListTools discovers tools from the connected MCP server.
func (c *Client) ListTools(ctx context.Context) ([]Tool, error) {
	if c.mcpClient == nil {
		return nil, fmt.Errorf("client not connected")
	}

	resp, err := c.mcpClient.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return nil, err
	}

	var tools []Tool
	for _, t := range resp.Tools {
		tools = append(tools, Tool{
			Name:        t.Name,
			Description: t.Description,
			Schema:      interfaceToMap(t.InputSchema),
		})
	}
	return tools, nil
}

// CallTool invokes a tool on the MCP server.
func (c *Client) CallTool(ctx context.Context, name string, args map[string]any) (any, error) {
	if c.mcpClient == nil {
		return nil, fmt.Errorf("client not connected")
	}

	resp, err := c.mcpClient.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      name,
			Arguments: args,
		},
	})
	if err != nil {
		return nil, err
	}

	return resp.Content, nil
}

func interfaceToMap(in any) map[string]any {
	if m, ok := in.(map[string]any); ok {
		return m
	}
	return nil
}
