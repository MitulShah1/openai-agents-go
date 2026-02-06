package mcp

import (
	"context"
	"testing"
)

func TestClientStub(t *testing.T) {
	client := NewClient("http://localhost:8080")
	if client == nil {
		t.Fatal("NewClient returned nil")
	}

	// Connect should return error (not implemented transport)
	err := client.Connect(context.Background())
	if err == nil {
		t.Error("Connect expected error, got nil")
	}
}
