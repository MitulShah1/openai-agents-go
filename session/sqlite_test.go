package session

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/openai/openai-go/v3"
)

func TestSQLiteSession_CreateAndGet(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	sess, err := NewSQLite(dbPath)
	if err != nil {
		t.Fatalf("Failed to create SQLite session: %v", err)
	}
	defer func() { _ = sess.(*SQLiteSession).Close() }()

	// Test Get on non-existent session
	_, err = sess.Get(context.Background(), "test-session")
	if err == nil {
		t.Fatal("Expected NotFoundError for non-existent session")
	}
	if _, ok := err.(*NotFoundError); !ok {
		t.Errorf("Expected NotFoundError, got %T", err)
	}
}

func TestSQLiteSession_AppendAndGet(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	sess, err := NewSQLite(dbPath)
	if err != nil {
		t.Fatalf("Failed to create SQLite session: %v", err)
	}
	defer func() { _ = sess.(*SQLiteSession).Close() }()

	//nolint:goconst // "test-session" used as example in multiple tests
	sessionID := "test-session"
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Hello"),
		openai.AssistantMessage("Hi there!"),
	}

	// Append messages
	err = sess.Append(context.Background(), sessionID, messages)
	if err != nil {
		t.Fatalf("Failed to append messages: %v", err)
	}

	// Get messages
	retrieved, err := sess.Get(context.Background(), sessionID)
	if err != nil {
		t.Fatalf("Failed to get messages: %v", err)
	}

	if len(retrieved) != len(messages) {
		t.Errorf("Expected %d messages, got %d", len(messages), len(retrieved))
	}
}

func TestSQLiteSession_MultipleAppends(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	sess, err := NewSQLite(dbPath)
	if err != nil {
		t.Fatalf("Failed to create SQLite session: %v", err)
	}
	defer func() { _ = sess.(*SQLiteSession).Close() }()

	sessionID := "test-session"

	// First append
	messages1 := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Message 1"),
	}
	err = sess.Append(context.Background(), sessionID, messages1)
	if err != nil {
		t.Fatalf("Failed to append first batch: %v", err)
	}

	// Second append
	messages2 := []openai.ChatCompletionMessageParamUnion{
		openai.AssistantMessage("Message 2"),
	}
	err = sess.Append(context.Background(), sessionID, messages2)
	if err != nil {
		t.Fatalf("Failed to append second batch: %v", err)
	}

	// Get all messages
	retrieved, err := sess.Get(context.Background(), sessionID)
	if err != nil {
		t.Fatalf("Failed to get messages: %v", err)
	}

	if len(retrieved) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(retrieved))
	}
}

func TestSQLiteSession_Clear(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	sess, err := NewSQLite(dbPath)
	if err != nil {
		t.Fatalf("Failed to create SQLite session: %v", err)
	}
	defer func() { _ = sess.(*SQLiteSession).Close() }()

	sessionID := "test-session"
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Hello"),
	}

	// Append messages
	_ = sess.Append(context.Background(), sessionID, messages)

	// Clear session
	err = sess.Clear(context.Background(), sessionID)
	if err != nil {
		t.Fatalf("Failed to clear session: %v", err)
	}

	// Verify empty
	retrieved, err := sess.Get(context.Background(), sessionID)
	if err != nil {
		t.Fatalf("Failed to get messages after clear: %v", err)
	}

	if len(retrieved) != 0 {
		t.Errorf("Expected 0 messages after clear, got %d", len(retrieved))
	}
}

func TestSQLiteSession_Delete(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	sess, err := NewSQLite(dbPath)
	if err != nil {
		t.Fatalf("Failed to create SQLite session: %v", err)
	}
	defer func() { _ = sess.(*SQLiteSession).Close() }()

	sessionID := "test-session"
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Hello"),
	}

	// Append messages
	_ = sess.Append(context.Background(), sessionID, messages)

	// Delete session
	err = sess.Delete(context.Background(), sessionID)
	if err != nil {
		t.Fatalf("Failed to delete session: %v", err)
	}

	// Verify deleted
	_, err = sess.Get(context.Background(), sessionID)
	if err == nil {
		t.Fatal("Expected NotFoundError after delete")
	}
	if _, ok := err.(*NotFoundError); !ok {
		t.Errorf("Expected NotFoundError, got %T", err)
	}
}

func TestSQLiteSession_DeleteNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	sess, err := NewSQLite(dbPath)
	if err != nil {
		t.Fatalf("Failed to create SQLite session: %v", err)
	}
	defer func() { _ = sess.(*SQLiteSession).Close() }()

	err = sess.Delete(context.Background(), "non-existent")
	if err == nil {
		t.Fatal("Expected error when deleting non-existent session")
	}
	if _, ok := err.(*NotFoundError); !ok {
		t.Errorf("Expected NotFoundError, got %T", err)
	}
}

func TestSQLiteSession_Persistence(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	sessionID := "persist-test"
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Persistent message"),
	}

	// Create session, add messages, close
	{
		sess, err := NewSQLite(dbPath)
		if err != nil {
			t.Fatalf("Failed to create SQLite session: %v", err)
		}
		_ = sess.Append(context.Background(), sessionID, messages)
		_ = sess.(*SQLiteSession).Close()
	}

	// Reopen and verify persistence
	{
		sess, err := NewSQLite(dbPath)
		if err != nil {
			t.Fatalf("Failed to reopen SQLite session: %v", err)
		}
		defer func() { _ = sess.(*SQLiteSession).Close() }()

		retrieved, err := sess.Get(context.Background(), sessionID)
		if err != nil {
			t.Fatalf("Failed to get messages after reopen: %v", err)
		}

		if len(retrieved) != len(messages) {
			t.Errorf("Expected %d persisted messages, got %d", len(messages), len(retrieved))
		}
	}
}

func TestSQLiteSession_InMemory(t *testing.T) {
	// Test in-memory database
	sess, err := NewSQLite(":memory:")
	if err != nil {
		t.Fatalf("Failed to create in-memory SQLite session: %v", err)
	}
	defer func() { _ = sess.(*SQLiteSession).Close() }()

	sessionID := "memory-test"
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("In-memory message"),
	}

	_ = sess.Append(context.Background(), sessionID, messages)
	retrieved, err := sess.Get(context.Background(), sessionID)
	if err != nil {
		t.Fatalf("Failed to get in-memory messages: %v", err)
	}

	if len(retrieved) != 1 {
		t.Errorf("Expected 1 in-memory message, got %d", len(retrieved))
	}
}

func TestSQLiteSession_ConcurrentAccess(t *testing.T) {
	t.Skip("Skipping concurrent access test - to be fixed in follow-up")
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	sess, err := NewSQLite(dbPath)
	if err != nil {
		t.Fatalf("Failed to create SQLite session: %v", err)
	}
	defer func() { _ = sess.(*SQLiteSession).Close() }()

	// Test concurrent appends to different sessions
	done := make(chan bool)
	for i := 0; i < 5; i++ {
		go func(id int) {
			sessionID := fmt.Sprintf("session-%d", id)
			messages := []openai.ChatCompletionMessageParamUnion{
				openai.UserMessage(fmt.Sprintf("Message from goroutine %d", id)),
			}
			_ = sess.Append(context.Background(), sessionID, messages)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 5; i++ {
		<-done
	}

	// Verify all sessions exist
	for i := 0; i < 5; i++ {
		sessionID := fmt.Sprintf("session-%d", i)
		_, err := sess.Get(context.Background(), sessionID)
		if err != nil {
			t.Errorf("Failed to get session %s: %v", sessionID, err)
		}
	}
}
