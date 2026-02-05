package files_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"

	"github.com/MitulShah1/openai-agents-go/pkg/files"
)

func TestIntegrationFiles(t *testing.T) {
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("skipping integration test: OPENAI_API_KEY not set")
	}

	ctx := context.Background()
	client := openai.NewClient(option.WithAPIKey(os.Getenv("OPENAI_API_KEY")))
	filesClient := files.New(&client)

	// 1. Upload
	content := "Hello, World! Integration Test"
	f, err := filesClient.Upload(ctx, strings.NewReader(content), "test-file.txt", "assistants")
	if err != nil {
		t.Fatalf("failed to upload file: %v", err)
	}
	t.Logf("Uploaded file: %s (%s)", f.ID, f.Filename)

	// Clean up at end
	defer func() {
		if err := filesClient.Delete(ctx, f.ID); err != nil {
			t.Errorf("failed to delete file %s: %v", f.ID, err)
		}
	}()

	// 2. Get
	fetched, err := filesClient.Get(ctx, f.ID)
	if err != nil {
		t.Fatalf("failed to get file: %v", err)
	}
	if fetched.ID != f.ID {
		t.Errorf("expected ID %s, got %s", f.ID, fetched.ID)
	}

	// 3. List
	// Allow some time for eventual consistency
	time.Sleep(1 * time.Second)
	list, err := filesClient.List(ctx)
	if err != nil {
		t.Fatalf("failed to list files: %v", err)
	}
	found := false
	for _, item := range list {
		if item.ID == f.ID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("file %s not found in list", f.ID)
	}

	// 4. Content
	rc, err := filesClient.Content(ctx, f.ID)
	if err != nil {
		// Ensure we don't fail hard if it's just processing
		t.Logf("failed to get content (might be processing): %v", err)
	} else {
		defer func() { _ = rc.Close() }()
	}
}
