package session

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/openai/openai-go"
)

func TestFileSession_CreateDir(t *testing.T) {
	tempDir := filepath.Join(os.TempDir(), "test-sessions-"+t.Name())
	defer os.RemoveAll(tempDir)

	_, err := NewFileSession(tempDir)
	if err != nil {
		t.Fatalf("failed to create file session: %v", err)
	}

	// Verify directory exists
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Error("session directory was not created")
	}
}

func TestFileSession_AppendAndGet(t *testing.T) {
	tempDir := filepath.Join(os.TempDir(), "test-sessions-"+t.Name())
	defer os.RemoveAll(tempDir)

	s, err := NewFileSession(tempDir)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	ctx := context.Background()
	sessionID := "test-session"

	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Hello"),
		openai.SystemMessage("Hi there"),
	}

	// Append
	if err := s.Append(ctx, sessionID, messages); err != nil {
		t.Fatalf("failed to append: %v", err)
	}

	// Get
	retrieved, err := s.Get(ctx, sessionID)
	if err != nil {
		t.Fatalf("failed to get: %v", err)
	}

	if len(retrieved) != len(messages) {
		t.Errorf("expected %d messages, got %d", len(messages), len(retrieved))
	}
}

func TestFileSession_Persistence(t *testing.T) {
	tempDir := filepath.Join(os.TempDir(), "test-sessions-"+t.Name())
	defer os.RemoveAll(tempDir)

	ctx := context.Background()
	sessionID := "test-session"

	// Create first session instance
	s1, _ := NewFileSession(tempDir)
	s1.Append(ctx, sessionID, []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Persisted message"),
	})

	// Create second session instance (simulates restart)
	s2, _ := NewFileSession(tempDir)
	retrieved, err := s2.Get(ctx, sessionID)
	if err != nil {
		t.Fatalf("failed to get from new instance: %v", err)
	}

	if len(retrieved) != 1 {
		t.Error("messages were not persisted across instances")
	}
}

func TestFileSession_Clear(t *testing.T) {
	tempDir := filepath.Join(os.TempDir(), "test-sessions-"+t.Name())
	defer os.RemoveAll(tempDir)

	s, _ := NewFileSession(tempDir)
	ctx := context.Background()
	sessionID := "test-session"

	// Add and clear
	s.Append(ctx, sessionID, []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Message"),
	})

	if err := s.Clear(ctx, sessionID); err != nil {
		t.Fatalf("failed to clear: %v", err)
	}

	// Verify empty
	retrieved, _ := s.Get(ctx, sessionID)
	if len(retrieved) != 0 {
		t.Errorf("expected 0 messages after clear, got %d", len(retrieved))
	}
}

func TestFileSession_Delete(t *testing.T) {
	tempDir := filepath.Join(os.TempDir(), "test-sessions-"+t.Name())
	defer os.RemoveAll(tempDir)

	s, _ := NewFileSession(tempDir)
	ctx := context.Background()
	sessionID := "test-session"

	// Add and delete
	s.Append(ctx, sessionID, []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Message"),
	})

	if err := s.Delete(ctx, sessionID); err != nil {
		t.Fatalf("failed to delete: %v", err)
	}

	// Verify not found
	_, err := s.Get(ctx, sessionID)
	if _, ok := err.(*NotFoundError); !ok {
		t.Errorf("expected NotFoundError after delete, got %T", err)
	}

	// Verify file is gone
	path := s.sessionPath(sessionID)
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("session file still exists after delete")
	}
}

func TestFileSession_MultipleAppends(t *testing.T) {
	tempDir := filepath.Join(os.TempDir(), "test-sessions-"+t.Name())
	defer os.RemoveAll(tempDir)

	s, _ := NewFileSession(tempDir)
	ctx := context.Background()
	sessionID := "test-session"

	// Multiple appends
	s.Append(ctx, sessionID, []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Message 1"),
	})
	s.Append(ctx, sessionID, []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Message 2"),
	})

	// Verify both
	retrieved, _ := s.Get(ctx, sessionID)
	if len(retrieved) != 2 {
		t.Errorf("expected 2 messages, got %d", len(retrieved))
	}
}
