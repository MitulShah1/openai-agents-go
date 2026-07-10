package session

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"testing"

	"github.com/openai/openai-go/v3"
)

func TestWithEncryption(t *testing.T) {
	ctx := context.Background()
	mem := NewMemorySession()

	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatal(err)
	}

	encrypted, err := WithEncryption(mem, key)
	if err != nil {
		t.Fatal(err)
	}
	sessionID := "enc-session"

	t.Run("basic append and get", func(t *testing.T) {
		msgs := []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("Secret Message"),
		}

		if err := encrypted.Append(ctx, sessionID, msgs); err != nil {
			t.Fatalf("Append failed: %v", err)
		}

		retrieved, err := encrypted.Get(ctx, sessionID)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}

		if len(retrieved) != 1 {
			t.Errorf("Expected 1 message, got %d", len(retrieved))
		}

		verifyMessageContent(t, retrieved[0], "user", "Secret Message")

		// Verify underlying storage
		raw, err := mem.Get(ctx, sessionID)
		if err != nil {
			t.Fatalf("Underlying Get failed: %v", err)
		}

		if len(raw) != 1 {
			t.Errorf("Expected 1 underlying message, got %d", len(raw))
		}
		verifyMessageContent(t, raw[0], "system", "AGENTS_AES_GCM_BASE64:")
	})

	t.Run("invalid key length", func(t *testing.T) {
		// Key must be 16, 24, or 32 bytes
		badKey := make([]byte, 10)
		if _, err := WithEncryption(mem, badKey); err == nil {
			t.Error("Expected error for invalid key length, got nil")
		}
	})

	t.Run("mixed content handling", func(t *testing.T) {
		// Append plain text
		plainMsg := openai.UserMessage("Plain Data")
		if err := mem.Append(ctx, sessionID, []openai.ChatCompletionMessageParamUnion{plainMsg}); err != nil {
			t.Fatal(err)
		}

		retrieved, err := encrypted.Get(ctx, sessionID)
		if err != nil {
			t.Fatal(err)
		}

		// [Encrypted(1 msg), Plain(1 msg)]
		if len(retrieved) != 2 {
			t.Errorf("Expected 2 messages, got %d", len(retrieved))
		}
		verifyMessageContent(t, retrieved[1], "user", "Plain Data")
	})

	t.Run("corrupt ciphertext", func(t *testing.T) {
		// Create a fake encrypted message with invalid ciphertext
		// The prefix is valid, but the data is garbage
		badData := base64.StdEncoding.EncodeToString([]byte("bad-data-not-encrypted"))
		badMsg := openai.SystemMessage("AGENTS_AES_GCM_BASE64:" + badData)

		if err := mem.Append(ctx, sessionID, []openai.ChatCompletionMessageParamUnion{badMsg}); err != nil {
			t.Fatal(err)
		}

		_, err := encrypted.Get(ctx, sessionID)
		if err == nil {
			t.Error("Expected error for corrupt ciphertext")
		} else {
			// Should be decryption error or "ciphertext too short"
			t.Logf("Got expected error: %v", err)
		}
	})

	t.Run("clear session", func(t *testing.T) {
		if err := encrypted.Clear(ctx, sessionID); err != nil {
			t.Fatal(err)
		}

		retrieved, err := encrypted.Get(ctx, sessionID)
		if err != nil {
			t.Fatal(err)
		}
		if len(retrieved) != 0 {
			t.Errorf("Expected 0 messages after clear, got %d", len(retrieved))
		}
	})
}
