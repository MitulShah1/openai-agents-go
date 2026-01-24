package embeddings_test

import (
	"context"
	"os"
	"testing"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"

	"github.com/MitulShah1/openai-agents-go/pkg/embeddings"
)

func TestIntegrationEmbeddings(t *testing.T) {
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("skipping integration test: OPENAI_API_KEY not set")
	}

	ctx := context.Background()
	client := openai.NewClient(option.WithAPIKey(os.Getenv("OPENAI_API_KEY")))
	embClient := embeddings.New(&client)

	texts := []string{"Hello, world!", "This is a test."}
	vectors, err := embClient.Generate(ctx, texts, "")
	if err != nil {
		t.Fatalf("failed to generate embeddings: %v", err)
	}

	if len(vectors) != 2 {
		t.Errorf("expected 2 vectors, got %d", len(vectors))
	}
	if len(vectors[0]) != 1536 { // text-embedding-3-small is 1536 dims
		t.Errorf("expected 1536 dimensions, got %d", len(vectors[0]))
	}
}
