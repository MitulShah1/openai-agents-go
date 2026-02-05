package vectorstore_test

import (
	"context"
	"os"
	"testing"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"

	"github.com/MitulShah1/openai-agents-go/pkg/vectorstore"
)

func TestIntegrationVectorStore(t *testing.T) {
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("skipping integration test: OPENAI_API_KEY not set")
	}

	ctx := context.Background()
	client := openai.NewClient(option.WithAPIKey(os.Getenv("OPENAI_API_KEY")))
	vsClient := vectorstore.New(&client)

	// 1. Create
	name := "test-vector-store"
	vs, err := vsClient.Create(ctx, name)
	if err != nil {
		t.Fatalf("failed to create vector store: %v", err)
	}
	t.Logf("Created vector store: %s", vs.ID)

	// Clean up? Vector stores persist. Usually we rely on external cleanup or test lifecycle.
	// No delete method exposed in client wrapper yet.

	// 2. Add Files (Requires real file IDs using Files API, effectively an integration chain)
	// We skip this part to keep unit test simple, or mock file ID if API allows invalid IDs (it doesn't).
	// So we stop here for basics.
}
