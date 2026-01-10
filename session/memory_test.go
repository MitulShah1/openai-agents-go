package session

import (
	"context"
	"testing"

	"github.com/openai/openai-go"
)

func TestMemorySession_GetNotFound(t *testing.T) {
	s := NewMemorySession()
	_, err := s.Get(context.Background(), "nonexistent")

	if err == nil {
		t.Fatal("expected error for nonexistent session")
	}

	if _, ok := err.(*NotFoundError); !ok {
		t.Errorf("expected NotFoundError, got %T", err)
	}
}

func TestMemorySession_AppendAndGet(t *testing.T) {
	s := NewMemorySession()
	ctx := context.Background()
	sessionID := "test-session"

	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Hello"),
		openai.SystemMessage("Hi there"),
	}

	// Append messages
	if err := s.Append(ctx, sessionID, messages); err != nil {
		t.Fatalf("failed to append: %v", err)
	}

	// Get messages
	retrieved, err := s.Get(ctx, sessionID)
	if err != nil {
		t.Fatalf("failed to get: %v", err)
	}

	if len(retrieved) != len(messages) {
		t.Errorf("expected %d messages, got %d", len(messages), len(retrieved))
	}
}

func TestMemorySession_Clear(t *testing.T) {
	s := NewMemorySession()
	ctx := context.Background()
	sessionID := "test-session"

	// Add messages
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Hello"),
	}
	s.Append(ctx, sessionID, messages)

	// Clear
	if err := s.Clear(ctx, sessionID); err != nil {
		t.Fatalf("failed to clear: %v", err)
	}

	// Verify empty
	retrieved, err := s.Get(ctx, sessionID)
	if err != nil {
		t.Fatalf("failed to get after clear: %v", err)
	}

	if len(retrieved) != 0 {
		t.Errorf("expected 0 messages after clear, got %d", len(retrieved))
	}
}

func TestMemorySession_Delete(t *testing.T) {
	s := NewMemorySession()
	ctx := context.Background()
	sessionID := "test-session"

	// Add messages
	s.Append(ctx, sessionID, []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Hello"),
	})

	// Delete
	if err := s.Delete(ctx, sessionID); err != nil {
		t.Fatalf("failed to delete: %v", err)
	}

	// Verify not found
	_, err := s.Get(ctx, sessionID)
	if err == nil {
		t.Error("expected error after delete")
	}
}

func TestMemorySession_MultipleAppends(t *testing.T) {
	s := NewMemorySession()
	ctx := context.Background()
	sessionID := "test-session"

	// First append
	s.Append(ctx, sessionID, []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Message 1"),
	})

	// Second append
	s.Append(ctx, sessionID, []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Message 2"),
	})

	// Verify both messages
	retrieved, err := s.Get(ctx, sessionID)
	if err != nil {
		t.Fatalf("failed to get: %v", err)
	}

	if len(retrieved) != 2 {
		t.Errorf("expected 2 messages, got %d", len(retrieved))
	}
}

func TestMemorySession_IsolatedSessions(t *testing.T) {
	s := NewMemorySession()
	ctx := context.Background()

	// Create two separate sessions
	s.Append(ctx, "session1", []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Session 1 message"),
	})

	s.Append(ctx, "session2", []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Session 2 message"),
	})

	// Verify isolation
	msg1, _ := s.Get(ctx, "session1")
	msg2, _ := s.Get(ctx, "session2")

	if len(msg1) != 1 || len(msg2) != 1 {
		t.Error("sessions are not isolated")
	}
}
