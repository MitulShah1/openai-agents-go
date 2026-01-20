package session

import (
	"context"
	"crypto/rand"
	"testing"

	"github.com/openai/openai-go/v3"
)

func TestWithCompression(t *testing.T) {
	ctx := context.Background()
	mem := NewMemorySession()
	compressed := WithCompression(mem)
	sessionID := testSessionID

	t.Run("basic append and get", func(t *testing.T) {
		msgs := []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("Hello World"),
			openai.SystemMessage("System instruction"),
		}

		if err := compressed.Append(ctx, sessionID, msgs); err != nil {
			t.Fatalf("Append failed: %v", err)
		}

		retrieved, err := compressed.Get(ctx, sessionID)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}

		if len(retrieved) != 2 {
			t.Errorf("Expected 2 messages, got %d", len(retrieved))
		}

		verifyMessageContent(t, retrieved[0], "user", "Hello World")
		verifyMessageContent(t, retrieved[1], "system", "System instruction")

		// Verify underlying storage
		raw, err := mem.Get(ctx, sessionID)
		if err != nil {
			t.Fatalf("Underlying Get failed: %v", err)
		}

		if len(raw) != 1 {
			t.Errorf("Expected 1 underlying message, got %d", len(raw))
		}
		verifyMessageContent(t, raw[0], "system", "AGENTS_GZIP_BASE64:")
	})

	t.Run("empty append", func(t *testing.T) {
		if err := compressed.Append(ctx, sessionID, nil); err != nil {
			t.Fatalf("Append nil failed: %v", err)
		}
		if err := compressed.Append(ctx, sessionID, []openai.ChatCompletionMessageParamUnion{}); err != nil {
			t.Fatalf("Append empty failed: %v", err)
		}

		// Should not add anything new
		raw, _ := mem.Get(ctx, sessionID)
		// Previous test added 1 blob
		if len(raw) != 1 {
			t.Errorf("Expected 1 underlying message, got %d", len(raw))
		}
	})

	t.Run("mixed content handling", func(t *testing.T) {
		// Append a plain uncompressed message to the underlying memory
		// This simulates a session that has both old (uncompressed) and new (compressed) data
		plainMsg := openai.UserMessage("I am plain text")
		if err := mem.Append(ctx, sessionID, []openai.ChatCompletionMessageParamUnion{plainMsg}); err != nil {
			t.Fatal(err)
		}

		retrieved, err := compressed.Get(ctx, sessionID)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}

		// content: [CompressedBlob(2 msgs), PlainMsg] => Total 3 messages
		if len(retrieved) != 3 {
			t.Errorf("Expected 3 messages, got %d", len(retrieved))
		}

		verifyMessageContent(t, retrieved[2], "user", "I am plain text")
	})

	t.Run("corrupt compressed data", func(t *testing.T) {
		// Inject a malformed compressed message
		badMsg := openai.SystemMessage("AGENTS_GZIP_BASE64:NotBase64Data!@#$")
		if err := mem.Append(ctx, sessionID, []openai.ChatCompletionMessageParamUnion{badMsg}); err != nil {
			t.Fatal(err)
		}

		_, err := compressed.Get(ctx, sessionID)
		if err == nil {
			t.Error("Expected error for corrupt data, got nil")
		}
	})

	t.Run("clear session", func(t *testing.T) {
		if err := compressed.Clear(ctx, sessionID); err != nil {
			t.Fatalf("Clear failed: %v", err)
		}

		retrieved, err := compressed.Get(ctx, sessionID)
		if err != nil {
			t.Fatal(err)
		}
		if len(retrieved) != 0 {
			t.Errorf("Expected empty session after clear, got %d messages", len(retrieved))
		}
	})
}

func TestCombinedUtilities(t *testing.T) {
	// Stack: Compression(Encryption(Memory))
	// Data -> Encrypt -> Compress -> Storage (Best practice for size? No, Entropy.)
	// Wait, code says:
	// compressed := WithCompression(encrypted)
	// compressed.Append calls encrypted.Append with "AGENTS_GZIP..." blob.
	// encrypted.Append calls memory.Append with "AGENTS_AES..." blob (encrypting the GZIP blob).
	// So Storage has "AGENTS_AES...{encrypted 'AGENTS_GZIP...'}"

	ctx := context.Background()
	mem := NewMemorySession()

	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatal(err)
	}

	encrypted := WithEncryption(mem, key)
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
	// Underlying memory should hold: Encrypted(Compressed(JSON))
	// So it should have ENCRYPTED prefix
	raw, _ := mem.Get(ctx, sessionID)

	if len(raw) != 1 {
		t.Fatalf("Expected 1 underlying message, got %d", len(raw))
	}

	verifyMessageContent(t, raw[0], "system", "AGENTS_AES_GCM_BASE64:")
}
