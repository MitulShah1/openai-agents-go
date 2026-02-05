package mcp

import "context"

// Server represents an MCP server that the agent can connect to.
type Server struct {
	URL string
	// connection management here
}

// NewServer creates a new MCP server definition.
func NewServer(url string) *Server {
	return &Server{URL: url}
}

// Connect starts the server (if local) or connects to it (remote).
// For now, assumes remote.
func (s *Server) Connect(ctx context.Context) (*Client, error) {
	client := NewClient(s.URL)
	if err := client.Connect(ctx); err != nil {
		return nil, err
	}
	return client, nil
}
