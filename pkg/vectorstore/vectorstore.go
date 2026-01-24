package vectorstore

import (
	"context"
	"fmt"

	"github.com/openai/openai-go/v3"
)

// Client wraps the OpenAI Vector Stores API.
type Client struct {
	client *openai.Client
}

// New creates a new VectorStores client.
func New(client *openai.Client) *Client {
	return &Client{
		client: client,
	}
}

// Store represents a vector store.
type Store = openai.VectorStore

// Create creates a new vector store.
func (c *Client) Create(ctx context.Context, name string) (*Store, error) {
	params := openai.VectorStoreNewParams{
		Name: openai.String(name),
	}
	vs, err := c.client.VectorStores.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create vector store: %w", err)
	}
	return vs, nil
}

// AddFiles adds files to a vector store.
// It uses CreateBatch but waits for completion logic is not implemented here fully (async).
// For simplicity, we trigger the batch.
func (c *Client) AddFiles(ctx context.Context, vectorStoreID string, fileIDs []string) error {
	params := openai.VectorStoreFileBatchNewParams{
		FileIDs: fileIDs,
	}

	_, err := c.client.VectorStores.FileBatches.New(ctx, vectorStoreID, params)
	if err != nil {
		return fmt.Errorf("failed to create file batch: %w", err)
	}

	// Normally we might want to poll for status, but this is an async operation.
	return nil
}

// Search performs a semantic search.
// Note: The OpenAI API does not expose a direct search endpoint for Vector Stores.
// Search is typically performed by an Assistant with the file_search tool enabled,
// attached to this vector store.
// This method is a placeholder or would require creating a temporary thread/run.
func (c *Client) Search(ctx context.Context, vectorStoreID string, query string) (string, error) {
	return "", fmt.Errorf("not implemented: vector store search requires assistant integration")
}
