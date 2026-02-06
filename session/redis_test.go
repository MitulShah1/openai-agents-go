//go:build redis

package session

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/openai/openai-go/v3"
)

func TestIntegrationRedis(t *testing.T) {
	// Need REDIS_ADDR env var
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		t.Skip("skipping redis integration test: REDIS_ADDR not set")
	}

	store, err := New(Options{
		Addr: addr,
	})
	if err != nil {
		t.Fatalf("failed to create redis store: %v", err)
	}

	ctx := context.Background()
	sessionID := "test-session-" + time.Now().String()

	// Test Append
	msgs := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Hello"),
	}
	if err := store.Append(ctx, sessionID, msgs); err != nil {
		t.Fatalf("failed to append: %v", err)
	}

	// Test Get
	got, err := store.Get(ctx, sessionID)
	if err != nil {
		t.Fatalf("failed to get: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("expected 1 message, got %d", len(got))
	}

	// Test Compact (Placeholder logic)
	if err := store.Compact(ctx, sessionID, 100); err != nil {
		t.Fatalf("failed to compact: %v", err)
	}

	// Test Delete
	if err := store.Delete(ctx, sessionID); err != nil {
		t.Fatalf("failed to delete: %v", err)
	}
}
