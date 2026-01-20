package session

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/openai/openai-go/v3"
)

func verifyMessageContent(t *testing.T, msg openai.ChatCompletionMessageParamUnion, role, contentContains string) {
	t.Helper()
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal message: %v", err)
	}
	jsonStr := string(data)

	// Verify role
	if !strings.Contains(jsonStr, fmt.Sprintf(`"role":"%s"`, role)) {
		t.Errorf("Expected role '%s', got JSON: %s", role, jsonStr)
	}

	// Verify content
	if !strings.Contains(jsonStr, contentContains) {
		t.Errorf("Expected content to contain '%s', got JSON: %s", contentContains, jsonStr)
	}
}

func TestWithCompression(t *testing.T) {
	ctx := context.Background()
	mem := NewMemorySession()
	compressed := WithCompression(mem)
	sessionID := "test-session"

	// 1. Append messages
	msgs := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Hello World"),
		openai.SystemMessage("System instruction"),
	}

	if err := compressed.Append(ctx, sessionID, msgs); err != nil {
		t.Fatalf("Append failed: %v", err)
	}

	// 2. Verify retrieval
	retrieved, err := compressed.Get(ctx, sessionID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if len(retrieved) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(retrieved))
	}

	verifyMessageContent(t, retrieved[0], "user", "Hello World")
	verifyMessageContent(t, retrieved[1], "system", "System instruction")

	// 3. Verify underlying storage
	raw, err := mem.Get(ctx, sessionID)
	if err != nil {
		t.Fatalf("Underlying Get failed: %v", err)
	}

	if len(raw) != 1 {
		t.Errorf("Expected 1 underlying message (compressed blob), got %d", len(raw))
	}

	// Should be system message with compressed prefix
	verifyMessageContent(t, raw[0], "system", "AGENTS_GZIP_BASE64:")
}

func TestWithEncryption(t *testing.T) {
	ctx := context.Background()
	mem := NewMemorySession()

	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatal(err)
	}

	encrypted := WithEncryption(mem, key)
	sessionID := "enc-session"

	// 1. Append messages
	msgs := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Secret Message"),
	}

	if err := encrypted.Append(ctx, sessionID, msgs); err != nil {
		t.Fatalf("Append failed: %v", err)
	}

	// 2. Verify retrieval
	retrieved, err := encrypted.Get(ctx, sessionID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if len(retrieved) != 1 {
		t.Errorf("Expected 1 message, got %d", len(retrieved))
	}

	verifyMessageContent(t, retrieved[0], "user", "Secret Message")

	// 3. Verify underlying storage
	raw, err := mem.Get(ctx, sessionID)
	if err != nil {
		t.Fatalf("Underlying Get failed: %v", err)
	}

	if len(raw) != 1 {
		t.Errorf("Expected 1 underlying message, got %d", len(raw))
	}

	// Should be system message with enc prefix
	verifyMessageContent(t, raw[0], "system", "AGENTS_AES_GCM_BASE64:")
}

func TestCombinedUtilities(t *testing.T) {
	// Stack: Compression(Encryption(Memory))
	// Data -> Encrypt -> Compress -> Storage (Best practice)
	ctx := context.Background()
	mem := NewMemorySession()

	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatal(err)
	}

	// Wrap memory with encryption (inner)
	encrypted := WithEncryption(mem, key)
	// Wrap encryption with compression (outer)
	compressed := WithCompression(encrypted)

	sessionID := "combined-session"
	msgs := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Combined Test"),
	}

	if err := compressed.Append(ctx, sessionID, msgs); err != nil {
		t.Fatalf("Append failed: %v", err)
	}

	// Verify Get (via outer wrapper)
	retrieved, err := compressed.Get(ctx, sessionID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	verifyMessageContent(t, retrieved[0], "user", "Combined Test")

	// Verify Underlying Content
	// Underlying memory should hold: Compressed(Encrypted(JSON))
	// So it should have COMPRESSED prefix
	raw, _ := mem.Get(ctx, sessionID)

	verifyMessageContent(t, raw[0], "system", "AGENTS_AES_GCM_BASE64:")
}
