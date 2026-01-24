package embeddings

import (
	"context"
	"fmt"

	"github.com/openai/openai-go/v3"
)

// Client wraps the OpenAI Embeddings API.
type Client struct {
	client *openai.Client
}

// New creates a new Embeddings client.
func New(client *openai.Client) *Client {
	return &Client{
		client: client,
	}
}

// DefaultModel is the default model used for embeddings.
const DefaultModel = "text-embedding-3-small"

// Generate creates embeddings for the given texts.
// If model is empty, uses DefaultModel.
func (c *Client) Generate(ctx context.Context, texts []string, model string) ([][]float64, error) {
	if model == "" {
		model = DefaultModel
	}

	params := openai.EmbeddingNewParams{
		Input: openai.EmbeddingNewParamsInputUnion{
			OfArrayOfStrings: texts,
		},
		Model: openai.EmbeddingModel(model),
	}

	resp, err := c.client.Embeddings.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embeddings: %w", err)
	}

	var results [][]float64
	for _, data := range resp.Data {
		results = append(results, data.Embedding)
	}

	return results, nil
}
